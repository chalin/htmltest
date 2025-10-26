---
title: Agent-helper guidance
date: 2025-10-18
lastmod: 2025-10-18
status: active
---

This is the htmltest project - a fast HTML validation and link checker written
in Go.

## Key Files to Reference

When starting a session or before making changes, review:

- `@docs/ops/session-start.md` - Session start checklist and context
- `@docs/tasks/summary.md` - Summary of all changes on dev/main
- `@docs/tasks/design.md` - Design decisions (status codes, etc.)
- `@CONTRIBUTING.md` - Development practices

## Development Practices

- **Use TDD**: Write tests first (RED), then implementation (GREEN)
- **Use Makefile targets**: `make test-tdd` and `make test-tdd-cache` for
  development

## Current Work

- Branch: `dev/main`
- Current feature: CacheAllExternal - **COMPLETE** ✅ (see
  `@docs/tasks/cache-unchecked-external-links.md`)

## Testing Commands

Examples:

```bash
make test-tdd TEST_RUN=TestName        # Run specific test with clean cache
make test-tdd TEST_RUN='.*Cache.*'     # Run all cache tests
make test-tdd-cache TEST_RUN=TestName  # Same but shows cache state
```
