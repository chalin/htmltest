# Respect Cached Error Status Codes

## Overview

Currently, when external URLs return error status codes (404, 403, 429, etc.),
htmltest saves them to the refcache but retries them on subsequent runs because
`statusCodeValid()` only accepts 200 and 206. This plan adds an
`IgnoreCachedErrors` config option to trust cached error responses.

## Problem

Users who post-process the refcache after htmltest runs want to prevent htmltest
from retrying known-bad URLs. Currently, htmltest rechecks every cached 4xx/5xx
status code.

## Solution

Add `IgnoreCachedErrors` boolean config option (default: `false` for backward
compatibility). When enabled, htmltest will trust ALL cached status codes
(including errors) and not retry them, but will still report errors
appropriately.

## Changes

### 1. Add IgnoreCachedErrors option (`htmltest/options.go`)

Add `IgnoreCachedErrors bool` field to the `Options` struct and set default to
`false` in `DefaultOptions()`.

### 2. Modify cache check logic (`htmltest/check-link.go`)

Update the `checkExternal()` function to accept cached results when:

- The status code is valid (200/206), OR
- `IgnoreCachedErrors` is enabled (trusts all cached status codes)

### 3. Add tests (`htmltest/check-link_test.go`)

Add tests to verify:

- Error status codes (404, etc.) are cached
- By default, cached errors are retried (backward compatibility)
- When `IgnoreCachedErrors: true`, cached errors are reused without retry
- Cached errors still report as errors (not silently ignored)

### 4. Update documentation

Update `README.md` to document the new `IgnoreCachedErrors` option.

## Files to modify

- `htmltest/options.go` (2 locations)
- `htmltest/check-link.go` (1 location)
- `htmltest/check-link_test.go` (new tests)
- `README.md` (documentation)

## Thread Safety

The refcache is already thread-safe, using `sync.RWMutex` for concurrent access:

- Multiple goroutines can read from cache simultaneously (RLock)
- Writes are exclusive (Lock)

This feature change only affects the read path (deciding whether to accept a
cached value) and does not introduce any new concurrency issues. The existing
mutex protection remains sufficient.

## Backward Compatibility

- **Default behavior unchanged:** `IgnoreCachedErrors` defaults to `false`
- **No existing tests will break:** No tests currently verify cached error retry
  behavior
- **Opt-in feature:** Users must explicitly enable to change behavior

## Benefits

- Prevents unnecessary retries of known-bad URLs
- Faster test runs when many cached errors exist
- Allows post-processing workflows that manipulate refcache
- Maintains backward compatibility (opt-in feature)
- Still reports errors so users know about broken links
- Safe for concurrent checking (TestFilesConcurrently option)

## To-dos

- [x] Add IgnoreCachedErrors to Options struct and defaults
- [x] Update checkExternal to accept all cached status codes when option enabled
- [x] Add tests for cached error handling behavior (3 comprehensive tests)
- [x] Document IgnoreCachedErrors option in README
