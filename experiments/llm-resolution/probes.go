package main

import (
	"hash/fnv"
	"sort"
	"strings"
)

// The probes browse a test again with one thing changed, and compare the
// answer with the first. The words the language model cited as its evidence
// are deleted from the test and from everything the tools show; two controls
// delete as many other words instead; the rebuild shows it the cited words
// alone; the repeat asks again unchanged. The incident has its own probes,
// which take an element or a link out of the systems model.

// The probes.
const (
	ProbeBase           = "base"           // the test as it is
	ProbeDeletion       = "deletion"       // the cited words, deleted everywhere
	ProbeControl        = "control"        // as many other words, deleted at random
	ProbeRareShared     = "rare-shared"    // as many shared uncited words, rarest first
	ProbeReconstruction = "reconstruction" // the test replaced by the cited words alone
	ProbeRepeat         = "repeat"         // the test browsed again, unchanged
)

// TestProbes are the probes on a test of the sample, in the order a run
// makes them.
var TestProbes = []string{ProbeReconstruction, ProbeDeletion, ProbeControl, ProbeRareShared}

// ProbeSpec is one probe's browse: its task, and what its tools hide.
type ProbeSpec struct {
	Task     Task
	Redact   map[string]bool
	HideTest string
	Removed  []string
	Note     string // why the probe asks nothing, when it doesn't
}

// evidenceTerms are the normalised words of the evidence.
func evidenceTerms(ev []string) map[string]bool {
	out := map[string]bool{}
	for _, e := range ev {
		for _, t := range terms(e) {
			out[t] = true
		}
	}
	return out
}

// fieldTerms are the words of a test's fields, each once, in order.
func fieldTerms(t Test) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range []string{t.Name, t.Package, t.File, t.Doc} {
		for _, w := range terms(f) {
			if !seen[w] {
				seen[w] = true
				out = append(out, w)
			}
		}
	}
	return out
}

// occurring are the cited words found in the test or in any tool answer of
// its first browse.
func occurring(t Test, cited map[string]bool, answers []string) map[string]bool {
	out := map[string]bool{}
	texts := append([]string{t.Name, t.Package, t.File, t.Doc}, answers...)
	for _, s := range texts {
		for _, w := range terms(s) {
			if cited[w] {
				out[w] = true
			}
		}
	}
	return out
}

// withoutWords is the test's task with the words deleted from its fields.
func withoutWords(t Test, drop map[string]bool) Task {
	v := viewOf(t)
	for _, f := range []*string{&v.Name, &v.Package, &v.File, &v.Doc} {
		*f, _ = removeTerms(*f, drop)
	}
	return Task{Kind: TaskTest, ID: t.ID, Text: taskText(v)}
}

func sortedWords(m map[string]bool) []string {
	var out []string
	for w := range m {
		out = append(out, w)
	}
	sort.Strings(out)
	return out
}

const nothingDeleted = "the deletion deleted nothing, so there is nothing to match"

// DeletionProbe deletes every cited word from the test and from every answer
// the tools give in the probe's browse.
func DeletionProbe(t Test, base TestAnswer, answers []string) ProbeSpec {
	cited := evidenceTerms(base.Evidence)
	spec := ProbeSpec{Redact: cited, Removed: sortedWords(cited)}
	if len(occurring(t, cited, answers)) == 0 {
		spec.Note = "none of the cited words occur in the test or in what the tools showed"
		return spec
	}
	spec.Task = withoutWords(t, cited)
	return spec
}

// ControlProbe deletes as many words as the deletion did, chosen from the
// test's uncited words with a fixed seed.
func ControlProbe(t Test, base TestAnswer, answers []string, seed int64) ProbeSpec {
	cited := evidenceTerms(base.Evidence)
	n := len(occurring(t, cited, answers))
	if n == 0 {
		return ProbeSpec{Note: nothingDeleted}
	}
	var pool []string
	for _, w := range fieldTerms(t) {
		if !cited[w] {
			pool = append(pool, w)
		}
	}
	if len(pool) == 0 {
		return ProbeSpec{Note: "the test has no uncited word to delete"}
	}
	chosen := map[string]bool{}
	for _, w := range shuffled(pool, seed^int64(hashString(t.ID)))[:min(n, len(pool))] {
		chosen[w] = true
	}
	return ProbeSpec{Task: withoutWords(t, chosen), Redact: chosen, Removed: sortedWords(chosen)}
}

// RareSharedProbe is the second control. The cited words are often the most
// telling words a test and its requirement share, so beating random words
// could mean only that. This control deletes the uncited words the two share,
// the rarest first, until at least as many are gone as the deletion took.
func RareSharedProbe(t Test, base TestAnswer, answers []string, reqs []Requirement, b *Baseline) ProbeSpec {
	cited := evidenceTerms(base.Evidence)
	n := len(occurring(t, cited, answers))
	if n == 0 {
		return ProbeSpec{Note: nothingDeleted}
	}
	var req *Requirement
	for _, l := range base.Links {
		if i := findRequirement(reqs, l.ID); i >= 0 {
			req = &reqs[i]
			break
		}
	}
	if req == nil {
		return ProbeSpec{Note: "the answer links no requirement"}
	}
	inReq := map[string]bool{}
	for _, w := range terms(req.Name + " " + req.Statement) {
		inReq[w] = true
	}
	var shared []string
	for _, w := range fieldTerms(t) {
		if inReq[w] && !cited[w] {
			shared = append(shared, w)
		}
	}
	sort.Slice(shared, func(i, j int) bool {
		wi, wj := b.Weight(shared[i]), b.Weight(shared[j])
		if wi != wj {
			return wi > wj
		}
		return shared[i] < shared[j]
	})
	if len(shared) == 0 {
		return ProbeSpec{Note: "the test and the linked requirement share no uncited word"}
	}
	chosen := map[string]bool{}
	for _, w := range shared[:min(n, len(shared))] {
		chosen[w] = true
	}
	return ProbeSpec{Task: withoutWords(t, chosen), Redact: chosen, Removed: sortedWords(chosen)}
}

// ReconstructionProbe shows the language model the cited words alone as the
// test, with the test's own code hidden from the tools.
func ReconstructionProbe(t Test, base TestAnswer) ProbeSpec {
	var ev []string
	for _, e := range base.Evidence {
		if s := strings.TrimSpace(e); s != "" {
			ev = append(ev, s)
		}
	}
	if len(ev) == 0 {
		return ProbeSpec{Note: "the answer cited no evidence"}
	}
	return ProbeSpec{Task: Task{Kind: TaskTest, ID: t.ID, Text: taskText(TestView{Doc: strings.Join(ev, "; ")})}, HideTest: t.ID}
}

func findRequirement(reqs []Requirement, key string) int {
	for i, r := range reqs {
		if r.Key == key {
			return i
		}
	}
	return -1
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
