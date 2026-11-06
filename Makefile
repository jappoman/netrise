GO ?= go
SEED ?= 17
TICKS ?= 600
FACTION ?= runner
DATA_DIR ?=

.PHONY: help fmt vet test race verify build bench validate-content play corp simulate

help:
	@printf '%s\n' 'NETRISE commands:' '  make play [DATA_DIR=/path]                         start the Runner TUI' '  make corp [DATA_DIR=/path]                         start the Corporation TUI' '  make simulate [FACTION=runner SEED=17 TICKS=600]     run a deterministic report' '  make test                                            run unit tests' '  make verify                                          format, vet, test and race-test'

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
