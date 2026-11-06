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

Richer prestige confirmation remains later work. The TUI now
exposes contract collection, Salvaged Rig purchase, one-bot fabrication,
three attack doctrines, a 75% default commitment, one reserve-bot
reinforcement, and persistent save/load. A save records UTC time and
reconciles only safe production on startup, capped at eight hours.
`--data-dir` overrides the default per-user configuration directory.

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
all five targets, three doctrines, four commitment levels and seeds 1..10. It
checks time-limit resolution, valid outcomes, bot/heat accounting and
non-negative resources for every case without bypassing the production engine.

Battle playback controls now support pause with `[space]` and 1x/2x/4x display
speed. These controls live entirely in the TUI: each update still advances the
shared engine through deterministic `AdvanceTicks` commands, and paused playback
cannot mutate simulation state.

Runtime amounts use a scale of 1,000 fixed units, with no floating-point
accountancy. Starter balance values remain external JSON hypotheses.
