package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEXPSR20_TheKeyNamesOnlyElementsThatExist(t *testing.T) {
	key, err := LoadKey("incident-key.json")
	if err != nil {
		t.Fatal(err)
	}
	realKit(t, TaskIncident) // loads the checkout's code
	if err := key.Check(realWiki(t), realCode); err != nil {
		t.Errorf("the committed key: %v", err)
	}
	if len(key.Items) < 8 || key.Report == "" || key.AlertOnly == "" || !strings.HasPrefix(key.Report, key.AlertOnly) {
		t.Errorf("the key's report, alert and items: %+v", key)
	}
	for _, field := range []string{"cause", "mechanism", "code", "consequences"} {
		found := false
		for _, it := range key.Items {
			found = found || it.Field == field
		}
		if !found {
			t.Errorf("the key has no item for %s", field)
		}
	}

	fixture, err := LoadKey(fixtureKey)
	if err != nil {
		t.Fatal(err)
	}
	w, root := fixtureWiki(t)
	code, err := LoadCode(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.Check(w, code); err != nil {
		t.Errorf("the fixture's key: %v", err)
	}
	fixture.Items[0].Any = append(fixture.Items[0].Any, "Fixture_LogicalArchitecture::NoSuchPart")
	if err := fixture.Check(w, code); err == nil || !strings.Contains(err.Error(), "NoSuchPart") {
		t.Errorf("a key naming a missing element: %v", err)
	}
	bad, err := LoadKey(fixtureKey)
	if err != nil {
		t.Fatal(err)
	}
	bad.Items[3].Any = []string{"adapter/serve/nothing.go"}
	if err := bad.Check(w, code); err == nil || !strings.Contains(err.Error(), "nothing.go") {
		t.Errorf("a key naming a missing file: %v", err)
	}
}

func TestEXPSR20_TheIncidentSeesTheWholeSystemsModel(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	if links := kit.Call("links", "VC_SR_02", ""); !strings.Contains(links, "TestSR02_ReturnsARankedPage") {
		t.Errorf("the incident's links hide the register:\n%s", links)
	}
	if doc := kit.Call("doc", "TestSR02_ReturnsARankedPage", ""); strings.Contains(doc, "unknown") {
		t.Errorf("the incident can't read a registered test's action:\n%s", doc)
	}
	if read := kit.Call("read", "adapter/parse/parse_test.go", "6"); !strings.Contains(read, "func TestSR01_RejectsAnEmptyQuery") {
		t.Errorf("the incident's read masks a key:\n%s", read)
	}
}

func TestEXPSR20_TheIncidentIsScoredItemByItem(t *testing.T) {
	key, err := LoadKey(fixtureKey)
	if err != nil {
		t.Fatal(err)
	}
	w, _ := fixtureWiki(t)
	a := IncidentAnswer{
		Cause:        []string{"system::parser"},
		Mechanism:    []string{"ServerStates"},
		Code:         []string{"adapter/serve/serve.go:3"},
		Consequences: []string{"SR-02", "SR-03"},
		Path:         []string{},
	}
	got := map[string]bool{}
	for _, r := range ScoreIncident(key, a, w) {
		got[r.Item] = r.Found
	}
	want := map[string]bool{
		"the parser is the cause":                     true,
		"the server's state machine is the mechanism": true,
		"the transition on the parser's exit":         false,
		"serve.go is the code to read":                true,
		"SR-02 is unmet":                              true,
		"the stakeholder story behind SR-02":          false,
	}
	if len(got) != len(want) {
		t.Fatalf("scored %v", got)
	}
	for item, found := range want {
		if got[item] != found {
			t.Errorf("%s: found %v, want %v", item, got[item], found)
		}
	}
	// An item cited in another field than its own is not found.
	a.Cause, a.Consequences = []string{"SR-02"}, []string{"system::parser"}
	for _, r := range ScoreIncident(key, a, w) {
		if (r.Item == "the parser is the cause" || r.Item == "SR-02 is unmet") && r.Found {
			t.Errorf("%s found in the wrong field", r.Item)
		}
	}
}

// citation finds one checked citation.
func citation(cs []Citation, field, ref string) (Citation, bool) {
	for _, c := range cs {
		if c.Field == field && c.Ref == ref {
			return c, true
		}
	}
	return Citation{}, false
}

func TestEXPSR23_AMissingElementIsFlagged(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	a := IncidentAnswer{
		Cause:        []string{"NoSuchElement"},
		Mechanism:    []string{"ServerStates"},
		Code:         []string{"adapter/serve/serve.go", "adapter/serve/gone.go"},
		Consequences: []string{"SR-02"},
		Path:         []string{},
	}
	cs := CheckCitations(a, kit)
	for _, c := range []struct {
		field, ref string
		found      bool
	}{
		{"cause", "NoSuchElement", false},
		{"mechanism", "ServerStates", true},
		{"code", "adapter/serve/serve.go", true},
		{"code", "adapter/serve/gone.go", false},
		{"consequences", "SR-02", true},
	} {
		got, ok := citation(cs, c.field, c.ref)
		if !ok || got.Found != c.found {
			t.Errorf("%s %s: %+v, want found %v", c.field, c.ref, got, c.found)
		}
	}
	tests := CheckTestCitations(TestAnswer{Links: []LinkAnswer{{ID: "SR-02", Relation: "verifies"}, {ID: "Madeup", Relation: "other"}}}, kit)
	if c, ok := citation(tests, "links", "Madeup"); !ok || c.Found {
		t.Errorf("a made-up link in a test's answer: %+v", tests)
	}
	if c, ok := citation(tests, "links", "SR-02"); !ok || !c.Found {
		t.Errorf("a real link in a test's answer: %+v", tests)
	}
}

func TestEXPSR23_ConsecutiveStepsMustBeLinked(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	kit.Call("read", "adapter/serve/serve.go", "3")
	check := func(path []string) []Citation {
		return CheckCitations(IncidentAnswer{Path: path}, kit)
	}
	linked := check([]string{"SR-02", "system::server", "Server", "ServerStates", "adapter/serve/serve.go"})
	for _, c := range linked {
		if !c.Found {
			t.Errorf("a linked path is flagged at %s: %s", c.Ref, c.Why)
		}
	}
	skipped := check([]string{"SR-02", "system::server", "ServerStates"})
	if c, ok := citation(skipped, "path", "ServerStates"); !ok || c.Found || c.Why == "" {
		t.Errorf("a step two links on is not flagged: %+v", skipped)
	}
	if c, ok := citation(skipped, "path", "system::server"); !ok || !c.Found {
		t.Errorf("the linked step before it is flagged: %+v", skipped)
	}
	unread := check([]string{"ServerStates", "adapter/parse/parse.go"})
	if c, ok := citation(unread, "path", "adapter/parse/parse.go"); !ok || c.Found {
		t.Errorf("a file the state machine doesn't name and the run didn't read is not flagged: %+v", unread)
	}
}

func TestEXPSR23_ACitationOfARemovedElementIsFlagged(t *testing.T) {
	key, err := LoadKey(fixtureKey)
	if err != nil {
		t.Fatal(err)
	}
	w, root := fixtureWiki(t)
	code, err := LoadCode(root)
	if err != nil {
		t.Fatal(err)
	}
	c, err := LoadCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	removed := w.Without(Removal{Elements: []string{key.RemoveElement}})
	kit := NewToolKit(removed, code, c.Requirements, Scope{Task: TaskIncident})
	a := IncidentAnswer{Mechanism: []string{"ServerStates::onParserExit", "ServerStates"}, Path: []string{"ServerStates", "ServerStates::onParserExit"}}
	cs := CheckCitations(a, kit)
	if got, ok := citation(cs, "mechanism", "ServerStates::onParserExit"); !ok || got.Found {
		t.Errorf("the removed transition is not flagged: %+v", cs)
	}
	if got, ok := citation(cs, "mechanism", "ServerStates"); !ok || !got.Found {
		t.Errorf("the state machine itself is flagged: %+v", cs)
	}
	if got, ok := citation(cs, "path", "ServerStates::onParserExit"); !ok || got.Found {
		t.Errorf("the removed transition in the path is not flagged: %+v", cs)
	}
	if _, err := os.Stat(filepath.Join(root, "adapter/serve/serve.go")); err != nil {
		t.Fatal(err)
	}
}
