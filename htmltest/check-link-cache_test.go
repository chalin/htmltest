package htmltest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for cache-related functionality
// Feature: RetryCachedErrors
// Added by @chalin

func TestCacheOptions(t *testing.T) {
	opts := DefaultOptions()

	// Sanity check: ensure that non-existent option doesn't exist
	_, exists := opts["NonExistentOption"]
	assert.False(t, exists, "NonExistentOption should not exist")

	// Verify cache-related option defaults
	assert.Equal(t, false, opts["CacheAllExternal"], "CacheAllExternal default")
	assert.Equal(t, true, opts["RetryCachedErrors"], "RetryCachedErrors default")
}

// ========================================
// External Error Caching Tests
// ========================================

// TestExternalErrorCached : Test that URLs with HTTP error status codes (such
// as 404) are saved to the refcache.
func TestExternalErrorCached(t *testing.T) {
	hT := tTestFileOptsFromCleanOutputDir("fixtures/images/imageExternal404.html",
		map[string]interface{}{"VCREnable": true, "EnableCache": true})
	tExpectIssueCount(t, hT, 1)
	tExpectCached(t, hT, "https://upload.wikimedia.org/wikipedia/en/404", 404)
}

// TestExternalErrorCachedRetried : Test that by default cached error responses
// (404, etc.) are retried on subsequent runs.
func TestExternalErrorCachedRetried(t *testing.T) {
	fixture := "fixtures/images/imageExternal404.html"
	opts := map[string]interface{}{"VCREnable": true, "EnableCache": true}

	// First run: populate cache with 404
	hT := tTestFileOptsFromCleanOutputDir(fixture, opts)
	tExpectIssueCount(t, hT, 1)

	// Second run: WITH VCR - should retry even though 404 is cached (default behavior)
	hT2 := tTestFileOpts(fixture, opts)
	tExpectIssueCount(t, hT2, 1)

	// Verify it did NOT use the cache (should see "fresh" not "from cache")
	tExpectIssue(t, hT2, "from cache", 0)
	tExpectIssue(t, hT2, "fresh", 1)
}

// TestExternalBrokenRetryCachedErrorsDisabled : Test that URLs with non-OK
// status are not retried when RetryCachedErrors is false.
func TestExternalBrokenRetryCachedErrorsDisabled(t *testing.T) {
	fixture := "fixtures/images/imageExternal404.html"
	opts := map[string]interface{}{"VCREnable": true, "EnableCache": true, "RetryCachedErrors": false}

	// First run WITH VCR: populate cache with 404
	hT := tTestFileOptsFromCleanOutputDir(fixture, opts)
	tExpectIssueCount(t, hT, 1)
	tExpectCached(t, hT, "https://upload.wikimedia.org/wikipedia/en/404", 404)

	// Second run: should use cached 404 (not retry because RetryCachedErrors is false)
	hT2 := tTestFileOpts(fixture, opts)
	tExpectIssueCount(t, hT2, 1)

	// Verify it used the cache (not retried)
	tExpectIssue(t, hT2, "from cache", 1)
	tExpectIssue(t, hT2, "hitting", 0)
}

// ========================================
// Timeout Caching Tests
// ========================================

// TestTimeoutNotCachedByDefault : Test that by default, timeouts are NOT cached
// and are retried on every run. This ensures backward compatibility with the
// original behavior.
func TestTimeoutNotCachedByDefault(t *testing.T) {
	tSkipSlow(t)
	tSkipShortExternal(t)
	fixture := "fixtures/links/ip_timeout.html"

	// First run: timeout occurs
	opts := map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true}
	hT := tTestFileOptsFromCleanOutputDir(fixture, opts)
	tExpectIssueCount(t, hT, 1)
	tExpectIssue(t, hT, "request exceeded our ExternalTimeout", 1)

	// Second run: should retry (not use cache) because RetryCachedErrors is true (default)
	hT2 := tTestFileOpts(fixture, opts)
	tExpectIssueCount(t, hT2, 1)

	// Verify it retried (should see "fresh" and "hitting", not "from cache")
	tExpectIssue(t, hT2, "fresh", 1)
	tExpectIssue(t, hT2, "hitting", 1)
	tExpectIssue(t, hT2, "from cache", 0)
}

// TestTimeoutNotCachedWithRetryCacheErrorsOnly : Test that timeouts are NOT
// cached when ONLY RetryCachedErrors is false (without CacheAllExternal). This
// ensures that the legacy behavior of RetryCachedErrors is no longer active.
func TestTimeoutNotCachedWithRetryCacheErrorsOnly(t *testing.T) {
	tSkipSlow(t)
	tSkipShortExternal(t)
	opts := map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "RetryCachedErrors": false}

	hT := tTestFileOptsFromCleanOutputDir("fixtures/links/ip_timeout.html", opts)
	tExpectIssueCount(t, hT, 1)
	tExpectIssue(t, hT, "request exceeded our ExternalTimeout", 1)
	tExpectCacheEmpty(t, hT)
}

// TestTimeoutIsCached : Test that URLs that timeout are saved to the refcache
// when CacheAllExternal is true.
func TestTimeoutIsCached(t *testing.T) {
	tSkipSlow(t)
	tSkipShortExternal(t)
	opts := map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "CacheAllExternal": true}
	hT := tTestFileOptsFromCleanOutputDir("fixtures/links/ip_timeout.html", opts)
	tExpectIssueCount(t, hT, 1)
	tExpectIssue(t, hT, "request exceeded our ExternalTimeout", 1)
	tExpectCached(t, hT, "http://5.6.7.8", StatusTimeout)
}

// TestTimeoutCachedButRetried : Test that timeouts are cached when CacheAllExternal
// is true, but still retried on subsequent runs when RetryCachedErrors is true (default).
// This verifies Row 3 of the behavior matrix.
func TestTimeoutCachedButRetried(t *testing.T) {
	tSkipSlow(t)
	tSkipShortExternal(t)
	fixture := "fixtures/links/ip_timeout.html"
	opts := map[string]interface{}{"ExternalTimeout": 1, "EnableCache": true, "CacheAllExternal": true, "RetryCachedErrors": true}

	// First run: timeout occurs and is cached
	hT := tTestFileOptsFromCleanOutputDir(fixture, opts)
	tExpectIssueCount(t, hT, 1)
	tExpectCached(t, hT, "http://5.6.7.8", StatusTimeout)

	// Second run: should retry despite being cached (RetryCachedErrors: true)
	hT2 := tTestFileOpts(fixture, opts)
	tExpectIssueCount(t, hT2, 1)

	// Verify it was retried (not reused from cache)
	tExpectIssue(t, hT2, "fresh", 1)
	tExpectIssue(t, hT2, "hitting", 1)
	tExpectIssue(t, hT2, "from cache", 0)
}

// TestTimeoutCachedReused : Test that cached timeout results are reused on
// subsequent runs without retrying the URL when CacheAllExternal is true and
// RetryCachedErrors is false.
func TestTimeoutCachedReused(t *testing.T) {
	tSkipSlow(t)
	tSkipShortExternal(t)
	fixture := "fixtures/links/ip_timeout.html"
	opts := map[string]interface{}{"EnableCache": true, "CacheAllExternal": true, "RetryCachedErrors": false}

	// First run: cause and cache a timeout
	opts["ExternalTimeout"] = 1
	hT := tTestFileOptsFromCleanOutputDir(fixture, opts)
	tExpectIssueCount(t, hT, 1)

	// Second run: WITHOUT timeout set - should use cached timeout, not actually try the request
	delete(opts, "ExternalTimeout")
	hT2 := tTestFileOpts(fixture, opts)
	tExpectIssueCount(t, hT2, 1)

	// Verify it used the cache (should see "from cache" not "hitting")
	tExpectIssue(t, hT2, "from cache", 1)
	tExpectIssue(t, hT2, "hitting", 0)
	tExpectIssue(t, hT2, "request exceeded our ExternalTimeout (cached)", 1)
}

// ========================================
// CacheAllExternal Tests
// ========================================

// TestCacheAllExternalDisabled : Test that external links are NOT cached when
// CacheAllExternal is false (default behavior).
// This is a regression test to ensure the default behavior doesn't change.
func TestCacheAllExternalDisabled(t *testing.T) {
	// Run with CheckExternal disabled and CacheAllExternal disabled (defaults)
	hT := tTestFileOptsFromCleanOutputDir("fixtures/links/brokenLinkExternalSingle.html",
		map[string]interface{}{"CheckExternal": false, "EnableCache": true})
	tExpectIssueCount(t, hT, 0) // No errors since external checking is disabled
	tExpectCacheEmpty(t, hT)
}

// TestCacheAllExternalDiscovery : Test that external links ARE cached with
// StatusUnchecked when CacheAllExternal is true and CheckExternal is false.
// This is the core discovery mode feature.
func TestCacheAllExternalDiscovery(t *testing.T) {
	fixture := "fixtures/links/brokenLinkExternalSingle.html"
	opts := map[string]interface{}{"CheckExternal": false, "EnableCache": true, "CacheAllExternal": true}

	hT := tTestFileOptsFromCleanOutputDir(fixture, opts)
	tExpectIssueCount(t, hT, 0) // No errors since external checking is disabled

	// Verify the external link WAS cached with StatusUnchecked
	tExpectCached(t, hT, "http://www.asdo3IRJ395295jsingrkrg4.com", StatusUnchecked)
}

// TestCacheAllExternalAndIgnoredURLs : Test that ignored URLs are NOT cached in
// discovery mode, verifying CacheAllExternal respects IgnoreURLs patterns.
func TestCacheAllExternalAndIgnoredURLs(t *testing.T) {
	fixture := "fixtures/links/brokenLinkExternalSingle.html"
	opts := map[string]interface{}{
		"CheckExternal":    false,
		"EnableCache":      true,
		"CacheAllExternal": true,
		"IgnoreURLs":       []interface{}{"asd.*\\.com"},
	}

	hT := tTestFileOptsFromCleanOutputDir(fixture, opts)
	tExpectIssueCount(t, hT, 0)
	tExpectCacheEmpty(t, hT)
}

// TestCacheAllExternalQueryString : Test that query strings are stripped before
// caching in discovery mode, verifying CacheAllExternal works with StripQueryString.
func TestCacheAllExternalQueryString(t *testing.T) {
	fixture := "fixtures/links/check_just_once.html"
	opts := map[string]interface{}{
		"CheckExternal":    false,
		"EnableCache":      true,
		"CacheAllExternal": true,
		"StripQueryString": true,
	}

	hT := tTestFileOptsFromCleanOutputDir(fixture, opts)
	tExpectIssueCount(t, hT, 0)

	// Verify the URL was cached WITHOUT the query string
	tExpectCached(t, hT, "https://github.com/contact", StatusUnchecked)

	// Verify the URL WITH query string is NOT in cache (proves stripping happened)
	tExpectNotCached(t, hT, "https://github.com/contact?form%5Bsubject%5D=New+Assigned+Events")
}
