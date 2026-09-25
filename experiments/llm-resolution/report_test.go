package main

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 0.005 }

func TestEXPSR11_WilsonIntervalsMatchKnownValues(t *testing.T) {
	cases := []struct {
		k, n   int
		lo, hi float64
	}{
		{8, 10, 0.49, 0.94},
		{80, 100, 0.71, 0.87},
		{800, 1000, 0.77, 0.82},
		{69, 69, 0.95, 1.00},
	}
	for _, c := range cases {
		lo, hi := wilson(c.k, c.n)
		if !near(lo, c.lo) || !near(hi, c.hi) {
			t.Errorf("wilson(%d, %d) = %.3f to %.3f, want %.2f to %.2f", c.k, c.n, lo, hi, c.lo, c.hi)
		}
	}
	if lo, hi := wilson(0, 0); lo != 0 || hi != 0 {
		t.Errorf("wilson(0, 0) = %v to %v, want 0 to 0", lo, hi)
	}
}

// final is a browse's final line for a test, linked to pick alone.
func final(probe, test, gold, basePick, pick string, evidence ...string) CallLine {
	a := &TestAnswer{Links: []LinkAnswer{}, Evidence: evidence, Reason: "r"}
	if pick != "none" {
		a.Links = []LinkAnswer{{ID: pick, Relation: "verifies"}}
	}
	return CallLine{Probe: probe, Test: test, Gold: gold, BasePick: basePick, Final: true, Pick: pick,
		Reply: &Reply{Raw: `{"links":"` + pick + `"}`}, Answer: a}
}

func findResolver(s Summary, name string) (ResolverScore, bool) {
	for _, r := range s.Resolvers {
		if r.Resolver == name {
			return r, true
		}
	}
	return ResolverScore{}, false
}

func TestEXPSR11_OnlyKeyedTestsAreScored(t *testing.T) {
	c := Contents{
		Baseline: []BaselineResult{
			{Test: "A", Gold: "SR-01", Pick: "SR-01"},
			{Test: "B", Gold: "SR-02", Pick: "SR-03"},
			{Test: "C", Pick: "SR-01"},
		},
		Calls: []CallLine{
			final("base", "A", "SR-01", "", "SR-01"),
			final("base", "B", "SR-02", "", "none"),
			final("base", "C", "", "", "SR-02"),
		},
	}
	s := Summarise(c)
	lm, ok := findResolver(s, "language model")
	if !ok {
		t.Fatalf("no language model row in %+v", s.Resolvers)
	}
	if lm.Proposed != 1 || lm.Correct != 1 || lm.Precision.K != 1 || lm.Precision.N != 1 || lm.Recall.K != 1 || lm.Recall.N != 2 {
		t.Errorf("language model = %+v, want 1 proposed, 1 correct, precision 1/1, recall 1/2", lm)
	}
	wo, ok := findResolver(s, "word overlap")
	if !ok || wo.Proposed != 2 || wo.Correct != 1 || wo.Precision.N != 2 || wo.Recall.N != 2 {
		t.Errorf("word overlap = %+v, want 2 proposed, 1 correct, over 2 keyed tests", wo)
	}
	kr, ok := findResolver(s, "key rule")
	if !ok || kr.Correct != 2 || kr.Recall.K != 2 || kr.Recall.N != 2 {
		t.Errorf("key rule = %+v, want both keyed tests", kr)
	}
}

func TestEXPSR11_ProbeRatesAndProposalsAreReported(t *testing.T) {
	c := Contents{
		Baseline: []BaselineResult{{Test: "A", Gold: "SR-01", Pick: "SR-01"}, {Test: "C", Pick: "SR-01"}},
		Calls: []CallLine{
			final("base", "A", "SR-01", "", "SR-01", "empty"),
			final("base", "C", "", "", "SR-02"),
			final("reconstruction", "A", "SR-01", "SR-01", "none"),
			{Probe: "reconstruction", Test: "C", BasePick: "SR-02", Final: true, Note: "the model cited no evidence"},
			final("repeat", "A", "SR-01", "SR-01", "SR-01", "empty"),
			final("deletion", "A", "SR-01", "SR-01", "SR-02"),
			final("control", "A", "SR-01", "SR-01", "SR-01"),
			final("rare-shared", "A", "SR-01", "SR-01", "SR-01"),
		},
	}
	s := Summarise(c)
	want := map[string][3]int{ // changed, asked, skipped
		"deletion": {1, 1, 0}, "control": {0, 1, 0}, "reconstruction": {1, 1, 1}, "rare-shared": {0, 1, 0}, "repeat": {0, 1, 0},
	}
	seen := map[string]bool{}
	for _, p := range s.Probes {
		w, ok := want[p.Probe]
		if !ok {
			continue
		}
		seen[p.Probe] = true
		if p.Changed.K != w[0] || p.Changed.N != w[1] || p.Skipped != w[2] {
			t.Errorf("%s: changed %d of %d, skipped %d, want %v", p.Probe, p.Changed.K, p.Changed.N, p.Skipped, w)
		}
		if p.Changed.N > 0 && !(p.Changed.Lo <= float64(p.Changed.K)/float64(p.Changed.N) && float64(p.Changed.K)/float64(p.Changed.N) <= p.Changed.Hi) {
			t.Errorf("%s: interval %v to %v doesn't hold its rate", p.Probe, p.Changed.Lo, p.Changed.Hi)
		}
	}
	if len(seen) != len(want) {
		t.Errorf("probes reported: %+v", s.Probes)
	}
	got := map[string]string{}
	for _, p := range s.Proposals {
		got[p.Resolver+"/"+p.Test] = p.Pick
	}
	if got["language model/C"] != "SR-02" || got["word overlap/C"] != "SR-01" {
		t.Errorf("proposals = %+v, want C proposed by both resolvers", s.Proposals)
	}
	if _, ok := got["language model/A"]; ok {
		t.Error("a keyed test was listed as an unjudged proposal")
	}
	md := s.Markdown()
	for _, needle := range []string{"deletion", "control", "language model", "word overlap", "C"} {
		if !strings.Contains(md, needle) {
			t.Errorf("the report lacks %q", needle)
		}
	}
}

func TestEXPSR11_TheResolversAreComparedTestByTest(t *testing.T) {
	top := func(k string) []Ranked { return []Ranked{{Key: k}} }
	c := Contents{
		Baseline: []BaselineResult{
			{Test: "A", Gold: "SR-01", Pick: "SR-01", Top: top("SR-01")},
			{Test: "B", Gold: "SR-02", Pick: "SR-03", Top: top("SR-03")},
			{Test: "C", Gold: "SR-03", Pick: "none", Top: top("SR-03")},
			{Test: "D", Gold: "SR-04", Pick: "SR-01", Top: top("SR-01")},
		},
		Calls: []CallLine{
			final("base", "A", "SR-01", "", "SR-01"),
			final("base", "B", "SR-02", "", "SR-02"),
			final("base", "C", "SR-03", "", "none"),
			final("base", "D", "SR-04", "", "SR-05"),
		},
	}
	s := Summarise(c)
	want := PairedComparison{BothRight: 1, OnlyLanguageModel: 1, OnlyWordOverlap: 1, NeitherRight: 1, McNemarP: 1}
	if s.Paired != want {
		t.Errorf("paired = %+v, want %+v", s.Paired, want)
	}
	best, ok := findResolver(s, "word overlap, best-ranked")
	if !ok || best.Proposed != 4 || best.Correct != 2 || best.Recall.K != 2 || best.Recall.N != 4 {
		t.Errorf("best-ranked row = %+v", best)
	}
	if p := mcnemarExact(10, 2); math.Abs(p-158.0/4096) > 1e-9 {
		t.Errorf("mcnemarExact(10, 2) = %v, want %v", p, 158.0/4096)
	}
	if p := mcnemarExact(0, 0); p != 1 {
		t.Errorf("mcnemarExact(0, 0) = %v, want 1", p)
	}
	if !strings.Contains(s.Markdown(), "McNemar") {
		t.Error("the report doesn't show the paired comparison")
	}
}

func findProbe(s Summary, name string) (ProbeRate, bool) {
	for _, p := range s.Probes {
		if p.Probe == name {
			return p, true
		}
	}
	return ProbeRate{}, false
}

func TestEXPSR11_ChangedAnswersAreSplitIntoNoneAndAnother(t *testing.T) {
	c := Contents{
		Baseline: []BaselineResult{{Test: "A", Gold: "SR-01", Pick: "SR-01"}},
		Calls: []CallLine{
			final("base", "A", "SR-01", "", "SR-01"),
			final("reconstruction", "A", "SR-01", "SR-01", "none"),
			final("deletion", "A", "SR-01", "SR-01", "SR-02"),
			final("control", "A", "SR-01", "SR-01", "SR-01"),
		},
	}
	s := Summarise(c)
	for name, want := range map[string][2]int{"reconstruction": {1, 0}, "deletion": {0, 1}, "control": {0, 0}} {
		p, ok := findProbe(s, name)
		if !ok || p.ToNone != want[0] || p.ToOther != want[1] {
			t.Errorf("%s: to none %d, to another %d, want %v", name, p.ToNone, p.ToOther, want)
		}
	}
	if !strings.Contains(s.Markdown(), "to none") {
		t.Error("the report doesn't show the split")
	}
}

func TestEXPSR11_EveryProbeIsAlsoReportedOnCorrectLinks(t *testing.T) {
	c := Contents{
		Baseline: []BaselineResult{{Test: "A", Gold: "SR-01", Pick: "SR-01"}, {Test: "C", Pick: "SR-01"}},
		Calls: []CallLine{
			final("base", "A", "SR-01", "", "SR-01"),
			final("base", "C", "", "", "SR-02"),
		},
	}
	for _, p := range []string{"deletion", "control", "rare-shared", "reconstruction"} {
		c.Calls = append(c.Calls, final(p, "A", "SR-01", "SR-01", "SR-03"), final(p, "C", "", "SR-02", "SR-03"))
	}
	s := Summarise(c)
	for _, p := range []string{"deletion", "control", "rare-shared", "reconstruction"} {
		all, ok1 := findProbe(s, p)
		right, ok2 := findProbe(s, p+", correct links only")
		if !ok1 || !ok2 || all.Changed.N != 2 || right.Changed.N != 1 || right.Changed.K != 1 {
			t.Errorf("%s: all %+v, correct links only %+v", p, all.Changed, right.Changed)
		}
	}
}

func TestEXPSR11_UnkeyedProposalsAreListedForBlindJudgement(t *testing.T) {
	c := Contents{
		Baseline: []BaselineResult{
			{Test: "K", Gold: "SR-01", Pick: "SR-01"},
			{Test: "U1", Pick: "SR-01", Reason: []string{"secretword"}},
			{Test: "U2", Pick: "SR-02"},
			{Test: "U3", Pick: "none"},
		},
		Calls: []CallLine{
			final("base", "K", "SR-01", "", "SR-01"),
			{Probe: "base", Test: "U1", Final: true, Pick: "SR-01", Reply: &Reply{Raw: "{}"}, Answer: &TestAnswer{Links: []LinkAnswer{{ID: "SR-01", Relation: "verifies"}}, Reason: "a telling sentence"}},
			{Probe: "base", Test: "U2", Final: true, Pick: "SR-03", Reply: &Reply{Raw: "{}"}, Answer: &TestAnswer{Links: []LinkAnswer{{ID: "SR-03", Relation: "verifies"}}, Reason: "another sentence"}},
			final("base", "U3", "", "", "none"),
		},
	}
	s := Summarise(c)
	got := map[string]bool{}
	for i, b := range s.Blind {
		if b.Number != i+1 {
			t.Errorf("item %d is numbered %d", i, b.Number)
		}
		got[b.Test+"/"+b.Pick] = true
	}
	// U1/SR-01 was proposed by both resolvers and appears once.
	if len(s.Blind) != 3 || !got["U1/SR-01"] || !got["U2/SR-02"] || !got["U2/SR-03"] {
		t.Fatalf("blind list = %+v", s.Blind)
	}
	if !reflect.DeepEqual(Summarise(c).Blind, s.Blind) {
		t.Error("the blind list's order changed between two reports on one file")
	}
	md := s.Markdown()
	blind := md[strings.Index(md, "blind"):]
	for _, leaked := range []string{"secretword", "a telling sentence", "language model", "word overlap"} {
		if strings.Contains(blind, leaked) {
			t.Errorf("the blind list shows %q", leaked)
		}
	}
}

func TestEXPSR11_EvidenceCountsAsFoundOnlyAsWritten(t *testing.T) {
	task := TestTask(Test{Name: "RejectsAnEmptyQuery", Package: "parse", File: "a/parse_test.go", Doc: "an empty query is refused"})
	line := final("base", "A", "SR-01", "", "SR-01",
		// found in the doc, found in the name, only in the question's fixed
		// wording, the right words in the wrong order, found only in SR-01
		"Empty Query", "Rejects", "requirement", "query empty", "parser shall reject")
	line.Task = task.Text
	c := Contents{
		Baseline: []BaselineResult{{Test: "A", Gold: "SR-01", Pick: "SR-01"}},
		Calls:    []CallLine{line},
	}
	if g := Summarise(c).Grounded; g.K != 2 || g.N != 5 {
		t.Errorf("grounded %d of %d, want 2 of 5", g.K, g.N)
	}
}

func TestEXPSR11_OtherKindsAreMeasuredAgainstABaseRate(t *testing.T) {
	a := final("base", "A", "SR-01", "", "SR-01")
	a.Facts = []LinkFact{{ID: "SR-01", Kind: "requirement", Requirement: true}, {ID: "demo::x", Kind: "part", Near: true}, {ID: "Act", Kind: "action"}}
	a.BaseRate = 0.25
	b := final("base", "B", "SR-02", "", "SR-02")
	b.Facts = []LinkFact{{ID: "SR-02", Kind: "requirement", Requirement: true}, {ID: "P2", Kind: "part", Near: true}}
	b.BaseRate = 0.35
	u := final("base", "C", "", "", "SR-02")
	u.Facts = []LinkFact{{ID: "Q", Kind: "part"}}
	s := Summarise(Contents{
		Baseline: []BaselineResult{{Test: "A", Gold: "SR-01"}, {Test: "B", Gold: "SR-02"}, {Test: "C"}},
		Calls:    []CallLine{a, b, u},
	})
	got := map[string][2]int{}
	for _, k := range s.OtherKinds {
		got[k.Kind] = [2]int{k.Links, k.Near}
	}
	if !reflect.DeepEqual(got, map[string][2]int{"part": {2, 2}, "action": {1, 0}}) {
		t.Errorf("other kinds = %+v", s.OtherKinds)
	}
	if s.OtherNear.K != 2 || s.OtherNear.N != 3 || !near(s.BaseRate, 0.30) {
		t.Errorf("near %+v, base rate %v", s.OtherNear, s.BaseRate)
	}
	md := s.Markdown()
	for _, want := range []string{"part", "action", "within three links", "every element"} {
		if !strings.Contains(md, want) {
			t.Errorf("the report lacks %q", want)
		}
	}
}

func TestEXPSR21_SuspectedMismatchesAreListed(t *testing.T) {
	a := final("base", "A", "SR-01", "", "SR-01")
	a.Answer.Mismatches = []Mismatch{{ID: "Adapter::serve", ModelSays: "the store is in serve", SystemShows: "store.go is in projection", Why: "the file sits elsewhere"}}
	inc := CallLine{Probe: ProbeIncident, Test: "incident", Variant: VariantReported, Final: true, Reply: &Reply{Raw: "{}"},
		Account: &IncidentAnswer{Mismatches: []Mismatch{{ID: "SupervisorStates", ModelSays: "nothing restarts it", SystemShows: "no restart policy", Why: "compose.yml"}}},
		Items:   []KeyResult{{Item: "the router is the cause", Found: true}, {Item: "the transition", Found: false}}}
	s := Summarise(Contents{Baseline: []BaselineResult{{Test: "A", Gold: "SR-01"}}, Calls: []CallLine{a, inc}})
	if len(s.Mismatches) != 2 {
		t.Fatalf("mismatches = %+v", s.Mismatches)
	}
	md := s.Markdown()
	for _, want := range []string{"Adapter::serve", "the store is in serve", "store.go is in projection", "the file sits elsewhere",
		"SupervisorStates", "no restart policy", "the router is the cause", VariantReported} {
		if !strings.Contains(md, want) {
			t.Errorf("the report lacks %q", want)
		}
	}
}
