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

func TestProgressionGuideExplainsNextStepForBothFactions(t *testing.T) {
	runner := New(sim.NewState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	updated, _ := runner.Update(tea.KeyPressMsg(tea.Key{Text: "h", Code: 'h'}))
	runnerView := updated.(Model).View().Content
	for _, expected := range []string{"PROGRESSION GUIDE", "Install the bot compiler", "NEXT", "[g] upgrades"} {
		if !strings.Contains(runnerView, expected) {
			t.Fatalf("Runner guide omitted %q: %s", expected, runnerView)
		}
	}
	corp := New(sim.NewCorporationState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	updated, _ = corp.Update(tea.KeyPressMsg(tea.Key{Text: "h", Code: 'h'}))
	corpView := updated.(Model).View().Content
	if !strings.Contains(corpView, "Build a second client asset") || !strings.Contains(corpView, "Security Desk") {
		t.Fatalf("Corporation guide omitted campaign steps: %s", corpView)
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

func TestPreflightAdjustsCommitmentAndLaunchesSelectedAmount(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Upgrades:      []content.Upgrade{{ID: "secondhand_compiler"}},
		Targets:       []content.Target{{ID: "target", Name: "Relay", Reachable: true, Layers: []content.Layer{{Integrity: 20, ICE: 1}}, TimeLimit: 30}},
	}
	state := sim.NewState(2)
	state.Upgrades = []string{"secondhand_compiler"}
	state.Resources.Bots = 8
	m := New(state, sim.New(catalog, 2), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "t", Code: 't'}))
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "+", Code: '+'}))
	m = updated.(Model)
	if !strings.Contains(m.preflightView(), "COMMITMENT: 8 bots (100%)") || !strings.Contains(m.preflightView(), "BASELINE:") {
		t.Fatalf("preflight did not show adjustable commitment: %s", m.preflightView())
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = updated.(Model)
	if m.state.ActiveBattle == nil || m.state.ActiveBattle.CommittedBots != 8 {
		t.Fatalf("preflight launched wrong commitment: %#v", m.state.ActiveBattle)
	}
}

func TestTargetScreenExplainsIntelAndStartsRecon(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "target", Name: "Relay", Corporation: "Holloway Transit", Reachable: true, Layers: []content.Layer{{Integrity: 20, ICE: 1}}, ReconCost: 10, ReconTicks: 5, IntelGain: 40}},
	}
	state := sim.NewState(5)
	state.Upgrades = []string{"street_contacts"}
	state.Resources.Credits = mathx.New(20)
	m := New(state, sim.New(catalog, 5), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "t", Code: 't'}))
	m = updated.(Model)
	if !strings.Contains(m.targetsView(), "intel:0% (unknown)") || !strings.Contains(m.targetsView(), "recon ready: 10 CR / 5 ticks [i]") {
		t.Fatalf("target screen omitted recon guidance: %s", m.targetsView())
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "i", Code: 'i'}))
	m = updated.(Model)
	if m.state.Recon == nil || !strings.Contains(m.targetsView(), "recon active: 5 ticks") {
		t.Fatalf("target screen did not start recon: state=%#v view=%s", m.state, m.targetsView())
	}
}

func TestBattlePlaybackShowsCountersAndRecentEvents(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "target", Name: "Relay", Layers: []content.Layer{{Integrity: 30, ICE: 2}, {Integrity: 20, ICE: 1}}, TimeLimit: 30}},
	}
	state := sim.NewState(3)
	state.ActiveBattle = &game.Battle{TargetID: "target", Doctrine: game.Stealth, CommittedBots: 8, LiveBots: 6, LayerIndex: 0, LayerIntegrity: 21, DefenderICE: 2, TraceBP: 1800, ElapsedTicks: 4, TimeLimit: 30}
	state.HeatBP = 1200
	state.EventLog = []game.Event{{Tick: 1, Type: "BattleStarted", Message: "Breach underway."}, {Tick: 4, Type: "BotsLost", Message: "Countermeasure destroyed one bot."}}
	m := New(state, sim.New(catalog, 3), "")
	m.screen = "battle"
	view := m.View().Content
	for _, expected := range []string{"BATTLE PLAYBACK", "LAYER: 1/2", "LIVE BOTS: 6/8", "TRACE: 18%", "Countermeasure destroyed one bot", "[r] reinforce once", "withdraw"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("battle playback omitted %q: %s", expected, view)
		}
	}
}

func TestBattleResultShowsReasonAndEconomySummary(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "target", Name: "Relay"}},
	}
	state := sim.NewState(4)
	state.LastBattle = &game.BattleResult{TargetID: "target", Doctrine: game.Balanced, Outcome: "victory", CommittedBots: 8, Survivors: 6, BotsLost: 2, TraceBP: 1200, HeatAddedBP: 700, HeatAfterBP: 1900, CreditsLooted: 40, DataLooted: 10, ElapsedTicks: 12}
	m := New(state, sim.New(catalog, 4), "")
	m.screen = "result"
	view := m.View().Content
	for _, expected := range []string{"OUTCOME: victory", "REASON: all security layers breached", "BOTS: 8 committed / 6 survived / 2 lost", "HEAT: +7% (total 19%)", "LOOT: 40 credits / 10 data", "review the archive"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("result view omitted %q: %s", expected, view)
		}
	}
}
