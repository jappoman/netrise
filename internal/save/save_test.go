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
