package main

import (
	"math"
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

func reply(pick string, evidence ...string) *Reply {
	return &Reply{Answer: Answer{Requirement: pick, Evidence: evidence, Reason: "r"}, Raw: `{"requirement":"` + pick + `"}`}
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
			{Probe: "base", Test: "A", Gold: "SR-01", Reply: reply("SR-01")},
			{Probe: "base", Test: "B", Gold: "SR-02", Reply: reply("none")},
			{Probe: "base", Test: "C", Reply: reply("SR-02")},
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
	same := reply("SR-01", "empty")
	c := Contents{
		Baseline: []BaselineResult{{Test: "A", Gold: "SR-01", Pick: "SR-01"}, {Test: "C", Pick: "SR-01"}},
		Calls: []CallLine{
			{Probe: "base", Test: "A", Gold: "SR-01", Reply: same},
			{Probe: "base", Test: "C", Reply: reply("SR-02")},
			{Probe: "reconstruction", Test: "A", BasePick: "SR-01", Reply: reply("none")},
			{Probe: "reconstruction", Test: "C", BasePick: "SR-02", Note: "the model cited no evidence"},
			{Probe: "repeat", Test: "A", Reply: same},
			{Probe: "deletion", Test: "A", BasePick: "SR-01", Reply: reply("SR-02")},
			{Probe: "control", Test: "A", BasePick: "SR-01", Reply: reply("SR-01")},
			{Probe: "order", Test: "A", BasePick: "SR-01", Reply: reply("SR-01", "empty")},
		},
	}
	s := Summarise(c)
	want := map[string][3]int{ // changed, asked, skipped
		"deletion": {1, 1, 0}, "control": {0, 1, 0}, "reconstruction": {1, 1, 1}, "order": {0, 1, 0}, "repeat": {0, 1, 0},
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
