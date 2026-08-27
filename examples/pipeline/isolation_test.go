package pipeline

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/client"

	"github.com/Roarge/sysml-federation/examples/pipeline/document"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// mutationOut absorbs a mutation's answer under whatever field name it
// carries. The test client refuses a key its target has no place for, so a
// mutation cannot be posted into an empty struct.
type mutationOut map[string]struct {
	Version int `json:"version"`
}

// TestSR36_DocumentOperationsLeaveTheModelVersion runs every editorial
// operation against the document service and reads back the version the
// service reports, which advances once per accepted operation.
func TestSR36_DocumentOperationsLeaveTheModelVersion(t *testing.T) {
	created, err := document.New()
	svc := assert.Must(t, created, err)
	c := client.New(document.Handler(svc), client.Path("/graphql"))
	operations := []string{
		`mutation { moveNode(id: "PIPE-R1.5", parentId: "PIPE-R1", index: 0) { version } }`,
		`mutation { moveNode(id: "PIPE-R2", parentId: "PIPE-R1", index: 5) { version } }`,
		`mutation { insertHeading(aboveId: "PIPE-R1", text: "Performance") { version } }`,
		`mutation { addProse(parentId: null, index: 0, text: "Why allocated limits fail.") { version } }`,
		`mutation { editText(id: "intro", text: "Rewritten.") { version } }`,
		`mutation { excludeRequirement(requirementId: "PIPE-R1.4") { version } }`,
		`mutation { includeRequirement(requirementId: "PIPE-R1.4") { version } }`,
		`mutation { resetDocument { version } }`,
	}
	for i, op := range operations {
		var out mutationOut
		c.MustPost(op, &out)
		assert.Equal(t, svc.Version(), i+2)
	}
}

// TestSR36_TheDocumentServiceReachesNoAdapterPackage parses the imports of
// every non-test Go source of the document service, its tree sub-package
// included, and fails on any import that leads into the adapter. The model
// version cannot move under a document operation because the state holding
// it is not reachable from here, not because nothing happened to write to
// it, and this test fails the moment the two are wired together.
func TestSR36_TheDocumentServiceReachesNoAdapterPackage(t *testing.T) {
	fset := token.NewFileSet()
	parsed := 0
	err := filepath.WalkDir("document", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		parsed++
		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if strings.Contains(imported, "/adapter/") || strings.HasSuffix(imported, "/adapter") {
				t.Errorf("%s: imports %q", path, imported)
			}
		}
		return nil
	})
	assert.NoError(t, err)
	// The document service's eight non-test sources are the floor. A walk that
	// parses fewer is looking in the wrong place, or has skipped the tree
	// sub-package, and would report a clean result it never earned.
	assert.True(t, parsed >= 8, "at least eight document service sources were parsed")
}
