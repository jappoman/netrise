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
