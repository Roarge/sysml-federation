package adapter_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// exampleNames are the identifiers examples/pipeline/model.sysml introduces
// that no adapter source may contain (SR-17). Each is a case-sensitive
// substring. The example's other names are absent from this list on purpose:
// parse, serve, Server, Query, target and the library names are ordinary
// words of the parser, of the standard library or of the SysML libraries, and
// the elements that carry them are caught through the names below.
var exampleNames = []string{
	"'PIPE'", "PIPE-",
	"Pipeline", "pipeline",
	"Component", "capacity",
	"QueryInput", "QueryOutput",
	"throughput", "Throughput",
	"latency", "Latency",
	"requiredRate",
	"ingest", "indexA", "indexB",
}

// sourceExtensions are the adapter's source kinds: Go, hand-written or
// generated, and the subgraph schema.
var sourceExtensions = map[string]bool{".go": true, ".graphql": true, ".graphqls": true}

func TestSR17_NoExampleIdentifiersInTheAdapter(t *testing.T) {
	checked := 0
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") || !sourceExtensions[filepath.Ext(path)] {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		checked++
		for i, line := range strings.Split(string(src), "\n") {
			for _, name := range exampleNames {
				if strings.Contains(line, name) {
					t.Errorf("%s:%d: contains the example identifier %q", path, i+1, name)
				}
			}
		}
		return nil
	})
	assert.NoError(t, err)
	// The adapter's twenty-two source files are the floor. A walk that finds
	// fewer is looking in the wrong place, or has skipped a whole package,
	// and would pass for the wrong reason.
	assert.True(t, checked >= 22, "at least twenty-two adapter source files were checked")
}

// The second fixture is SR-17's other verification: a model that shares no
// vocabulary with the example projects through the same code paths.
// capacity is the one name both models declare, because the fixture keeps
// the attribute the analysis service reads so its tests can reuse the file.
func TestSR17_SecondFixtureSharesNoNamesWithTheExample(t *testing.T) {
	read, err := os.ReadFile("model/testdata/warehouse.sysml")
	src := string(assert.Must(t, read, err))
	for _, name := range exampleNames {
		if name == "capacity" {
			continue
		}
		assert.True(t, !strings.Contains(src, name), "the warehouse fixture is free of "+name)
	}
}
