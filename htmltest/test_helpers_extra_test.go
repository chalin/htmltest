package htmltest

import (
	"os"
	"testing"

	"github.com/imdario/mergo"
	"github.com/stretchr/testify/assert"
)

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
	mergo.MergeWithOverwrite(&opts, tOpts)
	os.RemoveAll(opts["OutputDir"].(string))
}

// Cache assertion helpers

// tExpectCached asserts that a URL is present in the refcache.
// If statusCode is provided, also asserts the cached status code matches.
func tExpectCached(t *testing.T, hT *HTMLTest, url string, statusCode ...int) {
	cR, ok := hT.refCache.Get(url)
	assert.True(t, ok, "URL should be cached: "+url)
	if len(statusCode) > 0 {
		assert.Equal(t, statusCode[0], cR.StatusCode, "cached status code")
	}
}

// tExpectNotCached asserts that a URL is NOT present in the refcache.
func tExpectNotCached(t *testing.T, hT *HTMLTest, url string) {
	_, ok := hT.refCache.Get(url)
	assert.False(t, ok, "URL should not be cached: "+url)
}
