package sim

import (
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"math"
)

type Engine struct {
	Catalog content.Catalog
	rng     uint64
}

const OfflineCapTicks int64 = 8 * 60 * 60
const DefaultAutomationBudget int64 = 1000

// ReconcileOffline advances only safe production. Hostile actions and battles
// are intentionally not represented here, so closing the game cannot cause
// destructive outcomes. elapsedTicks is supplied by the persistence adapter.
func (e *Engine) ReconcileOffline(s game.State, elapsedTicks int64) (game.State, int64) {
	if elapsedTicks < 0 {
		return s, 0
	}
	if elapsedTicks > OfflineCapTicks {
		elapsedTicks = OfflineCapTicks
	}
	// Offline time only resolves safe production. Preserve hostile state and
	// the battle RNG stream so closing the game cannot decide an attack.
	activeBattle := s.ActiveBattle
	rngState := s.RNGState
	beforeCompute := s.Resources.Compute
	beforeCapital := mathx.Amount(0)
	if s.Corp != nil {
		beforeCapital = s.Corp.Capital
	}
	s.ActiveBattle = nil
	s, _ = e.advanceEvents(s, elapsedTicks, true)
	s.ActiveBattle = activeBattle
	s.RNGState = rngState
	computeProduced, _ := s.Resources.Compute.Add(-beforeCompute)
	capitalProduced := mathx.Amount(0)
	if s.Corp != nil {
		capitalProduced, _ = s.Corp.Capital.Add(-beforeCapital)
	}
	s.OfflineSummary = &game.OfflineSummary{Ticks: elapsedTicks, ComputeProduced: computeProduced, CapitalProduced: capitalProduced}
	return s, elapsedTicks
}

func New(c content.Catalog, seed int64) *Engine {
	return &Engine{Catalog: c, rng: uint64(seed) + 0x9e3779b97f4a7c15}
}
func (e *Engine) next() uint64 { e.rng ^= e.rng << 7; e.rng ^= e.rng >> 9; return e.rng }
func NewState(seed int64) game.State {
	return game.State{SchemaVersion: game.CurrentSchema, GameVersion: "0.1.0", CampaignID: "runner-01", Faction: game.Runner, WorldSeed: seed, RNGState: uint64(seed) + 0x9e3779b97f4a7c15, Resources: game.Resources{Credits: mathx.New(25), Compute: mathx.New(2), BotCapacity: 20}}
}

func NewCorporationState(seed int64) game.State {
	return game.State{
		SchemaVersion: game.CurrentSchema, GameVersion: "0.1.0", CampaignID: "corp-01",
		Faction: game.Corporation, WorldSeed: seed,
		RNGState: uint64(seed) + 0x9e3779b97f4a7c15,
		Corp: &game.CorpState{
			Capital: mathx.New(100), Compute: mathx.New(2), ICE: 3, ICECapacity: 5,
			DemandBP:     10000,
			ReputationBP: 8000, Assets: []game.CorpAsset{{ID: "starter_client", Name: "Starter Client", CapitalPerTick: 2, Integrity: 100, Operational: true}},
		},
	}
}

const EventLogLimit = 200

func (e *Engine) Apply(s game.State, c game.Command) (game.State, []game.Event, error) {
	next, events, err := e.apply(s, c)
	if err == nil && len(events) > 0 {
		next.EventLog = appendEventLog(next.EventLog, events...)
	}
	return next, events, err
}

func (e *Engine) apply(s game.State, c game.Command) (game.State, []game.Event, error) {
	if c.Kind == game.AdvanceTicks {
		if c.Ticks < 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "ticks cannot be negative"}
		}
		advanced, events := e.advanceEvents(s, c.Ticks, false)
		return advanced, events, nil
	}
	// Validate all spending before changing the snapshot.
	switch c.Kind {
	case game.CollectContract:
		if c.Quantity != 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "contract takes no quantity"}
		}
		n, err := s.Resources.Credits.Add(mathx.New(10))
		if err != nil {
			return s, nil, err
		}
		s.Resources.Credits = n
		return s, []game.Event{{Tick: s.LogicalTick, Type: "ResourceProduced", Message: "Street contract settled.", Amount: 10}}, nil
	case game.BuyGenerator:
		var g *content.Generator
		for i := range e.Catalog.Generators {
			if e.Catalog.Generators[i].ID == c.ID {
				g = &e.Catalog.Generators[i]
			}
		}
		if g == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown generator"}
		}
		cost := generatorCost(g.Cost, generatorCount(s, c.ID))
		if s.Resources.Credits < mathx.New(cost) {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: fmt.Sprintf("need %d credits", cost)}
		}
		s.Resources.Credits -= mathx.New(cost)
		upsertGenerator(&s, c.ID)
		s.Resources.BotCapacity += g.BotCapacity
		return s, []game.Event{{Tick: s.LogicalTick, Type: "GeneratorPurchased", Message: g.Name + " online."}}, nil
	case game.QueueFabrication:
		if !hasUpgrade(s, "secondhand_compiler") {
			return s, nil, &game.CommandError{Reason: game.ReasonUpgradeLocked, Detail: "buy the Secondhand Compiler first"}
		}
		if c.Quantity <= 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "quantity must be positive"}
		}
		if s.Fabrication != nil {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "fabrication already queued"}
		}
		if s.Resources.Bots+c.Quantity > s.Resources.BotCapacity {
			return s, nil, &game.CommandError{Reason: game.ReasonCapacityReached, Detail: "bot capacity reached"}
		}
		computeCost, err := mathx.New(10).MulUnits(c.Quantity)
		if err != nil || s.Resources.Compute < computeCost {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCompute, Detail: "fabrication requires 10 compute and 2 credits"}
		}
		creditCost, err := mathx.New(2).MulUnits(c.Quantity)
		if err != nil || s.Resources.Credits < creditCost {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: "fabrication requires 2 credits per bot"}
		}
		s.Resources.Compute -= computeCost
		s.Resources.Credits -= creditCost
		s.Fabrication = &game.Fabrication{RemainingTicks: e.fabricationTicks(s), Quantity: c.Quantity}
		return s, []game.Event{{Tick: s.LogicalTick, Type: "FabricationQueued", Message: "Bot compiler queued."}}, nil
	case game.BuyUpgrade:
		var upgrade *content.Upgrade
		for i := range e.Catalog.Upgrades {
			if e.Catalog.Upgrades[i].ID == c.ID {
				upgrade = &e.Catalog.Upgrades[i]
			}
		}
		if upgrade == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown upgrade"}
		}
		if hasUpgrade(s, c.ID) {
			return s, nil, &game.CommandError{Reason: game.ReasonDuplicatePurchase, Detail: "upgrade already purchased"}
		}
		if s.Resources.Credits < mathx.New(upgrade.Cost) {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: fmt.Sprintf("need %d credits", upgrade.Cost)}
		}
		s.Resources.Credits -= mathx.New(upgrade.Cost)
		s.Upgrades = append(s.Upgrades, c.ID)
		s.Resources.BotCapacity += upgrade.Effects.BotCapacityBonus
		return s, []game.Event{{Tick: s.LogicalTick, Type: "UpgradePurchased", Message: upgrade.Name + " installed."}}, nil
	case game.StartBattle:
		if s.ActiveBattle != nil {
			return s, nil, &game.CommandError{Reason: game.ReasonBattleActive, Detail: "a battle is already active"}
		}
		if !hasUpgrade(s, "secondhand_compiler") {
			return s, nil, &game.CommandError{Reason: game.ReasonUpgradeLocked, Detail: "buy the Secondhand Compiler first"}
		}
		if c.Quantity <= 0 || c.Quantity > s.Resources.Bots {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientBots, Detail: "not enough bots committed"}
		}
		target, ok := e.target(c.ID)
		if !ok {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown target"}
		}
		if !e.isReachable(s, target) {
			return s, nil, &game.CommandError{Reason: game.ReasonTargetUnreachable, Detail: "target route is not reachable"}
		}
		if c.Doctrine != game.Balanced && c.Doctrine != game.Stealth && c.Doctrine != game.Overload {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown doctrine"}
		}
		s.Resources.Bots -= c.Quantity
		s.LastBattle = nil
		s.ActiveBattle = &game.Battle{TargetID: c.ID, Doctrine: c.Doctrine, CommittedBots: c.Quantity, LiveBots: c.Quantity, LayerIntegrity: target.Layers[0].Integrity, DefenderICE: target.Layers[0].ICE, TimeLimit: target.TimeLimit, StartingHeatBP: s.HeatBP}
		return s, []game.Event{{Tick: s.LogicalTick, Type: "BattleStarted", Message: "Breach underway: " + target.Name}}, nil
	case game.ReinforceBattle:
		if s.ActiveBattle == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonNoBattle, Detail: "no active battle"}
		}
		if s.ActiveBattle.InterventionUsed {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "battle intervention already used"}
		}
		if c.Quantity <= 0 || c.Quantity > s.Resources.Bots {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientBots, Detail: "not enough reserve bots"}
		}
		s.Resources.Bots -= c.Quantity
		s.ActiveBattle.LiveBots += c.Quantity
		s.ActiveBattle.CommittedBots += c.Quantity
		s.ActiveBattle.InterventionUsed = true
		return s, []game.Event{{Tick: s.LogicalTick, Type: "BattleReinforced", Message: fmt.Sprintf("%d reserve bots joined the breach.", c.Quantity)}}, nil
	case game.WithdrawBattle:
		if s.ActiveBattle == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonNoBattle, Detail: "no active battle"}
		}
		s.Resources.Bots += s.ActiveBattle.LiveBots
		s.LastBattle = battleResult(s.ActiveBattle, "withdrawn", s.ActiveBattle.LiveBots, 0, 0, s.HeatBP)
		s.ActiveBattle = nil
		return s, []game.Event{{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Withdrawal complete. Surviving bots returned."}}, nil
	case game.StartRecon:
		if !hasUpgrade(s, "street_contacts") {
			return s, nil, &game.CommandError{Reason: game.ReasonReconLocked, Detail: "buy Street Contacts first"}
		}
		if s.Recon != nil {
			return s, nil, &game.CommandError{Reason: game.ReasonReconActive, Detail: "recon is already running"}
		}
		target, ok := e.target(c.ID)
		if !ok {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown target"}
		}
		if !e.isReachable(s, target) {
			return s, nil, &game.CommandError{Reason: game.ReasonTargetUnreachable, Detail: "target route is not reachable"}
		}
		if s.Resources.Credits < mathx.New(target.ReconCost) {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: fmt.Sprintf("need %d credits for recon", target.ReconCost)}
		}
		s.Resources.Credits -= mathx.New(target.ReconCost)
		s.Recon = &game.ReconJob{TargetID: target.ID, RemainingTicks: e.reconTicks(s, target.ReconTicks)}
		return s, []game.Event{{Tick: s.LogicalTick, Type: "ReconStarted", Message: "Recon package deployed: " + target.Name}}, nil
	case game.EnableAutomation:
		if c.Quantity < 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "automation budget cannot be negative"}
		}
		switch c.ID {
		case "contracts":
			if !hasUpgrade(s, "autodialer") {
				return s, nil, &game.CommandError{Reason: game.ReasonUpgradeLocked, Detail: "buy the Autodialer first"}
			}
			s.Automation.Contracts = true
			s.Automation.ContractBudget = configuredBudget(c.Quantity, s.Automation.ContractBudget)
		case "fabrication":
			if !hasUpgrade(s, "packet_scheduler") {
				return s, nil, &game.CommandError{Reason: game.ReasonUpgradeLocked, Detail: "buy the Packet Scheduler first"}
			}
			s.Automation.Fabrication = true
			s.Automation.FabricationBudget = configuredBudget(c.Quantity, s.Automation.FabricationBudget)
		case "recon":
			if !hasUpgrade(s, "signal_profiler") {
				return s, nil, &game.CommandError{Reason: game.ReasonUpgradeLocked, Detail: "buy the Signal Profiler first"}
			}
			target, ok := e.target(c.TargetID)
			if !ok || !e.isReachable(s, target) {
				return s, nil, &game.CommandError{Reason: game.ReasonTargetUnreachable, Detail: "recon target is not reachable"}
			}
			s.Automation.ReconTargetID = target.ID
			s.Automation.ReconBudget = configuredBudget(c.Quantity, s.Automation.ReconBudget)
		default:
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown automation policy"}
		}
		return s, []game.Event{{Tick: s.LogicalTick, Type: "AutomationEnabled", Message: "Automation policy enabled: " + c.ID}}, nil
	case game.BuyResearch:
		node, ok := e.researchNode(c.ID)
		if !ok {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown research node"}
		}
		if contains(s.Research, node.ID) {
			return s, nil, &game.CommandError{Reason: game.ReasonDuplicatePurchase, Detail: "research already completed"}
		}
		for _, prerequisite := range node.Prerequisites {
			if !contains(s.Research, prerequisite) {
				return s, nil, &game.CommandError{Reason: game.ReasonResearchLocked, Detail: "research prerequisite missing: " + prerequisite}
			}
		}
		costData := e.researchCost(s, node.CostData)
		if s.Resources.Data < mathx.New(costData) {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientData, Detail: fmt.Sprintf("need %d data", costData)}
		}
		s.Resources.Data -= mathx.New(costData)
		s.Research = append(s.Research, node.ID)
		return s, []game.Event{{Tick: s.LogicalTick, Type: "ResearchCompleted", Message: node.Name + " completed."}}, nil
	case game.IdentityBurn:
		if s.Faction != game.Runner {
			return s, nil, &game.CommandError{Reason: game.ReasonFactionMismatch, Detail: "Identity Burn is a Runner reset"}
		}
		reward := integerSqrt(s.RunExtractedData / 10000)
		if reward <= 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonPrestigeLocked, Detail: "extract at least 10,000 data before Identity Burn"}
		}
		ghostKeys := s.GhostKeys + reward
		newState := NewState(s.WorldSeed)
		newState.GhostKeys = ghostKeys
		newState.PrestigeCount = s.PrestigeCount + 1
		newState.Resources.Credits = mathx.New(25 + ghostKeys*5)
		newState.GameVersion = s.GameVersion
		return newState, []game.Event{{Tick: s.LogicalTick, Type: "PrestigeCompleted", Message: fmt.Sprintf("Identity Burn complete. %d Ghost Keys retained.", ghostKeys), Amount: reward}}, nil
	case game.CorpBuildAsset:
		if s.Faction != game.Corporation || s.Corp == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonFactionMismatch, Detail: "corporation campaign required"}
		}
		definition, ok := e.corpAsset(c.ID)
		if !ok {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown corporation asset"}
		}
		if s.Corp.Capital < mathx.New(definition.CostCapital) {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: fmt.Sprintf("need %d capital", definition.CostCapital)}
		}
		s.Corp.Capital -= mathx.New(definition.CostCapital)
		s.Corp.Assets = append(s.Corp.Assets, game.CorpAsset{ID: definition.ID, Name: definition.Name, CapitalPerTick: definition.CapitalPerTick, UpkeepPerTick: definition.UpkeepPerTick, Integrity: definition.Integrity, Operational: true})
		return s, []game.Event{{Tick: s.LogicalTick, Type: "AssetBuilt", Message: definition.Name + " commissioned."}}, nil
	case game.CorpBuyICE:
		if s.Faction != game.Corporation || s.Corp == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonFactionMismatch, Detail: "corporation campaign required"}
		}
		if c.Quantity <= 0 || s.Corp.ICE+c.Quantity > s.Corp.ICECapacity {
			return s, nil, &game.CommandError{Reason: game.ReasonCapacityReached, Detail: "ICE capacity reached"}
		}
		cost, err := mathx.New(10).MulUnits(c.Quantity)
		if err != nil || s.Corp.Capital < cost {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: "ICE requires 10 capital per unit"}
		}
		s.Corp.Capital -= cost
		s.Corp.ICE += c.Quantity
		return s, []game.Event{{Tick: s.LogicalTick, Type: "ICEBuilt", Message: fmt.Sprintf("%d ICE units restocked.", c.Quantity)}}, nil
	case game.CorpDeployICE:
		if s.Faction != game.Corporation || s.Corp == nil || s.ActiveBattle == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonNoBattle, Detail: "no active Corporation incursion"}
		}
		if c.Quantity <= 0 || c.Quantity > s.Corp.ICE {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: "not enough reserve ICE"}
		}
		if s.ActiveBattle.InterventionUsed {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "emergency response already used"}
		}
		s.Corp.ICE -= c.Quantity
		s.ActiveBattle.DefenderICE += c.Quantity
		s.ActiveBattle.InterventionUsed = true
		return s, []game.Event{{Tick: s.LogicalTick, Type: "ICEDeployed", Message: fmt.Sprintf("%d reserve ICE deployed.", c.Quantity)}}, nil
	case game.CorpRunResearch:
		if s.Faction != game.Corporation || s.Corp == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonFactionMismatch, Detail: "corporation campaign required"}
		}
		cost := mathx.New(200)
		if s.Corp.Capital < cost {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: "R&D requires 200 capital"}
		}
		s.Corp.Capital -= cost
		s.Corp.Research, _ = s.Corp.Research.Add(mathx.New(10))
		s.Corp.RunValuation += 200
		return s, []game.Event{{Tick: s.LogicalTick, Type: "CorpResearch", Message: "R&D department completed a systems study."}}, nil
	case game.CorpExpandMarket:
		if s.Faction != game.Corporation || s.Corp == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonFactionMismatch, Detail: "corporation campaign required"}
		}
		cost := mathx.New(150)
		if s.Corp.Capital < cost {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: "market expansion requires 150 capital"}
		}
		s.Corp.Capital -= cost
		s.Corp.Influence, _ = s.Corp.Influence.Add(mathx.New(10))
		s.Corp.DemandBP = minInt64(10000, s.Corp.DemandBP+1000)
		s.Corp.RunValuation += 150
		return s, []game.Event{{Tick: s.LogicalTick, Type: "MarketExpanded", Message: "New service demand secured."}}, nil
	case game.CorpEnableMaintenance:
		if s.Faction != game.Corporation || s.Corp == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonFactionMismatch, Detail: "corporation campaign required"}
		}
		if c.Quantity < 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "maintenance budget cannot be negative"}
		}
		s.Corp.MaintenanceAuto = true
		s.Corp.MaintenanceBudget = configuredBudget(c.Quantity, s.Corp.MaintenanceBudget)
		return s, []game.Event{{Tick: s.LogicalTick, Type: "MaintenanceEnabled", Message: "Automated maintenance budget enabled."}}, nil
	case game.CorpRestructure:
		if s.Faction != game.Corporation || s.Corp == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonFactionMismatch, Detail: "corporation campaign required"}
		}
		reward := integerSqrt(s.Corp.RunValuation / 1000)
		if reward <= 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonPrestigeLocked, Detail: "reach 1,000 valuation before Restructuring"}
		}
		newState := NewCorporationState(s.WorldSeed)
		newState.GameVersion = s.GameVersion
		newState.Corp.PatentShares = s.Corp.PatentShares + reward
		newState.Corp.Capital = mathx.New(100 + newState.Corp.PatentShares*10)
		newState.PrestigeCount = s.PrestigeCount + 1
		return newState, []game.Event{{Tick: s.LogicalTick, Type: "RestructuringCompleted", Message: fmt.Sprintf("Restructuring complete. %d Patent Shares retained.", newState.Corp.PatentShares), Amount: reward}}, nil
	default:
		return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unsupported command"}
	}
}

func appendEventLog(log []game.Event, events ...game.Event) []game.Event {
	log = append(log, events...)
	if len(log) > EventLogLimit {
		log = log[len(log)-EventLogLimit:]
	}
	return log
}
func (e *Engine) Advance(s game.State, ticks int64) game.State {
	advanced, _ := e.advanceEvents(s, ticks, false)
	return advanced
}

func (e *Engine) advanceEvents(s game.State, ticks int64, safeOnly bool) (game.State, []game.Event) {
	var events []game.Event
	for i := int64(0); i < ticks; i++ {
		s.LogicalTick++
		s.RNGState ^= s.RNGState << 7
		s.RNGState ^= s.RNGState >> 9
		if s.Faction == game.Corporation && s.Corp != nil {
			var corporationEvents []game.Event
			s, corporationEvents = e.advanceCorporation(s, safeOnly)
			events = append(events, corporationEvents...)
		}
		if s.Faction != game.Corporation {
			rate := e.ComputeRate(s)
			produced := mathx.New(rate)
			s.Resources.Compute, _ = s.Resources.Compute.Add(produced)
		}
		if s.Fabrication != nil {
			s.Fabrication.RemainingTicks--
			if s.Fabrication.RemainingTicks <= 0 {
				s.Resources.Bots += s.Fabrication.Quantity
				s.Fabrication = nil
			}
		}
		s, events = e.advanceAutomation(s, events)
		if s.Recon != nil {
			s.Recon.RemainingTicks--
			if s.Recon.RemainingTicks <= 0 {
				if s.Intel == nil {
					s.Intel = map[string]int64{}
				}
				if target, ok := e.target(s.Recon.TargetID); ok {
					s.Intel[target.ID] += target.IntelGain
					if s.Intel[target.ID] > 100 {
						s.Intel[target.ID] = 100
					}
					events = append(events, game.Event{Tick: s.LogicalTick, Type: "ReconCompleted", Message: fmt.Sprintf("Recon complete: %s intel %d%%.", target.Name, s.Intel[target.ID])})
				}
				s.Recon = nil
			}
		}
		if !safeOnly && s.Faction == game.Corporation && s.Corp != nil && s.ActiveBattle == nil {
			s.Corp.AttackCounter++
			if s.Corp.AttackCounter >= 30 {
				s, events = e.startCorporationIncursion(s, events)
			}
		}
		if s.ActiveBattle != nil {
			s, events = e.advanceBattle(s, events)
		}
	}
	return s, events
}

func (e *Engine) advanceAutomation(s game.State, events []game.Event) (game.State, []game.Event) {
	normalizeAutomationBudgets(&s.Automation)
	if s.Automation.Contracts && s.LogicalTick%e.automationInterval(s) == 0 {
		remaining := s.Automation.ContractBudget - s.Automation.ContractEarned
		if remaining <= 0 {
			s.Automation.Contracts = false
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "AutomationStopped", Message: "Autodialer stopped: contract budget reached."})
		} else {
			amount := minInt64(10, remaining)
			s.Resources.Credits, _ = s.Resources.Credits.Add(mathx.New(amount))
			s.Automation.ContractEarned += amount
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "AutomationProduced", Message: "Autodialer settled a safe contract.", Amount: amount})
			if s.Automation.ContractEarned >= s.Automation.ContractBudget {
				s.Automation.Contracts = false
				events = append(events, game.Event{Tick: s.LogicalTick, Type: "AutomationStopped", Message: "Autodialer stopped: contract budget reached."})
			}
		}
	}
	if s.Automation.Fabrication && s.Fabrication == nil && hasUpgrade(s, "secondhand_compiler") && s.Resources.Bots < s.Resources.BotCapacity && s.Resources.Compute >= mathx.New(10) && s.Resources.Credits >= mathx.New(2) {
		if s.Automation.FabricationSpent+2 > s.Automation.FabricationBudget {
			s.Automation.Fabrication = false
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "AutomationStopped", Message: "Packet Scheduler stopped: fabrication budget reached."})
		} else {
			s.Resources.Compute -= mathx.New(10)
			s.Resources.Credits -= mathx.New(2)
			s.Automation.FabricationSpent += 2
			s.Fabrication = &game.Fabrication{RemainingTicks: e.fabricationTicks(s), Quantity: 1}
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "AutomationQueued", Message: "Packet Scheduler queued one bot."})
		}
	}
	if s.Automation.ReconTargetID != "" && s.Recon == nil {
		if target, ok := e.target(s.Automation.ReconTargetID); ok && e.isReachable(s, target) && s.Intel[target.ID] < 100 && hasUpgrade(s, "street_contacts") && s.Resources.Credits >= mathx.New(target.ReconCost) {
			if s.Automation.ReconSpent+target.ReconCost > s.Automation.ReconBudget {
				s.Automation.ReconTargetID = ""
				events = append(events, game.Event{Tick: s.LogicalTick, Type: "AutomationStopped", Message: "Signal Profiler stopped: recon budget reached."})
			} else {
				s.Resources.Credits -= mathx.New(target.ReconCost)
				s.Automation.ReconSpent += target.ReconCost
				s.Recon = &game.ReconJob{TargetID: target.ID, RemainingTicks: e.reconTicks(s, target.ReconTicks)}
				events = append(events, game.Event{Tick: s.LogicalTick, Type: "AutomationQueued", Message: "Signal Profiler queued recon: " + target.Name})
			}
		}
	}
	return s, events
}

func configuredBudget(requested, existing int64) int64 {
	if requested > 0 {
		return requested
	}
	if existing > 0 {
		return existing
	}
	return DefaultAutomationBudget
}

func normalizeAutomationBudgets(automation *game.Automation) {
	if automation.Contracts && automation.ContractBudget <= 0 {
		automation.ContractBudget = DefaultAutomationBudget
	}
	if automation.Fabrication && automation.FabricationBudget <= 0 {
		automation.FabricationBudget = DefaultAutomationBudget
	}
	if automation.ReconTargetID != "" && automation.ReconBudget <= 0 {
		automation.ReconBudget = DefaultAutomationBudget
	}
}

func (e *Engine) advanceCorporation(s game.State, safeOnly bool) (game.State, []game.Event) {
	var events []game.Event
	if s.Corp.MaintenanceAuto && s.Corp.MaintenanceBudget <= 0 {
		s.Corp.MaintenanceBudget = DefaultAutomationBudget
	}
	rate := int64(2)
	upkeep := int64(0)
	for _, asset := range s.Corp.Assets {
		if asset.Operational {
			rate += asset.CapitalPerTick
			upkeep += asset.UpkeepPerTick
		}
	}
	if s.Corp.DemandBP <= 0 {
		s.Corp.DemandBP = 10000
	}
	rate = rate * s.Corp.DemandBP / 10000
	s.Corp.Capital, _ = s.Corp.Capital.Add(mathx.New(rate))
	if upkeep > 0 && s.Corp.Capital >= mathx.New(upkeep) {
		s.Corp.Capital -= mathx.New(upkeep)
	} else if upkeep > 0 {
		for i := range s.Corp.Assets {
			if s.Corp.Assets[i].UpkeepPerTick > 0 {
				s.Corp.Assets[i].Operational = false
			}
		}
	}
	s.Corp.RunValuation += rate
	s.Corp.Compute, _ = s.Corp.Compute.Add(mathx.New(2))
	if s.Corp.MaintenanceAuto && !safeOnly && s.Corp.ICE < s.Corp.ICECapacity {
		if s.Corp.MaintenanceSpent+10 > s.Corp.MaintenanceBudget {
			s.Corp.MaintenanceAuto = false
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "MaintenanceStopped", Message: "Automated maintenance stopped: budget reached."})
		} else if s.Corp.Capital >= mathx.New(10) {
			s.Corp.Capital -= mathx.New(10)
			s.Corp.ICE++
			s.Corp.MaintenanceSpent += 10
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "MaintenanceRestocked", Message: "Automated maintenance restocked one ICE unit."})
		}
	}
	return s, events
}

func (e *Engine) startCorporationIncursion(s game.State, events []game.Event) (game.State, []game.Event) {
	target, ok := e.target("t_01")
	if !ok || len(target.Layers) == 0 {
		return s, events
	}
	deployed := int64(0)
	if s.Corp.ICE > 0 {
		deployed = minInt64(s.Corp.ICE, 3)
		s.Corp.ICE -= deployed
	}
	s.Corp.AttackCounter = 0
	s.ActiveBattle = &game.Battle{TargetID: target.ID, AttackerFaction: game.Runner, CommittedBots: 8, LiveBots: 8, LayerIntegrity: target.Layers[0].Integrity, DefenderICE: target.Layers[0].ICE + deployed, TimeLimit: target.TimeLimit, StartingHeatBP: s.HeatBP}
	return s, append(events, game.Event{Tick: s.LogicalTick, Type: "BattleStarted", Message: "Emergency audit: suspicious ingress detected."})
}

func (e *Engine) advanceBattle(s game.State, events []game.Event) (game.State, []game.Event) {
	b := s.ActiveBattle
	target, ok := e.target(b.TargetID)
	if !ok {
		return s, events
	}
	layer := target.Layers[b.LayerIndex]
	attackMultiplier := int64(100)
	heatMultiplier := int64(100)
	effects := e.researchEffects(s)
	upgradeEffects := e.upgradeEffects(s)
	switch b.Doctrine {
	case game.Stealth:
		attackMultiplier, heatMultiplier = 82, 55
		heatMultiplier -= effects.StealthHeatReductionPercent
		attackMultiplier += upgradeEffects.StealthAttackBonusPercent
	case game.Overload:
		attackMultiplier, heatMultiplier = 125, 170
	}
	attackMultiplier += effects.AttackBonusPercent
	heatMultiplier -= upgradeEffects.HeatReductionPercent
	if heatMultiplier < 0 {
		heatMultiplier = 0
	}
	// Battle variation is deliberately bounded and comes from the persisted
	// state stream. It cannot decide an otherwise impossible battle, but keeps
	// equal commitments from producing identical-looking ticks.
	s.RNGState ^= s.RNGState << 7
	s.RNGState ^= s.RNGState >> 9
	variationBP := int64(9500 + s.RNGState%1001)
	damage := b.LiveBots * attackMultiplier / (100 + b.DefenderICE)
	damage = damage * variationBP / 10000
	if damage < 1 {
		damage = 1
	}
	b.LayerIntegrity -= damage
	if b.DefenderICE > 0 {
		b.DefenderICE -= maxInt64(1, b.LiveBots/10)
		if b.DefenderICE < 0 {
			b.DefenderICE = 0
		}
	}
	b.ElapsedTicks++
	b.TraceBP += heatMultiplier
	if b.TraceBP > 10000 {
		b.TraceBP = 10000
	}
	s.HeatBP += heatMultiplier
	if s.HeatBP > 10000 {
		s.HeatBP = 10000
	}
	if layer.Counter > 0 && b.ElapsedTicks%3 == 0 && b.LiveBots > 0 {
		b.LiveBots--
		events = append(events, game.Event{Tick: s.LogicalTick, Type: "BotsLost", Message: "Countermeasure destroyed one bot."})
	}
	if b.LiveBots <= 0 {
		if s.Faction == game.Corporation {
			s.Corp.ReputationBP = minInt64(10000, s.Corp.ReputationBP+500)
			s.LastBattle = battleResult(b, "defended", 0, 0, 0, s.HeatBP)
		} else {
			s.LastBattle = battleResult(b, "defeat", 0, 0, 0, s.HeatBP)
		}
		events = append(events, game.Event{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Breach failed. The committed bots were lost."})
		s.ActiveBattle = nil
		return s, events
	}
	if b.LayerIntegrity <= 0 {
		events = append(events, game.Event{Tick: s.LogicalTick, Type: "LayerBreached", Message: fmt.Sprintf("Layer %d breached.", b.LayerIndex+1)})
		b.LayerIndex++
		if b.LayerIndex >= len(target.Layers) {
			if s.Faction == game.Corporation {
				s.Corp.ExposureBP = minInt64(10000, s.Corp.ExposureBP+2000)
				s.Corp.ReputationBP = maxInt64(0, s.Corp.ReputationBP-1000)
				s.LastBattle = battleResult(b, "breach", b.LiveBots, 0, 0, s.HeatBP)
				s.ActiveBattle = nil
				events = append(events, game.Event{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Emergency audit failed. Sensitive asset exposed."})
				return s, events
			}
			s.Resources.Bots += b.LiveBots
			s.Resources.Credits, _ = s.Resources.Credits.Add(mathx.New(target.LootCredits))
			s.Resources.Data, _ = s.Resources.Data.Add(mathx.New(target.LootData))
			s.RunExtractedData += target.LootData
			s.LastBattle = battleResult(b, "victory", b.LiveBots, target.LootCredits, target.LootData, s.HeatBP)
			if !contains(s.CompletedTargets, target.ID) {
				s.CompletedTargets = append(s.CompletedTargets, target.ID)
			}
			s.ActiveBattle = nil
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Breach successful. Loot extracted."})
			return s, events
		}
		b.LayerIntegrity = target.Layers[b.LayerIndex].Integrity
		b.DefenderICE = target.Layers[b.LayerIndex].ICE
	}
	if b.ElapsedTicks >= b.TimeLimit {
		if s.Faction == game.Corporation {
			s.Corp.ReputationBP = minInt64(10000, s.Corp.ReputationBP+250)
			s.LastBattle = battleResult(b, "defended", 0, 0, 0, s.HeatBP)
			s.ActiveBattle = nil
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Audit timed out. The asset remains secured."})
			return s, events
		}
		s.Resources.Bots += b.LiveBots
		s.LastBattle = battleResult(b, "timeout", b.LiveBots, 0, 0, s.HeatBP)
		s.ActiveBattle = nil
		events = append(events, game.Event{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Breach timed out. Surviving bots returned."})
		return s, events
	}
	s.ActiveBattle = b
	return s, events
}

func battleResult(b *game.Battle, outcome string, survivors, credits, data, heatBP int64) *game.BattleResult {
	addedHeat := heatBP - b.StartingHeatBP
	if addedHeat < 0 {
		addedHeat = 0
	}
	return &game.BattleResult{
		TargetID: b.TargetID, Doctrine: b.Doctrine, Outcome: outcome,
		CommittedBots: b.CommittedBots, Survivors: survivors,
		BotsLost: b.CommittedBots - survivors, CreditsLooted: credits,
		DataLooted: data, TraceBP: b.TraceBP, HeatBeforeBP: b.StartingHeatBP,
		HeatAfterBP: heatBP, HeatAddedBP: addedHeat, ElapsedTicks: b.ElapsedTicks,
	}
}

func (e *Engine) target(id string) (content.Target, bool) {
	for _, target := range e.Catalog.Targets {
		if target.ID == id {
			return target, true
		}
	}
	return content.Target{}, false
}

func (e *Engine) corpAsset(id string) (content.CorpAssetDefinition, bool) {
	for _, asset := range e.Catalog.CorpAssets {
		if asset.ID == id {
			return asset, true
		}
	}
	return content.CorpAssetDefinition{}, false
}

func (e *Engine) researchNode(id string) (content.ResearchNode, bool) {
	for _, node := range e.Catalog.Research {
		if node.ID == id {
			return node, true
		}
	}
	return content.ResearchNode{}, false
}

func integerSqrt(value int64) int64 {
	if value <= 0 {
		return 0
	}
	var low, high int64 = 1, value
	for low <= high {
		mid := low + (high-low)/2
		if mid <= value/mid {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return high
}

// IdentityBurnReward exposes the deterministic preview used by the TUI.
func (e *Engine) IdentityBurnReward(s game.State) int64 {
	return integerSqrt(s.RunExtractedData / 10000)
}

func (e *Engine) isReachable(s game.State, target content.Target) bool {
	if target.Reachable {
		return true
	}
	return target.RequiresTargetID != "" && contains(s.CompletedTargets, target.RequiresTargetID)
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func hasUpgrade(s game.State, id string) bool {
	for _, upgrade := range s.Upgrades {
		if upgrade == id {
			return true
		}
	}
	return false
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// ComputeRate returns the safe passive production rate in whole compute units
// per logical second. It is presentation data derived from state, not a second
// source of truth.
func (e *Engine) ComputeRate(s game.State) int64 {
	rate := int64(2)
	for _, g := range e.Catalog.Generators {
		for _, owned := range s.Generators {
			if owned.ID == g.ID {
				rate += owned.Count * g.ComputePerSecond
			}
		}
	}
	return rate + e.researchEffects(s).ComputePerSecondBonus + e.upgradeEffects(s).ComputePerSecondBonus
}

// EstimateBattle provides a deterministic, intentionally conservative preflight
// projection. It ignores random variation and counter losses, so the TUI labels
// it as a baseline rather than promising an outcome.
func (e *Engine) EstimateBattle(s game.State, target content.Target, doctrine game.Doctrine, commitment int64) (int64, int64) {
	if commitment <= 0 || len(target.Layers) == 0 {
		return 0, 0
	}
	attack := int64(100) + e.researchEffects(s).AttackBonusPercent
	heat := int64(100)
	upgrades := e.upgradeEffects(s)
	switch doctrine {
	case game.Stealth:
		attack = 82 + e.researchEffects(s).AttackBonusPercent + upgrades.StealthAttackBonusPercent
		heat = 55 - e.researchEffects(s).StealthHeatReductionPercent
	case game.Overload:
		attack = 125 + e.researchEffects(s).AttackBonusPercent
		heat = 170
	}
	heat -= upgrades.HeatReductionPercent
	if heat < 0 {
		heat = 0
	}
	ticks := int64(0)
	for _, layer := range target.Layers {
		damage := commitment * attack / (100 + layer.ICE)
		if damage < 1 {
			damage = 1
		}
		ticks += (layer.Integrity + damage - 1) / damage
	}
	if ticks > target.TimeLimit {
		ticks = target.TimeLimit
	}
	return ticks, ticks * heat
}

func (e *Engine) upgradeEffects(s game.State) content.UpgradeEffects {
	var effects content.UpgradeEffects
	for _, upgrade := range e.Catalog.Upgrades {
		if !hasUpgrade(s, upgrade.ID) {
			continue
		}
		effects.ComputePerSecondBonus += upgrade.Effects.ComputePerSecondBonus
		effects.HeatReductionPercent += upgrade.Effects.HeatReductionPercent
		effects.StealthAttackBonusPercent += upgrade.Effects.StealthAttackBonusPercent
		effects.BotCapacityBonus += upgrade.Effects.BotCapacityBonus
		effects.FabricationTicksReduction += upgrade.Effects.FabricationTicksReduction
		effects.ReconTicksReduction += upgrade.Effects.ReconTicksReduction
		effects.ResearchDiscountPercent += upgrade.Effects.ResearchDiscountPercent
	}
	return effects
}

func (e *Engine) fabricationTicks(s game.State) int64 {
	ticks := int64(3) - e.upgradeEffects(s).FabricationTicksReduction
	if ticks < 1 {
		return 1
	}
	return ticks
}

func (e *Engine) reconTicks(s game.State, base int64) int64 {
	ticks := base - e.upgradeEffects(s).ReconTicksReduction
	if ticks < 1 {
		return 1
	}
	return ticks
}

func (e *Engine) researchCost(s game.State, base int64) int64 {
	discount := e.upgradeEffects(s).ResearchDiscountPercent
	multiplier := int64(100) - discount
	if multiplier <= 0 {
		return 0
	}
	if base > math.MaxInt64/multiplier {
		return math.MaxInt64 / 100
	}
	return base * multiplier / 100
}

// ResearchCost exposes the catalog cost after completed upgrade discounts for
// player-facing panels without duplicating economy rules in the TUI.
func (e *Engine) ResearchCost(s game.State, base int64) int64 {
	return e.researchCost(s, base)
}

func (e *Engine) researchEffects(s game.State) content.ResearchEffects {
	var effects content.ResearchEffects
	for _, node := range e.Catalog.Research {
		if !contains(s.Research, node.ID) {
			continue
		}
		effects.ComputePerSecondBonus += node.Effects.ComputePerSecondBonus
		effects.AttackBonusPercent += node.Effects.AttackBonusPercent
		effects.StealthHeatReductionPercent += node.Effects.StealthHeatReductionPercent
		effects.AutomationIntervalReduction += node.Effects.AutomationIntervalReduction
	}
	return effects
}

func (e *Engine) automationInterval(s game.State) int64 {
	interval := int64(4) - e.researchEffects(s).AutomationIntervalReduction
	if interval < 1 {
		return 1
	}
	return interval
}
func generatorCount(s game.State, id string) int64 {
	for _, g := range s.Generators {
		if g.ID == id {
			return g.Count
		}
	}
	return 0
}
func upsertGenerator(s *game.State, id string) {
	for i := range s.Generators {
		if s.Generators[i].ID == id {
			s.Generators[i].Count++
			return
		}
	}
	s.Generators = append(s.Generators, game.Generator{ID: id, Count: 1})
}
func generatorCost(base, n int64) int64 {
	cost := base
	for i := int64(0); i < n; i++ {
		if cost > math.MaxInt64/6 {
			return math.MaxInt64
		}
		cost = (cost*6 + 4) / 5
	}
	return cost
}
