package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"github.com/jappoman/netrise/internal/save"
	"github.com/jappoman/netrise/internal/sim"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type tickMsg struct{}
type tutorialStep struct {
	Title  string
	Action string
	Done   bool
}

type Model struct {
	state           game.State
	engine          *sim.Engine
	width, height   int
	dataDir         string
	status          string
	saveCounter     int64
	screen          string
	selectedTarget  int
	selectedUpgrade int
	commitmentPct   int64
	doctrine        game.Doctrine
}

func New(s game.State, e *sim.Engine, dataDir string) Model {
	return Model{state: s, engine: e, dataDir: dataDir, screen: "overview", doctrine: game.Balanced, commitmentPct: 75}
}
func (m Model) Init() tea.Cmd { return tea.Tick(1e9, func(time.Time) tea.Msg { return tickMsg{} }) }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		if v.String() == "q" || v.String() == "ctrl+c" {
			m.saveNow()
			return m, tea.Quit
		}
		if m.screen == "help" {
			if v.String() == "esc" || v.String() == "o" || v.String() == "?" {
				m.screen = "overview"
			}
			return m, nil
		}
		if v.String() == "?" {
			m.screen = "help"
			return m, nil
		}
		if m.screen == "tutorial" {
			if v.String() == "esc" || v.String() == "o" || v.String() == "h" {
				m.screen = "overview"
			}
			return m, nil
		}
		if m.screen == "targets" {
			switch v.String() {
			case "up":
				if m.selectedTarget > 0 {
					m.selectedTarget--
				}
			case "down":
				if m.selectedTarget+1 < len(m.engine.Catalog.Targets) {
					m.selectedTarget++
				}
			case "enter":
				m.screen = "preflight"
			case "i":
				if len(m.engine.Catalog.Targets) > 0 {
					m.applyCommand(game.Command{Kind: game.StartRecon, ID: m.engine.Catalog.Targets[m.selectedTarget].ID})
				}
			case "esc", "o":
				m.screen = "overview"
			}
			return m, nil
		}
		if m.screen == "preflight" {
			switch v.String() {
			case "+", "right":
				m.commitmentPct = nextCommitment(m.commitmentPct)
			case "-", "left":
				m.commitmentPct = previousCommitment(m.commitmentPct)
			case "1":
				m.doctrine = game.Balanced
			case "2":
				m.doctrine = game.Stealth
			case "3":
				m.doctrine = game.Overload
			case "enter":
				if len(m.engine.Catalog.Targets) > 0 {
					target := m.engine.Catalog.Targets[m.selectedTarget]
					m.applyCommand(game.Command{Kind: game.StartBattle, ID: target.ID, Quantity: commitmentBots(m.state.Resources.Bots, m.commitmentPct), Doctrine: m.doctrine})
				}
			case "esc", "t":
				m.screen = "targets"
			}
			return m, nil
		}
		if m.screen == "research" {
			switch v.String() {
			case "enter":
				if id := m.nextResearchID(); id != "" {
					m.applyCommand(game.Command{Kind: game.BuyResearch, ID: id})
				}
			case "esc", "o":
				m.screen = "overview"
			}
			return m, nil
		}
		if m.screen == "upgrades" {
			switch v.String() {
			case "up":
				if m.selectedUpgrade > 0 {
					m.selectedUpgrade--
				}
			case "down":
				if m.selectedUpgrade+1 < len(m.engine.Catalog.Upgrades) {
					m.selectedUpgrade++
				}
			case "enter":
				if len(m.engine.Catalog.Upgrades) > 0 {
					m.applyCommand(game.Command{Kind: game.BuyUpgrade, ID: m.engine.Catalog.Upgrades[m.selectedUpgrade].ID})
				}
			case "esc", "o":
				m.screen = "overview"
			}
			return m, nil
		}
		if m.screen == "archive" {
			if v.String() == "esc" || v.String() == "o" {
				m.screen = "overview"
			}
			return m, nil
		}
		if m.screen == "prestige" {
			switch v.String() {
			case "enter":
				m.applyCommand(game.Command{Kind: game.IdentityBurn})
				if m.state.PrestigeCount > 0 {
					m.screen = "overview"
				}
			case "esc", "o":
				m.screen = "overview"
			}
			return m, nil
		}
		if m.screen == "restructure" {
			switch v.String() {
			case "enter":
				m.applyCommand(game.Command{Kind: game.CorpRestructure})
				if m.state.PrestigeCount > 0 {
					m.screen = "overview"
				}
			case "esc", "o":
				m.screen = "overview"
			}
			return m, nil
		}
		if v.String() == "t" {
			m.screen = "targets"
			return m, nil
		}
		if v.String() == "o" {
			m.screen = "overview"
			return m, nil
		}
		if v.String() == "y" {
			m.screen = "research"
			return m, nil
		}
		if v.String() == "g" && m.state.Faction == game.Runner {
			m.screen = "upgrades"
			return m, nil
		}
		if v.String() == "l" {
			m.screen = "archive"
			return m, nil
		}
		if v.String() == "h" {
			m.screen = "tutorial"
			return m, nil
		}
		if v.String() == "p" {
			if m.state.Faction == game.Corporation {
				m.screen = "restructure"
			} else {
				m.screen = "prestige"
			}
			return m, nil
		}
		if m.screen == "result" && v.String() == "enter" {
			m.screen = "overview"
			return m, nil
		}
		var command game.Command
		switch v.String() {
		case "c":
			command = game.Command{Kind: game.CollectContract}
		case "b":
			if m.state.Faction == game.Corporation {
				command = game.Command{Kind: game.CorpBuildAsset, ID: "transit_hosting"}
			} else {
				command = game.Command{Kind: game.BuyGenerator, ID: "salvaged_rig"}
			}
		case "s":
			command = game.Command{Kind: game.CorpBuildAsset, ID: "security_desk"}
		case "i":
			command = game.Command{Kind: game.CorpBuyICE, Quantity: 1}
		case "e":
			command = game.Command{Kind: game.CorpDeployICE, Quantity: 1}
		case "d":
			command = game.Command{Kind: game.CorpRunResearch}
		case "m":
			command = game.Command{Kind: game.CorpExpandMarket}
		case "M":
			command = game.Command{Kind: game.CorpEnableMaintenance}
		case "f":
			command = game.Command{Kind: game.QueueFabrication, Quantity: 1}
		case "u":
			command = game.Command{Kind: game.BuyUpgrade, ID: "secondhand_compiler"}
		case "a":
			command = game.Command{Kind: game.StartBattle, ID: "t_01", Quantity: attackCommitment(m.state.Resources.Bots), Doctrine: game.Balanced}
		case "r":
			command = game.Command{Kind: game.ReinforceBattle, Quantity: 1}
		case "w":
			command = game.Command{Kind: game.WithdrawBattle}
		case "C":
			command = game.Command{Kind: game.EnableAutomation, ID: "contracts"}
		case "F":
			command = game.Command{Kind: game.EnableAutomation, ID: "fabrication"}
		case "I":
			command = game.Command{Kind: game.EnableAutomation, ID: "recon", TargetID: "t_01"}
		}
		if command.Kind != "" && !(command.Kind == game.StartBattle && command.Quantity == 0) {
			m.applyCommand(command)
		}
	case tickMsg:
		var events []game.Event
		m.state, events, _ = m.engine.Apply(m.state, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if len(events) > 0 {
			m.status = events[len(events)-1].Message
		}
		if m.state.ActiveBattle != nil {
			m.screen = "battle"
		}
		if m.state.ActiveBattle == nil && m.state.LastBattle != nil && m.screen == "battle" {
			m.screen = "result"
		}
		m.saveCounter++
		if m.saveCounter%10 == 0 {
			m.saveNow()
		}
		return m, m.Init()
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
	}
	return m, nil
}
func (m Model) View() tea.View {
	if m.screen == "battle" {
		if m.compact() {
			return tea.NewView(m.compactBattleView())
		}
		return tea.NewView(m.battleView())
	}
	if m.compact() && m.screen == "overview" {
		return tea.NewView(m.compactView())
	}
	titleStyle := lipgloss.NewStyle()
	if os.Getenv("NO_COLOR") == "" {
		titleStyle = titleStyle.Bold(true).Foreground(lipgloss.Color("42"))
	}
	title := titleStyle.Render("NETRISE // RUNNER TERMINAL")
	if m.screen == "targets" {
		return tea.NewView(m.targetsView())
	}
	if m.screen == "preflight" {
		return tea.NewView(m.preflightView())
	}
	if m.screen == "result" {
		return tea.NewView(m.resultView())
	}
	if m.screen == "research" {
		return tea.NewView(m.researchView())
	}
	if m.screen == "upgrades" {
		return tea.NewView(m.upgradesView())
	}
	if m.screen == "archive" {
		return tea.NewView(m.archiveView())
	}
	if m.screen == "prestige" {
		return tea.NewView(m.prestigeView())
	}
	if m.screen == "restructure" {
		return tea.NewView(m.restructureView())
	}
	if m.screen == "help" {
		return tea.NewView(m.helpView())
	}
	if m.screen == "tutorial" {
		return tea.NewView(m.tutorialView())
	}
	if m.state.Faction == game.Corporation {
		return tea.NewView(m.corpOverviewView())
	}
	battle := "NO ACTIVE BATTLE"
	if m.state.ActiveBattle != nil {
		battle = fmt.Sprintf("BATTLE %s | LAYER %d | BOTS %d | INTEGRITY %d | ICE %d | TRACE %d%%", m.state.ActiveBattle.TargetID, m.state.ActiveBattle.LayerIndex+1, m.state.ActiveBattle.LiveBots, m.state.ActiveBattle.LayerIntegrity, m.state.ActiveBattle.DefenderICE, m.state.ActiveBattle.TraceBP/100)
	}
	controls := "[c] contract [b] rig [u] compiler [f] bot [t] targets [g] upgrades [y] research [h] guide [l] archive [p] prestige [q] quit"
	if m.state.ActiveBattle != nil {
		controls = "[r] reinforce once [w] withdraw [q] quit"
	}
	body := fmt.Sprintf("CREDITS %s\nCOMPUTE %s (+%d/s)\nBOTS %d/%d\nHEAT %d%%\nTICK %d\n%s\n%s\n%s\n\n%s\nNEXT: %s\n%s", m.state.Resources.Credits, m.state.Resources.Compute, m.engine.ComputeRate(m.state), m.state.Resources.Bots, m.state.Resources.BotCapacity, m.state.HeatBP/100, m.state.LogicalTick, battle, m.reconLine(), m.offlineLine(), controls, m.nextHint(), m.status)
	return tea.NewView(title + "\n\n" + body + "\n")
}

func (m Model) compact() bool { return m.width > 0 && (m.width < 80 || m.height > 0 && m.height < 24) }

func (m Model) compactView() string {
	if m.state.Faction == game.Corporation && m.state.Corp != nil {
		controls := "[b] hosting [s] security [i] ICE [d] R&D [m] market [M] upkeep [l] archive [p] restructure [q] quit"
		if m.state.ActiveBattle != nil {
			controls = "[e] deploy ICE [q] quit"
		}
		return fmt.Sprintf("NETRISE // CORPORATION (COMPACT)\nCAPITAL %s  ICE %d/%d\nREP %d%%  EXPOSURE %d%%\nTICK %d\n%s\nNEXT: %s\n%s", m.state.Corp.Capital, m.state.Corp.ICE, m.state.Corp.ICECapacity, m.state.Corp.ReputationBP/100, m.state.Corp.ExposureBP/100, m.state.LogicalTick, compactBattle(m.state), m.corpNextHint(), controls)
	}
	controls := "[c] contract [b] rig [t] targets [g] upgrades [y] research [h] guide [l] archive [p] prestige [q] quit"
	if m.state.ActiveBattle != nil {
		controls = "[r] reinforce [w] withdraw [q] quit"
	}
	return fmt.Sprintf("NETRISE // RUNNER (COMPACT)\nCREDITS %s  COMPUTE %s\nBOTS %d/%d  HEAT %d%%\nTICK %d\n%s\n%s\nNEXT: %s\n%s", m.state.Resources.Credits, m.state.Resources.Compute, m.state.Resources.Bots, m.state.Resources.BotCapacity, m.state.HeatBP/100, m.state.LogicalTick, compactBattle(m.state), m.reconLine(), m.nextHint(), controls)
}

func compactBattle(s game.State) string {
	if s.ActiveBattle == nil {
		return "NO ACTIVE BATTLE"
	}
	return fmt.Sprintf("BATTLE %s | bots %d | integrity %d | ICE %d", s.ActiveBattle.TargetID, s.ActiveBattle.LiveBots, s.ActiveBattle.LayerIntegrity, s.ActiveBattle.DefenderICE)
}

func (m Model) compactBattleView() string {
	if m.state.ActiveBattle == nil {
		return "NETRISE // BATTLE\n\nNo active battle.\n[o] overview"
	}
	b := m.state.ActiveBattle
	controls := "[r] reinforce  [w] withdraw  [q] quit"
	if m.state.Faction == game.Corporation {
		controls = "[e] deploy ICE  [q] quit"
	}
	return fmt.Sprintf("NETRISE // BATTLE\n%s\nLAYER %d | BOTS %d | INTEGRITY %d\nICE %d | TRACE %d%% | TICK %d/%d\n%s\nNEXT: %s", battleTargetName(m.engine, b.TargetID), b.LayerIndex+1, b.LiveBots, b.LayerIntegrity, b.DefenderICE, b.TraceBP/100, b.ElapsedTicks, b.TimeLimit, controls, m.battleNextHint())
}

func (m Model) battleView() string {
	if m.state.ActiveBattle == nil {
		return "NETRISE // BATTLE PLAYBACK\n\nNo active battle.\n[o] overview"
	}
	b := m.state.ActiveBattle
	target := battleTargetName(m.engine, b.TargetID)
	layers := 0
	for _, candidate := range m.engine.Catalog.Targets {
		if candidate.ID == b.TargetID {
			layers = len(candidate.Layers)
			break
		}
	}
	lines := "NETRISE // BATTLE PLAYBACK\n\n"
	lines += fmt.Sprintf("TARGET: %s\nDOCTRINE: %s\nLAYER: %d/%d\nINTEGRITY: %d\nLIVE BOTS: %d/%d\nDEFENDER ICE: %d\nTRACE: %d%%\nELAPSED: %d/%d ticks\nHEAT: %d%%\n\n", target, b.Doctrine, b.LayerIndex+1, layers, b.LayerIntegrity, b.LiveBots, b.CommittedBots, b.DefenderICE, b.TraceBP/100, b.ElapsedTicks, b.TimeLimit, m.state.HeatBP/100)
	lines += "RECENT EVENTS:\n"
	events := recentBattleEvents(m.state.EventLog)
	if len(events) == 0 {
		lines += "  Battle counters are resolving.\n"
	}
	for _, event := range events {
		lines += fmt.Sprintf("  [t+%d] %s\n", event.Tick, event.Message)
	}
	controls := "[r] reinforce once  [w] withdraw  [q] quit"
	if m.state.Faction == game.Corporation {
		controls = "[e] deploy reserve ICE  [q] quit"
	}
	return lines + "\n" + controls + "\nNEXT: " + m.battleNextHint()
}

func battleTargetName(engine *sim.Engine, id string) string {
	for _, target := range engine.Catalog.Targets {
		if target.ID == id {
			return target.Name
		}
	}
	return id
}

func recentBattleEvents(events []game.Event) []game.Event {
	filtered := make([]game.Event, 0, 6)
	for _, event := range events {
		switch event.Type {
		case "BattleStarted", "BattleReinforced", "LayerBreached", "BotsLost", "BattleEnded":
			filtered = append(filtered, event)
		}
	}
	if len(filtered) > 6 {
		filtered = filtered[len(filtered)-6:]
	}
	return filtered
}

func (m Model) battleNextHint() string {
	if m.state.Faction == game.Corporation {
		return "watch exposure and deploy reserve ICE only when the incursion needs it"
	}
	return "watch layer integrity; reinforce once if needed or withdraw to save surviving bots"
}

func (m Model) targetsView() string {
	lines := "NETRISE // TARGETS\n\nChoose a reachable asset, then press Enter for preflight.\n"
	for i, target := range m.engine.Catalog.Targets {
		marker := "  "
		if i == m.selectedTarget {
			marker = "> "
		}
		reach := "LOCKED"
		if targetAvailable(m.state, target) {
			reach = "REACHABLE"
		}
		intel := targetIntel(m.state, target.ID)
		recon := targetReconStatus(m.state, target)
		lines += fmt.Sprintf("%s%-36s %-10s layers:%d\n  %s | intel:%d%% (%s) | %s\n", marker, target.Name, reach, len(target.Layers), target.Corporation, intel, intelConfidence(intel), recon)
	}
	return lines + "\n[up/down] select  [i] recon  [enter] preflight  [esc] overview\nNEXT: unlock a route, then run recon to replace uncertainty with intel before launch."
}

func (m Model) preflightView() string {
	if len(m.engine.Catalog.Targets) == 0 {
		return "NETRISE // PREFLIGHT\n\nNo targets available.\n[esc] targets"
	}
	target := m.engine.Catalog.Targets[m.selectedTarget]
	intel := int64(0)
	if m.state.Intel != nil {
		intel = m.state.Intel[target.ID]
	}
	commitment := commitmentBots(m.state.Resources.Bots, m.commitmentPct)
	estimatedTicks, estimatedHeat := m.engine.EstimateBattle(m.state, target, m.doctrine, commitment)
	return fmt.Sprintf("NETRISE // PREFLIGHT\n\nTARGET: %s\nROUTE: %s\nINTEL: %d%% (%s)\nLAYERS: %d\nCOMMITMENT: %d bots (%d%%)\nDOCTRINE: %s\nBASELINE: ~%d ticks / max heat %d%%\n\n[+/- or left/right] commitment  [1] balanced  [2] stealth  [3] overload\n[enter] launch  [esc] targets\nNEXT: reserve bots for intervention; review intel confidence, commitment, doctrine and baseline.", target.Name, reachability(targetAvailable(m.state, target)), intel, intelConfidence(intel), len(target.Layers), commitment, m.commitmentPct, m.doctrine, estimatedTicks, estimatedHeat/100)
}

func (m Model) reconLine() string {
	if m.state.Recon == nil {
		return "NO ACTIVE RECON"
	}
	return fmt.Sprintf("RECON %s | %d ticks remaining", m.state.Recon.TargetID, m.state.Recon.RemainingTicks)
}

func (m Model) resultView() string {
	r := m.state.LastBattle
	if r == nil {
		return "NETRISE // RESULT\n\nNo completed battle.\n[o] overview"
	}
	return fmt.Sprintf("NETRISE // RESULT\n\nTARGET: %s\nOUTCOME: %s\nREASON: %s\nDOCTRINE: %s\nBOTS: %d committed / %d survived / %d lost\nTRACE: %d%%\nHEAT: +%d%% (total %d%%)\nLOOT: %d credits / %d data\nELAPSED: %d ticks\n\nNEXT: %s\n\n[enter] overview  [t] targets", battleTargetName(m.engine, r.TargetID), r.Outcome, battleOutcomeReason(r.Outcome), r.Doctrine, r.CommittedBots, r.Survivors, r.BotsLost, r.TraceBP/100, r.HeatAddedBP/100, r.HeatAfterBP/100, r.CreditsLooted, r.DataLooted, r.ElapsedTicks, resultNextHint(r.Outcome))
}

func battleOutcomeReason(outcome string) string {
	switch outcome {
	case "victory":
		return "all security layers breached and loot extracted"
	case "defeat":
		return "countermeasures eliminated the committed bots"
	case "withdrawn":
		return "the Runner withdrew before final resolution"
	case "timeout":
		return "the time limit expired; surviving bots returned"
	case "defended":
		return "the Corporation kept the asset secured"
	case "breach":
		return "the Corporation asset was exposed"
	default:
		return "battle resolution completed"
	}
}

func resultNextHint(outcome string) string {
	if outcome == "victory" {
		return "review the archive, then inspect the next reachable target"
	}
	return "review the result, recover resources and prepare the next attempt"
}

func (m Model) researchView() string {
	lines := "NETRISE // RESEARCH\n\n"
	for _, node := range m.engine.Catalog.Research {
		state := "LOCKED"
		cost := m.engine.ResearchCost(m.state, node.CostData)
		if hasResearch(m.state, node.ID) {
			state = "COMPLETE"
		} else if researchPrerequisitesMet(m.state, node.Prerequisites) {
			state = fmt.Sprintf("AVAILABLE (%d data)", cost)
		}
		description := node.Description
		if effects := researchEffectText(node.Effects); effects != "" {
			description += " // " + effects
		}
		lines += fmt.Sprintf("%-22s %-18s %s\n  %s\n", node.Name, node.ID, state, description)
	}
	return lines + "\n[enter] buy next available node  [esc] overview"
}

func (m Model) upgradesView() string {
	lines := "NETRISE // UPGRADES\n\nSelect an upgrade, review its effect and press Enter to purchase.\n"
	for i, upgrade := range m.engine.Catalog.Upgrades {
		marker := "  "
		if i == m.selectedUpgrade {
			marker = "> "
		}
		state := fmt.Sprintf("AVAILABLE (%d CR)", upgrade.Cost)
		if hasUpgrade(m.state, upgrade.ID) {
			state = "INSTALLED"
		} else if m.state.Resources.Credits < mathx.New(upgrade.Cost) {
			state = fmt.Sprintf("NEED %d CR", upgrade.Cost)
		}
		description := upgrade.Description
		if effects := upgradeEffectText(upgrade.Effects); effects != "" {
			description += " // " + effects
		}
		lines += fmt.Sprintf("%s%-22s %-18s\n  %s\n", marker, upgrade.Name, state, description)
	}
	return lines + "\n[up/down] select  [enter] purchase  [esc] overview\nNEXT: install only upgrades you can afford; installed upgrades cannot be repurchased."
}

func researchEffectText(effects content.ResearchEffects) string {
	var parts []string
	if effects.ComputePerSecondBonus > 0 {
		parts = append(parts, fmt.Sprintf("+%d compute/s", effects.ComputePerSecondBonus))
	}
	if effects.AttackBonusPercent > 0 {
		parts = append(parts, fmt.Sprintf("+%d%% breach power", effects.AttackBonusPercent))
	}
	if effects.StealthHeatReductionPercent > 0 {
		parts = append(parts, fmt.Sprintf("-%d%% stealth heat", effects.StealthHeatReductionPercent))
	}
	if effects.AutomationIntervalReduction > 0 {
		parts = append(parts, fmt.Sprintf("automation every %ds faster", effects.AutomationIntervalReduction))
	}
	return strings.Join(parts, ", ")
}

func upgradeEffectText(effects content.UpgradeEffects) string {
	var parts []string
	if effects.ComputePerSecondBonus > 0 {
		parts = append(parts, fmt.Sprintf("+%d compute/s", effects.ComputePerSecondBonus))
	}
	if effects.HeatReductionPercent > 0 {
		parts = append(parts, fmt.Sprintf("-%d%% battle heat", effects.HeatReductionPercent))
	}
	if effects.StealthAttackBonusPercent > 0 {
		parts = append(parts, fmt.Sprintf("+%d%% stealth power", effects.StealthAttackBonusPercent))
	}
	if effects.BotCapacityBonus > 0 {
		parts = append(parts, fmt.Sprintf("+%d bot capacity", effects.BotCapacityBonus))
	}
	if effects.FabricationTicksReduction > 0 {
		parts = append(parts, fmt.Sprintf("-%ds fabrication", effects.FabricationTicksReduction))
	}
	if effects.ReconTicksReduction > 0 {
		parts = append(parts, fmt.Sprintf("-%ds recon", effects.ReconTicksReduction))
	}
	if effects.ResearchDiscountPercent > 0 {
		parts = append(parts, fmt.Sprintf("-%d%% research data", effects.ResearchDiscountPercent))
	}
	return strings.Join(parts, ", ")
}

func (m Model) archiveView() string {
	lines := "NETRISE // EVENT ARCHIVE\n\n"
	start := 0
	if len(m.state.EventLog) > 12 {
		start = len(m.state.EventLog) - 12
	}
	for _, event := range m.state.EventLog[start:] {
		lines += fmt.Sprintf("[t+%d] %s: %s\n", event.Tick, event.Type, event.Message)
	}
	if len(m.state.EventLog) == 0 {
		lines += "No events recorded yet.\n"
	}
	return lines + "\n[esc] overview"
}

func (m Model) helpView() string {
	if m.state.Faction == game.Corporation {
		return "NETRISE // HELP\n\nOBJECTIVE: grow a resilient information business.\n[b] build hosting  [s] security desk  [i] buy ICE\n[d] R&D  [m] expand market  [M] enable maintenance\n[e] deploy reserve ICE during an incursion\n[h] progression guide  [l] event archive  [p] restructuring  [q] save and quit\n\nRevenue and incursions advance only during active play.\n[esc] return"
	}
	return "NETRISE // HELP\n\nOBJECTIVE: prepare, breach, extract and automate.\n[c] contract  [b] rig  [u] compiler  [f] fabricate\n[t] targets  [g] upgrades  [i] recon  [y] research  [h] guide  [l] archive\n[p] Identity Burn  [r] reinforce  [w] withdraw\n[q] save and quit\n\nBattles are deterministic and resolve on 1-second ticks.\n[esc] return"
}

func (m Model) tutorialView() string {
	steps := m.tutorialSteps()
	current := len(steps)
	for i, step := range steps {
		if !step.Done {
			current = i
			break
		}
	}
	lines := "NETRISE // PROGRESSION GUIDE\n\n"
	if current == len(steps) {
		lines += "LOOP ESTABLISHED: keep improving the campaign or open the prestige screen.\n\n"
	}
	for i, step := range steps {
		marker := "[ ]"
		state := "LOCKED"
		if step.Done {
			marker, state = "[x]", "DONE"
		} else if i == current {
			marker, state = ">", "NEXT"
		}
		lines += fmt.Sprintf("%s %d. %-28s %-6s\n    %s\n", marker, i+1, step.Title, state, step.Action)
	}
	return lines + "\n[h/esc] overview  [?] help\nNEXT: complete the first NEXT item; the guide updates from the live campaign state."
}

func (m Model) tutorialSteps() []tutorialStep {
	if m.state.Faction == game.Corporation {
		return []tutorialStep{
			{Title: "Build a second client asset", Action: "press [b] to add Transit Hosting", Done: m.state.Corp != nil && len(m.state.Corp.Assets) >= 2},
			{Title: "Invest in security", Action: "press [s] to install a Security Desk", Done: m.corpHasAsset("security_desk")},
			{Title: "Fill the reserve ICE rack", Action: "press [i] until ICE reaches capacity", Done: m.state.Corp != nil && m.state.Corp.ICE >= m.state.Corp.ICECapacity},
			{Title: "Survive an active-play incursion", Action: "advance active play and review the battle playback", Done: m.state.LastBattle != nil},
			{Title: "Enable automated maintenance", Action: "press [M] after investing in R&D", Done: m.state.Corp != nil && m.state.Corp.MaintenanceAuto},
		}
	}
	return []tutorialStep{
		{Title: "Install the bot compiler", Action: "collect contracts with [c], then buy it in [g] upgrades", Done: hasUpgrade(m.state, "secondhand_compiler")},
		{Title: "Bring compute online", Action: "press [b] to buy a Salvaged Rig", Done: hasGenerator(m.state, "salvaged_rig")},
		{Title: "Fabricate your first bot", Action: "press [f] when compute and credits are available", Done: m.state.Resources.Bots > 0},
		{Title: "Replace uncertainty with intel", Action: "open [t], select t_01 and press [i] for recon", Done: targetIntel(m.state, "t_01") >= 45},
		{Title: "Complete the first breach", Action: "open [t], review preflight and press [enter] to launch", Done: containsID(m.state.CompletedTargets, "t_01")},
		{Title: "Fund the next capability", Action: "open [y] research and buy an available node", Done: len(m.state.Research) > 0},
	}
}

func (m Model) corpHasAsset(id string) bool {
	if m.state.Corp == nil {
		return false
	}
	for _, asset := range m.state.Corp.Assets {
		if asset.ID == id {
			return true
		}
	}
	return false
}

func containsID(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func (m Model) prestigeView() string {
	reward := m.engine.IdentityBurnReward(m.state)
	return fmt.Sprintf("NETRISE // IDENTITY BURN\n\nTHIS RUN DATA: %d\nGHOST KEYS GAINED: %d\nCURRENT GHOST KEYS: %d\n\nLOST: credits, bots, rigs, intel, heat, research and routes\nKEPT: Ghost Keys and lifetime prestige count\n\n[enter] confirm reset  [esc] cancel\n%s", m.state.RunExtractedData, reward, m.state.GhostKeys, m.status)
}

func (m Model) corpOverviewView() string {
	corp := m.state.Corp
	if corp == nil {
		return "NETRISE // CORPORATION\n\nCampaign state unavailable.\n[q] quit"
	}
	battle := "NO ACTIVE INCURSION"
	if m.state.ActiveBattle != nil {
		battle = fmt.Sprintf("INCursion %s | BOTS %d | INTEGRITY %d | ICE %d | TRACE %d%%", m.state.ActiveBattle.TargetID, m.state.ActiveBattle.LiveBots, m.state.ActiveBattle.LayerIntegrity, m.state.ActiveBattle.DefenderICE, m.state.ActiveBattle.TraceBP/100)
	}
	controls := "[b] hosting [s] security [i] ICE [d] R&D [m] market [M] upkeep [l] archive [p] restructure [q] quit"
	if m.state.ActiveBattle != nil {
		controls = "[e] deploy reserve ICE [q] quit"
	}
	return fmt.Sprintf("NETRISE // CORPORATION\n\nCAPITAL %s (+%d/tick)\nCOMPUTE %s\nICE %d/%d\nREPUTATION %d%%\nEXPOSURE %d%%\nASSETS %d\nTICK %d\n%s\n%s\n\n%s\nNEXT: %s\n%s", corp.Capital, corpCapitalRate(corp), corp.Compute, corp.ICE, corp.ICECapacity, corp.ReputationBP/100, corp.ExposureBP/100, len(corp.Assets), m.state.LogicalTick, battle, m.offlineLine(), controls, m.corpNextHint(), m.status)
}

func (m Model) offlineLine() string {
	if m.state.OfflineSummary == nil || m.state.OfflineSummary.Ticks == 0 {
		return ""
	}
	summary := m.state.OfflineSummary
	if m.state.Faction == game.Corporation {
		return fmt.Sprintf("OFFLINE: %d safe ticks / +%s capital", summary.Ticks, summary.CapitalProduced)
	}
	return fmt.Sprintf("OFFLINE: %d safe ticks / +%s compute", summary.Ticks, summary.ComputeProduced)
}

func (m Model) restructureView() string {
	reward := int64(0)
	if m.state.Corp != nil {
		reward = integerSqrtPreview(m.state.Corp.RunValuation / 1000)
	}
	patents := int64(0)
	if m.state.Corp != nil {
		patents = m.state.Corp.PatentShares
	}
	return fmt.Sprintf("NETRISE // RESTRUCTURING\n\nCURRENT VALUATION: %d\nPATENT SHARES GAINED: %d\nCURRENT PATENT SHARES: %d\n\nLOST: capital, current assets, ICE, demand and exposure\nKEPT: Patent Shares and Corporation campaign identity\n\n[enter] confirm reset  [esc] cancel\n%s", corpValuation(m.state), reward, patents, m.status)
}

func corpValuation(s game.State) int64 {
	if s.Corp == nil {
		return 0
	}
	return s.Corp.RunValuation
}
func integerSqrtPreview(value int64) int64 {
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

func corpCapitalRate(corp *game.CorpState) int64 {
	rate := int64(2)
	for _, asset := range corp.Assets {
		if asset.Operational {
			rate += asset.CapitalPerTick
		}
	}
	return rate
}

func (m Model) nextResearchID() string {
	for _, node := range m.engine.Catalog.Research {
		if !hasResearch(m.state, node.ID) && researchPrerequisitesMet(m.state, node.Prerequisites) && m.state.Resources.Data >= mathx.New(node.CostData) {
			return node.ID
		}
	}
	return ""
}

func hasResearch(s game.State, id string) bool { return containsString(s.Research, id) }

func researchPrerequisitesMet(s game.State, prerequisites []string) bool {
	return len(prerequisites) == 0 || allContained(s.Research, prerequisites)
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func allContained(values, wanted []string) bool {
	for _, id := range wanted {
		if !containsString(values, id) {
			return false
		}
	}
	return true
}

func reachability(reachable bool) string {
	if reachable {
		return "reachable"
	}
	return "locked"
}

func targetAvailable(s game.State, target content.Target) bool {
	if target.Reachable {
		return true
	}
	for _, id := range s.CompletedTargets {
		if id == target.RequiresTargetID {
			return true
		}
	}
	return false
}

func targetIntel(s game.State, id string) int64 {
	if s.Intel == nil {
		return 0
	}
	return s.Intel[id]
}

func intelConfidence(intel int64) string {
	switch {
	case intel >= 100:
		return "full"
	case intel >= 50:
		return "strong"
	case intel > 0:
		return "partial"
	default:
		return "unknown"
	}
}

func targetReconStatus(s game.State, target content.Target) string {
	intel := targetIntel(s, target.ID)
	if !targetAvailable(s, target) {
		if target.RequiresTargetID != "" {
			return "requires " + target.RequiresTargetID
		}
		return "route locked"
	}
	if intel >= 100 {
		return "intel complete"
	}
	if s.Recon != nil {
		if s.Recon.TargetID == target.ID {
			return fmt.Sprintf("recon active: %d ticks", s.Recon.RemainingTicks)
		}
		return "recon occupied"
	}
	return fmt.Sprintf("recon ready: %d CR / %d ticks [i]", target.ReconCost, target.ReconTicks)
}

func (m *Model) applyCommand(command game.Command) {
	var events []game.Event
	var err error
	m.state, events, err = m.engine.Apply(m.state, command)
	if err != nil {
		m.status = err.Error()
	} else if len(events) > 0 {
		m.status = events[len(events)-1].Message
	}
	if m.state.ActiveBattle != nil {
		m.screen = "battle"
	}
	m.saveNow()
}

func (m Model) nextHint() string {
	if m.state.ActiveBattle != nil {
		return "watch the counters; press [w] to withdraw"
	}
	if m.state.Recon != nil {
		return "recon is running; wait for intel before opening preflight"
	}
	if !hasUpgrade(m.state, "secondhand_compiler") {
		return "collect credits, then press [u] to install the compiler (75 CR)"
	}
	if !hasGenerator(m.state, "salvaged_rig") {
		return "press [b] to buy a Salvaged Rig (50 CR)"
	}
	if m.state.Resources.Bots == 0 {
		return "press [f] when you have 10 compute and 2 credits"
	}
	return "press [a] to launch the tutorial breach with all bots"
}

func (m Model) corpNextHint() string {
	if m.state.ActiveBattle != nil {
		return "review the intrusion and deploy reserve ICE only when needed"
	}
	if m.state.Corp == nil {
		return "campaign state unavailable"
	}
	if len(m.state.Corp.Assets) < 2 {
		return "build Transit Hosting, then invest in a Security Desk"
	}
	if m.state.Corp.ICE < m.state.Corp.ICECapacity {
		return "restock reserve ICE before the next emergency audit"
	}
	return "maintain capital and watch the next active-play intrusion"
}

func hasUpgrade(s game.State, id string) bool {
	for _, upgrade := range s.Upgrades {
		if upgrade == id {
			return true
		}
	}
	return false
}

func hasGenerator(s game.State, id string) bool {
	for _, generator := range s.Generators {
		if generator.ID == id && generator.Count > 0 {
			return true
		}
	}
	return false
}

func attackCommitment(bots int64) int64 {
	return commitmentBots(bots, 75)
}

func commitmentBots(bots, percent int64) int64 {
	if bots <= 0 {
		return 0
	}
	if percent < 25 {
		percent = 25
	}
	if percent > 100 {
		percent = 100
	}
	committed := bots * percent / 100
	if committed < 1 {
		return 1
	}
	return committed
}

func nextCommitment(percent int64) int64 {
	for _, option := range []int64{25, 50, 75, 100} {
		if option > percent {
			return option
		}
	}
	return 100
}

func previousCommitment(percent int64) int64 {
	for i := len([]int64{25, 50, 75, 100}) - 1; i >= 0; i-- {
		option := []int64{25, 50, 75, 100}[i]
		if option < percent {
			return option
		}
	}
	return 25
}

func (m *Model) saveNow() {
	if m.dataDir == "" {
		return
	}
	m.state.SavedAtUTC = time.Now().UTC().Format(time.RFC3339Nano)
	if err := save.Write(m.dataDir, m.state); err != nil {
		m.status = "save error: " + err.Error()
	}
}

// SavePath is kept here so callers can use the same path policy in diagnostics.
func SavePath(dir string) string { return filepath.Join(dir, "runner.json") }
