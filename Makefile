# Makefile
# cSpell:ignore TESTFLAGS coverprofile gopath golangci ldflags covermode coverpkg gofmt benchmem

.PHONY: build build-verify install run
.PHONY: test test-fast test-race test-coverage test-ci test-bench
.PHONY: test-tdd test-tdd-fast test-tdd-cache clean-cache refcache-check _test-tdd
.PHONY: lint fmt fmt-check vet check ci
.PHONY: clean deps help

# Binary name
BINARY := htmltest

# Version from git tags
VERSION := $(shell git describe --tags 2>/dev/null || echo "dev")
VERSION := $(VERSION)-chalin-dev
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -X main.version=$(VERSION) -X main.date=$(BUILD_DATE)

# Test flags (can be overridden: make test TESTFLAGS="-v -run TestName")
TESTFLAGS ?=

# TDD variables (can be overridden)
TEST_PKG ?= ./htmltest
TEST_RUN ?= .
REFCACHE := htmltest/tmp/.htmltest/refcache.json

# Default target
.DEFAULT_GOAL := help

## build: Build the binary
build:
	@echo "Building $(BINARY) $(VERSION)..."
	@go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) main.go
	@echo "Built bin/$(BINARY)"

## build-verify: Build and verify the binary works (smoke test like CI)
build-verify: build
	@echo "Verifying build..."
	@bin/$(BINARY) -v
	@bin/$(BINARY) -h > /dev/null
	@echo "Running smoke tests..."
	@bin/$(BINARY) -c htmldoc/fixtures/conf.yaml -l0
	@bin/$(BINARY) htmldoc/fixtures/documents/dir1
	@bin/$(BINARY) htmltest/fixtures/links/head_link_href.html
	@echo "Build verification complete!"

## test: Run all tests (use TESTFLAGS to pass additional flags, e.g., make test TESTFLAGS="-v")
test:
	@echo "Running ALL tests (includes slow and external tests)..."
	@go test ./... $(TESTFLAGS)

## test-fast: Run all tests but skip slow ones (timeout waits)
test-fast:
	@echo "Running tests (skipping slow tests)..."
	@go test ./htmldoc ./issues ./refcache
	@go test ./htmltest -skip-slow=true

## test-race: Run tests with race detector (recommended)
test-race:
	@echo "Running tests with race detector (includes slow and external tests)..."
	@go test -v -race ./...

## test-coverage: Generate and display test coverage
test-coverage:
	@echo "Generating coverage report (includes slow and external tests)..."
	@go test -coverprofile=coverage.txt ./... $(TESTFLAGS)
	@go tool cover -func=coverage.txt
	@echo ""
	@echo "To view HTML coverage report, run: go tool cover -html=coverage.txt"

## test-ci: Run tests exactly as CI does (race + coverage)
test-ci:
	@echo "Running tests with race detector and coverage (CI mode)..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...
	@go tool cover -func=coverage.txt


## test-bench: Run benchmark tests
test-bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./htmltest

## clean-cache: Remove test cache files
clean-cache:
	@echo "Cleaning test cache..."
	@rm -f $(REFCACHE)
	@echo "Cache cleaned"

## test-tdd: TDD mode - run specific test(s) with clean cache (use TEST_RUN to match tests).
## Examples:
##   make test-tdd TEST_RUN=TestTimeoutIsCached                      # Run single test
##   make test-tdd TEST_RUN='.*Cache.*'                              # Run all cache tests
##   make test-tdd-fast TEST_RUN='.*Cache.*'                         # Fast mode (skip slow tests)
##   make test-tdd TEST_RUN='.*Cache.*' TESTFLAGS=-skip-slow=true    # Same as -fast
##   make test-tdd-cache TEST_RUN=TestTimeout                        # With cache inspection
##   make clean-cache                                                # Just clean the cache
test-tdd: clean-cache _test-tdd

## test-tdd-fast: Fast TDD mode - same as test-tdd but skips slow tests
test-tdd-fast:
	@$(MAKE) test-tdd TESTFLAGS=-skip-slow=true TEST_RUN="$(TEST_RUN)" TEST_PKG="$(TEST_PKG)"

_test-tdd:
	@echo "TDD mode: Running $(TEST_RUN) in $(TEST_PKG)..."
	@go test -v -run $(TEST_RUN) $(TEST_PKG) $(TESTFLAGS) || true
	@echo ""
	@echo "Test completed."

refcache-check:
	@echo "Refcache state:"
	@if [ -f $(REFCACHE) ]; then \
		echo "=== $(REFCACHE) ==="; \
		cat $(REFCACHE) | python3 -m json.tool 2>/dev/null || cat $(REFCACHE); \
	else \
		echo "No refcache file created"; \
	fi

test-tdd-cache: clean-cache _test-tdd refcache-check

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

## clean: Remove build artifacts and temporary files (includes test cache)
clean: clean-cache
	@echo "Cleaning..."
	@go clean
	@rm -rf bin/
	@rm -rf tmp/
	@rm -f coverage.txt
	@echo "Clean complete"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY)..."
	@go install -ldflags "$(LDFLAGS)"
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY)"

## run: Run the application (requires arguments, e.g., make run ARGS="-h")
run:
	@go run main.go $(ARGS)

## deps: Download and verify dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@echo "Dependencies ready"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## fmt-check: Check if code is formatted (CI mode - fails if not formatted)
fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -d .)" || (echo "Code is not formatted. Run 'make fmt' to fix." && gofmt -d . && exit 1)

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## check: Run format check, vet, and tests (CI-like checks)
check: fmt-check vet test-ci
	@echo "All checks passed!"

## ci: Alias for check (matches CI pipeline)
ci: check

## help: Display this help message
help:
	@echo "$(BINARY) - Makefile commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "TDD Examples:"
	@echo "  make test-tdd TEST_RUN=TestTimeoutIsCached    # Run single test with clean cache"
	@echo "  make test-tdd TEST_RUN='.*Cache.*'            # Run all cache tests"
	@echo "  make test-tdd TEST_RUN=TestTimeout            # Run tests matching pattern"
	@echo "  make test-clean TEST_RUN=TestTimeout          # Same but without cache inspection"
	@echo "  make clean-cache                              # Just clean the cache"

