package tui

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"github.com/jappoman/netrise/internal/save"
	"github.com/jappoman/netrise/internal/sim"
	"math"
	"os"
	"path/filepath"
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

func TestPlayerViewUsesFullScreenAlternateBuffer(t *testing.T) {
	m := New(sim.NewState(2), sim.New(content.Catalog{SchemaVersion: 1}, 2), "")
	if view := m.View(); !view.AltScreen {
		t.Fatal("player-facing TUI did not request the terminal alternate screen")
	}
}

func TestUpgradeMenuFitsAtStandardTerminalHeight(t *testing.T) {
	upgrades := make([]content.Upgrade, 14)
	for i := range upgrades {
		upgrades[i] = content.Upgrade{ID: fmt.Sprintf("upgrade_%d", i), Name: fmt.Sprintf("Upgrade %d", i), Description: "A safe catalogued improvement.", Cost: 1}
	}
	m := New(sim.NewState(3), sim.New(content.Catalog{SchemaVersion: 1, Upgrades: upgrades}, 3), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(Model)
	m.screen = "upgrades"
	lines := strings.Split(strings.TrimSuffix(m.View().Content, "\n"), "\n")
	if len(lines) > 24 || !strings.Contains(m.View().Content, "SHOWING 1-7/14") || !strings.Contains(m.View().Content, "NEXT:") {
		t.Fatalf("upgrade menu is not usable at 80x24: lines=%d view=%s", len(lines), m.View().Content)
	}
	for i := 0; i < 13; i++ {
		updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "down", Code: tea.KeyDown}))
		m = updated.(Model)
	}
	if !strings.Contains(m.View().Content, "SHOWING 8-14/14") || !strings.Contains(m.View().Content, "> Upgrade 13") {
		t.Fatalf("upgrade menu did not scroll with selection: %s", m.View().Content)
	}
}

func TestOverviewActionRailsFitStandardTerminalWidth(t *testing.T) {
	for _, faction := range []game.Faction{game.Runner, game.Corporation} {
		state := sim.NewState(4)
		if faction == game.Corporation {
			state = sim.NewCorporationState(4)
		}
		m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 4), "")
		updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := updated.(Model).View().Content
		for _, line := range strings.Split(view, "\n") {
			if len(line) > 80 {
				t.Fatalf("%s overview line exceeds 80 columns (%d): %q", faction, len(line), line)
			}
		}
		if !strings.Contains(view, "[q] quit") || !strings.Contains(view, "NEXT:") {
			t.Fatalf("%s overview lost its action rail: %s", faction, view)
		}
	}
}

func TestPlayerScreensFitStandardTerminalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "target", Name: "Relay", Reachable: true, Layers: []content.Layer{{Integrity: 10}}, TimeLimit: 30}},
		Research:      []content.ResearchNode{{ID: "research", Name: "Research", Description: "A catalogued research option."}},
		Upgrades:      []content.Upgrade{{ID: "upgrade", Name: "Upgrade", Description: "A catalogued upgrade option.", Cost: 1}},
	}
	tests := []struct {
		name   string
		state  game.State
		screen string
	}{
		{name: "runner overview", state: sim.NewState(5), screen: "overview"},
		{name: "runner targets", state: sim.NewState(5), screen: "targets"},
		{name: "runner preflight", state: sim.NewState(5), screen: "preflight"},
		{name: "runner research", state: sim.NewState(5), screen: "research"},
		{name: "runner upgrades", state: sim.NewState(5), screen: "upgrades"},
		{name: "runner automation", state: sim.NewState(5), screen: "automation"},
		{name: "runner archive", state: sim.NewState(5), screen: "archive"},
		{name: "runner settings", state: sim.NewState(5), screen: "settings"},
		{name: "runner help", state: sim.NewState(5), screen: "help"},
		{name: "runner guide", state: sim.NewState(5), screen: "tutorial"},
		{name: "runner prestige", state: sim.NewState(5), screen: "prestige"},
		{name: "corporation overview", state: sim.NewCorporationState(5), screen: "overview"},
		{name: "corporation policy", state: sim.NewCorporationState(5), screen: "automation"},
		{name: "corporation report", state: sim.NewCorporationState(5), screen: "corp-report"},
		{name: "corporation archive", state: sim.NewCorporationState(5), screen: "archive"},
		{name: "corporation settings", state: sim.NewCorporationState(5), screen: "settings"},
		{name: "corporation help", state: sim.NewCorporationState(5), screen: "help"},
		{name: "corporation guide", state: sim.NewCorporationState(5), screen: "tutorial"},
		{name: "corporation restructure", state: sim.NewCorporationState(5), screen: "restructure"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := New(tc.state, sim.New(catalog, 5), "")
			updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			m = updated.(Model)
			m.screen = tc.screen
			content := m.View().Content
			lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
			if len(lines) > 24 {
				t.Fatalf("screen exceeds 24 rows (%d): %s", len(lines), content)
			}
			for _, line := range lines {
				if len(line) > 80 {
					t.Fatalf("line exceeds 80 columns (%d): %q", len(line), line)
				}
			}
		})
	}
}

func TestRunnerOverviewShowsDataCapacity(t *testing.T) {
	state := sim.NewState(20)
	state.Resources.Data = mathx.New(7)
	state.Resources.DataCapacity = 100
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 20), "")
	if !strings.Contains(m.View().Content, "DATA 7/100") {
		t.Fatalf("Runner overview hid data capacity: %s", m.View().Content)
	}
}

func TestOfflineSummaryExplainsCapInOverview(t *testing.T) {
	state := sim.NewState(409)
	state.OfflineSummary = &game.OfflineSummary{Ticks: sim.OfflineCapTicks, Capped: true}
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 409), "")
	if !strings.Contains(m.offlineLine(), "capped at 8h") {
		t.Fatalf("offline cap warning was not player-visible: %s", m.offlineLine())
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

func TestCtrlCPersistsActiveBattleForResume(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "target", Name: "Relay", Layers: []content.Layer{{Integrity: 30, ICE: 2}}, TimeLimit: 30}},
	}
	state := sim.NewState(14)
	state.LogicalTick = 7
	state.ActiveBattle = &game.Battle{TargetID: "target", AttackerFaction: game.Runner, Doctrine: game.Balanced, CommittedBots: 4, LiveBots: 3, LayerIntegrity: 22, DefenderICE: 1, ElapsedTicks: 2, TimeLimit: 30}
	dir := t.TempDir()
	m := New(state, sim.New(catalog, 14), dir)
	updated, quit := m.Update(tea.KeyPressMsg(tea.Key{Text: "ctrl+c"}))
	if updated == nil || quit == nil {
		t.Fatal("Ctrl+C did not return a model and quit command")
	}
	got, err := save.Read(dir)
	if err != nil || got.ActiveBattle == nil || got.ActiveBattle.LiveBots != 3 || got.LogicalTick != 7 {
		t.Fatalf("Ctrl+C did not persist resumable battle: state=%#v err=%v", got, err)
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

func TestCorporationOverviewShowsAssetIntegrity(t *testing.T) {
	state := sim.NewCorporationState(16)
	state.Corp.Assets[0].Integrity = 60
	state.Corp.Assets[0].MaxIntegrity = 100
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 16), "")
	view := m.corpOverviewView()
	if !strings.Contains(view, "Starter Client 60/100 ON") || !strings.Contains(view, "[r] to repair Starter Client") {
		t.Fatalf("Corporation overview hid asset integrity/recovery action: %s", view)
	}
}

func TestCorporationEmergencyIsolationOpensResultImmediately(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "t_01", Name: "Audit", Layers: []content.Layer{{Integrity: 30, ICE: 0}}, TimeLimit: 30}},
	}
	state := sim.NewCorporationState(17)
	state.Corp.Capital = mathx.New(100)
	state.Corp.Assets = append(state.Corp.Assets, game.CorpAsset{ID: "archive", Name: "Archive", CapitalPerTick: 7, Integrity: 100, MaxIntegrity: 100, Operational: true, IsolationCostCapital: 10, IsolationTicks: 5})
	state.ActiveBattle = &game.Battle{TargetID: "t_01", DefenderAssetID: "archive", AttackerFaction: game.Runner, CommittedBots: 8, LiveBots: 8, LayerIntegrity: 30, TimeLimit: 30}
	m := New(state, sim.New(catalog, 17), "")
	m.screen = "battle"
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "x", Code: 'x'}))
	m = updated.(Model)
	if m.screen != "result" || m.state.ActiveBattle != nil || m.state.LastBattle == nil || m.state.LastBattle.Outcome != "isolated" {
		t.Fatalf("isolation did not transition to result: screen=%q state=%#v", m.screen, m.state)
	}
}

func TestCorporationRepairKeyUsesEngineCommand(t *testing.T) {
	state := sim.NewCorporationState(18)
	state.Corp.Capital = mathx.New(50)
	state.Corp.Assets[0].Integrity = 60
	state.Corp.Assets[0].MaxIntegrity = 100
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 18), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "r", Code: 'r'}))
	m = updated.(Model)
	if m.state.Corp.Assets[0].Integrity != 70 || m.state.Corp.Capital != mathx.New(40) || !strings.Contains(m.status, "repaired") {
		t.Fatalf("repair key did not use atomic engine command: state=%#v status=%q", m.state.Corp, m.status)
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

func TestSettingsScreenShowsAccessibilityModesAndReturnsToOverview(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("NETRISE_HIGH_CONTRAST", "1")
	m := New(sim.NewState(21), sim.New(content.Catalog{SchemaVersion: 1}, 21), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "S", Code: 'S'}))
	m = updated.(Model)
	if m.screen != "settings" {
		t.Fatalf("settings key did not open screen: %q", m.screen)
	}
	view := m.View().Content
	if !strings.Contains(view, "NO_COLOR: ON") || !strings.Contains(view, "HIGH CONTRAST: ON") || !strings.Contains(view, "REDUCED MOTION: OFF") {
		t.Fatalf("settings screen hid accessibility state: %s", view)
	}
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "esc", Code: tea.KeyEscape}))
	if updated.(Model).screen != "overview" {
		t.Fatal("settings screen did not return to overview")
	}
}

func TestASCIIFallbackRemovesUnicodeFromPlayerView(t *testing.T) {
	t.Setenv("NETRISE_ASCII", "1")
	state := sim.NewState(22)
	e := sim.New(content.Catalog{SchemaVersion: 1, Targets: []content.Target{{ID: "target", Name: "Réseau", Reachable: true, Layers: []content.Layer{{Integrity: 1}}, TimeLimit: 1}}}, 22)
	m := New(state, e, "")
	m.screen = "targets"
	view := m.View().Content
	if strings.IndexFunc(view, func(r rune) bool { return r > 127 }) >= 0 {
		t.Fatalf("ASCII fallback left Unicode in player view: %s", view)
	}
	if !strings.Contains(view, "R?seau") {
		t.Fatalf("ASCII fallback did not preserve a readable replacement: %s", view)
	}
}

func TestReducedMotionResolvesBattleWithoutPlaybackFrames(t *testing.T) {
	t.Setenv("NETRISE_REDUCED_MOTION", "1")
	catalog := content.Catalog{SchemaVersion: 1, Targets: []content.Target{{ID: "target", Name: "Relay", Reachable: true, Layers: []content.Layer{{Integrity: 1}}, TimeLimit: 30}}}
	state := sim.NewState(19)
	state.Resources.Bots = 1
	state.Upgrades = []string{"secondhand_compiler"}
	e := sim.New(catalog, 19)
	var err error
	state, _, err = e.Apply(state, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 1, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	m := New(state, e, "")
	m.screen = "battle"
	updated, _ := m.Update(tickMsg{})
	got := updated.(Model)
	if got.state.ActiveBattle != nil || got.state.LastBattle == nil || got.screen != "result" {
		t.Fatalf("reduced motion left playback active: battle=%#v result=%#v screen=%q", got.state.ActiveBattle, got.state.LastBattle, got.screen)
	}
}

func TestRunnerKeyboardFlowReachesBattleResultEndToEnd(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	state := sim.NewState(17)
	state.Upgrades = []string{"secondhand_compiler"}
	state.Resources.Bots = 10
	m := New(state, sim.New(catalog, 17), t.TempDir())
	key := func(text string) tea.KeyPressMsg { return tea.KeyPressMsg(tea.Key{Text: text, Code: rune(text[0])}) }

	updated, _ := m.Update(key("t"))
	m = updated.(Model)
	if m.screen != "targets" {
		t.Fatalf("targets screen was not opened: %q", m.screen)
	}
	updated, _ = m.Update(key("enter"))
	m = updated.(Model)
	if m.screen != "preflight" {
		t.Fatalf("preflight screen was not opened: %q", m.screen)
	}
	updated, _ = m.Update(key("enter"))
	m = updated.(Model)
	if m.screen != "battle" || m.state.ActiveBattle == nil {
		t.Fatalf("keyboard launch did not start battle: screen=%q state=%#v", m.screen, m.state.ActiveBattle)
	}
	for ticks := 0; m.state.ActiveBattle != nil && ticks <= 30; ticks++ {
		updated, _ = m.Update(tickMsg{})
		m = updated.(Model)
	}
	if m.screen != "result" || m.state.LastBattle == nil || m.state.LastBattle.Outcome != "victory" {
		t.Fatalf("keyboard flow did not reach victory result: screen=%q result=%#v", m.screen, m.state.LastBattle)
	}
	if !strings.Contains(m.View().Content, "OUTCOME: victory") || !strings.Contains(m.View().Content, "NEXT:") {
		t.Fatalf("result screen did not explain outcome and next action: %s", m.View().Content)
	}
}

func TestCorporationKeyboardFlowRespondsToIncursionEndToEnd(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	state := sim.NewCorporationState(23)
	m := New(state, sim.New(catalog, 23), t.TempDir())
	key := func(text string) tea.KeyPressMsg { return tea.KeyPressMsg(tea.Key{Text: text, Code: rune(text[0])}) }
	updated, _ := m.Update(key("s"))
	m = updated.(Model)
	if len(m.state.Corp.Assets) != 2 || m.state.Corp.Capital != mathx.New(10) {
		t.Fatalf("security action did not build asset: %#v", m.state.Corp)
	}
	for ticks := 0; ticks < 30 && m.state.ActiveBattle == nil; ticks++ {
		updated, _ = m.Update(tickMsg{})
		m = updated.(Model)
	}
	if m.screen != "battle" || m.state.ActiveBattle == nil {
		t.Fatalf("active Corporation incursion did not reach battle screen: screen=%q battle=%#v", m.screen, m.state.ActiveBattle)
	}
	updated, _ = m.Update(key("x"))
	m = updated.(Model)
	if m.screen != "result" || m.state.ActiveBattle != nil || m.state.LastBattle == nil || m.state.LastBattle.Outcome != "isolated" {
		t.Fatalf("keyboard isolation did not resolve incursion: screen=%q battle=%#v result=%#v", m.screen, m.state.ActiveBattle, m.state.LastBattle)
	}
	if !strings.Contains(m.View().Content, "OUTCOME: isolated") || !strings.Contains(m.View().Content, "NEXT:") {
		t.Fatalf("Corporation result omitted outcome and next action: %s", m.View().Content)
	}
}

func TestCorporationCannotOpenRunnerTargets(t *testing.T) {
	m := New(sim.NewCorporationState(405), sim.New(content.Catalog{SchemaVersion: 1}, 405), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "t", Code: 't'}))
	got := updated.(Model)
	if got.screen != "overview" || !strings.Contains(got.status, "Runner operations") {
		t.Fatalf("Corporation crossed into Runner targets: screen=%q status=%q", got.screen, got.status)
	}
}

func TestCorporationCannotOpenRunnerResearch(t *testing.T) {
	m := New(sim.NewCorporationState(406), sim.New(content.Catalog{SchemaVersion: 1}, 406), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "y", Code: 'y'}))
	got := updated.(Model)
	if got.screen != "overview" || !strings.Contains(got.status, "Runner operation") {
		t.Fatalf("Corporation crossed into Runner research: screen=%q status=%q", got.screen, got.status)
	}
}

func TestSecondaryScreensExplainTheirNextAction(t *testing.T) {
	runner := New(sim.NewState(407), sim.New(content.Catalog{SchemaVersion: 1}, 407), "")
	for name, view := range map[string]string{
		"runner archive":  runner.archiveView(),
		"runner help":     runner.helpView(),
		"runner research": runner.researchView(),
		"runner prestige": runner.prestigeView(),
	} {
		if !strings.Contains(view, "NEXT:") {
			t.Fatalf("%s omitted next-action rail: %s", name, view)
		}
	}
	corp := New(sim.NewCorporationState(408), sim.New(content.Catalog{SchemaVersion: 1}, 408), "")
	for name, view := range map[string]string{
		"corp archive":     corp.archiveView(),
		"corp help":        corp.helpView(),
		"corp report":      corp.corpReportView(),
		"corp restructure": corp.restructureView(),
	} {
		if !strings.Contains(view, "NEXT:") {
			t.Fatalf("%s omitted next-action rail: %s", name, view)
		}
	}
}

func TestResearchScreenExplainsMissingPrerequisites(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Research: []content.ResearchNode{
			{ID: "hardware", Name: "Hardware", CostData: 10},
			{ID: "intrusion", Name: "Intrusion", CostData: 20, Prerequisites: []string{"hardware"}},
		},
	}
	m := New(sim.NewState(409), sim.New(catalog, 409), "")
	view := m.researchView()
	if !strings.Contains(view, "LOCKED (requires hardware)") {
		t.Fatalf("research screen hid missing prerequisite: %s", view)
	}
}

func TestResearchScreenExplainsInsufficientData(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Research:      []content.ResearchNode{{ID: "hardware", Name: "Hardware", CostData: 10}},
	}
	state := sim.NewState(410)
	state.Resources.Data = mathx.New(3)
	m := New(state, sim.New(catalog, 410), "")
	if !strings.Contains(m.researchView(), "DATA: 3 / 100 capacity") || !strings.Contains(m.researchView(), "NEED 7 DATA") {
		t.Fatalf("research screen hid insufficient data reason: %s", m.researchView())
	}
}

func TestUpgradeScreenShowsCurrentCredits(t *testing.T) {
	state := sim.NewState(411)
	state.Resources.Credits = mathx.New(42)
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1, Upgrades: []content.Upgrade{{ID: "u", Name: "Upgrade", Cost: 50}}}, 411), "")
	if !strings.Contains(m.upgradesView(), "CREDITS: 42") {
		t.Fatalf("upgrade screen hid current credits: %s", m.upgradesView())
	}
}

func TestEmptyPlayerScreensExplainTheirNextAction(t *testing.T) {
	runner := New(sim.NewState(410), sim.New(content.Catalog{SchemaVersion: 1}, 410), "")
	for name, view := range map[string]string{
		"compact battle": runner.compactBattleView(),
		"battle":         runner.battleView(),
		"result":         runner.resultView(),
	} {
		if !strings.Contains(view, "NEXT:") {
			t.Fatalf("empty %s omitted next-action rail: %s", name, view)
		}
	}
	corp := New(sim.NewCorporationState(411), sim.New(content.Catalog{SchemaVersion: 1}, 411), "")
	corp.state.Corp = nil
	if !strings.Contains(corp.corpOverviewView(), "NEXT:") || !strings.Contains(corp.corpReportView(), "NEXT:") || !strings.Contains(corp.corpMaintenanceView(), "NEXT:") {
		t.Fatal("empty Corporation screens omitted next-action rail")
	}
}

func TestRunnerGuideExplainsTemporaryAccessWindow(t *testing.T) {
	state := sim.NewState(20)
	state.CompletedTargets = []string{"t_01"}
	state.Access = map[string]int64{"t_01": 30}
	state.Resources.Bots = 1
	state.Upgrades = []string{"secondhand_compiler"}
	state.Generators = []game.Generator{{ID: "salvaged_rig", Count: 1}}
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 20), "")
	guide := m.tutorialView()
	if !strings.Contains(guide, "Use temporary access") || !strings.Contains(guide, "before the access window expires") {
		t.Fatalf("guide omitted temporary access step: %s", guide)
	}
	if !strings.Contains(m.nextHint(), "before temporary access expires") {
		t.Fatalf("next hint omitted active access window: %s", m.nextHint())
	}
	state.Access["t_01"] = 0
	m.state = state
	if !strings.Contains(m.nextHint(), "inspect adjacent routes") {
		t.Fatalf("next hint omitted expired access recovery: %s", m.nextHint())
	}
}

func TestTargetScreenNextHintTracksAccessWindow(t *testing.T) {
	state := sim.NewState(21)
	state.CompletedTargets = []string{"t_01"}
	state.Access = map[string]int64{"t_01": 12}
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 21), "")
	if !strings.Contains(m.targetsNextHint(), "launch t_02 before temporary access expires") {
		t.Fatalf("target rail omitted active access action: %s", m.targetsNextHint())
	}
	state.Access["t_01"] = 0
	m.state = state
	if !strings.Contains(m.targetsNextHint(), "replay t_01") {
		t.Fatalf("target rail omitted expired access recovery: %s", m.targetsNextHint())
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
	if !strings.Contains(corp.helpView(), "information business") || !strings.Contains(corp.helpView(), "[a] archive") || !strings.Contains(corp.helpView(), "[x] isolate") {
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
	for _, expected := range []string{"RESEARCH: 20", "TECHNOLOGIES: Network Efficiency", "ASSET STATUS: Starter Client 100/100 ON"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("Corporation report omitted %q: %s", expected, view)
		}
	}
}

func TestCorporationMarketStatusAndNextActionAreVisible(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		CorpMarkets:   []content.CorpMarketExpansion{{ID: "metro", Name: "Metro Contracts", CostCapital: 150, DemandBonusBP: 1000}},
	}
	m := New(sim.NewCorporationState(10), sim.New(catalog, 10), "")
	m.state.Corp.Assets = append(m.state.Corp.Assets,
		game.CorpAsset{ID: "security_desk", Name: "Security Desk", Operational: true, Integrity: 100, MaxIntegrity: 100},
		game.CorpAsset{ID: "archive_server", Name: "Archive Server", Operational: true, Integrity: 100, MaxIntegrity: 100},
	)
	m.state.Corp.ICE = m.state.Corp.ICECapacity
	view := m.corpOverviewView()
	if !strings.Contains(view, "MARKET 0/1 | next Metro Contracts (150 capital, +10% demand) | need 50 capital") || !strings.Contains(view, "press [m] to secure Metro Contracts") {
		t.Fatalf("Corporation market state was not self-explanatory: %s", view)
	}
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "m", Code: 'm'}))
	m = updated.(Model)
	if m.state.Corp.MarketTier != 0 || !strings.Contains(m.status, "requires 150 capital") {
		t.Fatalf("unaffordable market expansion was not explained atomically: state=%#v status=%q", m.state.Corp, m.status)
	}
}

func TestCorporationIncursionBudgetIsVisibleAcrossViews(t *testing.T) {
	state := sim.NewCorporationState(13)
	state.Corp.IncursionSpent = 8
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 13), "")
	overview := m.corpOverviewView()
	if !strings.Contains(overview, "NPC INCURSION BUDGET: 16/24 bot commitment remaining") {
		t.Fatalf("Corporation overview hid incursion budget: %s", overview)
	}
	m.width = 79
	m.height = 23
	compact := m.compactView()
	if !strings.Contains(compact, "NPC INCURSION BUDGET: 16/24 bot commitment remaining") {
		t.Fatalf("compact Corporation view hid incursion budget: %s", compact)
	}
	if !strings.Contains(m.helpView(), "finite NPC incursion budget") {
		t.Fatal("Corporation help omitted incursion budget explanation")
	}
}

func TestCorporationReportSaturatesExtremeProjection(t *testing.T) {
	state := sim.NewCorporationState(14)
	state.Corp.Capital = mathx.Amount(math.MaxInt64)
	state.Corp.Compute = mathx.Amount(math.MaxInt64)
	state.Corp.Assets = []game.CorpAsset{{ID: "extreme", Name: "Extreme", CapitalPerTick: math.MaxInt64, Operational: true}}
	m := New(state, sim.New(content.Catalog{SchemaVersion: 1}, 14), "")
	report := m.corpReportView()
	if !strings.Contains(report, "CAPITAL: "+state.Corp.Capital.String()) || !strings.Contains(report, "CAPITAL: 9223372036854775") {
		t.Fatalf("extreme capital projection was not rendered safely: %s", report)
	}
	if !strings.Contains(report, "COMPUTE: 9223372036854775") {
		t.Fatalf("extreme compute projection was not rendered safely: %s", report)
	}
}

func TestTargetEstimatesSaturateExtremeCatalogValues(t *testing.T) {
	target := content.Target{LootCredits: math.MaxInt64, LootData: math.MaxInt64, Layers: []content.Layer{{ICE: math.MaxInt64}, {ICE: math.MaxInt64}}}
	ice := targetICESummary(target, 0)
	loot := targetLootSummary(target, 0)
	if !strings.Contains(ice, "922337203685477") || !strings.Contains(loot, "922337203685477") {
		t.Fatalf("extreme target estimates were not rendered safely: ice=%q loot=%q", ice, loot)
	}
}

func TestCommitmentPreviewSaturatesExtremeBotCounts(t *testing.T) {
	maxBots := int64(math.MaxInt64)
	want := (maxBots/100)*75 + ((maxBots%100)*75)/100
	if got := commitmentBots(math.MaxInt64, 75); got != want {
		t.Fatalf("extreme commitment overflowed: got=%d want=%d", got, want)
	}
}

func TestPrestigeScreensExplainLockedAndReadyStates(t *testing.T) {
	corp := New(sim.NewCorporationState(11), sim.New(content.Catalog{SchemaVersion: 1}, 11), "")
	updated, _ := corp.Update(tea.KeyPressMsg(tea.Key{Text: "p", Code: 'p'}))
	corp = updated.(Model)
	if !strings.Contains(corp.View().Content, "STATUS: LOCKED // reach 1,000 valuation") {
		t.Fatalf("locked Corporation prestige state was not explained: %s", corp.View().Content)
	}
	corp.state.Corp.RunValuation = 1000
	corp.state.Corp.Assets = append(corp.state.Corp.Assets,
		game.CorpAsset{ID: "security_desk", Name: "Security Desk", Operational: true, Integrity: 100, MaxIntegrity: 100},
		game.CorpAsset{ID: "archive_server", Name: "Archive Server", Operational: true, Integrity: 100, MaxIntegrity: 100},
	)
	corp.state.Corp.ICE = corp.state.Corp.ICECapacity
	if !strings.Contains(corp.restructureView(), "STATUS: READY // [enter] confirms Restructuring") || !strings.Contains(corp.corpNextHint(), "[p] to Restructure") {
		t.Fatalf("ready Corporation prestige state was not explained: view=%s hint=%s", corp.restructureView(), corp.corpNextHint())
	}
	if !strings.Contains(corp.restructureView(), "PATENT SHARES: 0 -> 1") || !strings.Contains(corp.restructureView(), "LOST NOW:") {
		t.Fatalf("Corporation prestige confirmation omitted before/after values: %s", corp.restructureView())
	}

	runner := New(sim.NewState(12), sim.New(content.Catalog{SchemaVersion: 1}, 12), "")
	updated, _ = runner.Update(tea.KeyPressMsg(tea.Key{Text: "p", Code: 'p'}))
	runner = updated.(Model)
	if !strings.Contains(runner.View().Content, "STATUS: LOCKED // extract at least 10,000 data") {
		t.Fatalf("locked Runner prestige state was not explained: %s", runner.View().Content)
	}
	runner.state.RunExtractedData = 10000
	runner.state.GhostKeys = 2
	runner.state.PrestigeCount = 1
	view := runner.prestigeView()
	if !strings.Contains(view, "GHOST KEYS: 2 -> 3") || !strings.Contains(view, "LOST NOW:") || !strings.Contains(view, "NEXT RUN: 40 starter credits") {
		t.Fatalf("Runner prestige confirmation omitted before/after values: %s", view)
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
	state.Access = map[string]int64{"target": 25}
	m := New(state, sim.New(catalog, 2), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "t", Code: 't'}))
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "+", Code: '+'}))
	m = updated.(Model)
	if !strings.Contains(m.preflightView(), "COMMITMENT: 8 bots (100%)") || !strings.Contains(m.preflightView(), "ACCESS: access:25 ticks") || !strings.Contains(m.preflightView(), "BASELINE:") {
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
		Targets:       []content.Target{{ID: "target", Name: "Relay", Corporation: "Holloway Transit", Reachable: true, Layers: []content.Layer{{Integrity: 20, ICE: 1}}, LootCredits: 5, LootData: 3, ReconCost: 10, ReconTicks: 5, IntelGain: 40}},
	}
	state := sim.NewState(5)
	state.Upgrades = []string{"street_contacts"}
	state.Resources.Credits = mathx.New(20)
	m := New(state, sim.New(catalog, 5), "")
	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "t", Code: 't'}))
	m = updated.(Model)
	if !strings.Contains(m.targetsView(), "intel:0% (unknown)") || !strings.Contains(m.targetsView(), "ICE:1-2") || !strings.Contains(m.targetsView(), "reward:3-7 CR / 2-4 data") || !strings.Contains(m.targetsView(), "recon ready: 10 CR / 5 ticks [i]") {
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

func TestPreflightShowsIntelDrivenICESummary(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets:       []content.Target{{ID: "target", Name: "Relay", Reachable: true, Layers: []content.Layer{{Integrity: 20, ICE: 2}, {Integrity: 10, ICE: 3}}, LootCredits: 5, LootData: 3, TimeLimit: 30}},
	}
	state := sim.NewState(5)
	state.Intel = map[string]int64{"target": 100}
	state.Resources.Bots = 8
	state.Resources.BotCapacity = 8
	m := New(state, sim.New(catalog, 5), "")
	m.screen = "preflight"
	view := m.preflightView()
	if !strings.Contains(view, "ESTIMATED ICE: 5") || !strings.Contains(view, "ESTIMATED REWARD: 5 CR / 3 data") {
		t.Fatalf("preflight omitted exact intel-driven ICE estimate: %s", view)
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
	state.Access = map[string]int64{"target": 180}
	state.LastBattle = &game.BattleResult{TargetID: "target", Doctrine: game.Balanced, Outcome: "victory", CommittedBots: 8, Survivors: 6, BotsLost: 2, TraceBP: 1200, HeatAddedBP: 700, HeatAfterBP: 1900, CreditsLooted: 40, DataLooted: 10, ElapsedTicks: 12}
	m := New(state, sim.New(catalog, 4), "")
	m.screen = "result"
	view := m.View().Content
	for _, expected := range []string{"OUTCOME: victory", "REASON: all security layers breached", "BOTS: 8 committed / 6 survived / 2 lost", "HEAT: +7% (total 19%)", "LOOT: 40 credits / 10 data", "ACCESS: access:180 ticks", "review the archive"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("result view omitted %q: %s", expected, view)
		}
	}
}

func TestBattleResultExplainsCorporationIsolation(t *testing.T) {
	catalog := content.Catalog{SchemaVersion: 1, Targets: []content.Target{{ID: "target", Name: "Relay"}}}
	state := sim.NewCorporationState(19)
	state.LastBattle = &game.BattleResult{TargetID: "target", Outcome: "isolated", CommittedBots: 8, Survivors: 8, ElapsedTicks: 4}
	m := New(state, sim.New(catalog, 19), "")
	m.screen = "result"
	view := m.View().Content
	for _, expected := range []string{"OUTCOME: isolated", "REASON: the Corporation isolated the asset", "repair it when integrity"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("isolated result omitted %q: %s", expected, view)
		}
	}
}
