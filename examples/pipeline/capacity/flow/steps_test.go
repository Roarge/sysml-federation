package flow

import (
	"errors"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the children (.+)$`, w.theChildren)
	sc.Given(`^the connections (.+)$`, w.theConnections)
	sc.Given(`^a part (\w+) with no children, whose throughput is (\d+)$`, w.aLeaf)
	sc.Given(`^(\d+) series-parallel wirings, generated from a fixed seed$`, w.generatedWirings)
	sc.Given(`^the worked example with throughputs (.+)$`, w.theWorkedExample)
	sc.Given(`^a requirement on (\S+) (\S+) (\S+) of (.+), with verification case (\S+)$`, w.aRequirement)
	sc.When(`^the capacity is rolled up$`, w.theCapacityIsRolledUp)
	sc.When(`^the capacity is analysed$`, w.theCapacityIsAnalysed)
	sc.When(`^each is rolled up$`, w.eachIsRolledUp)
	sc.When(`^the bottleneck is reported$`, w.theCapacityIsAnalysed)
	sc.When(`^the bottleneck is reported for fifty orderings of its children and connections$`, w.fiftyOrderings)
	sc.When(`^its verdict is returned$`, w.itsVerdictIsReturned)
	sc.Then(`^the rollup gives (.+)$`, w.theRollupGives)
	sc.Then(`^the capacity is (\d+), with no bottleneck$`, w.theCapacityHasNoBottleneck)
	sc.Then(`^the capacity is (\d+), limited by (.+)$`, w.theCapacityIsLimitedBy)
	sc.Then(`^each capacity is the minimum over series and the sum over parallel$`, w.eachAgreesWithMinAndSum)
	sc.Then(`^each bottleneck's throughputs add up to its capacity$`, w.eachCutSumsToItsCapacity)
	sc.Then(`^every report is (\d+), limited by (.+)$`, w.everyReportIs)
	sc.Then(`^it is (\w+), because "([^"]*)"$`, w.itIs)
}

// world is what one scenario of this package knows.
type world struct {
	t        *testing.T
	nodes    []Node
	edges    []Edge
	subject  Subject
	result   Result
	err      error
	capacity *float64
	cut      []string
	wirings  []sp
	results  []Result
	reports  []string
	verdict  verdictIn
	kind     string
	reason   string
}

// listOf splits the way the scenarios list things: "a, b, c".
func listOf(s string) []string {
	if s == "" || s == "none" {
		return nil
	}
	return strings.Split(s, ", ")
}

// childrenOf reads "ingest 2000, parse ?" as nodes, ? for a child with no
// throughput.
func childrenOf(t *testing.T, s string) []Node {
	t.Helper()
	var nodes []Node
	for _, child := range listOf(s) {
		name, value, _ := strings.Cut(child, " ")
		if value == "?" {
			nodes = append(nodes, Node{ID: name, Name: name})
			continue
		}
		parsed, err := strconv.ParseFloat(value, 64)
		nodes = append(nodes, node(name, assert.Must(t, parsed, err)))
	}
	return nodes
}

func (w *world) theChildren(s string) { w.nodes = childrenOf(w.t, s) }

func (w *world) theConnections(s string) {
	for _, connection := range listOf(s) {
		from, to, _ := strings.Cut(connection, ">")
		w.edges = append(w.edges, Edge{from, to})
	}
}

func (w *world) theCapacityIsRolledUp() { w.result, w.err = Rollup(w.nodes, w.edges) }

// describe writes a rollup the way the scenarios do.
func describe(r Result, err error) string {
	var ve *ValueError
	switch {
	case errors.As(err, &ve) && ve.Negative:
		return "refused, " + ve.Name + " has a negative one"
	case errors.As(err, &ve):
		return "refused, " + ve.Name + " has no throughput"
	case err != nil:
		return "refused, " + err.Error()
	}
	return strconv.FormatFloat(r.Capacity, 'f', -1, 64) + ", limited by " + strings.Join(r.Cut, ", ")
}

func (w *world) theRollupGives(want string) { assert.Equal(w.t, describe(w.result, w.err), want) }

func (w *world) aLeaf(name string, throughput float64) { w.subject = leafSubject(name, v(throughput)) }

func (w *world) theCapacityIsAnalysed() { w.capacity, w.cut = Analyse(w.subject) }

func (w *world) theCapacityHasNoBottleneck(capacity float64) {
	assert.True(w.t, w.capacity != nil, "a capacity is reported")
	if w.capacity != nil {
		assert.Equal(w.t, *w.capacity, capacity)
	}
	assert.Len(w.t, w.cut, 0)
}

func (w *world) theCapacityIsLimitedBy(capacity float64, bottleneck string) {
	assert.True(w.t, w.capacity != nil, "a capacity is reported")
	if w.capacity != nil {
		assert.Equal(w.t, *w.capacity, capacity)
	}
	assert.SliceEqual(w.t, w.cut, listOf(bottleneck))
}

// generatedWirings wraps each generated wiring in a wide entry and a wide
// exit, so every node is connected, as the differential test does.
func (w *world) generatedWirings(count int) {
	r := rand.New(rand.NewPCG(7, 11))
	wide := float64(1 << 20)
	for range count {
		next := 0
		entry := sp{nodes: []Node{node("in", wide)}, entries: []string{"in"}, exits: []string{"in"}, want: wide}
		exit := sp{nodes: []Node{node("out", wide)}, entries: []string{"out"}, exits: []string{"out"}, want: wide}
		w.wirings = append(w.wirings, series(series(entry, generate(r, 4, &next)), exit))
	}
}

func (w *world) eachIsRolledUp() {
	for _, wiring := range w.wirings {
		result, err := Rollup(wiring.nodes, wiring.edges)
		w.results = append(w.results, assert.Must(w.t, result, err))
	}
}

func (w *world) eachAgreesWithMinAndSum() {
	for i, wiring := range w.wirings {
		if math.Abs(w.results[i].Capacity-wiring.want) > 1e-9 {
			w.t.Errorf("wiring %d: flow %v, minimum and sum %v, connections %v", i, w.results[i].Capacity, wiring.want, wiring.edges)
		}
	}
}

func (w *world) eachCutSumsToItsCapacity() {
	for i, wiring := range w.wirings {
		sum := 0.0
		for _, name := range w.results[i].Cut {
			for _, n := range wiring.nodes {
				if n.ID == name {
					sum += *n.Value
				}
			}
		}
		if math.Abs(sum-w.results[i].Capacity) > 1e-9 {
			w.t.Errorf("wiring %d: the bottleneck %v adds up to %v, the capacity is %v", i, w.results[i].Cut, sum, w.results[i].Capacity)
		}
	}
}

func (w *world) theWorkedExample(throughputs string) {
	w.nodes = childrenOf(w.t, throughputs)
	w.edges = slices.Clone(wiring)
	w.subject = subject(w.nodes, w.edges)
}

func (w *world) fiftyOrderings() {
	r := rand.New(rand.NewPCG(3, 5))
	for range 50 {
		nodes, edges := slices.Clone(w.nodes), slices.Clone(w.edges)
		r.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })
		r.Shuffle(len(edges), func(i, j int) { edges[i], edges[j] = edges[j], edges[i] })
		w.reports = append(w.reports, describe(Rollup(nodes, edges)))
	}
}

func (w *world) everyReportIs(capacity, bottleneck string) {
	want := capacity + ", limited by " + bottleneck
	for i, report := range w.reports {
		if report != want {
			w.t.Errorf("ordering %d reports %s, want %s", i, report, want)
		}
	}
}

// subjects are the requirement subjects the verdict scenarios name.
func subjects() map[string]Subject {
	shipped := pipeline(2000, 1200, 700, 700, 1800)
	return map[string]Subject{
		"the shipped pipeline": subject(shipped, wiring),
		"the shipped pipeline with 9000 of its own": {Name: "pipeline", HasAttribute: true, Attribute: v(9000),
			Children: shipped, Edges: wiring},
		"a pipeline whose indexA has no throughput": subject([]Node{node("ingest", 2000), {ID: "s3", Name: "indexA"}},
			[]Edge{{"ingest", "s3"}}),
		"two groups with no path between them": subject([]Node{node("a", 1), node("b", 1), node("c", 1), node("d", 1), node("e", 1), node("f", 1)},
			[]Edge{{"a", "b"}, {"b", "c"}, {"c", "b"}, {"d", "e"}, {"e", "d"}, {"e", "f"}}),
		"a part whose throughput is 700":  leafSubject("x", v(700)),
		"a part whose throughput is 2.25": leafSubject("x", v(2.25)),
	}
}

func (w *world) aRequirement(quantity, comparison, limit, subject, verificationCase string) {
	s, ok := subjects()[subject]
	if !ok {
		w.t.Fatalf("no subject is called %q", subject)
	}
	w.verdict = verdictIn{subject: s}
	if quantity != "none" {
		w.verdict.quantity = quantity
	}
	if comparison != "none" {
		w.verdict.comparison = comparison
	}
	if limit != "none" {
		parsed, err := strconv.ParseFloat(limit, 64)
		w.verdict.limit = v(assert.Must(w.t, parsed, err))
	}
	if verificationCase != "none" {
		w.verdict.vc = verificationCase
	}
}

func (w *world) itsVerdictIsReturned() {
	in := w.verdict
	w.kind, w.reason = Verdict(names, in.quantity, in.comparison, in.limit, in.subject, in.vc)
}

func (w *world) itIs(kind, reason string) {
	assert.Equal(w.t, w.kind, kind)
	assert.Equal(w.t, w.reason, reason)
}
