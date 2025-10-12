# RetryCachedErrors Config Option

## Overview

Currently, when external URLs return error status codes (404, 403, 429, etc.),
htmltest saves them to the refcache but retries them on subsequent runs because
`statusCodeValid()` only accepts 200 and 206. This plan adds a
`RetryCachedErrors` config option to control retry behavior for cached errors.

## Problem

Users who post-process the refcache after htmltest runs want to prevent htmltest
from retrying known-bad URLs. Currently, htmltest rechecks every cached 4xx/5xx
status code.

## Solution

Add `RetryCachedErrors` boolean config option (default: `true` to preserve
current behavior). When true, retries external links even if cached with error
status codes. When false, trusts all cached results without retry, but still
reports errors appropriately.

## Changes

### 1. Add RetryCachedErrors option (`htmltest/options.go`)

Add `RetryCachedErrors bool` field to the `Options` struct and set default to
`true` in `DefaultOptions()`.

### 2. Modify cache check logic (`htmltest/check-link.go`)

Update the `checkExternal()` function to accept cached results when:

- The status code is valid (200/206), OR
- `RetryCachedErrors` is disabled (trusts all cached status codes)

### 3. Add tests (`htmltest/check-link-cache_test.go`)

Add tests to verify:

- Error status codes (404, etc.) are cached
- By default (RetryCachedErrors=true), cached errors are retried (backward compatibility)
- When `RetryCachedErrors: false`, cached errors are reused without retry
- Cached errors still report as errors (not silently ignored)

### 4. Update documentation

Update `README.md` to document the new `RetryCachedErrors` option.

## Files to modify

- `htmltest/options.go` (2 locations)
- `htmltest/check-link.go` (1 location)
- `htmltest/check-link-cache_test.go` (new test file)
- `README.md` (documentation)

## Thread Safety

The refcache is already thread-safe, using `sync.RWMutex` for concurrent access:

- Multiple goroutines can read from cache simultaneously (RLock)
- Writes are exclusive (Lock)

This feature change only affects the read path (deciding whether to accept a
cached value) and does not introduce any new concurrency issues. The existing
mutex protection remains sufficient.

## Backward Compatibility

- **Default behavior unchanged:** `RetryCachedErrors` defaults to `true` (retries errors)
- **No existing tests will break:** No tests currently verify cached error retry
  behavior
- **Opt-in feature:** Users must set to `false` to skip retries

## Benefits

- Prevents unnecessary retries of known-bad URLs
- Faster test runs when many cached errors exist
- Allows post-processing workflows that manipulate refcache
- Maintains backward compatibility (opt-in feature)
- Still reports errors so users know about broken links
- Safe for concurrent checking (TestFilesConcurrently option)

## To-dos

- [x] Add RetryCachedErrors to Options struct and defaults
- [x] Update checkExternal to accept all cached status codes when option disabled
- [x] Add tests for cached error handling behavior (3 comprehensive tests)
- [x] Document RetryCachedErrors option in README
