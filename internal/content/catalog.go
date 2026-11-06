package content

import (
	"encoding/json"
	"fmt"
	"io"
)

type Generator struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Cost             int64  `json:"cost"`
	ComputePerSecond int64  `json:"compute_per_second"`
	BotCapacity      int64  `json:"bot_capacity"`
}
type Upgrade struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Cost        int64  `json:"cost"`
	Description string `json:"description"`
}
type ResearchNode struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	CostData      int64    `json:"cost_data"`
	Prerequisites []string `json:"prerequisites"`
	Description   string   `json:"description"`
}
type CorpAssetDefinition struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CostCapital    int64  `json:"cost_capital"`
	CapitalPerTick int64  `json:"capital_per_tick"`
	UpkeepPerTick  int64  `json:"upkeep_per_tick"`
	Integrity      int64  `json:"integrity"`
}
type Layer struct {
	Integrity int64 `json:"integrity"`
	ICE       int64 `json:"ice"`
	Counter   int64 `json:"counter"`
}
type Target struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Corporation      string  `json:"corporation"`
	Reachable        bool    `json:"reachable"`
	Layers           []Layer `json:"layers"`
	LootCredits      int64   `json:"loot_credits"`
	LootData         int64   `json:"loot_data"`
	TimeLimit        int64   `json:"time_limit"`
	RequiresTargetID string  `json:"requires_target_id,omitempty"`
	ReconCost        int64   `json:"recon_cost"`
	ReconTicks       int64   `json:"recon_ticks"`
	IntelGain        int64   `json:"intel_gain"`
}
type Catalog struct {
	SchemaVersion int                   `json:"schema_version"`
	Generators    []Generator           `json:"generators"`
	Upgrades      []Upgrade             `json:"upgrades"`
	Research      []ResearchNode        `json:"research"`
	CorpAssets    []CorpAssetDefinition `json:"corp_assets"`
	Targets       []Target              `json:"targets"`
}

func Load(r io.Reader) (Catalog, error) {
	var c Catalog
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return c, fmt.Errorf("decode content: %w", err)
	}
	if err := Validate(c); err != nil {
		return c, err
	}
	return c, nil
}
func Validate(c Catalog) error {
	if c.SchemaVersion != 1 {
		return fmt.Errorf("content: unsupported schema version %d", c.SchemaVersion)
	}
	seen := map[string]bool{}
	for _, g := range c.Generators {
		if g.ID == "" || seen[g.ID] || g.Cost < 0 || g.ComputePerSecond < 0 || g.BotCapacity < 0 {
			return fmt.Errorf("content: invalid generator %q", g.ID)
		}
		seen[g.ID] = true
	}
	seen = map[string]bool{}
	for _, u := range c.Upgrades {
		if u.ID == "" || seen[u.ID] || u.Cost < 0 {
			return fmt.Errorf("content: invalid upgrade %q", u.ID)
		}
		seen[u.ID] = true
	}
	seen = map[string]bool{}
	for _, r := range c.Research {
		if r.ID == "" || seen[r.ID] || r.Name == "" || r.CostData < 0 {
			return fmt.Errorf("content: invalid research node %q", r.ID)
		}
		seen[r.ID] = true
	}
	for _, r := range c.Research {
		for _, prerequisite := range r.Prerequisites {
			if !seen[prerequisite] || prerequisite == r.ID {
				return fmt.Errorf("content: research %q has invalid prerequisite %q", r.ID, prerequisite)
			}
		}
	}
	seen = map[string]bool{}
	for _, a := range c.CorpAssets {
		if a.ID == "" || seen[a.ID] || a.Name == "" || a.CostCapital < 0 || a.CapitalPerTick < 0 || a.UpkeepPerTick < 0 || a.Integrity <= 0 {
			return fmt.Errorf("content: invalid corporation asset %q", a.ID)
		}
		seen[a.ID] = true
	}
	seen = map[string]bool{}
	for _, t := range c.Targets {
		if t.ID == "" || seen[t.ID] || t.Name == "" || len(t.Layers) == 0 || t.TimeLimit <= 0 || t.LootCredits < 0 || t.LootData < 0 || t.ReconCost < 0 || t.ReconTicks < 0 || t.IntelGain < 0 {
			return fmt.Errorf("content: invalid target %q", t.ID)
		}
		seen[t.ID] = true
		for _, l := range t.Layers {
			if l.Integrity <= 0 || l.ICE < 0 || l.Counter < 0 {
				return fmt.Errorf("content: invalid layer in target %q", t.ID)
			}
		}
	}
	for _, t := range c.Targets {
		if t.RequiresTargetID != "" && !seen[t.RequiresTargetID] {
			return fmt.Errorf("content: target %q requires unknown target %q", t.ID, t.RequiresTargetID)
		}
		if t.ReconCost > 0 && (t.ReconTicks <= 0 || t.IntelGain <= 0) {
			return fmt.Errorf("content: target %q has incomplete recon definition", t.ID)
		}
	}
	return nil
}
