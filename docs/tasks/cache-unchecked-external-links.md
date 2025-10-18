---
title: CacheAllExternal Feature
date: 2025-10-18
lastmod: 2025-10-18
status: in-progress
---

# CacheAllExternal Feature

## Use Case

When running htmltest with external link checking disabled
(`CheckExternal: false`), it's useful to still discover and save all external
links to the cache file. This allows you to:

1. **Inventory external links** - Know which external URLs are referenced in
   your HTML files without the overhead of checking them
2. **Post-process discovered links** - Extract the list of external links from
   the cache file for analysis, reporting, or batch checking with other tools
3. **Build a cache incrementally** - Collect external links during
   development/build without slowing down the process, then validate them
   separately

Additionally, when checking is enabled (`CheckExternal: true`), this option also
caches timeout errors (as `StatusTimeout`), which were previously not cached.

Example workflow:

- Run htmltest with `CheckExternal: false` and `CacheAllExternal: true` during
  development
- External links are saved to refcache with `StatusUnchecked` (not checked)
- Later, extract the list of unchecked links from the cache for validation or
  reporting
- Optionally run htmltest with `CheckExternal: true` to validate cached links

## Feature Design

### Configuration Options (3 orthogonal dimensions)

1. **`CheckExternal: true | false`** (existing) - Whether to check external
   links
2. **`RetryCachedErrors: true | false`** (existing) - Whether to retry cached
   errors
3. **`CacheAllExternal: true | false`** (NEW) - Whether to cache everything
   including timeouts and unchecked links

### Complete Behavior Matrix

| CheckExternal | CacheAllExternal | RetryCachedErrors | Links Checked? | Errors Retried? | What Gets Cached  | Use Case                             |
| ------------- | ---------------- | ----------------- | -------------- | --------------- | ----------------- | ------------------------------------ |
| `true`        | `false`          | `true`            | ✓              | ✓               | 200, 4XX          | **Default/Legacy**                   |
| `true`        | `false`          | `false`           | ✓              | ✗               | 200, 4XX          | Reuse errors, retry timeouts         |
| `true`        | `true`           | `true`            | ✓              | ✓               | 200, 4XX, TSC[^1] | Cache timeouts, but retry            |
| `true`        | `true`           | `false`           | ✓              | ✗               | 200, 404, TSC[^1] | **Fast re-runs** - cache & reuse all |
| `false`       | `false`          | (N/A)             | ✗              | N/A             | Nothing           | **Default skip** - no cache          |
| `false`       | `true`           | (N/A)             | ✗              | N/A             | unchecked links   | **Link discovery**                   |

[^1]:
    TSC = Tool-specific status code used by htmltest. See
    `@docs/tasks/design.md` for details.

### Key Insights

- `CacheAllExternal` extends caching behavior in **both** modes:
  - When `CheckExternal: true` → Also caches timeouts (as `StatusTimeout`)
  - When `CheckExternal: false` → Caches discovered links (as `StatusUnchecked`)
- When `CheckExternal: false`, `RetryCachedErrors` has no effect (nothing to
  retry)
- See `@docs/tasks/design.md` for status code details

## Implementation Steps (TDD Approach)

**Incremental TDD**: Write one test at a time, implement, verify, then move to
next.

For each behavior in the matrix below:

1. Write ONE test for that specific behavior
2. Run test → **RED** (fails)
3. Write minimal code to make it pass
4. Run test → **GREEN** (passes)
5. Cleanup/refactor if needed
6. Move to next behavior

### Test Order & Behaviors

| #   | Behavior to Test    | Config                                             | Expected Result                  | Test Name                         |
| --- | ------------------- | -------------------------------------------------- | -------------------------------- | --------------------------------- |
| 1   | Default: no caching | `CheckExternal: false`, `CacheAllExternal: false`  | Nothing cached                   | `TestCacheAllExternalDisabled`    |
| 2   | **Discovery mode**  | `CheckExternal: false`, `CacheAllExternal: true`   | Cache with `StatusUnchecked`     | `TestCacheAllExternalDiscovery`   |
| 3   | **Timeout caching** | `CheckExternal: true`, `CacheAllExternal: true`    | Cache timeout as `StatusTimeout` | `TestCacheAllExternalTimeout`     |
| 4   | Ignored URLs        | `CacheAllExternal: true`, `IgnoreURLs: [pattern]`  | Ignored URLs NOT cached          | `TestCacheAllExternalIgnored`     |
| 5   | Query stripping     | `CacheAllExternal: true`, `StripQueryString: true` | Query stripped                   | `TestCacheAllExternalQueryString` |

**Commands**: Use `make test-tdd TEST_RUN=<TestName>` or
`make test-tdd-cache TEST_RUN=<TestName>`

### 1. Write Tests (One at a Time)

**File**: `htmltest/check-link-cache_test.go`

### 2. Add Configuration Option

**File**: `htmltest/options.go`

- Add `CacheAllExternal bool` field to the `Options` struct (around line 71,
  near other cache-related options)
- Add default value `"CacheAllExternal": false` in `DefaultOptions()` function
  (around line 143, near `EnableCache`)

### 3. Implement Caching Logic

**File**: `htmltest/check-link.go`

Two modifications needed:

**A) Discovery mode (CheckExternal: false):**

- Modify `checkExternal()` function (starting at line 129)
- When `!hT.opts.CheckExternal` is true:
  - Check if `hT.opts.CacheAllExternal` is enabled
  - If so, cache discovered links with `StatusUnchecked`
  - Apply URL processing (strip query string if configured)
  - Skip ignored URLs (respect `isURLIgnored()` check)

**B) Timeout caching (CheckExternal: true):**

- In timeout handling code (around line 201-210)
- Change condition from `if !hT.opts.RetryCachedErrors` to
  `if hT.opts.CacheAllExternal`
- Save with `StatusTimeout` instead of not caching

### 4. Update Documentation

**File**: `README.md`

- Add new row in the configuration options table (around line 145, after
  `CheckExternal`)
- Description: Cache all external links including timeouts (when checking is
  enabled) and unchecked links (when checking is disabled). Useful for fast
  re-runs and link discovery.
- Default: `false`

## Key Design Decisions

- **Minimal change**: Only ONE new option added, reuses existing `CheckExternal`
  and `RetryCachedErrors`
- **Orthogonal design**: Three independent dimensions that work together
  naturally
- **Clear semantics**: "CacheAll" clearly means "cache everything, not just
  successes"
- **Status codes**: See `@docs/tasks/design.md` for details
  - `StatusUnchecked` = unchecked/unknown (discovery mode)
  - `StatusTimeout` = timeout error (check mode)
  - Positive codes = actual HTTP responses
- **Respects existing patterns**:
  - URL ignore patterns (`IgnoreURLs`)
  - Query string stripping (`StripQueryString`)
- **Backward compatible**: Default `false` maintains existing behavior
- **No need for backward compat with dev/main features**: We can change
  `RetryCachedErrors` behavior if needed

## To-dos

- [ ] Write test cases for CacheAllExternal feature (TDD)
- [ ] Add CacheAllExternal field to Options struct and DefaultOptions()
- [ ] Implement discovery mode caching (CheckExternal: false)
- [ ] Implement timeout caching (CheckExternal: true)
- [ ] Add CacheAllExternal to README configuration table
