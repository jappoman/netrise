package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/sim"
	"strings"
	"testing"
)

func TestCompactResizeDoesNotPanic(t *testing.T) {
	m := New(sim.NewState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	if updated == nil {
		t.Fatal("nil model")
	}
}

func TestCorporationOverviewRendersAtCompactSize(t *testing.T) {
	m := New(sim.NewCorporationState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	if updated == nil {
		t.Fatal("nil Corporation model")
	}
	view := updated.(Model).View()
	if view.Content == "" {
		t.Fatal("empty Corporation view")
	}
	if !strings.Contains(view.Content, "[b] hosting") {
		t.Fatal("compact Corporation view hides the action rail")
	}
}

func TestNoColorOmitsANSIStyling(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New(sim.NewState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	view := updated.(Model).View()
	if strings.Contains(view.Content, "\x1b[") {
		t.Fatal("NO_COLOR view contains ANSI escape")
	}
}
