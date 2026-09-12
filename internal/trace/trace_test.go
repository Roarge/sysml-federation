package trace

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// TestSR46_ModelAndRepositoryAgree is the evidence for SR-46. The model under
// model/ names things that live in the repository around it: the identifiers of
// the light scheme, the decision records, the Go test functions, the check
// files, the published images, the inventory of live checks and the services of
// the check session. Nothing in the modelling language makes those names true.
// Each subtest below reads one register of the model and the part of the
// repository it describes, and fails on the first name one side carries and the
// other lacks.
//
// Every register these subtests read is written, so each of them asserts. The
// last one is the exception: the check session's compose file is not there yet,
// so it skips and names the file it is waiting for.
func TestSR46_ModelAndRepositoryAgree(t *testing.T) {
	found, err := ModuleRoot()
	root := assert.Must(t, found, err)

	t.Run("identifiers", func(t *testing.T) { identifiersAgree(t, root) })
	t.Run("requirementsAffected", func(t *testing.T) { requirementsAffectedAgree(t, root) })
	t.Run("goTests", func(t *testing.T) { goTestsAgree(t, root) })
	t.Run("checkFiles", func(t *testing.T) { checkFilesAgree(t, root) })
	t.Run("images", func(t *testing.T) { imagesAgree(t, root) })
	t.Run("coverage", func(t *testing.T) { coverageAgrees(t, root) })
	t.Run("checkInventory", func(t *testing.T) { checkInventoryAgrees(t, root) })
	t.Run("sessionParts", func(t *testing.T) { sessionPartsAgree(t, root) })
}

// identifiersAgree holds three agreements at once: every short name the model
// writes is declared exactly once, every decision record the model names exists
// under docs/decisions, and every record under docs/decisions is named
// somewhere in the model.
//
// The decision tags live in the logical architecture, so the second direction
// reads that register as much as this one does.
func identifiersAgree(t *testing.T, root string) {
	t.Helper()

	model := modelFiles(t, root)
	declared := Declarations(model)
	for _, id := range ShortNames(model) {
		switch declared[id] {
		case 1:
		case 0:
			t.Errorf("%s is written as a short name in the model, and no requirement declares it", id)
		default:
			t.Errorf("%s is declared %d times in the model, and one declaration is the whole of it", id, declared[id])
		}
	}

	records := decisionRecords(t, root)
	onDisk := make(map[string]string, len(records))
	for _, record := range records {
		onDisk[record.ID] = record.Path
	}
	named := DecisionIDs(model)
	for _, id := range slices.Sorted(maps.Keys(named)) {
		if _, ok := onDisk[id]; !ok {
			t.Errorf("%s is named in %s, and docs/decisions carries no such record", id, named[id][0])
		}
	}
	for _, record := range records {
		if len(named[record.ID]) == 0 {
			t.Errorf("%s (%s) is a decision record the model never names", record.ID, record.Path)
		}
	}
}

// requirementsAffectedAgree: every requirement a decision record claims to
// affect is a short name the model declares. A record naming SR-99 is either a
// typing slip or a requirement that was dropped without its record following.
func requirementsAffectedAgree(t *testing.T, root string) {
	t.Helper()

	inModel := nameSet(ShortNames(modelFiles(t, root)))
	for _, record := range decisionRecords(t, root) {
		for _, id := range record.RequirementsAffected {
			if !inModel[id] {
				t.Errorf("%s names %s under Requirements affected, and the model declares no such short name",
					record.Path, id)
			}
		}
	}
}

// goTestsAgree compares the Go test functions of the requirement scheme with
// the ones the verification register names, per file. Per file, because two
// functions share a name across two packages, and a set of names alone would
// hide one of them.
func goTestsAgree(t *testing.T, root string) {
	t.Helper()

	declarations, err := GoTestFunctions(root)
	repository := counted(assert.Must(t, declarations, err))
	model := counted(GoTestEvidence(text(t, root, VerificationCasesFile)))
	for _, test := range unionTests(repository, model) {
		if test.Name == "" {
			t.Errorf("an action of the verification register carries go-test evidence in %s without naming a test",
				test.File)
			continue
		}
		declared, named := repository[test], model[test]
		switch {
		case declared == named:
		case named == 0:
			t.Errorf("%s in %s is a test the verification register does not name", test.Name, test.File)
		case declared == 0:
			t.Errorf("%s is named by the verification register in %s, and there is no such test in that file",
				test.Name, test.File)
		default:
			t.Errorf("%s in %s: the repository declares it %d times and the register names it %d times",
				test.Name, test.File, declared, named)
		}
	}
}

// checkFilesAgree: the check files of the check project and the file names the
// two case registers quote are the same set, and every suite file is exercised
// by exactly one validation case. Before the check project exists both sides
// are empty, which is agreement rather than absence.
func checkFilesAgree(t *testing.T, root string) {
	t.Helper()

	files, err := CheckFiles(root)
	onDisk := assert.Must(t, files, err)
	validation := text(t, root, ValidationCasesFile)
	named := CheckFileNames(text(t, root, VerificationCasesFile) + "\n" + validation)
	bothWays(t, onDisk, named,
		func(name string) string {
			return fmt.Sprintf("%s/%s is a check file no case names", ChecksDir, name)
		},
		func(name string) string {
			return fmt.Sprintf("a case names %s, and %s carries no such file", name, ChecksDir)
		})

	cases := SpecFileCases(validation)
	for _, name := range onDisk {
		if !strings.HasSuffix(name, ".spec.ts") {
			continue
		}
		if times := cases[name]; times != 1 {
			t.Errorf("%s is exercised by %d validation cases, and one case is the whole of it", name, times)
		}
	}
}

// imagesAgree: every published image corresponds to a view that names it, and
// every image a view names is published. Exactly once each, so that two views
// cannot both claim to be the board a reader is looking at.
func imagesAgree(t *testing.T, root string) {
	t.Helper()

	images, err := PNGFiles(root)
	published := assert.Must(t, images, err)
	named := ImageNames(text(t, root, ViewsFile))
	times := make(map[string]int, len(named))
	for _, name := range named {
		times[name]++
	}
	for _, name := range published {
		switch times[name] {
		case 1:
		case 0:
			t.Errorf("%s/%s is published, and no view names it", ImagesDir, name)
		default:
			t.Errorf("%s/%s is named by %d views, and one view is the whole of it", ImagesDir, name, times[name])
		}
	}
	for _, name := range slices.Sorted(maps.Keys(times)) {
		if !slices.Contains(published, name) {
			t.Errorf("a view names %s/%s, and there is no such image", ImagesDir, name)
		}
	}
}

// coverageAgrees: a story called done is verified, satisfied and derived. A
// system story that is done has a verification case reading its acceptance and
// an allocation in the logical architecture, a stakeholder story that is done
// has a validation case, and every system story, done or not, is a derived end
// of a derivation connection. A story still in progress is exempt from the
// first two and reported, because an unfinished story with no evidence is the
// expected state rather than a defect.
func coverageAgrees(t *testing.T, root string) {
	t.Helper()

	stories := text(t, root, SystemStoriesFile)
	system := Stories(stories)
	verified := nameSet(VerifiedStories(text(t, root, VerificationCasesFile)))
	satisfied := nameSet(SatisfiedStories(text(t, root, ComponentsFile)))
	derived := nameSet(DerivedStories(stories))
	for _, story := range system {
		if !derived[story.Name] {
			t.Errorf("%s (%s) is a derived end of no derivation connection", story.ShortName, story.Name)
		}
		if !evidenceIsDue(t, story, "a verification case and an allocation") {
			continue
		}
		if !verified[story.Name] {
			t.Errorf("%s (%s) is done, and no verification case verifies its acceptance", story.ShortName, story.Name)
		}
		if !satisfied[story.Name] {
			t.Errorf("%s (%s) is done, and nothing in the logical architecture satisfies it", story.ShortName, story.Name)
		}
	}

	validation := counted(ValidationCases(text(t, root, ValidationCasesFile)))
	for _, story := range Stories(text(t, root, StakeholderStoriesFile)) {
		if !evidenceIsDue(t, story, "a validation case") {
			continue
		}
		want := "VAL_US_" + strings.TrimPrefix(story.ShortName, "US-")
		if times := validation[want]; times != 1 {
			t.Errorf("%s (%s) is done, and %s is declared %d times, want exactly one validation case",
				story.ShortName, story.Name, want, times)
		}
	}
}

// checkInventoryAgrees: what the check register says each check is, and what
// the manifest the check project reads says it is, are the same, and every
// check in the manifest is constructed once. A check the manifest carries and
// nothing constructs never runs, and one constructed twice runs twice under one
// name. Both sides are empty until the check project exists, as in
// checkFilesAgree.
func checkInventoryAgrees(t *testing.T, root string) {
	t.Helper()

	read, err := ReadManifest(root)
	entries := assert.Must(t, read, err)
	inManifest := byLogicalID(t, "manifest", ManifestTuples(entries))
	inRegister := byLogicalID(t, "check register", CheckCaseTuples(text(t, root, CheckCasesFile)))
	for _, id := range union(inManifest, inRegister) {
		manifest, fromManifest := inManifest[id]
		register, fromRegister := inRegister[id]
		switch {
		case !fromRegister:
			t.Errorf("%s is in the manifest, and the check register declares no such check", id)
		case !fromManifest:
			t.Errorf("%s is in the check register, and the manifest carries no such check", id)
		case manifest != register:
			t.Errorf("%s disagrees:\n  manifest: %+v\n  register: %+v", id, manifest, register)
		}
	}

	built, err := Constructions(root)
	constructed := assert.Must(t, built, err)
	for _, entry := range entries {
		// A suite file is not constructed by name: the check project builds one
		// check per manifest entry that names one, so the entry is its own
		// construction.
		times := constructed[entry.ID]
		if strings.HasSuffix(entry.File, ".spec.ts") {
			times++
		}
		if times != 1 {
			t.Errorf("%s is constructed %d times by the check project, and once is the whole of it",
				entry.ID, times)
		}
	}
}

// sessionPartsAgree: the services the compose file of the check session starts
// are the parts the session composite carries, and nothing else.
func sessionPartsAgree(t *testing.T, root string) {
	t.Helper()
	if !Exists(root, ComposeFile) {
		t.Skip("the compose file is not yet present, so the session has no services to be compared against")
	}

	services := ComposeServices(text(t, root, ComposeFile))
	parts := SessionComposeServices(text(t, root, ComponentsFile))
	bothWays(t, services, parts,
		func(name string) string {
			return fmt.Sprintf("the compose file starts %s, and no part of the session composite carries it", name)
		},
		func(name string) string {
			return fmt.Sprintf("the session composite carries %s, and the compose file starts no such service", name)
		})
}

// evidenceIsDue reports whether a story is far enough along that the evidence
// named by what is expected of it. A story in progress is exempt and logged,
// which is the exemption the requirement grants. Every other answer is a
// failure, including no status at all: the exemption is the one path that turns
// a check off, so a story may not reach it by being unreadable.
func evidenceIsDue(t *testing.T, story Story, what string) bool {
	t.Helper()
	switch story.Status {
	case StatusDone:
		return true
	case StatusInProgress:
		t.Logf("%s (%s) is %s, so %s is not due yet", story.ShortName, story.Name, story.Status, what)
		return false
	case "":
		t.Errorf("%s (%s) carries no readable @StoryMeta", story.ShortName, story.Name)
		return false
	default:
		t.Errorf("%s (%s) carries the status %s, and a story of this model is done or inProgress",
			story.ShortName, story.Name, story.Status)
		return false
	}
}

// text reads one file of the repository, or fails the subtest. It exists
// because assert.Must takes the value and the error as two arguments, Go having
// no way to spread a two-value call beside another one, and a read that appears
// eight times reads better as a call than as a pair of statements.
func text(t *testing.T, root, rel string) string {
	t.Helper()
	found, err := Text(root, rel)
	return assert.Must(t, found, err)
}

// modelFiles reads the whole model, or fails the subtest.
func modelFiles(t *testing.T, root string) []File {
	t.Helper()
	found, err := ModelFiles(root)
	return assert.Must(t, found, err)
}

// decisionRecords reads the decision records, or fails the subtest.
func decisionRecords(t *testing.T, root string) []DecisionRecord {
	t.Helper()
	found, err := DecisionRecords(root)
	return assert.Must(t, found, err)
}

// bothWays reports every name one side carries and the other lacks. The two
// functions turn a name into the failure that direction deserves, which is what
// makes a failure readable without the reader knowing which slice was which.
func bothWays(t *testing.T, got, want []string, missingFromWant, missingFromGot func(string) string) {
	t.Helper()
	inWant, inGot := nameSet(want), nameSet(got)
	for _, name := range sorted(got) {
		if !inWant[name] {
			t.Error(missingFromWant(name))
		}
	}
	for _, name := range sorted(want) {
		if !inGot[name] {
			t.Error(missingFromGot(name))
		}
	}
}

// nameSet turns a list of names into a set, for the membership tests that make
// up most of this file.
func nameSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

// sorted returns the names in order and without repeats, so that a failure
// reads the same way twice.
func sorted(names []string) []string {
	return slices.Compact(slices.Sorted(slices.Values(names)))
}

// counted counts the times each value appears, which is what a comparison of
// two multisets needs.
func counted[T comparable](values []T) map[T]int {
	times := make(map[T]int, len(values))
	for _, value := range values {
		times[value]++
	}
	return times
}

// union returns every key of either map, in order, so that the two can be
// compared in one pass and the failures come out the same way each run.
func union[T cmp.Ordered, V any](first, second map[T]V) []T {
	keys := slices.Sorted(maps.Keys(first))
	for _, key := range slices.Sorted(maps.Keys(second)) {
		if _, ok := first[key]; !ok {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	return keys
}

// unionTests is union for the pair of test name and file, which is a struct and
// therefore has no order of its own. It is ordered by file and then by name,
// which is the order a reader would grep in.
func unionTests(first, second map[TestFunc]int) []TestFunc {
	tests := append(slices.Collect(maps.Keys(first)), slices.Collect(maps.Keys(second))...)
	slices.SortFunc(tests, func(a, b TestFunc) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Name, b.Name))
	})
	return slices.Compact(tests)
}

// byLogicalID keys the check tuples by the identifier they carry, and reports a
// side that names one check twice rather than letting one of the two vanish.
func byLogicalID(t *testing.T, side string, tuples []CheckTuple) map[string]CheckTuple {
	t.Helper()
	byID := make(map[string]CheckTuple, len(tuples))
	for _, tuple := range tuples {
		if _, seen := byID[tuple.LogicalID]; seen {
			t.Errorf("the %s carries %s twice", side, tuple.LogicalID)
		}
		byID[tuple.LogicalID] = tuple
	}
	return byID
}
