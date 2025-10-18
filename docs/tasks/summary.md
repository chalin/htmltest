---
title: Summary of Changes on dev/main Branch
date: 2025-10-18
lastmod: 2025-10-18
status: active
---

# Summary of Changes on dev/main Branch

This document summarizes the changes made on the `dev/main` branch relative to
`master`.

## Features

So far all features are related to external link checking, and whether such
links are cached and/or retried on subsequent encounters.

### 1. RetryCachedErrors

This Introduced the `RetryCachedErrors` configuration option. When set to false,
This feature allows htmltest to reuse cached error status codes (e.g., 404)
without retrying them on subsequent runs.

- **Config parameter**: `RetryCachedErrors` (boolean, default: `true`)
- **Behavior**: When set to `false`, cached errors are reused instead of
  retrying
- **Backward compatibility**: Default `true` maintains existing behavior

### 2. Timeout Caching

**Status**: Completed

Extended caching behavior to save timeout errors to the refcache when
`RetryCachedErrors: false`.

- Timeouts are cached as status code `-10` (tool-specific timeout code)
- Prevents repeated timeout attempts on unreachable URLs
- Works in conjunction with `RetryCachedErrors` option

### 3. Status Code Migration

**Status**: Completed

Migrated timeout status codes from `408` to `-10` for clarity.

- Created `htmltest/statuscodes.go` with status code constants
- `StatusTimeout = -10` (tool-specific timeout)
- `StatusUnchecked = 0` (for future use)
- Helper functions: `IsHTTPStatus()`, `IsUnchecked()`, `IsToolError()`
- See `@docs/tasks/design.md` for status code conventions

### 4. URL Encoding in Cache

**Status**: Completed

Fixed URL escaping in the JSON refcache file.

- Disabled HTML escaping when encoding URLs to JSON
- URLs are now stored in their unescaped form in `refcache.json`
- Improves readability and prevents double-escaping issues

## Infrastructure & Tooling

### Version ID Suffix

Added repository-specific suffix to version IDs for better build tracking.

### CI/CD Improvements

- Added `workflow_dispatch` trigger to GitHub Actions CI workflow
- Allows manual workflow runs from GitHub UI

### Project Structure

- Added `Makefile` for build automation
- Added `CONTRIBUTING.md` with development guidelines

## Infrastructure & Tooling (continued)

### TDD Makefile Targets

**Status**: Completed

Added Makefile targets for TDD workflow:

- `make test-tdd TEST_RUN=TestName` - Run test with clean cache
- `make test-tdd-cache TEST_RUN=TestName` - Run test and show cache state
- `make clean-cache` - Remove refcache file
- Supports pattern matching for running multiple tests

### Documentation Structure

**Status**: Completed

Organized documentation under `docs/`:

- `docs/ops/` - Operational documentation (session-start, agent-guidance)
- `docs/tasks/` - Task and feature documentation
- `AGENTS.md` - AI agent guidance (Cursor auto-loads)
- `.github/copilot-instructions.md` - GitHub Copilot support

## Planned Features

Most features are documented under the `docs/tasks/` directory.

### CacheAllExternal (In Progress)

See `@docs/tasks/cache-unchecked-external-links.md` for details. This feature
will allow caching external links without checking them when
`CheckExternal: false`.
