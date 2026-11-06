GO ?= go
SEED ?= 17
TICKS ?= 600
SEED_FROM ?= 1
SEED_TO ?= 100
FACTION ?= runner
DATA_DIR ?=

.PHONY: help fmt vet test race verify build bench validate-content play corp simulate playtest balance smoke

help:
	@printf '%s\n' 'NETRISE commands:' '  make play [DATA_DIR=/path]                         start the Runner TUI' '  make corp [DATA_DIR=/path]                         start the Corporation TUI' '  make simulate [FACTION=runner SEED=17 TICKS=600]     run a deterministic report' '  make playtest                                       run deterministic first-loop reports' '  make balance [SEED_FROM=1 SEED_TO=100]             run the combat balance matrix' '  make smoke                                          run headless release smoke flow' '  make test                                            run unit tests' '  make verify                                          format, vet, test and race-test'

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

verify: fmt vet test race

build:
	$(GO) build ./cmd/netrise

bench:
	$(GO) test -bench . -benchmem ./internal/sim

validate-content:
	$(GO) run ./cmd/netrise validate-content

play:
	@if [ -n "$(DATA_DIR)" ]; then $(GO) run ./cmd/netrise --data-dir "$(DATA_DIR)"; else $(GO) run ./cmd/netrise; fi

corp:
	@if [ -n "$(DATA_DIR)" ]; then $(GO) run ./cmd/netrise corp --data-dir "$(DATA_DIR)"; else $(GO) run ./cmd/netrise corp; fi

simulate:
	$(GO) run ./cmd/netrise simulate --faction "$(FACTION)" --seed "$(SEED)" --ticks "$(TICKS)"

playtest:
	$(GO) run ./cmd/netrise playtest --faction runner --seed "$(SEED)"
	$(GO) run ./cmd/netrise playtest --faction corp --seed "$(SEED)"

balance:
	$(GO) run ./cmd/netrise balance --seed-from "$(SEED_FROM)" --seed-to "$(SEED_TO)"

smoke:
	$(GO) run ./cmd/netrise --help
	$(GO) run ./cmd/netrise validate-content
	$(GO) run ./cmd/netrise playtest --faction runner --seed "$(SEED)"
	$(GO) run ./cmd/netrise playtest --faction corp --seed "$(SEED)"
	$(GO) run ./cmd/netrise balance --seed-from 1 --seed-to 5
	$(GO) run ./cmd/netrise simulate --faction runner --seed "$(SEED)" --ticks "$(TICKS)"
	$(GO) run ./cmd/netrise simulate --faction corp --seed "$(SEED)" --ticks "$(TICKS)"
