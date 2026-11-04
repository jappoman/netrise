package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/sim"
	"time"
)

type tickMsg struct{}
type Model struct {
	state         game.State
	engine        *sim.Engine
	width, height int
}

func New(s game.State, e *sim.Engine) Model { return Model{state: s, engine: e} }
func (m Model) Init() tea.Cmd               { return tea.Tick(1e9, func(time.Time) tea.Msg { return tickMsg{} }) }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		if v.String() == "q" || v.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if v.String() == "c" {
			m.state, _, _ = m.engine.Apply(m.state, game.Command{Kind: game.CollectContract})
		}
	case tickMsg:
		m.state = m.engine.Advance(m.state, 1)
		return m, m.Init()
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
	}
	return m, nil
}
func (m Model) View() tea.View {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("NETRISE // RUNNER TERMINAL")
	body := fmt.Sprintf("CREDITS %s\nCOMPUTE %s/s\nBOTS %d/%d\nTICK %d\n\n[c] collect contract  [q] quit", m.state.Resources.Credits, m.state.Resources.Compute, m.state.Resources.Bots, m.state.Resources.BotCapacity, m.state.LogicalTick)
	return tea.NewView(title + "\n\n" + body + "\n")
}
