package htmltest

import (
	"encoding/json"
	"flag"
	"os"
	"path"
	"testing"

	"github.com/imdario/mergo"
	"github.com/stretchr/testify/assert"
)

var skipSlow = flag.Bool("skip-slow", false, "skip slow tests (timeouts, etc.)")

// Test fixture helpers

// tTestFileOptsFromCleanOutputDir runs a test after having removed the default
// output directory. This is ensures a clean fixture for tests that might be
// influenced by the output directory content, such as the refcache file.
func tTestFileOptsFromCleanOutputDir(filename string, tOpts map[string]interface{}) *HTMLTest {
	tRemoveOutputDir(tOpts)
	return tTestFileOpts(filename, tOpts)
}

func tRemoveOutputDir(tOpts map[string]interface{}) {
	opts := DefaultOptions()
	_ = mergo.MergeWithOverwrite(&opts, tOpts)
	_ = os.RemoveAll(opts["OutputDir"].(string))
}

// Test skip helpers

// tSkipSlow skips slow tests that involve actual timeouts (seconds of waiting).
// Controlled by the -skip-slow flag. Use -skip-slow=true to skip these tests
// during rapid development.
func tSkipSlow(t *testing.T) {
	if *skipSlow {
		t.Skip("Skipping slow test (involves actual timeout waits)")
	}
}

// Cache assertion helpers

// tExpectCached asserts that a URL is present in the refcache.
// If statusCode is provided, also asserts the cached status code matches.
func tExpectCached(t *testing.T, hT *HTMLTest, url string, statusCode ...int) {
	cR, ok := hT.refCache.Get(url)
	if !assert.True(t, ok, "URL should be cached: "+url) {
		return // Stop if URL not in cache to avoid nil pointer panic
	}
	if len(statusCode) > 0 {
		assert.Equal(t, statusCode[0], cR.StatusCode, "cached status code")
	}
}

// tExpectNotCached asserts that a URL is NOT present in the refcache.
func tExpectNotCached(t *testing.T, hT *HTMLTest, url string) {
	_, ok := hT.refCache.Get(url)
	assert.False(t, ok, "URL should not be cached: "+url)
}

// tExpectCacheEmpty asserts that the refcache is empty.
func tExpectCacheEmpty(t *testing.T, hT *HTMLTest) {
	cachePath := path.Join(hT.opts.OutputDir, hT.opts.OutputCacheFile)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return // Cache file doesn't exist - that's empty!
		}
		t.Fatalf("Error reading cache file: %v", err)
	}

	// Parse JSON to check if empty
	var cache map[string]interface{}
	if err := json.Unmarshal(data, &cache); err != nil {
		t.Fatalf("Cache file is not valid JSON: %v", err)
	}

	assert.Equal(t, 0, len(cache), "cache should be empty")
}
