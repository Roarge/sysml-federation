package main

import (
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
)

// The probes. Each asks the same question again with one thing changed, and
// compares the answer with the base answer.
const (
	ProbeBase           = "base"           // the test as it is
	ProbeDeletion       = "deletion"       // the words the model cited, deleted
	ProbeControl        = "control"        // as many other words, deleted at random
	ProbeRareShared     = "rare-shared"    // as many shared uncited words, rarest first
	ProbeReconstruction = "reconstruction" // the test replaced by the cited words alone
	ProbeOrder          = "order"          // the requirements in a shuffled order
	ProbeRepeat         = "repeat"         // the base question, asked again unchanged
)

// AllProbes lists them in the order a run makes them.
var AllProbes = []string{ProbeBase, ProbeDeletion, ProbeControl, ProbeRareShared, ProbeReconstruction, ProbeOrder, ProbeRepeat}

// Variant is one question: the requirements as listed and the test as shown.
type Variant struct {
	Requirements []Requirement
	View         TestView
	Removed      []string // words deleted, as written, for the record
	Note         string   // why no call was made, when none was
}

// evidenceTerms are the normalised words of the model's evidence.
func evidenceTerms(ev []string) map[string]bool {
	out := map[string]bool{}
	for _, e := range ev {
		for _, t := range terms(e) {
			out[t] = true
		}
	}
	return out
}

func findRequirement(reqs []Requirement, key string) int {
	for i, r := range reqs {
		if r.Key == key {
			return i
		}
	}
	return -1
}

// deletionVariant deletes every occurrence of the cited words from the test
// and from the requirement the model picked.
func deletionVariant(reqs []Requirement, t Test, base Answer) Variant {
	drop := evidenceTerms(base.Evidence)
	v := Variant{Requirements: append([]Requirement{}, reqs...), View: viewOf(t)}
	for w := range drop {
		v.Removed = append(v.Removed, w)
	}
	sort.Strings(v.Removed)
	total := 0
	for _, f := range []*string{&v.View.Name, &v.View.Doc, &v.View.File} {
		var n int
		*f, n = removeTerms(*f, drop)
		total += n
	}
	if i := findRequirement(v.Requirements, base.Requirement); i >= 0 {
		r := v.Requirements[i]
		var n int
		r.Name, n = removeTerms(r.Name, drop)
		total += n
		r.Statement, n = removeTerms(r.Statement, drop)
		total += n
		v.Requirements[i] = r
	}
	if total == 0 {
		v.Note = "none of the cited words occur in the test or the picked requirement"
	}
	return v
}

// deletionCount is how many word occurrences the deletion probe removed.
func deletionCount(reqs []Requirement, t Test, base Answer) int {
	drop := evidenceTerms(base.Evidence)
	total := 0
	count := func(s string) {
		for _, w := range terms(s) {
			if drop[w] {
				total++
			}
		}
	}
	count(t.Name)
	count(t.Doc)
	count(t.File)
	if i := findRequirement(reqs, base.Requirement); i >= 0 {
		count(reqs[i].Name)
		count(reqs[i].Statement)
	}
	return total
}

// controlVariant deletes the same number of word occurrences as the deletion
// probe, chosen at random from the words the model did not cite, in the same
// places. If explanations name what the model relied on, deleting them should
// change its answer far more often than this does.
func controlVariant(reqs []Requirement, t Test, base Answer, seed int64) Variant {
	drop := evidenceTerms(base.Evidence)
	n := deletionCount(reqs, t, base)
	v := Variant{Requirements: append([]Requirement{}, reqs...), View: viewOf(t)}
	if n == 0 {
		v.Note = "nothing was deleted in the deletion probe, so there is nothing to match"
		return v
	}
	ri := findRequirement(v.Requirements, base.Requirement)
	type slot struct{ field, pos int }
	fields := []*string{&v.View.Name, &v.View.Doc, &v.View.File}
	if ri >= 0 {
		fields = append(fields, &v.Requirements[ri].Name, &v.Requirements[ri].Statement)
	}
	var slots []slot
	for fi, f := range fields {
		for pos, w := range terms(*f) {
			if !drop[w] {
				slots = append(slots, slot{fi, pos})
			}
		}
	}
	rng := rand.New(rand.NewSource(seed ^ int64(hashString(t.ID))))
	rng.Shuffle(len(slots), func(i, j int) { slots[i], slots[j] = slots[j], slots[i] })
	if n > len(slots) {
		n = len(slots)
	}
	chosen := make([]map[int]bool, len(fields))
	for i := range chosen {
		chosen[i] = map[int]bool{}
	}
	for _, s := range slots[:n] {
		chosen[s.field][s.pos] = true
		v.Removed = append(v.Removed, terms(*fields[s.field])[s.pos])
	}
	for fi, f := range fields {
		*f = removeOccurrences(*f, chosen[fi])
	}
	sort.Strings(v.Removed)
	return v
}

// reconstructionVariant shows the model nothing of the test but the words it
// cited, as the test's doc.
func reconstructionVariant(reqs []Requirement, base Answer) Variant {
	var ev []string
	for _, e := range base.Evidence {
		if s := strings.TrimSpace(e); s != "" {
			ev = append(ev, s)
		}
	}
	v := Variant{Requirements: reqs, View: TestView{Doc: strings.Join(ev, "; ")}}
	if len(ev) == 0 {
		v.Note = "the model cited no evidence"
	}
	return v
}

// orderVariant lists the requirements in a shuffled order. The shuffle
// depends on the seed alone, so every test in a run gets the same one and
// the server can cache the shuffled list as it caches the original.
func orderVariant(reqs []Requirement, seed int64) Variant {
	shuffled := append([]Requirement{}, reqs...)
	rng := rand.New(rand.NewSource(seed + 1))
	rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	return Variant{Requirements: shuffled}
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

// jaccard compares two answers' evidence as sets of normalised words.
func jaccard(a, b []string) float64 {
	x, y := evidenceTerms(a), evidenceTerms(b)
	if len(x) == 0 && len(y) == 0 {
		return 1
	}
	inter := 0
	for w := range x {
		if y[w] {
			inter++
		}
	}
	return float64(inter) / float64(len(x)+len(y)-inter)
}

// rareSharedVariant is the second control. The words the model cites are
// often the most telling words the test and its requirement share, so beating
// a control of random words could mean only that. This control deletes the
// uncited words the two share instead, the rarest first, until at least as
// many words are gone as the deletion took. If explanations name what the
// answer depended on, deleting the cited words should still change it more.
func rareSharedVariant(reqs []Requirement, t Test, base Answer, b *Baseline) Variant {
	drop := evidenceTerms(base.Evidence)
	n := deletionCount(reqs, t, base)
	v := Variant{Requirements: append([]Requirement{}, reqs...), View: viewOf(t)}
	ri := findRequirement(v.Requirements, base.Requirement)
	if n == 0 || ri < 0 {
		v.Note = "nothing was deleted in the deletion probe, so there is nothing to match"
		return v
	}
	inTest, inReq := map[string]int{}, map[string]int{}
	for _, f := range []string{t.Name, t.Doc, t.File} {
		for _, w := range terms(f) {
			inTest[w]++
		}
	}
	for _, f := range []string{reqs[ri].Name, reqs[ri].Statement} {
		for _, w := range terms(f) {
			inReq[w]++
		}
	}
	var shared []string
	for w := range inTest {
		if inReq[w] > 0 && !drop[w] {
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
	chosen := map[string]bool{}
	for removed := 0; removed < n && len(v.Removed) < len(shared); {
		w := shared[len(v.Removed)]
		chosen[w] = true
		v.Removed = append(v.Removed, w)
		removed += inTest[w] + inReq[w]
	}
	if len(chosen) == 0 {
		v.Note = "the test and the picked requirement share no word the model didn't cite"
		return v
	}
	for _, f := range []*string{&v.View.Name, &v.View.Doc, &v.View.File} {
		*f, _ = removeTerms(*f, chosen)
	}
	r := v.Requirements[ri]
	r.Name, _ = removeTerms(r.Name, chosen)
	r.Statement, _ = removeTerms(r.Statement, chosen)
	v.Requirements[ri] = r
	return v
}
