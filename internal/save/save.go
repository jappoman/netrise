package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jappoman/netrise/internal/game"
	"github.com/jappoman/netrise/internal/mathx"
	"os"
	"path/filepath"
	"unicode"
)

var ErrCorrupt = errors.New("save is corrupt")
var ErrInvalidState = errors.New("save state is invalid")

func Path(dir string) string { return filepath.Join(dir, "runner.json") }
func PathForFaction(dir string, faction game.Faction) string {
	if faction == game.Corporation {
		return filepath.Join(dir, "corporation.json")
	}
	return Path(dir)
}
func Write(dir string, s game.State) error {
	return writeAt(dir, PathForFaction(dir, s.Faction), s)
}
func writeAt(dir, path string, s game.State) error {
	if err := validateState(s); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidState, err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".netrise-save-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err != nil {
		_ = f.Close()
		return err
	}
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); err == nil {
		backup := path + ".bak"
		if renameErr := os.Rename(path, backup); renameErr != nil {
			if removeErr := os.Remove(backup); removeErr != nil {
				return fmt.Errorf("rotate save backup: %w", renameErr)
			}
			if renameErr = os.Rename(path, backup); renameErr != nil {
				return fmt.Errorf("rotate save backup: %w", renameErr)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect existing save: %w", err)
	}
	return os.Rename(tmp, path)
}
func Read(dir string) (game.State, error) {
	return ReadFaction(dir, game.Runner)
}
func ReadFaction(dir string, faction game.Faction) (game.State, error) {
	path := PathForFaction(dir, faction)
	b, err := os.ReadFile(path)
	if err != nil {
		return game.State{}, err
	}
	var s game.State
	if err = json.Unmarshal(b, &s); err != nil {
		return recoverBackup(path, s, fmt.Errorf("%w: %v", ErrCorrupt, err))
	}
	if s.SchemaVersion == 0 {
		migrated, migrateErr := migrateV0(s)
		if migrateErr != nil {
			return recoverBackup(path, s, migrateErr)
		}
		return migrated, nil
	}
	if s.SchemaVersion != game.CurrentSchema {
		return recoverBackup(path, s, fmt.Errorf("unsupported save schema %d", s.SchemaVersion))
	}
	if s.Faction == game.Runner && s.Resources.DataCapacity <= 0 {
		s.Resources.DataCapacity = maxInt64(100, s.Resources.Data.Units())
	}
	if err := validateState(s); err != nil {
		return recoverBackup(path, s, fmt.Errorf("%w: %v", ErrCorrupt, err))
	}
	return s, nil
}

func recoverBackup(path string, primary game.State, primaryErr error) (game.State, error) {
	backup, backupErr := os.ReadFile(path + ".bak")
	if backupErr == nil {
		var recovered game.State
		if recoverErr := json.Unmarshal(backup, &recovered); recoverErr == nil && recovered.SchemaVersion == game.CurrentSchema {
			if validateErr := validateState(recovered); validateErr == nil {
				return recovered, nil
			}
		}
	}
	return primary, primaryErr
}

func validateState(s game.State) error {
	if s.SchemaVersion != game.CurrentSchema {
		return fmt.Errorf("unsupported schema %d", s.SchemaVersion)
	}
	if s.Faction != game.Runner && s.Faction != game.Corporation {
		return fmt.Errorf("unknown faction %q", s.Faction)
	}
	if !safeSaveText(s.GameVersion) || !safeSaveText(s.CampaignID) || !safeSaveText(s.Automation.ReconTargetID) {
		return errors.New("save metadata contains control characters")
	}
	if s.LogicalTick < 0 || s.Resources.Credits < 0 || s.Resources.Compute < 0 || s.Resources.Data < 0 || s.Resources.ComputeRemainder < 0 || int64(s.Resources.ComputeRemainder) >= mathx.Scale {
		return errors.New("negative tick/resource or invalid compute remainder")
	}
	if s.Resources.Bots < 0 || s.Resources.BotCapacity < 0 || s.Resources.Bots > s.Resources.BotCapacity || s.Resources.DataCapacity < 0 || s.Resources.DataCapacity > mathx.MaxUnits || s.Resources.Data.Units() > s.Resources.DataCapacity || s.HeatBP < 0 || s.HeatBP > 10000 {
		return errors.New("resource, capacity, or heat bounds exceeded")
	}
	if s.RunExtractedData < 0 || s.GhostKeys < 0 || s.PrestigeCount < 0 {
		return errors.New("negative progression value")
	}
	if s.OfflineSummary != nil && (s.OfflineSummary.Ticks < 0 || s.OfflineSummary.ComputeProduced < 0 || s.OfflineSummary.CapitalProduced < 0) {
		return errors.New("invalid offline summary")
	}
	if s.Automation.ContractBudget < 0 || s.Automation.ContractEarned < 0 || s.Automation.FabricationBudget < 0 || s.Automation.FabricationSpent < 0 || s.Automation.ReconBudget < 0 || s.Automation.ReconSpent < 0 {
		return errors.New("negative Runner automation budget or usage")
	}
	seenGenerators := make(map[string]struct{}, len(s.Generators))
	for _, generator := range s.Generators {
		if !requiredSaveText(generator.ID) || generator.Count <= 0 {
			return errors.New("invalid generator state")
		}
		if _, exists := seenGenerators[generator.ID]; exists {
			return errors.New("duplicate generator state")
		}
		seenGenerators[generator.ID] = struct{}{}
	}
	for _, id := range s.Upgrades {
		if !requiredSaveText(id) {
			return errors.New("invalid upgrade state")
		}
	}
	for _, id := range s.Research {
		if !requiredSaveText(id) {
			return errors.New("invalid research state")
		}
	}
	for _, id := range s.CompletedTargets {
		if !requiredSaveText(id) {
			return errors.New("invalid completed target state")
		}
	}
	for id, intel := range s.Intel {
		if !requiredSaveText(id) || intel < 0 || intel > 100 {
			return errors.New("intel bounds exceeded")
		}
	}
	for id, cooldown := range s.TargetCooldowns {
		if !requiredSaveText(id) || cooldown < 0 {
			return errors.New("target cooldown bounds exceeded")
		}
	}
	for id, access := range s.Access {
		if !requiredSaveText(id) || access < 0 {
			return errors.New("access bounds exceeded")
		}
	}
	if s.Recon != nil && (!requiredSaveText(s.Recon.TargetID) || s.Recon.RemainingTicks < 0) {
		return errors.New("invalid recon job")
	}
	if s.Fabrication != nil && (s.Fabrication.Quantity <= 0 || s.Fabrication.RemainingTicks <= 0 || s.Fabrication.Quantity > s.Resources.BotCapacity-s.Resources.Bots) {
		return errors.New("invalid fabrication job")
	}
	if s.ActiveBattle != nil {
		b := s.ActiveBattle
		if !requiredSaveText(b.TargetID) || !safeSaveText(b.DefenderAssetID) || !safeSaveText(string(b.Doctrine)) || b.CommittedBots < 0 || b.LiveBots < 0 || b.LiveBots > b.CommittedBots || (s.Faction == game.Runner && b.CommittedBots > s.Resources.BotCapacity-s.Resources.Bots) || b.LayerIndex < 0 || b.LayerIntegrity < 0 || b.DefenderICE < 0 || b.TraceBP < 0 || b.TraceBP > 10000 || b.ElapsedTicks < 0 || b.TimeLimit <= 0 || b.ElapsedTicks > b.TimeLimit || b.StartingHeatBP < 0 || b.StartingHeatBP > 10000 {
			return errors.New("invalid active battle")
		}
	}
	if s.LastBattle != nil {
		result := s.LastBattle
		if !requiredSaveText(result.TargetID) || !safeSaveText(string(result.Doctrine)) || !requiredSaveText(result.Outcome) || result.CommittedBots < 0 || result.Survivors < 0 || result.Survivors > result.CommittedBots || result.BotsLost < 0 || result.BotsLost > result.CommittedBots || result.BotsLost != result.CommittedBots-result.Survivors || result.CreditsLooted < 0 || result.DataLooted < 0 || result.TraceBP < 0 || result.TraceBP > 10000 || result.HeatBeforeBP < 0 || result.HeatBeforeBP > 10000 || result.HeatAfterBP < 0 || result.HeatAfterBP > 10000 || result.HeatAddedBP < 0 || result.HeatAddedBP > 10000 || result.ElapsedTicks < 0 {
			return errors.New("invalid battle result")
		}
	}
	if len(s.EventLog) > game.EventLogLimit {
		return errors.New("event log exceeds limit")
	}
	for _, event := range s.EventLog {
		if event.Tick < 0 || event.Tick > s.LogicalTick || !requiredSaveText(event.Type) || !requiredSaveText(event.Message) {
			return errors.New("invalid event log")
		}
	}
	if s.Corp != nil {
		c := s.Corp
		if c.Capital < 0 || c.Compute < 0 || c.Research < 0 || c.Influence < 0 || c.ICE < 0 || c.ICECapacity < 0 || c.ICE > c.ICECapacity || c.ReputationBP < 0 || c.ReputationBP > 10000 || c.ExposureBP < 0 || c.ExposureBP > 10000 || c.DemandBP < 0 || c.DemandBP > 20000 || c.AttackCounter < 0 || c.IncursionBudget < 0 || c.IncursionSpent < 0 || c.IncursionSpent > c.IncursionBudget || c.MaintenanceBudget < 0 || c.MaintenanceSpent < 0 || c.ResearchBudget < 0 || c.ResearchSpent < 0 || c.ComputeReserve < 0 || c.RunValuation < 0 || c.PatentShares < 0 || c.MarketTier < 0 {
			return errors.New("corporation bounds exceeded")
		}
		for _, id := range c.CompletedResearch {
			if !requiredSaveText(id) {
				return errors.New("invalid Corporation research state")
			}
		}
		for _, asset := range c.Assets {
			if !requiredSaveText(asset.ID) || !requiredSaveText(asset.Name) || asset.Integrity < 0 || asset.Integrity > mathx.MaxUnits || asset.MaxIntegrity < 0 || asset.MaxIntegrity > mathx.MaxUnits || (asset.MaxIntegrity > 0 && asset.Integrity > asset.MaxIntegrity) || (asset.MaxIntegrity > 0 && asset.Integrity == 0 && asset.Operational) || asset.CapitalPerTick < 0 || asset.CapitalPerTick > mathx.MaxUnits || asset.UpkeepPerTick < 0 || asset.UpkeepPerTick > mathx.MaxUnits || asset.IsolationCostCapital < 0 || asset.IsolationCostCapital > mathx.MaxUnits || asset.IsolationTicks < 0 || asset.IsolatedUntilTick < 0 {
				return errors.New("invalid corporation asset")
			}
		}
	}
	return nil
}

func safeSaveText(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func requiredSaveText(value string) bool { return value != "" && safeSaveText(value) }

func migrateV0(s game.State) (game.State, error) {
	s.SchemaVersion = game.CurrentSchema
	if s.Resources.DataCapacity <= 0 {
		s.Resources.DataCapacity = maxInt64(100, s.Resources.Data.Units())
	}
	if s.Faction == "" {
		s.Faction = game.Runner
	}
	if err := validateState(s); err != nil {
		return s, fmt.Errorf("%w: %v", ErrCorrupt, err)
	}
	return s, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
