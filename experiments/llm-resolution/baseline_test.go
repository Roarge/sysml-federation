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

// modelFixture is the systems model baseline over the fixture's test task,
// with the register's Go tests hidden as the tools hide them.
func modelFixture(t *testing.T) (*ModelBaseline, *Wiki) {
	t.Helper()
	w, root := fixtureWiki(t)
	c, err := LoadCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	tw := w.withoutGoTests()
	return NewModelBaseline(tw, c.Requirements), tw
}

// ranking lists elements best first, as the search would rank them.
func ranking(t *testing.T, w *Wiki, refs ...string) []Ranked {
	t.Helper()
	var out []Ranked
	for i, ref := range refs {
		e, ok := w.Get(ref)
		if !ok {
			t.Fatalf("the fixture has no element %s", ref)
		}
		out = append(out, Ranked{Key: e.ID, Score: 1 - float64(i)/100})
	}
	return out
}

func idOf(t *testing.T, w *Wiki, ref string) string {
	t.Helper()
	e, ok := w.Get(ref)
	if !ok {
		t.Fatalf("the fixture has no element %s", ref)
	}
	return e.ID
}

func TestEXPSR24_ARequirementRankedFirstIsTheAnswer(t *testing.T) {
	m, w := modelFixture(t)
	pick, via := m.pickFrom(ranking(t, w, "SR-01", "Fixture_VerificationCases::VC_SR_02"), Test{Name: "ParsesTokens"})
	if pick != "SR-01" || via != "" {
		t.Errorf("pick %q via %q, want SR-01 directly", pick, via)
	}
}

func TestEXPSR24_OneTraceLinkLeadsToARequirement(t *testing.T) {
	m, w := modelFixture(t)
	vc := idOf(t, w, "Fixture_VerificationCases::VC_SR_02")
	if pick, via := m.pickFrom(ranking(t, w, vc), Test{}); pick != "SR-02" || via != vc {
		t.Errorf("from the verification case: pick %q via %q, want SR-02 via %s", pick, via, vc)
	}
	parser := idOf(t, w, "Fixture_LogicalArchitecture::system::parser")
	if pick, via := m.pickFrom(ranking(t, w, parser), Test{}); pick != "SR-01" || via != parser {
		t.Errorf("from the parser: pick %q via %q, want SR-01 via %s", pick, via, parser)
	}
}

func TestEXPSR24_WordOverlapBreaksTies(t *testing.T) {
	m, w := modelFixture(t)
	// The server satisfies SR-02 and SR-03.
	server := ranking(t, w, "Fixture_LogicalArchitecture::system::server")
	if pick, _ := m.pickFrom(server, Test{Name: "ReportsLatency", Doc: "the latency of every page it returns"}); pick != "SR-03" {
		t.Errorf("a test about latency: pick %q, want SR-03", pick)
	}
	if pick, _ := m.pickFrom(server, Test{Name: "ReturnsARankedPage", Doc: "the top results as a ranked page"}); pick != "SR-02" {
		t.Errorf("a test about ranked pages: pick %q, want SR-02", pick)
	}
}

func TestEXPSR24_OwnershipIsNoTraceLink(t *testing.T) {
	m, w := modelFixture(t)
	// The package owns every system story, and the part definition only
	// types the server.
	if pick, via := m.pickFrom(ranking(t, w, "Fixture_SystemStories", "Fixture_LogicalArchitecture::Server"), Test{}); pick != "none" || via != "" {
		t.Errorf("pick %q via %q, want none", pick, via)
	}
	vc := idOf(t, w, "Fixture_VerificationCases::VC_SR_02")
	if pick, via := m.pickFrom(ranking(t, w, "Fixture_SystemStories", "Fixture_LogicalArchitecture::Server", vc), Test{}); pick != "SR-02" || via != vc {
		t.Errorf("passed over, they leave the answer to the next element: pick %q via %q", pick, via)
	}
	if pick, _ := m.pickFrom(nil, Test{}); pick != "none" {
		t.Errorf("an empty ranking: pick %q, want none", pick)
	}
}

func TestEXPSR24_TheFirstSearchsReachIsCounted(t *testing.T) {
	m, w := modelFixture(t)
	vc := ranking(t, w, "Fixture_VerificationCases::VC_SR_02")
	if !m.reach(vc, "SR-02") {
		t.Error("SR-02 is one verify link from its verification case")
	}
	if m.reach(vc, "SR-01") {
		t.Error("SR-01 is more than one trace link from VC_SR_02")
	}
	if !m.reach(ranking(t, w, "SR-03"), "SR-03") {
		t.Error("a requirement in the ranking is within reach")
	}
	// Only the first eight count: eight elements with no trace link to SR-01,
	// then SR-01 itself.
	var refs []string
	for _, e := range w.Elements {
		if len(refs) == LimitFind {
			break
		}
		if e.Short == "SR-01" {
			continue
		}
		linked := false
		for _, l := range w.LinksOf(e.ID) {
			if o, ok := w.Get(l.Other); ok && o.Short == "SR-01" && traceRelations[l.Rel] {
				linked = true
			}
		}
		if !linked {
			refs = append(refs, e.ID)
		}
	}
	if len(refs) < LimitFind {
		t.Fatalf("the fixture has only %d elements away from SR-01", len(refs))
	}
	if m.reach(ranking(t, w, append(refs, "SR-01")...), "SR-01") {
		t.Error("a requirement ninth in the ranking is out of reach")
	}
}
