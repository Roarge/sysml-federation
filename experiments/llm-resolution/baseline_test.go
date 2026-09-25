package main

import (
	"strings"
	"testing"
)

var toyRequirements = []Requirement{
	{Key: "R-1", Name: "Parse queries", Statement: "The parser shall reject an empty query."},
	{Key: "R-2", Name: "Serve pages", Statement: "The server shall return a page for every query."},
	{Key: "R-3", Name: "Report latency", Statement: "The server shall report the latency of every page."},
}

func TestEXPSR03_RareSharedWordsRankFirst(t *testing.T) {
	b := NewBaseline(toyRequirements)
	// "empty" is R-1's alone, while "query" is shared with R-2.
	got := b.Resolve(Test{ID: "a", Name: "RejectsEmptyQuery"})
	if got.Pick != "R-1" || len(got.Top) < 2 || got.Top[0].Key != "R-1" {
		t.Fatalf("pick %q, top %v, want R-1 first", got.Pick, got.Top)
	}
	// "latency" is R-3's alone, while "page" is shared with R-2.
	got = b.Resolve(Test{ID: "b", Name: "ReportsLatencyOfPage"})
	if got.Pick != "R-3" {
		t.Fatalf("pick %q, top %v, want R-3", got.Pick, got.Top)
	}
	if got.Top[0].Score <= got.Top[1].Score {
		t.Errorf("scores %v, want the first strictly ahead", got.Top)
	}
}

func TestEXPSR03_NoProposalBelowTheThreshold(t *testing.T) {
	b := NewBaseline(toyRequirements)
	got := b.Resolve(Test{ID: "c", Name: "Health", Package: "health", File: "cmd/health_test.go"})
	if got.Pick != "none" || len(got.Reason) != 0 {
		t.Fatalf("pick %q with reason %v, want none and no reason", got.Pick, got.Reason)
	}
	if len(got.Top) != 3 {
		t.Errorf("top = %v, want the three best whatever their scores", got.Top)
	}
	if BaselineThreshold <= 0 || BaselineThreshold >= 1 {
		t.Errorf("threshold %v is not a similarity between 0 and 1", BaselineThreshold)
	}
}

func TestEXPSR03_TheReasonIsTheSharedWords(t *testing.T) {
	b := NewBaseline(toyRequirements)
	got := b.Resolve(Test{ID: "a", Name: "RejectsEmptyQuery", Gold: "R-1"})
	// A word's weight is what it adds to the score. query is common, but R-1
	// uses it twice, so it adds most. empty and reject are R-1's alone and add
	// the same, so they follow in alphabetical order.
	if want := "query,empty,reject"; strings.Join(got.Reason, ",") != want {
		t.Fatalf("reason = %v, want %s", got.Reason, want)
	}
	if got.Rank != 1 || got.Gold != "R-1" {
		t.Errorf("gold %q ranked %d, want R-1 ranked 1", got.Gold, got.Rank)
	}
}

func TestEXPSR03_AnAcronymsPluralStaysOneWord(t *testing.T) {
	cases := map[string]string{
		"RoutingURLsAreLoopback": "Routing URLs Are Loopback",
		"IDsAndAPIs":             "IDs And APIs",
		"HTTPServer":             "HTTP Server",
		"parseURL":               "parse URL",
		"the routing URLs":       "the routing URLs",
		"ReadsRequirements":      "Reads Requirements",
	}
	for in, want := range cases {
		if got := strings.Join(splitIdentifier(in), " "); got != want {
			t.Errorf("splitIdentifier(%q) = %q, want %q", in, got, want)
		}
	}
}
