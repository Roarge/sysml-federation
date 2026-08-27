package model

import (
	"fmt"

	"github.com/Roarge/sysml-federation/adapter/syntax"
)

// constraints reads each requirement's require constraint into quantity,
// comparison and limit (AD-0008): one operand is a feature chain rooted at the
// subject's name, the other is an attribute of the requirement or a literal.
func (b *builder) constraints() {
	for _, r := range b.reqs {
		b.constraint(r)
	}
}

func (b *builder) constraint(r *reqNode) {
	c := r.ast.Constraint
	if c == nil && r.def != nil {
		c = r.def.Constraint
	}
	if c == nil {
		b.fail(r.ast.Span, fmt.Sprintf("requirement %q has no require constraint", r.ast.Name))
	}
	subjectName := ""
	if r.ast.Subject != nil {
		subjectName = r.ast.Subject.Name
	} else if r.def != nil && r.def.Subject != nil {
		subjectName = r.def.Subject.Name
	}
	if subjectName == "" {
		b.fail(r.ast.Span, fmt.Sprintf("requirement %q has no subject", r.ast.Name))
	}
	leftRooted, rightRooted := rooted(c.Left, subjectName), rooted(c.Right, subjectName)
	switch {
	case leftRooted && rightRooted:
		b.fail(c.Span, "the constraint compares two operands rooted at the subject")
	case !leftRooted && !rightRooted:
		b.fail(c.Span, fmt.Sprintf("no operand is a feature chain rooted at the subject %q", subjectName))
	}
	side, other, op := c.Left, c.Right, c.Op
	if rightRooted {
		side, other, op = c.Right, c.Left, flip(c.Op)
	}
	chain := side.(syntax.FeatureChain) // rooted has proved the type
	r.out.Quantity = chain.Names[len(chain.Names)-1]
	r.out.Comparison = op
	if r.subject != nil {
		b.checkQuantity(r.subject, chain)
	}
	switch o := other.(type) {
	case syntax.Literal:
		b.setLimit(r, quantity{o.Number, o.Unit}, &o)
	case syntax.FeatureChain:
		slot := (*attrSlot)(nil)
		if len(o.Names) == 1 {
			slot = findSlot(r.attrs, o.Names[0])
		}
		if slot == nil {
			b.fail(o.Span, fmt.Sprintf("%q is neither an attribute of the requirement nor a literal", b.f.Text(o.Span)))
		}
		if slot.bind == nil {
			b.fail(r.ast.Span, fmt.Sprintf("limit attribute %q of requirement %q has no value", slot.name, r.ast.Name))
		}
		lit, isLit := slot.bind.Value.(syntax.Literal)
		if isLit {
			b.setLimit(r, quantity{lit.Number, lit.Unit}, &lit)
		} else {
			b.setLimit(r, b.eval(nil, r, slot.bind), nil)
		}
	}
}

func (b *builder) setLimit(r *reqNode, q quantity, lit *syntax.Literal) {
	r.out.Limit, r.out.LimitUnit = q.num, q.unit
	if lit != nil {
		r.out.LimitEditable, r.out.limitSpan = true, lit.Span
		b.m.literals[lit.Span] = true
	}
}

// rooted reports whether e is a chain of at least two segments starting at name.
func rooted(e syntax.Expr, name string) bool {
	c, ok := e.(syntax.FeatureChain)
	return ok && len(c.Names) >= 2 && c.Names[0] == name
}

// flip mirrors a comparison for the operand order being swapped.
func flip(c syntax.Comparison) syntax.Comparison {
	switch c {
	case syntax.GE:
		return syntax.LE
	case syntax.LE:
		return syntax.GE
	case syntax.GT:
		return syntax.LT
	case syntax.LT:
		return syntax.GT
	}
	return c
}

// checkQuantity confirms the subject declares the quantity the chain names,
// walking child parts for the middle segments.
func (b *builder) checkQuantity(subject *partNode, chain syntax.FeatureChain) {
	cur := subject
	middle, last := chain.Names[1:len(chain.Names)-1], chain.Names[len(chain.Names)-1]
	for _, name := range middle {
		next := cur.child(name)
		if next == nil {
			b.fail(chain.Span, fmt.Sprintf("subject %q has no part %q", subject.out.ID, name))
		}
		cur = next
	}
	if findSlot(cur.attrs, last) == nil {
		b.fail(chain.Span, fmt.Sprintf("subject %q declares no attribute %q", cur.out.ID, last))
	}
}
