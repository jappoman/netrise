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
func PathForFaction(dir string, faction game.Faction) string {
	if faction == game.Corporation {
		return filepath.Join(dir, "corporation.json")
	}
	return Path(dir)
}
func Write(dir string, s game.State) error {
	return writeAt(dir, PathForFaction(dir, s.Faction), s)
}
func writeAt(dir, path string, s game.State) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
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
	if _, err = os.Stat(path); err == nil {
		_ = os.Rename(path, path+".bak")
	}
	return os.Rename(tmp, path)
}
func Read(dir string) (game.State, error) {
	return ReadFaction(dir, game.Runner)
}
func ReadFaction(dir string, faction game.Faction) (game.State, error) {
	path := PathForFaction(dir, faction)
	b, err := os.ReadFile(path)
	if err != nil {
		return game.State{}, err
	}
	var s game.State
	if err = json.Unmarshal(b, &s); err != nil {
		backup, backupErr := os.ReadFile(path + ".bak")
		if backupErr == nil {
			var recovered game.State
			if recoverErr := json.Unmarshal(backup, &recovered); recoverErr == nil && recovered.SchemaVersion == game.CurrentSchema {
				return recovered, nil
			}
		}
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
