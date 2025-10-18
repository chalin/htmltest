package htmltest

import (
	"testing"

	"github.com/wjdp/htmltest/issues"
)

// Tests for cache-related functionality
// Feature: RetryCachedErrors
// Added by @chalin

// TestExternalErrorCached : Test that URLs with HTTP error status codes (such
// as 404) are saved to the refcache.
func TestExternalErrorCached(t *testing.T) {
	hT := tTestFileOpts("fixtures/images/imageExternal404.html",
		map[string]interface{}{"VCREnable": true, "EnableCache": true})
	tExpectIssueCount(t, hT, 1)

	// Verify the 404 was saved to cache
	cR, ok := hT.refCache.Get("https://upload.wikimedia.org/wikipedia/en/404")
	if !ok {
		t.Error("expected 404 response to be cached, but it wasn't")
	}
	// Verify it's actually a 404
	if cR.StatusCode != 404 {
		t.Errorf("expected status code 404 in cache, got %d", cR.StatusCode)
	}
}

// TestExternalErrorCachedRetried : Test that by default cached error responses
// (404, etc.) are retried on subsequent runs.
func TestExternalErrorCachedRetried(t *testing.T) {
	// First run: populate cache with 404
	hT := tTestFileOpts("fixtures/images/imageExternal404.html",
		map[string]interface{}{"VCREnable": true, "EnableCache": true})
	tExpectIssueCount(t, hT, 1)

	// Second run: WITH VCR - should retry even though 404 is cached (default behavior)
	hT2 := tTestFileOpts("fixtures/images/imageExternal404.html",
		map[string]interface{}{"VCREnable": true, "EnableCache": true, "LogLevel": issues.LevelDebug})
	tExpectIssueCount(t, hT2, 1)

	// Verify it did NOT use the cache (should see "fresh" not "from cache")
	if hT2.issueStore.MessageMatchCount("from cache") > 0 {
		t.Error("expected cached 404 to be retried (should NOT see 'from cache' message by default)")
	}
	if hT2.issueStore.MessageMatchCount("fresh") == 0 {
		t.Error("expected cached 404 to be retried (should see 'fresh' message)")
	}
}

// TestExternalBrokenRetryCachedErrorsDisabled : Test that URLs with non-OK
// status are not retried when RetryCachedErrors is set to false.
// Uses <q cite="..."> elements for variety.
func TestExternalBrokenRetryCachedErrorsDisabled(t *testing.T) {
	// First run: populate cache with 404
	hT := tTestFileOpts("fixtures/generic/citeBroken.html",
		map[string]interface{}{"VCREnable": true, "EnableCache": true, "RetryCachedErrors": false})
	tExpectIssueCount(t, hT, 4) // 4 broken citations in the fixture

	// Second run: WITHOUT VCR - should use cached 404s, not retry (which would fail without VCR)
	hT2 := tTestFileOpts("fixtures/generic/citeBroken.html",
		map[string]interface{}{"EnableCache": true, "RetryCachedErrors": false, "LogLevel": issues.LevelDebug})
	tExpectIssueCount(t, hT2, 4)

	// Verify it used the cache by checking for "from cache" messages
	if hT2.issueStore.MessageMatchCount("from cache") == 0 {
		t.Error("expected cached 404s to be reused (should see 'from cache' messages)")
	}
}

// ========================================
// Timeout Caching Tests
// ========================================

// TestTimeoutNotCachedByDefault : Test that by default, timeouts are NOT cached
// and are retried on every run. This ensures backward compatibility with the
// original behavior.
func TestTimeoutNotCachedByDefault(t *testing.T) {
	tSkipShortExternal(t)

	// First run: timeout occurs
	hT := tTestFileOpts("fixtures/links/ip_timeout.html",
		map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "LogLevel": issues.LevelDebug})
	tExpectIssueCount(t, hT, 1)
	tExpectIssue(t, hT, "request exceeded our ExternalTimeout", 1)

	// Second run: should retry (not use cache) because RetryCachedErrors is true (default)
	hT2 := tTestFileOpts("fixtures/links/ip_timeout.html",
		map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "LogLevel": issues.LevelDebug})
	tExpectIssueCount(t, hT2, 1)

	// Verify it retried (should see "fresh" and "hitting", not "from cache")
	if hT2.issueStore.MessageMatchCount("fresh") == 0 {
		t.Error("timeout should be retried by default (should see 'fresh' message)")
	}
	if hT2.issueStore.MessageMatchCount("hitting") == 0 {
		t.Error("timeout should be retried by default (should see 'hitting' message)")
	}
}

// TestTimeoutIsCached : Test that URLs that timeout are saved to the refcache
// when RetryCachedErrors is false.
func TestTimeoutIsCached(t *testing.T) {
	tSkipShortExternal(t)
	hT := tTestFileOpts("fixtures/links/ip_timeout.html",
		map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "RetryCachedErrors": false})
	tExpectIssueCount(t, hT, 1)
	tExpectIssue(t, hT, "request exceeded our ExternalTimeout", 1)

	// Verify the timeout was saved to cache with StatusTimeout
	cR, ok := hT.refCache.Get("http://5.6.7.8")
	if !ok {
		t.Error("expected timeout to be cached when RetryCachedErrors is false, but it wasn't")
	}
	if cR.StatusCode != StatusTimeout {
		t.Errorf("expected status code %d (StatusTimeout), got %d", StatusTimeout, cR.StatusCode)
	}
}

// TestTimeoutCachedReused : Test that cached timeout results are reused on
// subsequent runs without retrying the URL when RetryCachedErrors is false.
func TestTimeoutCachedReused(t *testing.T) {
	tSkipShortExternal(t)

	// First run: cause and cache a timeout
	hT := tTestFileOpts("fixtures/links/ip_timeout.html",
		map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "RetryCachedErrors": false})
	tExpectIssueCount(t, hT, 1)

	// Second run: WITHOUT timeout set - should use cached 524, not actually try the request
	hT2 := tTestFileOpts("fixtures/links/ip_timeout.html",
		map[string]interface{}{"EnableCache": true, "RetryCachedErrors": false, "LogLevel": issues.LevelDebug})
	tExpectIssueCount(t, hT2, 1)

	// Verify it used the cache (should see "from cache" not "hitting")
	if hT2.issueStore.MessageMatchCount("from cache") == 0 {
		t.Error("expected cached timeout to be reused (should see 'from cache' message)")
	}
	if hT2.issueStore.MessageMatchCount("hitting") > 0 {
		t.Error("should not retry when timeout is cached and RetryCachedErrors is false")
	}
}

// TestTimeoutCachedMessage : Test that a URL that previously timed out will be
// reported as "(cached)" on subsequent runs. when RetryCachedErrors is false.
func TestTimeoutCachedMessage(t *testing.T) {
	tSkipShortExternal(t)

	// First run: cause and cache a timeout
	hT := tTestFileOpts("fixtures/links/ip_timeout.html",
		map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "RetryCachedErrors": false})
	tExpectIssueCount(t, hT, 1)

	// Second run: with RetryCachedErrors disabled, should use cached timeout and report with "(cached)" message
	hT2 := tTestFileOpts("fixtures/links/ip_timeout.html",
		map[string]interface{}{"EnableCache": true, "RetryCachedErrors": false})
	tExpectIssueCount(t, hT2, 1)
	tExpectIssue(t, hT2, "request exceeded our ExternalTimeout (cached)", 1)
}
