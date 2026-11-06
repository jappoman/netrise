package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
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
	if s.Corp.Capital < 0 || s.Corp.ReputationBP < 0 || s.Corp.ReputationBP > 10000 || s.Corp.ExposureBP < 0 || s.Corp.ExposureBP > 10000 {
		t.Fatalf("corporation meter invariant failed: %#v", s.Corp)
	}
	if s.Corp.ICE != 0 {
		t.Fatalf("reserve ICE accounting failed: %d", s.Corp.ICE)
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
