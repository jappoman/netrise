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
