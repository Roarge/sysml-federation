// Package scenario reads and runs the Gherkin scenarios that give each
// acceptance criterion of a system story its executable form (AD-0034).
//
// A package whose tests run scenarios keeps them in a features directory beside
// its tests, one feature file per system story. The feature carries the story's
// key as its one tag, @SR-04, and each scenario carries the name of the
// criterion it gives steps to, as the model writes it, @fourPathsAnswer. A
// scenario for a criterion that go test cannot check also carries the kind of
// evidence that verifies it today, @record or @workflow, and is read and not
// run.
//
// The package's one test, TestScenarios, hands its steps to Run. Run makes a
// subtest per story and per criterion, TestScenarios/SR-04/fourPathsAnswer, and
// runs each criterion as a godog suite of its own inside that subtest. The steps
// are bound to the subtest's *testing.T, so a step can call the helpers the Go
// tests of its package already use. A helper that stops the test with t.Fatal
// ends the subtest there, and go test reports it failed. godog's own report of
// that scenario is lost, since the step never returned to it.
//
// Steps are written as regular expressions, and shared by convention rather
// than by a shared package, because each package drives its own part of the
// demo. The phrases the features share are "the demo is running", "the shipped
// model", "the second fixture", "the mutation is rejected" and "the model
// version is still".
package scenario

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v34"
)

// Dir is the directory beside a package's tests that holds its feature files.
const Dir = "features"

// elsewhere are the kinds of evidence a scenario may name in place of running,
// the kinds of the verification register that go test does not produce.
var elsewhere = []string{"analysis", "checkly", "inspection", "make-target", "record", "validator", "workflow"}

// ElsewhereKinds returns the kinds of evidence a scenario may name in place of
// running, in order.
func ElsewhereKinds() []string { return slices.Clone(elsewhere) }

// Feature is one feature file as the scenarios read it: its path as it was
// given, the tags of its feature without their @, its scenarios in the order
// they are written, and anything in it the agreement cannot read.
type Feature struct {
	Path      string
	Tags      []string
	Scenarios []Scenario
	Problems  []string
}

// Scenario is one scenario of a feature file: the line it starts on, its
// title, the criterion its tag names, and the kinds of evidence it names in
// place of running. A scenario naming no kind of evidence is one go test runs.
type Scenario struct {
	Line      int
	Title     string
	Criterion string
	Elsewhere []string
}

// Runs reports whether go test runs the scenario.
func (s Scenario) Runs() bool { return len(s.Elsewhere) == 0 }

// Story returns the key of the story the feature file tags, or the file's
// name where it carries no single tag, which the agreement reports.
func (f Feature) Story() string {
	if len(f.Tags) == 1 {
		return f.Tags[0]
	}
	return filepath.Base(f.Path)
}

// ReadFile reads one feature file with godog's own parser, the one the
// scenarios run with. A file godog cannot parse is an error naming it. A file
// it parses with nothing in it to run is returned with a problem, since godog
// passes over such a file in silence.
func ReadFile(path string) (Feature, error) {
	feature := Feature{Path: path}
	if _, err := os.Stat(path); err != nil {
		return feature, err
	}
	parsed, err := godog.TestSuite{Options: &godog.Options{Paths: []string{path}}}.RetrieveFeatures()
	if err != nil {
		return feature, err
	}
	if len(parsed) == 0 || parsed[0].Feature == nil {
		feature.Problems = append(feature.Problems, "the file holds no scenario")
		return feature, nil
	}
	document := parsed[0].Feature
	feature.Tags = tagNames(document.Tags)
	for _, child := range document.Children {
		switch {
		case child.Rule != nil:
			feature.Problems = append(feature.Problems,
				fmt.Sprintf("line %d: a Rule groups scenarios, and a scenario here sits directly under its feature", child.Rule.Location.Line))
		case child.Scenario != nil:
			feature.Scenarios = append(feature.Scenarios, readScenario(child.Scenario, &feature))
		}
	}
	return feature, nil
}

// readScenario reads one scenario's tags into the criterion it names and the
// kinds of evidence it names in place of running.
func readScenario(s *messages.Scenario, feature *Feature) Scenario {
	scenario := Scenario{Line: int(s.Location.Line), Title: s.Name}
	var criteria []string
	for _, tag := range tagNames(s.Tags) {
		if slices.Contains(elsewhere, tag) {
			scenario.Elsewhere = append(scenario.Elsewhere, tag)
		} else {
			criteria = append(criteria, tag)
		}
	}
	switch len(criteria) {
	case 1:
		scenario.Criterion = criteria[0]
	case 0:
		feature.Problems = append(feature.Problems,
			fmt.Sprintf("line %d: the scenario names no criterion", scenario.Line))
	default:
		feature.Problems = append(feature.Problems,
			fmt.Sprintf("line %d: the scenario names %d criteria, %s, and one scenario gives steps to one",
				scenario.Line, len(criteria), strings.Join(criteria, ", ")))
	}
	for _, examples := range s.Examples {
		if len(examples.Tags) > 0 {
			feature.Problems = append(feature.Problems,
				fmt.Sprintf("line %d: the examples carry tags, and a criterion is named on its scenario", examples.Location.Line))
		}
	}
	return scenario
}

// Read reads every feature file directly under dir, in order.
func Read(dir string) ([]Feature, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.feature"))
	if err != nil {
		return nil, err
	}
	slices.Sort(paths)
	features := make([]Feature, 0, len(paths))
	for _, path := range paths {
		feature, err := ReadFile(path)
		if err != nil {
			return nil, err
		}
		features = append(features, feature)
	}
	return features, nil
}

// Steps registers a package's steps for one scenario, bound to the subtest the
// scenario runs in.
type Steps func(t *testing.T, sc *godog.ScenarioContext)

// Run runs the scenarios of the package whose test calls it, from the features
// directory beside that test. Each story is a subtest and each criterion a
// subtest of it. A criterion verified elsewhere is skipped with the reason.
func Run(t *testing.T, steps Steps) {
	t.Helper()
	features, err := Read(Dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(features) == 0 {
		t.Fatalf("no feature file under %s", Dir)
	}
	for _, feature := range features {
		for _, problem := range feature.Problems {
			t.Errorf("%s: %s", feature.Path, problem)
		}
		t.Run(feature.Story(), func(t *testing.T) {
			for _, s := range feature.Scenarios {
				if s.Criterion == "" {
					continue
				}
				t.Run(s.Criterion, func(t *testing.T) {
					if !s.Runs() {
						t.Skipf("verified by the %s evidence of its verification case, not by go test",
							strings.Join(s.Elsewhere, " and "))
					}
					ran, status, report := Suite(feature.Path, s.Criterion, func(sc *godog.ScenarioContext) { steps(t, sc) })
					switch {
					case ran == 0:
						t.Errorf("godog ran no scenario tagged @%s in %s", s.Criterion, feature.Path)
					case status != 0:
						t.Errorf("the scenario did not pass:\n%s", report)
					}
				})
			}
		})
	}
}

// Suite runs the scenarios of one file that carry one tag, strictly, so that a
// step with no definition fails rather than passing as undefined. It returns
// how many scenarios godog started, its exit status and its report. A tag no
// scenario carries runs nothing and returns status 0, which is why the count
// is returned: a run of nothing is not a pass.
func Suite(path, tag string, initialise func(*godog.ScenarioContext)) (ran, status int, report string) {
	var out bytes.Buffer
	status = godog.TestSuite{
		Name: path,
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			ran++
			initialise(sc)
		},
		Options: &godog.Options{
			Format:      "pretty",
			Output:      &out,
			NoColors:    true,
			Paths:       []string{path},
			Tags:        "@" + tag,
			Strict:      true,
			Concurrency: 1,
		},
	}.Run()
	return ran, status, out.String()
}

// tagNames returns the names of a list of tags without their @.
func tagNames(tags []*messages.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, strings.TrimPrefix(tag.Name, "@"))
	}
	return names
}
