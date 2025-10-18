---
title: Session Start Checklist
date: 2025-10-18
lastmod: 2025-10-18
status: active
cSpell:ignore: oneline
---

# Session Start Checklist

At the beginning of each work session, reference these key files to provide
context:

## Essential Context Files

- `@AGENTS.md` - AI agent guidance and conventions (auto-loaded by Cursor)
- `@docs/tasks/summary.md` - Overview of all changes on dev/main branch
- `@docs/tasks/design.md` - Key design decisions (status codes, etc.)
- `@CONTRIBUTING.md` - Development workflow and practices
- `@README.md` - User-facing documentation and config options

## Current Work

- `@docs/tasks/cache-unchecked-external-links.md` - CacheAllExternal feature (in
  progress)
- `@docs/tasks/migrate-status-codes.md` - Status code migration (completed)

## How to Use

At the start of a session, say:

```
Review @AGENTS.md
```

Cursor auto-loads `docs/AGENTS.md`, but explicitly reviewing ensures full
context.

Then mention the specific task you're working on:

```
We're implementing CacheAllExternal. Review:
@docs/tasks/cache-unchecked-external-links.md
```

## Quick Status Check

Use these commands to see current state:

```bash
git status                    # What files are changed
git log --oneline -10        # Recent commits
make help                     # Available commands
```

## Key Design Decisions to Remember

1. **Status codes**: 0 = unchecked, -10 = timeout, >0 = HTTP codes
2. **TDD workflow**: Use `make test-tdd` and `make test-tdd-cache`
3. **Three config dimensions**: CheckExternal, RetryCachedErrors,
   CacheAllExternal
