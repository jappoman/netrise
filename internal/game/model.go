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
	SchemaVersion    int              `json:"schema_version"`
	GameVersion      string           `json:"game_version"`
	CampaignID       string           `json:"campaign_id"`
	Faction          Faction          `json:"faction"`
	WorldSeed        int64            `json:"world_seed"`
	RNGState         uint64           `json:"rng_state"`
	LogicalTick      int64            `json:"logical_tick"`
	SavedAtUTC       string           `json:"saved_at_utc"`
	Resources        Resources        `json:"resources"`
	Generators       []Generator      `json:"generators"`
	Upgrades         []string         `json:"upgrades"`
	Fabrication      *Fabrication     `json:"fabrication,omitempty"`
	HeatBP           int64            `json:"heat_bp"`
	ActiveBattle     *Battle          `json:"active_battle,omitempty"`
	LastBattle       *BattleResult    `json:"last_battle,omitempty"`
	Intel            map[string]int64 `json:"intel,omitempty"`
	Recon            *ReconJob        `json:"recon,omitempty"`
	CompletedTargets []string         `json:"completed_targets,omitempty"`
	Automation       Automation       `json:"automation"`
}

type Automation struct {
	Contracts     bool   `json:"contracts"`
	Fabrication   bool   `json:"fabrication"`
	ReconTargetID string `json:"recon_target_id,omitempty"`
}

type ReconJob struct {
	TargetID       string `json:"target_id"`
	RemainingTicks int64  `json:"remaining_ticks"`
}

type BattleResult struct {
	TargetID      string `json:"target_id"`
	Outcome       string `json:"outcome"`
	Survivors     int64  `json:"survivors"`
	BotsLost      int64  `json:"bots_lost"`
	CreditsLooted int64  `json:"credits_looted"`
	DataLooted    int64  `json:"data_looted"`
	HeatAddedBP   int64  `json:"heat_added_bp"`
	ElapsedTicks  int64  `json:"elapsed_ticks"`
}

type Doctrine string

const (
	Balanced Doctrine = "balanced"
	Stealth  Doctrine = "stealth"
	Overload Doctrine = "overload"
)

type Battle struct {
	TargetID         string   `json:"target_id"`
	Doctrine         Doctrine `json:"doctrine"`
	CommittedBots    int64    `json:"committed_bots"`
	LiveBots         int64    `json:"live_bots"`
	LayerIndex       int      `json:"layer_index"`
	LayerIntegrity   int64    `json:"layer_integrity"`
	DefenderICE      int64    `json:"defender_ice"`
	TraceBP          int64    `json:"trace_bp"`
	ElapsedTicks     int64    `json:"elapsed_ticks"`
	TimeLimit        int64    `json:"time_limit"`
	InterventionUsed bool     `json:"intervention_used"`
}

type CommandKind string

const (
	AdvanceTicks     CommandKind = "advance_ticks"
	CollectContract  CommandKind = "collect_contract"
	BuyGenerator     CommandKind = "buy_generator"
	QueueFabrication CommandKind = "queue_fabrication"
	BuyUpgrade       CommandKind = "buy_upgrade"
	StartBattle      CommandKind = "start_battle"
	ReinforceBattle  CommandKind = "reinforce_battle"
	WithdrawBattle   CommandKind = "withdraw_battle"
	StartRecon       CommandKind = "start_recon"
	EnableAutomation CommandKind = "enable_automation"
)

type Command struct {
	Kind     CommandKind
	Ticks    int64
	ID       string
	TargetID string
	Quantity int64
	Doctrine Doctrine
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
	ReasonUpgradeLocked       Reason = "upgrade_locked"
	ReasonInsufficientBots    Reason = "insufficient_bots"
	ReasonTargetUnreachable   Reason = "target_unreachable"
	ReasonBattleActive        Reason = "battle_active"
	ReasonNoBattle            Reason = "no_battle"
	ReasonReconActive         Reason = "recon_active"
	ReasonReconLocked         Reason = "recon_locked"
)

type CommandError struct {
	Reason Reason
	Detail string
}

func (e *CommandError) Error() string { return string(e.Reason) + ": " + e.Detail }
