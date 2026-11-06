package balance

import (
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/sim"
)

const (
	DefaultSeedFrom = 1
	DefaultSeedTo   = 100
	MaxSeedCount    = 1000
)

var commitments = []int64{10, 20, 30, 40}
var doctrines = []game.Doctrine{game.Balanced, game.Stealth, game.Overload}

type Report struct {
	SeedFrom int64                   `json:"seed_from"`
	SeedTo   int64                   `json:"seed_to"`
	Attempts int64                   `json:"attempts"`
	Results  []StrategyTargetSummary `json:"results"`
}

type StrategyTargetSummary struct {
	TargetID             string        `json:"target_id"`
	Doctrine             game.Doctrine `json:"doctrine"`
	Attempts             int64         `json:"attempts"`
	Victories            int64         `json:"victories"`
	Defeats              int64         `json:"defeats"`
	Timeouts             int64         `json:"timeouts"`
	AverageDurationTicks int64         `json:"average_duration_ticks"`
	AverageBotsLost      int64         `json:"average_bots_lost"`
	AverageHeatAddedBP   int64         `json:"average_heat_added_bp"`
}

func Run(catalog content.Catalog, seedFrom, seedTo int64) (Report, error) {
	if seedFrom < 1 || seedTo < seedFrom || seedTo-seedFrom+1 > MaxSeedCount {
		return Report{}, fmt.Errorf("seed range must be 1..N and contain at most %d seeds", MaxSeedCount)
	}
	accessible := catalog
	for i := range accessible.Targets {
		accessible.Targets[i].Reachable = true
	}
	report := Report{SeedFrom: seedFrom, SeedTo: seedTo}
	for _, target := range accessible.Targets {
		for _, doctrine := range doctrines {
			summary := StrategyTargetSummary{TargetID: target.ID, Doctrine: doctrine}
			for seed := seedFrom; ; seed++ {
				for _, commitment := range commitments {
					result, err := resolve(accessible, target.ID, seed, commitment, doctrine)
					if err != nil {
						return Report{}, err
					}
					summary.Attempts++
					report.Attempts++
					summary.AverageDurationTicks += result.ElapsedTicks
					summary.AverageBotsLost += result.BotsLost
					summary.AverageHeatAddedBP += result.HeatAddedBP
					switch result.Outcome {
					case "victory":
						summary.Victories++
					case "defeat":
						summary.Defeats++
					case "timeout":
						summary.Timeouts++
					default:
						return Report{}, fmt.Errorf("target %s produced unknown outcome %q", target.ID, result.Outcome)
					}
				}
				if seed == seedTo {
					break
				}
			}
			summary.AverageDurationTicks /= summary.Attempts
			summary.AverageBotsLost /= summary.Attempts
			summary.AverageHeatAddedBP /= summary.Attempts
			report.Results = append(report.Results, summary)
		}
	}
	return report, nil
}

func resolve(catalog content.Catalog, targetID string, seed, commitment int64, doctrine game.Doctrine) (game.BattleResult, error) {
	engine := sim.New(catalog, seed)
	state := sim.NewState(seed)
	state.Upgrades = []string{"secondhand_compiler"}
	state.Resources.Bots = commitment
	state.Resources.BotCapacity = commitment
	var err error
	state, _, err = engine.Apply(state, game.Command{Kind: game.StartBattle, ID: targetID, Quantity: commitment, Doctrine: doctrine})
	if err != nil {
		return game.BattleResult{}, fmt.Errorf("start %s seed %d commitment %d doctrine %s: %w", targetID, seed, commitment, doctrine, err)
	}
	target, ok := findTarget(catalog, targetID)
	if !ok {
		return game.BattleResult{}, fmt.Errorf("balance target %q disappeared from catalog", targetID)
	}
	for ticks := int64(0); state.ActiveBattle != nil && ticks <= target.TimeLimit; ticks++ {
		state, _, err = engine.Apply(state, game.Command{Kind: game.AdvanceTicks, Ticks: 1})
		if err != nil {
			return game.BattleResult{}, fmt.Errorf("advance %s seed %d commitment %d doctrine %s: %w", targetID, seed, commitment, doctrine, err)
		}
	}
	if state.ActiveBattle != nil || state.LastBattle == nil {
		return game.BattleResult{}, fmt.Errorf("target %s did not resolve within %d ticks", targetID, target.TimeLimit)
	}
	return *state.LastBattle, nil
}

func findTarget(catalog content.Catalog, id string) (content.Target, bool) {
	for _, target := range catalog.Targets {
		if target.ID == id {
			return target, true
		}
	}
	return content.Target{}, false
}
