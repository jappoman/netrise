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
}
