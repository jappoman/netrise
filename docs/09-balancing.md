# 09. Balancing, simulations and playtest methodology

**All target numbers are hypotheses** pending playtest. Build the headless simulator before tuning visual polish.

## Desired pacing
| Checkpoint | First-run target |
| --- | --- |
| First generator | 30–90 seconds |
| First target intel | 2–4 minutes |
| First successful breach | 3–7 minutes |
| First meaningful bot losses | 5–10 minutes |
| First automation unlock | 15–25 minutes |
| First second-tier sector | 30–60 minutes |
| First optional prestige | 90–180 minutes |

Do not add artificial waits merely to hit targets. Player should always have an interesting choice even when waiting.

## Metrics to collect locally in balance runs
Time to afford next critical purchase, target win rate by relative strength, expected bot attrition, target ROI including replacement cost, resource bottleneck duration, heat burden, time between active choices, upgrades purchased and reliance on manual contract. Use headless simulations, not real user tracking in MVP.

## Battle test matrix
Sweep seed [1..100], attacker commitment levels 25/50/75/100%, doctrines [balanced, stealth, overload], target tier 1..5, intel 0/50/100 and relevant tech levels. Assert:
- Power advantage generally increases win rate, controlling other variables.
- Greater commitment does not reduce win chance except where a clearly documented rule justifies it.
- Stealth usually lowers heat and bot attrition but often takes longer or fails against strong layers.
- Overload usually wins faster under equal investment but increases heat and/or attrition.
- Purely unlucky outcomes should remain rare at substantial advantage.
- Losses and loot are never negative or duplicated.
- Time limit and reserve capacity never create indefinite battles.
- Any target can be overcome through feasible progression without perfect luck.

**Warning:** a simple attrition formula can create runaway collapse (bots die so damage falls, extending fight). Use automated tests to detect impossible or oppressive low-level battles and adjust layer health, damage, ICE attack or early bot income.

The integration suite includes a deterministic smoke sweep over all five catalog
targets, all three doctrines, four commitment levels and seeds 1..100. It makes
targets reachable only inside the test, then asserts that every battle resolves
within its time limit, records a valid Runner outcome, preserves bot/heat
accounting and never produces negative resources. The broader intel/tech matrix
remains a playtest expansion rather than a release gate.

The Corporation integration suite also runs a scripted 3,600-tick campaign
with all three starter assets, security capacity, maintenance and automated
R&D enabled. It asserts bounded capital/compute/research/ICE/meters, event-log
limits and actual strategic progress rather than only survival.

Both factions also have a 21,600-tick (six-hour equivalent) headless
regression with safe automation enabled. It asserts bounded resources, capped
event history and continued strategic progress.

Runner integration also walks the real catalog from `t_01` through `t_05`,
asserting that each victory grants the next temporary access window and that
the final route is reachable without bypassing route validation.

## Economy experiments
- Compare optimal and plausible novice strategies; novice should make progress without reading a wiki.
- Sample buying one generator repeatedly, mixed purchases, saving for milestones and research allocation.
- Every generator tier must provide at least one relevant period where purchasing it beats waiting for the next tier.
- Detect softlocks from spending all compute on upkeep, having no bot fabricator or a depleted loot market.
- Evaluate long offline sessions, caps and fractional production accumulation.
- Ensure Corp can recover from an incursion and Runner can recover after losing 100% of attack bots.

## Fixed arithmetic properties
`cost(i,n+1) >= cost(i,n)`; cumulative purchases equal repeated purchases; `stock >= 0`; `remaining <= cap`; conversion/upgrade never creates value via round-trip loops unless explicitly designed. No overflow after 1e6 headless ticks or at extreme but valid ownership counts. Add property tests and fuzzing.

## Prestige experiments
For each simulated reset strategy, record time back to first breach and new content reached. Avoid trivial rapid-reset farming, but don't require a long dead period. Compare first-run and second-run timelines; prestige should increase strategic options in addition to speed.

## Human testing questionnaire
After 15 minutes:
- Did you understand the difference between bots, compute and credits?
- Could you explain why a raid succeeded or failed?
- Did you know what to do next?
- Did you feel forced to click repeatedly?
After 60 minutes:
- Has gameplay introduced anything meaningfully new?
- Do stealth and overload feel distinct?
- Is it appealing to watch a battle unfold?
- Is the interface legible in an 80x24 terminal?
For Corp tests:
- Are business expansion and security investment genuinely different from Runner raids?
- Can you understand the tradeoff of leaving value undefended?

## Acceptance thresholds for release candidates (provisional)
- No deterministic crash, save corruption, progression dead end or unbounded event log in a 6-hour simulated playthrough.
- 90% of fresh blind testers complete initial breach without external help.
- No mandatory repeated manual clicking after first automation unlock.
- 100% of core actions accessible by keyboard and in 80x24 mode.
- Tests demonstrate 100% replay determinism for fixed seeds/commands within same game version.
