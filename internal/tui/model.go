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
		}
		if command.Kind != "" {
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
		m.state = m.engine.Advance(m.state, 1)
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
	body := fmt.Sprintf("CREDITS %s\nCOMPUTE %s (+%d/s)\nBOTS %d/%d\nTICK %d\n\n[c] contract  [b] buy rig  [f] fabricate bot  [q] quit\n%s", m.state.Resources.Credits, m.state.Resources.Compute, m.engine.ComputeRate(m.state), m.state.Resources.Bots, m.state.Resources.BotCapacity, m.state.LogicalTick, m.status)
	return tea.NewView(title + "\n\n" + body + "\n")
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
