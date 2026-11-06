# NETRISE

**Build your network. Break theirs.**

NETRISE is a planned single-player, offline-first, retrofuturistic cyberpunk incremental strategy game. Players choose one of two genuinely asymmetric campaigns: a clandestine **Runner** who infiltrates corporate systems, or a **Corporation** that grows a profitable information empire while defending its assets.

The game runs as a **native terminal user interface (TUI)** on real terminals (Bash, Fish, PowerShell, CMD/Windows Terminal). The same game core and TUI are intended to run inside a bundled terminal window for a future Steam edition.

> **Status:** active playable prototype. The repository contains the deterministic Runner/Corporation engine, terminal UI, persistence, offline reconciliation and release packaging foundation. Numerical values and schedules remain balance hypotheses requiring playtests.

## Core experience

1. Grow an economy through generators, upgrades and strategic allocation.
2. Prepare a breach (Runner) or a security posture and business expansion (Corp).
3. Make a few meaningful decisions, then watch an **automatically simulated battle** unfold through counters, status reports and terminal animations.
4. React with limited reinforcements, emergency defenses or withdrawal.
5. Earn permanent capabilities, research technologies and gradually automate routine work.
6. Decide when to perform a **prestige reset** for permanent metaprogression.

**Reference inspirations, not assets or implementations to reproduce:** Cookie Clicker (escalating economy), Universal Paperclips (phase transitions), Antimatter Dimensions (layered progression), Melvor Idle (interdependent and offline systems), NGU Idle (resource allocation), Bitburner and Terminal.exe (hacking-flavored progression), OpenFront (commit strength and watch battles unfold), Android: Netrunner (Corp/Runner asymmetry, hidden information, ICE), Blade Runner (analog noir atmosphere). NETRISE must use original names, art, fiction, UI and rules.

## Design principles

- **Corp vs Runner above everything.** Do not turn this into a generic territory-conquest game.
- **Preparation over puzzles.** Players select target, commitment and doctrine; simulation resolves the conflict.
- **Two economies, one world.** Distinct player verbs and resources, shared security / conflict rules.
- **Idle with agency.** Automation is earned; interaction should matter without demanding constant clicking.
- **Offline without punishment.** No unattended catastrophic losses while closed.
- **Native terminal first.** Keyboard accessibility, small-terminal fallbacks and no network dependency.
- **Readable numbers.** Display why attacks succeed or fail, support speed controls and accessible palettes.
- **Data-driven and testable.** Game rules separated from presentation, deterministic simulation with seeded RNG.

## Documentation map

| File | Purpose |
| --- | --- |
| [AGENTS.md](AGENTS.md) | Mandatory coding-agent contract and first tasks |
| [docs/01-vision.md](docs/01-vision.md) | Vision, worldbuilding, audience, pillars, boundaries |
| [docs/02-core-loops.md](docs/02-core-loops.md) | Session rhythm, player journeys, shared world |
| [docs/03-economies.md](docs/03-economies.md) | Resources, generators, upgrade math, factions |
| [docs/04-conflict.md](docs/04-conflict.md) | Fully specified combat model, intel, ICE, outcomes |
| [docs/05-progression.md](docs/05-progression.md) | Unlocks, automation, prestige and offline rewards |
| [docs/06-ui-ux.md](docs/06-ui-ux.md) | TUI layout, interaction, accessibility, aesthetics |
| [docs/07-architecture.md](docs/07-architecture.md) | Go modules, state model, engine, persistence, builds |
| [docs/08-content.md](docs/08-content.md) | Initial content catalog and sample scenarios |
| [docs/09-balancing.md](docs/09-balancing.md) | Balancing targets, formulas, simulations and tests |
| [docs/10-roadmap.md](docs/10-roadmap.md) | Milestones, ticket-ready tasks, acceptance criteria |
| [docs/11-release.md](docs/11-release.md) | Steam and standalone distribution, safety and IP |
| [docs/13-playtest-checklist.md](docs/13-playtest-checklist.md) | Reproducible local Runner/Corporation playtesting |

## Implementation posture

The current implementation provides a narrow but playable **Runner vertical slice** and a first-class asymmetric **Corporation campaign** on one shared deterministic engine. Balance iteration, extended content, human playtesting and the future Steam host remain outside the MVP foundation.

Recommended stack: **Go**, **Bubble Tea v2**, **Lip Gloss**, standard-library JSON for initial saves and content definitions. All packages and APIs must be verified against their installed versions by the coding agent before implementation.

See [AGENTS.md](AGENTS.md) for the sequence and definition of done.

## Licensing

**No license selected yet.** Do not assume the source, documentation, logo or game content may be reused by third parties. The owner should select a license before accepting external contributions or shipping builds.
