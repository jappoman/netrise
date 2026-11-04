package game

import "github.com/jappoman/netrise/internal/mathx"

const CurrentSchema = 1

type Faction string

const Runner Faction = "runner"

type Resources struct {
	Credits          mathx.Amount `json:"credits"`
	Compute          mathx.Amount `json:"compute"`
	Data             mathx.Amount `json:"data"`
	Bots             int64        `json:"bots"`
	BotCapacity      int64        `json:"bot_capacity"`
	ComputeRemainder mathx.Amount `json:"compute_remainder"`
}

type Generator struct {
	ID    string `json:"id"`
	Count int64  `json:"count"`
}
type Fabrication struct {
	RemainingTicks int64 `json:"remaining_ticks"`
	Quantity       int64 `json:"quantity"`
}

type State struct {
	SchemaVersion int          `json:"schema_version"`
	GameVersion   string       `json:"game_version"`
	CampaignID    string       `json:"campaign_id"`
	Faction       Faction      `json:"faction"`
	WorldSeed     int64        `json:"world_seed"`
	RNGState      uint64       `json:"rng_state"`
	LogicalTick   int64        `json:"logical_tick"`
	SavedAtUTC    string       `json:"saved_at_utc"`
	Resources     Resources    `json:"resources"`
	Generators    []Generator  `json:"generators"`
	Upgrades      []string     `json:"upgrades"`
	Fabrication   *Fabrication `json:"fabrication,omitempty"`
}

type CommandKind string

const (
	AdvanceTicks     CommandKind = "advance_ticks"
	CollectContract  CommandKind = "collect_contract"
	BuyGenerator     CommandKind = "buy_generator"
	QueueFabrication CommandKind = "queue_fabrication"
)

type Command struct {
	Kind     CommandKind
	Ticks    int64
	ID       string
	Quantity int64
}

type Event struct {
	Tick    int64  `json:"tick"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Amount  int64  `json:"amount,omitempty"`
}

type Reason string

const (
	ReasonInvalidCommand      Reason = "invalid_command"
	ReasonInsufficientCredits Reason = "insufficient_credits"
	ReasonInsufficientCompute Reason = "insufficient_compute"
	ReasonCapacityReached     Reason = "capacity_reached"
	ReasonDuplicatePurchase   Reason = "duplicate_purchase"
)

type CommandError struct {
	Reason Reason
	Detail string
}

func (e *CommandError) Error() string { return string(e.Reason) + ": " + e.Detail }
