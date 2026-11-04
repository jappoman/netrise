package sim

import (
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
