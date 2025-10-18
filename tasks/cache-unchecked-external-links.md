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
caches timeout errors (408), which were previously not cached.

Example workflow:

- Run htmltest with `CheckExternal: false` and `CacheAllExternal: true` during
  development
- External links are saved to refcache with status code 0 (unchecked)
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

| CheckExternal | CacheAllExternal | RetryCachedErrors | Links Checked? | Errors Retried? | What Gets Cached   | Use Case                             |
| ------------- | ---------------- | ----------------- | -------------- | --------------- | ------------------ | ------------------------------------ |
| `true`        | `false`          | `true`            | ✓              | ✓               | 200, 404 (NOT 408) | **Default/Legacy**                   |
| `true`        | `false`          | `false`           | ✓              | ✗               | 200, 404 (NOT 408) | Reuse errors, retry timeouts         |
| `true`        | `true`           | `true`            | ✓              | ✓               | 200, 404, 408      | Cache timeouts, but retry            |
| `true`        | `true`           | `false`           | ✓              | ✗               | 200, 404, 408      | **Fast re-runs** - cache & reuse all |
| `false`       | `false`          | (N/A)             | ✗              | N/A             | Nothing            | **Default skip** - no cache          |
| `false`       | `true`           | (N/A)             | ✗              | N/A             | 0 (unchecked)      | **Link discovery**                   |

### Key Insights

- `CacheAllExternal` extends caching behavior in **both** modes:
  - When `CheckExternal: true` → Also caches timeouts (408)
  - When `CheckExternal: false` → Caches discovered links (0)
- When `CheckExternal: false`, `RetryCachedErrors` has no effect (nothing to
  retry)
- Status code `0` indicates "unchecked/unknown" status
- Status code `408` indicates timeout error

## Implementation Steps (TDD Approach)

### 1. Write Tests First

**File**: `htmltest/check-link-cache_test.go`

Add test cases to verify:

- **Discovery mode**: Links cached with status 0 when `CheckExternal: false` and
  `CacheAllExternal: true`
- **Default behavior**: Links NOT cached when `CacheAllExternal: false`
- **Ignored URLs**: Ignored URLs still not cached even with
  `CacheAllExternal: true`
- **Query string handling**: Query string stripping applied correctly
- **Timeout caching**: Timeouts cached as 408 when `CheckExternal: true` and
  `CacheAllExternal: true`

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
  - If so, cache discovered links with status 0
  - Apply URL processing (strip query string if configured)
  - Skip ignored URLs (respect `isURLIgnored()` check)

**B) Timeout caching (CheckExternal: true):**

- In timeout handling code (around line 201-210)
- Change condition from `if !hT.opts.RetryCachedErrors` to
  `if hT.opts.CacheAllExternal`
- This caches timeouts when CacheAllExternal is enabled

### 4. Update Documentation

**File**: `README.md`

- Add new row in the configuration options table (around line 145, after
  `CheckExternal`)
- Format: `| \`CacheAllExternal\` | Cache all external links including timeouts
  (when checking) and unchecked links (when not checking). Useful for fast
  re-runs and link discovery. | \`false\` |`

## Key Design Decisions

- **Minimal change**: Only ONE new option added, reuses existing `CheckExternal`
  and `RetryCachedErrors`
- **Orthogonal design**: Three independent dimensions that work together
  naturally
- **Clear semantics**: "CacheAll" clearly means "cache everything, not just
  successes"
- **Status codes**:
  - `0` = unchecked/unknown (discovery mode)
  - `408` = timeout error (check mode)
  - Other codes = actual HTTP responses
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
