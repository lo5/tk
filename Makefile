.PHONY: help
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=tk
BUILD_DIR=.
INSTALL_PATH=~/.local/bin

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Test flags
TEST_FLAGS=-v -race
COVERAGE_FLAGS=-coverprofile=coverage.out -covermode=atomic
INTEGRATION_FLAGS=-tags integration

help: ## Show this help message
	@echo "GoTK - Go Ticket Keeper"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the binary
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v

build-all: ## Build for all platforms
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64

install: build ## Install binary to system
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

uninstall: ## Uninstall binary from system
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)

clean: ## Remove build artifacts and test files
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-*
	rm -f coverage.out coverage.html
	rm -rf .tickets-test-*

# Test targets
test: ## Run all unit tests
	$(GOTEST) $(TEST_FLAGS) ./...

test-unit: ## Run unit tests only (exclude integration)
	$(GOTEST) $(TEST_FLAGS) -short ./...

test-integration: ## Run integration tests only
	$(GOTEST) $(TEST_FLAGS) $(INTEGRATION_FLAGS) ./...

test-coverage: ## Run tests with coverage report
	$(GOTEST) $(TEST_FLAGS) $(COVERAGE_FLAGS) ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-summary: ## Show coverage summary
	$(GOTEST) $(COVERAGE_FLAGS) ./...
	$(GOCMD) tool cover -func=coverage.out

test-verbose: ## Run tests with verbose output
	$(GOTEST) -v ./...

# Code quality targets
fmt: ## Format code with gofmt
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

lint: ## Run golangci-lint (requires golangci-lint installed)
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/"; exit 1)
	golangci-lint run

# Dependency management
deps: ## Download dependencies
	$(GOMOD) download

deps-tidy: ## Tidy dependencies
	$(GOMOD) tidy

deps-verify: ## Verify dependencies
	$(GOMOD) verify

deps-update: ## Update dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Development targets
run: build ## Build and run with example command
	./$(BINARY_NAME) ls

dev: ## Quick development cycle (fmt, vet, test, build)
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) test-unit
	@$(MAKE) build

check: ## Run all checks (fmt, vet, test, coverage)
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) test
	@$(MAKE) test-coverage-summary

# CI targets
ci: ## Run CI pipeline (all checks and tests)
	@$(MAKE) deps-verify
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) test-coverage
	@$(MAKE) test-integration
	@$(MAKE) build

ci-short: ## Run quick CI pipeline (no integration tests)
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) test-unit
	@$(MAKE) build

# Benchmarking targets
bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

bench-cpu: ## Run CPU benchmarks with profiling
	$(GOTEST) -bench=. -benchmem -cpuprofile=cpu.prof ./...
	@echo "View with: go tool pprof cpu.prof"

bench-mem: ## Run memory benchmarks with profiling
	$(GOTEST) -bench=. -benchmem -memprofile=mem.prof ./...
	@echo "View with: go tool pprof mem.prof"

# Documentation targets
docs: ## Generate documentation
	@echo "Generating documentation..."
	$(GOCMD) doc -all ./...

docs-server: ## Start godoc server
	@echo "Starting documentation server at http://localhost:6060"
	godoc -http=:6060

# Utility targets
version: ## Show Go version
	@$(GOCMD) version

list-tools: ## List required development tools
	@echo "Required tools for development:"
	@echo "  - go ($(shell go version))"
	@echo "  - golangci-lint (optional, for linting)"
	@echo "  - godoc (optional, for documentation server)"

todo: ## Show TODO comments in code
	@grep -rn "TODO\|FIXME\|XXX" --include="*.go" . || echo "No TODOs found"

lines: ## Count lines of code
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1

# Watch targets (requires entr or similar)
watch-test: ## Watch files and run tests on change (requires entr)
	@which entr > /dev/null || (echo "entr not installed. Install with: brew install entr"; exit 1)
	find . -name "*.go" | entr -c make test-unit

watch-build: ## Watch files and rebuild on change (requires entr)
	@which entr > /dev/null || (echo "entr not installed. Install with: brew install entr"; exit 1)
	find . -name "*.go" | entr -c make build
