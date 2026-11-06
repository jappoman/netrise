package sim

import (
	"encoding/json"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"reflect"
	"testing"
)

func catalog() content.Catalog {
	return content.Catalog{SchemaVersion: 1, Generators: []content.Generator{{ID: "rig", Name: "Rig", Cost: 50, ComputePerSecond: 1}}, Upgrades: []content.Upgrade{{ID: "u", Cost: 10}, {ID: "secondhand_compiler", Name: "Compiler", Cost: 75}, {ID: "street_contacts", Name: "Contacts", Cost: 150}}, Research: []content.ResearchNode{{ID: "hardware", Name: "Hardware", CostData: 10}, {ID: "intrusion", Name: "Intrusion", CostData: 20, Prerequisites: []string{"hardware"}}}, Targets: []content.Target{{ID: "target", Name: "Target", Reachable: true, Layers: []content.Layer{{Integrity: 20, ICE: 1, Counter: 0}, {Integrity: 15, ICE: 1, Counter: 0}}, LootCredits: 5, LootData: 2, TimeLimit: 30, ReconCost: 10, ReconTicks: 15, IntelGain: 50}}}
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

func TestRunnerDataLootRespectsCapacityAndCacheExpansion(t *testing.T) {
	c := catalog()
	c.Generators = append(c.Generators, content.Generator{ID: "cache", Name: "Cache", Cost: 50, DataCapacity: 100})
	e := New(c, 105)
	s := NewState(105)
	s.Resources.Credits = mathx.New(100)
	s.Resources.DataCapacity = 1
	s.Resources.Bots = 20
	s.Upgrades = []string{"secondhand_compiler"}
	var err error
	s, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 20, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	for ticks := int64(0); s.ActiveBattle != nil && ticks <= 30; ticks++ {
		s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			t.Fatal(err)
		}
	}
	if s.LastBattle == nil || s.LastBattle.Outcome != "victory" || s.LastBattle.DataLooted != 1 || s.Resources.Data.Units() != 1 || s.RunExtractedData != 1 {
		t.Fatalf("data loot exceeded storage capacity: result=%#v resources=%#v", s.LastBattle, s.Resources)
	}
	s.Resources.Credits = mathx.New(100)
	s, _, err = e.Apply(s, game.Command{Kind: game.BuyGenerator, ID: "cache"})
	if err != nil || s.Resources.DataCapacity != 101 {
		t.Fatalf("data cache did not expand capacity: resources=%#v err=%v", s.Resources, err)
	}
}

func TestVictoryGrantsTemporaryAccessAndOfflineKeepsItSafe(t *testing.T) {
	c := catalog()
	c.Targets = []content.Target{
		{ID: "t_01", Name: "Relay", Reachable: true, AccessTicks: 3, Layers: []content.Layer{{Integrity: 1}}, TimeLimit: 10},
		{ID: "t_02", Name: "Vault", RequiresTargetID: "t_01", Layers: []content.Layer{{Integrity: 1}}, TimeLimit: 10},
	}
	e := New(c, 108)
	s := NewState(108)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 2
	s, events, err := e.Apply(s, game.Command{Kind: game.StartBattle, ID: "t_01", Quantity: 1, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	s, events, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.Access["t_01"] != 3 || !hasEvent(events, "AccessGranted") {
		t.Fatalf("victory did not grant access: state=%#v events=%#v err=%v", s, events, err)
	}
	offline, _ := e.ReconcileOffline(s, 100)
	if offline.Access["t_01"] != 3 {
		t.Fatalf("offline reconciliation consumed temporary access: %#v", offline.Access)
	}
	s, events, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 3})
	if err != nil || s.Access["t_01"] != 0 || !hasEvent(events, "AccessExpired") {
		t.Fatalf("active ticks did not expire access: state=%#v events=%#v err=%v", s.Access, events, err)
	}
	if _, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "t_02", Quantity: 1, Doctrine: game.Balanced}); err == nil {
		t.Fatal("expired access still allowed the adjacent target")
	}
	s.CompletedTargets = []string{"legacy-route"}
	legacyTarget := content.Target{ID: "legacy-next", RequiresTargetID: "legacy-route"}
	if !e.isReachable(s, legacyTarget) {
		t.Fatal("legacy completed route was lost when access map was introduced")
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

func TestTargetCooldownPreventsRepeatedLootUntilElapsed(t *testing.T) {
	c := catalog()
	c.Targets = []content.Target{{ID: "target", Name: "Target", Reachable: true, Layers: []content.Layer{{Integrity: 1, ICE: 0}}, LootCredits: 5, TimeLimit: 10, CooldownTicks: 3, ReconCost: 10, ReconTicks: 1, IntelGain: 10}}
	e := New(c, 99)
	s := NewState(99)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 8
	var err error
	s, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 8, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.LastBattle == nil || s.LastBattle.Outcome != "victory" || s.TargetCooldowns["target"] != 3 {
		t.Fatalf("victory did not establish cooldown: state=%#v err=%v", s, err)
	}
	credits := s.Resources.Credits
	bots := s.Resources.Bots
	_, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 1, Doctrine: game.Balanced})
	if err == nil || err.(*game.CommandError).Reason != game.ReasonTargetCooldown {
		t.Fatalf("cooldown raid was not rejected: err=%v", err)
	}
	if s.Resources.Credits != credits || s.Resources.Bots != bots {
		t.Fatal("cooldown rejection partially changed resources")
	}
	s.Upgrades = append(s.Upgrades, "street_contacts")
	reconCredits := s.Resources.Credits
	_, _, err = e.Apply(s, game.Command{Kind: game.StartRecon, ID: "target"})
	if err == nil || err.(*game.CommandError).Reason != game.ReasonTargetCooldown || s.Resources.Credits != reconCredits {
		t.Fatalf("cooldown recon was not rejected atomically: err=%v state=%#v", err, s)
	}
	s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 3})
	if err != nil || len(s.TargetCooldowns) != 0 {
		t.Fatalf("cooldown did not expire: cooldowns=%v err=%v", s.TargetCooldowns, err)
	}
	if _, _, err = e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 1, Doctrine: game.Balanced}); err != nil {
		t.Fatalf("target remained blocked after cooldown: %v", err)
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
	if applied != OfflineCapTicks || got.LogicalTick != OfflineCapTicks || got.OfflineSummary == nil || got.OfflineSummary.Ticks != OfflineCapTicks {
		t.Fatalf("offline cap failed: ticks=%d state=%d", applied, got.LogicalTick)
	}
	got, applied = e.ReconcileOffline(s, -1)
	if applied != 0 || got.LogicalTick != 0 {
		t.Fatal("clock reversal advanced the simulation")
	}
}

func TestCorporationOfflineSummaryIsSafeAndNonHostile(t *testing.T) {
	e := New(catalog(), 5)
	s := NewCorporationState(5)
	before := s.Corp.Capital
	got, applied := e.ReconcileOffline(s, 100)
	if applied != 100 || got.OfflineSummary == nil || got.OfflineSummary.CapitalProduced <= 0 || got.Corp.Capital <= before || got.ActiveBattle != nil {
		t.Fatalf("invalid Corporation offline summary: before=%s after=%s summary=%#v", before, got.Corp.Capital, got.OfflineSummary)
	}
	if got.Corp.AttackCounter != s.Corp.AttackCounter {
		t.Fatal("offline time advanced hostile attack counter")
	}
}

func TestCorporationIncursionUsesActiveBudgetOnly(t *testing.T) {
	c := catalog()
	c.Targets[0].ID = "t_01"
	e := New(c, 106)
	s := NewCorporationState(106)
	s.Corp.IncursionBudget = 8
	s, events, err := e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 30})
	if err != nil || s.ActiveBattle == nil || s.ActiveBattle.CommittedBots != 8 || s.Corp.IncursionSpent != 8 || len(events) == 0 {
		t.Fatalf("budgeted incursion did not start: state=%#v events=%#v err=%v", s, events, err)
	}
	s.ActiveBattle = nil
	s.Corp.AttackCounter = 29
	s, events, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.ActiveBattle != nil || s.Corp.IncursionSpent != 8 || len(events) != 0 {
		t.Fatalf("incursion exceeded budget: state=%#v events=%#v err=%v", s, events, err)
	}

	offline := NewCorporationState(107)
	offline.Corp.IncursionBudget = 8
	offline, _ = e.ReconcileOffline(offline, 60)
	if offline.Corp.IncursionSpent != 0 || offline.ActiveBattle != nil {
		t.Fatalf("offline reconciliation launched or spent an incursion: %#v", offline.Corp)
	}
}

func TestCorporationIncursionBudgetUsesCatalogConfiguration(t *testing.T) {
	c := catalog()
	c.Targets[0].ID = "t_01"
	c.CorpIncursion = content.CorpIncursion{IntervalTicks: 1, BotBudget: 12, BotCommitment: 6}
	e := New(c, 109)
	s := NewCorporationState(109)
	s.Corp.IncursionBudget = 0
	s, _, err := e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.Corp.IncursionBudget != 12 || s.ActiveBattle == nil || s.ActiveBattle.CommittedBots != 6 {
		t.Fatalf("catalogued incursion budget was not applied: corp=%#v battle=%#v err=%v", s.Corp, s.ActiveBattle, err)
	}
}

func TestOfflinePassiveAggregationMatchesTickPath(t *testing.T) {
	e := New(catalog(), 104)
	start := NewState(104)
	start.Resources.Compute = mathx.New(3)
	start.TargetCooldowns = map[string]int64{"target": 80}
	aggregated, applied := e.ReconcileOffline(start, 600)
	if applied != 600 {
		t.Fatalf("unexpected offline ticks: %d", applied)
	}
	tickStart := start
	tickStart.TargetCooldowns = map[string]int64{"target": 80}
	ticked := e.Advance(tickStart, 600)
	if aggregated.LogicalTick != ticked.LogicalTick || aggregated.Resources.Compute != ticked.Resources.Compute || aggregated.TargetCooldowns["target"] != ticked.TargetCooldowns["target"] {
		t.Fatalf("aggregated offline state diverged: aggregated=%#v ticked=%#v", aggregated, ticked)
	}
	if aggregated.OfflineSummary == nil || aggregated.OfflineSummary.Ticks != 600 {
		t.Fatalf("offline summary missing from aggregate path: %#v", aggregated.OfflineSummary)
	}
}

func TestOfflineAggregationPreservesActiveBattle(t *testing.T) {
	e := New(catalog(), 105)
	s := NewState(105)
	s.Upgrades = []string{"secondhand_compiler"}
	s.Resources.Bots = 8
	s, _, err := e.Apply(s, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 8, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	before := *s.ActiveBattle
	got, applied := e.ReconcileOffline(s, 600)
	if applied != 600 || got.ActiveBattle == nil || *got.ActiveBattle != before || got.RNGState != s.RNGState {
		t.Fatalf("offline aggregation mutated hostile state: applied=%d battle=%#v before=%#v rng=%d/%d", applied, got.ActiveBattle, before, got.RNGState, s.RNGState)
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

func TestRunnerVerticalSliceEndToEnd(t *testing.T) {
	e := New(catalog(), 42)
	s := NewState(42)
	apply := func(command game.Command) {
		var err error
		s, _, err = e.Apply(s, command)
		if err != nil {
			t.Fatal(err)
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
	apply(game.Command{Kind: game.StartRecon, ID: "target"})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 15})
	if s.Recon != nil || s.Intel["target"] != 50 {
		t.Fatalf("recon did not complete: recon=%#v intel=%#v", s.Recon, s.Intel)
	}
	apply(game.Command{Kind: game.CollectContract})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 5})
	apply(game.Command{Kind: game.QueueFabrication, Quantity: 5})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 3})
	if s.Resources.Bots != 5 {
		t.Fatalf("fabrication did not produce bots: %d", s.Resources.Bots)
	}
	apply(game.Command{Kind: game.StartBattle, ID: "target", Quantity: 5, Doctrine: game.Balanced})
	for s.ActiveBattle != nil {
		apply(game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	}
	if s.LastBattle == nil || s.LastBattle.Outcome != "victory" || !contains(s.CompletedTargets, "target") {
		t.Fatalf("vertical slice did not finish: result=%#v completed=%v", s.LastBattle, s.CompletedTargets)
	}

	locked := catalog().Targets[0]
	locked.ID = "next"
	locked.RequiresTargetID = "target"
	locked.Reachable = false
	other := New(catalog(), 42)
	other.Catalog.Targets = append(other.Catalog.Targets, locked)
	if _, _, err := other.Apply(s, game.Command{Kind: game.StartBattle, ID: "next", Quantity: 1, Doctrine: game.Balanced}); err != nil {
		t.Fatalf("victory did not unlock next target: %v", err)
	}
}

func TestAutomationPoliciesRunInsideDeterministicTicks(t *testing.T) {
	e := New(catalog(), 91)
	s := NewState(91)
	s.Upgrades = []string{"autodialer", "secondhand_compiler", "packet_scheduler", "street_contacts", "signal_profiler"}
	s.Resources.Credits = mathx.New(50)
	s.Resources.Compute = mathx.New(100)
	s.Resources.BotCapacity = 5
	apply := func(command game.Command) {
		var err error
		s, _, err = e.Apply(s, command)
		if err != nil {
			t.Fatal(err)
		}
	}
	apply(game.Command{Kind: game.EnableAutomation, ID: "contracts"})
	apply(game.Command{Kind: game.EnableAutomation, ID: "fabrication"})
	apply(game.Command{Kind: game.EnableAutomation, ID: "recon", TargetID: "target"})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 20})
	if !s.Automation.Contracts || !s.Automation.Fabrication || s.Automation.ReconTargetID != "target" {
		t.Fatalf("automation policies not enabled: %#v", s.Automation)
	}
	if s.Resources.Bots == 0 || s.Intel["target"] == 0 {
		t.Fatalf("automation did not make progress: bots=%d intel=%v", s.Resources.Bots, s.Intel)
	}
	if s.Resources.Credits < 0 || s.Resources.Compute < 0 {
		t.Fatalf("automation created negative resources: %#v", s.Resources)
	}
}

func TestAutomationBudgetsStopPoliciesAtConfiguredCaps(t *testing.T) {
	c := catalog()
	c.Upgrades = append(c.Upgrades,
		content.Upgrade{ID: "autodialer"},
		content.Upgrade{ID: "packet_scheduler"},
		content.Upgrade{ID: "signal_profiler"},
	)
	e := New(c, 92)
	contracts := NewState(92)
	contracts.Upgrades = []string{"autodialer"}
	var err error
	contracts, _, err = e.Apply(contracts, game.Command{Kind: game.EnableAutomation, ID: "contracts", Quantity: 25})
	if err != nil {
		t.Fatal(err)
	}
	contracts, events, err := e.Apply(contracts, game.Command{Kind: game.AdvanceTicks, Ticks: 12})
	if err != nil || contracts.Automation.Contracts || contracts.Automation.ContractEarned != 25 || contracts.Resources.Credits != mathx.New(50) {
		t.Fatalf("contract budget did not stop atomically: state=%#v events=%#v err=%v", contracts, events, err)
	}

	fabrication := NewState(92)
	fabrication.Upgrades = []string{"secondhand_compiler", "packet_scheduler"}
	fabrication.Resources.Compute = mathx.New(100)
	fabrication.Resources.Credits = mathx.New(10)
	fabrication.Resources.BotCapacity = 5
	fabrication, _, err = e.Apply(fabrication, game.Command{Kind: game.EnableAutomation, ID: "fabrication", Quantity: 2})
	if err != nil {
		t.Fatal(err)
	}
	fabrication, _, err = e.Apply(fabrication, game.Command{Kind: game.AdvanceTicks, Ticks: 4})
	if err != nil || fabrication.Automation.Fabrication || fabrication.Automation.FabricationSpent != 2 || fabrication.Resources.Bots != 1 {
		t.Fatalf("fabrication budget did not stop atomically: state=%#v err=%v", fabrication, err)
	}

	recon := NewState(92)
	recon.Upgrades = []string{"street_contacts", "signal_profiler"}
	recon.Resources.Credits = mathx.New(20)
	recon, _, err = e.Apply(recon, game.Command{Kind: game.EnableAutomation, ID: "recon", TargetID: "target", Quantity: 10})
	if err != nil {
		t.Fatal(err)
	}
	recon, _, err = e.Apply(recon, game.Command{Kind: game.AdvanceTicks, Ticks: 16})
	if err != nil || recon.Automation.ReconTargetID != "" || recon.Automation.ReconSpent != 10 || recon.Intel["target"] == 0 {
		t.Fatalf("recon budget did not stop atomically: state=%#v err=%v", recon, err)
	}
}

func TestCorporationMaintenanceBudgetRestocksOnlyDuringActivePlay(t *testing.T) {
	e := New(catalog(), 93)
	s := NewCorporationState(93)
	s.Corp.ICE = 0
	s.Corp.ICECapacity = 2
	s, _, err := e.Apply(s, game.Command{Kind: game.CorpEnableMaintenance, Quantity: 10})
	if err != nil {
		t.Fatal(err)
	}
	s, events, err := e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.Corp.ICE != 1 || s.Corp.MaintenanceSpent != 10 || len(events) == 0 {
		t.Fatalf("maintenance did not restock during active play: state=%#v events=%#v err=%v", s.Corp, events, err)
	}
	s, events, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.Corp.MaintenanceAuto || s.Corp.ICE != 1 || len(events) == 0 || events[len(events)-1].Type != "MaintenanceStopped" {
		t.Fatalf("maintenance cap did not stop policy: state=%#v events=%#v err=%v", s.Corp, events, err)
	}

	offline := NewCorporationState(93)
	offline.Corp.ICE = 0
	offline, _, err = e.Apply(offline, game.Command{Kind: game.CorpEnableMaintenance, Quantity: 10})
	if err != nil {
		t.Fatal(err)
	}
	offline, _ = e.ReconcileOffline(offline, 10)
	if offline.Corp.ICE != 0 || offline.Corp.MaintenanceSpent != 0 || !offline.Corp.MaintenanceAuto {
		t.Fatalf("offline reconciliation spent maintenance budget: %#v", offline.Corp)
	}
}

func TestSecurityAssetExpandsICECapacityAndMaintenanceReserve(t *testing.T) {
	c := catalog()
	c.CorpAssets = []content.CorpAssetDefinition{{ID: "security", Name: "Security Desk", CostCapital: 50, Integrity: 100, ICECapacityBonus: 3}}
	e := New(c, 97)
	s := NewCorporationState(97)
	s.Corp.Capital = mathx.New(100)
	s.Corp.MaintenanceReserve = s.Corp.ICECapacity
	s, _, err := e.Apply(s, game.Command{Kind: game.CorpBuildAsset, ID: "security"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Corp.ICECapacity != 8 || s.Corp.MaintenanceReserve != 8 || s.Corp.Capital != mathx.New(50) {
		t.Fatalf("security asset did not expand capacity atomically: %#v", s.Corp)
	}
}

func TestCorporationReactivatesAssetsAfterUpkeepRecovery(t *testing.T) {
	e := New(catalog(), 98)
	s := NewCorporationState(98)
	s.Corp.Assets = []game.CorpAsset{{ID: "costly", Name: "Costly Service", UpkeepPerTick: 5, Operational: true}}
	s.Corp.Capital = 0
	s, events, err := e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.Corp.Assets[0].Operational {
		t.Fatalf("unaffordable asset was not throttled: state=%#v events=%#v err=%v", s.Corp, events, err)
	}
	for ticks := 0; ticks < 3 && !s.Corp.Assets[0].Operational; ticks++ {
		s, events, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			t.Fatal(err)
		}
	}
	if !s.Corp.Assets[0].Operational || s.Corp.Capital < 0 {
		t.Fatalf("asset did not recover after capital accumulation: state=%#v", s.Corp)
	}
	found := false
	for _, event := range events {
		if event.Type == "AssetReactivated" {
			found = true
		}
	}
	if !found {
		t.Fatalf("reactivation event missing: %#v", events)
	}
}

func TestCorporationResearchAllocationHonorsBudgetAndComputeReserve(t *testing.T) {
	e := New(catalog(), 94)
	s := NewCorporationState(94)
	s.Corp.Capital = mathx.New(500)
	s.Corp.Compute = mathx.New(2)
	s, _, err := e.Apply(s, game.Command{Kind: game.CorpConfigurePolicy, ID: "compute_reserve", Quantity: 1})
	if err != nil {
		t.Fatal(err)
	}
	s, _, err = e.Apply(s, game.Command{Kind: game.CorpConfigurePolicy, ID: "research", Quantity: 200})
	if err != nil {
		t.Fatal(err)
	}
	s, events, err := e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.Corp.ResearchAuto || s.Corp.ResearchSpent != 200 || s.Corp.Research != mathx.New(10) || len(events) < 2 {
		t.Fatalf("automated R&D did not allocate resources: state=%#v events=%#v err=%v", s.Corp, events, err)
	}
	reserve := NewCorporationState(94)
	reserve.Corp.Capital = mathx.New(500)
	reserve.Corp.Compute = mathx.New(2)
	reserve, _, err = e.Apply(reserve, game.Command{Kind: game.CorpConfigurePolicy, ID: "compute_reserve", Quantity: 2})
	if err != nil {
		t.Fatal(err)
	}
	reserve, _, err = e.Apply(reserve, game.Command{Kind: game.CorpConfigurePolicy, ID: "research", Quantity: 200})
	if err != nil {
		t.Fatal(err)
	}
	reserve, _, err = e.Apply(reserve, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || reserve.Corp.Research != 0 {
		t.Fatalf("compute reserve threshold was ignored: state=%#v err=%v", reserve.Corp, err)
	}
}

func TestCorporationResearchCompletesDataDrivenTechnology(t *testing.T) {
	c := catalog()
	c.CorpResearch = []content.CorpResearchNode{{
		ID: "efficiency", Name: "Efficiency", Threshold: 20,
		Effects: content.CorpResearchEffects{CapitalPerTickBonus: 3, UpkeepReductionPercent: 25},
	}}
	e := New(c, 95)
	s := NewCorporationState(95)
	s.Corp.Assets = append(s.Corp.Assets, game.CorpAsset{ID: "costly", Name: "Costly", UpkeepPerTick: 4, Operational: true})
	s.Corp.Capital = mathx.New(500)
	baseRate := e.CorporationGrossRate(s)
	baseUpkeep := e.CorporationUpkeepRate(s)
	s, _, err := e.Apply(s, game.Command{Kind: game.CorpRunResearch})
	if err != nil || len(s.Corp.CompletedResearch) != 0 {
		t.Fatalf("technology completed before threshold: state=%#v err=%v", s.Corp, err)
	}
	s, events, err := e.Apply(s, game.Command{Kind: game.CorpRunResearch})
	if err != nil || len(s.Corp.CompletedResearch) != 1 || s.Corp.CompletedResearch[0] != "efficiency" {
		t.Fatalf("technology did not complete at threshold: state=%#v events=%#v err=%v", s.Corp, events, err)
	}
	if e.CorporationGrossRate(s) != baseRate+3 {
		t.Fatalf("technology effect did not change gross rate: before=%d after=%d", baseRate, e.CorporationGrossRate(s))
	}
	if e.CorporationUpkeepRate(s) != baseUpkeep*75/100 {
		t.Fatalf("technology effect did not reduce upkeep: before=%d after=%d", baseUpkeep, e.CorporationUpkeepRate(s))
	}
	found := false
	for _, event := range events {
		if event.Type == "CorpTechnologyCompleted" {
			found = true
		}
	}
	if !found {
		t.Fatalf("technology completion event missing: %#v", events)
	}
}

func TestCorporationReputationAndExposureBoundRevenue(t *testing.T) {
	e := New(catalog(), 100)
	healthy := NewCorporationState(100)
	base := e.CorporationGrossRate(healthy)
	lowReputation := NewCorporationState(100)
	lowReputation.Corp.ReputationBP = 4000
	exposed := NewCorporationState(100)
	exposed.Corp.ExposureBP = 10000
	crisis := NewCorporationState(100)
	crisis.Corp.ReputationBP = 4000
	crisis.Corp.ExposureBP = 10000
	if e.CorporationRevenueFactor(healthy) != 100 || e.CorporationRevenueFactor(lowReputation) != 90 || e.CorporationRevenueFactor(exposed) != 50 || e.CorporationRevenueFactor(crisis) != 45 {
		t.Fatalf("unexpected revenue factors: healthy=%d low_rep=%d exposed=%d crisis=%d", e.CorporationRevenueFactor(healthy), e.CorporationRevenueFactor(lowReputation), e.CorporationRevenueFactor(exposed), e.CorporationRevenueFactor(crisis))
	}
	if base <= e.CorporationGrossRate(lowReputation) || base <= e.CorporationGrossRate(exposed) || e.CorporationGrossRate(lowReputation) <= e.CorporationGrossRate(crisis) || e.CorporationGrossRate(exposed) <= e.CorporationGrossRate(crisis) {
		t.Fatalf("reputation/exposure did not produce monotonic bounded revenue: healthy=%d low_rep=%d exposed=%d crisis=%d", base, e.CorporationGrossRate(lowReputation), e.CorporationGrossRate(exposed), e.CorporationGrossRate(crisis))
	}
	if e.CorporationGrossRate(crisis) < 1 {
		t.Fatalf("reputation/exposure penalty created a dead economy: %d", e.CorporationGrossRate(crisis))
	}
}

func TestCorporationCanIsolateAssetDuringIncursion(t *testing.T) {
	c := catalog()
	c.Targets = append(c.Targets, content.Target{ID: "t_01", Name: "Audit Target", Layers: []content.Layer{{Integrity: 30, ICE: 1, Counter: 1}}, TimeLimit: 30})
	e := New(c, 101)
	s := NewCorporationState(101)
	s.Corp.Capital = mathx.New(500)
	s.Corp.Assets = append(s.Corp.Assets, game.CorpAsset{ID: "archive_server", Name: "Archive Server", CapitalPerTick: 7, UpkeepPerTick: 2, Integrity: 140, Operational: true, IsolationCostCapital: 25, IsolationTicks: 3})
	var err error
	s, _, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 30})
	if err != nil || s.ActiveBattle == nil {
		t.Fatalf("incursion did not start: state=%#v err=%v", s, err)
	}
	beforeCapital := s.Corp.Capital
	s, events, err := e.Apply(s, game.Command{Kind: game.CorpIsolateAsset, ID: "archive_server"})
	if err != nil || s.ActiveBattle != nil || s.LastBattle == nil || s.LastBattle.Outcome != "isolated" {
		t.Fatalf("asset isolation did not end incursion: state=%#v events=%#v err=%v", s, events, err)
	}
	asset := s.Corp.Assets[len(s.Corp.Assets)-1]
	if asset.Operational || asset.IsolatedUntilTick != s.LogicalTick+3 || s.Corp.Capital != beforeCapital-mathx.New(25) {
		t.Fatalf("isolation state/cost incorrect: asset=%#v capital=%s", asset, s.Corp.Capital)
	}
	if len(events) != 2 || events[0].Type != "AssetIsolated" || events[1].Type != "BattleEnded" {
		t.Fatalf("isolation events incorrect: %#v", events)
	}
	s, events, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 2})
	if err != nil || s.Corp.Assets[len(s.Corp.Assets)-1].Operational {
		t.Fatalf("isolated asset returned too early: state=%#v events=%#v err=%v", s.Corp.Assets[len(s.Corp.Assets)-1], events, err)
	}
	s, events, err = e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || !s.Corp.Assets[len(s.Corp.Assets)-1].Operational {
		t.Fatalf("isolated asset did not recover: state=%#v events=%#v err=%v", s.Corp.Assets[len(s.Corp.Assets)-1], events, err)
	}
	found := false
	for _, event := range events {
		if event.Type == "AssetIsolationEnded" {
			found = true
		}
	}
	if !found {
		t.Fatalf("isolation recovery event missing: %#v", events)
	}
}

func TestCorporationAssetDamageAndRepairUseSharedBattle(t *testing.T) {
	c := catalog()
	c.Targets = append(c.Targets, content.Target{ID: "t_01", Name: "Audit Target", Layers: []content.Layer{{Integrity: 1, ICE: 0, Counter: 0}}, TimeLimit: 30, AssetDamage: 100})
	e := New(c, 102)
	s := NewCorporationState(102)
	s.Corp.Capital = mathx.New(500)
	s.Corp.Assets = append(s.Corp.Assets, game.CorpAsset{ID: "archive_server", Name: "Archive Server", CapitalPerTick: 7, Integrity: 100, MaxIntegrity: 100, Operational: true})
	s.ActiveBattle = &game.Battle{TargetID: "t_01", DefenderAssetID: "archive_server", AttackerFaction: game.Runner, CommittedBots: 8, LiveBots: 8, LayerIntegrity: 1, TimeLimit: 30}
	s, events, err := e.Apply(s, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil || s.ActiveBattle != nil || s.Corp.Assets[1].Integrity != 0 || s.Corp.Assets[1].Operational {
		t.Fatalf("breach did not damage/disable asset: state=%#v events=%#v err=%v", s, events, err)
	}
	if !hasEvent(events, "AssetDamaged") || !hasEvent(events, "AssetDisabled") {
		t.Fatalf("asset damage events missing: %#v", events)
	}
	s, events, err = e.Apply(s, game.Command{Kind: game.CorpRepairAsset, ID: "archive_server", Quantity: 40})
	if err != nil || s.Corp.Assets[1].Integrity != 40 || !s.Corp.Assets[1].Operational || !hasEvent(events, "AssetRepaired") {
		t.Fatalf("asset repair failed: state=%#v events=%#v err=%v", s.Corp.Assets[1], events, err)
	}
}

func TestCorporationRepairMigratesLegacyAssetIntegrity(t *testing.T) {
	c := catalog()
	c.CorpAssets = []content.CorpAssetDefinition{{ID: "legacy", Name: "Legacy Service", Integrity: 100}}
	e := New(c, 103)
	s := NewCorporationState(103)
	s.Corp.Capital = mathx.New(100)
	s.Corp.Assets = []game.CorpAsset{{ID: "legacy", Name: "Legacy Service", Integrity: 0, Operational: false}}
	s, events, err := e.Apply(s, game.Command{Kind: game.CorpRepairAsset, ID: "legacy", Quantity: 25})
	if err != nil || s.Corp.Assets[0].Integrity != 25 || s.Corp.Assets[0].MaxIntegrity != 100 || !s.Corp.Assets[0].Operational || !hasEvent(events, "AssetRepaired") {
		t.Fatalf("legacy asset repair did not use catalog maximum: state=%#v events=%#v err=%v", s.Corp.Assets[0], events, err)
	}
}

func hasEvent(events []game.Event, eventType string) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func TestResearchPrerequisitesAndIdentityBurnAreAtomic(t *testing.T) {
	e := New(catalog(), 33)
	s := NewState(33)
	s.Resources.Data = mathx.New(20)
	if _, _, err := e.Apply(s, game.Command{Kind: game.BuyResearch, ID: "intrusion"}); err == nil {
		t.Fatal("research prerequisite bypassed")
	}
	got, _, err := e.Apply(s, game.Command{Kind: game.BuyResearch, ID: "hardware"})
	if err != nil || got.Resources.Data != mathx.New(10) || len(got.Research) != 1 {
		t.Fatalf("research purchase failed: %#v, %v", got, err)
	}
	got.RunExtractedData = 10000
	got.Resources.Credits = mathx.New(999)
	got.Resources.Bots = 4
	got.Generators = []game.Generator{{ID: "rig", Count: 2}}
	reset, _, err := e.Apply(got, game.Command{Kind: game.IdentityBurn})
	if err != nil {
		t.Fatal(err)
	}
	if reset.GhostKeys != 1 || reset.PrestigeCount != 1 || reset.RunExtractedData != 0 || reset.Resources.Bots != 0 || len(reset.Generators) != 0 || len(reset.Research) != 0 {
		t.Fatalf("identity burn did not reset run state: %#v", reset)
	}
	if _, _, err := e.Apply(reset, game.Command{Kind: game.IdentityBurn}); err == nil {
		t.Fatal("tiny reset unexpectedly granted a second reward")
	}
}

func TestResearchEffectsChangeSimulationThroughCatalogData(t *testing.T) {
	c := catalog()
	c.Research[0].Effects.ComputePerSecondBonus = 2
	c.Research[0].Effects.StealthHeatReductionPercent = 15
	c.Research[1].Effects.AttackBonusPercent = 8
	c.Research[1].Effects.AutomationIntervalReduction = 1
	e := New(c, 52)
	s := NewState(52)
	s.Research = []string{"hardware", "intrusion"}
	if got := e.ComputeRate(s); got != 4 {
		t.Fatalf("research compute effect not applied: got %d", got)
	}
	base := NewState(52)
	base.Upgrades = []string{"secondhand_compiler"}
	base.Resources.Bots = 8
	base, _, err := e.Apply(base, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 8, Doctrine: game.Stealth})
	if err != nil {
		t.Fatal(err)
	}
	withResearch := NewState(52)
	withResearch.Upgrades = []string{"secondhand_compiler"}
	withResearch.Resources.Bots = 8
	withResearch.Research = []string{"hardware"}
	withResearch, _, err = e.Apply(withResearch, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 8, Doctrine: game.Stealth})
	if err != nil {
		t.Fatal(err)
	}
	base, _, err = e.Apply(base, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil {
		t.Fatal(err)
	}
	withResearch, _, err = e.Apply(withResearch, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil {
		t.Fatal(err)
	}
	if base.HeatBP-withResearch.HeatBP != 15 {
		t.Fatalf("stealth research effect not applied: base=%d researched=%d", base.HeatBP, withResearch.HeatBP)
	}
	attackBase := NewState(52)
	attackBase.Upgrades = []string{"secondhand_compiler"}
	attackBase.Resources.Bots = 8
	attackResearch := attackBase
	attackResearch.Research = []string{"intrusion"}
	attackBase, _, err = e.Apply(attackBase, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 8, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	attackResearch, _, err = e.Apply(attackResearch, game.Command{Kind: game.StartBattle, ID: "target", Quantity: 8, Doctrine: game.Balanced})
	if err != nil {
		t.Fatal(err)
	}
	attackBase, _, err = e.Apply(attackBase, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil {
		t.Fatal(err)
	}
	attackResearch, _, err = e.Apply(attackResearch, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if err != nil {
		t.Fatal(err)
	}
	if attackResearch.ActiveBattle.LayerIntegrity >= attackBase.ActiveBattle.LayerIntegrity {
		t.Fatalf("attack research effect not applied: base=%d researched=%d", attackBase.ActiveBattle.LayerIntegrity, attackResearch.ActiveBattle.LayerIntegrity)
	}
	automation := NewState(52)
	automation.Upgrades = []string{"autodialer"}
	automation.Automation.Contracts = true
	automation.Research = []string{"intrusion"}
	automation, events, err := e.Apply(automation, game.Command{Kind: game.AdvanceTicks, Ticks: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != "AutomationProduced" || automation.Resources.Credits != mathx.New(35) {
		t.Fatalf("automation research effect not applied: events=%#v resources=%#v", events, automation.Resources)
	}
}

func TestUpgradeEffectsChangeEconomyAndPreparation(t *testing.T) {
	c := catalog()
	c.Upgrades[0].Effects = content.UpgradeEffects{
		ComputePerSecondBonus:     2,
		BotCapacityBonus:          5,
		FabricationTicksReduction: 2,
		ReconTicksReduction:       5,
		ResearchDiscountPercent:   50,
	}
	e := New(c, 64)
	s := NewState(64)
	s.Resources.Credits = mathx.New(100)
	s.Resources.Data = mathx.New(5)
	s.Upgrades = []string{"secondhand_compiler", "street_contacts", "u"}
	if got := e.ComputeRate(s); got != 4 {
		t.Fatalf("upgrade compute effect not applied: got %d", got)
	}
	if _, _, err := e.Apply(s, game.Command{Kind: game.BuyResearch, ID: "hardware"}); err != nil {
		t.Fatalf("upgrade research discount not applied: %v", err)
	}
	s.Resources.Compute = mathx.New(10)
	s.Resources.Bots = 1
	s.Resources.Credits = mathx.New(10)
	queued, _, err := e.Apply(s, game.Command{Kind: game.QueueFabrication, Quantity: 1})
	if err != nil || queued.Fabrication == nil || queued.Fabrication.RemainingTicks != 1 {
		t.Fatalf("fabrication upgrade effect not applied: state=%#v err=%v", queued, err)
	}
	s.Fabrication = nil
	s.Resources.Credits = mathx.New(20)
	recon, _, err := e.Apply(s, game.Command{Kind: game.StartRecon, ID: "target"})
	if err != nil || recon.Recon == nil || recon.Recon.RemainingTicks != 10 {
		t.Fatalf("recon upgrade effect not applied: state=%#v err=%v", recon, err)
	}
}

func TestEventArchiveIsBoundedAndRejectsFailedCommands(t *testing.T) {
	e := New(catalog(), 44)
	s := NewState(44)
	for i := 0; i < EventLogLimit+25; i++ {
		var err error
		s, _, err = e.Apply(s, game.Command{Kind: game.CollectContract})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(s.EventLog) != EventLogLimit || s.EventLog[len(s.EventLog)-1].Type != "ResourceProduced" {
		t.Fatalf("event archive bound failed: len=%d last=%#v", len(s.EventLog), s.EventLog[len(s.EventLog)-1])
	}
	before := len(s.EventLog)
	if _, _, err := e.Apply(s, game.Command{Kind: game.BuyGenerator, ID: "missing"}); err == nil {
		t.Fatal("invalid command accepted")
	}
	if len(s.EventLog) != before {
		t.Fatal("failed command changed event archive")
	}
}

func TestCorporationMarketExpansionUsesCatalogTiersAndIsAtomic(t *testing.T) {
	c := catalog()
	c.CorpMarkets = []content.CorpMarketExpansion{
		{ID: "metro", Name: "Metro Contracts", CostCapital: 150, DemandBonusBP: 1000, InfluenceGain: 10, ValuationGain: 150},
		{ID: "regional", Name: "Regional Exchange", CostCapital: 300, DemandBonusBP: 1000, InfluenceCost: 10, InfluenceGain: 20, ValuationGain: 300},
	}
	e := New(c, 104)
	s := NewCorporationState(104)
	s.Corp.Capital = mathx.New(149)
	before := s
	before.Corp = &*s.Corp
	before.Corp.Assets = append([]game.CorpAsset(nil), s.Corp.Assets...)
	if _, _, err := e.Apply(s, game.Command{Kind: game.CorpExpandMarket}); err == nil {
		t.Fatal("unaffordable first market tier was accepted")
	}
	if s.Corp.MarketTier != 0 || s.Corp.DemandBP != 10000 {
		t.Fatalf("failed command mutated market state: %#v", s.Corp)
	}
	s.Corp.Capital = mathx.New(150)
	got, events, err := e.Apply(s, game.Command{Kind: game.CorpExpandMarket})
	if err != nil {
		t.Fatal(err)
	}
	if got.Corp.MarketTier != 1 || got.Corp.DemandBP != 11000 || got.Corp.Influence != mathx.New(10) || got.Corp.RunValuation != 150 || got.Corp.Capital != mathx.New(0) || !hasEvent(events, "MarketExpanded") {
		t.Fatalf("first market tier was not applied from catalog: state=%#v events=%#v", got.Corp, events)
	}
	if !reflect.DeepEqual(before, s) {
		t.Fatalf("failed market command was not atomic: before=%#v after=%#v", before, s)
	}
	if _, _, err := e.Apply(got, game.Command{Kind: game.CorpExpandMarket}); err == nil {
		t.Fatal("unaffordable next market tier was accepted")
	}
	got.Corp.Capital = mathx.New(300)
	got, _, err = e.Apply(got, game.Command{Kind: game.CorpExpandMarket})
	if err != nil || got.Corp.MarketTier != 2 || got.Corp.DemandBP != 12000 || got.Corp.Influence != mathx.New(20) {
		t.Fatalf("second market tier was not applied: state=%#v err=%v", got.Corp, err)
	}
	if _, _, err := e.Apply(got, game.Command{Kind: game.CorpExpandMarket}); err == nil {
		t.Fatal("market expanded beyond catalog tiers")
	}
}
