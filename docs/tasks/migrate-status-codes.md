---
title: Migrate Status Codes from 408 to -10
date: 2025-10-18
lastmod: 2025-10-18
status: in-progress
---

# Migrate Status Codes from 408 to -10

## Overview

Migrate timeout status codes from `408` (HTTP Request Timeout) to `-10`
(tool-specific timeout) to clarify the distinction between server responses and
client-side errors.

## Rationale

- **Old**: `408` - Ambiguous (is it from the server or our timeout?)
- **New**: `-10` - Clearly a tool-specific timeout error
- Aligns with new status code conventions (see `tasks/design.md`)

## Implementation Steps (TDD Approach)

### 1. Write Tests First (RED)

**File**: `htmltest/statuscodes_test.go` (new file)

Write tests for the new constants and helper functions before they exist:

```go
func TestStatusCodeConstants(t *testing.T) {
    // Test constant values
    if StatusUnchecked != 0 {
        t.Errorf("Expected StatusUnchecked to be 0, got %d", StatusUnchecked)
    }
    if StatusTimeout != -10 {
        t.Errorf("Expected StatusTimeout to be -10, got %d", StatusTimeout)
    }
}

func TestIsHTTPStatus(t *testing.T) {
    // Positive numbers are HTTP
    if !IsHTTPStatus(200) { t.Error("200 should be HTTP status") }
    if !IsHTTPStatus(404) { t.Error("404 should be HTTP status") }

    // Zero and negative are not
    if IsHTTPStatus(0) { t.Error("0 should not be HTTP status") }
    if IsHTTPStatus(-10) { t.Error("-10 should not be HTTP status") }
}

func TestIsUnchecked(t *testing.T) {
    if !IsUnchecked(0) { t.Error("0 should be unchecked") }
    if IsUnchecked(200) { t.Error("200 should not be unchecked") }
    if IsUnchecked(-10) { t.Error("-10 should not be unchecked") }
}

func TestIsToolError(t *testing.T) {
    if !IsToolError(-10) { t.Error("-10 should be tool error") }
    if IsToolError(0) { t.Error("0 should not be tool error") }
    if IsToolError(200) { t.Error("200 should not be tool error") }
}
```

**File**: `htmltest/check-link-cache_test.go`

Update existing timeout test to expect -10 instead of 408:

```go
func TestTimeoutIsCached(t *testing.T) {
    // ... existing test setup ...

    // Verify the timeout was saved with StatusTimeout (-10)
    cR, ok := hT.refCache.Get("http://5.6.7.8")
    if !ok {
        t.Error("expected timeout to be cached")
    }
    if cR.StatusCode != StatusTimeout {  // Changed from http.StatusRequestTimeout
        t.Errorf("expected status code %d (StatusTimeout), got %d",
            StatusTimeout, cR.StatusCode)
    }
}
```

**Run tests** - They should FAIL (constants and functions don't exist yet)

### 2. Create Status Code Constants File (GREEN)

**File**: `htmltest/statuscodes.go` (new file)

```go
package htmltest

// Status code conventions:
//   > 0  : Real HTTP status codes (200, 404, etc.)
//   = 0  : Unchecked/undiscovered (neutral state)
//   < 0  : Tool-specific error states
//
// Negative codes use -10 spacing to allow for future expansion

const (
	StatusUnchecked = 0   // Link discovered but not checked
	StatusTimeout   = -10 // Client-side timeout

	// Future expansion:
	//   -11 to -19: Timeout variations (connection, read, DNS)
	//   -20 to -29: Network errors
	//   -30 to -39: SSL/TLS errors
)

// IsHTTPStatus returns true if the status code is a real HTTP response
func IsHTTPStatus(code int) bool {
	return code > 0
}

// IsUnchecked returns true if the link was discovered but not checked
func IsUnchecked(code int) bool {
	return code == 0
}

// IsToolError returns true if the status is a tool-specific error
func IsToolError(code int) bool {
	return code < 0
}
```

**Run tests** - They should now PASS

### 3. Update Cache Writing Code (GREEN)

Update implementation to write -10 for timeouts:

**File**: `htmltest/check-link.go`

Find timeout handling code (around lines 201-210):

```go
// OLD
if !hT.opts.RetryCachedErrors {
    hT.refCache.Save(urlStr, http.StatusRequestTimeout)
}

// NEW
if !hT.opts.RetryCachedErrors {
    hT.refCache.Save(urlStr, StatusTimeout)
}
```

**Run tests** - Should still pass

### 4. Update Cache Reading Code (GREEN)

Add backward compatibility for reading old 408 codes:

**File**: `htmltest/check-link.go`

Update the status code switch statement (around lines 253-288) to handle both
old and new timeout codes for backward compatibility:

```go
case http.StatusRequestTimeout:
    // Legacy: old cache files may have 408
    fallthrough
case StatusTimeout:
    hT.issueStore.AddIssue(issues.Issue{
        Level:     issueLevel,
        Message:   "request exceeded our ExternalTimeout (cached)",
        Reference: ref,
    })
```

**Run tests** - Should still pass

### 5. Update statusCodeValid Function (GREEN)

**File**: `htmltest/check-link.go` (or wherever this helper lives)

Update the `statusCodeValid` function to recognize negative status codes as
invalid (errors):

```go
func statusCodeValid(code int) bool {
    return IsHTTPStatus(code) && code >= 200 && code < 400
}
```

**Run tests** - All tests should pass

### 6. Run Full Test Suite

```bash
make test
```

All tests should pass. Verify:

- New constants work correctly
- Helper functions behave as expected
- Timeout caching writes -10
- Cache reading handles both 408 (old) and -10 (new)
- No regressions in existing functionality

## TDD Cycle Summary

1. **RED**: Write tests for new constants/functions → Tests fail
2. **GREEN**: Create `statuscodes.go` → Tests pass
3. **GREEN**: Update cache writing to use `-10` → Tests pass
4. **GREEN**: Update cache reading for backward compat → Tests pass
5. **GREEN**: Update `statusCodeValid` → Tests pass
6. **REFACTOR**: (if needed) Clean up any code duplication

## Checklist (TDD Order)

- [ ] Write tests for constants/helpers in `htmltest/statuscodes_test.go` (RED)
- [ ] Update `htmltest/check-link-cache_test.go` to expect -10 (RED)
- [ ] Run tests → should FAIL
- [ ] Create `htmltest/statuscodes.go` with constants and helpers (GREEN)
- [ ] Run tests → should PASS
- [ ] Update `htmltest/check-link.go` to write -10 for timeouts (GREEN)
- [ ] Update `htmltest/check-link.go` to read both 408 and -10 (GREEN)
- [ ] Update `statusCodeValid` function if needed (GREEN)
- [ ] Run full test suite → should PASS
- [ ] Manual verification: check cache file contains -10

## Backward Compatibility

- **Reading**: Accept both 408 (old) and -10 (new) when reading cache
- **Writing**: Always write -10 for new timeout entries
- **No migration needed**: Old cache entries with 408 will continue to work
