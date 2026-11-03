# 02. Core loops and session design

## Shared simulation model
The fictional network comprises **assets** owned by organizations, connected by **routes**. An asset has an owner, economic output, security deployment, visibility, access prerequisites, reserve capacity and optional hidden information. Connections represent reachability and intelligence paths, not generic map territory. Runner and Corp see different partial projections of the same underlying state.

## Runner loop
1. Generate passive **compute** and acquire **credits** through low-risk contracts.
2. Spend credits/compute on rigs, bot capacity, software and intel.
3. Inspect available targets; recon improves information quality without requiring a puzzle.
4. Select target, commit a proportion of deployable attack bots and choose one doctrine.
5. Watch seeded automatic combat with clear phases, damage, losses and trace.
6. On success extract data/credits and install limited access; on failure lose resources and gain heat.
7. Research and automate contracts/recon, reduce routine interaction; eventually gain new markets or prestige.

**Runner primary decisions:** which target, how many bots, which doctrine. Optional preflight: buy intel. During combat: once per battle, reinforce or withdraw; skip is allowed.

## Corporation loop
1. Earn recurring **revenue** from legitimate contracts, compute services and data products.
2. Choose between additional capacity, R&D, hiring, security/ICE and reputation.
3. Provision business assets; locate sensitive data in selected protected server tiers.
4. Respond to simulated Runner incursions by security posture, reserve ICE and limited emergency response.
5. Conduct **corporate acquisitions / market pushes** using capital and influence; these are economic operations, not Runner-style hacks.
6. Maintain an efficient portfolio: protect valuable servers without overspending.
7. Unlock security automation, diversification, new industries and prestige.

**Corp primary decisions:** where to invest, what to defend, how aggressively to expand. During an intrusion: deploy reserve ICE or isolate an asset; no forced real-time response.

## Shared world / NPC rules
- World generated from a seed: named organizations, industries, assets, routes, defenders, contracts and flavor.
- AI runs only on fixed simulation ticks, with fixed budgets, deterministic priorities and capped growth. AI shall not have hidden infinite income during active play.
- A Runner player sees incomplete defense composition until purchasing recon; Corp player knows own defenses but has uncertain information about future attack vectors.
- AI must not require a remote service or LLM at runtime.
- Attacks may affect NPC organizations; the player must never be permanently softlocked by losing a single asset.

## Session pacing (initial targets, validate)
| Time since fresh start | Expected meaningful event |
| --- | --- |
| 0–2 min | First manual contract/credit, first generator |
| 3–6 min | First recon and low-tier attack |
| 8–12 min | First recovery from losses, first impactful upgrade |
| 15–25 min | First automated production or recon |
| 30–60 min | First sector unlock and difficult target |
| 90–180 min | First reset opportunity, not forced |

Avoid alert fatigue: only notify on completed research, battles, unlocks and resource caps. The active screen updates at 4–10 FPS; simulation ticks at 1 Hz. Visual FPS is independent from economy time.

## Design of a typical breach
Preflight: 5 seconds for selecting a known target and commitment. Execution: 15–90 seconds at 1x, speed up to 4x or instant resolve once unlocked. Result: loot, trace, bot casualties, access duration, log of decisive events. Repeat only if player chooses or automation has been explicitly unlocked.

## Failure state
There is **no permanent game over**. Failure costs tactical resources and time, creates a recovery decision, and may increase threat. Heat decays; the economy continues to function. No arbitrary permanent confiscation of all accumulated upgrades.

## Offline design
Production and eligible timers progress with time elapsed; player raids, NPC intrusions and other adverse events do NOT resolve while closed. NPC economic growth may progress to a bounded cap without removing owned assets. On returning, show a digest of elapsed time, resources produced, caps reached and queued safe task completions.

## Engagement ethics
No daily login streaks, intentionally punitive notifications, monetized timers or manipulative FOMO. Short sessions and stop-anytime behavior are first-class requirements.
