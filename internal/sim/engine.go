package sim

import (
	"fmt"
	"github.com/jappoman/netrise/internal/content"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"math"
)

type Engine struct {
	Catalog content.Catalog
	rng     uint64
}

const OfflineCapTicks int64 = 8 * 60 * 60

// ReconcileOffline advances only safe production. Hostile actions and battles
// are intentionally not represented here, so closing the game cannot cause
// destructive outcomes. elapsedTicks is supplied by the persistence adapter.
func (e *Engine) ReconcileOffline(s game.State, elapsedTicks int64) (game.State, int64) {
	if elapsedTicks < 0 {
		return s, 0
	}
	if elapsedTicks > OfflineCapTicks {
		elapsedTicks = OfflineCapTicks
	}
	return e.Advance(s, elapsedTicks), elapsedTicks
}

func New(c content.Catalog, seed int64) *Engine {
	return &Engine{Catalog: c, rng: uint64(seed) + 0x9e3779b97f4a7c15}
}
func (e *Engine) next() uint64 { e.rng ^= e.rng << 7; e.rng ^= e.rng >> 9; return e.rng }
func NewState(seed int64) game.State {
	return game.State{SchemaVersion: game.CurrentSchema, GameVersion: "0.1.0", CampaignID: "runner-01", Faction: game.Runner, WorldSeed: seed, RNGState: uint64(seed) + 0x9e3779b97f4a7c15, Resources: game.Resources{Credits: mathx.New(25), Compute: mathx.New(2), BotCapacity: 20}}
}

func (e *Engine) Apply(s game.State, c game.Command) (game.State, []game.Event, error) {
	if c.Kind == game.AdvanceTicks {
		if c.Ticks < 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "ticks cannot be negative"}
		}
		return e.Advance(s, c.Ticks), nil, nil
	}
	// Validate all spending before changing the snapshot.
	switch c.Kind {
	case game.CollectContract:
		if c.Quantity != 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "contract takes no quantity"}
		}
		n, err := s.Resources.Credits.Add(mathx.New(10))
		if err != nil {
			return s, nil, err
		}
		s.Resources.Credits = n
		return s, []game.Event{{Tick: s.LogicalTick, Type: "ResourceProduced", Message: "Street contract settled.", Amount: 10}}, nil
	case game.BuyGenerator:
		var g *content.Generator
		for i := range e.Catalog.Generators {
			if e.Catalog.Generators[i].ID == c.ID {
				g = &e.Catalog.Generators[i]
			}
		}
		if g == nil {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unknown generator"}
		}
		cost := generatorCost(g.Cost, generatorCount(s, c.ID))
		if s.Resources.Credits < mathx.New(cost) {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: fmt.Sprintf("need %d credits", cost)}
		}
		s.Resources.Credits -= mathx.New(cost)
		upsertGenerator(&s, c.ID)
		s.Resources.BotCapacity += g.BotCapacity
		return s, []game.Event{{Tick: s.LogicalTick, Type: "GeneratorPurchased", Message: g.Name + " online."}}, nil
	case game.QueueFabrication:
		if c.Quantity <= 0 {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "quantity must be positive"}
		}
		if s.Fabrication != nil {
			return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "fabrication already queued"}
		}
		if s.Resources.Bots+c.Quantity > s.Resources.BotCapacity {
			return s, nil, &game.CommandError{Reason: game.ReasonCapacityReached, Detail: "bot capacity reached"}
		}
		computeCost, err := mathx.New(10).MulUnits(c.Quantity)
		if err != nil || s.Resources.Compute < computeCost {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCompute, Detail: "fabrication requires 10 compute and 2 credits"}
		}
		creditCost, err := mathx.New(2).MulUnits(c.Quantity)
		if err != nil || s.Resources.Credits < creditCost {
			return s, nil, &game.CommandError{Reason: game.ReasonInsufficientCredits, Detail: "fabrication requires 2 credits per bot"}
		}
		s.Resources.Compute -= computeCost
		s.Resources.Credits -= creditCost
		s.Fabrication = &game.Fabrication{RemainingTicks: 3, Quantity: c.Quantity}
		return s, []game.Event{{Tick: s.LogicalTick, Type: "FabricationQueued", Message: "Bot compiler queued."}}, nil
	default:
		return s, nil, &game.CommandError{Reason: game.ReasonInvalidCommand, Detail: "unsupported command"}
	}
}
func (e *Engine) Advance(s game.State, ticks int64) game.State {
	for i := int64(0); i < ticks; i++ {
		s.LogicalTick++
		s.RNGState ^= s.RNGState << 7
		s.RNGState ^= s.RNGState >> 9
		rate := int64(2)
		for _, g := range e.Catalog.Generators {
			for _, owned := range s.Generators {
				if owned.ID == g.ID {
					rate += owned.Count * g.ComputePerSecond
				}
			}
		}
		produced := mathx.New(rate)
		s.Resources.Compute, _ = s.Resources.Compute.Add(produced)
		if s.Fabrication != nil {
			s.Fabrication.RemainingTicks--
			if s.Fabrication.RemainingTicks <= 0 {
				s.Resources.Bots += s.Fabrication.Quantity
				s.Fabrication = nil
			}
		}
	}
	return s
}
func generatorCount(s game.State, id string) int64 {
	for _, g := range s.Generators {
		if g.ID == id {
			return g.Count
		}
	}
	return 0
}
func upsertGenerator(s *game.State, id string) {
	for i := range s.Generators {
		if s.Generators[i].ID == id {
			s.Generators[i].Count++
			return
		}
	}
	s.Generators = append(s.Generators, game.Generator{ID: id, Count: 1})
}
func generatorCost(base, n int64) int64 {
	cost := base
	for i := int64(0); i < n; i++ {
		if cost > math.MaxInt64/6 {
			return math.MaxInt64
		}
		cost = (cost*6 + 4) / 5
	}
	return cost
}
