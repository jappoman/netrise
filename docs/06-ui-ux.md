# 06. Native TUI design and UX

## Product requirement
The exact same TUI must work in a **real user's terminal** and later within a **bundled terminal window** launched by Steam. Do not implement two separate user interfaces or rely on Bash-specific behavior. Use Bubble Tea v2 + Lip Gloss with version-specific verified APIs.

## Core navigation
Top bar: faction emblem, campaign name, uptime, key notifications.
Primary tabs: `Overview`, `Operations`, `Assets`, `Research`, `Archive`, `Settings`.
Runner labels: `Targets`, `Rig`, `Intel`; Corp labels: `Portfolio`, `Security`, `Markets`.
Global: `Tab` switch panel, arrows/jk navigate, `Enter` confirm, `Esc` back, `?` help, `q` quit with safe save. Printable keys must not collide with active text input. Offer remapping later.

## Screen layouts
### Overview
- Runner: credits, compute/s, bots stock/cap, data, heat, active fabrication/recon, next unlock.
- Corp: capital/s, compute allocation, research/s, stationed/reserve ICE, reputation/exposure, contracts.
- Show both stock and rate; short explanations for caps/throttling.

### Target browser / Corp portfolio
- Target name, industry, known security range, loot estimate, route reachability, intelligence quality, cooldown.
- Hide confidential Corp internals from Runner unless legitimately revealed by recon.
- Corp: assets list, current protections, profit margins, sensitive data and possible security alerts.

### Preflight
- Target, bot commitment (slider via +/- or cursor and quick percentages), doctrine selection, intel confidence.
- Confirmation must show exact bots committed and a bounded risk estimate. Deny unaffordable actions clearly.

### Active battle
- Side-by-side combatants where terminal width allows; otherwise vertical.
- Prominent bars: attacking bots, ICE, layer integrity, trace.
- Timeline: limited events with UTC-free fictional timestamps or tick counter.
- Controls: `r` reinforce once, `w` withdraw, `space` pause, `1/2/4` playback speeds, `?` explain values.
- Explicit result screen including loot, survivors, net changes and explanation.
- Visual rendering must **never** determine game ticks.

### Research and automation
- Locked/unlocked/affordable states; show prerequisites and exact effects.
- Automation policies list budgets, thresholds, safe operation types, current state.

## Example screen (visual direction, not pixel-exact contract)
```text
┌ NETRISE // RUNNER TERMINAL ────────────────── 00:17:42 ┐
│ CREDITS 1.24K (+12/s)   COMPUTE 48/s   BOTS 220/400 │
│ HEAT 18%               DATA 320       GHOST KEYS 0 │
├ TARGET: KUROSAWA INDUSTRIES / ARCHIVE 03 ───────────┤
│ Intel: MODERATE  ICE: 400–520  Reward: 220–310 CR   │
│ Doctrine: STEALTH       Commitment: 160 bots (73%) │
│                                                        │
│ PERIMETER    [########.......] 53%                     │
│ RUNNER BOTS  136          DEFENDER ICE  410            │
│ Trace        [###............] 19%                     │
│ [t+18] Layer 01 counterattack: 4 bots lost              │
│ [t+19] Runner pressure: 12 integrity damage            │
│                                                        │
│ [SPACE] Pause  [R] Reinforce  [W] Withdraw  [?] Help   │
└────────────────────────────────────────────────────────┘
```

## Responsive behavior
- Preferred terminal at least 100x30; functional minimum target **80x24**.
- At 80x24 collapse secondary panels and show key values.
- Below 80x24 show a safe compact view with instructions, never panic.
- Handle resize events, SIGINT, terminal focus changes, Unicode width and Windows ANSI support.
- No hard dependency on Nerd Fonts, truecolor, mouse or braille glyphs. ASCII fallback where Unicode unsupported.

## Visual specification
Runner: phosphor green, muted teal, soot black, emergency red. Corp: amber, pale paper, charcoal, red security alerts. Noir retrofuture, deliberately restrained neon.
- Avoid continuously flashing visuals, audio surprises or fake destructive system warnings.
- Respect NO_COLOR and provide 16-color, high-contrast and reduced-motion modes.
- `NO_COLOR=1` disables styling; `NETRISE_HIGH_CONTRAST=1` switches the title
  rail to a bold white-on-black treatment, with `NO_COLOR` taking precedence.
- Do not overwrite real shell history, execute user shell or access clipboard by default.
- Clear distinction between fictional in-game commands and real CLI.
- Text must remain selectable where host permits; avoid high-frequency full-screen disruptive repaints.

## Information design
A player should answer at a glance: What is increasing? Why? What can I afford? What can I attack? What do I risk? What just happened?
- Numbers use compact suffixes in panels and full values on inspect (e.g. 12.4K / 12,400).
- Display uncertainty honestly (`estimated ICE 400–520`).
- Include event explanation: `Countermeasure: +20% ICE attrition`.
- Support pause and instant skip of lore typing. All important messages persist in Archive.

## Self-explanatory screens

Every player-facing screen must explain itself without requiring a manual or
prior memory of the controls. Keep a compact, persistent action rail visible
at the bottom or beside the primary content. It should show:

- the current objective or state (`NO ACTIVE BATTLE`, `FABRICATION RUNNING`);
- the actions currently available and their keys;
- the next useful action for a new or returning player;
- concise reasons for locked or unaffordable actions.

The rail is contextual, not a modal tutorial: it must remain visible while
the player can still inspect the main screen. Detailed explanations belong to
the `?` help view and must not obscure the simulation by default.

## Steam host decision record
Later evaluate embedded terminal host (e.g. a lightweight native window with VT emulator) versus OS terminal launcher. The terminal host must launch local game executable, propagate resize/keyboard input correctly, and not run arbitrary shell instructions. Investigate Steam Overlay, achievements and close behavior on each platform separately. **Do not build host during MVP.**
