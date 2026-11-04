# 12. Implementation notes

The first implementation slice covers M0 and the economy portion of M1: a
Runner state, validated starter content, deterministic 1 Hz resource
production, contracts, generator purchases, bot fabrication, versioned atomic
JSON saves, capped safe offline reconciliation, a headless report, and a
minimal Bubble Tea v2 terminal adapter.

The battle model, Corporation campaign, recon, upgrades, and save timestamp
clock adapter remain M2/M3 work. The TUI currently uses an in-memory campaign;
`--data-dir` is accepted for portable-build compatibility, but interactive
load/save wiring is intentionally deferred until the complete save-resume flow
is added. This avoids presenting an apparently persistent game that does not
yet reconcile timestamps correctly.

Runtime amounts use a scale of 1,000 fixed units, with no floating-point
accountancy. Starter balance values remain external JSON hypotheses.
