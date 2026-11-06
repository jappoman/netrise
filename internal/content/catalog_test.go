package content

import (
	"strings"
	"testing"
)

func TestLoadRejectsDuplicateIDs(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"generators":[{"id":"x"},{"id":"x"}]}`))
	if err == nil {
		t.Fatal("expected duplicate rejection")
	}
}

func TestValidateRejectsResearchCycles(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"research":[{"id":"a","name":"A","prerequisites":["b"]},{"id":"b","name":"B","prerequisites":["a"]}]}`))
	if err == nil {
		t.Fatal("research cycle accepted")
	}
}

func TestValidateRejectsTargetRouteCycles(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"targets":[{"id":"a","name":"A","requires_target_id":"b","layers":[{"integrity":1}],"time_limit":1},{"id":"b","name":"B","requires_target_id":"a","layers":[{"integrity":1}],"time_limit":1}]}`))
	if err == nil {
		t.Fatal("target route cycle accepted")
	}
}

func TestValidateRejectsNegativeResearchEffects(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"research":[{"id":"a","name":"A","effects":{"compute_per_second_bonus":-1}}]}`))
	if err == nil {
		t.Fatal("negative research effect accepted")
	}
}

func TestValidateRejectsInvalidUpgradeEffects(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"upgrades":[{"id":"u","effects":{"heat_reduction_percent":101}}]}`))
	if err == nil {
		t.Fatal("invalid upgrade effect accepted")
	}
}

func TestValidateRejectsInvalidCorporationResearch(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"corp_research":[{"id":"efficiency","name":"Efficiency","threshold":0}]}`))
	if err == nil {
		t.Fatal("zero-threshold Corporation research accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"corp_research":[{"id":"lean","name":"Lean","threshold":10,"effects":{"upkeep_reduction_percent":101}}]}`))
	if err == nil {
		t.Fatal("over-cap Corporation research effect accepted")
	}
}

func TestValidateRejectsNegativeSecurityCapacityBonus(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"corp_assets":[{"id":"security","name":"Security","integrity":100,"ice_capacity_bonus":-1}]}`))
	if err == nil {
		t.Fatal("negative ICE capacity bonus accepted")
	}
}

func TestValidateRejectsNegativeIsolationPolicy(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"corp_assets":[{"id":"archive","name":"Archive","integrity":100,"isolation_cost_capital":-1}]}`))
	if err == nil {
		t.Fatal("negative isolation cost accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"corp_assets":[{"id":"archive","name":"Archive","integrity":100,"isolation_ticks":-1}]}`))
	if err == nil {
		t.Fatal("negative isolation duration accepted")
	}
}

func TestValidateRejectsNegativeTargetCooldown(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"targets":[{"id":"target","name":"Target","layers":[{"integrity":1}],"time_limit":1,"cooldown_ticks":-1}]}`))
	if err == nil {
		t.Fatal("negative target cooldown accepted")
	}
}

func TestValidateRejectsNegativeAssetDamage(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"targets":[{"id":"target","name":"Target","layers":[{"integrity":1}],"time_limit":1,"asset_damage":-1}]}`))
	if err == nil {
		t.Fatal("negative asset damage accepted")
	}
}

func TestValidateRejectsInvalidCorporationMarketExpansion(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"corp_markets":[{"id":"metro","name":"Metro","cost_capital":0,"demand_bonus_bp":1000}]}`))
	if err == nil {
		t.Fatal("zero-cost Corporation market expansion accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"corp_markets":[{"id":"metro","name":"Metro","cost_capital":10,"demand_bonus_bp":-1}]}`))
	if err == nil {
		t.Fatal("negative Corporation market demand bonus accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"corp_markets":[{"id":"metro","name":"Metro","cost_capital":10,"demand_bonus_bp":1000,"influence_cost":-1}]}`))
	if err == nil {
		t.Fatal("negative Corporation market influence cost accepted")
	}
}

func TestValidateRejectsUnrepresentableAmountCosts(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"generators":[{"id":"rig","cost":9223372036854775807}]}`))
	if err == nil {
		t.Fatal("generator cost beyond fixed-point range accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"targets":[{"id":"target","name":"Target","layers":[{"integrity":1}],"time_limit":1,"loot_credits":9223372036854775807}]}`))
	if err == nil {
		t.Fatal("target loot beyond fixed-point range accepted")
	}
}

func TestValidateRejectsUnrepresentableCombatValues(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"targets":[{"id":"target","name":"Target","layers":[{"integrity":9223372036854775807}]}]}`))
	if err == nil {
		t.Fatal("layer integrity beyond fixed-point range accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"targets":[{"id":"target","name":"Target","layers":[{"integrity":1,"ice":9223372036854775807}]}]}`))
	if err == nil {
		t.Fatal("layer ICE beyond fixed-point range accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"targets":[{"id":"target","name":"Target","layers":[{"integrity":1,"counter":9223372036854775807}]}]}`))
	if err == nil {
		t.Fatal("layer counter beyond fixed-point range accepted")
	}
}

func TestValidateRejectsNegativeDataCapacity(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"generators":[{"id":"cache","name":"Cache","data_capacity":-1}]}`))
	if err == nil {
		t.Fatal("negative generator data capacity accepted")
	}
}

func TestValidateRejectsUnrepresentableCapacityValues(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"generators":[{"id":"rig","bot_capacity":9223372036854775807}]}`))
	if err == nil {
		t.Fatal("generator bot capacity beyond fixed-point range accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"upgrades":[{"id":"u","effects":{"bot_capacity_bonus":9223372036854775807}}]}`))
	if err == nil {
		t.Fatal("upgrade bot capacity beyond fixed-point range accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"corp_assets":[{"id":"security","name":"Security","integrity":100,"ice_capacity_bonus":9223372036854775807}]}`))
	if err == nil {
		t.Fatal("Corporation ICE capacity beyond fixed-point range accepted")
	}
}

func TestValidateRejectsNegativeAccessDuration(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"targets":[{"id":"target","name":"Target","layers":[{"integrity":1}],"time_limit":1,"access_ticks":-1}]}`))
	if err == nil {
		t.Fatal("negative access duration accepted")
	}
}

func TestValidateRejectsInvalidCorporationIncursion(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"corp_incursion":{"interval_ticks":0,"bot_budget":8,"bot_commitment":9}}`))
	if err == nil {
		t.Fatal("incursion commitment above budget accepted")
	}
	_, err = Load(strings.NewReader(`{"schema_version":1,"corp_incursion":{"interval_ticks":-1}}`))
	if err == nil {
		t.Fatal("negative incursion interval accepted")
	}
}
