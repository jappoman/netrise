package main

import (
	tea "charm.land/bubbletea/v2"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/sim"
	"github.com/jappoman/netrise/internal/tui"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--data-dir" {
		if len(os.Args) < 3 {
			fatal(fmt.Errorf("--data-dir requires a directory"))
		}
		// Accepted for portable builds; interactive save wiring follows with
		// the complete timestamp reconciliation flow.
		os.Args = append([]string{os.Args[0]}, os.Args[3:]...)
	}
	if len(os.Args) > 1 && os.Args[1] == "simulate" {
		simulate(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("netrise [--data-dir DIR]\nnetrise simulate --seed N --ticks N\n\nA deterministic fictional Runner strategy game.")
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}
	c, err := loadCatalog()
	if err != nil {
		fatal(err)
	}
	e := sim.New(c, 1)
	p := tea.NewProgram(tui.New(sim.NewState(1), e))
	if _, err = p.Run(); err != nil {
		fatal(err)
	}
}
func simulate(args []string) {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	seed := fs.Int64("seed", 1, "world seed")
	ticks := fs.Int64("ticks", 0, "logical seconds")
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
	s := sim.NewState(*seed)
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
