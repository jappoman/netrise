# 10. Roadmap and issue-ready implementation plan

Prioritize **a functioning, fun vertical slice** over large amounts of generic infrastructure. A milestone is complete only when its tests and player-facing acceptance criteria pass. Estimates intentionally omitted.

## M0: repository and build skeleton
**Tasks**
1. Initialize Go module, README alignment and license decision note; pin Bubble Tea v2/Lip Gloss versions.
2. Implement `netrise --help`, `--version`, `--data-dir`, headless mode routing.
3. Define domain types, interfaces and content schema; create a minimal validated sample catalog.
4. Add CI for lint/vet/test/build on initial platforms.

**Acceptance:** `go test ./...` passes, `go run ./cmd/netrise --help` works, headless and interactive modes launch without network, invalid content produces clear errors.

## M1: deterministic economy
**Tasks**
1. State and event model, command validation, fixed-point numeric utilities.
2. 1Hz logical clock and deterministic seeded RNG, no wall clock in engine.
3. Runner credits, compute, fabrication, bot caps, 3 generators and first 3 upgrades.
4. Headless `simulate --seed --ticks` with reproducible structured report.
5. JSON save v1, backup/atomic write, corruption recovery, injected time service.

**Acceptance:** for same seed + commands, exact same state/events; no negative resources; tests for cost scaling, duplicate buys, caps, integer remainder, clock reversal, 8h offline cap and damaged save.

## M2: first playable conflict
**Tasks**
1. Generate 1 Corp opponent with 2 assets and fictional route.
2. Implement 3 preflight choices, two defense layers, seeded battle tick and loot.
3. Optional reinforcement/withdrawal, timeouts, idempotent resolution.
4. First complete TUI: Overview, Target, Preflight, Battle, Result; 80x24 handling.
5. Add combat matrix test and event-driven battle playback.

**Acceptance:** fresh Runner can buy bots, inspect target, launch an attack, watch counters change, win or lose fairly, recover, save/reopen with same result. Game playable solely on keyboard. Battle results independent of render speed.

## M3: Runner vertical slice
**Tasks**
1. 5 targets, recon/intel 0–100, locked routes and variable ICE.
2. 10 upgrades, research tree and 3 doctrines.
3. First automation: contracts, fabrication, recon; budgets and caps.
4. Progression guide, event archive, offline summary, basic prestige/Identity Burn.
5. Headless scripted progression for at least first 60 minutes.

**Acceptance:** complete coherent 30–60 minute Runner experience; at least two distinct strategies succeed; automation removes manual grind; save and prestige invariants tested; no softlock.

## M4: Corporation campaign
**Tasks**
1. Separate Corp economy: capital, services, compute allocation, research, ICE, reputation.
2. Asset construction, security layers, protected data and upkeep.
3. Budget-constrained NPC Runner attacks in active play only; Corp emergency responses.
4. Acquisitions/market competition, R&D and automated maintenance.
5. Corporation prestige/Restructuring and differentiated theme/UI.

**Acceptance:** Corp is not a reskin of Runner. Playable 30–60 minute loop includes expansion, deliberate security investment, at least one attempted incursion and recovery. Shared combat engine used unchanged.

## M5: polish and cross-platform packaging
**Tasks**
1. Test Linux Bash/Fish, Windows Terminal/PowerShell/CMD and macOS terminals.
2. High contrast, NO_COLOR, ASCII fallback, motion controls, help/tutorial.
3. Stable save migrations and performance regression benchmarks.
4. Build release artifacts and checksums; license and original IP review.
5. Playtest feedback and balance iteration.

**Acceptance:** release binary runs on target OS without Go installed; menus usable at 80x24; recoverable save failures; no terminal state corruption when exiting unexpectedly.

## M6: Steam-ready host (future)
**Tasks**
1. Prototype desktop terminal wrapper with identical TUI binary and no shell subprocess escalation.
2. Test application lifecycle, keyboard handling, resize, Overlay, achievements and Steam Cloud in sandboxed setup.
3. Store assets, trailer/screenshots, platform claims, first-run UX, compliance.
4. Validate Steamworks review, revenue/tax paperwork and fee before launch.

**Acceptance:** clicking Play opens native window and works on the advertised OS. Closing game cleanly saves progress. No reliance on user's default terminal or shell configuration.

## First 10 proposed GitHub issue titles
1. `M0: Initialize Go module and dual CLI/TUI entry point`
2. `M0: Establish content schema and validation`
3. `M1: Implement deterministic tick engine and fixed-point arithmetic`
4. `M1: Add Runner economy, generators and fabrication`
5. `M1: Add atomic versioned saves and offline reconciliation`
6. `M2: Model NPC corporations, routes and ICE layers`
7. `M2: Implement seeded automatic breach resolution`
8. `M2: Build keyboard-only 80x24 TUI battle flow`
9. `M3: Add recon, doctrines and data-driven upgrades`
10. `M3: Add safe automation and Runner prestige`

## Recommended first coding-agent prompt
```text
Read README.md, AGENTS.md, and docs/01-vision.md through docs/11-release.md.
Implement milestone M0 and the smallest useful part of M1 as a working,
testable vertical foundation. Keep a strict boundary between simulation,
persistence and TUI. Follow docs/04-conflict.md for later combat.
Do not implement multiplayer, networking, Steam packaging or real hacking.
Commit focused changes, run go test/go vet, and report deviations from specs.
```

## Definition of done for every milestone
Tests and reproducible manual instructions; update docs when architecture changes; no unverified claims; understandable UX errors; no regression of deterministic behavior, save integrity, original IP boundaries or Corp/Runner design.
