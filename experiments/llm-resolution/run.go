package main

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Runner puts every browse of a run to the server, in a fixed order, and
// writes each reply as it arrives: the tests, then the incident five ways,
// then the probes on the sample of linked tests, then the repeats.
type Runner struct {
	Corpus   Corpus
	Tests    []Test
	Wiki     *Wiki
	Code     *CodeBase
	Key      Key
	Client   *Client
	Out      lineWriter
	Log      io.Writer
	Seed     int64
	Quick    bool
	Previous map[string]CallLine // replies of a run being resumed, by key

	keys        map[string]bool
	base        map[string]firstBrowse
	done        int
	start       time.Time
	lastSeconds float64
	lastReused  bool
}

// firstBrowse is what the probes need of a test's first browse.
type firstBrowse struct {
	answer  TestAnswer
	pick    string
	answers []string // every tool answer the browse saw
}

// job is one browse to make.
type job struct {
	probe, variant, test, gold, basePick string
	task                                 Task
	kit                                  *ToolKit
	removed                              []string
}

func (j job) prefix() string {
	if j.variant != "" {
		return j.probe + "/" + j.test + "/" + j.variant
	}
	return j.probe + "/" + j.test
}

// Run makes every browse of the run.
func (r *Runner) Run(ctx context.Context) error {
	r.start = time.Now()
	r.base = map[string]firstBrowse{}
	r.keys = map[string]bool{}
	for _, k := range r.Corpus.Keys() {
		r.keys[k] = true
	}
	reqs := r.Corpus.Requirements
	testKit := func(sc Scope) *ToolKit {
		sc.Task = TaskTest
		return NewToolKit(r.Wiki, r.Code, reqs, sc)
	}

	r.phase("the tests", len(r.Tests))
	for _, t := range r.Tests {
		res, err := r.browse(ctx, job{probe: ProbeBase, test: t.ID, gold: t.Gold, task: TestTask(t), kit: testKit(Scope{})})
		if err != nil {
			return err
		}
		fb := firstBrowse{pick: "none"}
		if res.Final.Test != nil {
			fb.answer = *res.Final.Test
			fb.pick = r.requirementAnswer(res.Final.Test)
		} else {
			fb.pick = ""
		}
		for _, s := range res.Steps {
			fb.answers = append(fb.answers, s.Answer)
		}
		r.base[t.ID] = fb
	}

	variants := IncidentVariants
	if r.Quick {
		variants = variants[:1]
	}
	r.phase("the incident", len(variants))
	for _, v := range variants {
		w, text := r.Wiki, r.Key.Report
		switch v {
		case VariantAlertOnly:
			text = r.Key.AlertOnly
		case VariantRemoved:
			w = r.Wiki.Without(Removal{Elements: []string{r.Key.RemoveElement}})
		case VariantControl:
			w = r.Wiki.Without(Removal{Links: []LinkRef{r.Key.ControlLink}})
		}
		kit := NewToolKit(w, r.Code, reqs, Scope{Task: TaskIncident})
		task := Task{Kind: TaskIncident, ID: ProbeIncident, Text: text}
		if _, err := r.browse(ctx, job{probe: ProbeIncident, variant: v, test: ProbeIncident, task: task, kit: kit}); err != nil {
			return err
		}
	}

	picks := map[string]string{}
	for id, fb := range r.base {
		picks[id] = fb.pick
	}
	sample := probeSample(r.Tests, picks)
	weights := NewBaseline(reqs)
	r.phase("the probes", len(sample)*len(TestProbes))
	for _, t := range sample {
		fb := r.base[t.ID]
		for _, p := range TestProbes {
			var spec ProbeSpec
			switch p {
			case ProbeReconstruction:
				spec = ReconstructionProbe(t, fb.answer)
			case ProbeDeletion:
				spec = DeletionProbe(t, fb.answer, fb.answers)
			case ProbeControl:
				spec = ControlProbe(t, fb.answer, fb.answers, r.Seed)
			case ProbeRareShared:
				spec = RareSharedProbe(t, fb.answer, fb.answers, reqs, weights)
			}
			j := job{probe: p, test: t.ID, gold: t.Gold, basePick: fb.pick, task: spec.Task, removed: spec.Removed}
			if spec.Note != "" {
				if err := r.record(CallLine{Key: j.prefix() + "/final", Probe: p, Test: t.ID, Gold: t.Gold, BasePick: fb.pick,
					Final: true, Note: spec.Note, Removed: spec.Removed}); err != nil {
					return err
				}
				continue
			}
			j.kit = testKit(Scope{Redact: spec.Redact, HideTest: spec.HideTest})
			if _, err := r.browse(ctx, j); err != nil {
				return err
			}
		}
	}

	repeats := repeatSample(r.Tests)
	r.phase("the repeats", len(repeats))
	for _, t := range repeats {
		j := job{probe: ProbeRepeat, test: t.ID, gold: t.Gold, basePick: r.base[t.ID].pick, task: TestTask(t), kit: testKit(Scope{})}
		if _, err := r.browse(ctx, j); err != nil {
			return err
		}
	}
	fmt.Fprintf(r.Log, "-- done: %d questions in %s\n", r.done, time.Since(r.start).Round(time.Second))
	return nil
}

func (r *Runner) phase(name string, n int) {
	fmt.Fprintf(r.Log, "-- %s: %d browses (elapsed %s)\n", name, n, time.Since(r.start).Round(time.Second))
}

// browse makes one browse, asking the server only what the run being
// resumed didn't answer, and writing each reply as it arrives.
func (r *Runner) browse(ctx context.Context, j job) (BrowseResult, error) {
	prefix := j.prefix()
	line := func(q Question) CallLine {
		return CallLine{Key: prefix + "/" + q.Key, Probe: j.probe, Variant: j.variant, Test: j.test, Gold: j.gold, BasePick: j.basePick, User: q.User}
	}
	ask := func(ctx context.Context, q Question) (Reply, error) {
		if prev, ok := r.Previous[prefix+"/"+q.Key]; ok && prev.Reply != nil && prev.User == q.User {
			r.lastSeconds, r.lastReused = prev.Seconds, true
			return *prev.Reply, nil
		}
		began := time.Now()
		reply, err := r.Client.Chat(ctx, q.System, q.User, q.Format)
		r.lastSeconds, r.lastReused = time.Since(began).Seconds(), false
		if err != nil && ctx.Err() == nil {
			failed := line(q)
			failed.Error, failed.Seconds, failed.Final = err.Error(), r.lastSeconds, q.Final
			_ = r.record(failed)
		}
		return reply, err
	}
	observe := func(q Question, reply Reply, step *StepRecord, whole *BrowseResult) error {
		cl := line(q)
		cl.Reply, cl.Seconds, cl.Reused = &reply, r.lastSeconds, r.lastReused
		if step != nil {
			cl.Step, _ = strconv.Atoi(q.Key)
			cl.Error = step.Error
			cl.Tool, cl.Arg, cl.Arg2, cl.ToolAnswer = step.Step.Tool, step.Step.Arg, step.Step.Arg2, step.Answer
			return r.record(cl)
		}
		cl.Final, cl.Task, cl.Removed, cl.Steps = true, j.task.Text, j.removed, len(whole.Steps)
		cl.Error = whole.Final.Error
		view := whole.View
		cl.View = &view
		name := "answer"
		if j.task.Kind == TaskIncident {
			name = "incident"
		}
		cl.SysML = view.SysML(j.kit.Wiki(), name)
		if a := whole.Final.Test; a != nil {
			cl.Answer = a
			cl.Pick = r.requirementAnswer(a)
			cl.Facts, cl.BaseRate = r.linkFacts(a, j.gold)
			cl.Citations = CheckTestCitations(*a, j.kit)
		}
		if a := whole.Final.Incident; a != nil {
			cl.Account = a
			cl.Items = ScoreIncident(r.Key, *a, r.Wiki)
			cl.Citations = CheckCitations(*a, j.kit)
		}
		return r.record(cl)
	}
	return browse(ctx, j.task, j.kit, ask, observe)
}

// requirementAnswer is the answer a test's final reply gives as a
// requirement link: the first link to a system requirement, or none.
func (r *Runner) requirementAnswer(a *TestAnswer) string {
	for _, l := range a.Links {
		if e, ok := r.Wiki.Get(l.ID); ok && r.keys[e.Short] {
			return e.Short
		}
		if r.keys[strings.TrimSpace(l.ID)] {
			return strings.TrimSpace(l.ID)
		}
	}
	return "none"
}

// linkFacts places every link of an answer in the whole systems model: its
// kind, whether it is a system requirement, and whether it lies within three
// links of the recorded requirement. The base rate is the share of every
// element that lies that near.
func (r *Runner) linkFacts(a *TestAnswer, gold string) ([]LinkFact, float64) {
	var near map[*Element]bool
	rate := 0.0
	if g, ok := r.Wiki.Get(gold); ok && gold != "" {
		near = r.Wiki.near(g)
		rate = float64(len(near)) / float64(len(r.Wiki.Elements)-1)
	}
	var out []LinkFact
	for _, l := range a.Links {
		f := LinkFact{ID: l.ID, Kind: "unknown"}
		if e, ok := r.Wiki.Get(l.ID); ok {
			f.ID, f.Kind = e.ID, e.Kind
			f.Requirement = r.keys[e.Short]
			f.Near = near[e]
		}
		out = append(out, f)
	}
	return out, rate
}

func (r *Runner) record(cl CallLine) error {
	r.done++
	what := cl.Tool + " " + cl.Arg
	switch {
	case cl.Note != "":
		what = "skipped: " + cl.Note
	case cl.Reply == nil && cl.Error != "":
		what = "error: " + cl.Error
	case cl.Final && cl.Answer != nil:
		what = "final: " + cl.Pick
	case cl.Final && cl.Account != nil:
		found := 0
		for _, it := range cl.Items {
			if it.Found {
				found++
			}
		}
		what = fmt.Sprintf("final: %d of %d key items", found, len(cl.Items))
	case cl.Error != "":
		what = "not understood"
	}
	if cl.Reused {
		what += " (reused)"
	}
	fmt.Fprintf(r.Log, "%5d %6.1fs %-14s %-50s %s\n", r.done, cl.Seconds, cl.Probe, trimTo(cl.Test+" "+cl.Variant, 50), clip(what, 70))
	return r.Out.write(cl)
}

func trimTo(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n+1:]
}

// previousCalls loads the replies of an earlier run, if it read the same
// tests and systems model with the same settings.
func previousCalls(path, corpusHash, modelHash string, s Settings) (map[string]CallLine, error) {
	c, err := ReadResults(path)
	if err != nil {
		return nil, err
	}
	if c.Header.CorpusHash != corpusHash {
		return nil, fmt.Errorf("%s was made from a different state of the tests (corpus %s, now %s)", path, c.Header.CorpusHash, corpusHash)
	}
	if c.Header.ModelHash != modelHash {
		return nil, fmt.Errorf("%s was made from a different systems model (%s, now %s)", path, c.Header.ModelHash, modelHash)
	}
	if c.Header.Settings != s {
		return nil, fmt.Errorf("%s was made with different settings: %+v", path, c.Header.Settings)
	}
	out := map[string]CallLine{}
	for _, cl := range c.Calls {
		if cl.Reply != nil {
			out[cl.Key] = cl
		}
	}
	return out, nil
}

// probeSample is the first test and every fourth after it among those the
// first browse linked to a requirement, in the order the tests are read
// (EXP-SR-05).
func probeSample(tests []Test, picks map[string]string) []Test {
	var out []Test
	n := 0
	for _, t := range tests {
		if p := picks[t.ID]; p == "" || p == "none" {
			continue
		}
		if n%4 == 0 {
			out = append(out, t)
		}
		n++
	}
	return out
}

// quickSample is the fixed sample of a quick run: eight tests that carry a
// key and four that do not, spread evenly over the sorted list (EXP-SR-14).
// A corpus with fewer gives all it has.
func quickSample(tests []Test) []Test {
	var keyed, unkeyed []Test
	for _, t := range tests {
		if t.Gold != "" {
			keyed = append(keyed, t)
		} else {
			unkeyed = append(unkeyed, t)
		}
	}
	chosen := map[string]bool{}
	for _, t := range append(spread(keyed, 8), spread(unkeyed, 4)...) {
		chosen[t.ID] = true
	}
	var out []Test
	for _, t := range tests {
		if chosen[t.ID] {
			out = append(out, t)
		}
	}
	return out
}

func spread(tests []Test, n int) []Test {
	if len(tests) <= n {
		return tests
	}
	out := make([]Test, n)
	for i := range out {
		out[i] = tests[i*len(tests)/n]
	}
	return out
}

// repeatSample is the first test and every fourth after it (EXP-SR-08).
func repeatSample(tests []Test) []Test {
	var out []Test
	for i, t := range tests {
		if i%4 == 0 {
			out = append(out, t)
		}
	}
	return out
}
