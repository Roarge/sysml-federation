package syntax

import "strconv"

// parseExpr reads an expression. Task 1.4 adds the operator levels; until
// then it reads one primary.
func (p *parser) parseExpr() Expr { return p.parsePrimary() }

// parsePrimary reads a literal, a feature chain, or a parenthesised expression.
func (p *parser) parsePrimary() Expr {
	t := p.peek()
	switch {
	case t.Kind == Number || (t.Kind == Punct && t.Text == "-"):
		return p.parseLiteral()
	case t.Kind == Ident || t.Kind == Name:
		return p.parseChain()
	case t.Kind == Punct && t.Text == "(":
		p.next()
		e := p.parseExpr()
		p.expect(")")
		return e
	default:
		p.failExpected("a number, a feature chain or '('")
		return nil
	}
}

// parseLiteral reads `[-]number [ '[' qualifiedName ']' ]`. The span covers the
// sign and the digits only, so a patch replaces exactly them.
func (p *parser) parseLiteral() Literal {
	start := p.peek().Span.Start
	neg := p.accept("-")
	t := p.peek()
	if t.Kind != Number {
		p.failExpected("a number after '-'")
	}
	p.next()
	n, _ := strconv.ParseFloat(t.Text, 64) // the lexer already proved it parses
	if neg {
		n = -n
	}
	lit := Literal{Number: n, Span: p.end(start)}
	if p.accept("[") {
		lit.Unit = p.qualifiedName()
		p.expect("]")
	}
	return lit
}

// parseChain reads `a.b.c`.
func (p *parser) parseChain() FeatureChain {
	start := p.peek().Span.Start
	c := FeatureChain{Names: []string{p.name()}}
	for p.accept(".") {
		c.Names = append(c.Names, p.name())
	}
	c.Span = p.end(start)
	return c
}
