package htmltest

import (
	"testing"

	"github.com/wjdp/htmltest/issues"
)

// Tests for cache-related functionality
// Feature: IgnoreCachedErrors
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

func TestAnchorExternalBrokenCachedRetried(t *testing.T) {
	// By default (IgnoreCachedErrors=false), cached 404s should be retried
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

func TestAnchorExternalBrokenIgnoreCachedErrors(t *testing.T) {
	// When IgnoreCachedErrors is enabled, cached 404s should be reused without retry
	// First run: populate cache with 404
	hT := tTestFileOpts("fixtures/images/imageExternal404.html",
		map[string]interface{}{"VCREnable": true, "EnableCache": true, "IgnoreCachedErrors": true})
	tExpectIssueCount(t, hT, 1)

	// Second run: WITHOUT VCR - should use cached 404, not retry (which would fail without VCR)
	hT2 := tTestFileOpts("fixtures/images/imageExternal404.html",
		map[string]interface{}{"EnableCache": true, "IgnoreCachedErrors": true, "LogLevel": issues.LevelDebug})
	tExpectIssueCount(t, hT2, 1)

	// Verify it used the cache by checking for "from cache" message
	if hT2.issueStore.MessageMatchCount("from cache") == 0 {
		t.Error("expected cached 404 to be reused (should see 'from cache' message)")
	}
}
