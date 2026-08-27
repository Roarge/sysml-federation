package syntax

import "strconv"

// isArrow reports whether the next two tokens are the '-' and '>' of SysML's
// '->' operator. The lexer has no '->' token, so the additive level asks here
// before it reads a '-' as subtraction, and `y->reduce min` is refused where
// the statement's terminator was due rather than inside the expression.
func (p *parser) isArrow() bool {
	if !p.is("-") {
		return false
	}
	next := p.toks[p.pos+1] // a '-' is never the last token: EOF follows it
	return next.Kind == Punct && next.Text == ">"
}

// parseExpr reads `term (('+'|'-') term)*`.
func (p *parser) parseExpr() Expr {
	start := p.peek().Span.Start
	left := p.parseTerm()
	for (p.is("+") || p.is("-")) && !p.isArrow() {
		op := Add
		if p.next().Text == "-" {
			op = Sub
		}
		left = Binary{Op: op, Left: left, Right: p.parseTerm(), Span: p.end(start)}
	}
	return left
}

// parseTerm reads `primary (('*'|'/') primary)*`.
func (p *parser) parseTerm() Expr {
	start := p.peek().Span.Start
	left := p.parsePrimary()
	for p.is("*") || p.is("/") {
		op := Mul
		if p.next().Text == "/" {
			op = Div
		}
		left = Binary{Op: op, Left: left, Right: p.parsePrimary(), Span: p.end(start)}
	}
	return left
}

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
