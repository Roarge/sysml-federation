package model

import (
	"fmt"

	"github.com/Roarge/sysml-federation/adapter/syntax"
)

// findPart resolves a chain to a part. The first segment is looked up among
// the children of scope and of each enclosing part outward, then among the
// package-level parts. Each later segment is a child of the part before it.
func (b *builder) findPart(c syntax.FeatureChain, scope *partNode) *partNode {
	var cur *partNode
	for s := scope; s != nil && cur == nil; s = s.parent {
		cur = s.child(c.Names[0])
	}
	if cur == nil {
		cur = b.byName[c.Names[0]]
	}
	if cur == nil {
		if _, isReq := b.reqByName[c.Names[0]]; isReq {
			b.fail(c.Span, fmt.Sprintf("%q is not a part", c.Names[0]))
		}
		b.fail(c.Span, fmt.Sprintf("unresolved name %q", c.Names[0]))
	}
	for _, name := range c.Names[1:] {
		next := cur.child(name)
		if next == nil {
			b.fail(c.Span, fmt.Sprintf("%q has no part %q", cur.out.Name, name))
		}
		cur = next
	}
	return cur
}

func (n *partNode) child(name string) *partNode {
	for _, c := range n.children {
		if c.ast.Name == name {
			return c
		}
	}
	return nil
}

// findReq resolves a single-segment chain to a package-level requirement usage.
func (b *builder) findReq(c syntax.FeatureChain) *reqNode {
	if len(c.Names) == 1 {
		if r, ok := b.reqByName[c.Names[0]]; ok {
			return r
		}
		if _, isPart := b.byName[c.Names[0]]; isPart {
			b.fail(c.Span, fmt.Sprintf("%q is not a requirement", c.Names[0]))
		}
	}
	b.fail(c.Span, fmt.Sprintf("unresolved name %q", c.Names[0]))
	return nil
}

// portEnd resolves `child.port` within owner and returns the child and the port.
func (b *builder) portEnd(c syntax.FeatureChain, owner *partNode) (*partNode, Port) {
	child := owner.child(c.Names[0])
	if child == nil {
		b.fail(c.Span, fmt.Sprintf("unresolved name %q", c.Names[0]))
	}
	for _, p := range child.out.Ports {
		if p.Name == c.Names[1] {
			return child, p
		}
	}
	b.fail(c.Span, fmt.Sprintf("%q has no port %q", c.Names[0], c.Names[1]))
	return nil, Port{}
}

func (b *builder) connect(owner *partNode, c syntax.Connect) {
	from, fp := b.portEnd(c.From, owner)
	if fp.Direction != syntax.DirectionOut {
		b.fail(c.From.Span, fmt.Sprintf("first end %q is not an out port", b.f.Text(c.From.Span)))
	}
	to, tp := b.portEnd(c.To, owner)
	if tp.Direction != syntax.DirectionIn {
		b.fail(c.To.Span, fmt.Sprintf("second end %q is not an in port", b.f.Text(c.To.Span)))
	}
	conn := Connection{From: from.out.ID, FromPort: fp.Name, To: to.out.ID, ToPort: tp.Name}
	conn.ID = conn.From + "." + conn.FromPort + "->" + conn.To + "." + conn.ToPort
	for _, existing := range owner.out.Connections {
		if existing.ID == conn.ID {
			b.fail(c.Span, fmt.Sprintf("duplicate connection %q", conn.ID))
		}
	}
	owner.out.Connections = append(owner.out.Connections, conn)
}
