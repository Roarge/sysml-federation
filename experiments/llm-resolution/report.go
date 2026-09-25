package main

import "math"

// Rate is k of n, with its 95% Wilson interval.
type Rate struct {
	K  int     `json:"k"`
	N  int     `json:"n"`
	Lo float64 `json:"lo"`
	Hi float64 `json:"hi"`
}

// wilson is the 95% Wilson score interval for k successes in n trials.
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

// mcnemarExact is not built yet.
func mcnemarExact(b, c int) float64 { return 0 }

// ResolverScore is one resolver's result on the tests that carry a key.
type ResolverScore struct {
	Resolver  string
	Proposed  int
	Correct   int
	Precision Rate
	Recall    Rate
}

// ProbeRate is how often a probe changed the answer.
type ProbeRate struct {
	Probe   string
	Changed Rate
	ToNone  int
	ToOther int
	Skipped int
	Errors  int
}

// PairedComparison counts the keyed tests by which resolver got each right.
type PairedComparison struct {
	BothRight         int
	OnlyLanguageModel int
	OnlyWordOverlap   int
	NeitherRight      int
	McNemarP          float64
}

// Proposal is a link proposed for a test that carries no key.
type Proposal struct {
	Test     string
	Resolver string
	Pick     string
	Reason   string
}

// BlindItem is one proposal with nothing to say who made it.
type BlindItem struct {
	Number int
	Test   string
	Pick   string
}

// KindCount counts links to elements of one kind.
type KindCount struct {
	Kind  string
	Links int
	Near  int
}

// MismatchItem is one suspicion, with where it was raised.
type MismatchItem struct {
	Where string
	Mismatch
}

// Summary is the report on one results file.
type Summary struct {
	Resolvers  []ResolverScore
	Probes     []ProbeRate
	Grounded   Rate
	Proposals  []Proposal
	Blind      []BlindItem
	Paired     PairedComparison
	OtherKinds []KindCount
	OtherNear  Rate
	BaseRate   float64
	Mismatches []MismatchItem
}

// Summarise is not built yet.
func Summarise(c Contents) Summary { return Summary{} }

// Markdown is not built yet.
func (s Summary) Markdown() string { return "" }
