# 04. Intrusion and defense: automatic combat specification

This is the signature system. It is **not** a command-line hacking puzzle, card combat or a physically accurate cyberattack. Everything is fictional and mechanically abstract. The player prepares, chooses target/commitment/doctrine, and watches numbers resolve.

## Ontology
A **route** connects two fictional network assets. A **target** has a public security tier, protected value and 1–3 defense layers. A **layer** is a security stage with integrity and optional ICE unit pool. A **Runner strike force** commits bot units and software-derived damage modifiers. A **Corp defensive posture** commits stationary ICE and policy bonuses. Units and damage are game abstractions.

### Runner preflight decisions (exactly three mandatory)
1. **Target:** which reachable corp asset to breach.
2. **Commitment:** a whole number of bots, shown also as 10/25/50/75/100% quick presets.
3. **Doctrine:** balanced, stealth or overload. All three available once combat unlocks.

Optional intel purchase/job before preflight, not another mandatory decision. Target info defaults to intervals (e.g. ICE 300–450), improving toward exact values as intel rises.

### Doctrine values, starting hypothesis
| Doctrine | Damage dealt | Bot losses | Heat | Speed |
| --- | --- | --- | --- | --- |
| Balanced | 1.00 | 1.00 | 1.00 | 1.00 |
| Stealth | 0.82 | 0.75 | 0.55 | 0.90 |
| Overload | 1.25 | 1.30 | 1.70 | 1.15 |

A doctrine may never dominate every other one. Review strategies using deterministic sweeps over target tiers.

## Combat timeline
Combat uses **1-second authoritative simulation ticks**, displayed at selectable 1x, 2x or 4x (speed alters real-time playback, not results). Keep combat and economic simulation event order explicit:
1. Validate eligibility/affordability/access; reserve committed bots atomically.
2. Snapshot defender layers and modifiers, derive battle-local seeded RNG.
3. Each tick: process pending player intervention, attacker damage, defender damage, status/trace and event emission.
4. Finish upon all layers breached, attacking force depleted, explicit withdrawal, or battle time limit.
5. Reconcile surviving bots, deterministic rewards and heat; emit immutable result event; persist atomically.

### Initial discrete deterministic formula
All values below are **integers** and computed with a scale constant `S=10000` (basis points). Avoid fractional state corruption.

For tick t against current layer L:
- `effectiveAttack = max(1, floor(liveBots * botAttack * doctrineAttackBP * attackUpgradeBP / (S*S)))`
- `effectiveDefense = max(1, floor(liveICE * iceDefense + layerBaseResist))`
- `layerDamage = max(1, floor(effectiveAttack * S / (S + effectiveDefense)))`
- `botLosses = min(liveBots, max(1, floor((liveICE * iceAttack + layerCounter) * doctrineLossBP / (S * max(1, botArmor)))))` **only when defender has counterattack capability**
- `layerIntegrity = max(0, layerIntegrity - layerDamage)`
- `liveBots = max(0, liveBots - botLosses)`
- `trace = clamp(trace + baseTraceBP * doctrineHeatBP / S + triggeredICETrace, 0, traceCap)`

Starter combat constants: `botAttack=1`, `botArmor=1`, `iceAttack=1`, `iceDefense=1`, `baseTraceBP=100`; layer integrity / resistance / ICE quantities are tuned in data. Compute with checked wide intermediates to avoid overflow. Distinguish basis-point trace from percentage display.

**Important balancing note:** these equations are a baseline, not a proof of good gameplay. Before implementation of more content, run outcome matrix tests (commitment vs defense) and ensure low-tier battles finish in 15–90 ticks. If rates are too fast or too slow, adjust layer parameters in content data first; change formulas only with tests and documented rationale.

### Layer transitions and outcomes
On layer integrity reaching zero, emit `LayerBreached`, optionally release trapped intel/access and move to next layer in next tick. On last layer: victory and a single loot transaction. On depletion: defeat with no loot. On withdrawal: survivors return minus explicit retreat penalty (initially 0), no loot, less heat than fighting to destruction. Timeout: defender victory, residual bots return (not wiped).

## Controlled uncertainty
- Seeded PRNG used only for bounded variation in damage (initially ±5%) or named events, **not** a secret coin-flip deciding overall victory.
- Same state+seed+commands+ticks must produce byte-for-byte equivalent battle results.
- Intel affects visibility and preparation opportunities, **not** direct RNG manipulation.
- Show a preflight **estimate range**, never a falsely precise win percentage unless measured by the actual model over a seed distribution.

## Intervention budget
At most **one optional intervention** per fight in MVP:
- `Reinforce(n)`: once, append up to available uncommitted bots, atomic reserve. Extra heat +5 percentage points.
- `Withdraw`: always available; ends engagement immediately.
A Corp player can later `DeployReserveICE(n)` or `IsolateAsset`; these have shared costed action rules. No repeated rapid-click buffs.

## Reward logic
A victory earns credits and data bounded by the target's actual **loot pool**. Repeated raids cannot conjure unlimited cash: each target has a data-driven cooldown/refill window and temporarily locks down after a resolution. Win unlocks time-limited access to adjacent assets. High-tier corp archives may give unique data once per run. Failure loses destroyed bots, increases heat, advances the same bounded cooldown. Recovery always remains possible.

## Corp defense mechanics
Defenders choose protection budgets and deploy ICE on 1–3 layers, hiding actual placement from poorly informed attackers. Layers may have different countermeasures (high integrity, bot attrition, trace gain) and an upkeep cost. Stationary ICE consumed in battle must be restocked from real corp resources; passive upgrades do not produce infinite armor.

## Telemetry presented to player
- Before: commitment, doctrine, known uncertainty in target security, estimated duration, projected max heat, spend/reserve.
- During: current layer, bots remaining, ICE remaining, integrity bar, trace bar, numerical deltas per tick, milestone events. Avoid dumping every tick into an unbounded log.
- After: victory/defeat/withdrawal, costs, net profit, bot casualties, actual defense encountered, new access, heat changes and why.

## Invariants and tests
- Each unit is in precisely one place: stock, reserved in battle, surviving/returned, or destroyed.
- A battle resolves **once** and rewards only **once**; replaying finalization is idempotent.
- Defender integrity and ICE never become negative.
- No attack on unreachable targets, no over-commitment, no attack on oneself.
- Same seed produces identical events and results, even at different render FPS.
- Pausing saves safely; restart resumes deterministically.
- Connection to real hosts or execution of exploit code is impossible by construction.
