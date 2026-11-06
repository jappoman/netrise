package playtest

import (
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func loadCatalog(t *testing.T) content.Catalog {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestRunnerAndCorporationPlaytestScriptsAreDeterministic(t *testing.T) {
	catalog := loadCatalog(t)
	for _, faction := range []game.Faction{game.Runner, game.Corporation} {
		first, err := Run(catalog, faction, 17)
		if err != nil {
			t.Fatalf("%s playtest failed: %v", faction, err)
		}
		second, err := Run(catalog, faction, 17)
		if err != nil {
			t.Fatalf("%s replay failed: %v", faction, err)
		}
		if !reflect.DeepEqual(first, second) {
			t.Fatalf("%s playtest was not deterministic: first=%#v second=%#v", faction, first, second)
		}
		if first.Events == 0 || first.LogicalTick <= 0 {
			t.Fatalf("%s playtest made no progress: %#v", faction, first)
		}
		if first.BattleStartedAt <= 0 || first.BattleEndedAt < first.BattleStartedAt {
			t.Fatalf("%s playtest did not expose ordered battle timing: %#v", faction, first)
		}
		if faction == game.Runner && first.ReconCompletedAt <= 0 {
			t.Fatalf("Runner playtest did not expose recon timing: %#v", first)
		}
	}
}
