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
	if len(s.Corp.Assets) != 3 || s.Corp.ICE != 5 || s.Corp.Capital < 0 {
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
	if s.Corp.ICE <= 0 || s.Corp.MaintenanceSpent <= 0 || s.Corp.MaintenanceSpent > s.Corp.MaintenanceBudget {
		t.Fatalf("maintenance reserve accounting failed: ice=%d spent=%d budget=%d", s.Corp.ICE, s.Corp.MaintenanceSpent, s.Corp.MaintenanceBudget)
	}
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 10})
	apply(game.Command{Kind: game.CorpRunResearch})
	apply(game.Command{Kind: game.AdvanceTicks, Ticks: 25})
	apply(game.Command{Kind: game.CorpExpandMarket})
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
