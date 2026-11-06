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

func TestFreshCorporationCampaignEndToEnd(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	e := sim.New(catalog, 23)
	s := sim.NewCorporationState(23)
	apply := func(command game.Command) {
		var commandErr error
		s, _, commandErr = e.Apply(s, command)
		if commandErr != nil {
			t.Fatal(commandErr)
		}
	}
	apply(game.Command{Kind: game.CorpBuildAsset, ID: "transit_hosting"})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 20})
	apply(game.Command{Kind: game.CorpBuildAsset, ID: "security_desk"})
	apply(game.Command{Kind: game.CorpBuyICE, Quantity: 2})
	apply(game.Command{Kind: game.CorpEnableMaintenance})
	if len(s.Corp.Assets) != 3 || s.Corp.ICECapacity != 8 || s.Corp.ICE != 5 || s.Corp.Capital < 0 {
		t.Fatalf("corporate investment state invalid: %#v", s.Corp)
	}
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 9})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	if s.ActiveBattle == nil || s.Corp.AttackCounter != 0 {
		t.Fatalf("active incursion was not generated: battle=%#v corp=%#v", s.ActiveBattle, s.Corp)
	}
	apply(game.Command{Kind: game.CorpDeployICE, Quantity: 2})
	for s.ActiveBattle != nil {
		apply(game.Command{Kind: game.AdvanceTicks, Ticks: 1})
	}
	if s.LastBattle == nil || (s.LastBattle.Outcome != "defended" && s.LastBattle.Outcome != "breach") {
		t.Fatalf("invalid Corporation battle result: %#v", s.LastBattle)
	}
	if len(s.EventLog) == 0 {
		t.Fatal("Corporation E2E produced no event archive")
	}
	if s.Corp.Capital < 0 || s.Corp.ReputationBP < 0 || s.Corp.ReputationBP > 10000 || s.Corp.ExposureBP < 0 || s.Corp.ExposureBP > 10000 {
		t.Fatalf("corporation meter invariant failed: %#v", s.Corp)
	}
	if s.Corp.ICE != 8 || s.Corp.ICECapacity != 8 || s.Corp.MaintenanceSpent <= 0 || s.Corp.MaintenanceSpent > s.Corp.MaintenanceBudget {
		t.Fatalf("maintenance reserve accounting failed: ice=%d spent=%d budget=%d", s.Corp.ICE, s.Corp.MaintenanceSpent, s.Corp.MaintenanceBudget)
	}
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 10})
	apply(game.Command{Kind: game.CorpRunResearch})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 25})
	apply(game.Command{Kind: game.CorpExpandMarket})
	if s.Corp.MarketTier != 1 || s.Corp.DemandBP != 11000 || s.Corp.Influence <= 0 {
		t.Fatalf("catalogued market expansion did not create strategic growth: %#v", s.Corp)
	}
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 50})
	if s.Corp.Research <= 0 || s.Corp.Influence <= 0 || s.Corp.DemandBP <= 8000 || s.Corp.RunValuation < 1000 {
		t.Fatalf("corporate growth systems did not progress: %#v", s.Corp)
	}
	apply(game.Command{Kind: game.CorpRestructure})
	if s.PrestigeCount != 1 || s.Corp.PatentShares != 1 || len(s.Corp.Assets) != 1 || s.Corp.ExposureBP != 0 {
		t.Fatalf("restructuring did not reset/retain correctly: state=%#v", s)
	}
}

func TestCorporationResearchTechnologyEndToEnd(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	e := sim.New(catalog, 96)
	s := sim.NewCorporationState(96)
	s.Corp.Capital = mathx.New(500)
	apply := func(command game.Command) {
		var commandErr error
		s, _, commandErr = e.Apply(s, command)
		if commandErr != nil {
			t.Fatal(commandErr)
		}
	}
	baseRate := e.CorporationGrossRate(s)
	apply(game.Command{Kind: game.CorpRunResearch})
	apply(game.Command{Kind: game.CorpRunResearch})
	if len(s.Corp.CompletedResearch) != 1 || s.Corp.CompletedResearch[0] != "network_efficiency" {
		t.Fatalf("technology did not unlock through R&D commands: %#v", s.Corp)
	}
	if e.CorporationGrossRate(s) != baseRate+1 {
		t.Fatalf("technology effect missing from economic rate: before=%d after=%d", baseRate, e.CorporationGrossRate(s))
	}
}

func TestCorporationSixtyMinuteScriptedProgressionRemainsBounded(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "assets", "content", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	e := sim.New(catalog, 97)
	s := sim.NewCorporationState(97)
	s.Corp.Capital = mathx.New(500)
	s.Corp.Compute = mathx.New(100)
	apply := func(command game.Command) {
		var commandErr error
		s, _, commandErr = e.Apply(s, command)
		if commandErr != nil {
			t.Fatal(commandErr)
		}
	}
	apply(game.Command{Kind: game.CorpBuildAsset, ID: "transit_hosting"})
	apply(game.Command{Kind: game.CorpBuildAsset, ID: "security_desk"})
	apply(game.Command{Kind: game.CorpBuildAsset, ID: "archive_server"})
	apply(game.Command{Kind: game.CorpConfigurePolicy, ID: "research", Quantity: 1000})
	apply(game.Command{Kind: game.CorpConfigurePolicy, ID: "compute_reserve", Quantity: 5})
	apply(game.Command{Kind: game.CorpConfigurePolicy, ID: "maintenance", Quantity: 8})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 3600})
	if s.LogicalTick != 3600 || len(s.EventLog) > sim.EventLogLimit {
		t.Fatalf("60-minute Corporation script exceeded bounds: tick=%d events=%d", s.LogicalTick, len(s.EventLog))
	}
	if s.Corp == nil || s.Corp.Capital < 0 || s.Corp.Compute < 0 || s.Corp.Research < 0 || s.Corp.Influence < 0 || s.Corp.ICE < 0 || s.Corp.ICE > s.Corp.ICECapacity {
		t.Fatalf("60-minute Corporation resources violated invariants: %#v", s.Corp)
	}
	if len(s.Corp.Assets) != 4 || s.Corp.DemandBP < 0 || s.Corp.DemandBP > 10000 || s.Corp.ReputationBP < 0 || s.Corp.ReputationBP > 10000 || s.Corp.ExposureBP < 0 || s.Corp.ExposureBP > 10000 {
		t.Fatalf("60-minute Corporation progression violated campaign bounds: %#v", s.Corp)
	}
	if s.Corp.Research <= 0 || len(s.Corp.CompletedResearch) == 0 || s.Corp.RunValuation <= 0 {
		t.Fatalf("60-minute Corporation script produced no strategic progress: %#v", s.Corp)
	}
}
