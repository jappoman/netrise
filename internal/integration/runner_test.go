package integration

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"github.com/jappoman/netrise/internal/save"
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
	if len(s.EventLog) == 0 {
		t.Fatal("runner E2E produced no event archive")
	}
	if len(s.CompletedTargets) != 1 || s.CompletedTargets[0] != "t_01" {
		t.Fatalf("first route did not unlock: %v", s.CompletedTargets)
	}
	if s.TargetCooldowns["t_01"] != 60 {
		t.Fatalf("first target cooldown was not persisted: %v", s.TargetCooldowns)
	}
	if s.Access["t_01"] != 180 {
		t.Fatalf("first breach did not grant catalogued temporary access: %v", s.Access)
	}
	if s.Resources.Bots < 1 {
		t.Fatal("first breach left no reserve bot for the adjacent route check")
	}
	apply(game.Command{Kind: game.StartBattle, ID: "t_02", Quantity: 1, Doctrine: game.Stealth})
	if s.ActiveBattle == nil {
		t.Fatal("temporary access did not unlock the adjacent catalog target")
	}
	apply(game.Command{Kind: game.WithdrawBattle})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 180})
	if s.Access["t_01"] != 0 {
		t.Fatalf("active ticks did not expire first route access: %v", s.Access)
	}
	if _, _, err := e.Apply(s, game.Command{Kind: game.StartBattle, ID: "t_02", Quantity: 1, Doctrine: game.Stealth}); err == nil {
		t.Fatal("expired catalogued access still unlocked the adjacent target")
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

func TestActiveBattleSaveReopenResumesIdentically(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	e := sim.New(catalog, 311)
	original := sim.NewState(311)
	original.Upgrades = []string{"secondhand_compiler"}
	original.Resources.Bots = 12
	original, _, err = e.Apply(original, game.Command{Kind: game.StartBattle, ID: "t_01", Quantity: 10, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	original, _, err = e.Apply(original, game.Command{Kind: game.AdvanceTicks, Ticks: 3})
	if err != nil || original.ActiveBattle == nil {
		t.Fatalf("battle did not remain resumable after setup: state=%#v err=%v", original, err)
	}
	dir := t.TempDir()
	if err := save.Write(dir, original); err != nil {
		t.Fatal(err)
	}
	resumed, err := save.Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	var originalEvents, resumedEvents []game.Event
	for original.ActiveBattle != nil || resumed.ActiveBattle != nil {
		if original.ActiveBattle == nil || resumed.ActiveBattle == nil {
			t.Fatalf("save/reopen diverged on battle lifetime: original=%#v resumed=%#v", original.ActiveBattle, resumed.ActiveBattle)
		}
		var events []game.Event
		original, events, err = e.Apply(original, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			t.Fatal(err)
		}
		originalEvents = append(originalEvents, events...)
		resumed, events, err = e.Apply(resumed, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			t.Fatal(err)
		}
		resumedEvents = append(resumedEvents, events...)
	}
	if !reflect.DeepEqual(original, resumed) || !reflect.DeepEqual(originalEvents, resumedEvents) {
		t.Fatalf("save/reopen changed deterministic result:\noriginal=%#v\nresumed=%#v\noriginal events=%#v\nresumed events=%#v", original, resumed, originalEvents, resumedEvents)
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

func TestDoctrineMatrixPreservesDistinctHeatAndAttrition(t *testing.T) {
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
	doctrines := []game.Doctrine{game.Stealth, game.Balanced, game.Overload}
	var heat [3]int64
	var losses [3]int64
	for seed := int64(1); seed <= 100; seed++ {
		for i, doctrine := range doctrines {
			e := sim.New(catalog, seed)
			s := sim.NewState(seed)
			s.Upgrades = []string{"secondhand_compiler"}
			s.Resources.Bots = 20
			s, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: target.ID, Quantity: 20, Doctrine: doctrine})
			if err != nil {
				t.Fatalf("start %s seed %d: %v", doctrine, seed, err)
			}
			for ticks := int64(0); s.ActiveBattle != nil && ticks <= target.TimeLimit; ticks++ {
				s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
				if err != nil {
					t.Fatalf("advance %s seed %d: %v", doctrine, seed, err)
				}
			}
			if s.LastBattle == nil {
				t.Fatalf("missing %s result for seed %d", doctrine, seed)
			}
			heat[i] += s.LastBattle.HeatAddedBP
			losses[i] += s.LastBattle.BotsLost
		}
	}
	if !(heat[0] < heat[1] && heat[1] < heat[2]) {
		t.Fatalf("doctrine heat profiles collapsed: stealth=%d balanced=%d overload=%d", heat[0], heat[1], heat[2])
	}
	if !(losses[0] < losses[1] && losses[2] >= losses[1]) {
		t.Fatalf("doctrine attrition profiles collapsed: stealth=%d balanced=%d overload=%d", losses[0], losses[1], losses[2])
	}
}

func TestBattleBalanceSweepResolvesEveryTargetAndDoctrine(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	// Route progression is covered by the vertical-slice tests. This sweep
	// isolates combat balance by making every catalog target reachable.
	for i := range catalog.Targets {
		catalog.Targets[i].Reachable = true
	}

	doctrines := []game.Doctrine{game.Balanced, game.Stealth, game.Overload}
	commitments := []int64{10, 20, 30, 40}
	for _, target := range catalog.Targets {
		for _, doctrine := range doctrines {
			for _, commitment := range commitments {
				for seed := int64(1); seed <= 100; seed++ {
					e := sim.New(catalog, seed)
					s := sim.NewState(seed)
					s.Upgrades = []string{"secondhand_compiler"}
					s.Resources.Bots = commitment
					s, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: target.ID, Quantity: commitment, Doctrine: doctrine})
					if err != nil {
						t.Fatalf("start %s/%s/%d/seed%d: %v", target.ID, doctrine, commitment, seed, err)
					}
					for ticks := int64(0); s.ActiveBattle != nil && ticks <= target.TimeLimit; ticks++ {
						s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
						if err != nil {
							t.Fatalf("advance %s/%s/%d/seed%d: %v", target.ID, doctrine, commitment, seed, err)
						}
					}
					if s.ActiveBattle != nil {
						t.Fatalf("battle exceeded time limit: target=%s doctrine=%s commitment=%d seed=%d active=%#v", target.ID, doctrine, commitment, seed, s.ActiveBattle)
					}
					if s.LastBattle == nil {
						t.Fatalf("missing result: target=%s doctrine=%s commitment=%d seed=%d", target.ID, doctrine, commitment, seed)
					}
					result := s.LastBattle
					if result.TargetID != target.ID || result.Doctrine != doctrine || result.CommittedBots != commitment {
						t.Fatalf("result identity mismatch: target=%s doctrine=%s commitment=%d seed=%d result=%#v", target.ID, doctrine, commitment, seed, result)
					}
					if result.Outcome != "victory" && result.Outcome != "defeat" && result.Outcome != "timeout" {
						t.Fatalf("invalid runner outcome: target=%s doctrine=%s commitment=%d seed=%d outcome=%q", target.ID, doctrine, commitment, seed, result.Outcome)
					}
					if result.ElapsedTicks < 1 || result.ElapsedTicks > target.TimeLimit {
						t.Fatalf("invalid battle duration: target=%s doctrine=%s commitment=%d seed=%d ticks=%d limit=%d", target.ID, doctrine, commitment, seed, result.ElapsedTicks, target.TimeLimit)
					}
					if result.Survivors < 0 || result.BotsLost < 0 || result.Survivors+result.BotsLost != result.CommittedBots {
						t.Fatalf("bot accounting mismatch: target=%s doctrine=%s commitment=%d seed=%d result=%#v", target.ID, doctrine, commitment, seed, result)
					}
					if result.HeatBeforeBP < 0 || result.HeatAfterBP < result.HeatBeforeBP || result.HeatAfterBP > 10000 || result.HeatAddedBP < 0 {
						t.Fatalf("heat invariant violated: target=%s doctrine=%s commitment=%d seed=%d result=%#v", target.ID, doctrine, commitment, seed, result)
					}
					if s.Resources.Bots < 0 || s.Resources.Credits < 0 || s.Resources.Compute < 0 || s.Resources.Data < 0 {
						t.Fatalf("negative resource after battle: target=%s doctrine=%s commitment=%d seed=%d resources=%#v", target.ID, doctrine, commitment, seed, s.Resources)
					}
				}
			}
		}
	}
}

func TestRunnerSixtyMinuteScriptedProgressionRemainsBounded(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	e := sim.New(catalog, 61)
	s := sim.NewState(61)
	s.Resources.Credits = mathx.New(1000)
	s.Resources.Compute = mathx.New(1000)
	s.Resources.BotCapacity = 20
	s.Upgrades = []string{"secondhand_compiler", "street_contacts", "autodialer", "packet_scheduler", "signal_profiler"}
	s.Automation = game.Automation{Contracts: true, Fabrication: true, ReconTargetID: "t_01"}
	var commandErr error
	s, _, commandErr = e.Apply(s, game.Command{Kind: game.StartRecon, ID: "t_01"})
	if commandErr != nil {
		t.Fatal(commandErr)
	}
	s, _, commandErr = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 3600})
	if commandErr != nil {
		t.Fatal(commandErr)
	}
	if s.LogicalTick != 3600 || len(s.EventLog) > sim.EventLogLimit || s.Resources.Credits < 0 || s.Resources.Compute < 0 || s.Resources.Bots < 0 {
		t.Fatalf("60-minute script violated bounded progression: tick=%d events=%d resources=%#v", s.LogicalTick, len(s.EventLog), s.Resources)
	}
	if s.Intel["t_01"] == 0 {
		t.Fatal("60-minute script produced no recon progress")
	}
}

func TestRunnerTwoDistinctStrategiesCompleteFirstBreach(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	strategies := []struct {
		name     string
		doctrine game.Doctrine
		bots     int64
		seed     int64
	}{
		{name: "stealth conservation", doctrine: game.Stealth, bots: 10, seed: 117},
		{name: "overload pressure", doctrine: game.Overload, bots: 12, seed: 118},
	}
	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			e := sim.New(catalog, strategy.seed)
			s := sim.NewState(strategy.seed)
			s.Upgrades = []string{"secondhand_compiler"}
			s.Resources.Bots = strategy.bots
			var commandErr error
			s, _, commandErr = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "t_01", Quantity: strategy.bots, Doctrine: strategy.doctrine})
			if commandErr != nil {
				t.Fatal(commandErr)
			}
			for ticks := int64(0); s.ActiveBattle != nil && ticks <= 30; ticks++ {
				s, _, commandErr = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
				if commandErr != nil {
					t.Fatal(commandErr)
				}
			}
			if s.ActiveBattle != nil || s.LastBattle == nil || s.LastBattle.Outcome != "victory" {
				t.Fatalf("%s did not complete first breach: battle=%#v result=%#v", strategy.name, s.ActiveBattle, s.LastBattle)
			}
			if s.LastBattle.Doctrine != strategy.doctrine || len(s.CompletedTargets) != 1 || s.CompletedTargets[0] != "t_01" {
				t.Fatalf("%s did not preserve strategy outcome: result=%#v completed=%v", strategy.name, s.LastBattle, s.CompletedTargets)
			}
		})
	}
}

func TestRunnerSixHourSimulationRemainsSafeAndProgresses(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	e := sim.New(catalog, 401)
	s := sim.NewState(401)
	s.Resources.Credits = mathx.New(1000)
	s.Resources.Compute = mathx.New(1000)
	s.Resources.BotCapacity = 20
	s.Upgrades = []string{"secondhand_compiler", "street_contacts", "autodialer", "packet_scheduler", "signal_profiler"}
	s.Automation = game.Automation{Contracts: true, Fabrication: true, ReconTargetID: "t_01"}
	s, _, err = e.Apply(s, game.Command{Kind: game.StartRecon, ID: "t_01"})
	if err != nil {
		t.Fatal(err)
	}
	s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 6 * 60 * 60})
	if err != nil {
		t.Fatal(err)
	}
	if s.LogicalTick != 6*60*60 || len(s.EventLog) > sim.EventLogLimit || s.Resources.Credits < 0 || s.Resources.Compute < 0 || s.Resources.Bots < 0 || s.Resources.Data < 0 {
		t.Fatalf("six-hour Runner simulation violated bounds: tick=%d events=%d resources=%#v", s.LogicalTick, len(s.EventLog), s.Resources)
	}
	if s.Intel["t_01"] == 0 || s.Resources.BotCapacity < s.Resources.Bots {
		t.Fatalf("six-hour Runner simulation produced invalid progression: intel=%v resources=%#v", s.Intel, s.Resources)
	}
}

func TestRunnerProgressesThroughEveryCatalogRouteWithTemporaryAccess(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	e := sim.New(catalog, 509)
	s := sim.NewState(509)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 500
	s.Resources.BotCapacity = 500
	for routeIndex, target := range catalog.Targets {
		var commandErr error
		s, _, commandErr = e.Apply(s, game.Command{Kind: game.StartBattle, ID: target.ID, Quantity: 100, Doctrine: game.Balanced})
		if commandErr != nil {
			t.Fatalf("could not launch route %s: %v", target.ID, commandErr)
		}
		for ticks := int64(0); s.ActiveBattle != nil && ticks <= target.TimeLimit; ticks++ {
			s, _, commandErr = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
			if commandErr != nil {
				t.Fatalf("route %s tick failed: %v", target.ID, commandErr)
			}
		}
		if s.ActiveBattle != nil || s.LastBattle == nil || s.LastBattle.Outcome != "victory" {
			t.Fatalf("route %s did not resolve as a victory: battle=%#v result=%#v", target.ID, s.ActiveBattle, s.LastBattle)
		}
		if len(s.CompletedTargets) != routeIndex+1 || s.CompletedTargets[len(s.CompletedTargets)-1] != target.ID {
			t.Fatalf("route progression order changed after %s: %v", target.ID, s.CompletedTargets)
		}
		if s.Access[target.ID] <= 0 && target.AccessTicks > 0 {
			t.Fatalf("route %s did not grant temporary access: %v", target.ID, s.Access)
		}
	}
	if len(s.CompletedTargets) != len(catalog.Targets) || s.CompletedTargets[len(s.CompletedTargets)-1] != "t_05" {
		t.Fatalf("Runner did not reach final catalog route: %v", s.CompletedTargets)
	}
}
