package main

import (
	"reflect"
	"strings"
	"testing"
)

var probeRequirements = []Requirement{
	{Key: "SR-01", Name: "Parse queries", Statement: "The parser shall reject an empty query and report the error to the caller."},
	{Key: "SR-02", Name: "Serve results", Statement: "The server shall return a ranked page for every query."},
}

var probeTest = Test{
	ID: "adapter/parse/parse_test.go#TestSR01_RejectsAnEmptyQuery", FuncName: "TestSR01_RejectsAnEmptyQuery", Name: "RejectsAnEmptyQuery",
	Package: "parse", File: "adapter/parse/parse_test.go", Doc: "RejectsAnEmptyQuery: an empty query is refused.", Gold: "SR-01",
}

// wordsIn are the normalised words of a text.
func wordsIn(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range terms(s) {
		out[w] = true
	}
	return out
}

func sameWordsBut(a, b string, gone map[string]bool) bool {
	return redact(a, gone) == b
}

func TestEXPSR05_DeletionRedactsTheCitedWordsFromEveryToolAnswer(t *testing.T) {
	base := TestAnswer{Links: []LinkAnswer{{ID: "SR-01", Relation: "verifies"}}, Evidence: []string{"empty query"}}
	spec := DeletionProbe(probeTest, base, []string{"no answer shows these words"})
	if spec.Note != "" {
		t.Fatalf("note %q, want a browse", spec.Note)
	}
	if strings.Join(spec.Removed, ",") != "empty,query" || !spec.Redact["empty"] || !spec.Redact["query"] || len(spec.Redact) != 2 {
		t.Fatalf("removed %v, redact %v", spec.Removed, spec.Redact)
	}
	words := wordsIn(spec.Task.Text)
	if words["empty"] || words["query"] || !words["reject"] || !words["refus"] || !words["adapter"] {
		t.Errorf("task = %q", spec.Task.Text)
	}

	plain, _, _ := fixtureKit(t, TaskTest, Scope{})
	redacted, _, _ := fixtureKit(t, TaskTest, Scope{Redact: spec.Redact})
	for _, call := range [][3]string{
		{"doc", "SR-01", ""}, {"links", "SR-01", ""}, {"read", "adapter/parse/parse_test.go", "6"},
		{"find", "empty query parser", ""}, {"grep", "empty", ""}, {"code", "Parser", ""},
	} {
		got := redacted.Call(call[0], call[1], call[2])
		w := wordsIn(got)
		if w["empty"] || w["query"] {
			t.Errorf("%s %s still shows a cited word:\n%s", call[0], call[1], got)
		}
		if call[0] == "doc" || call[0] == "links" || call[0] == "read" {
			if want := plain.Call(call[0], call[1], call[2]); !sameWordsBut(want, got, spec.Redact) {
				t.Errorf("%s %s changed more than the cited words:\n%s\n---\n%s", call[0], call[1], want, got)
			}
		}
	}
	if answer := redacted.Call("grep", "empty", ""); !strings.Contains(answer, "no match") {
		t.Errorf("a search for a deleted word still finds it:\n%s", answer)
	}
}

func TestEXPSR05_ControlRemovesAsManyOtherWords(t *testing.T) {
	base := TestAnswer{Links: []LinkAnswer{{ID: "SR-01", Relation: "verifies"}}, Evidence: []string{"empty query"}}
	deletion := DeletionProbe(probeTest, base, nil)
	spec := ControlProbe(probeTest, base, nil, 42)
	if spec.Note != "" || len(spec.Removed) != len(deletion.Removed) {
		t.Fatalf("control removed %v with note %q, the deletion %v", spec.Removed, spec.Note, deletion.Removed)
	}
	task := wordsIn(TestTask(probeTest).Text)
	for _, w := range spec.Removed {
		if deletion.Redact[w] || !task[w] || !spec.Redact[w] {
			t.Errorf("control removed %q", w)
		}
	}
	after := wordsIn(spec.Task.Text)
	if !after["empty"] || !after["query"] {
		t.Errorf("the control deleted a cited word: %q", spec.Task.Text)
	}
	if again := ControlProbe(probeTest, base, nil, 42); !reflect.DeepEqual(spec, again) {
		t.Error("the same seed chose different words")
	}
	if none := ControlProbe(probeTest, TestAnswer{Evidence: []string{"zzzz"}}, nil, 42); none.Note == "" {
		t.Error("no note when the deletion deletes nothing")
	}
}

func TestEXPSR05_TheRareSharedControlDeletesSharedWordsRarestFirst(t *testing.T) {
	b := NewBaseline(probeRequirements)
	link := []LinkAnswer{{ID: "SR-01", Relation: "verifies"}}
	// Cited "reject": the shared uncited words are empty and parse, which
	// only SR-01 uses, and query, which SR-02 uses too. Empty and parse tie
	// and go alphabetically, and one word matches the deletion's one.
	spec := RareSharedProbe(probeTest, TestAnswer{Links: link, Evidence: []string{"reject"}}, nil, probeRequirements, b)
	if spec.Note != "" || strings.Join(spec.Removed, ",") != "empty" {
		t.Fatalf("removed %v with note %q, want empty alone", spec.Removed, spec.Note)
	}
	// Cited "empty query": two words, so parse and reject, the two rarest
	// shared uncited words, alphabetical within their tie.
	spec = RareSharedProbe(probeTest, TestAnswer{Links: link, Evidence: []string{"empty query"}}, nil, probeRequirements, b)
	if spec.Note != "" || strings.Join(spec.Removed, ",") != "parse,reject" {
		t.Fatalf("removed %v with note %q, want parse then reject", spec.Removed, spec.Note)
	}
	if w := wordsIn(spec.Task.Text); w["parse"] || w["reject"] || !w["empty"] {
		t.Errorf("task = %q", spec.Task.Text)
	}
	if !reflect.DeepEqual(spec, RareSharedProbe(probeTest, TestAnswer{Links: link, Evidence: []string{"empty query"}}, nil, probeRequirements, b)) {
		t.Error("two calls chose differently")
	}
	if n := RareSharedProbe(probeTest, TestAnswer{Links: link, Evidence: []string{"zzzz"}}, nil, probeRequirements, b); n.Note == "" {
		t.Error("no note when the deletion deleted nothing")
	}
	if b.Weight("empty") <= b.Weight("query") || b.Weight("nonexistent") != 0 {
		t.Errorf("weights: empty %v, query %v, unknown %v", b.Weight("empty"), b.Weight("query"), b.Weight("nonexistent"))
	}
}

func TestEXPSR06_ReconstructionShowsOnlyTheCitedWords(t *testing.T) {
	spec := ReconstructionProbe(probeTest, TestAnswer{Evidence: []string{"empty query", " reject "}})
	if spec.Note != "" || spec.HideTest != probeTest.ID {
		t.Fatalf("spec = %+v", spec)
	}
	for _, leaked := range []string{"RejectsAn", "parse_test", "adapter", "refused"} {
		if strings.Contains(spec.Task.Text, leaked) {
			t.Errorf("the task shows %q from the test:\n%s", leaked, spec.Task.Text)
		}
	}
	if !strings.Contains(spec.Task.Text, "empty query; reject") {
		t.Errorf("the task lacks the evidence:\n%s", spec.Task.Text)
	}
	if n := ReconstructionProbe(probeTest, TestAnswer{}); n.Note == "" {
		t.Error("no note when there is no evidence to show")
	}
}

func TestEXPSR06_TheRebuiltBrowseHidesTheTestsOwnCode(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskTest, Scope{HideTest: probeTest.ID})
	if answer := kit.Call("grep", "RejectsAnEmptyQuery", ""); strings.Contains(answer, "parse_test.go:5") || strings.Contains(answer, "parse_test.go:6") {
		t.Errorf("grep shows the hidden test:\n%s", answer)
	}
	read := kit.Call("read", "adapter/parse/parse_test.go", "6")
	if strings.Contains(read, "RejectsAnEmptyQuery") || !strings.Contains(read, "ParsesTokens") {
		t.Errorf("read shows the hidden test, or hides its neighbours:\n%s", read)
	}
	if other := kit.Call("grep", "ParsesTokens", ""); !strings.Contains(other, "parse_test.go:8") {
		t.Errorf("grep hides another test:\n%s", other)
	}
}

func TestEXPSR22_ARemovalTakesAwayOnlyItsTarget(t *testing.T) {
	w, _ := fixtureWiki(t)
	key, err := LoadKey(fixtureKey)
	if err != nil {
		t.Fatal(err)
	}
	target, ok := w.Get(key.RemoveElement)
	if !ok {
		t.Fatalf("no %s", key.RemoveElement)
	}
	cut := w.Without(Removal{Elements: []string{key.RemoveElement}})
	if _, ok := cut.Get(target.ID); ok {
		t.Error("the removed transition is still there")
	}
	if len(cut.Elements) != len(w.Elements)-1 {
		t.Errorf("%d elements left of %d", len(cut.Elements), len(w.Elements))
	}
	for _, e := range w.Elements {
		if e.ID == target.ID {
			continue
		}
		var want []LinkView
		for _, l := range w.LinksOf(e.ID) {
			if l.Other != target.ID {
				want = append(want, l)
			}
		}
		if got := cut.LinksOf(e.ID); !reflect.DeepEqual(got, want) {
			t.Errorf("%s's links changed beyond the removal:\n%v\n%v", e.ID, want, got)
		}
		if other, _ := cut.Get(e.ID); other == nil || other.ID != e.ID {
			t.Errorf("%s changed its identifier", e.ID)
		}
	}
	if hasLink(cut, "ParserExited", "accepted by", target.ID) || !hasLink(w, "ParserExited", "accepted by", target.ID) {
		t.Error("the transition's accept doesn't go with it")
	}

	from, _ := w.Get(key.ControlLink.From)
	to, _ := w.Get(key.ControlLink.To)
	unlinked := w.Without(Removal{Links: []LinkRef{key.ControlLink}})
	if hasLink(unlinked, from.ID, key.ControlLink.Rel, to.ID) || hasLink(unlinked, to.ID, Reverse(key.ControlLink.Rel), from.ID) {
		t.Error("the control link is still there")
	}
	if len(unlinked.Elements) != len(w.Elements) {
		t.Error("removing a link removed an element")
	}
	for _, e := range w.Elements {
		if e.ID == from.ID || e.ID == to.ID {
			continue
		}
		if !reflect.DeepEqual(unlinked.LinksOf(e.ID), w.LinksOf(e.ID)) {
			t.Errorf("%s's links changed", e.ID)
		}
	}
	if len(w.LinksOf(target.ID)) == 0 {
		t.Error("removing from a copy changed the original")
	}
}
