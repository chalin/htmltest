---
title: CacheAllExternal Feature
date: 2025-10-18
lastmod: 2025-10-19
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

### RetryCachedErrors behavior cleanup

`RetryCachedErrors` currently conflates two concerns:

1. Whether to retry cached errors (its intended purpose)
2. Whether to cache timeouts (side effect when set to `false`)

With `CacheAllExternal`, we can separate these orthogonal concerns:

- `CacheAllExternal` controls **what gets cached** (just checked results, or
  everything including timeouts/unchecked)
- `RetryCachedErrors` controls **retry behavior only** (use cached errors, or
  retry them)

This cleanup must happen BEFORE implementing the discovery mode feature.

### Complete Behavior Matrix

| CheckExternal | CacheAllExternal | RetryCachedErrors | Links Checked? | Errors Retried? | What Gets Cached  | Use Case                             |
| ------------- | ---------------- | ----------------- | -------------- | --------------- | ----------------- | ------------------------------------ |
| `true`        | `false`          | `true`            | ✓              | ✓               | 200, 4XX          | **Default/Legacy**                   |
| `true`        | `false`          | `false`           | ✓              | ✗               | 200, 4XX          | Failed links are not retried         |
| `true`        | `true`           | `true`            | ✓              | ✓               | 200, 4XX, TSC[^1] | Cache timeouts, but retry            |
| `true`        | `true`           | `false`           | ✓              | ✗               | 200, 4XX, TSC[^1] | **Fast re-runs** - cache & reuse all |
| `false`       | `false`          | (N/A)             | ✗              | N/A             | Nothing           | **Default skip** - no cache          |
| `false`       | `true`           | (N/A)             | ✗              | N/A             | unchecked links   | **Link discovery**                   |

[^1]:
    TSC = Tool-specific status code used by htmltest. See
    `@docs/tasks/design.md` for details.

### Key Insights

- `CacheAllExternal` extends caching behavior in **both** modes:
  - When `CheckExternal: true` → Also caches timeouts (as `StatusTimeout`)
  - When `CheckExternal: false` → Caches discovered links (as `StatusUnchecked`)
- `RetryCachedErrors` now has clean, single-purpose semantics (retry vs. reuse
  cached errors)
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

**Best Practice**: Use `tTestFileOptsFromCleanOutputDir()` for the first test
call in each test function to ensure test isolation with a clean cache
directory. Use `tTestFileOpts()` for subsequent calls within the same test.

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

## Incremental Implementation Strategy

### Phase 0: Cleanup `RetryCachedErrors` Semantics

**Purpose**: Separate concerns before adding new functionality

#### Step 0a: Add config option

- Add `CacheAllExternal bool` to `Options` struct
- Add default value `false` in `DefaultOptions()`
- Run existing tests: Should PASS (option exists but unused)

#### Step 0b: Update existing timeout tests

- Update 3 existing tests to use `CacheAllExternal: true` instead of
  `RetryCachedErrors: false`:
  - `TestTimeoutIsCached`
  - `TestTimeoutCachedReused`
  - `TestTimeoutCachedMessage`
- Run tests: RED (tests expect new behavior, code uses old)

#### Step 0c: Refactor timeout caching code

- In `check-link.go`, change timeout caching condition from:
  - OLD: `if !hT.opts.RetryCachedErrors`
  - NEW: `if hT.opts.CacheAllExternal`
- Run tests: GREEN (existing tests verify behavior)

#### Step 0d: Verify row 2 behavior

- Ensure `TestTimeoutNotCachedByDefault` still passes
- This test verifies row 2: `CacheAllExternal: false` +
  `RetryCachedErrors: false` = NO timeout caching

**Benefit**: Clean semantics established, existing tests ensure no regression

---

### Phase 1: Discovery Mode Feature

#### Increment 1: Test #1 - Default Behavior (Baseline)

**Purpose**: Establish regression test for row 5

- Write test: Verify `CacheAllExternal: false` doesn't cache discovered links
- Run test: Should PASS immediately (tests current behavior)
- No code needed: This is baseline
- Benefit: Protects against future regressions

#### Increment 2: Test #2 - Discovery Mode (Core Feature)

**Purpose**: Implement row 6 - the main use case

- Write test: Discovery mode caches with `StatusUnchecked`
- Run test: RED (config exists but no discovery caching code)
- Implement: Add discovery caching in `checkExternal()`
- Run test: GREEN
- **This is the core feature**

#### Increment 3: Test #4 - Ignored URLs

**Purpose**: Verify edge case works

- Write test: Ignored URLs not cached
- Run test: Likely PASS (existing `isURLIgnored()` should work)
- If RED: Fix discovery code to respect ignore patterns
- Benefit: Validates design assumption

#### Increment 4: Test #5 - Query String Stripping

**Purpose**: Verify edge case works

- Write test: Query strings stripped before caching
- Run test: Likely PASS (existing logic should work)
- If RED: Adjust operation order
- Benefit: Validates design assumption

#### Increment 5: Verify Row 3 & Row 4 Coverage

**Purpose**: Ensure timeout+retry combinations work

- Verify `TestTimeoutCachedReused` covers row 4 (cache + reuse)
- Consider if we need explicit test for row 3 (cache + retry)
- Existing test infrastructure may already cover this

**Rationale for order**: Clean up semantics first (Phase 0), then add
infrastructure and core discovery feature (Phase 1), validate assumptions while
fresh, verify complete matrix coverage.

## To-dos

### Phase 0: Cleanup

- [ ] Step 0a: Add `CacheAllExternal` config option
- [ ] Step 0b: Update 3 existing timeout tests
- [ ] Step 0c: Refactor timeout caching code
- [ ] Step 0d: Verify row 2 behavior

### Phase 1: Discovery Mode

- [ ] Increment 1: Baseline test (default behavior - row 5)
- [ ] Increment 2: Discovery mode test + implementation (row 6)
- [ ] Increment 3: Ignored URLs edge case
- [ ] Increment 4: Query stripping edge case
- [ ] Increment 5: Verify row 3 & 4 coverage

### Documentation

- [ ] Update README configuration table
