package main

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

var probeRequirements = []Requirement{
	{Key: "SR-01", Name: "Parse queries", Statement: "The parser shall reject an empty query and report the error to the caller."},
	{Key: "SR-02", Name: "Serve results", Statement: "The server shall return a ranked page for every query."},
}

var probeTest = Test{
	ID: "adapter/parse/parse_test.go#TestSR01_RejectsAnEmptyQuery", Name: "RejectsAnEmptyQuery",
	Package: "parse", File: "adapter/parse/parse_test.go", Doc: "RejectsAnEmptyQuery: an empty query is refused.", Gold: "SR-01",
}

// occurrences counts the words of fields that are in words.
func occurrences(words map[string]bool, fields ...string) int {
	n := 0
	for _, f := range fields {
		for _, w := range terms(f) {
			if words[w] {
				n++
			}
		}
	}
	return n
}

func TestEXPSR05_DeletionRemovesTheCitedWordsFromBothSides(t *testing.T) {
	base := Answer{Requirement: "SR-01", Evidence: []string{"empty query"}}
	v := deletionVariant(probeRequirements, probeTest, base)
	if v.Note != "" {
		t.Fatalf("note %q, want a question", v.Note)
	}
	want := TestView{Name: "RejectsAn", Package: "parse", File: "adapter/parse/parse_test.go", Doc: "RejectsAn: an is refused."}
	if v.View != want {
		t.Errorf("view = %+v, want %+v", v.View, want)
	}
	if len(v.Requirements) != 2 {
		t.Fatalf("%d requirements, want 2", len(v.Requirements))
	}
	picked := v.Requirements[0]
	if picked.Name != "Parse" || picked.Statement != "The parser shall reject an and report the error to the caller." {
		t.Errorf("picked requirement became %q / %q", picked.Name, picked.Statement)
	}
	if v.Requirements[1] != probeRequirements[1] {
		t.Errorf("a requirement that wasn't picked changed: %+v", v.Requirements[1])
	}
	if probeRequirements[0].Name != "Parse queries" {
		t.Error("the probe changed the caller's requirements")
	}
	if strings.Join(v.Removed, ",") != "empty,query" {
		t.Errorf("removed = %v", v.Removed)
	}
}

func TestEXPSR05_ControlRemovesAsManyOtherWords(t *testing.T) {
	base := Answer{Requirement: "SR-01", Evidence: []string{"empty query"}}
	cited := map[string]bool{"empty": true, "query": true}
	deleted := occurrences(cited, probeTest.Name, probeTest.Doc, probeTest.File,
		probeRequirements[0].Name, probeRequirements[0].Statement)
	v := controlVariant(probeRequirements, probeTest, base, 42)
	if v.Note != "" || deleted == 0 {
		t.Fatalf("note %q with %d cited occurrences", v.Note, deleted)
	}
	if len(v.Removed) != deleted {
		t.Fatalf("control removed %d words (%v), the deletion removes %d", len(v.Removed), v.Removed, deleted)
	}
	for _, w := range v.Removed {
		if cited[w] {
			t.Errorf("control removed the cited word %q", w)
		}
	}
	// The cited words are all still there.
	after := occurrences(cited, v.View.Name, v.View.Doc, v.View.File, v.Requirements[0].Name, v.Requirements[0].Statement)
	if after != deleted {
		t.Errorf("%d cited occurrences left, want %d", after, deleted)
	}
	again := controlVariant(probeRequirements, probeTest, base, 42)
	if !reflect.DeepEqual(v, again) {
		t.Error("the same seed chose different words")
	}
}

func TestEXPSR06_ReconstructionShowsOnlyTheCitedWords(t *testing.T) {
	base := Answer{Requirement: "SR-01", Evidence: []string{"empty query", " reject "}}
	v := reconstructionVariant(probeRequirements, base)
	if v.View != (TestView{Doc: "empty query; reject"}) {
		t.Fatalf("view = %+v", v.View)
	}
	user := userPrompt(v.View)
	for _, leaked := range []string{"RejectsAn", "parse_test", "adapter"} {
		if strings.Contains(user, leaked) {
			t.Errorf("the question shows %q from the test:\n%s", leaked, user)
		}
	}
	if !strings.Contains(user, "empty query; reject") {
		t.Errorf("the question lacks the evidence:\n%s", user)
	}
	if n := reconstructionVariant(probeRequirements, Answer{Requirement: "SR-01"}); n.Note == "" {
		t.Error("no note when there is no evidence to show")
	}
}

func TestEXPSR07_OneShuffleForEveryTest(t *testing.T) {
	c, err := LoadCorpus(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	a := orderVariant(c.Requirements, 42)
	b := orderVariant(c.Requirements, 42)
	if !reflect.DeepEqual(a.Requirements, b.Requirements) {
		t.Fatal("two tests got different orders from the same seed")
	}
	if reflect.DeepEqual(a.Requirements, c.Requirements) {
		t.Fatal("the shuffled order is the original order")
	}
	var x, y []string
	for i := range c.Requirements {
		x = append(x, c.Requirements[i].Key)
		y = append(y, a.Requirements[i].Key)
	}
	sort.Strings(x)
	sort.Strings(y)
	if !reflect.DeepEqual(x, y) {
		t.Error("the shuffle isn't a permutation of the requirements")
	}
	if reflect.DeepEqual(orderVariant(c.Requirements, 43).Requirements, a.Requirements) {
		t.Error("another seed gave the same order")
	}
}
