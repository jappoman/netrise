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
}

func New(s game.State, e *sim.Engine, dataDir string) Model {
	return Model{state: s, engine: e, dataDir: dataDir}
}
func (m Model) Init() tea.Cmd { return tea.Tick(1e9, func(time.Time) tea.Msg { return tickMsg{} }) }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		if v.String() == "q" || v.String() == "ctrl+c" {
			m.saveNow()
			return m, tea.Quit
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
			command = game.Command{Kind: game.StartBattle, ID: "t_01", Quantity: m.state.Resources.Bots, Doctrine: game.Balanced}
		case "w":
			command = game.Command{Kind: game.WithdrawBattle}
		}
		if command.Kind != "" && !(command.Kind == game.StartBattle && command.Quantity == 0) {
			var events []game.Event
			var err error
			m.state, events, err = m.engine.Apply(m.state, command)
			if err != nil {
				m.status = err.Error()
			} else if len(events) > 0 {
				m.status = events[len(events)-1].Message
			}
			m.saveNow()
		}
	case tickMsg:
		var events []game.Event
		m.state, events, _ = m.engine.Apply(m.state, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if len(events) > 0 {
			m.status = events[len(events)-1].Message
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
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("NETRISE // RUNNER TERMINAL")
	battle := "NO ACTIVE BATTLE"
	if m.state.ActiveBattle != nil {
		battle = fmt.Sprintf("BATTLE %s | LAYER %d | BOTS %d | INTEGRITY %d | ICE %d | TRACE %d%%", m.state.ActiveBattle.TargetID, m.state.ActiveBattle.LayerIndex+1, m.state.ActiveBattle.LiveBots, m.state.ActiveBattle.LayerIntegrity, m.state.ActiveBattle.DefenderICE, m.state.ActiveBattle.TraceBP/100)
	}
	body := fmt.Sprintf("CREDITS %s\nCOMPUTE %s (+%d/s)\nBOTS %d/%d\nTICK %d\n%s\n\n[c] contract  [b] rig  [u] compiler  [f] bot  [a] attack  [w] withdraw  [q] quit\nNEXT: %s\n%s", m.state.Resources.Credits, m.state.Resources.Compute, m.engine.ComputeRate(m.state), m.state.Resources.Bots, m.state.Resources.BotCapacity, m.state.LogicalTick, battle, m.nextHint(), m.status)
	return tea.NewView(title + "\n\n" + body + "\n")
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
