package syntax_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/adapter/projection"
	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^a model file whose third line declares an action, which the subset does not support$`, w.aModelWithAnAction)
	sc.When(`^the adapter starts on it$`, w.theAdapterStarts)
	sc.Then(`^it refuses to start$`, w.itRefuses)
	sc.Then(`^the refusal names the file, line (\d+) and column (\d+)$`, w.theRefusalNamesThePlace)
	sc.Given(`^the parser's rejection fixtures, one for each rejection it distinguishes$`, w.theRejectionFixtures)
	sc.When(`^each fixture is read$`, w.eachFixtureIsRead)
	sc.Then(`^each is refused at the line and column its fixture gives, with the message it gives$`, w.eachIsRefusedAtItsPlace)
}

// world is what one scenario of this package knows.
type world struct {
	t        *testing.T
	path     string
	err      error
	fixtures []tabletest.Case[string, rejection]
	got      []rejection
}

func (w *world) aModelWithAnAction() {
	w.path = filepath.Join(w.t.TempDir(), "unsupported.sysml")
	assert.NoError(w.t, os.WriteFile(w.path, []byte("package P {\n  part def A;\n  action a;\n}\n"), 0o600))
}

// theAdapterStarts loads the file as the adapter subcommand does before it
// serves anything.
func (w *world) theAdapterStarts() { _, w.err = projection.Load(w.path) }

func (w *world) itRefuses() { assert.Error(w.t, w.err) }

func (w *world) theRefusalNamesThePlace(line, column int) {
	e := assert.ErrorAs[*syntax.Error](w.t, w.err)
	assert.Equal(w.t, e.File, w.path)
	assert.Equal(w.t, e.Line, line)
	assert.Equal(w.t, e.Column, column)
}

func (w *world) theRejectionFixtures() {
	w.fixtures = append(append(w.fixtures, structuralRejections...), expressionAndPortRejections...)
}

func (w *world) eachFixtureIsRead() {
	for _, fixture := range w.fixtures {
		_, err := syntax.Parse("m.sysml", fixture.In)
		e := assert.ErrorAs[*syntax.Error](w.t, err)
		w.got = append(w.got, rejection{e.Line, e.Column, e.Message})
	}
}

func (w *world) eachIsRefusedAtItsPlace() {
	assert.Len(w.t, w.got, len(w.fixtures))
	for i, fixture := range w.fixtures {
		assert.Equal(w.t, w.got[i], fixture.Want)
	}
}
