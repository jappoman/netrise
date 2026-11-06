package main

import (
	tea "charm.land/bubbletea/v2"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/playtest"
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
	if err := validateTopLevelArgs(args); err != nil {
		fatal(err)
	}
	if len(args) > 0 && args[0] == "simulate" {
		simulate(args[1:])
		return
	}
	if len(args) > 0 && args[0] == "playtest" {
		playtestCommand(args[1:])
		return
	}
	if len(args) > 0 && args[0] == "validate-content" {
		if _, err := loadCatalog(); err != nil {
			fatal(err)
		}
		fmt.Println("content: valid")
		return
	}
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "corp" && len(args) > 1 && (args[1] == "--help" || args[1] == "-h")) {
		fmt.Println(topLevelUsage())
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
	lock, err := save.AcquireSessionLock(dataDir, faction)
	if err != nil {
		fatal(fmt.Errorf("open campaign: %w", err))
	}
	defer lock.Close()
	s := loadState(dataDir, e, faction)
	p := tea.NewProgram(tui.New(s, e, dataDir))
	if _, err = p.Run(); err != nil {
		fatal(err)
	}
}

func topLevelUsage() string {
	return "netrise [--data-dir DIR]\nnetrise corp [--data-dir DIR]\nnetrise simulate --faction runner|corp --seed N --ticks N\nnetrise playtest --faction runner|corp --seed N\nnetrise validate-content\n\nA deterministic fictional Runner/Corporation strategy game."
}

func validateTopLevelArgs(args []string) error {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "simulate", "playtest":
		return nil
	case "corp":
		if len(args) > 1 && args[1] != "--help" && args[1] != "-h" {
			return fmt.Errorf("corp does not accept %q (use --help)", args[1])
		}
		if len(args) > 2 {
			return fmt.Errorf("corp help accepts no additional arguments")
		}
		return nil
	case "--help", "-h", "--version":
		if len(args) > 1 {
			return fmt.Errorf("%s accepts no additional arguments", args[0])
		}
		return nil
	case "validate-content":
		if len(args) > 1 {
			return fmt.Errorf("validate-content does not accept arguments")
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q (use --help)", args[0])
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
	return loadStateAt(dataDir, e, faction, time.Now)
}

func loadStateAt(dataDir string, e *sim.Engine, faction game.Faction, now func() time.Time) game.State {
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
	current := now()
	elapsed := int64(current.Sub(savedAt).Seconds())
	if elapsed <= 0 {
		return s
	}
	reconciled, _ := e.ReconcileOffline(s, elapsed)
	reconciled.SavedAtUTC = current.UTC().Format(time.RFC3339Nano)
	if err := save.Write(dataDir, reconciled); err != nil {
		fatal(fmt.Errorf("persist offline reconciliation: %w", err))
	}
	return reconciled
}
func simulate(args []string) {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	seed := fs.Int64("seed", 1, "world seed")
	ticks := fs.Int64("ticks", 0, "logical seconds")
	faction := fs.String("faction", "runner", "campaign faction: runner or corp")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return
		}
		fatal(err)
	}
	if err := rejectTrailingArgs("simulate", fs.Args()); err != nil {
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

func playtestCommand(args []string) {
	fs := flag.NewFlagSet("playtest", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	seed := fs.Int64("seed", 17, "world seed")
	faction := fs.String("faction", "runner", "campaign faction: runner or corp")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return
		}
		fatal(err)
	}
	if err := rejectTrailingArgs("playtest", fs.Args()); err != nil {
		fatal(err)
	}
	var selected game.Faction
	switch *faction {
	case "runner":
		selected = game.Runner
	case "corp", "corporation":
		selected = game.Corporation
	default:
		fatal(fmt.Errorf("unknown faction %q (use runner or corp)", *faction))
	}
	catalog, err := loadCatalog()
	if err != nil {
		fatal(err)
	}
	report, err := playtest.Run(catalog, selected, *seed)
	if err != nil {
		fatal(err)
	}
	b, err := json.Marshal(report)
	if err != nil {
		fatal(err)
	}
	fmt.Println(string(b))
}

func rejectTrailingArgs(command string, args []string) error {
	if len(args) == 0 {
		return nil
	}
	return fmt.Errorf("%s does not accept positional arguments: %q", command, args[0])
}

func loadCatalog() (content.Catalog, error) {
	paths := []string{"assets/content/catalog.json"}
	if executable, err := os.Executable(); err == nil {
		paths = append(paths, catalogBundlePaths(executable)...)
	}
	var lastErr error
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			lastErr = err
			continue
		}
		catalog, loadErr := content.Load(f)
		_ = f.Close()
		if loadErr != nil {
			return content.Catalog{}, fmt.Errorf("load content %s: %w", path, loadErr)
		}
		return catalog, nil
	}
	return content.Catalog{}, fmt.Errorf("open content catalog from working directory or executable bundle: %w", lastErr)
}

func catalogBundlePaths(executable string) []string {
	executableDir := filepath.Dir(executable)
	return []string{
		filepath.Join(executableDir, "assets", "content", "catalog.json"),
		filepath.Join(executableDir, "catalog.json"),
	}
}
func fatal(err error) { fmt.Fprintln(os.Stderr, "netrise:", err); os.Exit(1) }
