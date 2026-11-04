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
type Catalog struct {
	SchemaVersion int         `json:"schema_version"`
	Generators    []Generator `json:"generators"`
	Upgrades      []Upgrade   `json:"upgrades"`
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
	return nil
}
