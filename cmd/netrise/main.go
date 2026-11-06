package main

import (
	tea "charm.land/bubbletea/v2"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/save"
	"github.com/jappoman/netrise/internal/sim"
	"github.com/jappoman/netrise/internal/tui"
	"os"
	"path/filepath"
	"time"
)

const version = "0.1.0"

func main() {
	args, dataDir := extractDataDir(os.Args[1:])
	if len(args) > 0 && args[0] == "simulate" {
		simulate(args[1:])
		return
	}
	if len(args) > 0 && args[0] == "validate-content" {
		if _, err := loadCatalog(); err != nil {
			fatal(err)
		}
		fmt.Println("content: valid")
		return
	}
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Println("netrise [--data-dir DIR]\nnetrise corp [--data-dir DIR]\nnetrise simulate --faction runner|corp --seed N --ticks N\nnetrise validate-content\n\nA deterministic fictional Runner/Corporation strategy game.")
		return
	}
	if len(args) > 0 && args[0] == "--version" {
		fmt.Println(version)
		return
	}
	c, err := loadCatalog()
	if err != nil {
		fatal(err)
	}
	e := sim.New(c, 1)
	if dataDir == "" {
		configDir, configErr := os.UserConfigDir()
		if configErr != nil {
			fatal(fmt.Errorf("find user config directory: %w", configErr))
		}
		dataDir = filepath.Join(configDir, "netrise")
	}
	faction := game.Runner
	if len(args) > 0 && args[0] == "corp" {
		faction = game.Corporation
	}
	s := loadState(dataDir, e, faction)
	p := tea.NewProgram(tui.New(s, e, dataDir))
	if _, err = p.Run(); err != nil {
		fatal(err)
	}
}

func extractDataDir(args []string) ([]string, string) {
	filtered := make([]string, 0, len(args))
	dataDir := ""
	for i := 0; i < len(args); i++ {
		if args[i] != "--data-dir" {
			filtered = append(filtered, args[i])
			continue
		}
		if i+1 >= len(args) || args[i+1] == "" {
			fatal(fmt.Errorf("--data-dir requires a directory"))
		}
		dataDir = args[i+1]
		i++
	}
	return filtered, dataDir
}

func loadState(dataDir string, e *sim.Engine, faction game.Faction) game.State {
	s, err := save.ReadFaction(dataDir, faction)
	if err != nil {
		if !os.IsNotExist(err) {
			fatal(fmt.Errorf("load save: %w", err))
		}
		if faction == game.Corporation {
			return sim.NewCorporationState(1)
		}
		return sim.NewState(1)
	}
	if s.SavedAtUTC == "" {
		return s
	}
	savedAt, err := time.Parse(time.RFC3339Nano, s.SavedAtUTC)
	if err != nil {
		return s
	}
	elapsed := int64(time.Since(savedAt).Seconds())
	if elapsed <= 0 {
		return s
	}
	reconciled, _ := e.ReconcileOffline(s, elapsed)
	return reconciled
}
func simulate(args []string) {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	seed := fs.Int64("seed", 1, "world seed")
	ticks := fs.Int64("ticks", 0, "logical seconds")
	faction := fs.String("faction", "runner", "campaign faction: runner or corp")
	if err := fs.Parse(args); err != nil {
		fatal(err)
	}
	if *ticks < 0 {
		fatal(fmt.Errorf("ticks cannot be negative"))
	}
	c, err := loadCatalog()
	if err != nil {
		fatal(err)
	}
	var s game.State
	switch *faction {
	case "runner":
		s = sim.NewState(*seed)
	case "corp", "corporation":
		s = sim.NewCorporationState(*seed)
	default:
		fatal(fmt.Errorf("unknown faction %q (use runner or corp)", *faction))
	}
	s = sim.New(c, *seed).Advance(s, *ticks)
	out := struct {
		Seed  int64      `json:"seed"`
		Ticks int64      `json:"ticks"`
		State game.State `json:"state"`
	}{*seed, *ticks, s}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}
func loadCatalog() (content.Catalog, error) {
	f, err := os.Open("assets/content/catalog.json")
	if err != nil {
		return content.Catalog{}, fmt.Errorf("open content: %w", err)
	}
	defer f.Close()
	return content.Load(f)
}
func fatal(err error) { fmt.Fprintln(os.Stderr, "netrise:", err); os.Exit(1) }
