# 07. Technical architecture and data contracts

## Recommended stack
Go (stable currently supported toolchain), Bubble Tea v2, Lip Gloss compatible version, standard library `encoding/json`, `testing` and `math/big` only if numeric limits demand it. Versions and APIs must be verified when implementing.

## Core invariants
Engine runs without terminal, wall clock, shell, network or Steam. It consumes an immutable-ish snapshot + Command + deterministic RNG input and returns updated state + ordered Events + error. User interfaces are adapters only.

### Suggested directories
```text
cmd/netrise/             executable; interactive and headless modes
internal/game/           state, commands, domain events, errors
internal/sim/            tick engine, economy, battles, offline replay
internal/content/        load, validate, version and catalog content
internal/save/           atomic JSON save, migrations, lock policy
internal/tui/            Bubble Tea model, views, keybindings and theme
internal/mathx/          fixed-point safe math and number formatting
assets/content/          JSON game definitions
tests/                   integration, replay fixtures, balancing tests
docs/                    design reference and decisions
```

## State model (conceptual)
```text
GameState {
  schemaVersion, gameVersion, campaignId, faction, worldSeed,
  logicalTick, savedAtUTC, resources, capacities, generators,
  upgrades, research, automationPolicies, worldAssets,
  activeBattle?, metaProgress, eventCursor, rngState
}
WorldAsset {
  id, ownerId, industry, routeIds[], securityLayers[], lootPool,
  lootRegenRate, businessYield, accessRequirements, intelVisibility
}
Battle {
  id, attackerId, defenderAssetId, doctrine, committedBots,
  survivors, layerIndex, remainingIntegrity, defenderICE,
  traceBP, startedTick, elapsedTicks, interventionUsed, rngState
}
```
Use stable IDs, typed enums and explicit versioning, not ad-hoc dynamic maps for rules. Serialize using documented schema field names and snapshot tests. Separate runtime-only UI state from save state.

## Commands and events
Commands (examples): `CollectContract`, `BuyGenerator`, `QueueFabrication`, `StartRecon`, `StartBattle`, `ReinforceBattle`, `WithdrawBattle`, `SetAutomationPolicy`, `AllocateCompute`, `Prestige`, `AdvanceTicks`.
Events: `ResourceProduced`, `ResourceSpent`, `UpgradePurchased`, `ReconCompleted`, `BattleStarted`, `LayerBreached`, `BotsLost`, `BattleEnded`, `AutomationProduced`, `AutomationQueued`, `AutomationStopped`, `MaintenanceRestocked`, `MaintenanceStopped`, `AssetReactivated`, `CorpPolicyConfigured`, `CorpResearch`, `CorpResearchAutomated`, `CorpResearchStopped`, `CorpTechnologyCompleted`, `AccessGranted`, `AccessExpired`, `PrestigeCompleted`, `OfflineReconciled`. Target cooldowns and temporary access are persisted state; cooldowns advance on logical ticks and access expires only during active play.
Each command validates before mutation; return typed error/reason, e.g. `ErrInsufficientBots`, `ErrTargetUnreachable`.

## Tick order and RNG
At every logical second:
1. validate and apply already queued player commands,
2. advance income and consumption,
3. process safe automation jobs according to budgets,
4. process active battle (if any),
5. apply safe meter recovery and bounded NPC actions,
6. emit events, update logicalTick.
RNG seeded from campaign seed and persisted stream/counter. For reproducibility across systems, implement a documented PRNG algorithm, not a version-dependent implicit generator. Never call global random or current time inside deterministic game logic.

## Offline reconciliation
Separate fast-forward production from real-time battle ticking. On load, compute sanitized elapsed seconds using injected Clock service, clamp 0..8h initially, aggregate predictable generator output and eligible tasks. Do not resolve battles or hostile AI operations while closed. Document exactly how mid-battle suspension resumes without consuming offline turns. Save *after* reconciliation, then show resume summary.

## Numeric architecture
Currency/resource stocks: integer fixed point with a documented scale, checked arithmetic and bounded rules. Battle damage, ratios and multipliers: integer basis points with checked intermediate multiplication. Large incremental values eventually need `math/big`, decimal representation or custom mantissa/exponent; decide before adding content beyond int64 limits. Never silently overflow, go negative or produce NaN/Inf.

## Save handling
- Default OS app data location (`os.UserConfigDir` or suitable XDG/Windows user data path); allow `--data-dir` for testing and portable builds.
- Versioned JSON with explicit schema version; update via migration chain.
- Write temporary file, flush/sync where supported, atomic rename, backup previous valid save; verify on load.
- Graceful handling of corruption, malformed data, future schema, duplicate processes and read-only dirs.
- One campaign save per faction plus optional save slots; separate meta-state clearly.
- No secrets or user-specific personal data collected.

## Execution modes
`netrise`: normal TUI; `netrise simulate --seed 17 --ticks 600`: headless deterministic report for tests/balance; `netrise --version`; `netrise --help`. Optional `netrise validate-content`.
Production build should work without requiring Python/Node, shell scripting, runtime network access or an external DB.

## CI and quality
- Build and test Linux amd64, Windows amd64, macOS amd64/arm64 as supported.
- `go test ./...`, `go vet ./...`, race checks where supported, format checks.
- Golden test for deterministic battle with same seed, headless batch balance suite, fuzz malformed saves/commands, terminal resize tests.
- Performance targets (provisional): TUI input latency below 100ms on ordinary hardware; 10,000 headless ticks under 1s with typical small world; stable memory use in 8h equivalent offline simulation.
- Avoid performance optimizations that sacrifice determinism before benchmarks.

## Architecture decisions needing ADRs
1. Numeric representation for late-game scale.
2. Persistence backup and concurrency policy.
3. Steam window host choice, tested on all targets.
4. Localization and content packaging.
5. PRNG specification and compatibility guarantees.
