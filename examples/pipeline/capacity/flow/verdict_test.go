package flow

import (
	"testing"

	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

var names = Names{Quantity: "capacity", Attribute: "throughput"}

func subject(nodes []Node, edges []Edge) Subject {
	return Subject{Name: "pipeline", Children: nodes, Edges: edges}
}

func leafSubject(name string, value *float64) Subject {
	return Subject{Name: name, HasAttribute: true, Attribute: value}
}

type verdictIn struct {
	quantity, comparison string
	limit                *float64
	subject              Subject
	vc                   string
}

type verdictOut struct{ Kind, Reason string }

func TestSR30_Verdict(t *testing.T) {
	shipped := subject(pipeline(2000, 1200, 700, 700, 1800), wiring)
	cases := []tabletest.Case[verdictIn, verdictOut]{
		{Name: "other quantity with a verification case", In: verdictIn{"latency", "LE", v(200), shipped, "PIPE-VC1"},
			Want: verdictOut{KindInconclusive, "PIPE-VC1 is declared and no service runs it"}},
		{Name: "other quantity without one", In: verdictIn{"latency", "LE", v(200), shipped, ""},
			Want: verdictOut{KindInconclusive, "no service computes latency"}},
		{Name: "other quantity wins over a bad child", In: verdictIn{"latency", "LE", v(200),
			subject([]Node{{ID: "a", Name: "a"}, node("b", 1)}, []Edge{{"a", "b"}}), ""},
			Want: verdictOut{KindInconclusive, "no service computes latency"}},
		{Name: "no limit", In: verdictIn{"capacity", "GE", nil, shipped, ""},
			Want: verdictOut{KindInconclusive, "no limit to compare against"}},
		{Name: "no comparison", In: verdictIn{"capacity", "", v(1500), shipped, ""},
			Want: verdictOut{KindInconclusive, "no comparison to apply"}},
		// An entry needs no incoming connection and an exit no outgoing one, and a
		// child with neither is left out of the network, so the two must be
		// different children and the smallest wiring that separates them puts each
		// in its own group of three.
		{Name: "an entry and an exit with no path between them", In: verdictIn{"capacity", "GE", v(1500),
			subject([]Node{node("a", 1), node("b", 1), node("c", 1), node("d", 1), node("e", 1), node("f", 1)},
				[]Edge{{"a", "b"}, {"b", "c"}, {"c", "b"}, {"d", "e"}, {"e", "d"}, {"e", "f"}}), ""},
			Want: verdictOut{KindFail, "capacity 0 against 1500, no path from entry to exit"}},
		{Name: "children win over the subject's own attribute", In: verdictIn{"capacity", "GE", v(1500),
			Subject{Name: "pipeline", HasAttribute: true, Attribute: v(9000), Children: pipeline(2000, 1200, 700, 700, 1800), Edges: wiring}, ""},
			Want: verdictOut{KindFail, "capacity 1200 against 1500, limited by parse"}},
		{Name: "missing child value", In: verdictIn{"capacity", "GE", v(1500),
			subject([]Node{node("ingest", 2000), {ID: "s3", Name: "indexA"}}, []Edge{{"ingest", "s3"}}), ""},
			Want: verdictOut{KindError, "indexA has missing throughput"}},
		{Name: "negative child value", In: verdictIn{"capacity", "GE", v(1500),
			subject([]Node{node("ingest", 2000), node("parse", -1)}, []Edge{{"ingest", "parse"}}), ""},
			Want: verdictOut{KindError, "parse has negative throughput"}},
		{Name: "error wins over no exit", In: verdictIn{"capacity", "GE", v(1500),
			subject([]Node{{ID: "a", Name: "a"}, node("b", 1)}, []Edge{{"a", "b"}, {"b", "a"}}), ""},
			Want: verdictOut{KindError, "a has missing throughput"}},
		{Name: "error wins over a missing limit", In: verdictIn{"capacity", "GE", nil,
			subject([]Node{node("parse", -1)}, nil), ""},
			Want: verdictOut{KindError, "parse has negative throughput"}},
		{Name: "no entry", In: verdictIn{"capacity", "GE", v(1500),
			subject([]Node{node("a", 1), node("b", 1)}, []Edge{{"a", "b"}, {"b", "a"}}), ""},
			Want: verdictOut{KindInconclusive, "no entry part"}},
		{Name: "no exit", In: verdictIn{"capacity", "GE", v(1500),
			subject([]Node{node("a", 1), node("b", 1), node("c", 1)}, []Edge{{"a", "b"}, {"b", "c"}, {"c", "b"}}), ""},
			Want: verdictOut{KindInconclusive, "no exit part"}},
		{Name: "empty subject", In: verdictIn{"capacity", "GE", v(1500), Subject{Name: "empty"}, ""},
			Want: verdictOut{KindInconclusive, "no children to analyse"}},
		{Name: "shipped fails on parse", In: verdictIn{"capacity", "GE", v(1500), shipped, ""},
			Want: verdictOut{KindFail, "capacity 1200 against 1500, limited by parse"}},
		{Name: "parse to 1700 still fails", In: verdictIn{"capacity", "GE", v(1500), subject(pipeline(2000, 1700, 700, 700, 1800), wiring), ""},
			Want: verdictOut{KindFail, "capacity 1400 against 1500, limited by indexA, indexB"}},
		{Name: "then indexA to 900 passes", In: verdictIn{"capacity", "GE", v(1500), subject(pipeline(2000, 1700, 900, 700, 1800), wiring), ""},
			Want: verdictOut{KindPass, "capacity 1600 against 1500, limited by indexA, indexB"}},
		{Name: "parse to zero", In: verdictIn{"capacity", "GE", v(1500), subject(pipeline(2000, 0, 700, 700, 1800), wiring), ""},
			Want: verdictOut{KindFail, "capacity 0 against 1500, limited by parse"}},
		{Name: "limit lowered to 1000 passes", In: verdictIn{"capacity", "GE", v(1000), shipped, ""},
			Want: verdictOut{KindPass, "capacity 1200 against 1000, limited by parse"}},
		{Name: "leaf passes", In: verdictIn{"capacity", "GE", v(1500), leafSubject("ingest", v(2000)), ""},
			Want: verdictOut{KindPass, "throughput 2000 against 1500"}},
		{Name: "leaf fails", In: verdictIn{"capacity", "GE", v(750), leafSubject("indexB", v(700)), ""},
			Want: verdictOut{KindFail, "throughput 700 against 750"}},
		{Name: "leaf with a missing value", In: verdictIn{"capacity", "GE", v(750), leafSubject("indexB", nil), ""},
			Want: verdictOut{KindError, "indexB has missing throughput"}},
		{Name: "leaf with a negative value", In: verdictIn{"capacity", "GE", v(750), leafSubject("indexB", v(-1)), ""},
			Want: verdictOut{KindError, "indexB has negative throughput"}},
		{Name: "GT", In: verdictIn{"capacity", "GT", v(700), leafSubject("x", v(700)), ""}, Want: verdictOut{KindFail, "throughput 700 against 700"}},
		{Name: "LE", In: verdictIn{"capacity", "LE", v(700), leafSubject("x", v(700)), ""}, Want: verdictOut{KindPass, "throughput 700 against 700"}},
		{Name: "LT", In: verdictIn{"capacity", "LT", v(700), leafSubject("x", v(700)), ""}, Want: verdictOut{KindFail, "throughput 700 against 700"}},
		{Name: "EQ", In: verdictIn{"capacity", "EQ", v(700), leafSubject("x", v(700)), ""}, Want: verdictOut{KindPass, "throughput 700 against 700"}},
		{Name: "fractional values print plainly", In: verdictIn{"capacity", "GE", v(1.5), leafSubject("x", v(2.25)), ""},
			Want: verdictOut{KindPass, "throughput 2.25 against 1.5"}},
	}
	tabletest.Run(t, cases, func(_ *testing.T, in verdictIn) verdictOut {
		kind, reason := Verdict(names, in.quantity, in.comparison, in.limit, in.subject, in.vc)
		return verdictOut{kind, reason}
	})
}

func TestSR29_Analyse(t *testing.T) {
	c, cut := Analyse(subject(pipeline(2000, 1700, 700, 700, 1800), wiring))
	assert.Equal(t, *c, 1400)
	assert.SliceEqual(t, cut, []string{"indexA", "indexB"})
	c, cut = Analyse(leafSubject("ingest", v(2000)))
	assert.Equal(t, *c, 2000)
	assert.Len(t, cut, 0)
	c, _ = Analyse(Subject{Name: "empty"})
	assert.True(t, c == nil, "no capacity for an empty subject")
	c, _ = Analyse(subject([]Node{node("a", 1), node("b", 1)}, []Edge{{"a", "b"}, {"b", "a"}}))
	assert.True(t, c == nil, "no capacity without an entry")
}
