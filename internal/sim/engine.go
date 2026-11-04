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
	return e.Advance(s, elapsedTicks), elapsedTicks
}

func New(c content.Catalog, seed int64) *Engine {
	return &Engine{Catalog: c, rng: uint64(seed) + 0x9e3779b97f4a7c15}
}
func (e *Engine) next() uint64 { e.rng ^= e.rng << 7; e.rng ^= e.rng >> 9; return e.rng }
func NewState(seed int64) game.State {
	return game.State{SchemaVersion: game.CurrentSchema, GameVersion: "0.1.0", CampaignID: "runner-01", Faction: game.Runner, WorldSeed: seed, RNGState: uint64(seed) + 0x9e3779b97f4a7c15, Resources: game.Resources{Credits: mathx.New(25), Compute: mathx.New(2), BotCapacity: 20}}
}

func (e *Engine) Apply(s game.State, c game.Command) (game.State, []game.Event, error) {
	if c.Kind == game.AdvanceTicks {
		if c.Ticks < 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "ticks cannot be negative"}
		}
		advanced, events := e.advanceEvents(s, c.Ticks)
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
		s.Fabrication = &game.Fabrication{RemainingTicks: 3, Quantity: c.Quantity}
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
		if !target.Reachable {
			return s, nil, &game.CommandError{Reason: game.ReasonTargetUnreachable, Detail: "target route is not reachable"}
		}
		if c.Doctrine != game.Balanced && c.Doctrine != game.Stealth && c.Doctrine != game.Overload {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown doctrine"}
		}
		s.Resources.Bots -= c.Quantity
		s.ActiveBattle = &game.Battle{TargetID: c.ID, Doctrine: c.Doctrine, CommittedBots: c.Quantity, LiveBots: c.Quantity, LayerIntegrity: target.Layers[0].Integrity, DefenderICE: target.Layers[0].ICE, TimeLimit: target.TimeLimit}
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
		s.ActiveBattle = nil
		return s, []game.Event{{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Withdrawal complete. Surviving bots returned."}}, nil
	default:
		return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unsupported command"}
	}
}
func (e *Engine) Advance(s game.State, ticks int64) game.State {
	advanced, _ := e.advanceEvents(s, ticks)
	return advanced
}

func (e *Engine) advanceEvents(s game.State, ticks int64) (game.State, []game.Event) {
	var events []game.Event
	for i := int64(0); i < ticks; i++ {
		s.LogicalTick++
		s.RNGState ^= s.RNGState << 7
		s.RNGState ^= s.RNGState >> 9
		rate := int64(2)
		for _, g := range e.Catalog.Generators {
			for _, owned := range s.Generators {
				if owned.ID == g.ID {
					rate += owned.Count * g.ComputePerSecond
				}
			}
		}
		produced := mathx.New(rate)
		s.Resources.Compute, _ = s.Resources.Compute.Add(produced)
		if s.Fabrication != nil {
			s.Fabrication.RemainingTicks--
			if s.Fabrication.RemainingTicks <= 0 {
				s.Resources.Bots += s.Fabrication.Quantity
				s.Fabrication = nil
			}
		}
		if s.ActiveBattle != nil {
			s, events = e.advanceBattle(s, events)
		}
	}
	return s, events
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
	switch b.Doctrine {
	case game.Stealth:
		attackMultiplier, heatMultiplier = 82, 55
	case game.Overload:
		attackMultiplier, heatMultiplier = 125, 170
	}
	damage := b.LiveBots * attackMultiplier / (100 + b.DefenderICE)
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
	if layer.Counter > 0 && b.ElapsedTicks%3 == 0 && b.LiveBots > 0 {
		b.LiveBots--
		events = append(events, game.Event{Tick: s.LogicalTick, Type: "BotsLost", Message: "Countermeasure destroyed one bot."})
	}
	if b.LiveBots <= 0 {
		events = append(events, game.Event{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Breach failed. The committed bots were lost."})
		s.ActiveBattle = nil
		return s, events
	}
	if b.LayerIntegrity <= 0 {
		events = append(events, game.Event{Tick: s.LogicalTick, Type: "LayerBreached", Message: fmt.Sprintf("Layer %d breached.", b.LayerIndex+1)})
		b.LayerIndex++
		if b.LayerIndex >= len(target.Layers) {
			s.Resources.Bots += b.LiveBots
			s.Resources.Credits, _ = s.Resources.Credits.Add(mathx.New(target.LootCredits))
			s.Resources.Data, _ = s.Resources.Data.Add(mathx.New(target.LootData))
			s.ActiveBattle = nil
			events = append(events, game.Event{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Breach successful. Loot extracted."})
			return s, events
		}
		b.LayerIntegrity = target.Layers[b.LayerIndex].Integrity
		b.DefenderICE = target.Layers[b.LayerIndex].ICE
	}
	if b.ElapsedTicks >= b.TimeLimit {
		s.Resources.Bots += b.LiveBots
		s.ActiveBattle = nil
		events = append(events, game.Event{Tick: s.LogicalTick, Type: "BattleEnded", Message: "Breach timed out. Surviving bots returned."})
		return s, events
	}
	s.ActiveBattle = b
	return s, events
}

func (e *Engine) target(id string) (content.Target, bool) {
	for _, target := range e.Catalog.Targets {
		if target.ID == id {
			return target, true
		}
	}
	return content.Target{}, false
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
	return rate
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
