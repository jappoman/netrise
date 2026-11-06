# 03. Economies, generators and upgrades

Numbers here are **starting balance hypotheses**, not promised final values. Use content definitions so playtests can rebalance without changing code. Accounts use integer micro-units where needed; avoid float precision for stored resources.

## Design model
Separate **stocks** (credits, data, bot units, influence), **capacities** (compute throughput, bot slots, storage, defense slots), **flows** (per-second generation, upkeep) and **meters** (heat, reputation). Every item has an in-world source and at least one meaningful sink.

### Runner accounts
| Resource | Source | Sink | Property |
| --- | --- | --- | --- |
| Credits | street contracts, extraction, data sales | rigs, bot manufacturing, tools, recon | persistent stock |
| Compute | rigs and rented nodes | bot fabrication, research, background jobs | capacity per second |
| Bot units | compute fabrication with credit cost | raids; units survive or are lost | stock, capped by bot storage |
| Intel | recon jobs, contacts, recovered logs | reduces uncertainty; intel expires per target | target-scoped quality 0–100 |
| Data | extraction from secured assets | permanent research, black-market exchanges | persistent stock |
| Heat | attacks, noisy recon | decay, laundering, stealth | 0–100 pressure meter |
| Access | successful breaches | temporary contracts, extraction bonus, new reachable targets | expiring permission token |

### Corp accounts
| Resource | Source | Sink | Property |
| --- | --- | --- | --- |
| Capital | service contracts and product revenue | servers, ICE, acquisitions, R&D | persistent stock |
| Compute | owned racks | services, analytics, security systems | capacity per second |
| Research | R&D labs and proprietary archives | software and infrastructure technologies | persistent stock |
| Influence | PR, partnerships, acquisitions | market expansion and negotiation | persistent stock |
| ICE capacity | security appliances | stationed defenses and emergency reserve | integer units, capped |
| Reputation | uptime and contract fulfillment | impacts service income and influence | 0–100 meter |
| Exposure | discovered intrusions and leaks | increases ongoing cost and contract risk | meter with recovery path |

Credits and capital can share an internal numeric primitive without being interchangeable currencies between campaigns. Intel and access are informational advantages; not just another generator currency.

## Baseline formulas
For any generator tier i, `cost(i,n) = ceil(baseCost_i * growth_i^n)`, with a bulk-purchase algorithm that exactly matches buying sequentially. For early tiers, growth_i between 1.12 and 1.20; each tier has a meaningfully higher starting cost and yield. Avoid assuming floating-point pow remains exact for large n; use deterministic integer or decimal algorithms.

`grossRate = sum(count_i * baseRate_i * purchasedModifiers_i) * globalMultiplier`

`netRate = max(0, grossRate - mandatoryUpkeep)`

`storageLimitedGain = min(netRate * elapsedTicks, storageCap - existingStock)`

Use a fractional remainder ledger for rates expressed in fixed-point units so small positive production is never lost due to repeated rounding.

### Scaling policy
- Upgrade milestones at generator counts 10, 25, 50 and 100 may unlock multipliers, not just passive flat additions.
- Cost growth should periodically outrun single-tier output, encouraging diversification.
- Soft caps should be visible and explained, never hidden in code.
- If a resource has no useful spend in a phase, introduce a conversion, optional upgrade or research sink.
- Defense upkeep should not push Corp into a permanent debt death spiral; nonessential defenses throttle or deactivate gracefully.

## Starter Runner content (tunable)
| Generator / action | Initial cost | Output / effect |
| --- | --- | --- |
| Manual contract | free, 4s cooldown | +10 credits |
| Salvaged rig | 50 credits | +1 compute/s |
| Rental compute | 250 credits | +6 compute/s |
| Bot compiler | 75 credits | converts 10 compute + 2 credits into 1 bot, job 3s |
| Contact network | 150 credits | starts 15s intel job |
| Local data cache | 100 credits | +100 data capacity |

Initial wallet 25 credits, initial compute capacity 2/s from starter machine, 20 bot slots; starting low-tier target within 2–6 minutes. No material should require constant manual spam after first automation unlock.

### Runner tradeoffs
- Spending compute on bot fabrication reduces compute available for research.
- Saving bots enables larger raids but delays income and progression.
- Recon uses time and credits but improves estimate of enemy defense and rewards.
- Heat may temporarily raise contract friction or recon cost; never terminate the run.

## Starter Corp content
| Generator / action | Initial cost | Output / effect |
| --- | --- | --- |
| Starter client contract | free | +2 capital/s while serviced |
| Rack expansion | 75 capital | +5 compute capacity/s |
| Hosting subscription | 150 capital | +4 capital/s, consumes 2 compute/s |
| Security desk | 90 capital | +10 ICE slot capacity |
| ICE assembler | 125 capital | builds 1 ICE/5s for capital + compute |
| Research department | 200 capital | +1 research/s, consumes 1 compute/s |

Corp economics must include **market demand** and marginal operating costs in later tiers; not only bigger income multipliers. Every successful expansion exposes additional value, incentivizing better security.

Corporation R&D converts capital into threshold-based technologies. Network
Efficiency improves capital flow, while Lean Operations reduces upkeep on
operational assets. These technologies are data-driven, persist during the
campaign and reset on Restructuring with the rest of the run state.
The Security Desk also adds three ICE capacity slots; when maintenance was
already reserving the previous full rack, the reserve expands with the asset.
If upkeep throttles an asset, the engine reactivates it automatically once
available capital can cover the operational costs again; the recovery is
recorded as an event and surfaced in the Corporation overview.
Low reputation and high exposure reduce effective service revenue with bounded
integer multipliers; healthy operations recover the full demand-adjusted rate.

## Anti-exploit rules
- No negative balances; spending is validated and atomic.
- Production is bounded by capacities, upkeep and storage.
- Never let clock rollback create negative or infinite earnings.
- Bulk purchases reproduce single-purchase totals exactly.
- Named upgrades may be unique; duplicate purchase is rejected.
- On prestige, apply reset tables explicitly (see docs/05-progression.md).
- Every resource conversion rounds consistently and documents fractional carry.

## UI requirements
Display current stock, effective rate, capacity use, next affordable upgrade, estimated time to affordability, and why a generator is throttled. Explain all multipliers in inspect/details views.
