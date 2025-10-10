# Makefile
# cSpell:ignore htmltest TESTFLAGS coverprofile gopath golangci covermode coverpkg gofmt benchmem

.PHONY: build build-verify test test-race test-coverage test-ci test-bench lint clean install run fmt fmt-check vet check ci deps help

# Binary name
BINARY := htmltest

# Version from git tags
VERSION := $(shell git describe --tags 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -X main.version=$(VERSION) -X main.date=$(BUILD_DATE)

# Test flags (can be overridden: make test TESTFLAGS="-v -run TestName")
TESTFLAGS ?=

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
	@echo "Running tests..."
	@go test $(TESTFLAGS) ./...

## test-race: Run tests with race detector (recommended)
test-race:
	@echo "Running tests with race detector..."
	@go test -v -race ./...

## test-coverage: Generate and display test coverage
test-coverage:
	@echo "Generating coverage report..."
	@go test $(TESTFLAGS) -coverprofile=coverage.txt ./...
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

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

## clean: Remove build artifacts and temporary files
clean:
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

