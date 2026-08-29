package pipeline

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
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

// TestEveryEditorialOperationIsAcceptedThroughTheGraph posts each of the
// eight editorial operations to the served handler and checks that the
// document's own version advances once per operation. It carries no
// requirement identifier because it looks no further than acceptance: what
// each operation does to the tree, the restore as last child and the
// renumbering are checked by the document package's own tests, and the
// change notice reaching every subscriber by the test that drives the
// subscription.
func TestEveryEditorialOperationIsAcceptedThroughTheGraph(t *testing.T) {
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

// service is one of the three that stand behind the router: a name to read
// in a failure, the directory holding its sources relative to this test,
// the import path of that directory, and the number of non-test sources a
// walk has to find before its silence means anything.
type service struct {
	name    string
	dir     string
	pkg     string
	sources int
}

// The three services. The adapter sits outside this example, so its
// directory is reached upwards while its import path is written out like the
// other two. The counts are floors rather than equalities: a service that
// gains a source keeps every walk below honest, and a service that loses a
// whole sub-package is what the floor is there to catch.
var (
	adapterService = service{
		name:    "the adapter",
		dir:     filepath.Join("..", "..", "adapter"),
		pkg:     "github.com/Roarge/sysml-federation/adapter",
		sources: 21,
	}
	capacityService = service{
		name:    "the capacity service",
		dir:     "capacity",
		pkg:     "github.com/Roarge/sysml-federation/examples/pipeline/capacity",
		sources: 9,
	}
	documentService = service{
		name:    "the document service",
		dir:     "document",
		pkg:     "github.com/Roarge/sysml-federation/examples/pipeline/document",
		sources: 8,
	}
	pipelineServices = []service{adapterService, capacityService, documentService}
)

// walkSources parses every non-test Go source under one service and hands
// each to visit, as the path it was read from and its syntax tree. It fails
// the test when the walk parses fewer files than the service is known to
// hold, since a walk that reached the wrong directory would otherwise
// report a clean result it never earned.
func walkSources(t *testing.T, s service, mode parser.Mode, visit func(path string, file *ast.File)) {
	t.Helper()
	fset := token.NewFileSet()
	parsed := 0
	err := filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, mode)
		if err != nil {
			return err
		}
		parsed++
		visit(path, file)
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, parsed >= s.sources,
		fmt.Sprintf("%s: %d sources parsed, at least %d expected", s.name, parsed, s.sources))
}

// importPaths returns the paths a parsed file imports, and fails the test on
// a path it cannot unquote rather than passing over it.
func importPaths(t *testing.T, path string, file *ast.File) []string {
	t.Helper()
	paths := make([]string, 0, len(file.Imports))
	for _, spec := range file.Imports {
		imported, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Errorf("%s: unquoting the import %s: %v", path, spec.Path.Value, err)
			continue
		}
		paths = append(paths, imported)
	}
	return paths
}

// owns reports whether the import path leads into the service's own tree,
// either to its top package or to a package under it.
func (s service) owns(imported string) bool {
	return imported == s.pkg || strings.HasPrefix(imported, s.pkg+"/")
}

// TestSR41_NoServiceImportsAnother walks the import graph of all three
// services and fails on an import that crosses out of one of them into
// another. Each service is read once and held against both of the others,
// so all six ordered pairs are covered and neither direction of a pair rests
// on the other. A call inside this process needs an import to name what it
// calls, so a graph with no edge between two services is also a graph in
// which neither can call the other except over the wire, and the test below
// covers that.
func TestSR41_NoServiceImportsAnother(t *testing.T) {
	for _, from := range pipelineServices {
		walkSources(t, from, parser.ImportsOnly, func(path string, file *ast.File) {
			for _, imported := range importPaths(t, path, file) {
				for _, to := range pipelineServices {
					if to.pkg == from.pkg || !to.owns(imported) {
						continue
					}
					t.Errorf("%s: %s imports %q, which belongs to %s",
						path, from.name, imported, to.name)
				}
			}
		})
	}
}

// addressLike matches a string that could name something to connect to: a
// URL, the loopback name, a dotted quad, or the bare colon and port a listen
// address is written with.
var addressLike = regexp.MustCompile(`(?i)https?://|localhost|[0-9]{1,3}(\.[0-9]{1,3}){3}|:[0-9]{2,5}(\D|$)`)

// reachesOut names, per package it is spelled with, the calls that would
// fetch an address from outside the source or open a connection to one. The
// qualifier is checked as well as the name, so a method of the same name on
// a value of the service's own is not mistaken for one of these.
var reachesOut = map[string][]string{
	"os":   {"Getenv", "LookupEnv", "Environ"},
	"net":  {"Dial", "DialTimeout", "Dialer"},
	"http": {"Get", "Head", "Post", "PostForm", "NewRequest", "NewRequestWithContext", "Client", "DefaultClient", "DefaultTransport", "Transport"},
}

// TestSR41_NoServiceIsConfiguredWithAnAddress reads the hand-written sources
// of all three services for a host, a port, a URL, a command-line flag, a
// lookup in the process environment and any means of opening an outbound
// connection, and finds none of them. A service that could reach another
// over the wire would need an address and a way to dial it, and neither is
// anywhere in these sources.
//
// What the walk cannot see is an address handed to a service by whoever
// starts it, since that would be a value at run time and not a word in the
// source. Nothing hands one over today: each service is constructed from the
// state it serves alone, the model store, the two quantity names and the
// document tree.
//
// Generated files are left out, because what they carry comes from the
// schemas and the generator rather than from anyone writing here, and one of
// them quotes the federation specification's own URL.
func TestSR41_NoServiceIsConfiguredWithAnAddress(t *testing.T) {
	for _, s := range pipelineServices {
		walkSources(t, s, parser.ParseComments, func(path string, file *ast.File) {
			if ast.IsGenerated(file) {
				return
			}
			for _, imported := range importPaths(t, path, file) {
				if imported == "flag" {
					t.Errorf("%s: %s imports %q, a way of being handed an address",
						path, s.name, imported)
				}
			}
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.BasicLit:
					if node.Kind != token.STRING {
						return true
					}
					text, err := strconv.Unquote(node.Value)
					if err != nil {
						t.Errorf("%s: unquoting the literal %s: %v", path, node.Value, err)
						return true
					}
					if addressLike.MatchString(text) {
						t.Errorf("%s: %s holds %q, which reads as an address",
							path, s.name, text)
					}
				case *ast.SelectorExpr:
					qualifier, ok := node.X.(*ast.Ident)
					if !ok {
						return true
					}
					for _, name := range reachesOut[qualifier.Name] {
						if node.Sel.Name == name {
							t.Errorf("%s: %s names %s.%s, which reaches outside the process",
								path, s.name, qualifier.Name, name)
						}
					}
				}
				return true
			})
		})
	}
}

// TestSR36_TheDocumentServiceReachesNoAdapterPackage parses the imports of
// every non-test Go source of the document service, its tree sub-package
// included, and fails on any import that leads into the adapter. The model
// version cannot move under a document operation because the state holding
// it is not reachable from here, not because nothing happened to write to
// it, and this test fails the moment the two are wired together. The sweep
// above holds the same pair among the other five, and this one names the
// pair the model version rests on.
func TestSR36_TheDocumentServiceReachesNoAdapterPackage(t *testing.T) {
	walkSources(t, documentService, parser.ImportsOnly, func(path string, file *ast.File) {
		for _, imported := range importPaths(t, path, file) {
			if adapterService.owns(imported) {
				t.Errorf("%s: imports %q", path, imported)
			}
		}
	})
}
