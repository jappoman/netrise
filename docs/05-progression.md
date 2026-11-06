# 05. Progression, automation, prestige and offline play

## Four progression horizons
1. **Local terminal:** manual credits, rented compute, first low-value corporate server.
2. **Underground network / local enterprise:** basic automation, layered security, specialized technology and meaningful intel.
3. **Citywide operations / regional corporation:** parallel jobs, complex targets, defensive planning, mid-scale economic synergies.
4. **Global information power:** systemic pressure and rare assets, advanced AI agents, long-term meta unlocks.

These are *mechanical changes*, not merely larger integer displays. Every stage introduces at most 1–2 new resources or decisions and explains them at first use.

## Runner unlock ordering
- Starter contract + salvaged rig → bot compiler.
- First target discovered → preflight and battle replay.
- First success → access token and adjacent targets.
- First data extraction → R&D screen.
- Three victories or early research → automatic contracts.
- Mid-tier intel tech → automatic recon jobs with player-defined budget.
- Sector clearance → parallel background fabrication and queued safe raids (raids while actively running only).
- First notable archive → prestige option.

## Corp unlock ordering
- First contract → new revenue capacity.
- Security desk → layered ICE assignment.
- First defended attack → intrusion analytics, security report.
- First acquisition → ownership/economic portfolio view.
- R&D milestone → automated ICE maintenance, configurable reserve.
- Regional scale → sector-specific projects and new market policies.
- First high-value secret secured → prestige option.

## Automation contract
Automation must be explicitly purchased and configured, with resource budgets and a stop condition. Initially:
- Runner: automatic low-risk contract collection, bot manufacturing, target intel renewal.
- Corp: contract renewal, ICE restocking and resource allocation thresholds.
- Runner automation uses explicit cumulative credit budgets and stops when a cap is reached; the TUI exposes policy, budget and spent totals.
- Corporation automated ICE maintenance uses the same explicit capital budget, restocks only during active play and stops at its cap.
- **Never auto-spend scarce resources without a configurable budget.**
- **No auto-raids during offline time** and no offline enemy attacks in any campaign.
- Auto-combat on active screens can be unlocked later; avoid autoplay that removes all meaningful decisions.

## Prestige: Runner "Identity Burn"
Narrative: abandon compromised identity, retain encrypted know-how.
- Resets: credits, bots, local rigs, intel snapshots, access tokens, heat and most sector control.
- Keeps: permanently earned **Ghost Keys**, lifetime achievements, cosmetic unlocks and account settings.
- Unlocks: starter credit boosts, new intel tactics, more efficient bot economy and alternate approaches.
- Proposed currency: `ghostKeysEarned = floor(sqrt(runExtractedData / 10000))`; award once at reset, then set the next run's extracted-data counter to zero. This is provisional and must be tested against repeated tiny resets.
- Only cumulative **this-run** data counts for an upcoming reset; keep lifetime and current-run counters separately to avoid inflation and exploits.

## Prestige: Corp "Restructuring"
Narrative: spin off a new enterprise with retained patents.
- Resets: capital, regular contracts, current assets, stationed ICE, exposure and sector acquisitions.
- Keeps: **Patent Shares**, unlocked corporate profiles/cosmetics and account settings.
- Unlocks: starting capital, better early racks, long-term market opportunities and R&D efficiencies.
- Use an analogous convex/concave reward curve tuned to *this-run corporate valuation*, not extracted Runner data.
- Neither faction should be obliged to prestige merely to resolve the first hour of progression.

## Prestige invariants
- Confirmation shows exactly what is lost, retained and gained.
- Persist transaction atomically; cannot grant rewards twice.
- Free or low-cost respec of permanent perks is preferred to irreversible traps.
- Prestige must add a new strategy or capability, rather than only a 2x all-stats multiplier.
- Changing faction creates a separate campaign save; prestige currencies never cross factions in MVP.

## Offline economy
- Calculate from wall-clock elapsed time but validate monotonic bounds against stored save timestamp.
- Initial cap: 8 hours effective offline production; upgradeable later up to 24 hours.
- Cap credits/data/stocks at their storage limits, carry bounded fractional remainder, and apply explicitly allowed safe queued work.
- Zero offline adverse outcomes: no raids, no bot loss, no disappearing assets, no heat growth.
- Safe decay of heat and safe Corp exposure recovery may occur offline.
- Clamp suspicious clock differences and report reconciliation rather than granting absurd amounts.
- On resume display production totals and cap warnings with an option to inspect.
- Long offline calculations must be **O(number of generators + queued safe jobs)**, not iterate every historical second.

## Long-term goals
Optional achievements and collection: rare intelligence, recovered analog archives, corporate dossiers, monochrome terminal themes, short lore fragments. Avoid daily quests and streak coercion.

## Balance gates
Playtest first automation before ~25 min, first optional prestige around ~2–3h (hypotheses). A reset should noticeably accelerate the next run and reveal at least one fresh decision within 15–30 minutes.
