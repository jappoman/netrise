# 08. First content catalog and world design

This is an **illustrative content seed**. All numeric values are balancing hypotheses and should live in schema-validated data, never embedded in UI handlers.

## Fiction style and original naming
Rainfall megacity **Vesper City**, neon logistics, night-shift analysts, pneumatic archives, cold data farms, analog surveillance and multinational corporate power. Corporate NPCs are original. Never use direct Android: Netrunner faction names, named programs, real-world companies as fictional adversaries, or Blade Runner characters/brands.

## NPC corporations
| ID | Name | Industry | Identity |
| --- | --- | --- | --- |
| c_01 | Holloway Transit | Logistics | Cheap routing infrastructure, thin perimeter |
| c_02 | Kuroda Dataworks | Archiving | Valuable caches, layered archive ICE |
| c_03 | Meridian Civic Systems | Identity | Public contracts, trace-heavy security |
| c_04 | Asterion Biologics | Biotech | High-value unique R&D records |
| c_05 | Obsidian Securities | Security | Hardened network, counterattack specialty |

## Five initial Runner targets
| ID | Corp | Reach | Security theme | Expected reward tier | Novelty |
| --- | --- | --- | --- | --- | --- |
| t_01 | Holloway | start | single layer, low ICE | 1 | tutorial raid |
| t_02 | Holloway | from t_01 | more attrition | 2 | demonstrates bot loss |
| t_03 | Kuroda | after basic recon | hidden archive/2 layers | 3 | intel value |
| t_04 | Meridian | after access | high trace pressure | 4 | stealth tradeoff |
| t_05 | Obsidian | research gate | thick ICE/3 layers | 5 | commitment planning |

Each target needs: layer integrity, defense resist, starting/max ICE, trace modifier, loot pool, regeneration, route constraints, description, reveal thresholds and a validated `cooldown_ticks` refill window. Tutorial target should be forgiving but not free of losses.

## Starting upgrades (10 Runner examples)
| ID | Name | Function |
| --- | --- | --- |
| u_01 | Salvaged Clock | +compute rate |
| u_02 | Secondhand Compiler | unlock bot fabrication |
| u_03 | Heat Sink Array | lower heat generation |
| u_04 | Proxy Ledger | improve stealth doctrine |
| u_05 | Stolen Memory Bank | +bot capacity |
| u_06 | Street Contacts | unlock recon |
| u_07 | Packet Scheduler | faster fabrication |
| u_08 | Autodialer | automate low-risk contracts |
| u_09 | Signal Profiler | narrower ICE estimate |
| u_10 | Black Archive | unlock data research |

Targets also define an `access_ticks` duration for temporary route permissions granted
after a successful breach. Use explicit costs, unlock predicates and multiplier definitions in content JSON. The
initial catalog uses a validated upgrade effect block for compute, heat, stealth,
bot capacity, safe-job duration and research discounts. Upgrade UX should display
exact before/after effect.

## Initial research branches
Runner: **Hardware**, **Intrusion**, **Anonymity**, **Automation**.
Corp: **Infrastructure**, **Monetization**, **Security**, **Intelligence**.
Runner research nodes have stable IDs, costs, prerequisite lists, a validated effect DSL
(`compute_per_second_bonus`, `attack_bonus_percent`,
`stealth_heat_reduction_percent`, and `automation_interval_reduction`) and
descriptive text. Effects are aggregated by the simulation only after a node is
completed.

Corporation R&D uses a separate threshold catalog (`corp_research`) rather than
sharing Runner prerequisites. Completed technology IDs are persisted on the
Corporation state; the current seed includes Network Efficiency (capital per
tick) and Lean Operations (asset upkeep reduction). Effects are aggregated by
the engine, so the report and active economy use the same source of truth.

## Three doctrines
Balanced (safe default), Stealth (lower heat/attrition with weaker attack), Overload (faster power with extra trace and losses). Each receives a distinct short textual description and color-neutral icon/text cue.

## Corp business sectors (later milestone)
- Transit hosting: high volume, modest profit, weakly sensitive data.
- Data archiving: recurring profits with increasingly valuable secrets.
- Identity services: lucrative contracts, high exposure costs, specialized trace defense.
- Biotech R&D: slow but valuable projects, differentiated research rewards.
- Private security: profitable ICE products and stronger countermeasures.

Corporate progression includes client demand, operating margin, reputation, information value and defensive exposure. The Security Desk carries a data-driven ICE capacity bonus, so defensive investment changes the reserve ceiling rather than only adding income. An acquisition is an economic investment, not a magic hack performed by Corp.

## Event writing style
Concise industrial noir with minimal forced jokes:
- `ARCHIVE CONNECTED // Security index incomplete.`
- `Countermeasure activated: Warden-class ICE.`
- `Transfer complete. You were never here.`
- `BOARD REPORT // Quarterly uptime exceeds forecast.`
- `Emergency audit: suspicious ingress detected.`

## Content validation
Unique IDs, no cyclical research prerequisites, reachable starter target, positive rates, costs representable, resolvable effects, valid resource names, reachable progression milestones, no broken links, meaningful tooltips and English source strings.
