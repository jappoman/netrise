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

Runtime amounts use a scale of 1,000 fixed units, with no floating-point
accountancy. Starter balance values remain external JSON hypotheses.
