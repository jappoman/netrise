package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
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

func TestEventArchiveViewShowsPersistedEvents(t *testing.T) {
	m := New(sim.NewState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	m.state.EventLog = []game.Event{{Tick: 9, Type: "BattleEnded", Message: "Breach successful."}}
	view := m.archiveView()
	if !strings.Contains(view, "Breach successful") || !strings.Contains(view, "BattleEnded") {
		t.Fatalf("archive view omitted persisted event: %s", view)
	}
}

func TestHelpViewExplainsBothCampaigns(t *testing.T) {
	runner := New(sim.NewState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	if !strings.Contains(runner.helpView(), "prepare, breach") {
		t.Fatal("Runner help omitted objective")
	}
	corp := New(sim.NewCorporationState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	if !strings.Contains(corp.helpView(), "information business") {
		t.Fatal("Corporation help omitted objective")
	}
}

func TestResearchEffectTextIsPlayerReadable(t *testing.T) {
	text := researchEffectText(content.ResearchEffects{
		ComputePerSecondBonus:       1,
		AttackBonusPercent:          8,
		StealthHeatReductionPercent: 15,
		AutomationIntervalReduction: 1,
	})
	for _, expected := range []string{"+1 compute/s", "+8% breach power", "-15% stealth heat", "automation every 1s faster"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("research effect text omitted %q: %s", expected, text)
		}
	}
}

func TestUpgradeScreenNavigatesAndPurchases(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Upgrades:      []content.Upgrade{{ID: "clock", Name: "Salvaged Clock", Cost: 10, Description: "Compute discipline.", Effects: content.UpgradeEffects{ComputePerSecondBonus: 1}}},
	}
	state := sim.NewState(1)
	state.Resources.Credits = mathx.New(100)
	m := New(state, sim.New(catalog, 1), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	m = updated.(Model)
	if m.screen != "upgrades" || !strings.Contains(m.upgradesView(), "Salvaged Clock") || !strings.Contains(m.upgradesView(), "+1 compute/s") {
		t.Fatalf("upgrade screen did not render selected catalog upgrade: %#v", m)
	}
	m.width, m.height = 40, 10
	if !strings.Contains(m.View().Content, "NETRISE // UPGRADES") {
		t.Fatal("compact terminal hid the upgrade screen")
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = updated.(Model)
	if !hasUpgrade(m.state, "clock") || m.state.Resources.Credits != mathx.New(90) {
		t.Fatalf("upgrade screen did not purchase selected upgrade: state=%#v status=%s", m.state, m.status)
	}
}
