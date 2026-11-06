package playtest

import (
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/sim"
)

type Report struct {
	Seed             int64        `json:"seed"`
	Faction          game.Faction `json:"faction"`
	LogicalTick      int64        `json:"logical_tick"`
	Events           int          `json:"events"`
	FirstBattle      string       `json:"first_battle,omitempty"`
	ReconCompletedAt int64        `json:"recon_completed_at,omitempty"`
	BattleStartedAt  int64        `json:"battle_started_at,omitempty"`
	BattleEndedAt    int64        `json:"battle_ended_at,omitempty"`
	CompletedTargets []string     `json:"completed_targets,omitempty"`
	RunnerCredits    int64        `json:"runner_credits_units,omitempty"`
	RunnerBots       int64        `json:"runner_bots,omitempty"`
	CorpAssets       int          `json:"corp_assets,omitempty"`
	CorpMarketTier   int64        `json:"corp_market_tier,omitempty"`
	CorpCapital      int64        `json:"corp_capital_units,omitempty"`
}

func Run(catalog content.Catalog, faction game.Faction, seed int64) (Report, error) {
	e := sim.New(catalog, seed)
	var state game.State
	if faction == game.Corporation {
		state = sim.NewCorporationState(seed)
	} else if faction == game.Runner {
		state = sim.NewState(seed)
	} else {
		return Report{}, fmt.Errorf("unknown faction %q", faction)
	}

	apply := func(command game.Command) error {
		var err error
		state, _, err = e.Apply(state, command)
		return err
	}
	if faction == game.Runner {
		if err := runRunner(apply, &state); err != nil {
			return Report{}, err
		}
	} else if err := runCorporation(apply, &state); err != nil {
		return Report{}, err
	}

	report := Report{Seed: seed, Faction: faction, LogicalTick: state.LogicalTick, Events: len(state.EventLog), FirstBattle: battleOutcome(state), CompletedTargets: append([]string(nil), state.CompletedTargets...)}
	for _, event := range state.EventLog {
		switch event.Type {
		case "ReconCompleted":
			if report.ReconCompletedAt == 0 {
				report.ReconCompletedAt = event.Tick
			}
		case "BattleStarted":
			if report.BattleStartedAt == 0 {
				report.BattleStartedAt = event.Tick
			}
		case "BattleEnded":
			if report.BattleEndedAt == 0 {
				report.BattleEndedAt = event.Tick
			}
		}
	}
	if faction == game.Runner {
		report.RunnerCredits = state.Resources.Credits.Units()
		report.RunnerBots = state.Resources.Bots
	} else if state.Corp != nil {
		report.CorpAssets = len(state.Corp.Assets)
		report.CorpMarketTier = state.Corp.MarketTier
		report.CorpCapital = state.Corp.Capital.Units()
	}
	return report, nil
}

func runRunner(apply func(game.Command) error, state *game.State) error {
	for i := 0; i < 13; i++ {
		if err := apply(game.Command{Kind: game.CollectContract}); err != nil {
			return err
		}
	}
	if err := apply(game.Command{Kind: game.BuyUpgrade, ID: "street_contacts"}); err != nil {
		return err
	}
	for i := 0; i < 7; i++ {
		if err := apply(game.Command{Kind: game.CollectContract}); err != nil {
			return err
		}
	}
	for _, command := range []game.Command{
		{Kind: game.BuyUpgrade, ID: "secondhand_compiler"},
		{Kind: game.AdvanceTicks, Ticks: 4},
		{Kind: game.CollectContract},
		{Kind: game.StartRecon, ID: "t_01"},
		{Kind: game.AdvanceTicks, Ticks: 15},
		{Kind: game.CollectContract},
		{Kind: game.AdvanceTicks, Ticks: 5},
		{Kind: game.QueueFabrication, Quantity: 5},
		{Kind: game.AdvanceTicks, Ticks: 28},
		{Kind: game.CollectContract},
		{Kind: game.CollectContract},
		{Kind: game.QueueFabrication, Quantity: 5},
		{Kind: game.AdvanceTicks, Ticks: 3},
		{Kind: game.StartBattle, ID: "t_01", Quantity: 10, Doctrine: game.Balanced},
	} {
		if err := apply(command); err != nil {
			return err
		}
	}
	for state.ActiveBattle != nil {
		if err := apply(game.Command{Kind: game.AdvanceTicks, Ticks: 1}); err != nil {
			return err
		}
	}
	return nil
}

func runCorporation(apply func(game.Command) error, state *game.State) error {
	for _, command := range []game.Command{
		{Kind: game.CorpBuildAsset, ID: "transit_hosting"},
		{Kind: game.AdvanceTicks, Ticks: 20},
		{Kind: game.CorpBuildAsset, ID: "security_desk"},
		{Kind: game.CorpBuyICE, Quantity: 2},
		{Kind: game.CorpEnableMaintenance},
		{Kind: game.AdvanceTicks, Ticks: 10},
		{Kind: game.CorpDeployICE, Quantity: 2},
	} {
		if err := apply(command); err != nil {
			return err
		}
	}
	for state.ActiveBattle != nil {
		if err := apply(game.Command{Kind: game.AdvanceTicks, Ticks: 1}); err != nil {
			return err
		}
	}
	for _, command := range []game.Command{
		{Kind: game.AdvanceTicks, Ticks: 10},
		{Kind: game.CorpRunResearch},
		{Kind: game.AdvanceTicks, Ticks: 25},
		{Kind: game.CorpExpandMarket},
	} {
		if err := apply(command); err != nil {
			return err
		}
	}
	return nil
}

func battleOutcome(state game.State) string {
	if state.LastBattle == nil {
		return ""
	}
	return state.LastBattle.Outcome
}
