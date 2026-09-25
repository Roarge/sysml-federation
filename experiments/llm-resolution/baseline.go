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
}

// NewBaseline builds the weights from the requirements alone.
func NewBaseline(reqs []Requirement) *Baseline {
	b := &Baseline{idf: map[string]float64{}}
	df := map[string]int{}
	docs := make([][]string, len(reqs))
	for i, r := range reqs {
		docs[i] = terms(r.Name + " " + r.Statement)
		seen := map[string]bool{}
		for _, t := range docs[i] {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}
	n := float64(len(reqs))
	for t, d := range df {
		b.idf[t] = math.Log((n+1)/(float64(d)+1)) + 1
	}
	for i, r := range reqs {
		b.keys = append(b.keys, r.Key)
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
func (b *Baseline) Rank(t Test) []Ranked {
	q := b.weigh(terms(testText(t)))
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
