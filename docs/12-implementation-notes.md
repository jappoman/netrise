# 12. Implementation notes

The first implementation slice covers M0 and the economy portion of M1: a
Runner state, validated starter content, deterministic 1 Hz resource
production, contracts, generator purchases, bot fabrication, versioned atomic
JSON saves, capped safe offline reconciliation, a headless report, and a
minimal Bubble Tea v2 terminal adapter.

The battle model, Corporation campaign, recon, and research remain M2/M3 work.
The TUI now exposes contract collection, Salvaged Rig purchase, one-bot
fabrication, and persistent save/load. A save records UTC time and reconciles
only safe production on startup, capped at eight hours. `--data-dir` overrides
the default per-user configuration directory.

Runtime amounts use a scale of 1,000 fixed units, with no floating-point
accountancy. Starter balance values remain external JSON hypotheses.
