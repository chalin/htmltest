# Contributing to htmltest

This document provides guidelines and information for developers.

## Project Overview

**htmltest** is a fast HTML validation and link checker written in Go. It's
designed as a faster alternative to the Ruby-based html-proofer tool.

### What htmltest Does

- Validates HTML files for broken links, images, and script references
- Checks internal and external links
- Verifies image alt attributes
- Tests meta refresh tags
- Checks for valid favicon references
- Validates DOCTYPE declarations
- Caches external link checks for faster subsequent runs

### Performance

On sites with 2000+ files, htmltest runs in seconds compared to minutes with
other tools, making it ideal for CI/CD pipelines.

## Development Setup

### Prerequisites

- **Go 1.13 or higher** (check with `go version`)
- **Git**
- **golangci-lint** (optional, for linting)

### Getting Started

1. **Clone the repository:**

   ```bash
   git clone https://github.com/wjdp/htmltest.git
   cd htmltest
   ```

2. **Install dependencies:**

   ```bash
   go mod download
   ```

   Or using Make:

   ```bash
   make deps
   ```

3. **Build the project:**

   ```bash
   make build
   ```

   Or directly:

   ```bash
   go build -o bin/htmltest main.go
   ```

4. **Run tests:**
   ```bash
   make test-race
   ```

## Project Structure

```
htmltest/
├── main.go             # Entry point
├── htmldoc/            # HTML document parsing and storage
│   ├── document.go     # Document representation
│   ├── reference.go    # Reference (link/image/script) handling
│   └── attr.go         # HTML attribute utilities
├── htmltest/           # Core testing logic
│   ├── htmltest.go     # Main test orchestration
│   ├── options.go      # Configuration options
│   ├── check-*.go      # Specific checkers (links, images, etc.)
│   └── fixtures/       # Test fixtures
├── issues/             # Issue tracking and reporting
├── refcache/           # External reference caching
└── output/             # Output formatting and logging
```

## Running Tests

### Unit Tests

Run all tests across all packages:

```bash
make test
```

### Tests with Race Detection (Recommended)

Catches concurrency issues:

```bash
make test-race
```

### Tests Exactly as CI Runs Them

Run tests with race detection AND coverage (matches GitHub Actions CI):

```bash
make test-ci
```

### Test Specific Package

```bash
go test ./htmldoc
go test ./htmltest
go test ./issues
go test ./refcache
```

### Test Coverage

```bash
make test-coverage
```

View HTML coverage report:

```bash
go tool cover -html=coverage.txt
```

### Benchmark Tests

```bash
make test-bench
```

### Run Specific Test

```bash
go test -v -run TestMissingOptions ./htmltest
```

## Test Structure

### Fixtures

Tests use fixture files located in `*/fixtures/` directories:

- HTML test files with various scenarios (broken links, valid links, etc.)
- Configuration files
- Sample resources (images, scripts)

### VCR Cassettes

External HTTP requests are mocked using VCR cassettes (recorded HTTP interactions) in `htmltest/fixtures/vcr/*.cassette`. This allows tests to run offline and consistently without hitting real external URLs.

### Test Naming Convention

- Test files: `*_test.go`
- Test functions: `Test<Feature>(t *testing.T)`
- Benchmark functions: `Benchmark<Feature>(b *testing.B)`
- Helper functions: `t<Helper>()` (lowercase t prefix)

## Building

### Development Build

```bash
make build
```

This creates `bin/htmltest` with version information from git tags.

### Manual Build

```bash
go build -ldflags "-X main.version=$(git describe --tags)" -o bin/htmltest main.go
```

### Install to GOPATH

```bash
make install
```

## Code Quality

### Format Code

```bash
make fmt
```

### Check Formatting (CI Mode)

Check if code is properly formatted without modifying files (fails if not formatted):

```bash
make fmt-check
```

### Run go vet

```bash
make vet
```

### Lint

Requires [golangci-lint](https://golangci-lint.run/usage/install/):

```bash
make lint
```

### Run All CI Checks Locally

Run format check, vet, and tests exactly as CI does:

```bash
make check
# or
make ci
```

This is the best command to run before committing to ensure CI will pass.

## Testing Your Changes

### End-to-End Testing

After building, test the binary manually:

```bash
# Show help
./bin/htmltest -h

# Test a single file
./bin/htmltest htmltest/fixtures/links/head_link_href.html

# Test a directory
./bin/htmltest htmldoc/fixtures/documents/

# Test with config
./bin/htmltest -c htmldoc/fixtures/conf.yaml
```

### Testing with Fixtures

When adding new features:

1. Add appropriate test fixtures in `htmltest/fixtures/`
2. Create test cases in the relevant `*_test.go` file
3. Use the test helper functions (see `test_helpers_test.go`)
4. Ensure tests pass both locally and don't require external network access

Example test pattern:

```go
func TestNewFeature(t *testing.T) {
    hT := tTestFileOpts("fixtures/feature/test.html",
        map[string]interface{}{"NewOption": true})
    tExpectIssueCount(t, hT, 0)
}
```

## Makefile Commands

Run `make help` to see all available commands:

```bash
make build          # Build the binary
make test           # Run all tests
make test-race      # Run tests with race detector
make test-coverage  # Generate coverage report
make test-ci        # Run tests exactly as CI does (race + coverage)
make test-bench     # Run benchmarks
make lint           # Run linter
make clean          # Remove build artifacts
make install        # Install to GOPATH/bin
make fmt            # Format code
make fmt-check      # Check formatting (CI mode - fails if not formatted)
make vet            # Run go vet
make check          # Run all CI checks locally (fmt-check, vet, test-ci)
make ci             # Alias for check
make help           # Show all commands
```

## Submitting Changes

### Before Submitting

1. **Run all CI checks locally:** `make check` or `make ci`
   - This runs format checking, vet, and tests with race detection + coverage
   - Ensures your code will pass CI
2. **Update documentation** if needed
3. **Add tests** for new features

If you need to fix formatting issues, run `make fmt` before running checks again.

### Pull Request Guidelines

1. Create a descriptive branch name (e.g., `feature/check-canonical-links`)
2. Write clear commit messages
3. Include tests for new functionality
4. Update README.md if adding user-facing features
5. Ensure all CI checks pass

## Need Help?

- **Issues:** [Submit an issue](https://github.com/wjdp/htmltest/issues/new)
- **Documentation:** Check the [README](README.md) for user documentation
- **Code patterns:** Look at existing tests and checkers for examples

## Coding Conventions

- Follow standard Go conventions
- Use `go fmt` for formatting
- Write tests for new features
- Keep functions focused and small
- Document exported functions and types
- Use descriptive variable names

## License

By contributing to htmltest, you agree that your contributions will be licensed under the same license as the project (see [LICENCE](LICENCE)).

---

Thank you for contributing to htmltest! 🎉
