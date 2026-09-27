package trace

import (
	"regexp"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// check is one of the agreement test's checks, run against a root.
type check func(t *testing.T, root string)

// agreement names the check each scenario's Given sets up and each Then
// states. A Then passes only after the check it states has run, so a
// scenario cannot pass on the strength of another check.
type agreement struct {
	name   string
	run    check
	given  string
	claims []string
}

var agreements = []agreement{
	{"identifiers", identifiersAgree, "the model and the decisions folder", []string{
		"every identifier the model writes is declared exactly once in it",
		"every decision record the model names exists under the decisions folder, and every record there is named in the model"}},
	{"requirementsAffected", requirementsAffectedAgree, "the decision records", []string{
		"every one is a short name the model declares"}},
	{"goTests", goTestsAgree, "the Go tests of the requirement scheme and the verification register", []string{
		"every Go test the model names exists in the file it names, and every Go test of the scheme is named by the model"}},
	{"checkFiles", checkFilesAgree, "the check project and the case registers", []string{
		"every check file a case names exists in the check project, and every check file of that project is named by a case"}},
	{"images", imagesAgree, "the published images and the views", []string{
		"every published image is named by exactly one view, and every image a view names is published"}},
	{"coverage", coverageAgrees, "the stories, the cases and the logical architecture", []string{
		"every story that is done is verified by a case, satisfied by something in the logical architecture, and a derived end of a derivation connection"}},
	{"checkInventory", checkInventoryAgrees, "the check register and the manifest the check project reads", []string{
		"both carry the same checks with the same fields"}},
	{"sessionParts", sessionPartsAgree, "the compose file of the check session and the session composite", []string{
		"the services the compose file starts are the parts the composite carries, and nothing else"}},
	{"scenarios", scenariosAgree, "the feature files, the system stories and the verification register", []string{
		"every feature file is named after the one system story its tag names, and every scenario names one acceptance criterion of that story",
		"every criterion of a story that is done has exactly one scenario",
		"every kind of evidence a scenario names in place of running is offered by its story's case",
		"every scenario go test runs is named by that case and has a runner in its package"}},
}

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	for _, a := range agreements {
		sc.Given("^"+regexp.QuoteMeta(a.given)+"$", func() { w.chosen = a })
		for _, claim := range a.claims {
			sc.Then("^"+regexp.QuoteMeta(claim)+"$", func() { w.claims(a.name) })
		}
	}
	sc.When(`^the agreement test (?:reads|compares) them$`, w.theCheckRuns)
	sc.When(`^the agreement test reads the requirements each names as affected$`, w.theCheckRuns)
}

// world is what one scenario of this package knows.
type world struct {
	t      *testing.T
	chosen agreement
	ran    string
}

func (w *world) theCheckRuns() {
	found, err := ModuleRoot()
	root := assert.Must(w.t, found, err)
	w.chosen.run(w.t, root)
	w.ran = w.chosen.name
}

func (w *world) claims(name string) { assert.Equal(w.t, w.ran, name) }
