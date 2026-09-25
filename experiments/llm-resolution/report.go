package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
)

// The report is computed from the results file alone, so a report rebuilt
// from a file is the report the run wrote (EXP-SR-11, EXP-SR-15).

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

// mcnemarExact is the two-sided exact McNemar test on the discordant pairs:
// b tests only one resolver got right, c only the other. It is the binomial
// test of b against b+c at one half.
func mcnemarExact(b, c int) float64 {
	n := b + c
	if n == 0 {
		return 1
	}
	k := min(b, c)
	var tail float64
	for i := 0; i <= k; i++ {
		tail += math.Exp(lchoose(n, i) - float64(n)*math.Ln2)
	}
	return math.Min(1, 2*tail)
}

func lchoose(n, k int) float64 {
	a, _ := math.Lgamma(float64(n + 1))
	b, _ := math.Lgamma(float64(k + 1))
	c, _ := math.Lgamma(float64(n - k + 1))
	return a - b - c
}

// ResolverScore is one resolver's result on the tests that carry a key.
type ResolverScore struct {
	Resolver  string `json:"resolver"`
	Proposed  int    `json:"proposed"`
	Correct   int    `json:"correct"`
	Precision Rate   `json:"precision"`
	Recall    Rate   `json:"recall"`
}

// ProbeRate is how often a probe changed the answer.
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

// Proposal is a link proposed for a test that carries no key. Nobody recorded
// an answer for these, so they are listed for a person to judge.
type Proposal struct {
	Test     string `json:"test"`
	Resolver string `json:"resolver"`
	Pick     string `json:"pick"`
	Reason   string `json:"reason"`
}

// BlindItem is one proposal with nothing to say who made it or why.
type BlindItem struct {
	Number int    `json:"number"`
	Test   string `json:"test"`
	Pick   string `json:"pick"`
}

// KindCount counts the links to elements of one kind that aren't system
// requirements, and how many of them lie near the recorded requirement.
type KindCount struct {
	Kind  string `json:"kind"`
	Links int    `json:"links"`
	Near  int    `json:"near"`
}

// MismatchItem is one suspicion, with where it was raised.
type MismatchItem struct {
	Where string `json:"where"`
	Mismatch
}

// IncidentResult is one browse of the incident.
type IncidentResult struct {
	Variant   string      `json:"variant"`
	Found     int         `json:"found"`
	Items     []KeyResult `json:"items"`
	Cited     int         `json:"cited"`
	Missing   []Citation  `json:"missing"`
	Mechanism []string    `json:"mechanism"`
	Why       string      `json:"why"`
	Steps     int         `json:"steps"`
	Seconds   float64     `json:"seconds"`
	Error     string      `json:"error,omitempty"`
	SysML     string      `json:"view_sysml"`
}

// ProbeTiming is what the questions of one probe cost.
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
	Header     RunHeader        `json:"-"`
	Tests      int              `json:"tests"`
	Keyed      int              `json:"keyed_tests"`
	Resolvers  []ResolverScore  `json:"resolvers"`
	Paired     PairedComparison `json:"paired"`
	Probes     []ProbeRate      `json:"probes"`
	Grounded   Rate             `json:"evidence_grounded"`
	OtherKinds []KindCount      `json:"other_kinds"`
	OtherNear  Rate             `json:"other_near"`
	BaseRate   float64          `json:"base_rate"`
	Incident   []IncidentResult `json:"incident"`
	Mismatches []MismatchItem   `json:"mismatches"`
	Proposals  []Proposal       `json:"proposals"`
	Blind      []BlindItem      `json:"blind"`
	Timing     []ProbeTiming    `json:"timing"`
	BaseErrors int              `json:"base_errors"`
}

// Summarise computes the report from a results file's contents alone.
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
		if cl.Final && cl.Probe == ProbeBase {
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

	// The same search's best-ranked answer, whatever its score, which is
	// also the first step every browse can take.
	best := ResolverScore{Resolver: "word overlap, best-ranked"}
	woRight := map[string]bool{}
	for _, b := range c.Baseline {
		if g := gold[b.Test]; g != "" && len(b.Top) > 0 {
			best.Proposed++
			if b.Top[0].Key == g {
				best.Correct++
				woRight[b.Test] = true
			}
		}
	}
	best.Precision, best.Recall = rate(best.Correct, best.Proposed), rate(best.Correct, s.Keyed)
	s.Resolvers = append(s.Resolvers, best)

	lm := ResolverScore{Resolver: "language model"}
	var grounded, pieces int
	kinds := map[string]*KindCount{}
	var near, links int
	var expected float64
	for _, t := range sortedKeys(base) {
		cl := base[t]
		if cl.Reply == nil || cl.Answer == nil {
			s.BaseErrors++
			continue
		}
		a := cl.Answer
		g := gold[t]
		if cl.Pick != "none" && cl.Pick != "" {
			for _, ok := range foundAsWritten(a.Evidence, cl.Task) {
				pieces++
				if ok {
					grounded++
				}
			}
		}
		if g == "" {
			for _, l := range a.Links {
				s.Proposals = append(s.Proposals, Proposal{Test: t, Resolver: lm.Resolver, Pick: l.ID, Reason: a.Reason})
			}
			continue
		}
		if cl.Pick != "none" && cl.Pick != "" {
			lm.Proposed++
			if cl.Pick == g {
				lm.Correct++
			}
		}
		for _, f := range cl.Facts {
			if f.Requirement {
				continue
			}
			k, ok := kinds[f.Kind]
			if !ok {
				k = &KindCount{Kind: f.Kind}
				kinds[f.Kind] = k
			}
			k.Links++
			links++
			expected += cl.BaseRate
			if f.Near {
				k.Near++
				near++
			}
		}
	}
	lm.Precision, lm.Recall = rate(lm.Correct, lm.Proposed), rate(lm.Correct, s.Keyed)
	s.Resolvers = append(s.Resolvers, lm)
	s.Grounded = rate(grounded, pieces)
	for _, k := range kinds {
		s.OtherKinds = append(s.OtherKinds, *k)
	}
	sort.Slice(s.OtherKinds, func(i, j int) bool {
		if s.OtherKinds[i].Links != s.OtherKinds[j].Links {
			return s.OtherKinds[i].Links > s.OtherKinds[j].Links
		}
		return s.OtherKinds[i].Kind < s.OtherKinds[j].Kind
	})
	s.OtherNear = rate(near, links)
	if links > 0 {
		s.BaseRate = expected / float64(links)
	}

	for t, g := range gold {
		lmRight := false
		if cl, ok := base[t]; ok {
			lmRight = cl.Pick == g
		}
		switch {
		case lmRight && woRight[t]:
			s.Paired.BothRight++
		case lmRight:
			s.Paired.OnlyLanguageModel++
		case woRight[t]:
			s.Paired.OnlyWordOverlap++
		default:
			s.Paired.NeitherRight++
		}
	}
	s.Paired.McNemarP = mcnemarExact(s.Paired.OnlyLanguageModel, s.Paired.OnlyWordOverlap)

	s.Probes = probeRates(c.Calls, base, gold)
	s.Incident = incidentResults(c.Calls)
	for _, t := range sortedKeys(base) {
		if a := base[t].Answer; a != nil {
			for _, m := range a.Mismatches {
				s.Mismatches = append(s.Mismatches, MismatchItem{Where: t, Mismatch: m})
			}
		}
	}
	for _, cl := range c.Calls {
		if cl.Final && cl.Probe == ProbeIncident && cl.Account != nil {
			for _, m := range cl.Account.Mismatches {
				s.Mismatches = append(s.Mismatches, MismatchItem{Where: "the incident, " + cl.Variant, Mismatch: m})
			}
		}
	}
	s.Timing = timings(c.Calls)
	sort.SliceStable(s.Proposals, func(i, j int) bool {
		if s.Proposals[i].Test != s.Proposals[j].Test {
			return s.Proposals[i].Test < s.Proposals[j].Test
		}
		return s.Proposals[i].Resolver > s.Proposals[j].Resolver
	})
	s.Blind = blindList(s.Proposals)
	return s
}

// blindList gives each distinct proposal once, in an order that depends on
// nothing a judge could read anything into, with no resolver and no reason.
func blindList(ps []Proposal) []BlindItem {
	seen := map[string]bool{}
	var out []BlindItem
	for _, p := range ps {
		if key := p.Test + "\x00" + p.Pick; !seen[key] {
			seen[key] = true
			out = append(out, BlindItem{Test: p.Test, Pick: p.Pick})
		}
	}
	h := func(b BlindItem) uint64 {
		f := fnv.New64a()
		_, _ = f.Write([]byte(b.Test + "\x00" + b.Pick))
		return f.Sum64()
	}
	sort.Slice(out, func(i, j int) bool { return h(out[i]) < h(out[j]) })
	for i := range out {
		out[i].Number = i + 1
	}
	return out
}

func sortedKeys(m map[string]CallLine) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// foundAsWritten checks each piece of evidence against the test's fields as
// the task showed them. Only the text as written counts, apart from case.
// Words found one by one, or in the task's fixed wording, or only in what the
// tools showed, do not.
func foundAsWritten(evidence []string, task string) []bool {
	text := strings.ToLower(strings.Join(taskFields(task), "\n"))
	out := make([]bool, len(evidence))
	for i, e := range evidence {
		e = strings.ToLower(strings.TrimSpace(e))
		out[i] = e != "" && strings.Contains(text, e)
	}
	return out
}

func probeRates(calls []CallLine, base map[string]CallLine, gold map[string]string) []ProbeRate {
	type tally struct{ changed, toNone, toOther, asked, skipped, errors int }
	order := []string{ProbeDeletion, ProbeControl, ProbeRareShared, ProbeReconstruction, ProbeRepeat}
	all := map[string]*tally{}
	correct := map[string]*tally{}
	for _, p := range order {
		all[p], correct[p] = &tally{}, &tally{}
	}
	for _, cl := range calls {
		t, ok := all[cl.Probe]
		if !ok || !cl.Final {
			continue
		}
		b := base[cl.Test]
		basePick := cl.BasePick
		if basePick == "" {
			basePick = b.Pick
		}
		onCorrect := gold[cl.Test] != "" && basePick == gold[cl.Test]
		switch {
		case cl.Note != "":
			t.skipped++
			if onCorrect {
				correct[cl.Probe].skipped++
			}
			continue
		case cl.Reply == nil || cl.Answer == nil:
			t.errors++
			continue
		}
		changed := cl.Pick != basePick
		if cl.Probe == ProbeRepeat {
			changed = b.Reply == nil || cl.Reply.Raw != b.Reply.Raw
		}
		count := func(t *tally) {
			t.asked++
			if !changed {
				return
			}
			t.changed++
			if cl.Probe == ProbeRepeat {
				return
			}
			if cl.Pick == "none" || cl.Pick == "" {
				t.toNone++
			} else {
				t.toOther++
			}
		}
		count(t)
		if onCorrect {
			count(correct[cl.Probe])
		}
	}
	var out []ProbeRate
	for _, p := range order {
		t := all[p]
		out = append(out, ProbeRate{Probe: p, Changed: rate(t.changed, t.asked), ToNone: t.toNone, ToOther: t.toOther, Skipped: t.skipped, Errors: t.errors})
	}
	for _, p := range order[:4] {
		t := correct[p]
		out = append(out, ProbeRate{Probe: p + ", correct links only", Changed: rate(t.changed, t.asked), ToNone: t.toNone, ToOther: t.toOther, Skipped: t.skipped})
	}
	return out
}

func incidentResults(calls []CallLine) []IncidentResult {
	seconds := map[string]float64{}
	for _, cl := range calls {
		if cl.Probe == ProbeIncident {
			seconds[cl.Variant] += cl.Seconds
		}
	}
	var out []IncidentResult
	for _, cl := range calls {
		if !cl.Final || cl.Probe != ProbeIncident {
			continue
		}
		r := IncidentResult{Variant: cl.Variant, Items: cl.Items, Steps: cl.Steps, Seconds: seconds[cl.Variant], Error: cl.Error, SysML: cl.SysML}
		for _, it := range cl.Items {
			if it.Found {
				r.Found++
			}
		}
		for _, c := range cl.Citations {
			r.Cited++
			if !c.Found {
				r.Missing = append(r.Missing, c)
			}
		}
		if cl.Account != nil {
			r.Mechanism = cl.Account.Mechanism
			r.Why = cl.Account.Why
		}
		out = append(out, r)
	}
	return out
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
	ProbeDeletion:       "the cited words deleted from the test and from every tool answer",
	ProbeControl:        "as many other words of the test deleted at random",
	ProbeRareShared:     "uncited words the test and its requirement share deleted, rarest first",
	ProbeReconstruction: "the test replaced by the cited words alone, its code hidden",
	ProbeRepeat:         "the same browse again (changed means the final reply's text differs)",
}

func cell(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, "|", "/"), "\n", " ") }

// Markdown is the report as a page a person can read. It depends on the
// results file alone, so a report rebuilt from the file is the same page.
func (s Summary) Markdown() string {
	var b strings.Builder
	h := s.Header
	b.WriteString("# Language model experiment: report\n\n")
	fmt.Fprintf(&b, "Run started %s on commit `%s`. Language model `%s` (digest `%s`, %s, %s) on Ollama %s.\n",
		orUnknown(h.Started), orUnknown(h.Commit), h.Settings.Model, orUnknown(h.Server.Digest),
		orUnknown(h.Server.ParameterSize), orUnknown(h.Server.Quantization), orUnknown(h.Server.Version))
	fmt.Fprintf(&b, "Temperature %g, seed %d, context %d tokens. %d tests, %d of them carrying a key, %d requirements and %d elements of the systems model. At most %d tool calls a test and %d for the incident.",
		h.Settings.Temperature, h.Settings.Seed, h.Settings.NumCtx, s.Tests, s.Keyed, h.Requirements, h.Elements, h.BudgetTest, h.BudgetIncident)
	if h.Quick {
		b.WriteString(" A quick run.")
	}
	b.WriteString("\n\nTo judge the links proposed for tests with no key without being swayed by the reasons given, start with the last list, which has none, before reading the rest.\n")

	b.WriteString("\n## The incident\n\n")
	if len(s.Incident) == 0 {
		b.WriteString("Not browsed.\n")
	} else {
		b.WriteString("| Browse | Key items found | Citations checked | Citations missing | Steps | Seconds |\n|---|---|---|---|---|---|\n")
		for _, r := range s.Incident {
			fmt.Fprintf(&b, "| %s | %d of %d | %d | %d | %d | %.0f |\n", r.Variant, r.Found, len(r.Items), r.Cited, len(r.Missing), r.Steps, r.Seconds)
		}
		for _, r := range s.Incident {
			fmt.Fprintf(&b, "\n### %s\n\n", r.Variant)
			if r.Error != "" {
				fmt.Fprintf(&b, "The final reply wasn't the JSON asked for: %s\n\n", r.Error)
			}
			if r.Why != "" {
				fmt.Fprintf(&b, "> %s\n\n", cell(r.Why))
			}
			if len(r.Mechanism) > 0 {
				fmt.Fprintf(&b, "Mechanism cited: %s.\n\n", strings.Join(r.Mechanism, ", "))
			}
			for _, it := range r.Items {
				mark := "missed"
				if it.Found {
					mark = "found"
				}
				fmt.Fprintf(&b, "- %s: %s\n", it.Item, mark)
			}
			for _, c := range r.Missing {
				fmt.Fprintf(&b, "- cited in %s but missing: `%s`, %s\n", c.Field, c.Ref, c.Why)
			}
			if r.SysML != "" {
				fmt.Fprintf(&b, "\nThe view it ended with:\n\n```\n%s```\n", r.SysML)
			}
		}
	}

	b.WriteString("\n## Resolvers on the tests that carry a key\n\n")
	b.WriteString("The keys were hidden from the word overlap baseline and the language model. The key rule reads the names as written.\n\n")
	b.WriteString("| Resolver | Proposed | Correct | Precision (95% interval) | Recall (95% interval) |\n|---|---|---|---|---|\n")
	for _, r := range s.Resolvers {
		fmt.Fprintf(&b, "| %s | %d | %d | %s | %s |\n", r.Resolver, r.Proposed, r.Correct, r.Precision, r.Recall)
	}
	if s.BaseErrors > 0 {
		fmt.Fprintf(&b, "\n%d browses ended without a final answer and count as no proposal.\n", s.BaseErrors)
	}
	p := s.Paired
	b.WriteString("\nTest by test, the language model against the word overlap's best-ranked answer:\n\n")
	b.WriteString("| Both right | Only the language model | Only the word overlap | Neither |\n|---|---|---|---|\n")
	fmt.Fprintf(&b, "| %d | %d | %d | %d |\n\nMcNemar's exact test on the %d tests only one got right: p = %.3f.\n",
		p.BothRight, p.OnlyLanguageModel, p.OnlyWordOverlap, p.NeitherRight, p.OnlyLanguageModel+p.OnlyWordOverlap, p.McNemarP)

	b.WriteString("\n## Links to elements of other kinds\n\n")
	if s.OtherNear.N == 0 {
		b.WriteString("None, on the tests that carry a key.\n")
	} else {
		b.WriteString("| Kind | Links | Within three links of the recorded requirement |\n|---|---|---|\n")
		for _, k := range s.OtherKinds {
			fmt.Fprintf(&b, "| %s | %d | %d |\n", k.Kind, k.Links, k.Near)
		}
		fmt.Fprintf(&b, "\n%d of %d such links lie within three links of the recorded requirement, %s. Of every element, the share that lies as near, taken over the same links, is %.2f.\n",
			s.OtherNear.K, s.OtherNear.N, s.OtherNear, s.BaseRate)
	}

	b.WriteString("\n## The explanation tests\n\n")
	b.WriteString("Each probe browses again for a test the language model linked, with one thing changed, and counts how often the answer changed.\n\n")
	b.WriteString("| Probe | What changed | Asked | Answer changed (95% interval) | to none | to another | Skipped | Errors |\n|---|---|---|---|---|---|---|---|\n")
	for _, p := range s.Probes {
		meaning := probeMeaning[strings.TrimSuffix(p.Probe, ", correct links only")]
		fmt.Fprintf(&b, "| %s | %s | %d | %s | %d | %d | %d | %d |\n", p.Probe, meaning, p.Changed.N, p.Changed, p.ToNone, p.ToOther, p.Skipped, p.Errors)
	}
	fmt.Fprintf(&b, "\nEvidence found as written, apart from case, in the test's fields: %d of %d pieces, %s.\n", s.Grounded.K, s.Grounded.N, s.Grounded)

	b.WriteString("\n## Suspected mismatches\n\n")
	if len(s.Mismatches) == 0 {
		b.WriteString("None.\n")
	} else {
		b.WriteString("| Raised in | Element | The systems model says | The system shows | Why |\n|---|---|---|---|---|\n")
		for _, m := range s.Mismatches {
			fmt.Fprintf(&b, "| %s | `%s` | %s | %s | %s |\n", cell(m.Where), cell(m.ID), cell(m.ModelSays), cell(m.SystemShows), cell(m.Why))
		}
	}
	if len(h.Key.KnownMismatches) > 0 {
		b.WriteString("\nKnown before the run:\n\n")
		for _, m := range h.Key.KnownMismatches {
			fmt.Fprintf(&b, "- `%s`: %s\n", m.ID, m.Note)
		}
	}

	b.WriteString("\n## Links proposed for tests with no key\n\n")
	b.WriteString("Nobody recorded an answer for these tests, so these are for a person to judge and count in no score.\n\n")
	if len(s.Proposals) == 0 {
		b.WriteString("None.\n")
	} else {
		b.WriteString("| Test | Resolver | Proposed | Reason given |\n|---|---|---|---|\n")
		for _, p := range s.Proposals {
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", p.Test, p.Resolver, p.Pick, cell(p.Reason))
		}
	}
	b.WriteString("\n## For blind judgement\n\n")
	b.WriteString("The same proposals again, each once, in a scrambled order and with nothing to say where they came from. Judge these first, then compare with the reasons further up.\n\n")
	if len(s.Blind) == 0 {
		b.WriteString("None.\n")
	} else {
		b.WriteString("| Number | Test | Proposed | Right? |\n|---|---|---|---|\n")
		for _, x := range s.Blind {
			fmt.Fprintf(&b, "| %d | `%s` | %s | |\n", x.Number, x.Test, x.Pick)
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
