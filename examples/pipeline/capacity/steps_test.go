package capacity

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the capacity service, asked for the worked example with parse at 1700$`, w.askedOnce)
	sc.When(`^the service is started again and asked the same$`, w.askedAgain)
	sc.Then(`^the two answers are the same, byte for byte$`, w.theAnswersAreTheSame)
	sc.Given(`^the capacity service's hand-written source, the generated files left out$`, func() {})
	sc.When(`^it is read$`, w.theSourceIsRead)
	sc.Then(`^its resolver holds the configured names and nothing else$`, w.theResolverHoldsOnlyTheNames)
	sc.Then(`^no hand-written file of the service declares a variable at package level$`, w.noPackageVariable)
	sc.Then(`^the flow package declares only its errors at package level$`, w.flowDeclaresOnlyErrors)
}

// world is what one scenario of this package knows.
type world struct {
	t       *testing.T
	answers [][]byte
	service []*ast.File
	flow    []*ast.File
}

func (w *world) ask() {
	vars := `{"reps":[` + pipelineRep(2000, 1700, 700, 700, 1800) + `]}`
	w.answers = append(w.answers, post(w.t, Handler(names), entities, vars))
}

func (w *world) askedOnce()  { w.ask() }
func (w *world) askedAgain() { w.ask() }

func (w *world) theAnswersAreTheSame() {
	assert.Len(w.t, w.answers, 2)
	assert.Equal(w.t, string(w.answers[1]), string(w.answers[0]))
	assert.True(w.t, bytes.Contains(w.answers[0], []byte(`"capacity":1400`)), "the answer carries the capacity")
}

// handWritten parses the Go files of dir that are neither tests nor
// generated.
func handWritten(t *testing.T, dir string) []*ast.File {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	assert.NoError(t, err)
	var files []*ast.File
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		src := assert.Must(t, raw, err)
		if strings.HasSuffix(path, "_test.go") || bytes.HasPrefix(src, []byte("// Code generated")) {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, src, 0)
		files = append(files, assert.Must(t, parsed, err))
	}
	return files
}

func (w *world) theSourceIsRead() {
	w.service, w.flow = handWritten(w.t, "."), handWritten(w.t, "flow")
	assert.True(w.t, len(w.service) >= 4, "at least four hand-written service files were read")
}

func (w *world) theResolverHoldsOnlyTheNames() {
	found := false
	for _, file := range w.service {
		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "Resolver" {
				return true
			}
			found = true
			fields := spec.Type.(*ast.StructType).Fields.List
			assert.Len(w.t, fields, 1)
			selector, ok := fields[0].Type.(*ast.SelectorExpr)
			assert.True(w.t, ok && selector.Sel.Name == "Names", "the resolver's one field is the configured names")
			return false
		})
	}
	assert.True(w.t, found, "the service declares its resolver")
}

// packageVariables returns every value a file declares at package level.
func packageVariables(file *ast.File) []*ast.ValueSpec {
	var specs []*ast.ValueSpec
	for _, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				specs = append(specs, spec.(*ast.ValueSpec))
			}
		}
	}
	return specs
}

func (w *world) noPackageVariable() {
	for _, file := range w.service {
		for _, spec := range packageVariables(file) {
			w.t.Errorf("the service declares %s at package level", spec.Names[0].Name)
		}
	}
}

func (w *world) flowDeclaresOnlyErrors() {
	for _, file := range w.flow {
		for _, spec := range packageVariables(file) {
			for i, value := range spec.Values {
				if !isErrorsNew(value) {
					w.t.Errorf("flow declares %s at package level, and it is no error", spec.Names[i].Name)
				}
			}
		}
	}
}

// isErrorsNew reports whether a value is a call of errors.New.
func isErrorsNew(value ast.Expr) bool {
	call, ok := value.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "New"
}
