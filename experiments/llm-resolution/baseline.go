package main

import (
	"math"
	"sort"
)

// BaselineThreshold is the similarity below which the word-overlap baseline
// proposes no link. It was set before the baseline was first run and is not
// tuned on the results.
const BaselineThreshold = 0.15

// Baseline ranks requirements for a test by the words they share, weighted
// so that a word found in few requirements counts for more (TF-IDF with
// cosine similarity). Its reason for a link is the shared words themselves.
type Baseline struct {
	keys    []string
	vectors []map[string]float64
	idf     map[string]float64
}

// Ranked is one candidate requirement with its score and the shared words
// that produced it, heaviest first.
type Ranked struct {
	Key    string   `json:"key"`
	Score  float64  `json:"score"`
	Shared []string `json:"shared"`
}

// BaselineResult is the baseline's answer for one test.
type BaselineResult struct {
	Test   string   `json:"test"`
	Gold   string   `json:"gold,omitempty"`
	Pick   string   `json:"pick"` // "none" below the threshold
	Top    []Ranked `json:"top"`  // the best three, whatever their scores
	Rank   int      `json:"gold_rank,omitempty"`
	Reason []string `json:"reason"`

	// The systems model baseline's answer, empty in a file from before it.
	ModelPick string `json:"model_pick,omitempty"`
	ModelVia  string `json:"model_via,omitempty"` // the element it came through
	Reach     bool   `json:"reach,omitempty"`     // the recorded requirement within the search's reach
}

// NewBaseline builds the weights from the requirements alone.
func NewBaseline(reqs []Requirement) *Baseline {
	keys := make([]string, len(reqs))
	texts := make([]string, len(reqs))
	for i, r := range reqs {
		keys[i], texts[i] = r.Key, r.Name+" "+r.Statement
	}
	return newIndex(keys, texts)
}

// newIndex weighs any set of texts the way the baseline weighs the
// requirements, so the search tool ranks elements as the baseline ranks
// requirements.
func newIndex(keys, texts []string) *Baseline {
	b := &Baseline{idf: map[string]float64{}}
	df := map[string]int{}
	docs := make([][]string, len(texts))
	for i, text := range texts {
		docs[i] = terms(text)
		seen := map[string]bool{}
		for _, t := range docs[i] {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}
	n := float64(len(texts))
	for t, d := range df {
		b.idf[t] = math.Log((n+1)/(float64(d)+1)) + 1
	}
	for i := range texts {
		b.keys = append(b.keys, keys[i])
		b.vectors = append(b.vectors, b.weigh(docs[i]))
	}
	return b
}

func (b *Baseline) weigh(ts []string) map[string]float64 {
	v := map[string]float64{}
	for _, t := range ts {
		if w, ok := b.idf[t]; ok {
			v[t] += w
		}
	}
	return v
}

// testText is everything the baseline reads about a test.
func testText(t Test) string {
	return t.Name + " " + t.Package + " " + t.File + " " + t.Doc
}

// Rank scores every requirement for the test, best first.
func (b *Baseline) Rank(t Test) []Ranked { return b.rankTerms(terms(testText(t))) }

// rankTerms scores every text for the words given, best first.
func (b *Baseline) rankTerms(words []string) []Ranked {
	q := b.weigh(words)
	out := make([]Ranked, len(b.keys))
	for i, v := range b.vectors {
		out[i] = Ranked{Key: b.keys[i], Score: cosine(q, v), Shared: shared(q, v)}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Key < out[j].Key
	})
	return out
}

// Resolve gives the baseline's answer for one test.
func (b *Baseline) Resolve(t Test) BaselineResult {
	ranked := b.Rank(t)
	r := BaselineResult{Test: t.ID, Gold: t.Gold, Pick: "none", Top: ranked[:min(3, len(ranked))]}
	if len(ranked) > 0 && ranked[0].Score >= BaselineThreshold {
		r.Pick = ranked[0].Key
		r.Reason = ranked[0].Shared
	}
	for i, x := range ranked {
		if x.Key == t.Gold {
			r.Rank = i + 1
		}
	}
	return r
}

func cosine(a, b map[string]float64) float64 {
	var dot, na, nb float64
	for t, x := range a {
		na += x * x
		if y, ok := b[t]; ok {
			dot += x * y
		}
	}
	for _, y := range b {
		nb += y * y
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return math.Round(dot/math.Sqrt(na*nb)*10000) / 10000
}

func shared(a, b map[string]float64) []string {
	var ts []string
	for t := range a {
		if _, ok := b[t]; ok {
			ts = append(ts, t)
		}
	}
	sort.Slice(ts, func(i, j int) bool {
		wi, wj := a[ts[i]]*b[ts[i]], a[ts[j]]*b[ts[j]]
		if wi != wj {
			return wi > wj
		}
		return ts[i] < ts[j]
	})
	if len(ts) > 5 {
		ts = ts[:5]
	}
	return ts
}

// Weight is how much a normalised word counts: more the fewer requirements
// use it, and 0 for a word no requirement uses.
func (b *Baseline) Weight(term string) float64 { return b.idf[term] }

// traceRelations are the links the systems model baseline follows from an
// element to a requirement: the ones a systems engineer writes to say what
// meets, checks or gives rise to a requirement.
var traceRelations = map[string]bool{
	"satisfies": true, "satisfied by": true,
	"verifies": true, "verified by": true,
	"derives": true, "derived from": true,
}

// ModelBaseline resolves a test from the systems model without a language
// model: the search's ranking of every element, then at most one trace link
// to a requirement (EXP-SR-24).
type ModelBaseline struct {
	w    *Wiki
	all  *Baseline       // every element, weighted as the search weighs them
	reqs *Baseline       // the requirements alone, to choose among several
	keys map[string]bool // the requirements a pick may name
}

// NewModelBaseline builds the baseline over the systems model a test task
// shows.
func NewModelBaseline(w *Wiki, reqs []Requirement) *ModelBaseline {
	m := &ModelBaseline{w: w, reqs: NewBaseline(reqs), keys: map[string]bool{}}
	for _, r := range reqs {
		m.keys[r.Key] = true
	}
	var ids, texts []string
	for _, e := range w.Elements {
		ids = append(ids, e.ID)
		texts = append(texts, pageText(e))
	}
	m.all = newIndex(ids, texts)
	return m
}

// Resolve gives the baseline's answer for one test, and whether its recorded
// requirement lies within the search's reach.
func (m *ModelBaseline) Resolve(t Test) (pick, via string, reach bool) {
	ranked := m.all.rankTerms(terms(testText(t)))
	pick, via = m.pickFrom(ranked, t)
	return pick, via, t.Gold != "" && m.reach(ranked, t.Gold)
}

// pickFrom takes the first ranked element that is a requirement, or that one
// trace link joins to one.
func (m *ModelBaseline) pickFrom(ranked []Ranked, t Test) (pick, via string) {
	for _, r := range ranked {
		if r.Score <= 0 {
			break
		}
		e, ok := m.w.Get(r.Key)
		if !ok {
			continue
		}
		if m.keys[e.Short] {
			return e.Short, ""
		}
		if linked := m.traced(e); len(linked) > 0 {
			for _, x := range m.reqs.Rank(t) {
				if linked[x.Key] {
					return x.Key, e.ID
				}
			}
		}
	}
	return "none", ""
}

// reach says whether gold is among the first eight ranked elements, or one
// trace link from one of them.
func (m *ModelBaseline) reach(ranked []Ranked, gold string) bool {
	for i, r := range ranked {
		if i == LimitFind || r.Score <= 0 {
			break
		}
		e, ok := m.w.Get(r.Key)
		if !ok {
			continue
		}
		if e.Short == gold || m.traced(e)[gold] {
			return true
		}
	}
	return false
}

// traced gives the requirements one trace link joins to e.
func (m *ModelBaseline) traced(e *Element) map[string]bool {
	out := map[string]bool{}
	for _, l := range m.w.links[e] {
		if traceRelations[l.rel] && m.keys[l.other.Short] {
			out[l.other.Short] = true
		}
	}
	return out
}
