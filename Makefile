# Everything CI enforces, runnable locally.
BASE  ?= master
GOBIN := $(shell go env GOPATH)/bin
WORK  := $(shell mktemp -d)

# What counts as a regression: slower by both THRESH percent and MINNS
# nanoseconds. Needing both is what keeps a 15ns check growing to 85ns from
# failing a parse that still costs half a microsecond, while a parse that doubles
# does fail. Comparing two runs on one machine is what makes the percentage mean
# anything on a CI runner.
THRESH ?= 30
MINNS  ?= 300

.PHONY: all test lint fix bench benchstat benchguard

## all: everything that has to pass before a change goes up
all: fix lint test

## test: the suite, with coverage
test:
	go test -cover ./...

## lint: formatting, vet and the linter CI runs
lint: $(GOBIN)/goimports
	@test -z "$$($(GOBIN)/goimports -l .)" || { echo "goimports:"; $(GOBIN)/goimports -l .; exit 1; }
	golangci-lint run ./...

## fix: apply the Go modernizers, then format and fix up imports
fix: $(GOBIN)/goimports
	go fix ./...
	@$(GOBIN)/goimports -w .

## bench: the benchmarks, on their own
bench:
	go test -run=XXX -bench=. -benchmem .

## benchstat: the benchmarks against BASE, which is the only number worth quoting
benchstat: $(GOBIN)/benchstat
	@git worktree add -q --detach $(WORK)/base $(BASE)
	@echo "measuring $(BASE)..."
	@cd $(WORK)/base && go test -run=XXX -bench=. . > $(WORK)/base.out
	@echo "measuring the working tree..."
	@go test -run=XXX -bench=. . > $(WORK)/head.out
	@git worktree remove --force $(WORK)/base
	@cd $(WORK) && $(GOBIN)/benchstat base.out head.out

## benchguard: the same comparison, as CI grades it
#
# Six samples of a tenth of a second each, where a local run takes one of a full
# second: benchstat will not call a difference below four samples, and the gate
# then has nothing to grade.
benchguard: $(GOBIN)/benchstat
	@git worktree add -q --detach $(WORK)/base $(BASE)
	@echo "measuring $(BASE)..."
	@cd $(WORK)/base && go test -run=XXX -bench=. -count=6 -benchtime=100ms . > $(WORK)/base.out
	@echo "measuring the working tree..."
	@go test -run=XXX -bench=. -count=6 -benchtime=100ms . > $(WORK)/head.out
	@git worktree remove --force $(WORK)/base
	@cd $(WORK) && $(GOBIN)/benchstat base.out head.out
	@cd $(WORK) && $(GOBIN)/benchstat -format csv base.out head.out 2>/dev/null \
		| awk -v thresh=$(THRESH) -v minns=$(MINNS) -f $(CURDIR)/scripts/benchguard.awk

$(GOBIN)/benchstat:
	go install golang.org/x/perf/cmd/benchstat@latest

$(GOBIN)/goimports:
	go install golang.org/x/tools/cmd/goimports@latest
