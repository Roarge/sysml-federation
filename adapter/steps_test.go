package adapter_test

import (
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the adapter's Go and schema files, its tests left out$`, func() {})
	sc.When(`^they are searched for the example's names$`, w.theSourceIsSearched)
	sc.Then(`^none is found$`, w.noneIsFound)
	sc.Then(`^at least twenty-two files were searched$`, w.atLeastTwentyTwoFiles)
	sc.Given(`^the second fixture of SR-16$`, func() {})
	sc.When(`^it is projected$`, w.theSecondFixtureIsProjected)
	sc.Then(`^it projects five parts, four requirements and one verification case$`, w.itProjectsEverything)
	sc.Then(`^its text carries no name of the example but capacity, which both models declare$`, w.noNameOfTheExample)
}

// world is what one scenario of this package knows.
type world struct {
	t       *testing.T
	checked int
	hits    []string
	m       *model.Model
}

func (w *world) theSourceIsSearched() {
	checked, hits, err := exampleNameHits()
	assert.NoError(w.t, err)
	w.checked, w.hits = checked, hits
}

func (w *world) noneIsFound() {
	for _, hit := range w.hits {
		w.t.Error(hit)
	}
}

func (w *world) atLeastTwentyTwoFiles() {
	assert.True(w.t, w.checked >= 22, "at least twenty-two adapter source files were searched")
}

func (w *world) theSecondFixtureIsProjected() {
	loaded, err := model.Load("model/testdata/warehouse.sysml")
	w.m = assert.Must(w.t, loaded, err)
}

func (w *world) itProjectsEverything() {
	assert.Len(w.t, w.m.Parts, 5)
	assert.Len(w.t, w.m.Requirements, 4)
	assert.Len(w.t, w.m.VerificationCases, 1)
}

func (w *world) noNameOfTheExample() {
	read, err := os.ReadFile("model/testdata/warehouse.sysml")
	src := string(assert.Must(w.t, read, err))
	for _, name := range exampleNames {
		if name != "capacity" {
			assert.True(w.t, !strings.Contains(src, name), "the warehouse fixture is free of "+name)
		}
	}
}
