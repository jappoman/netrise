# 12. Implementation notes

The current implementation covers M0, the economy portion of M1, M2, and the
first M3 progression slice: a Runner state, validated starter content,
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

The Corporation campaign, broader research effects, and richer prestige
confirmation remain later work. The TUI now
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
Save reads now recover a valid `.bak` snapshot when the primary JSON is
corrupted. A simulation benchmark covers 10,000 logical ticks, and tag-based
release automation builds Linux, Windows, and macOS artifacts with SHA-256
checksums. Headless reports support both campaigns through
`simulate --faction runner|corp`, with deterministic output verified for each.
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

Runtime amounts use a scale of 1,000 fixed units, with no floating-point
accountancy. Starter balance values remain external JSON hypotheses.
