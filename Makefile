SHELL := /bin/sh
.DEFAULT_GOAL := help

GO ?= go
MODE ?= client
BINARY_NAME ?= discoverd
CMD_PATH ?= ./cmd/discoverd
BIN_DIR ?= ./bin
DIST_DIR ?= ./dist
BINARY ?= $(BIN_DIR)/$(BINARY_NAME)
PKGS ?= ./...
VERSION ?= dev
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS ?= -X github.com/napilab/discoverd/version.Version=$(VERSION) -X github.com/napilab/discoverd/version.GitCommit=$(GIT_COMMIT) -X github.com/napilab/discoverd/version.BuildTime=$(BUILD_TIME)
LDFLAGS_RELEASE := -s -w
EXT ?=

.PHONY: help build version release xgo deb release-deb run run-client run-server test test-race vet lint fmt tidy check ci clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nAvailable targets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-14s %s\n", $$1, $$2} END {printf "\n"}' $(MAKEFILE_LIST)

build: ## Build binary into ./bin
	@mkdir -p $(BIN_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD_PATH)

release: clean ## Cross-compile release binaries for major platforms
	@echo "Cross-compiling for all platforms..."
	@$(MAKE) xgo GOOS=linux GOARCH=amd64
	@$(MAKE) xgo GOOS=linux GOARCH=arm64
	@$(MAKE) xgo GOOS=windows GOARCH=amd64 EXT=.exe
	@$(MAKE) xgo GOOS=windows GOARCH=arm64 EXT=.exe
	@$(MAKE) xgo GOOS=darwin GOARCH=arm64
	@$(MAKE) xgo GOOS=darwin GOARCH=amd64
	@echo "Cross-compilation completed!"
	@$(MAKE) release-deb


deb: ## Build .deb package for linux (requires nfpm; set GOARCH and VERSION)
	@test -n "$(GOARCH)" || { echo "GOARCH is required"; exit 1; }
	@command -v nfpm >/dev/null 2>&1 || { echo "nfpm is required. Install: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest"; exit 1; }
	@mkdir -p $(DIST_DIR) .deb-staging
	@cp bin/linux-$(GOARCH)/discoverd .deb-staging/discoverd
	GOARCH=$(GOARCH) VERSION=$(VERSION) nfpm pkg --packager deb --target $(DIST_DIR)/discoverd_$(VERSION)_linux_$(GOARCH).deb
	@rm -rf .deb-staging

release-deb: ## Build .deb packages for linux/amd64 and linux/arm64
	@$(MAKE) deb GOARCH=amd64
	@$(MAKE) deb GOARCH=arm64

xgo: ## Build single target binary (requires GOOS and GOARCH)
	@test -n "$(GOOS)" || { echo "GOOS is required"; exit 1; }
	@test -n "$(GOARCH)" || { echo "GOARCH is required"; exit 1; }
	@mkdir -p $(BIN_DIR)/$(GOOS)-$(GOARCH)
	@echo "Building for $(GOOS)/$(GOARCH)..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build \
		-trimpath \
		-ldflags "$(LDFLAGS) $(LDFLAGS_RELEASE)" \
		-o $(BIN_DIR)/$(GOOS)-$(GOARCH)/$(BINARY_NAME)$(EXT) $(CMD_PATH)
	@ls -lh $(BIN_DIR)/$(GOOS)-$(GOARCH)/$(BINARY_NAME)$(EXT)

version: ## Print build/runtime version info
	$(GO) run -ldflags "$(LDFLAGS)" $(CMD_PATH) --version

run: ## Run app (override with MODE=client|server)
	$(GO) run $(CMD_PATH) --mode $(MODE)

run-client: ## Run discovery client mode
	$(GO) run $(CMD_PATH) --mode client

run-server: ## Run discovery server mode
	$(GO) run $(CMD_PATH) --mode server

test: ## Run unit tests
	$(GO) test $(PKGS)

test-race: ## Run tests with race detector
	$(GO) test -race $(PKGS)

vet: ## Run go vet
	$(GO) vet $(PKGS)

lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint is required. Install from https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run

fmt: ## Format Go code
	$(GO) fmt $(PKGS)

tidy: ## Tidy Go module dependencies
	$(GO) mod tidy

check: fmt test vet ## Run fast local checks

ci: tidy fmt test-race vet lint ## Run full validation pipeline

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) $(DIST_DIR) .deb-staging
