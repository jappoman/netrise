package game

import "github.com/jappoman/netrise/internal/mathx"

const CurrentSchema = 1

type Faction string

const (
	Runner      Faction = "runner"
	Corporation Faction = "corporation"
)

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
type CorpAsset struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	CapitalPerTick       int64  `json:"capital_per_tick"`
	UpkeepPerTick        int64  `json:"upkeep_per_tick"`
	Integrity            int64  `json:"integrity"`
	MaxIntegrity         int64  `json:"max_integrity,omitempty"`
	Operational          bool   `json:"operational"`
	IsolationCostCapital int64  `json:"isolation_cost_capital,omitempty"`
	IsolationTicks       int64  `json:"isolation_ticks,omitempty"`
	IsolatedUntilTick    int64  `json:"isolated_until_tick,omitempty"`
}
type CorpState struct {
	Capital            mathx.Amount `json:"capital"`
	Compute            mathx.Amount `json:"compute"`
	Research           mathx.Amount `json:"research"`
	CompletedResearch  []string     `json:"completed_research,omitempty"`
	Influence          mathx.Amount `json:"influence"`
	ICE                int64        `json:"ice"`
	ICECapacity        int64        `json:"ice_capacity"`
	ReputationBP       int64        `json:"reputation_bp"`
	ExposureBP         int64        `json:"exposure_bp"`
	Assets             []CorpAsset  `json:"assets"`
	AttackCounter      int64        `json:"attack_counter"`
	DemandBP           int64        `json:"demand_bp"`
	MaintenanceAuto    bool         `json:"maintenance_auto"`
	MaintenanceBudget  int64        `json:"maintenance_budget"`
	MaintenanceSpent   int64        `json:"maintenance_spent"`
	MaintenanceReserve int64        `json:"maintenance_reserve"`
	ResearchAuto       bool         `json:"research_auto"`
	ResearchBudget     int64        `json:"research_budget"`
	ResearchSpent      int64        `json:"research_spent"`
	ComputeReserve     int64        `json:"compute_reserve"`
	RunValuation       int64        `json:"run_valuation"`
	PatentShares       int64        `json:"patent_shares"`
	MarketTier         int64        `json:"market_tier"`
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
	TargetCooldowns  map[string]int64 `json:"target_cooldowns,omitempty"`
	Automation       Automation       `json:"automation"`
	Research         []string         `json:"research,omitempty"`
	RunExtractedData int64            `json:"run_extracted_data"`
	GhostKeys        int64            `json:"ghost_keys"`
	PrestigeCount    int64            `json:"prestige_count"`
	Corp             *CorpState       `json:"corp,omitempty"`
	EventLog         []Event          `json:"event_log,omitempty"`
	OfflineSummary   *OfflineSummary  `json:"offline_summary,omitempty"`
}

type OfflineSummary struct {
	Ticks           int64        `json:"ticks"`
	ComputeProduced mathx.Amount `json:"compute_produced"`
	CapitalProduced mathx.Amount `json:"capital_produced"`
}

type Automation struct {
	Contracts         bool   `json:"contracts"`
	Fabrication       bool   `json:"fabrication"`
	ReconTargetID     string `json:"recon_target_id,omitempty"`
	ContractBudget    int64  `json:"contract_budget"`
	ContractEarned    int64  `json:"contract_earned"`
	FabricationBudget int64  `json:"fabrication_budget"`
	FabricationSpent  int64  `json:"fabrication_spent"`
	ReconBudget       int64  `json:"recon_budget"`
	ReconSpent        int64  `json:"recon_spent"`
}

type ReconJob struct {
	TargetID       string `json:"target_id"`
	RemainingTicks int64  `json:"remaining_ticks"`
}

type BattleResult struct {
	TargetID      string   `json:"target_id"`
	Doctrine      Doctrine `json:"doctrine"`
	Outcome       string   `json:"outcome"`
	CommittedBots int64    `json:"committed_bots"`
	Survivors     int64    `json:"survivors"`
	BotsLost      int64    `json:"bots_lost"`
	CreditsLooted int64    `json:"credits_looted"`
	DataLooted    int64    `json:"data_looted"`
	TraceBP       int64    `json:"trace_bp"`
	HeatBeforeBP  int64    `json:"heat_before_bp"`
	HeatAfterBP   int64    `json:"heat_after_bp"`
	HeatAddedBP   int64    `json:"heat_added_bp"`
	ElapsedTicks  int64    `json:"elapsed_ticks"`
}

type Doctrine string

const (
	Balanced Doctrine = "balanced"
	Stealth  Doctrine = "stealth"
	Overload Doctrine = "overload"
)

type Battle struct {
	TargetID         string   `json:"target_id"`
	DefenderAssetID  string   `json:"defender_asset_id,omitempty"`
	AttackerFaction  Faction  `json:"attacker_faction"`
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
	StartingHeatBP   int64    `json:"starting_heat_bp"`
}

type CommandKind string

const (
	AdvanceTicks          CommandKind = "advance_ticks"
	CollectContract       CommandKind = "collect_contract"
	BuyGenerator          CommandKind = "buy_generator"
	QueueFabrication      CommandKind = "queue_fabrication"
	BuyUpgrade            CommandKind = "buy_upgrade"
	StartBattle           CommandKind = "start_battle"
	ReinforceBattle       CommandKind = "reinforce_battle"
	WithdrawBattle        CommandKind = "withdraw_battle"
	StartRecon            CommandKind = "start_recon"
	EnableAutomation      CommandKind = "enable_automation"
	BuyResearch           CommandKind = "buy_research"
	IdentityBurn          CommandKind = "identity_burn"
	CorpBuildAsset        CommandKind = "corp_build_asset"
	CorpBuyICE            CommandKind = "corp_buy_ice"
	CorpDeployICE         CommandKind = "corp_deploy_ice"
	CorpIsolateAsset      CommandKind = "corp_isolate_asset"
	CorpRepairAsset       CommandKind = "corp_repair_asset"
	CorpRunResearch       CommandKind = "corp_run_research"
	CorpExpandMarket      CommandKind = "corp_expand_market"
	CorpEnableMaintenance CommandKind = "corp_enable_maintenance"
	CorpConfigurePolicy   CommandKind = "corp_configure_policy"
	CorpRestructure       CommandKind = "corp_restructure"
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
	ReasonInvalidCommand        Reason = "invalid_command"
	ReasonInsufficientCredits   Reason = "insufficient_credits"
	ReasonInsufficientCompute   Reason = "insufficient_compute"
	ReasonInsufficientData      Reason = "insufficient_data"
	ReasonInsufficientInfluence Reason = "insufficient_influence"
	ReasonCapacityReached       Reason = "capacity_reached"
	ReasonDuplicatePurchase     Reason = "duplicate_purchase"
	ReasonUpgradeLocked         Reason = "upgrade_locked"
	ReasonInsufficientBots      Reason = "insufficient_bots"
	ReasonTargetUnreachable     Reason = "target_unreachable"
	ReasonTargetCooldown        Reason = "target_cooldown"
	ReasonBattleActive          Reason = "battle_active"
	ReasonNoBattle              Reason = "no_battle"
	ReasonReconActive           Reason = "recon_active"
	ReasonReconLocked           Reason = "recon_locked"
	ReasonResearchLocked        Reason = "research_locked"
	ReasonPrestigeLocked        Reason = "prestige_locked"
	ReasonFactionMismatch       Reason = "faction_mismatch"
)

type CommandError struct {
	Reason Reason
	Detail string
}

func (e *CommandError) Error() string { return string(e.Reason) + ": " + e.Detail }
