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
	"time"
)

type tickMsg struct{}
type Model struct {
	state          game.State
	engine         *sim.Engine
	width, height  int
	dataDir        string
	status         string
	saveCounter    int64
	screen         string
	selectedTarget int
	doctrine       game.Doctrine
}

func New(s game.State, e *sim.Engine, dataDir string) Model {
	return Model{state: s, engine: e, dataDir: dataDir, screen: "overview", doctrine: game.Balanced}
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
			case "1":
				m.doctrine = game.Balanced
			case "2":
				m.doctrine = game.Stealth
			case "3":
				m.doctrine = game.Overload
			case "enter":
				if len(m.engine.Catalog.Targets) > 0 {
					target := m.engine.Catalog.Targets[m.selectedTarget]
					m.applyCommand(game.Command{Kind: game.StartBattle, ID: target.ID, Quantity: attackCommitment(m.state.Resources.Bots), Doctrine: m.doctrine})
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
		if v.String() == "l" {
			m.screen = "archive"
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
	if m.compact() {
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
	if m.state.Faction == game.Corporation {
		return tea.NewView(m.corpOverviewView())
	}
	battle := "NO ACTIVE BATTLE"
	if m.state.ActiveBattle != nil {
		battle = fmt.Sprintf("BATTLE %s | LAYER %d | BOTS %d | INTEGRITY %d | ICE %d | TRACE %d%%", m.state.ActiveBattle.TargetID, m.state.ActiveBattle.LayerIndex+1, m.state.ActiveBattle.LiveBots, m.state.ActiveBattle.LayerIntegrity, m.state.ActiveBattle.DefenderICE, m.state.ActiveBattle.TraceBP/100)
	}
	controls := "[c] contract [b] rig [u] compiler [f] bot [t] targets [y] research [l] archive [p] prestige [q] quit"
	if m.state.ActiveBattle != nil {
		controls = "[r] reinforce once [w] withdraw [q] quit"
	}
	body := fmt.Sprintf("CREDITS %s\nCOMPUTE %s (+%d/s)\nBOTS %d/%d\nHEAT %d%%\nTICK %d\n%s\n%s\n\n%s\nNEXT: %s\n%s", m.state.Resources.Credits, m.state.Resources.Compute, m.engine.ComputeRate(m.state), m.state.Resources.Bots, m.state.Resources.BotCapacity, m.state.HeatBP/100, m.state.LogicalTick, battle, m.offlineLine(), controls, m.nextHint(), m.status)
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
	controls := "[c] contract [b] rig [t] targets [y] research [l] archive [p] prestige [q] quit"
	if m.state.ActiveBattle != nil {
		controls = "[r] reinforce [w] withdraw [q] quit"
	}
	return fmt.Sprintf("NETRISE // RUNNER (COMPACT)\nCREDITS %s  COMPUTE %s\nBOTS %d/%d  HEAT %d%%\nTICK %d\n%s\nNEXT: %s\n%s", m.state.Resources.Credits, m.state.Resources.Compute, m.state.Resources.Bots, m.state.Resources.BotCapacity, m.state.HeatBP/100, m.state.LogicalTick, compactBattle(m.state), m.nextHint(), controls)
}

func compactBattle(s game.State) string {
	if s.ActiveBattle == nil {
		return "NO ACTIVE BATTLE"
	}
	return fmt.Sprintf("BATTLE %s | bots %d | integrity %d | ICE %d", s.ActiveBattle.TargetID, s.ActiveBattle.LiveBots, s.ActiveBattle.LayerIntegrity, s.ActiveBattle.DefenderICE)
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
		lines += fmt.Sprintf("%s%-36s %-10s layers:%d\n", marker, target.Name, reach, len(target.Layers))
	}
	return lines + "\n[up/down] select  [i] recon  [enter] preflight  [esc] overview"
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
	return fmt.Sprintf("NETRISE // PREFLIGHT\n\nTARGET: %s\nROUTE: %s\nINTEL: %d%%\nLAYERS: %d\nCOMMITMENT: %d bots (75%% default)\nDOCTRINE: %s\n\n[1] balanced  [2] stealth  [3] overload\n[enter] launch  [esc] targets\nNEXT: ensure the commitment is affordable and review the doctrine.", target.Name, reachability(targetAvailable(m.state, target)), intel, len(target.Layers), attackCommitment(m.state.Resources.Bots), m.doctrine)
}

func (m Model) resultView() string {
	r := m.state.LastBattle
	if r == nil {
		return "NETRISE // RESULT\n\nNo completed battle.\n[o] overview"
	}
	return fmt.Sprintf("NETRISE // RESULT\n\nTARGET: %s\nOUTCOME: %s\nSURVIVORS: %d\nBOTS LOST: %d\nLOOT: %d credits / %d data\nELAPSED: %d ticks\n\n[enter] overview  [t] targets", r.TargetID, r.Outcome, r.Survivors, r.BotsLost, r.CreditsLooted, r.DataLooted, r.ElapsedTicks)
}

func (m Model) researchView() string {
	lines := "NETRISE // RESEARCH\n\n"
	for _, node := range m.engine.Catalog.Research {
		state := "LOCKED"
		if hasResearch(m.state, node.ID) {
			state = "COMPLETE"
		} else if researchPrerequisitesMet(m.state, node.Prerequisites) {
			state = fmt.Sprintf("AVAILABLE (%d data)", node.CostData)
		}
		lines += fmt.Sprintf("%-22s %-18s %s\n", node.Name, node.ID, state)
	}
	return lines + "\n[enter] buy next available node  [esc] overview"
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
		return "NETRISE // HELP\n\nOBJECTIVE: grow a resilient information business.\n[b] build hosting  [s] security desk  [i] buy ICE\n[d] R&D  [m] expand market  [M] enable maintenance\n[e] deploy reserve ICE during an incursion\n[l] event archive  [p] restructuring  [q] save and quit\n\nRevenue and incursions advance only during active play.\n[esc] return"
	}
	return "NETRISE // HELP\n\nOBJECTIVE: prepare, breach, extract and automate.\n[c] contract  [b] rig  [u] compiler  [f] fabricate\n[t] targets  [i] recon  [y] research  [l] archive\n[p] Identity Burn  [r] reinforce  [w] withdraw\n[q] save and quit\n\nBattles are deterministic and resolve on 1-second ticks.\n[esc] return"
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
	if bots <= 1 {
		return bots
	}
	return (bots * 3) / 4
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
