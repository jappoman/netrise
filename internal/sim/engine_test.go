package sim

import (
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"testing"
)

func catalog() content.Catalog {
	return content.Catalog{SchemaVersion: 1, Generators: []content.Generator{{ID: "rig", Name: "Rig", Cost: 50, ComputePerSecond: 1}}, Upgrades: []content.Upgrade{{ID: "u", Cost: 10}}}
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
