package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/sim"
	"testing"
)

func TestCompactResizeDoesNotPanic(t *testing.T) {
	m := New(sim.NewState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	if updated == nil {
		t.Fatal("nil model")
	}
}
