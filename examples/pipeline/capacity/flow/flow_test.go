package flow

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

func v(f float64) *float64 { return &f }

func node(id string, value float64) Node { return Node{ID: id, Name: id, Value: v(value)} }

// pipeline is the worked example: ingest, parse, the index pair, serve.
func pipeline(ingest, parse, indexA, indexB, serve float64) []Node {
	return []Node{node("ingest", ingest), node("parse", parse), node("indexA", indexA), node("indexB", indexB), node("serve", serve)}
}

var wiring = []Edge{{"ingest", "parse"}, {"parse", "indexA"}, {"parse", "indexB"}, {"indexA", "serve"}, {"indexB", "serve"}}

type in struct {
	nodes []Node
	edges []Edge
}

func TestSR28_Rollup(t *testing.T) {
	empty := []cmp.Option{cmpopts.EquateEmpty()}
	cases := []tabletest.ErrCase[in, Result]{
		{Name: "a chain takes the minimum", In: in{[]Node{node("a", 2000), node("b", 1200), node("c", 1800)}, []Edge{{"a", "b"}, {"b", "c"}}},
			Want: Result{Capacity: 1200, Cut: []string{"b"}, Entry: []string{"a"}, Exit: []string{"c"}}},
		{Name: "shipped: fan-out and fan-in add the branches", In: in{pipeline(2000, 1200, 700, 700, 1800), wiring},
			Want: Result{Capacity: 1200, Cut: []string{"parse"}, Entry: []string{"ingest"}, Exit: []string{"serve"}}},
		{Name: "ingest raised: nothing moves", In: in{pipeline(3000, 1200, 700, 700, 1800), wiring},
			Want: Result{Capacity: 1200, Cut: []string{"parse"}, Entry: []string{"ingest"}, Exit: []string{"serve"}}},
		{Name: "parse to 1700: the cut moves to the index pair", In: in{pipeline(2000, 1700, 700, 700, 1800), wiring},
			Want: Result{Capacity: 1400, Cut: []string{"indexA", "indexB"}, Entry: []string{"ingest"}, Exit: []string{"serve"}}},
		{Name: "then indexA to 900", In: in{pipeline(2000, 1700, 900, 700, 1800), wiring},
			Want: Result{Capacity: 1600, Cut: []string{"indexA", "indexB"}, Entry: []string{"ingest"}, Exit: []string{"serve"}}},
		{Name: "a tie reports the source side", In: in{pipeline(2000, 1600, 900, 700, 1800), wiring},
			Want: Result{Capacity: 1600, Cut: []string{"parse"}, Entry: []string{"ingest"}, Exit: []string{"serve"}}},
		{Name: "nested branches", In: in{
			[]Node{node("a", 5000), node("b", 600), node("c", 500), node("d", 300), node("e", 400), node("f", 900), node("g", 2000)},
			[]Edge{{"a", "b"}, {"a", "c"}, {"b", "d"}, {"b", "e"}, {"c", "f"}, {"d", "g"}, {"e", "g"}, {"f", "g"}}},
			Want: Result{Capacity: 1100, Cut: []string{"b", "c"}, Entry: []string{"a"}, Exit: []string{"g"}}},
		{Name: "a cycle is ordinary wiring", In: in{[]Node{node("a", 100), node("b", 50), node("c", 80), node("d", 200)},
			[]Edge{{"a", "b"}, {"b", "c"}, {"c", "b"}, {"c", "d"}}},
			Want: Result{Capacity: 50, Cut: []string{"b"}, Entry: []string{"a"}, Exit: []string{"d"}}},
		{Name: "a disconnected child is left out", In: in{append(pipeline(2000, 1200, 700, 700, 1800), node("spare", 1)), wiring},
			Want: Result{Capacity: 1200, Cut: []string{"parse"}, Entry: []string{"ingest"}, Exit: []string{"serve"}}},
		{Name: "an unreachable pair is left out", In: in{append(pipeline(2000, 1200, 700, 700, 1800), node("x", 1), node("y", 1)),
			append(slices.Clone(wiring), Edge{"x", "y"}, Edge{"y", "x"})},
			Want: Result{Capacity: 1200, Cut: []string{"parse"}, Entry: []string{"ingest"}, Exit: []string{"serve"}}},
		{Name: "zero on the serial path", In: in{[]Node{node("a", 0), node("b", 5)}, []Edge{{"a", "b"}}},
			Want: Result{Capacity: 0, Cut: []string{"a"}, Entry: []string{"a"}, Exit: []string{"b"}}},
		{Name: "several entries and exits", In: in{[]Node{node("a", 10), node("b", 20), node("c", 5), node("d", 40)},
			[]Edge{{"a", "c"}, {"b", "c"}, {"c", "d"}, {"b", "d"}}},
			Want: Result{Capacity: 25, Cut: []string{"b", "c"}, Entry: []string{"a", "b"}, Exit: []string{"d"}}},
		{Name: "no children", In: in{}, ErrIs: ErrNoNodes},
		{Name: "no entry", In: in{[]Node{node("a", 1), node("b", 1)}, []Edge{{"a", "b"}, {"b", "a"}}}, ErrIs: ErrNoEntry},
		{Name: "no exit", In: in{[]Node{node("a", 1), node("b", 1), node("c", 1)}, []Edge{{"a", "b"}, {"b", "c"}, {"c", "b"}}}, ErrIs: ErrNoExit},
		{Name: "missing value", In: in{[]Node{{ID: "a", Name: "a"}, node("b", 1)}, []Edge{{"a", "b"}}}, WantErr: true},
		{Name: "negative value", In: in{[]Node{node("a", -1), node("b", 1)}, []Edge{{"a", "b"}}}, WantErr: true},
	}
	for i := range cases {
		cases[i].Opts = empty
	}
	tabletest.RunErr(t, cases, func(_ *testing.T, in in) (Result, error) { return Rollup(in.nodes, in.edges) })
}

func TestRollupNamesTheFaultyChild(t *testing.T) {
	_, err := Rollup([]Node{{ID: "s3", Name: "indexA"}, node("b", 1)}, []Edge{{"s3", "b"}})
	ve := assert.ErrorAs[*ValueError](t, err)
	assert.Equal(t, ve.Name, "indexA")
	assert.Equal(t, ve.Negative, false)
	_, err = Rollup([]Node{{ID: "s4", Name: "indexB", Value: v(-5)}}, nil)
	ve = assert.ErrorAs[*ValueError](t, err)
	assert.Equal(t, ve.Negative, true)
}

// sp is a series-parallel wiring with the capacity the README's two rules give it.
type sp struct {
	nodes          []Node
	edges          []Edge
	entries, exits []string
	want           float64
}

func leaf(r *rand.Rand, next *int) sp {
	id := fmt.Sprintf("n%d", *next)
	*next++
	value := float64(r.IntN(1000))
	return sp{nodes: []Node{node(id, value)}, entries: []string{id}, exits: []string{id}, want: value}
}

func series(a, b sp) sp {
	out := sp{nodes: slices.Concat(a.nodes, b.nodes), edges: slices.Concat(a.edges, b.edges), entries: a.entries, exits: b.exits, want: math.Min(a.want, b.want)}
	for _, x := range a.exits {
		for _, y := range b.entries {
			out.edges = append(out.edges, Edge{x, y})
		}
	}
	return out
}

func parallel(a, b sp) sp {
	return sp{nodes: slices.Concat(a.nodes, b.nodes), edges: slices.Concat(a.edges, b.edges),
		entries: slices.Concat(a.entries, b.entries), exits: slices.Concat(a.exits, b.exits), want: a.want + b.want}
}

func generate(r *rand.Rand, depth int, next *int) sp {
	if depth == 0 || r.IntN(3) == 0 {
		return leaf(r, next)
	}
	a, b := generate(r, depth-1, next), generate(r, depth-1, next)
	if r.IntN(2) == 0 {
		return series(a, b)
	}
	return parallel(a, b)
}

// TestSR28_FlowAgreesWithMinAndSum is the differential test of the capacity
// model. A wide entry and exit server wrap each wiring so every node is
// connected.
func TestSR28_FlowAgreesWithMinAndSum(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 11))
	for i := range 300 {
		next := 0
		wide := float64(1 << 20)
		w := series(series(sp{nodes: []Node{node("in", wide)}, entries: []string{"in"}, exits: []string{"in"}, want: wide}, generate(r, 4, &next)),
			sp{nodes: []Node{node("out", wide)}, entries: []string{"out"}, exits: []string{"out"}, want: wide})
		result, err := Rollup(w.nodes, w.edges)
		got := assert.Must(t, result, err)
		if math.Abs(got.Capacity-w.want) > 1e-9 {
			t.Fatalf("case %d: flow %v, min/sum %v, wiring %v", i, got.Capacity, w.want, w.edges)
		}
	}
}
