package main

import (
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/save"
	"github.com/jappoman/netrise/internal/sim"
	"path/filepath"
	"testing"
	"time"
)

func TestCatalogBundlePathsAreExecutableRelative(t *testing.T) {
	paths := catalogBundlePaths(filepath.Join("release", "netrise.exe"))
	want := []string{
		filepath.Join("release", "assets", "content", "catalog.json"),
		filepath.Join("release", "catalog.json"),
	}
	if len(paths) != len(want) {
		t.Fatalf("candidate count = %d, want %d", len(paths), len(want))
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("candidate %d = %q, want %q", i, paths[i], want[i])
		}
	}
}

func TestValidateTopLevelArgsRejectsUnknownCommands(t *testing.T) {
	if err := validateTopLevelArgs([]string{"unknown"}); err == nil {
		t.Fatal("unknown top-level command was accepted")
	}
	if err := validateTopLevelArgs([]string{"validate-content", "extra"}); err == nil {
		t.Fatal("validate-content accepted an unexpected argument")
	}
	if err := validateTopLevelArgs([]string{"corp", "--unexpected"}); err == nil {
		t.Fatal("Corporation command accepted an unexpected argument")
	}
	for _, args := range [][]string{{}, {"corp"}, {"simulate", "--help"}, {"playtest", "--help"}, {"validate-content"}, {"--help"}} {
		if err := validateTopLevelArgs(args); err != nil {
			t.Fatalf("valid top-level args rejected: %v: %v", args, err)
		}
	}
}

func TestLoadStateUsesInjectedClockForOfflineReconciliation(t *testing.T) {
	dir := t.TempDir()
	saved := time.Date(2026, time.November, 6, 10, 0, 0, 0, time.UTC)
	state := sim.NewState(17)
	state.SavedAtUTC = saved.Format(time.RFC3339Nano)
	if err := save.Write(dir, state); err != nil {
		t.Fatal(err)
	}
	e := sim.New(content.Catalog{SchemaVersion: 1}, 17)
	got := loadStateAt(dir, e, game.Runner, func() time.Time { return saved.Add(90 * time.Second) })
	if got.LogicalTick != 90 || got.OfflineSummary == nil || got.OfflineSummary.Ticks != 90 {
		t.Fatalf("injected clock did not reconcile expected ticks: %#v", got)
	}
}

func TestLoadStateDoesNotAdvanceOnClockReversal(t *testing.T) {
	dir := t.TempDir()
	saved := time.Date(2026, time.November, 6, 10, 0, 0, 0, time.UTC)
	state := sim.NewState(17)
	state.LogicalTick = 12
	state.SavedAtUTC = saved.Format(time.RFC3339Nano)
	if err := save.Write(dir, state); err != nil {
		t.Fatal(err)
	}
	e := sim.New(content.Catalog{SchemaVersion: 1}, 17)
	got := loadStateAt(dir, e, game.Runner, func() time.Time { return saved.Add(-time.Second) })
	if got.LogicalTick != 12 {
		t.Fatalf("clock reversal advanced state: %#v", got)
	}
}

func TestLoadStatePersistsOfflineCapWarning(t *testing.T) {
	dir := t.TempDir()
	saved := time.Date(2026, time.November, 6, 10, 0, 0, 0, time.UTC)
	state := sim.NewState(18)
	state.SavedAtUTC = saved.Format(time.RFC3339Nano)
	if err := save.Write(dir, state); err != nil {
		t.Fatal(err)
	}
	e := sim.New(content.Catalog{SchemaVersion: 1}, 18)
	got := loadStateAt(dir, e, game.Runner, func() time.Time {
		return saved.Add(time.Duration(sim.OfflineCapTicks+100) * time.Second)
	})
	if got.OfflineSummary == nil || !got.OfflineSummary.Capped || got.OfflineSummary.Ticks != sim.OfflineCapTicks {
		t.Fatalf("offline startup did not preserve cap warning: %#v", got.OfflineSummary)
	}
}

func TestLoadStatePersistsOfflineReconciliationWithoutDoubleGrant(t *testing.T) {
	dir := t.TempDir()
	saved := time.Date(2026, time.November, 6, 10, 0, 0, 0, time.UTC)
	state := sim.NewState(19)
	state.SavedAtUTC = saved.Format(time.RFC3339Nano)
	if err := save.Write(dir, state); err != nil {
		t.Fatal(err)
	}
	e := sim.New(content.Catalog{SchemaVersion: 1}, 19)
	now := saved.Add(90 * time.Second)
	first := loadStateAt(dir, e, game.Runner, func() time.Time { return now })
	if first.LogicalTick != 90 || first.SavedAtUTC != now.Format(time.RFC3339Nano) {
		t.Fatalf("offline reconciliation was not persisted: %#v", first)
	}
	persisted, err := save.ReadFaction(dir, game.Runner)
	if err != nil {
		t.Fatal(err)
	}
	second := loadStateAt(dir, e, game.Runner, func() time.Time { return now })
	if second.LogicalTick != persisted.LogicalTick || second.LogicalTick != 90 {
		t.Fatalf("reopening at the same timestamp double-granted offline time: persisted=%d reopened=%d", persisted.LogicalTick, second.LogicalTick)
	}
}

func TestCorporationOfflineReconciliationPersistsWithoutDoubleGrant(t *testing.T) {
	dir := t.TempDir()
	saved := time.Date(2026, time.November, 6, 10, 0, 0, 0, time.UTC)
	state := sim.NewCorporationState(20)
	state.SavedAtUTC = saved.Format(time.RFC3339Nano)
	if err := save.Write(dir, state); err != nil {
		t.Fatal(err)
	}
	e := sim.New(content.Catalog{SchemaVersion: 1}, 20)
	now := saved.Add(90 * time.Second)
	first := loadStateAt(dir, e, game.Corporation, func() time.Time { return now })
	if first.LogicalTick != 90 || first.SavedAtUTC != now.Format(time.RFC3339Nano) || first.Corp == nil {
		t.Fatalf("Corporation offline reconciliation was not persisted: %#v", first)
	}
	second := loadStateAt(dir, e, game.Corporation, func() time.Time { return now })
	if second.LogicalTick != 90 || second.Corp == nil || second.Corp.Capital != first.Corp.Capital {
		t.Fatalf("Corporation reopening double-granted offline time: first=%#v second=%#v", first.Corp, second.Corp)
	}
}
