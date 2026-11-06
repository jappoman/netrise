# 12. Implementation notes

The current implementation covers the M0 foundation, M1 deterministic economy,
M2 conflict, the M3 Runner progression slice, and the core M4 Corporation
campaign: a Runner state, validated starter content,
deterministic 1 Hz resource production, contracts, generator purchases, bot
fabrication, versioned atomic JSON saves, capped safe offline reconciliation,
a headless report, and a minimal Bubble Tea v2 terminal adapter. The combat
slice now includes five data-driven Corporation targets, bounded seeded damage
variation, doctrine selection, heat, timeout handling, and one atomic battle
reinforcement. Runner progression now includes target-scoped intel, timed
recon, route unlocking after victories, and a tested economy-to-recon-to-battle
vertical slice. The first safe automation policies are now available through
the Autodialer, Packet Scheduler, and Signal Profiler upgrades; they run on
fixed ticks with bounded resource checks and no hostile offline actions.
The progression slice also includes four validated research nodes with
prerequisites, a data-funded research command and TUI panel, plus the
transactional Runner `Identity Burn` prestige reset with Ghost Keys and
run-scoped extracted-data accounting.

Prestige confirmation now quantifies the current run's losses, retained
currency and next-run starting values for both factions. The TUI also
exposes contract collection, Salvaged Rig purchase, one-bot fabrication,
three attack doctrines, a 75% default commitment, one reserve-bot
reinforcement, and persistent save/load. A save records UTC time and
reconciles only safe production on startup, capped at eight hours.
`--data-dir` overrides the default per-user configuration directory.

Retained prestige currency now has mechanical identity: each Ghost Key adds
two starting Runner bot-capacity slots, while every two Patent Shares add one
starting Corporation ICE-capacity slot. The reset preview exposes both
bonuses, and extreme retained values are clamped before arithmetic.

The first Corporation slice is now playable through `netrise corp`: it has a
separate campaign save, capital-producing client assets, reserve ICE,
reputation/exposure meters, and a deterministic active-play NPC intrusion.
Corporation emergency ICE uses the shared battle state and resolution path,
while the outcome applies Corporation-specific defense or exposure effects.
The Corporation slice now also includes capital-funded R&D, market demand and
influence expansion, upkeep-aware automated maintenance, and a separate
`Restructuring` prestige that retains Patent Shares.

The initial M5 quality layer is in place: `NO_COLOR` removes ANSI styling,
terminals below 80×24 receive a compact safe view, CI verifies Linux/Windows/
macOS builds, and Make targets expose Runner, Corporation, simulation, build,
vet, test, and race workflows.
`NETRISE_HIGH_CONTRAST=1` provides an explicit bold white-on-black title rail
for low-contrast terminals, while `NO_COLOR` remains the higher-priority mode.
`NETRISE_REDUCED_MOTION=1` resolves an active battle in one UI update while
preserving the same one-second engine ticks and deterministic result.
Save reads now recover a valid `.bak` snapshot when the primary JSON is
corrupted. A simulation benchmark covers 10,000 logical ticks, and tag-based
release automation builds Linux, Windows, and macOS artifacts with SHA-256
checksums. Each release artifact is a self-contained `.zip` or `.tar.gz`
bundle containing the executable and catalog. Headless reports support both campaigns through
`simulate --faction runner|corp`, with deterministic output verified for each.
Save reads and writes also validate semantic state bounds at the persistence
boundary; syntactically valid but impossible JSON is rejected or recovered from
the last valid backup.
Offline load reconciliation now accepts an injected clock at the CLI boundary,
so elapsed-time behavior and clock reversal handling are deterministic in tests
while normal play continues to use the system clock.
When no battle, timer, automation or policy job is active, offline passive
production now aggregates directly over elapsed ticks instead of iterating each
second; equivalence tests keep the result identical to the authoritative tick
path, while queued safe jobs retain the detailed reconciliation path.
Seed fuzz harnesses cover arbitrary save payloads and unknown command kinds,
checking that malformed inputs cannot panic or partially mutate a snapshot.
Corporation emergency actions issued from the TUI now transition directly from
Battle to Result when they resolve an incursion; keyboard tests cover isolation
and repair through the same engine command path used in play. The Result screen
also explains `isolated` as a successful containment outcome and points to
asset recovery rather than presenting an unknown battle result. The Corporation
report repeats each asset's integrity and operational state, so the economic
projection screen also exposes the operational risk that can reduce revenue.
The catalog loader first checks the working directory and then the executable
bundle, so packaged binaries can run from any current directory. Release
artifacts include `assets/content/catalog.json` beside the binary.
Player-facing events are retained in a bounded 200-entry archive, persisted in
saves and available from the TUI with `[l] archive`.
The content validator rejects missing links and cyclic research prerequisites;
an integration script exercises 3,600 Runner ticks with automation enabled and
asserts bounded resources, event history and recon progress.
The CLI also exposes `validate-content`, and the TUI `?` view explains the
current objective and available actions separately for Runner and Corporation.
Offline reconciliation now persists a safe-production digest (ticks and
resource gains) shown in Overview, while active battles and hostile
Corporation attack counters remain frozen during closed time.

Research nodes now carry validated, data-driven effects in the content catalog.
The initial Runner branch affects compute production, breach power, stealth heat,
and safe automation cadence; the Research screen displays each node's effect.
The simulation aggregates only completed nodes, so the same catalog data drives
headless and TUI behavior without faction-specific UI mutations.

Runner upgrades now also expose validated catalog effects. Salvaged Clock and
Heat Sink Array affect passive compute and battle heat; Proxy Ledger improves
stealth power; Stolen Memory Bank increases capacity; Packet Scheduler and
Signal Profiler shorten safe jobs; Black Archive discounts research. Existing
upgrade IDs remain the source of unlock predicates, while numeric bonuses are
aggregated by the engine.

The Runner TUI now has a keyboard-only Upgrade screen (`[g]`) with selection,
purchase feedback, installed/affordability states, catalog descriptions and
effect text. Secondary screens remain reachable below 80x24; only the overview
and active battle use the compact dashboard.

Runner preflight now supports 25/50/75/100% commitment selection with `+/-` or
left/right, shows the selected bot count and a deterministic baseline duration
and heat projection, and launches exactly the selected commitment. Active recon
is surfaced in the overview so the next useful action is visible while intel
is being collected.

Active battles now use a dedicated playback view instead of falling back to the
overview. It shows target, doctrine, layer progress, integrity, live/committed
bots, ICE, trace, heat, elapsed limit and the last six battle events, with
faction-specific intervention controls for Runner and Corporation.

Battle results now persist doctrine, committed bots, trace, heat before/after
and heat delta. The Result screen turns those fields into an explicit reason,
bot accounting, trace/heat summary, loot, duration and a next-action hint.

The Target screen now exposes route prerequisites, corporation identity, intel
confidence, recon cost/duration and whether recon is ready, active, occupied or
complete. The same screen action starts recon through the engine command path;
the TUI never mutates the simulation directly.

The `[h]` progression guide now derives a concise first-loop checklist from
live Runner or Corporation state. Completed steps, the current next step and
later locked steps are shown separately, with actionable keys and no duplicate
progress state outside the simulation/save model.

Runner automation now persists cumulative credit budgets and usage counters for
contracts, fabrication and recon. Enabling a policy with `Command.Quantity` set
configures its cap; zero retains an existing cap or uses the 1,000-credit
default. Reaching a cap emits `AutomationStopped` and disables only that policy.
The TUI `[A]` screen exposes the three policies, preset budgets and spent totals.

Corporation `MaintenanceAuto` now has the same persisted budget semantics. During
active ticks it safely restocks one reserve ICE unit at a time, emits
`MaintenanceRestocked`, and stops with `MaintenanceStopped` when the capital
budget is exhausted. Offline reconciliation never spends that budget; the TUI
Corporation maintenance panel exposes the cap, usage and reserve state.

The same Corporation policy panel now configures automated R&D capital budget
and a compute reserve threshold. R&D spends 200 capital and 1 compute per active
study, never consumes the protected compute reserve, and emits explicit
configuration/completion/stop events.

Corporation now also has a derived Economic Report (`[v]`) showing demand-scaled
gross revenue, active upkeep, operating margin, compute flow, policy states,
reserve ICE and next-tick capital/compute projections. It stores no duplicate
economy values and is available only through the shared state/TUI read path.

Corporation R&D now completes validated threshold technologies from the
`corp_research` catalog. Network Efficiency adds capital per tick and Lean
Operations reduces operational upkeep; completed IDs are persisted, displayed
in the Corporation report and reset by Restructuring. Rate calculations are
exposed by the shared engine so the TUI cannot drift from simulation rules.

The Corporation overview now exposes all three initial business assets:
`[b]` Transit Hosting, `[a]` Archive Server and `[s]` Security Desk. The
progression guide and contextual next-step hint identify the Archive Server as
the next expansion after security investment, while Runner's `[a]` action
remains its tutorial breach.

Security Desk now carries a validated `ice_capacity_bonus` from the catalog.
Building it expands Corporation ICE capacity and, when the maintenance policy
was filling the previous rack, expands the protected reserve atomically too.
Upkeep-throttled Corporation assets now recover automatically once accumulated
capital can cover the active costs; `AssetReactivated` is archived and the TUI
shows the operational asset ratio plus a recovery hint.

Runner target loot now respects catalog-defined cooldowns. A resolved raid
stores the target cooldown before returning loot, a repeated launch is rejected
atomically with `target_cooldown`, and logical ticks expire the lock. Recon and
recon automation respect the same lock. Targets and preflight display the
remaining cooldown so the next useful action is explicit rather than an
unexplained failed command.

The integration suite now covers a deterministic 3,600-tick Corporation
campaign with all starter assets, security, maintenance and automated R&D. It
checks resource and meter bounds, event archive limits and research progress,
giving the Corporation loop the same long-run regression coverage as Runner.
It also saves and reopens an active Runner battle from the real catalog, then
asserts byte-equivalent state and event results after deterministic resume.

Corporation gross revenue now derives from demand, reputation and exposure in
the shared engine. Reputation below 50% and exposure above zero apply bounded
integer penalties, so the Economic Report and active production expose the same
operating-risk tradeoff without floating-point accountancy.
The report now prints reputation, exposure and the effective revenue factor so
the penalty is explainable to the player rather than hidden in the rate.

Corporation market expansion is now a data-driven progression rather than a
repeatable fixed-cost action. Catalogued tiers persist through `MarketTier`, add
demand above the initial 100% service baseline, grant influence and valuation,
and later tiers consume catalogued influence as well as capital. They stop
atomically when the next tier is unaffordable or the catalog is exhausted.
Overview, compact mode, report and next-step hints show the next market name,
costs and demand gain.
Save tests cover round-tripping market tiers and demand above 100%, defaulting
the new tier field for legacy Corporation JSON and rejecting impossible demand
values at the persistence boundary.
CI now runs the player-facing CLI smoke flow for both factions on every platform
job. The release workflow also executes the Linux amd64 binary from its bundled
directory, outside the repository working directory, before publishing the
archive and checksum.

Prestige confirmation screens now expose their lock reason before confirmation:
Runner Identity Burn states the 10,000-data threshold, while Corporation
Restructuring states the 1,000-valuation threshold and the resulting Patent
Shares. Once ready, the same screen shows an explicit confirmation status and
the Corporation overview points to `[p]` as the next useful action.

Corporation incursions now support the second specified emergency response:
`[x] isolate asset`. The selected highest-value operational asset is suspended
for its data-driven catalog duration, capital is charged atomically, the
incursion ends as `isolated`, and an `AssetIsolationEnded` event marks automatic
recovery. Reserve ICE and isolation share the single intervention budget.
Incursions also apply catalog `asset_damage` to that selected asset; zero
integrity disables it and `[r]` repairs a chosen damaged asset at one capital
per integrity, restoring a recoverable Corporation security loop. Legacy saves
without `max_integrity` migrate that ceiling from the validated asset catalog.

The integration suite now runs a deterministic combat-balance smoke sweep across
all five targets, three doctrines, four commitment levels and seeds 1..100. It
checks time-limit resolution, valid outcomes, bot/heat accounting and
non-negative resources for every case without bypassing the production engine.

Battle playback controls now support pause with `[space]` and 1x/2x/4x display
speed. These controls live entirely in the TUI: each update still advances the
shared engine through deterministic `AdvanceTicks` commands, and paused playback
cannot mutate simulation state.

Counterattack attrition now follows the conflict specification's doctrine
profiles: Stealth 75%, Balanced 100% and Overload 130% of baseline losses,
accumulated deterministically across counter ticks. A 100-seed integration
matrix verifies that heat and bot-loss profiles remain distinct.

Runtime amounts use a scale of 1,000 fixed units, with no floating-point
accountancy. `mathx.NewChecked` rejects unit values that cannot be represented;
the legacy `mathx.New` constructor now panics instead of silently wrapping, and
content validation rejects out-of-range amount costs before simulation.
`MulUnits` now checks all signed operand directions, including the `MinInt64`
edge case. Starter balance values remain external JSON hypotheses.

Runner data is now capacity-limited: the starter state has 100 data slots,
`Local Data Cache` adds 100 catalogued slots, and successful loot records only
the amount that fits. `DataCapacity` is persisted, legacy saves default it
safely, and Overview/compact mode show `DATA current/capacity`.
Corporation incursions are now catalog-driven and budget-constrained: interval, attacker commitment, and total bot budget live in content; the persistent spent counter prevents unlimited active-play attacks while offline reconciliation remains non-hostile. Legacy saves with missing budget fields use the documented default budget.
Runner victories now grant catalogued temporary access tokens. Adjacent routes require an active token, tokens expire only during active play, and expiration is observable through `AccessGranted`/`AccessExpired` events and the target browser; legacy content without an access duration retains its previous completed-route behavior.
Save writes now use unique permissioned temporary files and report backup-rotation failures instead of silently continuing with an uncertain snapshot.

The Runner integration suite now proves two distinct first-breach strategies
end to end: a lower-attrition Stealth commitment and a higher-pressure
Overload commitment both complete the initial target through the shared
deterministic battle engine.

CI packaging checks now exercise both factions from the Linux release bundle,
while Windows jobs also run the player-facing CLI smoke flow through both
PowerShell's default runner and CMD.
Linux CI additionally installs Fish and runs the same two-faction smoke flow
under Fish, covering the documented shell hosts without changing game logic.

The TUI suite now includes a real keyboard-driven Runner flow on the shipped
catalog: Overview to Targets, Preflight, Battle playback and Result, including
the player-facing outcome and next-action rail.

It also covers the Corporation keyboard loop on the shipped catalog: security
investment, active NPC incursion, emergency asset isolation and the resulting
player-facing resolution screen.

Corporation active-incursion saves now round-trip without importing Runner-only
data capacity, and deterministic resume is covered through the complete battle
resolution path.

Long-run integration coverage now exercises 21,600 active ticks for both
campaigns with automation enabled, checking six-hour-equivalent resource,
event-log and progression invariants.

Faction boundaries are now enforced centrally by the command engine: Runner
operations cannot mutate Corporation state and Corporation operations cannot
mutate Runner state. The TUI also keeps Corporation navigation out of the
Runner target and research screens.

Secondary TUI screens now all expose an explicit `NEXT:` rail, including Help,
Archive, Research, prestige confirmation and the Corporation report, so the
player can always return to an actionable campaign state without guessing.
Empty battle/result states and Corporation recovery/error panels carry the
same rail, so this remains true after a resume or an unavailable state.

Offline reconciliation now persists whether the eight-hour safety cap was
applied, and Runner/Corporation overviews show that cap warning alongside the
safe production digest.

The Runner integration suite now walks all five catalog targets in order,
using each victory's temporary access token to reach the next route and
verifying the final target without bypassing engine reachability checks.

Generator counts are validated at the save boundary, including duplicate-ID
rejection, and passive production uses saturating fixed-point addition. This
prevents malformed state from double-counting a generator or wrapping a
resource to zero on the next tick.

Content validation now rejects cyclic target-route prerequisites as well as
missing links, preventing catalog-authored progression softlocks.

Active play now recovers Runner heat and Corporation exposure by 25 basis
points per safe tick when no battle is running. Hostile/offline reconciliation
does not add a new adverse action or resolve a battle while closed.

Target lists and Runner preflight now show a deterministic ICE estimate range
derived from target layers and current intel. The range narrows to the exact
catalog value at 100% intel, keeping uncertainty visible without exposing
hidden defense values prematurely.

The same intel-driven range is now shown for catalogued credit and data loot,
so preflight communicates both security risk and expected reward without a
false precision claim.

Content validation also bounds layer integrity, ICE and counter values to the
fixed-point range before they reach combat arithmetic or preflight estimates.

Corporation valuation gains now use saturating arithmetic across offline
aggregation, active revenue, R&D and market expansion, so a valid upper-bound
save cannot wrap its prestige progress back to a negative value.

Prestige counter accumulation and its TUI preview use the same saturating
progression arithmetic, keeping Ghost Keys, Patent Shares and lifetime reset
counts stable even at persisted integer limits.

Battle and preflight damage calculations now use checked non-negative
multiply/divide helpers and safe ceiling division, preventing extreme but
validated content or resumed battle state from overflowing intermediate
values.

Corporation gross revenue and upkeep now use the same saturating additions and
checked multiply/divide path when aggregating assets and applying demand,
reputation and exposure factors.

Logical ticks, fabrication/intel gains, extracted data, demand and aggregated
technology effects also use bounded increments, preserving capacity and meter
invariants when resuming unusually large but structurally valid state.

The Runner research panel now names missing prerequisite IDs beside locked
nodes, keeping progression choices self-explanatory without requiring players
to infer the research graph.

Research nodes with satisfied prerequisites now distinguish affordable data
costs from `NEED N DATA`, so an unavailable purchase explains its resource
shortfall before the player presses Enter.

The Corporation market summary now exposes capital and influence shortfalls
for the next expansion directly in the overview, before an acquisition key is
pressed.

Automation budget guards now compare spend and cost without adding first, so
maximum-size persisted budgets cannot wrap and authorize work beyond their
configured cap.

Corporation compute-reserve policy values are bounded to the fixed-point range,
and automated R&D compares available compute directly against the reserve
without an overflow-prone `reserve + 1` expression.

Research and Upgrade panels now show the current data/capacity or credit
balance above their catalog rows, making affordability visible before a
purchase attempt.

Catalog validation now bounds generator, upgrade and Corporation asset
capacity values to the fixed-point range; capacity mutations and incursion ICE
composition use saturating additions at runtime as well.

Attack-power bonuses from research and upgrades are now bounded in content and
combined with saturating arithmetic before combat or preflight multiplication.

Runner battle withdrawals, victories and timeouts now return surviving bots with
capacity-aware saturating arithmetic; extreme persisted balances cannot wrap and
returned units cannot exceed the configured bot capacity.

Battle elapsed time, trace and heat meters, Corporation attack counters and
incursion totals now saturate at their valid upper bounds; extreme resumed
states cannot wrap these counters into negative or early-reset values.

The save boundary now rejects Runner snapshots with more reserve bots than the
configured capacity, preventing semantically impossible balances from entering
the simulation even when their JSON is syntactically valid.

The same boundary now rejects non-representable data capacities and Corporation
asset economic, integrity and isolation-cost values before they reach fixed-point
arithmetic paths.

Generator purchases now reject a saturated generator count before charging the
player, preventing progression-counter wraparound and preserving command atomicity.

The Corporation Economic Report now uses saturating projections for capital and
compute, and overflow-safe budget deductions, so extreme but valid state values
remain readable in the player-facing TUI.

Target ICE and loot estimates now saturate both multi-layer aggregation and
uncertainty multiplication, keeping the Target screen safe for extreme catalog
values without changing the underlying simulation.

Runner automation budgets and cumulative usage counters are now validated at the
save boundary, matching Corporation policy validation and rejecting negative
state before a policy can resume.

Persisted fabrication jobs, battle results and event archive entries now also
have semantic checks for positive timers, coherent bot accounting, bounded
meters and logical timestamps before they are resumed by the TUI.

Active Runner battles now validate their committed units against reserve capacity
and starting heat, while persisted fabrication quantities must fit remaining bot
capacity before a save can be reopened.

Offline reconciliation summaries are also checked for non-negative ticks and
resource deltas, so the Overview cannot display fabricated offline gains after a
corrupted save.

The 200-entry event archive limit is now defined in the shared game model and
enforced at both engine append and save-load boundaries.

Runner preflight commitment previews now avoid multiplying the full bot reserve
before dividing by percentage, keeping the player-facing commitment selector safe
for maximum-size persisted reserves.

Event archive text is now rejected at the save boundary when it contains control
characters, preventing untrusted JSON from injecting terminal escape sequences
into the player-facing TUI.

The same control-character policy now applies to data-driven catalog IDs, names,
descriptions and target labels before content can reach any TUI screen.

Save validation extends that policy to persisted target, asset, research and
generator labels, closing indirect TUI injection paths through resumed state.

The headless `simulate --help` path now exits successfully after printing its
flags, matching the top-level CLI help behavior instead of reporting an error.

The `corp --help` alias now follows the same clean help path and never starts a
TUI session when a user is asking for command usage.

The TUI now exposes a keyboard-accessible Settings screen (`[S]`) for both
campaigns. It reports the active `NO_COLOR`, high-contrast, reduced-motion and
ASCII fallback modes and explains how to configure them before relaunching,
while keeping all simulation state and deterministic behavior in the shared
engine. `NETRISE_ASCII=1` replaces non-ASCII content at the final render
boundary for terminals without Unicode support.

Offline reconciliation is now written immediately at the CLI load boundary with
an updated UTC timestamp. Reopening a save before the first TUI autosave cannot
grant the same elapsed interval twice; regression tests cover the full
Runner and Corporation load/reopen paths.

Top-level CLI validation now rejects unknown commands and extra arguments to
`validate-content` instead of falling through to an interactive TUI session.

Integration coverage now exercises the full eight-hour offline cap against the
shipped catalog for both factions, asserting bounded resources and no hostile
events or state changes while the game is closed.

Interactive campaigns now acquire a faction-scoped advisory session lock before
loading state. Unix builds use `Flock` and Windows builds use `LockFileEx`, so a
second process receives a clear error while independent Runner and Corporation
saves remain available; the lock is released when the TUI exits. The save test
suite also holds a lock in a separate helper process to verify the OS-level
exclusion rather than only same-process handles.

The headless `playtest` command now runs the deterministic first-loop script
for either faction and emits a compact JSON baseline (progress tick, battle
outcome, events and faction-specific progression). It reuses the production
engine and catalog, giving balance iterations a repeatable comparison point
without adding telemetry or a second gameplay implementation.

The headless commands now reject unexpected positional arguments instead of
silently ignoring them. `make playtest` runs both scripted campaign baselines,
while `make smoke` exercises help, content validation, both playtests and both
600-tick simulations as a local release smoke flow.

The Windows CI smoke matrix now exercises both CMD and PowerShell, and packaged
release smoke checks cover top-level, Corporation and simulation help before
running both faction simulations.

The player-facing Bubble Tea view now opts into the terminal alternate buffer,
so the shipped TUI is genuinely full-screen and automatically restores the
previous terminal contents when the program exits.

Long Runner catalog menus now use a height-aware scrolling window. At the
standard 80x24 size the Upgrades screen keeps the selection, action rail and
`NEXT:` guidance visible while `[up/down]` moves through the complete catalog.
Overview action rails also wrap at a safe terminal width, keeping all
faction-specific keyboard actions visible instead of clipping the right side.
