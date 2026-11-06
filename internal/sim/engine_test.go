package sim

import (
	"encoding/json"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"testing"
)

func catalog() content.Catalog {
	return content.Catalog{SchemaVersion: 1, Generators: []content.Generator{{ID: "rig", Name: "Rig", Cost: 50, ComputePerSecond: 1}}, Upgrades: []content.Upgrade{{ID: "u", Cost: 10}, {ID: "secondhand_compiler", Name: "Compiler", Cost: 75}}, Targets: []content.Target{{ID: "target", Name: "Target", Reachable: true, Layers: []content.Layer{{Integrity: 20, ICE: 1, Counter: 0}, {Integrity: 15, ICE: 1, Counter: 0}}, LootCredits: 5, LootData: 2, TimeLimit: 30}}}
}
func TestDeterministicAdvance(t *testing.T) {
	a := NewState(17)
	b := NewState(17)
	e := New(catalog(), 17)
	a = e.Advance(a, 600)
	b = New(catalog(), 17).Advance(b, 600)
	if a.LogicalTick != 600 || a.Resources.Compute != b.Resources.Compute {
		t.Fatalf("not deterministic: %#v %#v", a, b)
	}
}
func TestPurchaseAtomicAndCostScaling(t *testing.T) {
	e := New(catalog(), 1)
	s := NewState(1)
	s.Resources.Credits = mathx.New(49)
	_, _, err := e.Apply(s, game.Command{Kind: game.BuyGenerator, ID: "rig"})
	if err == nil || s.Resources.Credits != mathx.New(49) {
		t.Fatal("failed purchase mutated state")
	}
	s.Resources.Credits = mathx.New(50)
	s, _, err = e.Apply(s, game.Command{Kind: game.BuyGenerator, ID: "rig"})
	if err != nil || s.Resources.Credits != 0 {
		t.Fatal(err)
	}
}
func TestFabricationCap(t *testing.T) {
	e := New(catalog(), 1)
	s := NewState(1)
	s.Upgrades = append(s.Upgrades, "secondhand_compiler")
	s.Resources.Credits = mathx.New(2)
	s.Resources.Compute = mathx.New(10)
	_, _, err := e.Apply(s, game.Command{Kind: game.QueueFabrication, Quantity: 21})
	if err == nil {
		t.Fatal("over-cap fabrication accepted")
	}
}

func TestFabricationReservesInputsAtomically(t *testing.T) {
	e := New(catalog(), 1)
	s := NewState(1)
	s.Upgrades = append(s.Upgrades, "secondhand_compiler")
	s.Resources.Credits = mathx.New(4)
	s.Resources.Compute = mathx.New(20)
	got, _, err := e.Apply(s, game.Command{Kind: game.QueueFabrication, Quantity: 2})
	if err != nil || got.Resources.Credits != 0 || got.Resources.Compute != 0 {
		t.Fatalf("inputs were not reserved: %#v, %v", got.Resources, err)
	}
	_, _, err = e.Apply(s, game.Command{Kind: game.QueueFabrication, Quantity: 3})
	if err == nil || s.Resources.Credits != mathx.New(4) || s.Resources.Compute != mathx.New(20) {
		t.Fatal("failed fabrication partially mutated source state")
	}
}

func TestCompilerUnlockAndBattleResolution(t *testing.T) {
	e := New(catalog(), 17)
	s := NewState(17)
	s.Resources.Credits = mathx.New(75)
	if _, _, err := e.Apply(s, game.Command{Kind: game.QueueFabrication, Quantity: 1}); err == nil {
		t.Fatal("fabrication bypassed compiler unlock")
	}
	var err error
	s, _, err = e.Apply(s, game.Command{Kind: game.BuyUpgrade, ID: "secondhand_compiler"})
	if err != nil {
		t.Fatal(err)
	}
	s.Resources.Bots = 10
	s, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 10, Doctrine: game.Balanced})
	if err != nil || s.ActiveBattle == nil || s.Resources.Bots != 0 {
		t.Fatalf("battle did not start atomically: %#v, %v", s, err)
	}
	for i := 0; i < 30 && s.ActiveBattle != nil; i++ {
		s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			t.Fatal(err)
		}
	}
	if s.ActiveBattle != nil || s.Resources.Credits <= mathx.New(0) || s.Resources.Data != mathx.New(2) {
		t.Fatalf("battle did not resolve with loot: %#v", s)
	}
}

func TestWithdrawReturnsSurvivors(t *testing.T) {
	e := New(catalog(), 1)
	s := NewState(1)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 4
	s, _, err := e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 3, Doctrine: game.Stealth})
	if err != nil {
		t.Fatal(err)
	}
	s, _, err = e.Apply(s, game.Command{Kind: game.WithdrawBattle})
	if err != nil || s.ActiveBattle != nil || s.Resources.Bots != 4 {
		t.Fatalf("withdraw did not return survivors: %#v, %v", s, err)
	}
}

func TestBattleDeterminismIncludesEvents(t *testing.T) {
	e1, e2 := New(catalog(), 77), New(catalog(), 77)
	start := NewState(77)
	start.Upgrades = []string{"secondhand_compiler"}
	start.Resources.Bots = 12
	a, eventsA, err := e1.Apply(start, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 10, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	b, eventsB, err := e2.Apply(start, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 10, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	for a.ActiveBattle != nil || b.ActiveBattle != nil {
		var ea, eb []game.Event
		a, ea, err = e1.Apply(a, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			t.Fatal(err)
		}
		b, eb, err = e2.Apply(b, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			t.Fatal(err)
		}
		eventsA, eventsB = append(eventsA, ea...), append(eventsB, eb...)
	}
	aa, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	if string(aa) != string(bb) {
		t.Fatalf("same seed diverged:\n%s\n%s", aa, bb)
	}
	if len(eventsA) != len(eventsB) {
		t.Fatalf("event streams differ: %d vs %d", len(eventsA), len(eventsB))
	}
}

func TestReinforceIsAtomicAndSingleUse(t *testing.T) {
	e := New(catalog(), 2)
	s := NewState(2)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 5
	var err error
	s, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 3, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	before := s.Resources.Bots
	s, _, err = e.Apply(s, game.Command{Kind: game.ReinforceBattle, Quantity: 2})
	if err != nil || s.Resources.Bots != 0 || s.ActiveBattle.CommittedBots != 5 {
		t.Fatalf("reinforce failed: %#v, %v", s, err)
	}
	if _, _, err = e.Apply(s, game.Command{Kind: game.ReinforceBattle, Quantity: 1}); err == nil {
		t.Fatal("second intervention accepted")
	}
	if before != 2 {
		t.Fatalf("unexpected reserve before reinforcement: %d", before)
	}
}

func TestBattleTimeoutReturnsSurvivorsWithoutLoot(t *testing.T) {
	c := catalog()
	c.Targets[0].Layers[0].Integrity = 100000
	c.Targets[0].TimeLimit = 1
	e := New(c, 4)
	s := NewState(4)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 3
	s, _, err := e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 2, Doctrine: game.Stealth})
	if err != nil {
		t.Fatal(err)
	}
	s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.ActiveBattle != nil || s.Resources.Bots != 3 || s.Resources.Credits != mathx.New(25) {
		t.Fatalf("timeout invariant failed: %#v, %v", s, err)
	}
}

func TestBattleResultIsPersistedOnce(t *testing.T) {
	e := New(catalog(), 9)
	s := NewState(9)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 10
	var err error
	s, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 8, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	for s.ActiveBattle != nil {
		s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			t.Fatal(err)
		}
	}
	if s.LastBattle == nil || s.LastBattle.Outcome != "victory" || s.LastBattle.CreditsLooted != 5 {
		t.Fatalf("missing or invalid result: %#v", s.LastBattle)
	}
	result := *s.LastBattle
	s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 10})
	if err != nil || *s.LastBattle != result {
		t.Fatalf("result changed after finalization: %#v, %v", s.LastBattle, err)
	}
}

func TestOfflineCapAndClockReversal(t *testing.T) {
	e := New(catalog(), 1)
	s := NewState(1)
	got, applied := e.ReconcileOffline(s, OfflineCapTicks+100)
	if applied != OfflineCapTicks || got.LogicalTick != OfflineCapTicks {
		t.Fatalf("offline cap failed: ticks=%d state=%d", applied, got.LogicalTick)
	}
	got, applied = e.ReconcileOffline(s, -1)
	if applied != 0 || got.LogicalTick != 0 {
		t.Fatal("clock reversal advanced the simulation")
	}
}

func TestOfflineReconciliationDoesNotResolveBattle(t *testing.T) {
	e := New(catalog(), 12)
	s := NewState(12)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 10
	var err error
	s, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 8, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	before := *s.ActiveBattle
	rngBefore := s.RNGState
	got, applied := e.ReconcileOffline(s, 600)
	if applied != 600 || got.ActiveBattle == nil {
		t.Fatalf("offline reconciliation resolved battle: %#v", got)
	}
	if *got.ActiveBattle != before || got.RNGState != rngBefore {
		t.Fatalf("offline reconciliation changed hostile state: before=%#v after=%#v", before, *got.ActiveBattle)
	}
	if got.LogicalTick != 600 {
		t.Fatalf("safe production did not advance: %d", got.LogicalTick)
	}
}
