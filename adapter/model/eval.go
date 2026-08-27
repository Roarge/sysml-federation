package model

import (
	"fmt"
	"math"

	"github.com/Roarge/sysml-federation/adapter/syntax"
)

// quantity is an evaluated value with the unit it carries.
type quantity struct {
	num  float64
	unit string
}

// eval evaluates the value bound by a, in the scope of the part or requirement
// that owns it, with a guard against cycles. Results are memoised per binding
// and scope, so a definition's expression is evaluated once for each usage
// whose own values it reads.
func (b *builder) eval(part *partNode, req *reqNode, a *syntax.AttributeUsage) quantity {
	k := binding{a, part, req}
	if q, ok := b.values[k]; ok {
		return q
	}
	if b.busy[k] {
		b.fail(exprSpan(a.Value), fmt.Sprintf("cyclic binding of %q", a.Name))
	}
	b.busy[k] = true
	q := b.expr(part, req, a.Value)
	delete(b.busy, k)
	b.values[k] = q
	return q
}

func (b *builder) expr(part *partNode, req *reqNode, e syntax.Expr) quantity {
	switch x := e.(type) {
	case syntax.Literal:
		return quantity{x.Number, x.Unit}
	case syntax.FeatureChain:
		return b.chainValue(part, req, x)
	case syntax.Binary:
		l, r := b.expr(part, req, x.Left), b.expr(part, req, x.Right)
		unit := l.unit
		if unit == "" {
			unit = r.unit
		} else if r.unit != "" && r.unit != unit {
			b.fail(x.Span, fmt.Sprintf("units differ: %q and %q", l.unit, r.unit))
		}
		var n float64
		switch x.Op {
		case syntax.Add:
			n = l.num + r.num
		case syntax.Sub:
			n = l.num - r.num
		case syntax.Mul:
			n = l.num * r.num
		default:
			if r.num == 0 {
				b.fail(x.Span, "division by zero")
			}
			n = l.num / r.num
		}
		if math.IsInf(n, 0) || math.IsNaN(n) {
			b.fail(x.Span, "the result is not a finite number")
		}
		return quantity{n, unit}
	}
	return quantity{}
}

// chainValue resolves a chain to an attribute and evaluates it. The first
// segment is an attribute of the owner, the owner's subject (for a
// requirement), a child part, a part visible from an enclosing part, or a
// package-level part or requirement. Middle segments are child parts, and
// the last is an attribute.
func (b *builder) chainValue(part *partNode, req *reqNode, c syntax.FeatureChain) quantity {
	first, rest := c.Names[0], c.Names[1:]
	if slot, owner := b.ownAttr(part, req, first); slot != nil {
		if len(rest) > 0 {
			b.fail(c.Span, fmt.Sprintf("%q is not a part", first))
		}
		return b.slotValue(owner, req, slot, c.Span)
	}
	var cur *partNode
	var curReq *reqNode
	switch {
	case req != nil && req.subject != nil && req.ast.Subject != nil && first == req.ast.Subject.Name:
		cur = req.subject
	case b.reqByName[first] != nil:
		curReq = b.reqByName[first]
	default:
		cur = b.findPart(syntax.FeatureChain{Names: []string{first}, Span: c.Span}, part)
	}
	for i, name := range rest {
		last := i == len(rest)-1
		if curReq != nil {
			slot := findSlot(curReq.attrs, name)
			if slot == nil || !last {
				b.fail(c.Span, fmt.Sprintf("%q has no attribute %q", first, name))
			}
			return b.slotValue(nil, curReq, slot, c.Span)
		}
		if last {
			if slot := findSlot(cur.attrs, name); slot != nil {
				return b.slotValue(cur, nil, slot, c.Span)
			}
		}
		next := cur.child(name)
		if next == nil {
			b.fail(c.Span, fmt.Sprintf("%q has no attribute %q", cur.out.Name, name))
		}
		cur = next
	}
	b.fail(c.Span, fmt.Sprintf("%q is not an attribute", c.Names[len(c.Names)-1]))
	return quantity{}
}

// ownAttr finds name among the owner's own attribute slots: the requirement's
// when evaluating inside a requirement, the part's otherwise. The returned
// part is the slot's owner when it is a part, nil when it is the requirement.
func (b *builder) ownAttr(part *partNode, req *reqNode, name string) (*attrSlot, *partNode) {
	if req != nil {
		return findSlot(req.attrs, name), nil
	}
	if part != nil {
		return findSlot(part.attrs, name), part
	}
	return nil, nil
}

func findSlot(slots []attrSlot, name string) *attrSlot {
	for i := range slots {
		if slots[i].name == name {
			return &slots[i]
		}
	}
	return nil
}

func (b *builder) slotValue(part *partNode, req *reqNode, s *attrSlot, at syntax.Span) quantity {
	if s.bind == nil {
		b.fail(at, fmt.Sprintf("%q has no value", s.name))
	}
	return b.eval(part, req, s.bind)
}
