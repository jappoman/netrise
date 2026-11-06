package save

import (
	"github.com/jappoman/netrise/internal/game"
	"os"
	"testing"
)

func TestRoundTripAndCorruption(t *testing.T) {
	d := t.TempDir()
	s := game.State{SchemaVersion: 1, Faction: game.Runner}
	if err := Write(d, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read(d)
	if err != nil || got.SchemaVersion != 1 {
		t.Fatal(got, err)
	}
	if err := os.WriteFile(Path(d), []byte("{"), 0600); err == nil {
		_, err = Read(d)
		if err == nil {
			t.Fatal("corruption accepted")
		}
	}
}

func TestRoundTripPreservesProgressionState(t *testing.T) {
	d := t.TempDir()
	s := game.State{SchemaVersion: game.CurrentSchema, Faction: game.Runner, Research: []string{"hardware_01"}, RunExtractedData: 120, GhostKeys: 2, Automation: game.Automation{Contracts: true}, Intel: map[string]int64{"t_01": 45}}
	if err := Write(d, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read(d)
	if err != nil || got.RunExtractedData != 120 || got.GhostKeys != 2 || !got.Automation.Contracts || got.Intel["t_01"] != 45 || len(got.Research) != 1 {
		t.Fatalf("progression state was not preserved: %#v, %v", got, err)
	}
}
func TestMigrate(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(Path(d), []byte(`{"faction":"runner"}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Read(d)
	if err != nil || s.SchemaVersion != 1 {
		t.Fatal(s, err)
	}
}
