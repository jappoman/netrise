# 11. Distribution, safety, legal and commercial checklist

## Dual-mode distribution
**Terminal-native edition:** cross-compile Go standalone binaries. Run from Bash, Fish, PowerShell, CMD/Windows Terminal and macOS Terminal where ANSI/VT support is appropriate. The user shell is a host, **not the game scripting engine**.

**Steam edition (future):** wrap the same TUI/game binary in a dedicated terminal host, launched from Steam as an application window. Don't assume Steam itself opens a usable console or provides a PTY. Evaluate OS-specific PTY support, suspend/resume lifecycle, input, close/exit, accessibility, Steam Overlay and controller expectations. Steam integration and Steamworks SDK are not needed for initial prototype.

## Publishing
Steamworks onboarding, store page approval, build review, legal entity/tax and banking details, required disclosure, applicable platform rules and Steam Direct fee must be checked against current official documentation near launch. A $100 per-app Steam Direct fee and recoupment criteria are commonly listed; policies may change and are not legal advice. Do not promise platform approval before testing.

## IP clearance
- NETRISE name is provisional. Search trademarks and game/store naming conflicts before public branding.
- All corporations, lore, UI copy, diagrams, audio and art must be original or properly licensed.
- **Do not** use Blade Runner characters, locations, logos or named entities; do not use Android: Netrunner cards, copyrighted factions, ICE names, flavor text, logos or art.
- Treat reference games as genre inspirations only.
- Decide source license and asset license separately; current repository deliberately has no LICENSE.

## Product model
Tentative one-time purchase on Steam; local standalone build availability and pricing remain undecided. No ads, no microtransactions, no time-limited spending. Community mod support may arrive after content schema hardens.

## Security
The hacking vocabulary is fictional. Executing game actions must never connect to real third-party devices, run a system shell, provide exploit tooling or require administrative permissions. Only locally simulated IP-free fictional networks. Process spawning for future terminal host must launch vetted local executable only; sanitize paths/arguments and protect save data from arbitrary injection.

## Privacy
No account, telemetry or server needed in the MVP. Local save data only. Steam Cloud opt-in depends on actual integration; no invented cloud feature claims. No personal data required.

## Release gates
- Final names and legal/IP review complete.
- Core balance and 60-min blind playtest.
- Cross-platform input and window compatibility.
- Successful 80x24 and NO_COLOR tests.
- Robust saves, corruption recovery, migration.
- License and third-party notices complete.
- Clear accessibility settings and minimum requirements.
- Correct screenshots showing actual gameplay.
- Automated builds, tests, checksums and release notes.

## Future possibilities, not MVP
Multiplayer asymmetry, asynchronous competitions, user-authored networks, modding, achievements, Steam Cloud, soundtrack, full localization, controller mode, richer city simulation and advanced Corp espionage. Each must justify additional complexity and preserve offline-first gameplay.
