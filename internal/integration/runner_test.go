package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"github.com/jappoman/netrise/internal/sim"
)

func TestFreshRunnerCatalogVerticalSlice(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Targets) != 5 {
		t.Fatalf("expected five targets, got %d", len(catalog.Targets))
	}

	e := sim.New(catalog, 17)
	s := sim.NewState(17)
	apply := func(command game.Command) {
		var commandErr error
		s, _, commandErr = e.Apply(s, command)
		if commandErr != nil {
			t.Fatal(commandErr)
		}
	}
	for i := 0; i < 13; i++ {
		apply(game.Command{Kind: game.CollectContract})
	}
	apply(game.Command{Kind: game.BuyUpgrade, ID: "street_contacts"})
	for i := 0; i < 7; i++ {
		apply(game.Command{Kind: game.CollectContract})
	}
	apply(game.Command{Kind: game.BuyUpgrade, ID: "secondhand_compiler"})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 4})
	apply(game.Command{Kind: game.CollectContract})
	apply(game.Command{Kind: game.StartRecon, ID: "t_01"})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 15})
	if s.Intel["t_01"] != 45 || s.Recon != nil {
		t.Fatalf("recon did not complete: intel=%v recon=%#v", s.Intel, s.Recon)
	}
	apply(game.Command{Kind: game.CollectContract})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 5})
	apply(game.Command{Kind: game.QueueFabrication, Quantity: 5})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 3})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 25})
	apply(game.Command{Kind: game.CollectContract})
	apply(game.Command{Kind: game.CollectContract})
	apply(game.Command{Kind: game.QueueFabrication, Quantity: 5})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 3})
	apply(game.Command{Kind: game.StartBattle, ID: "t_01", Quantity: 10, Doctrine: game.Balanced})
	for s.ActiveBattle != nil {
		apply(game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	}
	if s.LastBattle == nil || s.LastBattle.Outcome != "victory" {
		t.Fatalf("fresh runner did not complete first breach: %#v", s.LastBattle)
	}
	if len(s.CompletedTargets) != 1 || s.CompletedTargets[0] != "t_01" {
		t.Fatalf("first route did not unlock: %v", s.CompletedTargets)
	}
	apply(game.Command{Kind: game.BuyResearch, ID: "hardware_01"})
	if len(s.Research) != 1 || s.Research[0] != "hardware_01" {
		t.Fatalf("first research node was not completed: %v", s.Research)
	}
	s.RunExtractedData = 10000
	apply(game.Command{Kind: game.IdentityBurn})
	if s.PrestigeCount != 1 || s.GhostKeys != 1 || s.RunExtractedData != 0 || len(s.CompletedTargets) != 0 || len(s.Research) != 0 {
		t.Fatalf("identity burn did not reset the run atomically: %#v", s)
	}
	if s.Resources.Credits != mathx.New(30) {
		t.Fatalf("identity burn starter boost incorrect: %s", s.Resources.Credits)
	}
	if _, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "t_02", Quantity: 1, Doctrine: game.Stealth}); err == nil {
		// The reset intentionally removes route access; the target must be
		// rediscovered in the new run rather than remaining unlocked for free.
		t.Fatal("identity burn retained route access")
	}
}

func TestCombatCommitmentMatrixDoesNotRewardLessPower(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	target := catalog.Targets[0]
	wins := func(commitment int64) int {
		count := 0
		for seed := int64(1); seed <= 20; seed++ {
			e := sim.New(catalog, seed)
			s := sim.NewState(seed)
			s.Upgrades = []string{"secondhand_compiler"}
			s.Resources.Bots = commitment
			var commandErr error
			s, _, commandErr = e.Apply(s, game.Command{Kind: game.StartBattle, ID: target.ID, Quantity: commitment, Doctrine: game.Balanced})
			if commandErr != nil {
				t.Fatal(commandErr)
			}
			for tick := int64(0); s.ActiveBattle != nil && tick <= target.TimeLimit; tick++ {
				s, _, commandErr = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
				if commandErr != nil {
					t.Fatal(commandErr)
				}
			}
			if s.ActiveBattle != nil {
				t.Fatalf("battle exceeded time limit for seed %d and commitment %d", seed, commitment)
			}
			if s.LastBattle != nil && s.LastBattle.Outcome == "victory" {
				count++
			}
		}
		return count
	}
	low, high := wins(5), wins(15)
	if high < low {
		t.Fatalf("more committed bots reduced wins: low=%d high=%d", low, high)
	}
}
