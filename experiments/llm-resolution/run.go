package main

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Runner puts the questions to the model and writes each answer as it comes.
type Runner struct {
	Corpus   Corpus
	Tests    []Test
	Client   *Client
	Out      lineWriter
	Log      io.Writer
	Seed     int64
	Probes   map[string]bool
	Previous map[string]CallLine // completed calls from a run being resumed

	calls map[string]CallLine
	done  int
	start time.Time
}

func (r *Runner) key(probe string, t Test) string { return probe + "/" + t.ID }

// ask makes one call, or reuses the answer from the run being resumed.
func (r *Runner) ask(ctx context.Context, probe string, t Test, base string, v Variant) (CallLine, error) {
	line := CallLine{Key: r.key(probe, t), Probe: probe, Test: t.ID, Gold: t.Gold, BasePick: base, Removed: v.Removed}
	if probe == ProbeDeletion || probe == ProbeControl {
		if i := findRequirement(v.Requirements, base); i >= 0 {
			edited := v.Requirements[i]
			line.EditedReq = &edited
		}
	}
	line.Shuffled = probe == ProbeOrder
	if v.Note != "" {
		line.Note = v.Note
		return line, r.record(line)
	}
	line.User = userPrompt(v.View)
	if prev, ok := r.Previous[line.Key]; ok && prev.Reply != nil && prev.Error == "" {
		prev.Reused = true
		return prev, r.record(prev)
	}
	began := time.Now()
	reply, err := r.Client.Chat(ctx, systemPrompt(v.Requirements), line.User)
	line.Seconds = time.Since(began).Seconds()
	if err != nil {
		line.Error = err.Error()
		if reply.Raw != "" {
			line.Reply = &reply
		}
	} else {
		line.Reply = &reply
	}
	if ctx.Err() != nil {
		return line, ctx.Err()
	}
	return line, r.record(line)
}

func (r *Runner) record(line CallLine) error {
	r.calls[line.Key] = line
	r.done++
	pick := "-"
	switch {
	case line.Note != "":
		pick = "skipped: " + line.Note
	case line.Error != "":
		pick = "error: " + line.Error
	case line.Reply != nil:
		pick = line.Reply.Answer.Requirement
	}
	fmt.Fprintf(r.Log, "%5d  %6.1fs  %-14s %-60s %s\n", r.done, line.Seconds, line.Probe, trimTo(line.Test, 60), pick)
	return r.Out.write(line)
}

func trimTo(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n+1:]
}

func (r *Runner) pick(t Test) (Answer, bool) {
	line, ok := r.calls[r.key(ProbeBase, t)]
	if !ok || line.Reply == nil || line.Error != "" {
		return Answer{}, false
	}
	return line.Reply.Answer, true
}

// Run asks the base question for every test, then the probes for every test
// the model linked. Calls that share a prompt prefix run together, so the
// server's prompt cache does most of the reading.
func (r *Runner) Run(ctx context.Context) error {
	r.calls = map[string]CallLine{}
	r.start = time.Now()
	reqs := r.Corpus.Requirements
	phase := func(name string, n int) {
		fmt.Fprintf(r.Log, "-- %s: %d questions (elapsed %s)\n", name, n, time.Since(r.start).Round(time.Second))
	}

	phase(ProbeBase, len(r.Tests))
	for _, t := range r.Tests {
		if _, err := r.ask(ctx, ProbeBase, t, "", Variant{Requirements: reqs, View: viewOf(t)}); err != nil {
			return err
		}
	}

	var linked []Test
	for _, t := range r.Tests {
		if a, ok := r.pick(t); ok && a.Requirement != "none" {
			linked = append(linked, t)
		}
	}
	type probeStep struct {
		name  string
		tests []Test
		make  func(Test, Answer) Variant
	}
	repeat := repeatSample(r.Tests)
	steps := []probeStep{
		{ProbeReconstruction, linked, func(_ Test, a Answer) Variant { return reconstructionVariant(reqs, a) }},
		{ProbeRepeat, repeat, func(t Test, _ Answer) Variant { return Variant{Requirements: reqs, View: viewOf(t)} }},
		{ProbeDeletion, linked, func(t Test, a Answer) Variant { return deletionVariant(reqs, t, a) }},
		{ProbeControl, linked, func(t Test, a Answer) Variant { return controlVariant(reqs, t, a, r.Seed) }},
		{ProbeOrder, linked, func(t Test, _ Answer) Variant {
			v := orderVariant(reqs, r.Seed)
			v.View = viewOf(t)
			return v
		}},
	}
	for _, s := range steps {
		if !r.Probes[s.name] {
			continue
		}
		phase(s.name, len(s.tests))
		for _, t := range s.tests {
			a, ok := r.pick(t)
			if !ok {
				continue
			}
			if _, err := r.ask(ctx, s.name, t, a.Requirement, s.make(t, a)); err != nil {
				return err
			}
		}
	}
	fmt.Fprintf(r.Log, "-- done: %d questions in %s\n", r.done, time.Since(r.start).Round(time.Second))
	return nil
}

// Calls returns every call made or skipped, in the order they were recorded.
func (r *Runner) Calls() []CallLine {
	out := make([]CallLine, 0, len(r.calls))
	for _, t := range r.Tests {
		for _, p := range AllProbes {
			if c, ok := r.calls[p+"/"+t.ID]; ok {
				out = append(out, c)
			}
		}
	}
	return out
}

// previousCalls loads the completed calls of an earlier run, if it read the
// same corpus with the same settings.
func previousCalls(path string, hash string, s Settings) (map[string]CallLine, error) {
	c, err := ReadResults(path)
	if err != nil {
		return nil, err
	}
	if c.Header.CorpusHash != hash {
		return nil, fmt.Errorf("%s was made from a different state of the repository (corpus %s, now %s)", path, c.Header.CorpusHash, hash)
	}
	if c.Header.Settings != s {
		return nil, fmt.Errorf("%s was made with different settings: %+v", path, c.Header.Settings)
	}
	out := map[string]CallLine{}
	for _, cl := range c.Calls {
		if cl.Reply != nil && cl.Error == "" {
			out[cl.Key] = cl
		}
	}
	return out, nil
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
