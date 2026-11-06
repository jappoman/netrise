# NETRISE playtest checklist

This checklist is intentionally local and manual. It collects no telemetry and
does not require a network account. Record observations in a scratch note, not
inside save files.

## Build and launch

From the repository root:

```text
go run ./cmd/netrise validate-content
go run ./cmd/netrise playtest --faction runner --seed 17
go run ./cmd/netrise playtest --faction corp --seed 17
make smoke SEED=17 TICKS=600
go run ./cmd/netrise --data-dir ./tmp/playtest-runner
go run ./cmd/netrise corp --data-dir ./tmp/playtest-corp
```

The `playtest` reports are deterministic JSON baselines for the scripted first
loop, including recon and battle checkpoint ticks. Use them to compare catalog
changes before and after a manual session; they are not a replacement for
observing the player-facing TUI.

Use separate data directories for the two factions. Delete those temporary
directories between fresh-run sessions; keep one copy when testing reopen and
offline reconciliation.

## Runner first loop

Start a fresh Runner campaign and note the time for each checkpoint:

1. Collect contracts until the overview explains the compiler purchase.
2. Open `[g]` Upgrades, buy `Secondhand Compiler`, and confirm the installed
   state remains after quitting and reopening.
3. Fabricate bots, open `[t]` Targets, inspect the ICE/reward uncertainty, and
   run recon before launching the first breach.
4. Use preflight to compare at least two doctrines or commitment levels; record
   why the chosen option is affordable and reachable.
5. Watch the battle playback, then inspect Result and Archive. Confirm the next
   action is explicit after both victory and withdrawal.
6. Enable one safe automation policy with a finite budget and verify it stops
   at the configured cap.

## Corporation first loop

Start a fresh Corporation campaign and note the time for each checkpoint:

1. Build `Transit Hosting`, then install `Security Desk` and inspect the ICE
   capacity increase.
2. Open the report and compare gross revenue, upkeep and operating margin.
3. Allow an active-play incursion to start. Deploy reserve ICE or isolate an
   asset, then inspect the Result and Archive screens.
4. Repair or recover a damaged asset and confirm the overview explains the
   recovery action.
5. Configure maintenance or R&D with a finite budget and verify it does not
   spend during offline time.

## Terminal and accessibility matrix

Run the short launch/quit flow on each available host:

- Linux Bash and Fish
- Windows Terminal with PowerShell and CMD
- macOS Terminal

Also repeat the overview and battle flow with:

```text
NO_COLOR=1
NETRISE_HIGH_CONTRAST=1
NETRISE_REDUCED_MOTION=1
NETRISE_ASCII=1
```

At 80×24, confirm the action rail, current objective, unavailable-action
reason and `NEXT:` hint remain visible. Open Upgrades and move through the
full list with `[up/down]`; the visible window should scroll without hiding
the action rail. At a smaller size, confirm the compact view remains usable
and quitting still writes a valid save.

## Feedback to record

For each session record only:

- time to first meaningful purchase, recon and breach/incursion;
- whether the next action was obvious without opening help;
- whether resource caps and risk estimates were understandable;
- whether automation removed repetitive work without removing decisions;
- whether any command felt unexpectedly unavailable or destructive;
- terminal size, shell, accessibility environment and faction.

The first balance pass should use these observations to adjust catalog data,
not simulation invariants or the Runner/Corporation asymmetry.
