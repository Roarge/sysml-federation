package scenario

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/internal/assert"
)

const fixture = "testdata/features/sr90-a-fixture-story.feature"

// fixtureSteps defines the fixture's steps, with a world of its own for each
// scenario, as a package's steps would.
func fixtureSteps(sc *godog.ScenarioContext) {
	var noted []string
	sc.Given(`^a defined step$`, func() {})
	sc.Given(`^the number (\d+) is noted$`, func(n string) { noted = append(noted, n) })
	sc.Then(`^one number has been noted$`, func() error {
		if len(noted) != 1 {
			return errors.New("the world was carried over from another example: " + strings.Join(noted, ", "))
		}
		return nil
	})
}

func TestReadFindsEachScenariosCriterionAndItsEvidence(t *testing.T) {
	features, err := Read("testdata/features")
	assert.NoError(t, err)
	assert.Len(t, features, 1)
	feature := features[0]
	assert.SliceEqual(t, feature.Tags, []string{"SR-90"})
	assert.Equal(t, feature.Story(), "SR-90")
	assert.Len(t, feature.Problems, 0)

	var criteria []string
	for _, s := range feature.Scenarios {
		criteria = append(criteria, s.Criterion)
	}
	assert.SliceEqual(t, criteria, []string{"passes", "undefined", "recordedElsewhere", "eachExample"})
	elsewhere := feature.Scenarios[2]
	assert.SliceEqual(t, elsewhere.Elsewhere, []string{"record"})
	assert.Equal(t, elsewhere.Runs(), false)
	assert.Equal(t, feature.Scenarios[0].Runs(), true)
	assert.Equal(t, feature.Scenarios[0].Line, 6)
}

func TestReadReportsWhatTheAgreementCannotRead(t *testing.T) {
	feature, err := ReadFile("testdata/problems/problems.feature")
	assert.NoError(t, err)
	assert.Equal(t, feature.Story(), "problems.feature")
	want := []string{
		"line 4: the scenario names no criterion",
		"line 8: the scenario names 2 criteria, one, two, and one scenario gives steps to one",
		"line 16: the examples carry tags, and a criterion is named on its scenario",
		"line 20: a Rule groups scenarios, and a scenario here sits directly under its feature",
	}
	for _, problem := range want {
		assert.Contains(t, feature.Problems, problem)
	}
	assert.Len(t, feature.Problems, len(want))
}

func TestReadReportsAFileWithNothingToRun(t *testing.T) {
	feature, err := ReadFile("testdata/empty/empty.feature")
	assert.NoError(t, err)
	assert.SliceEqual(t, feature.Problems, []string{"the file holds no scenario"})
}

func TestReadNamesAFileGodogCannotParse(t *testing.T) {
	_, err := ReadFile("testdata/broken/broken.feature")
	if err == nil || !strings.Contains(err.Error(), "broken.feature") {
		t.Fatalf("got %v, want an error naming the file", err)
	}
}

func TestSuiteRunsTheTaggedScenarioWithItsSteps(t *testing.T) {
	ran, status, report := Suite(fixture, "passes", fixtureSteps)
	assert.Equal(t, ran, 1)
	if status != 0 {
		t.Errorf("status %d, report:\n%s", status, report)
	}
}

func TestSuiteFailsAStepWithNoDefinition(t *testing.T) {
	ran, status, report := Suite(fixture, "undefined", fixtureSteps)
	assert.Equal(t, ran, 1)
	assert.NotEqual(t, status, 0)
	if !strings.Contains(report, "a step nobody defined") {
		t.Errorf("the report does not name the undefined step:\n%s", report)
	}
}

func TestSuiteRunsNothingForATagNoScenarioCarries(t *testing.T) {
	ran, status, _ := Suite(fixture, "noSuchCriterion", fixtureSteps)
	assert.Equal(t, ran, 0)
	assert.Equal(t, status, 0)
}

func TestSuiteGivesEachExampleAFreshWorld(t *testing.T) {
	ran, status, report := Suite(fixture, "eachExample", fixtureSteps)
	assert.Equal(t, ran, 2)
	if status != 0 {
		t.Errorf("status %d, report:\n%s", status, report)
	}
}

func TestRunRunsEachCriterionInASubtestOfItsStory(t *testing.T) {
	t.Chdir("testdata/run")
	var ran []string
	Run(t, func(t *testing.T, sc *godog.ScenarioContext) {
		ran = append(ran, t.Name())
		sc.Given(`^a defined step$`, func() {})
	})
	assert.SliceEqual(t, ran, []string{"TestRunRunsEachCriterionInASubtestOfItsStory/SR-92/runs"})
}

func TestElsewhereKindsAreNotGoTestsOrScenarios(t *testing.T) {
	kinds := ElsewhereKinds()
	for _, kind := range []string{"go-test", "scenario"} {
		if slices.Contains(kinds, kind) {
			t.Errorf("%s is a kind go test produces, and a scenario verified by it runs", kind)
		}
	}
	kinds[0] = "changed"
	assert.NotEqual(t, ElsewhereKinds()[0], "changed")
}
