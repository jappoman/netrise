package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/save"
	"github.com/jappoman/netrise/internal/sim"
	"path/filepath"
	"time"
)

type tickMsg struct{}
type Model struct {
	state         game.State
	engine        *sim.Engine
	width, height int
	dataDir       string
	status        string
	saveCounter   int64
	screen        string
	selectedTarget int
	doctrine      game.Doctrine
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
		if m.screen == "targets" {
			switch v.String() {
			case "up":
				if m.selectedTarget > 0 { m.selectedTarget-- }
			case "down":
				if m.selectedTarget+1 < len(m.engine.Catalog.Targets) { m.selectedTarget++ }
			case "enter":
				m.screen = "preflight"
			case "esc", "o":
				m.screen = "overview"
			}
			return m, nil
		}
		if m.screen == "preflight" {
			switch v.String() {
			case "1": m.doctrine = game.Balanced
			case "2": m.doctrine = game.Stealth
			case "3": m.doctrine = game.Overload
			case "enter":
				if len(m.engine.Catalog.Targets) > 0 {
					target := m.engine.Catalog.Targets[m.selectedTarget]
					m.applyCommand(game.Command{Kind: game.StartBattle, ID: target.ID, Quantity: attackCommitment(m.state.Resources.Bots), Doctrine: m.doctrine})
				}
			case "esc", "t": m.screen = "targets"
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
		if m.screen == "result" && v.String() == "enter" {
			m.screen = "overview"
			return m, nil
		}
		var command game.Command
		switch v.String() {
		case "c":
			command = game.Command{Kind: game.CollectContract}
		case "b":
			command = game.Command{Kind: game.BuyGenerator, ID: "salvaged_rig"}
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
		}
		if command.Kind != "" && !(command.Kind == game.StartBattle && command.Quantity == 0) { m.applyCommand(command) }
	case tickMsg:
		var events []game.Event
		m.state, events, _ = m.engine.Apply(m.state, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if len(events) > 0 {
			m.status = events[len(events)-1].Message
		}
		if m.state.ActiveBattle != nil { m.screen = "battle" }
		if m.state.ActiveBattle == nil && m.state.LastBattle != nil && m.screen == "battle" { m.screen = "result" }
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
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("NETRISE // RUNNER TERMINAL")
	if m.screen == "targets" { return tea.NewView(m.targetsView()) }
	if m.screen == "preflight" { return tea.NewView(m.preflightView()) }
	if m.screen == "result" { return tea.NewView(m.resultView()) }
	battle := "NO ACTIVE BATTLE"
	if m.state.ActiveBattle != nil {
		battle = fmt.Sprintf("BATTLE %s | LAYER %d | BOTS %d | INTEGRITY %d | ICE %d | TRACE %d%%", m.state.ActiveBattle.TargetID, m.state.ActiveBattle.LayerIndex+1, m.state.ActiveBattle.LiveBots, m.state.ActiveBattle.LayerIntegrity, m.state.ActiveBattle.DefenderICE, m.state.ActiveBattle.TraceBP/100)
	}
	controls := "[c] contract [b] rig [u] compiler [f] bot [t] targets [q] quit"
	if m.state.ActiveBattle != nil {
		controls = "[r] reinforce once [w] withdraw [q] quit"
	}
	body := fmt.Sprintf("CREDITS %s\nCOMPUTE %s (+%d/s)\nBOTS %d/%d\nHEAT %d%%\nTICK %d\n%s\n\n%s\nNEXT: %s\n%s", m.state.Resources.Credits, m.state.Resources.Compute, m.engine.ComputeRate(m.state), m.state.Resources.Bots, m.state.Resources.BotCapacity, m.state.HeatBP/100, m.state.LogicalTick, battle, controls, m.nextHint(), m.status)
	return tea.NewView(title + "\n\n" + body + "\n")
}

func (m Model) targetsView() string {
	lines := "NETRISE // TARGETS\n\nChoose a reachable asset, then press Enter for preflight.\n"
	for i, target := range m.engine.Catalog.Targets {
		marker := "  "
		if i == m.selectedTarget { marker = "> " }
		reach := "LOCKED"
		if target.Reachable { reach = "REACHABLE" }
		lines += fmt.Sprintf("%s%-36s %-10s layers:%d\n", marker, target.Name, reach, len(target.Layers))
	}
	return lines + "\n[up/down] select  [enter] preflight  [esc] overview"
}

func (m Model) preflightView() string {
	if len(m.engine.Catalog.Targets) == 0 { return "NETRISE // PREFLIGHT\n\nNo targets available.\n[esc] targets" }
	target := m.engine.Catalog.Targets[m.selectedTarget]
	return fmt.Sprintf("NETRISE // PREFLIGHT\n\nTARGET: %s\nROUTE: %s\nLAYERS: %d\nCOMMITMENT: %d bots (75%% default)\nDOCTRINE: %s\n\n[1] balanced  [2] stealth  [3] overload\n[enter] launch  [esc] targets\nNEXT: ensure the commitment is affordable and review the doctrine.", target.Name, reachability(target.Reachable), len(target.Layers), attackCommitment(m.state.Resources.Bots), m.doctrine)
}

func (m Model) resultView() string {
	r := m.state.LastBattle
	if r == nil { return "NETRISE // RESULT\n\nNo completed battle.\n[o] overview" }
	return fmt.Sprintf("NETRISE // RESULT\n\nTARGET: %s\nOUTCOME: %s\nSURVIVORS: %d\nBOTS LOST: %d\nLOOT: %d credits / %d data\nELAPSED: %d ticks\n\n[enter] overview  [t] targets", r.TargetID, r.Outcome, r.Survivors, r.BotsLost, r.CreditsLooted, r.DataLooted, r.ElapsedTicks)
}

func reachability(reachable bool) string { if reachable { return "reachable" }; return "locked" }

func (m *Model) applyCommand(command game.Command) {
	var events []game.Event
	var err error
	m.state, events, err = m.engine.Apply(m.state, command)
	if err != nil { m.status = err.Error() } else if len(events) > 0 { m.status = events[len(events)-1].Message }
	if m.state.ActiveBattle != nil { m.screen = "battle" }
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
