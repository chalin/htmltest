package htmltest

import (
	"os"

	"github.com/imdario/mergo"
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
