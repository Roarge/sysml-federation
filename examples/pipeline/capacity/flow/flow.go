// Package flow computes the capacity of a wiring as a maximum flow and reports
// the source-side minimum cut as the bottleneck. It is the mechanism of the
// capacity model and knows nothing about GraphQL or the example.
package flow

import (
	"errors"
	"fmt"
	"math"
)

// Node is one child part with the configured attribute's value, nil when the
// attribute has no value.
type Node struct {
	ID, Name string
	Value    *float64
}

// Edge is one directed connection between two children.
type Edge struct{ From, To string }

// Result carries the capacity and, as IDs in child order, the cut, the entry
// children and the exit children.
type Result struct {
	Capacity         float64
	Cut, Entry, Exit []string
}

// The three ways a wiring gives no number.
var (
	ErrNoNodes = errors.New("no children to analyse")
	ErrNoEntry = errors.New("no entry part")
	ErrNoExit  = errors.New("no exit part")
)

// ValueError names a child whose attribute is missing or negative.
type ValueError struct {
	ID, Name string
	Negative bool
}

// Error names the child and says which of the two faults it carries.
func (e *ValueError) Error() string {
	if e.Negative {
		return fmt.Sprintf("%s has negative value", e.Name)
	}
	return fmt.Sprintf("%s has missing value", e.Name)
}

// eps is the residual below which an edge counts as saturated, so a rounding
// remainder cannot open a spurious augmenting path or extend the cut.
const eps = 1e-9

// Rollup builds the network of the capacity model and runs Dinic's algorithm.
// Every child becomes an in-node and an out-node joined by an edge of its
// value. Connections are unlimited edges from out to in. A super-source feeds
// every child nobody feeds and a super-sink drains every child that feeds
// nobody. A child with no connections is left out.
func Rollup(nodes []Node, edges []Edge) (Result, error) {
	if len(nodes) == 0 {
		return Result{}, ErrNoNodes
	}
	index := make(map[string]int, len(nodes))
	for i, n := range nodes {
		if n.Value == nil {
			return Result{}, &ValueError{ID: n.ID, Name: n.Name}
		}
		if *n.Value < 0 {
			return Result{}, &ValueError{ID: n.ID, Name: n.Name, Negative: true}
		}
		index[n.ID] = i
	}
	hasIn, hasOut := make([]bool, len(nodes)), make([]bool, len(nodes))
	var wires [][2]int
	for _, e := range edges {
		from, okFrom := index[e.From]
		to, okTo := index[e.To]
		if !okFrom || !okTo {
			continue
		}
		hasOut[from], hasIn[to] = true, true
		wires = append(wires, [2]int{from, to})
	}
	n := len(nodes)
	source, sink := 2*n, 2*n+1
	g := newGraph(2*n + 2)
	var res Result
	for i, nd := range nodes {
		if !hasIn[i] && !hasOut[i] {
			continue
		}
		g.add(2*i, 2*i+1, *nd.Value)
		if !hasIn[i] {
			g.add(source, 2*i, math.Inf(1))
			res.Entry = append(res.Entry, nd.ID)
		}
		if !hasOut[i] {
			g.add(2*i+1, sink, math.Inf(1))
			res.Exit = append(res.Exit, nd.ID)
		}
	}
	for _, w := range wires {
		g.add(2*w[0]+1, 2*w[1], math.Inf(1))
	}
	if len(res.Entry) == 0 {
		return Result{}, ErrNoEntry
	}
	if len(res.Exit) == 0 {
		return Result{}, ErrNoExit
	}
	res.Capacity = g.maxFlow(source, sink)
	reach := g.reachable(source)
	for i, nd := range nodes {
		if reach[2*i] && !reach[2*i+1] {
			res.Cut = append(res.Cut, nd.ID)
		}
	}
	return res, nil
}

type edge struct {
	to, rev int
	cap     float64
}

type graph struct {
	adj   [][]edge
	level []int
	next  []int
}

func newGraph(n int) *graph { return &graph{adj: make([][]edge, n)} }

func (g *graph) add(u, v int, c float64) {
	g.adj[u] = append(g.adj[u], edge{to: v, rev: len(g.adj[v]), cap: c})
	g.adj[v] = append(g.adj[v], edge{to: u, rev: len(g.adj[u]) - 1})
}

func (g *graph) bfs(s, t int) bool {
	g.level = make([]int, len(g.adj))
	for i := range g.level {
		g.level[i] = -1
	}
	g.level[s] = 0
	for queue := []int{s}; len(queue) > 0; queue = queue[1:] {
		u := queue[0]
		for _, e := range g.adj[u] {
			if e.cap > eps && g.level[e.to] < 0 {
				g.level[e.to] = g.level[u] + 1
				queue = append(queue, e.to)
			}
		}
	}
	return g.level[t] >= 0
}

func (g *graph) dfs(u, t int, f float64) float64 {
	if u == t {
		return f
	}
	for ; g.next[u] < len(g.adj[u]); g.next[u]++ {
		e := &g.adj[u][g.next[u]]
		if e.cap <= eps || g.level[e.to] != g.level[u]+1 {
			continue
		}
		if d := g.dfs(e.to, t, math.Min(f, e.cap)); d > eps {
			e.cap -= d
			g.adj[e.to][e.rev].cap += d
			return d
		}
	}
	return 0
}

func (g *graph) maxFlow(s, t int) float64 {
	var flow float64
	for g.bfs(s, t) {
		g.next = make([]int, len(g.adj))
		for d := g.dfs(s, t, math.Inf(1)); d > eps; d = g.dfs(s, t, math.Inf(1)) {
			flow += d
		}
	}
	return flow
}

// reachable is the residual reachability from s, the same set for every
// maximum flow, which is what makes the reported cut canonical.
func (g *graph) reachable(s int) []bool {
	seen := make([]bool, len(g.adj))
	seen[s] = true
	for queue := []int{s}; len(queue) > 0; queue = queue[1:] {
		for _, e := range g.adj[queue[0]] {
			if e.cap > eps && !seen[e.to] {
				seen[e.to] = true
				queue = append(queue, e.to)
			}
		}
	}
	return seen
}
