# Cache Timed-Out External Links

## Overview

Currently, when external link checks timeout, the URL is not cached, causing
htmltest to retry the same slow URLs on every run. This change caches timeout
results using `http.StatusRequestTimeout` (408).

## Changes

### 1. Save timeouts to refcache (`htmltest/check-link.go`)

When a timeout error occurs in `checkExternal()`, save the URL to cache with
`http.StatusRequestTimeout` (408) **only if `RetryCachedErrors` is disabled**.
This preserves original behavior by default.

### 2. ~~Recognize 408 as valid cached status~~ (Not needed)

No change to `statusCodeValid()` - it should only accept truly successful codes
(200, 206). The `RetryCachedErrors` feature handles accepting cached 408s when
disabled.

### 3. Handle 408 in status code switch (`htmltest/check-link.go`)

Add a case for `http.StatusRequestTimeout` in the status code switch to report
cached timeouts as errors with message "request exceeded our ExternalTimeout
(cached)".

### 4. Add tests (`htmltest/check-link-cache_test.go`)

Add tests to verify:

- Timeouts are cached with `http.StatusRequestTimeout` (408) when
  `RetryCachedErrors: false`
- Timeouts are NOT cached by default (backward compatibility)
- Cached timeout results are retrieved on subsequent runs (with
  `RetryCachedErrors: false`)
- Errors are still reported for cached timeouts

## Benefits

- Avoids repeatedly checking URLs that consistently timeout
- Speeds up subsequent test runs when timeouts occur
- Reduces load on external servers that are slow/unreliable
- Maintains error reporting while improving performance
- Works automatically with `RetryCachedErrors` feature (when disabled, cached
  408s are trusted like other errors)

## Notes

- Thread-safe: Uses existing refcache mutex protection
- Status code 408: `http.StatusRequestTimeout` - standard HTTP status code for
  request timeouts
- Uses stdlib constant, no magic numbers
- Conditional caching: Timeouts only cached when `RetryCachedErrors: false`
- Backward compatible: Default behavior unchanged (timeouts always retried)

## To-dos

- [x] Save timeout errors to refcache with `http.StatusRequestTimeout` (408)
      when `RetryCachedErrors: false`
- [x] Add case for `http.StatusRequestTimeout` to report cached timeouts
- [x] Add tests in check-link-cache_test.go (4 tests added)
