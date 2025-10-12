package htmltest

import (
	"net/http"
	"testing"

	"github.com/wjdp/htmltest/issues"
)

// Tests for cache-related functionality
// Feature: RetryCachedErrors
// Added by @chalin

func TestAnchorExternalBrokenCached(t *testing.T) {
	// URLs returning HTTP error status codes (404) should be saved to cache
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

// By default, cached non-ok status codes are retried
func TestAnchorExternalNonOkCachedRetried(t *testing.T) {

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

func TestAnchorExternalBrokenRetryCachedErrorsDisabled(t *testing.T) {
	// When RetryCachedErrors is false, cached 404s should be reused without retry
	// First run: populate cache with 404
	hT := tTestFileOpts("fixtures/images/imageExternal404.html",
		map[string]interface{}{"VCREnable": true, "EnableCache": true, "RetryCachedErrors": false})
	tExpectIssueCount(t, hT, 1)

	// Second run: WITHOUT VCR - should use cached 404, not retry (which would fail without VCR)
	hT2 := tTestFileOpts("fixtures/images/imageExternal404.html",
		map[string]interface{}{"EnableCache": true, "RetryCachedErrors": false, "LogLevel": issues.LevelDebug})
	tExpectIssueCount(t, hT2, 1)

	// Verify it used the cache by checking for "from cache" message
	if hT2.issueStore.MessageMatchCount("from cache") == 0 {
		t.Error("expected cached 404 to be reused (should see 'from cache' message)")
	}
}

// ========================================
// Timeout Caching Tests
// ========================================

func TestTimeoutNotCachedByDefault(t *testing.T) {
	// By default timeouts are NOT cached and are retried every run (backward compatibility)
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

func TestTimeoutIsCached(t *testing.T) {
	// URLs that timeout should be saved to cache when RetryCachedErrors is false
	tSkipShortExternal(t)
	hT := tTestFileOpts("fixtures/links/ip_timeout.html",
		map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "RetryCachedErrors": false})
	tExpectIssueCount(t, hT, 1)
	tExpectIssue(t, hT, "request exceeded our ExternalTimeout", 1)

	// Verify the timeout was saved to cache with http.StatusRequestTimeout (408)
	cR, ok := hT.refCache.Get("http://5.6.7.8")
	if !ok {
		t.Error("expected timeout to be cached when RetryCachedErrors is false, but it wasn't")
	}
	if cR.StatusCode != http.StatusRequestTimeout {
		t.Errorf("expected status code %d (http.StatusRequestTimeout) for timeout, got %d", http.StatusRequestTimeout, cR.StatusCode)
	}
}

func TestTimeoutCachedReused(t *testing.T) {
	// When RetryCachedErrors is false, cached timeouts should be reused
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

func TestTimeoutCachedMessage(t *testing.T) {
	// Cached timeouts should have a specific error message when reused
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
