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

.PHONY: help build test bench install lint clean

help: ## Display available make targets
	@awk 'BEGIN {FS=":.*##"; printf "\nUsage: make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_\-]+:.*##/ {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Compile the abacus CLI into ./abacus
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

test: ## Run all unit tests
	$(GO) test ./...

bench: ## Run benchmarks
	$(GO) test -run=^$$ -bench=. ./...

install: ## Install the CLI into GOPATH/bin
	$(GO) install -ldflags "$(LDFLAGS)" $(PKG)

lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run ./...

clean: ## Remove build artifacts
	rm -f $(BINARY)
