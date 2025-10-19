---
title: CacheAllExternal Feature
date: 2025-10-18
lastmod: 2025-10-19
status: in-progress
cSpell:ignore: statuscodes
---

## Status

- ✅ **Phase 0 Complete**: `RetryCachedErrors` semantics cleaned up, timeout
  caching now controlled by `CacheAllExternal`
- ✅ **Test Infrastructure**: Migrated to `testify/assert`, cache helpers, fast
  test targets
- ✅ **Phase 1 Increments 0-3 Complete**: Core discovery mode + feature
  interactions verified (IgnoreURLs, StripQueryString)
- 🚧 **Phase 1 In Progress**: Verifying timeout test coverage (Increment 4)

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

| Row | CheckExternal | CacheAllExternal | RetryCachedErrors | Links Checked? | Errors Retried? | What Gets Cached    | Use Case                             |
| --- | ------------- | ---------------- | ----------------- | -------------- | --------------- | ------------------- | ------------------------------------ |
| 1   | `true`        | `false`          | `true`            | ✓              | ✓               | 200, 4XX            | **Default/Legacy**                   |
| 2   | `true`        | `false`          | `false`           | ✓              | ✗               | 200, 4XX            | Failed links are not retried         |
| 3   | `true`        | `true`           | `true`            | ✓              | ✓               | 200, 4XX, TSC[^TSC] | Cache timeouts, but retry            |
| 4   | `true`        | `true`           | `false`           | ✓              | ✗               | 200, 4XX, TSC[^TSC] | **Fast re-runs** - cache & reuse all |
| 5   | `false`       | `false`          | (N/A)             | ✗              | N/A             | Nothing             | **Default skip** - no cache          |
| 6   | `false`       | `true`           | (N/A)             | ✗              | N/A             | unchecked links     | **Link discovery**                   |

[^TSC]:
    TSC = Tool-specific status code used by htmltest. See
    `@docs/tasks/design.md` for details.

### Key Insights

- `CacheAllExternal` extends caching behavior in **both** modes:
  - When `CheckExternal: true` → Caches **all tool-specific errors** (timeouts,
    network failures, etc.)
  - When `CheckExternal: false` → Caches discovered links (as `StatusUnchecked`)
- `RetryCachedErrors` now has clean, single-purpose semantics (retry vs. reuse
  cached errors)
- When `CheckExternal: false`, `RetryCachedErrors` has no effect (nothing to
  retry)
- See `@docs/tasks/design.md` for status code details

### Tool-Specific Errors to Cache

When `CacheAllExternal: true` and `CheckExternal: true`, **all** non-HTTP errors
should be cached with appropriate status codes:

1. **Timeout errors** (`StatusTimeout = -10`):
   - Request exceeds `ExternalTimeout` setting
   - Currently: ✅ Already implemented in Phase 0

2. **DNS/Network errors** (`StatusNetworkError = -20`):
   - "dial tcp" failures
   - DNS lookup failures
   - Connection refused
   - Currently: ❌ Not cached (returns early without caching)

3. **Certificate errors** (`StatusCertError = -30`):
   - x509.UnknownAuthorityError
   - Invalid/expired certificates
   - Incomplete certificate chains
   - Currently: ❌ Not cached (returns early without caching)

4. **Generic client errors** (`StatusClientError = -40`):
   - Other unhandled HTTP client errors
   - Currently: ❌ Not cached (returns early without caching)

**Note**: HTTP status codes (200, 404, etc.) are already cached in all modes.

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

Phase 1 focuses on implementing **Discovery Mode** (rows 5-6 from Complete
Behavior Matrix). Timeout caching (rows 3-4) was already implemented in Phase 0.

| Test# | Incr | Matrix Row | Behavior to Test               | CheckExternal-related config[^Config]            | Expected Result              | Test Name                                  | Status       |
| ----- | ---- | ---------- | ------------------------------ | ------------------------------------------------ | ---------------------------- | ------------------------------------------ | ------------ |
| —     | —    | Row 1      | Default/Legacy                 | Check: true, All: false, Retry: true             | 200, 4XX cached & retried    | `TestExternalErrorCachedRetried`           | ✅ Existing  |
| —     | —    | Row 2      | Errors not retried             | Check: true, All: false, Retry: false            | 200, 4XX cached & reused     | `TestExternalBrokenRetryCachedErrorsDisabled` | ✅ Existing  |
| —     | —    | Row 3      | Timeout cached & retried       | Check: true, All: true, Retry: true              | Timeout cached but retried   | _(no explicit test yet)_                   | ⚠️ TODO Inc 4 |
| —     | —    | Row 4      | Timeout cached & reused        | Check: true, All: true, Retry: false             | Timeout cached & reused      | `TestTimeoutCachedReused`                  | ✅ Phase 0   |
| 1     | 0    | Row 5      | Default: no caching            | Check: false, All: false                         | Nothing cached               | `TestCacheAllExternalDisabled`             | ✅ Inc 0     |
| 2     | 1    | Row 6      | **Discovery mode**             | Check: false, All: true                          | Cache with `StatusUnchecked` | `TestCacheAllExternalDiscovery`            | ✅ Inc 1     |
| 3     | 2    | —          | IgnoreURLs interaction         | Check: false, All: true, IgnoreURLs: `[pattern]` | Ignored URLs NOT cached      | `TestCacheAllExternalAndIgnoredURLs`       | ✅ Inc 2     |
| 4     | 3    | —          | StripQueryString interaction   | Check: false, All: true, StripQueryString: true  | Query stripped before cache  | `TestCacheAllExternalQueryString`          | ✅ Inc 3     |

[^Config]:
    Abbreviations: `Check` = `CheckExternal`, `All` = `CacheAllExternal`,
    `Retry` = `RetryCachedErrors`.

**Commands**:
- `make test-tdd-fast TEST_RUN=<TestName>` - Fast TDD (skip slow tests, ~1.5s)
- `make test-tdd TEST_RUN=<TestName>` - Full TDD (includes slow tests, ~7.8s)
- `make test-tdd-cache TEST_RUN=<TestName>` - With cache inspection

**Note**: Phase 1 focuses on discovery mode (row 6) and its edge cases. Rows 1-4
from the Complete Behavior Matrix are already covered by existing tests.

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

### Phase 0: Cleanup `RetryCachedErrors` Semantics ✅ COMPLETE

**Purpose**: Separate concerns before adding new functionality

**Step 0.1**: Write test for new semantics

- ✅ Created `TestTimeoutNotCachedWithRetryCacheErrorsOnly` - verifies timeouts
  are NOT cached when only `RetryCachedErrors: false` (without
  `CacheAllExternal`)
- ✅ Test run: **GREEN** (existing code already has correct behavior for this
  case)

**Step 0.2**: Update existing test to expect new behavior (RED)

- ✅ Updated `TestTimeoutIsCached` to use `CacheAllExternal: true` instead of
  `RetryCachedErrors: false`
- ✅ Test cannot run yet (references non-existent option)

**Step 0.3**: Change implementation to use new option (RED)

- ✅ Changed `check-link.go` timeout caching condition from
  `if !hT.opts.RetryCachedErrors` to `if hT.opts.CacheAllExternal`
- ✅ Result: **Compilation error** - option doesn't exist yet

**Step 0.4**: Add option to make code compile (GREEN)

- ✅ Added `CacheAllExternal bool` to `Options` struct (`htmltest/options.go`)
- ✅ Added default value `"CacheAllExternal": false` in `DefaultOptions()`
- ✅ Test run: **GREEN** - code compiles and test passes

**Step 0.5**: Update remaining timeout tests

- ✅ Updated `TestTimeoutCachedReused` to use `CacheAllExternal: true`
- ✅ Updated `TestTimeoutCachedMessage` to use `CacheAllExternal: true`
- ✅ All cache tests: **GREEN**

**Step 0.6**: Add sanity check test

- ✅ Created `TestCacheOptions` to verify default option values
- ✅ Test run: **GREEN**
- ✅ Full test suite: All packages pass

**Result**: Clean semantics established, timeout caching now controlled by
`CacheAllExternal` (not `RetryCachedErrors`), all tests passing, no regression

**Key TDD Insight**: We changed the implementation BEFORE the option existed,
causing a compilation error (deep RED), then added just enough to compile and
pass (GREEN)

**Infrastructure Improvements**:

1. **Testify migration**:
   - ✅ Migrated `check-link-cache_test.go` to
     `github.com/stretchr/testify/assert`
   - ✅ Benefits: Industry-standard library, better error messages, active
     maintenance
   - ✅ Cleaner assertion syntax: `assert.Equal(t, expected, actual)`

2. **Cache assertion helpers** (`test_helpers_extra_test.go`):
   - ✅ `tExpectCached(t, hT, url, statusCode...)` - Assert URL is cached
     (optionally with status)
   - ✅ `tExpectNotCached(t, hT, url)` - Assert URL is not cached
   - ✅ Reduces boilerplate from ~7 lines to 1 line per assertion

3. **Slow test control**:
   - ✅ Added `tSkipSlow(t)` helper with `-skip-slow` flag
   - ✅ Applied to 4 cache timeout tests (wait 1s each)
   - ✅ Updated Makefile with `TESTFLAGS` support
   - ✅ Added fast test targets:
     - `make test-fast` - All packages, skip slow tests (~7s)
     - `make test-tdd-fast TEST_RUN='...'` - TDD mode, skip slow tests (~1.5s)
   - ✅ Result: 76% faster cache test runs (1.5s vs 7.8s)

4. **Test refactoring** (DRY patterns):
   - ✅ Extract `fixture` variables for paths
   - ✅ Extract `opts` maps and reuse across test runs
   - ✅ Consistent test structure across all cache tests

---

### Phase 1: Discovery Mode Feature

**Scope**: Implement `StatusUnchecked` caching when `CheckExternal: false` and
`CacheAllExternal: true`.

**Note**: Phase 1 focuses on discovery mode only. Caching additional tool-specific
errors (network failures, cert errors) will be addressed in future phases.

#### Increment 0: Test #1 - Default Behavior (Baseline) ✅

**Purpose**: Establish regression test for row 5

- ✅ Test already exists: `TestCacheAllExternalDisabled`
- ✅ Verifies `CacheAllExternal: false` doesn't cache discovered links
- ✅ Test: PASS (baseline behavior confirmed)
- ✅ No work needed: Baseline already covered

#### Increment 1: Test #2 - Discovery Mode (Core Feature) ✅

**Purpose**: Implement row 6 - the main use case

**Actual TDD Steps**:

- ✅ Wrote `TestCacheAllExternalDiscovery` test
- ✅ Test run: **RED** (URL not cached, feature doesn't exist)
- ✅ Fixed `tExpectCached` helper to avoid panic on failed assertion
- ✅ Implemented discovery mode in `checkExternal()`:
  - Early return when `CheckExternal: false && CacheAllExternal: false`
  - Fall through to URL processing when `CacheAllExternal: true`
  - Cache with `StatusUnchecked` and return
- ✅ Test run: **GREEN**
- ✅ Refactored: Moved URL processing before discovery mode check (DRY)
- ✅ Added invariant check to document control flow assumption
- ✅ All cache tests: **GREEN**

**Result**: Core discovery mode feature working! External links cached with
`StatusUnchecked` when `CheckExternal: false` and `CacheAllExternal: true`.

#### Increment 2: Test #5 - IgnoreURLs Feature Interaction ✅

**Purpose**: Verify CacheAllExternal respects IgnoreURLs patterns

- ✅ Wrote `TestCacheAllExternalAndIgnoredURLs`
- ✅ Created `tExpectCacheEmpty` helper (more robust than checking specific URL)
- ✅ Test run: **GREEN** (URL processing order is correct - ignored URLs filtered
  before caching)
- ✅ No code changes needed
- ✅ Updated `TestCacheAllExternalDisabled` and
  `TestTimeoutNotCachedWithRetryCacheErrorsOnly` to use `tExpectCacheEmpty`

**Result**: Feature interaction verified - ignored URLs correctly excluded from
cache

#### Increment 3: Test #6 - StripQueryString Feature Interaction ✅

**Purpose**: Verify CacheAllExternal works with query string stripping

- ✅ Wrote `TestCacheAllExternalQueryString` using `check_just_once.html` fixture
- ✅ Tests `github.com` URLs (not in default `StripQueryExcludes`)
- ✅ Test run: **GREEN** (query stripping happens before caching)
- ✅ No code changes needed
- ✅ Verifies: Multiple URLs with different query params → single cached entry

**Result**: Query string stripping correctly applied before caching in discovery
mode

#### Increment 4: Verify Row 3 & Row 4 Coverage

**Purpose**: Ensure timeout+retry combinations work

- Verify `TestTimeoutCachedReused` covers row 4 (cache + reuse)
- Consider if we need explicit test for row 3 (cache + retry)
- Existing test infrastructure may already cover this

**Rationale for order**: Clean up semantics first (Phase 0), then add
infrastructure and core discovery feature (Phase 1), validate assumptions while
fresh, verify complete matrix coverage.

## To-dos

### Phase 0: Cleanup

- [x] Step 0.1-0.6: All steps complete ✅

### Phase 1: Discovery Mode

- [x] Increment 0: Baseline test (default behavior - row 5) ✅
- [x] Increment 1: Discovery mode test + implementation (row 6) ✅
- [x] Increment 2: IgnoreURLs feature interaction ✅
- [x] Increment 3: StripQueryString feature interaction ✅
- [ ] Increment 4: Verify row 3 & 4 coverage

### Documentation

- [ ] Update README configuration table

## Future Phases

### Phase 2: Network Error Caching (Future)

**Scope**: Cache DNS and network failures with `StatusNetworkError = -20`

- Add `StatusNetworkError` constant to `statuscodes.go`
- Modify "dial tcp" error handling in `check-link.go` to cache when
  `CacheAllExternal: true`
- Add tests for network error caching
- Covers: DNS lookup failures, connection refused, network unreachable

### Phase 3: Certificate Error Caching (Future)

**Scope**: Cache certificate validation errors with `StatusCertError = -30`

- Add `StatusCertError` constant to `statuscodes.go`
- Modify x509 error handling in `check-link.go` to cache when `CacheAllExternal:
  true`
- Add tests for certificate error caching
- Covers: Unknown authority, expired certs, incomplete chains

### Phase 4: Generic Client Error Caching (Future)

**Scope**: Cache other HTTP client errors with `StatusClientError = -40`

- Add `StatusClientError` constant to `statuscodes.go`
- Modify generic error handling in `check-link.go` to cache when
  `CacheAllExternal: true`
- Add tests for generic error caching
- Covers: All other unhandled HTTP client errors

**Note**: These phases follow the same TDD approach as Phases 0 and 1. Each error
type gets its own status code and test coverage.
