---
title: Summary of Changes on dev/main Branch
date: 2025-10-18
lastmod: 2025-10-19
status: active
---

# Summary of Changes on dev/main Branch

This document summarizes the changes made on the `dev/main` branch relative to
`master`.

## Features

All features enhance external link checking and caching behavior.

### 1. CacheAllExternal

Cache all external links, including timeouts and tool-specific errors, for fast
re-runs and link discovery mode.

- **Config parameter**: `CacheAllExternal` (boolean, default: `false`)
- **Discovery mode**: When `CheckExternal: false`, discovered links are cached
  with `StatusUnchecked = 0`
- **Error caching**: When `CheckExternal: true`, caches timeouts, network
  errors, certificate errors, and generic client errors
- **Status codes**:
  - `StatusUnchecked = 0` - Discovered but not checked
  - `StatusTimeout = -10` - Timeout errors
  - `StatusNetworkError = -20` - DNS/connection failures
  - `StatusCertError = -30` - Certificate errors (expired, untrusted, etc.)
  - `StatusClientError = -40` - Generic client errors
- **Tests**: 16+ tests covering discovery mode, feature interactions, and error
  caching
- **Fixtures**: `link_expired_cert.html` for cert error testing

See `@docs/tasks/cache-unchecked-external-links.md` for full details.

### 2. RetryCachedErrors

Control whether cached errors are retried on subsequent runs.

- **Config parameter**: `RetryCachedErrors` (boolean, default: `true`)
- **Behavior**: When set to `false`, cached errors are reused without retrying
- **Use case**: Fast re-runs using only cached results (combine with
  `CacheAllExternal: true`)
- **Backward compatibility**: Default `true` maintains existing behavior

### 3. URL Encoding in Cache

Fixed URL escaping in the JSON refcache file.

- Disabled HTML escaping when encoding URLs to JSON
- URLs stored in their unescaped form in `refcache.json`
- Improves readability and prevents double-escaping issues

## Infrastructure & Tooling

### Status Code System

Custom status code system for tool-specific states:

- `htmltest/statuscodes.go` with constants and helpers
- Positive integers: HTTP status codes
- Zero: Unchecked/undiscovered links
- Negative integers: Tool-specific errors (timeout, network, cert, client)
- Helper functions: `IsHTTPStatus()`, `IsUnchecked()`, `IsToolError()`
- See `@docs/tasks/design.md` for conventions

### TDD Makefile Targets

Makefile targets for TDD workflow:

- `make test-tdd TEST_RUN=TestName` - Run test with clean cache
- `make test-tdd-fast` - Skip slow tests (timeouts)
- `make test-tdd-cache TEST_RUN=TestName` - Run test and show cache state
- `make clean-cache` - Remove refcache file
- Supports pattern matching for running multiple tests

### Documentation Structure

Organized documentation under `docs/`:

- `docs/ops/` - Operational documentation (session-start)
- `docs/tasks/` - Task and feature documentation
- `AGENTS.md` - AI agent guidance (Cursor auto-loads)
- `.github/copilot-instructions.md` - GitHub Copilot support

### Project Structure

- Added `Makefile` for build automation
- Added `CONTRIBUTING.md` with development guidelines
- Version ID suffix for better build tracking

### CI/CD

- Added `workflow_dispatch` trigger to GitHub Actions
- Updated action versions
- Prevents duplicate runs on PRs
