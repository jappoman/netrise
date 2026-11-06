# Changelog

All notable changes to NETRISE are documented here. The project is an active
prototype; version `0.1.0` is a development milestone, not a public release.

## [0.1.0] — development milestone

### Added

- Deterministic offline-first Runner and Corporation campaigns on one shared
  simulation engine.
- Fixed-unit economy, seeded automatic conflict resolution, recon, doctrines,
  automation, research, prestige and safe capped offline production.
- Keyboard-first Bubble Tea terminal UI with contextual guidance, event archive,
  accessibility modes and 80×24/narrow-terminal handling.
- Versioned atomic saves with migrations, backup recovery, semantic validation,
  corruption tests and faction-scoped process locking.
- Headless `simulate`, `playtest`, `balance` and `validate-content` commands.
- Cross-platform CI and standalone release-bundle verification for Windows,
  Linux and macOS targets, including catalog packaging and SHA-256 checksums.

### Verification

- Go unit, integration, fuzz, race and benchmark coverage is maintained in the
  repository.
- Player-facing Runner and Corporation flows have been exercised through the
  terminal UI and deterministic headless E2E paths.
- Numerical balance values and final product/legal review remain provisional;
  see [docs/09-balancing.md](docs/09-balancing.md) and
  [docs/11-release.md](docs/11-release.md).
