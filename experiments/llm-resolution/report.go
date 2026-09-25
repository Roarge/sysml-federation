package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Rate is k of n, with its 95% Wilson interval.
type Rate struct {
	K  int     `json:"k"`
	N  int     `json:"n"`
	Lo float64 `json:"lo"`
	Hi float64 `json:"hi"`
}

func rate(k, n int) Rate {
	lo, hi := wilson(k, n)
	return Rate{K: k, N: n, Lo: lo, Hi: hi}
}

func (r Rate) String() string {
	if r.N == 0 {
		return "no cases"
	}
	return fmt.Sprintf("%.2f (%.2f to %.2f)", float64(r.K)/float64(r.N), r.Lo, r.Hi)
}

// wilson is the 95% Wilson score interval for k successes in n trials, the
// interval Brown, Cai and DasGupta recommend for small samples.
func wilson(k, n int) (float64, float64) {
	if n == 0 {
		return 0, 0
	}
	const z = 1.959963984540054
	p := float64(k) / float64(n)
	nf := float64(n)
	denom := 1 + z*z/nf
	centre := (p + z*z/(2*nf)) / denom
	half := z * math.Sqrt(p*(1-p)/nf+z*z/(4*nf*nf)) / denom
	return math.Max(0, centre-half), math.Min(1, centre+half)
}

// ResolverScore is one resolver's result on the tests that carry a key.
type ResolverScore struct {
	Resolver  string `json:"resolver"`
	Proposed  int    `json:"proposed"`
	Correct   int    `json:"correct"`
	Precision Rate   `json:"precision"`
	Recall    Rate   `json:"recall"`
}

// ProbeRate is how often a probe changed the model's answer.
type ProbeRate struct {
	Probe   string `json:"probe"`
	Changed Rate   `json:"changed"`
	ToNone  int    `json:"to_none"`
	ToOther int    `json:"to_another"`
	Skipped int    `json:"skipped"`
	Errors  int    `json:"errors"`
}

// PairedComparison counts the keyed tests by which hidden-key resolver got
// each right: the language model, or the word overlap's best-ranked answer.
type PairedComparison struct {
	BothRight         int     `json:"both_right"`
	OnlyLanguageModel int     `json:"only_language_model"`
	OnlyWordOverlap   int     `json:"only_word_overlap"`
	NeitherRight      int     `json:"neither_right"`
	McNemarP          float64 `json:"mcnemar_exact_p"`
}

// BlindItem is one proposal for an unkeyed test, with nothing to say who
// proposed it or why.
type BlindItem struct {
	Number int    `json:"number"`
	Test   string `json:"test"`
	Pick   string `json:"pick"`
}

// mcnemarExact is not built yet.
func mcnemarExact(b, c int) float64 { return 0 }

// Proposal is a link proposed for a test that carries no key. Nobody recorded
// an answer for these, so they are listed for a person to judge.
type Proposal struct {
	Test     string `json:"test"`
	Resolver string `json:"resolver"`
	Pick     string `json:"pick"`
	Reason   string `json:"reason"`
}

// ProbeTiming is what the calls of one probe cost.
type ProbeTiming struct {
	Probe        string  `json:"probe"`
	Questions    int     `json:"questions"`
	MeanSeconds  float64 `json:"mean_seconds"`
	MeanPrompt   float64 `json:"mean_prompt_tokens"`
	MaxPrompt    int     `json:"max_prompt_tokens"`
	TotalSeconds float64 `json:"total_seconds"`
}

// Summary is the report on one results file.
type Summary struct {
	Header       RunHeader        `json:"-"`
	Tests        int              `json:"tests"`
	Keyed        int              `json:"keyed_tests"`
	Resolvers    []ResolverScore  `json:"resolvers"`
	Probes       []ProbeRate      `json:"probes"`
	Grounded     Rate             `json:"evidence_grounded"`
	OrderOverlap float64          `json:"order_evidence_overlap"`
	Proposals    []Proposal       `json:"proposals"`
	Blind        []BlindItem      `json:"blind"`
	Paired       PairedComparison `json:"paired"`
	Timing       []ProbeTiming    `json:"timing"`
	BaseErrors   int              `json:"base_errors"`
}

// Summarise computes the report from a results file's contents alone
// (EXP-SR-11, EXP-SR-15).
func Summarise(c Contents) Summary {
	s := Summary{Header: c.Header}
	gold := map[string]string{}
	tests := map[string]bool{}
	for _, b := range c.Baseline {
		tests[b.Test] = true
		if b.Gold != "" {
			gold[b.Test] = b.Gold
		}
	}
	base := map[string]CallLine{}
	for _, cl := range c.Calls {
		if cl.Probe == ProbeBase {
			base[cl.Test] = cl
			tests[cl.Test] = true
			if cl.Gold != "" {
				gold[cl.Test] = cl.Gold
			}
		}
	}
	s.Tests, s.Keyed = len(tests), len(gold)

	// The key rule reads the names as written, so it finds every keyed test.
	s.Resolvers = append(s.Resolvers, ResolverScore{
		Resolver: "key rule", Proposed: s.Keyed, Correct: s.Keyed,
		Precision: rate(s.Keyed, s.Keyed), Recall: rate(s.Keyed, s.Keyed),
	})

	wo := ResolverScore{Resolver: "word overlap"}
	for _, b := range c.Baseline {
		g := gold[b.Test]
		switch {
		case g == "" && b.Pick != "none" && b.Pick != "":
			s.Proposals = append(s.Proposals, Proposal{Test: b.Test, Resolver: wo.Resolver, Pick: b.Pick, Reason: strings.Join(b.Reason, ", ")})
		case g != "" && b.Pick != "none" && b.Pick != "":
			wo.Proposed++
			if b.Pick == g {
				wo.Correct++
			}
		}
	}
	wo.Precision, wo.Recall = rate(wo.Correct, wo.Proposed), rate(wo.Correct, s.Keyed)
	s.Resolvers = append(s.Resolvers, wo)

	lm := ResolverScore{Resolver: "language model"}
	var groundedItems, items int
	for _, t := range sortedKeys(base) {
		cl := base[t]
		if cl.Reply == nil || cl.Error != "" {
			s.BaseErrors++
			continue
		}
		a := cl.Reply.Answer
		if a.Requirement == "none" {
			continue
		}
		for _, ok := range groundedInLine(a, cl.User, c.Header.RequirementList) {
			items++
			if ok {
				groundedItems++
			}
		}
		g := gold[t]
		if g == "" {
			s.Proposals = append(s.Proposals, Proposal{Test: t, Resolver: lm.Resolver, Pick: a.Requirement, Reason: a.Reason})
			continue
		}
		lm.Proposed++
		if a.Requirement == g {
			lm.Correct++
		}
	}
	lm.Precision, lm.Recall = rate(lm.Correct, lm.Proposed), rate(lm.Correct, s.Keyed)
	s.Resolvers = append(s.Resolvers, lm)
	s.Grounded = rate(groundedItems, items)

	s.Probes, s.OrderOverlap = probeRates(c.Calls, base, gold)
	s.Timing = timings(c.Calls)
	sort.SliceStable(s.Proposals, func(i, j int) bool {
		if s.Proposals[i].Test != s.Proposals[j].Test {
			return s.Proposals[i].Test < s.Proposals[j].Test
		}
		return s.Proposals[i].Resolver > s.Proposals[j].Resolver
	})
	return s
}

func sortedKeys(m map[string]CallLine) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// groundedInLine checks the evidence against the question's test part and
// the picked requirement's line in the list the run recorded.
func groundedInLine(a Answer, user, list string) []bool {
	req := ""
	for _, line := range strings.Split(list, "\n") {
		if strings.HasPrefix(line, a.Requirement+" ") {
			req = line
		}
	}
	return grounded(a.Evidence, TestView{Doc: user}, &Requirement{Statement: req})
}

func probeRates(calls []CallLine, base map[string]CallLine, gold map[string]string) ([]ProbeRate, float64) {
	type tally struct{ changed, asked, skipped, errors int }
	order := []string{ProbeDeletion, ProbeControl, ProbeReconstruction, ProbeOrder, ProbeRepeat}
	all := map[string]*tally{}
	correct := map[string]*tally{}
	for _, p := range order {
		all[p], correct[p] = &tally{}, &tally{}
	}
	var overlap float64
	var overlapN int
	for _, cl := range calls {
		t, ok := all[cl.Probe]
		if !ok {
			continue
		}
		b := base[cl.Test]
		basePick := cl.BasePick
		if basePick == "" && b.Reply != nil {
			basePick = b.Reply.Answer.Requirement
		}
		onCorrect := gold[cl.Test] != "" && basePick == gold[cl.Test]
		switch {
		case cl.Note != "":
			t.skipped++
			if onCorrect {
				correct[cl.Probe].skipped++
			}
			continue
		case cl.Reply == nil || cl.Error != "":
			t.errors++
			continue
		}
		changed := cl.Reply.Answer.Requirement != basePick
		if cl.Probe == ProbeRepeat {
			changed = b.Reply == nil || cl.Reply.Raw != b.Reply.Raw
		}
		if cl.Probe == ProbeOrder && b.Reply != nil {
			overlap += jaccard(b.Reply.Answer.Evidence, cl.Reply.Answer.Evidence)
			overlapN++
		}
		t.asked++
		if changed {
			t.changed++
		}
		if onCorrect {
			correct[cl.Probe].asked++
			if changed {
				correct[cl.Probe].changed++
			}
		}
	}
	var out []ProbeRate
	for _, p := range order {
		t := all[p]
		out = append(out, ProbeRate{Probe: p, Changed: rate(t.changed, t.asked), Skipped: t.skipped, Errors: t.errors})
	}
	for _, p := range []string{ProbeDeletion, ProbeControl} {
		t := correct[p]
		out = append(out, ProbeRate{Probe: p + ", correct links only", Changed: rate(t.changed, t.asked), Skipped: t.skipped})
	}
	if overlapN > 0 {
		overlap /= float64(overlapN)
	}
	return out, math.Round(overlap*100) / 100
}

func timings(calls []CallLine) []ProbeTiming {
	byProbe := map[string]*ProbeTiming{}
	var order []string
	for _, cl := range calls {
		if cl.Reply == nil {
			continue
		}
		pt, ok := byProbe[cl.Probe]
		if !ok {
			pt = &ProbeTiming{Probe: cl.Probe}
			byProbe[cl.Probe] = pt
			order = append(order, cl.Probe)
		}
		pt.Questions++
		pt.TotalSeconds += cl.Seconds
		pt.MeanPrompt += float64(cl.Reply.Timing.PromptTokens)
		if cl.Reply.Timing.PromptTokens > pt.MaxPrompt {
			pt.MaxPrompt = cl.Reply.Timing.PromptTokens
		}
	}
	var out []ProbeTiming
	for _, p := range order {
		pt := byProbe[p]
		pt.MeanSeconds = pt.TotalSeconds / float64(pt.Questions)
		pt.MeanPrompt /= float64(pt.Questions)
		out = append(out, *pt)
	}
	return out
}

var probeMeaning = map[string]string{
	ProbeDeletion:       "the cited words deleted from the test and the picked requirement",
	ProbeControl:        "as many other words deleted at random",
	ProbeReconstruction: "the test replaced by the cited words alone",
	ProbeOrder:          "the requirements listed in a shuffled order",
	ProbeRepeat:         "the same question asked again (changed means the reply's text differs)",
}

// Markdown is the report as a page a person can read. It depends on the
// results file alone, so a report rebuilt from the file is the same page.
func (s Summary) Markdown() string {
	var b strings.Builder
	h := s.Header
	b.WriteString("# Language model experiment: report\n\n")
	fmt.Fprintf(&b, "Run started %s on commit `%s`. Language model `%s` (digest `%s`, %s, %s) on Ollama %s.\n",
		orUnknown(h.Started), orUnknown(h.Commit), h.Settings.Model, orUnknown(h.Server.Digest),
		orUnknown(h.Server.ParameterSize), orUnknown(h.Server.Quantization), orUnknown(h.Server.Version))
	fmt.Fprintf(&b, "Temperature %g, seed %d, context %d tokens. %d tests, %d of them carrying a key, and %d requirements.",
		h.Settings.Temperature, h.Settings.Seed, h.Settings.NumCtx, s.Tests, s.Keyed, h.Requirements)
	if h.Quick {
		b.WriteString(" A quick run.")
	}
	b.WriteString("\n\n## Resolvers on the tests that carry a key\n\n")
	b.WriteString("The keys were hidden from the word overlap baseline and the language model. The key rule reads the names as written.\n\n")
	b.WriteString("| Resolver | Proposed | Correct | Precision (95% interval) | Recall (95% interval) |\n|---|---|---|---|---|\n")
	for _, r := range s.Resolvers {
		fmt.Fprintf(&b, "| %s | %d | %d | %s | %s |\n", r.Resolver, r.Proposed, r.Correct, r.Precision, r.Recall)
	}
	if s.BaseErrors > 0 {
		fmt.Fprintf(&b, "\n%d base questions ended in an error and count as no proposal.\n", s.BaseErrors)
	}
	b.WriteString("\n## The explanation tests\n\n")
	b.WriteString("Each probe asks again about a test the language model linked, with one thing changed, and counts how often the answer changed.\n\n")
	b.WriteString("| Probe | What changed in the question | Asked | Answer changed (95% interval) | Skipped | Errors |\n|---|---|---|---|---|---|\n")
	for _, p := range s.Probes {
		meaning := probeMeaning[strings.TrimSuffix(p.Probe, ", correct links only")]
		fmt.Fprintf(&b, "| %s | %s | %d | %s | %d | %d |\n", p.Probe, meaning, p.Changed.N, p.Changed, p.Skipped, p.Errors)
	}
	fmt.Fprintf(&b, "\nEvidence found in what the language model was shown: %d of %d pieces, %s.\n", s.Grounded.K, s.Grounded.N, s.Grounded)
	fmt.Fprintf(&b, "Mean overlap of the cited words before and after the shuffle (Jaccard): %.2f.\n", s.OrderOverlap)
	b.WriteString("\n## Links proposed for tests with no key\n\n")
	b.WriteString("Nobody recorded an answer for these tests, so these are for a person to judge and count in no score.\n\n")
	if len(s.Proposals) == 0 {
		b.WriteString("None.\n")
	} else {
		b.WriteString("| Test | Resolver | Proposed | Reason given |\n|---|---|---|---|\n")
		for _, p := range s.Proposals {
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", p.Test, p.Resolver, p.Pick, strings.ReplaceAll(p.Reason, "|", "/"))
		}
	}
	b.WriteString("\n## Time\n\n| Probe | Questions | Mean seconds | Mean prompt tokens | Largest prompt |\n|---|---|---|---|---|\n")
	var total float64
	for _, t := range s.Timing {
		total += t.TotalSeconds
		fmt.Fprintf(&b, "| %s | %d | %.1f | %.0f | %d |\n", t.Probe, t.Questions, t.MeanSeconds, t.MeanPrompt, t.MaxPrompt)
	}
	fmt.Fprintf(&b, "\nTotal time spent on questions: %.0f minutes.\n", total/60)
	return b.String()
}

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}
