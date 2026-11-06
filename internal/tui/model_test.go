package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"github.com/jappoman/netrise/internal/save"
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

func TestQuitPersistsActiveBattleForResume(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "target", Name: "Relay", Layers: []content.Layer{{Integrity: 30, ICE: 2}}, TimeLimit: 30}},
	}
	state := sim.NewState(13)
	state.LogicalTick = 42
	state.ActiveBattle = &game.Battle{TargetID: "target", AttackerFaction: game.Runner, Doctrine: game.Stealth, CommittedBots: 8, LiveBots: 6, LayerIntegrity: 18, DefenderICE: 1, ElapsedTicks: 4, TimeLimit: 30}
	dir := t.TempDir()
	m := New(state, sim.New(catalog, 13), dir)
	updated, quit := m.Update(tea.KeyPressMsg(tea.Key{Text: "q", Code: 'q'}))
	if updated == nil || quit == nil {
		t.Fatal("quit did not return a model and quit command")
	}
	got, err := save.Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.LogicalTick != 42 || got.ActiveBattle == nil || got.ActiveBattle.TargetID != "target" || got.ActiveBattle.LiveBots != 6 {
		t.Fatalf("quit did not persist resumable battle: %#v", got)
	}
	if got.SavedAtUTC == "" {
		t.Fatal("quit save did not record save time")
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

func TestCorporationCanBuildArchiveServerFromOverview(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		CorpAssets:    []content.CorpAssetDefinition{{ID: "archive_server", Name: "Archive Server", CostCapital: 150, CapitalPerTick: 7, UpkeepPerTick: 2, Integrity: 140}},
	}
	state := sim.NewCorporationState(14)
	state.Corp.Capital = mathx.New(200)
	m := New(state, sim.New(catalog, 14), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "a", Code: 'a'}))
	m = updated.(Model)
	if !m.corpHasAsset("archive_server") || m.state.Corp.Capital != mathx.New(50) {
		t.Fatalf("archive server action did not build asset atomically: state=%#v status=%s", m.state.Corp, m.status)
	}
}

func TestCorporationOverviewShowsDisabledAssetRecoveryHint(t *testing.T) {
	state := sim.NewCorporationState(15)
	state.Corp.Assets = append(state.Corp.Assets, game.CorpAsset{ID: "archive", Name: "Archive", UpkeepPerTick: 5, Operational: false})
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 15), "")
	view := m.corpOverviewView()
	if !strings.Contains(view, "ASSETS 1/2 operational") || !strings.Contains(view, "capital is recovering") {
		t.Fatalf("Corporation recovery state was not explained: %s", view)
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

func TestHighContrastModeIsSelectableAndNoColorWins(t *testing.T) {
	t.Setenv("NETRISE_HIGH_CONTRAST", "1")
	contrast := terminalTitleStyle().Render("TITLE")
	if !strings.Contains(contrast, "TITLE") {
		t.Fatal("high-contrast mode did not select a foreground color")
	}
	t.Setenv("NO_COLOR", "1")
	plain := terminalTitleStyle().Render("TITLE")
	if strings.Contains(plain, "\x1b[") {
		t.Fatal("NO_COLOR did not override high-contrast styling")
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
	if !strings.Contains(runner.helpView(), "prepare, breach") || !strings.Contains(runner.helpView(), "[a] breach") {
		t.Fatal("Runner help omitted objective")
	}
	corp := New(sim.NewCorporationState(1), sim.New(content.Catalog{SchemaVersion: 1}, 1), "")
	if !strings.Contains(corp.helpView(), "information business") || !strings.Contains(corp.helpView(), "[a] archive") {
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

func TestAutomationScreenConfiguresBudgetAndPolicy(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Upgrades:      []content.Upgrade{{ID: "autodialer", Name: "Autodialer"}},
	}
	state := sim.NewState(6)
	state.Upgrades = []string{"autodialer"}
	m := New(state, sim.New(catalog, 6), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "A", Code: 'A'}))
	m = updated.(Model)
	if m.screen != "automation" || !strings.Contains(m.automationView(), "SELECTED BUDGET: 1000 credits") {
		t.Fatalf("automation screen did not open with default budget: %s", m.automationView())
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "+", Code: '+'}))
	m = updated.(Model)
	if !strings.Contains(m.automationView(), "SELECTED BUDGET: 5000 credits") {
		t.Fatalf("automation budget preset did not advance: %s", m.automationView())
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = updated.(Model)
	if !m.state.Automation.Contracts || m.state.Automation.ContractBudget != 5000 {
		t.Fatalf("automation policy did not persist configured budget: %#v", m.state.Automation)
	}
}

func TestCorporationMaintenanceScreenConfiguresBudget(t *testing.T) {
	m := New(sim.NewCorporationState(7), sim.New(content.Catalog{SchemaVersion: 1}, 7), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "A", Code: 'A'}))
	m = updated.(Model)
	if m.screen != "automation" || !strings.Contains(m.View().Content, "CORPORATION POLICIES") {
		t.Fatalf("Corporation maintenance screen did not open: %s", m.View().Content)
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "+", Code: '+'}))
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = updated.(Model)
	if m.state.Corp == nil || !m.state.Corp.MaintenanceAuto || m.state.Corp.MaintenanceBudget != 5000 {
		t.Fatalf("Corporation maintenance budget did not persist: %#v", m.state.Corp)
	}
}

func TestCorporationReportShowsDerivedMarginAndProjection(t *testing.T) {
	state := sim.NewCorporationState(8)
	state.Corp.DemandBP = 8000
	state.Corp.MaintenanceAuto = true
	state.Corp.MaintenanceBudget = 1000
	state.Corp.MaintenanceReserve = 5
	state.Corp.ICE = 3
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 8), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "v", Code: 'v'}))
	m = updated.(Model)
	view := m.View().Content
	for _, expected := range []string{"CORPORATION REPORT", "GROSS REVENUE: +3/tick", "UPKEEP: -0/tick", "OPERATING MARGIN: +3/tick", "REPUTATION: 80%  EXPOSURE: 0%", "REVENUE FACTOR: 100%", "NEXT TICK PROJECTION", "MAINTENANCE: ON"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("Corporation report omitted %q: %s", expected, view)
		}
	}
}

func TestCorporationReportShowsCompletedTechnology(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		CorpResearch:  []content.CorpResearchNode{{ID: "efficiency", Name: "Network Efficiency", Threshold: 20}},
	}
	state := sim.NewCorporationState(9)
	state.Corp.Research = mathx.New(20)
	state.Corp.CompletedResearch = []string{"efficiency"}
	m := New(state, sim.New(catalog, 9), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "v", Code: 'v'}))
	view := updated.(Model).View().Content
	for _, expected := range []string{"RESEARCH: 20", "TECHNOLOGIES: Network Efficiency"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("Corporation report omitted %q: %s", expected, view)
		}
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
	m.state.Recon = nil
	m.state.TargetCooldowns = map[string]int64{"target": 7}
	if !strings.Contains(m.targetsView(), "COOLDOWN") || !strings.Contains(m.targetsView(), "cooldown active: 7 ticks") {
		t.Fatalf("target screen omitted cooldown guidance: %s", m.targetsView())
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
	for _, expected := range []string{"BATTLE PLAYBACK", "LAYER: 1/2", "LIVE BOTS: 6/8", "TRACE: 18%", "Countermeasure destroyed one bot", "[space] pause", "[1/2/4] speed", "[r] reinforce once", "withdraw", "RUNNING at 1x"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("battle playback omitted %q: %s", expected, view)
		}
	}
}

func TestBattlePlaybackPauseSpeedAndResume(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "target", Name: "Relay", Layers: []content.Layer{{Integrity: 30, ICE: 2}}, TimeLimit: 30}},
	}
	state := sim.NewState(12)
	state.ActiveBattle = &game.Battle{TargetID: "target", Doctrine: game.Balanced, CommittedBots: 8, LiveBots: 8, LayerIntegrity: 30, DefenderICE: 2, TimeLimit: 30}
	m := New(state, sim.New(catalog, 12), "")
	m.screen = "battle"
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: " ", Code: tea.KeySpace}))
	m = updated.(Model)
	if !m.paused || !strings.Contains(m.battleView(), "PAUSED") {
		t.Fatalf("space did not pause playback: paused=%v view=%s", m.paused, m.battleView())
	}
	before := m.state.LogicalTick
	updated, _ = m.Update(tickMsg{})
	m = updated.(Model)
	if m.state.LogicalTick != before {
		t.Fatalf("paused playback advanced simulation: before=%d after=%d", before, m.state.LogicalTick)
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "4", Code: '4'}))
	m = updated.(Model)
	if m.playbackSpeed != 4 {
		t.Fatalf("speed control did not select 4x: %d", m.playbackSpeed)
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: " ", Code: tea.KeySpace}))
	m = updated.(Model)
	if m.paused {
		t.Fatal("space did not resume playback")
	}
	updated, _ = m.Update(tickMsg{})
	m = updated.(Model)
	if m.state.LogicalTick != before+4 {
		t.Fatalf("4x playback advanced %d ticks, want 4", m.state.LogicalTick-before)
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
