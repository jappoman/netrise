# NETRISE: coding agent instructions

Read **README.md** and **all docs/ files** before major architectural work. This is a design-first repository. Never mistake proposals for implemented features.

## Non-negotiable product constraints
- Retain original **Corporation vs Runner** asymmetry; no generic conquest or card-game clone.
- Native full-screen terminal UI and a single shared simulation engine. A future Steam wrapper must reuse the same engine, not fork gameplay.
- Hybrid incremental: passive production + preparation + 2–3 tactical choices + automatically resolved, observable attacks. No hacking puzzles or real intrusion tooling.
- Single-player, offline-first, deterministic simulation; no authentication, cloud account, telemetry, payment, or backend for MVP.
- No destructive offline attacks. Offline elapsed time only advances explicitly safe production/research and capped AI growth.
- Honor accessibility, headless testing and save compatibility.
- Original branding and fiction. Do not copy copyrighted game mechanics text, fictional factions, characters, artwork or trademarks.

## Working rules
1. **Inspect repository first.** Don't overwrite hand-edited files. Prefer small PR-sized commits and tests.
2. Implement **milestones in docs/10-roadmap.md order**. Begin with M0 scaffolding and M1 deterministic headless simulation, not Steam or extra content.
3. Public engine functions must be deterministic for the same state, input and RNG seed. No time.Now(), filesystem, global RNG or terminal I/O inside simulation logic.
4. Represent money/resources with overflow-safe fixed units or explicitly justified large-number arithmetic. Centralize formatting and arithmetic. No float-based currency accountancy.
5. Content values and IDs must be data-driven, schema-validated, versioned and tested.
6. The engine accepts semantic Commands and emits Events; the TUI never mutates internal state directly.
7. Define exact invariants and reason codes. Reject impossible commands atomically; never partially spend resources on validation failures.
8. Use safe atomic save writes and migration tests. Never trust save JSON or user terminal sequences.
9. Use external packages only where clearly valuable; pin versions and run their own supported APIs (Bubble Tea v2 differs from v1).
10. Do not silently change locked design decisions. Record tradeoffs and suggest a documented amendment.
11. Prefer local tests, race tests and benchmarks; add CI for supported OS targets. Never claim a build or test passed unless executed.
12. Keep game copy **English by default** but make all displayed strings localizable later. Write source comments and technical identifiers in English.
13. After each coherent implementation increment, verify it end to end (automated checks plus the relevant player-facing flow), then create a focused Git commit that records the verified progress. Sync the commit to the configured remote when possible; if sync fails, report the exact reason and keep the local commit safe.

## First implementation PR
Produce:
- go.mod, cmd/netrise/main.go, internal/game state/command/events, internal/sim, internal/content, internal/save, internal/tui
- a fixed 1-second simulation tick with elapsed-time reconciliation and seeded RNG
- initial Runner economy, simulated Corp node, one attack flow (per docs/04-conflict.md)
- headless CLI option `netrise simulate --seed 17 --ticks 600` for repeatability
- tests for arithmetic, affordability, combat consistency, determinism, offline cap, migration, corruption and terminal resize
- a simple playable TUI for terminal-only distribution

Check acceptance criteria in docs/10-roadmap.md. First PR can land a narrower functioning vertical slice if explicitly documented.

## Suggested commands (after code exists)
```sh
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
go run ./cmd/netrise
go run ./cmd/netrise simulate --seed 17 --ticks 600
```

## Security and runtime boundaries
Real-world hacking is **fictional**. Never execute shell commands derived from in-game fiction, contact targets, scan real networks, launch subprocesses to simulate attacks, or require elevated privileges. A bundled terminal host should run only the game binary, sandboxed to application data paths.

## Agent handoff checklist
Document changed files, design decisions, player-visible behavior, commands run, test results and outstanding risks. A task is complete only when its acceptance criteria pass or deficiencies are clearly disclosed.
