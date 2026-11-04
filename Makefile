GO ?= go
SEED ?= 17
TICKS ?= 600
DATA_DIR ?=

.PHONY: help fmt vet test race verify play simulate

help:
	@printf '%s\n' 'NETRISE commands:' '  make play [DATA_DIR=/path]        start the TUI' '  make simulate [SEED=17 TICKS=600] run a deterministic report' '  make test                         run unit tests' '  make verify                       format, vet, test and race-test'

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

verify: fmt vet test race

play:
	@if [ -n "$(DATA_DIR)" ]; then $(GO) run ./cmd/netrise --data-dir "$(DATA_DIR)"; else $(GO) run ./cmd/netrise; fi

simulate:
	$(GO) run ./cmd/netrise simulate --seed "$(SEED)" --ticks "$(TICKS)"
