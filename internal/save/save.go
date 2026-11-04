package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jappoman/netrise/internal/game"
	"os"
	"path/filepath"
)

var ErrCorrupt = errors.New("save is corrupt")

func Path(dir string) string { return filepath.Join(dir, "runner.json") }
func Write(dir string, s game.State) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := Path(dir) + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	if _, err = os.Stat(Path(dir)); err == nil {
		_ = os.Rename(Path(dir), Path(dir)+".bak")
	}
	return os.Rename(tmp, Path(dir))
}
func Read(dir string) (game.State, error) {
	b, err := os.ReadFile(Path(dir))
	if err != nil {
		return game.State{}, err
	}
	var s game.State
	if err = json.Unmarshal(b, &s); err != nil {
		return s, fmt.Errorf("%w: %v", ErrCorrupt, err)
	}
	if s.SchemaVersion == 0 {
		return migrateV0(s)
	}
	if s.SchemaVersion != game.CurrentSchema {
		return s, fmt.Errorf("unsupported save schema %d", s.SchemaVersion)
	}
	return s, nil
}
func migrateV0(s game.State) (game.State, error) {
	s.SchemaVersion = game.CurrentSchema
	if s.Faction == "" {
		s.Faction = game.Runner
	}
	return s, nil
}
