# NETRISE handoff

This document records the verified state of the prototype so development can
resume without reconstructing the previous session.

## Current state

The repository contains a playable, deterministic single-player prototype with
one shared simulation engine and two asymmetric campaigns:

- Runner: passive economy, fabrication, recon, five routes, doctrines,
  automatic battles, automation, research, event archive and Identity Burn.
- Corporation: capital/services economy, assets, ICE, NPC incursions, repair,
  maintenance, R&D, market expansion, reports and Restructuring.
- TUI: full-screen alternate buffer, keyboard-only controls, contextual
  `NEXT:` guidance, help/tutorial, archive browsing, accessibility modes,
  80x24 support and narrow-terminal wrapping.
- Persistence: versioned atomic saves, migration, backup recovery, semantic
  validation, offline-safe capped reconciliation and faction-scoped locking.
- Distribution: deterministic headless CLI, Windows PowerShell packaging,
  Unix packaging helper, bundled catalog and SHA-256 verification.

The latest synchronized commit is `06ce5d7 Align Unix package CLI verification`.
The working tree is clean and `main` matches `origin/main`.

## Verified checks

The following checks have passed locally on Windows with Go 1.24.2:

```powershell
go test ./...
go vet ./...
go test -race ./...
./scripts/smoke.ps1 -Seed 17 -Ticks 600
./scripts/package-windows.ps1 -OutputDirectory $env:TEMP/netrise-package
bash -n scripts/package-unix.sh
```

The smoke flow covers help, content validation, Runner/Corporation playtests,
balance and 600-tick simulations. The Windows package flow builds and extracts
a standalone archive, validates the bundled catalog and checksum, runs the
headless commands without Go, and confirms unknown commands fail.

Real PTY sessions have also verified Runner and Corporation launch, battle or
incursion playback, archive/report navigation, repair, `q`/Ctrl+C exit,
alternate-screen restoration and valid JSON saves.

## How to test the game manually

From `E:/repos/netrise` in PowerShell:

```powershell
./scripts/smoke.ps1
go run ./cmd/netrise --data-dir ./tmp/playtest-runner
go run ./cmd/netrise corp --data-dir ./tmp/playtest-corp
```

Use separate data directories for the two factions. Quit with `q` or Ctrl+C;
the save is written under the selected data directory. In Runner, follow the
overview `NEXT:` hint: collect contracts, buy the compiler, fabricate bots,
inspect targets, run recon and launch a doctrine-based breach. In Corporation,
build Transit Hosting, install Security Desk, buy ICE, observe the active
incursion, then repair the damaged asset and inspect Report/Archive.

Useful deterministic reports:

```powershell
go run ./cmd/netrise playtest --faction runner --seed 17
go run ./cmd/netrise playtest --faction corp --seed 17
go run ./cmd/netrise balance --seed-from 1 --seed-to 100
go run ./cmd/netrise simulate --faction runner --seed 17 --ticks 600
go run ./cmd/netrise simulate --faction corp --seed 17 --ticks 600
```

Accessibility/environment checks can be repeated with `NO_COLOR=1`,
`NETRISE_HIGH_CONTRAST=1`, `NETRISE_REDUCED_MOTION=1` and `NETRISE_ASCII=1`.

## Known limits and next resumption point

- Unix package runtime still needs an actual Linux/macOS host; this Windows
  session only performed Bash syntax validation for the Unix helper. The CI
  workflow is configured to run the extracted host-native bundle on Unix.
- Human blind-playtest feedback is still required before changing provisional
  balance values. The 100-seed matrix currently shows distinct doctrine
  profiles and requires higher commitments for advanced targets.
- Final product name, license/IP review, screenshots and commercial release
  decisions remain intentionally unresolved.
- The Steam host is explicitly future work (M6) and must not fork the engine.

When resuming, first run `git status --short --branch`, read this file and
`docs/10-roadmap.md`, then start with the outstanding Unix-host CI result or a
recorded human playtest. Keep the shared engine and deterministic command/event
boundaries intact.
