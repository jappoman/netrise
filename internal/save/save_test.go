package save

import (
	"errors"
	"github.com/jappoman/netrise/internal/game"
	"os"
	"testing"
)

func TestRoundTripAndCorruption(t *testing.T) {
	d := t.TempDir()
	s := game.State{SchemaVersion: 1, Faction: game.Runner}
	if err := Write(d, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read(d)
	if err != nil || got.SchemaVersion != 1 {
		t.Fatal(got, err)
	}
	if err := os.WriteFile(Path(d), []byte("{"), 0600); err == nil {
		_, err = Read(d)
		if err == nil {
			t.Fatal("corruption accepted")
		}
	}
}

func TestRoundTripPreservesProgressionState(t *testing.T) {
	d := t.TempDir()
	s := game.State{SchemaVersion: game.CurrentSchema, Faction: game.Runner, LogicalTick: 4, Research: []string{"hardware_01"}, RunExtractedData: 120, GhostKeys: 2, Automation: game.Automation{Contracts: true}, Intel: map[string]int64{"t_01": 45}, EventLog: []game.Event{{Tick: 4, Type: "ResourceProduced", Message: "saved"}}}
	if err := Write(d, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read(d)
	if err != nil || got.RunExtractedData != 120 || got.GhostKeys != 2 || !got.Automation.Contracts || got.Intel["t_01"] != 45 || len(got.Research) != 1 || len(got.EventLog) != 1 {
		t.Fatalf("progression state was not preserved: %#v, %v", got, err)
	}
}

func TestCorporationUsesSeparateSaveFile(t *testing.T) {
	d := t.TempDir()
	s := game.State{SchemaVersion: game.CurrentSchema, Faction: game.Corporation, Corp: &game.CorpState{Capital: 100}}
	if err := Write(d, s); err != nil {
		t.Fatal(err)
	}
	if PathForFaction(d, game.Corporation) == Path(d) {
		t.Fatal("campaign save paths collided")
	}
	got, err := ReadFaction(d, game.Corporation)
	if err != nil || got.Faction != game.Corporation || got.Corp == nil {
		t.Fatalf("corporation save was not isolated: %#v, %v", got, err)
	}
	if _, err := Read(d); err == nil {
		t.Fatal("corporation save leaked into runner save")
	}
}

func TestCorruptPrimaryRecoversLastValidBackup(t *testing.T) {
	d := t.TempDir()
	first := game.State{SchemaVersion: game.CurrentSchema, Faction: game.Runner, LogicalTick: 10}
	second := first
	second.LogicalTick = 20
	if err := Write(d, first); err != nil {
		t.Fatal(err)
	}
	if err := Write(d, second); err != nil {
		t.Fatal(err)
	}
	third := second
	third.LogicalTick = 30
	if err := Write(d, third); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(d), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := Read(d)
	if err != nil || got.LogicalTick != second.LogicalTick {
		t.Fatalf("backup recovery failed: %#v, %v", got, err)
	}
}

func TestSemanticallyInvalidPrimaryRecoversLastValidBackup(t *testing.T) {
	d := t.TempDir()
	first := game.State{SchemaVersion: game.CurrentSchema, Faction: game.Runner, LogicalTick: 10}
	second := first
	second.LogicalTick = 20
	if err := Write(d, first); err != nil {
		t.Fatal(err)
	}
	if err := Write(d, second); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(d), []byte(`{"schema_version":1,"faction":"runner","logical_tick":-1}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := Read(d)
	if err != nil || got.LogicalTick != first.LogicalTick {
		t.Fatalf("semantic backup recovery failed: %#v, %v", got, err)
	}
}

func TestSemanticallyInvalidSaveIsRejected(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(Path(d), []byte(`{"schema_version":1,"faction":"runner","heat_bp":10001}`), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Read(d)
	if err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("semantic corruption was accepted: %v", err)
	}
}

func TestSaveRejectsBotsAboveCapacity(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(Path(d), []byte(`{"schema_version":1,"faction":"runner","resources":{"bots":2,"bot_capacity":1}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("impossible bot balance was accepted: %v", err)
	}
}

func TestSaveRejectsUnrepresentableFixedPointAssetValues(t *testing.T) {
	d := t.TempDir()
	payload := `{"schema_version":1,"faction":"corporation","corp":{"assets":[{"id":"asset","upkeep_per_tick":9223372036854775807}]}}`
	if err := os.WriteFile(PathForFaction(d, game.Corporation), []byte(payload), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFaction(d, game.Corporation); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("unrepresentable asset value was accepted: %v", err)
	}
}

func TestSaveRejectsNegativeRunnerAutomationCounters(t *testing.T) {
	d := t.TempDir()
	payload := `{"schema_version":1,"faction":"runner","automation":{"contracts":true,"contract_budget":100,"contract_earned":-1}}`
	if err := os.WriteFile(Path(d), []byte(payload), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("negative Runner automation usage was accepted: %v", err)
	}
}

func TestSaveRejectsInvalidBattleAndJobSnapshots(t *testing.T) {
	d := t.TempDir()
	invalidBattle := `{"schema_version":1,"faction":"runner","logical_tick":4,"last_battle":{"target_id":"t_01","outcome":"victory","committed_bots":2,"survivors":3,"bots_lost":0}}`
	if err := os.WriteFile(Path(d), []byte(invalidBattle), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("invalid battle result was accepted: %v", err)
	}
	invalidJob := `{"schema_version":1,"faction":"runner","logical_tick":4,"fabrication":{"remaining_ticks":0,"quantity":1}}`
	if err := os.WriteFile(Path(d), []byte(invalidJob), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("invalid fabrication job was accepted: %v", err)
	}
	invalidBattleState := `{"schema_version":1,"faction":"runner","resources":{"bots":1,"bot_capacity":2},"active_battle":{"target_id":"t_01","committed_bots":2,"live_bots":2,"layer_integrity":1,"time_limit":10,"starting_heat_bp":10001}}`
	if err := os.WriteFile(Path(d), []byte(invalidBattleState), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("invalid active battle state was accepted: %v", err)
	}
}

func TestInvalidGeneratorStateIsRejected(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(Path(d), []byte(`{"schema_version":1,"faction":"runner","generators":[{"id":"rig","count":-1}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("invalid generator state was accepted: %v", err)
	}
}

func TestDuplicateGeneratorStateIsRejected(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(Path(d), []byte(`{"schema_version":1,"faction":"runner","generators":[{"id":"rig","count":1},{"id":"rig","count":2}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("duplicate generator state was accepted: %v", err)
	}
}

func TestTemporaryAccessRoundTripsAndInvalidAccessIsRejected(t *testing.T) {
	d := t.TempDir()
	s := game.State{SchemaVersion: game.CurrentSchema, Faction: game.Runner, Access: map[string]int64{"t_01": 120}}
	if err := Write(d, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read(d)
	if err != nil || got.Access["t_01"] != 120 {
		t.Fatalf("temporary access did not round-trip: %#v, %v", got.Access, err)
	}
	if err := os.WriteFile(Path(d), []byte(`{"schema_version":1,"faction":"runner","access":{"t_01":-1}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("negative temporary access was accepted: %v", err)
	}
}

func TestRoundTripPreservesActiveBattleForResume(t *testing.T) {
	d := t.TempDir()
	s := game.State{SchemaVersion: game.CurrentSchema, Faction: game.Runner, LogicalTick: 17, Resources: game.Resources{Bots: 2, BotCapacity: 20}, ActiveBattle: &game.Battle{TargetID: "t_01", AttackerFaction: game.Runner, CommittedBots: 8, LiveBots: 6, LayerIndex: 1, LayerIntegrity: 12, DefenderICE: 3, TraceBP: 900, ElapsedTicks: 4, TimeLimit: 90}}
	if err := Write(d, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read(d)
	if err != nil || got.ActiveBattle == nil || *got.ActiveBattle != *s.ActiveBattle || got.LogicalTick != s.LogicalTick {
		t.Fatalf("active battle was not resumable: %#v, %v", got, err)
	}
}
func TestMigrate(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(Path(d), []byte(`{"faction":"runner"}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Read(d)
	if err != nil || s.SchemaVersion != 1 {
		t.Fatal(s, err)
	}
}

func TestCorporationMarketProgressionRoundTripsAndLegacyDefaults(t *testing.T) {
	d := t.TempDir()
	s := game.State{SchemaVersion: game.CurrentSchema, Faction: game.Corporation, Corp: &game.CorpState{
		Capital: 9000, DemandBP: 12000, Influence: 20000, MarketTier: 2,
	}}
	if err := Write(d, s); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFaction(d, game.Corporation)
	if err != nil || got.Corp == nil || got.Corp.MarketTier != 2 || got.Corp.DemandBP != 12000 || got.Corp.Influence != 20000 {
		t.Fatalf("Corporation market progression did not round-trip: %#v, %v", got, err)
	}

	legacyDir := t.TempDir()
	if err := os.WriteFile(PathForFaction(legacyDir, game.Corporation), []byte(`{"schema_version":1,"faction":"corporation","corp":{"capital":1000,"demand_bp":10000}}`), 0600); err != nil {
		t.Fatal(err)
	}
	legacy, err := ReadFaction(legacyDir, game.Corporation)
	if err != nil || legacy.Corp == nil || legacy.Corp.MarketTier != 0 {
		t.Fatalf("legacy Corporation save did not default market tier: %#v, %v", legacy, err)
	}
}

func TestCorporationMarketDemandBoundsAreEnforced(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(PathForFaction(d, game.Corporation), []byte(`{"schema_version":1,"faction":"corporation","corp":{"demand_bp":20001}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFaction(d, game.Corporation); err == nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("oversaturated Corporation save was accepted: %v", err)
	}
}
