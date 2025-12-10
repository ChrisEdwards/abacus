GO ?= go
GOLANGCI_LINT ?= golangci-lint
BINARY ?= abacus
PKG ?= ./cmd/abacus

# Version information
VERSION ?= dev
BUILD ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILDTIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Linker flags for version injection
LDFLAGS := -X main.Version=$(VERSION) -X main.Build=$(BUILD) -X main.BuildTime=$(BUILDTIME)

.PHONY: help build test test-verbose bench install lint clean check check-verbose check-test

help: ## Display available make targets
	@awk 'BEGIN {FS=":.*##"; printf "\nUsage: make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_\-]+:.*##/ {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Compile the abacus CLI into ./abacus
	@if [ -n "$$VERBOSE" ]; then \
		$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG); \
	else \
		. ./hack/run_silent.sh && run_silent "Building $(BINARY)" "$(GO) build -ldflags \"$(LDFLAGS)\" -o $(BINARY) $(PKG)"; \
	fi

## Check targets (linting and static analysis)

check: ## Run all checks (quiet output)
	@if [ -n "$$VERBOSE" ]; then \
		$(GO) fmt ./... && $(GO) vet ./... && $(GOLANGCI_LINT) run ./...; \
	else \
		$(MAKE) check-quiet; \
	fi

check-quiet:
	@. ./hack/run_silent.sh && print_main_header "Running Checks"
	@. ./hack/run_silent.sh && print_header "abacus" "Static analysis"
	@. ./hack/run_silent.sh && run_with_quiet "Format check passed" "$(GO) fmt ./..."
	@. ./hack/run_silent.sh && run_with_quiet "Vet check passed" "$(GO) vet ./..."
	@. ./hack/run_silent.sh && run_with_quiet "Lint check passed" "$(GOLANGCI_LINT) run ./..."

check-verbose: ## Run checks with verbose output
	@VERBOSE=1 $(MAKE) check

## Test targets

test: ## Run all unit tests (quiet output)
	@if [ -n "$$VERBOSE" ]; then \
		$(GO) test -v ./...; \
	else \
		$(MAKE) test-quiet; \
	fi

test-quiet:
	@. ./hack/run_silent.sh && print_main_header "Running Tests"
	@. ./hack/run_silent.sh && print_header "abacus" "Unit tests"
	@. ./hack/run_silent.sh && run_silent_with_test_count "Unit tests passed" "$(GO) test -json ./..." "go"

test-verbose: ## Run tests with verbose output
	@VERBOSE=1 $(MAKE) test

## Combined targets

check-test: ## Run all checks and tests
	@$(MAKE) check
	@$(MAKE) test

## Other targets

bench: ## Run benchmarks
	$(GO) test -run=^$$ -bench=. ./...

install: ## Install the CLI into GOPATH/bin
	$(GO) install -ldflags "$(LDFLAGS)" $(PKG)

lint: ## Run golangci-lint (verbose)
	$(GOLANGCI_LINT) run ./...

clean: ## Remove build artifacts
	rm -f $(BINARY)
