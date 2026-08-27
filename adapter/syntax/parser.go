package syntax

import (
	"fmt"
	"strings"
)

type parser struct {
	f    *File
	toks []Token
	pos  int
}

// Parse lexes and parses one file. The error is always a *Error.
func Parse(name, src string) (f *File, err error) {
	toks, err := Lex(name, src)
	if err != nil {
		return nil, err
	}
	p := &parser{f: &File{Name: name, Src: src}, toks: toks}
	defer func() {
		if r := recover(); r != nil {
			se, ok := r.(*Error)
			if !ok {
				panic(r)
			}
			f, err = nil, se
		}
	}()
	p.f.Package = p.parsePackage()
	if p.peek().Kind != EOF {
		p.failExpected("end of file")
	}
	return p.f, nil
}

// --- cursor -----------------------------------------------------------------

func (p *parser) peek() Token { return p.toks[p.pos] }

func (p *parser) next() Token {
	t := p.toks[p.pos]
	if t.Kind != EOF {
		p.pos++
	}
	return t
}

func (p *parser) fail(at Span, msg string) {
	panic(p.f.ErrorAt(at, msg))
}

func (p *parser) failExpected(what string) {
	p.fail(p.peek().Span, fmt.Sprintf("expected %s, found %s", what, describe(p.peek())))
}

func describe(t Token) string {
	switch t.Kind {
	case EOF:
		return "end of file"
	case Keyword:
		return fmt.Sprintf("keyword '%s'", t.Text)
	case Ident:
		return fmt.Sprintf("identifier '%s'", t.Text)
	case Name:
		return fmt.Sprintf("name '%s'", t.Text)
	case Number:
		return fmt.Sprintf("number '%s'", t.Text)
	case String:
		return "string"
	case Comment:
		return "comment"
	default:
		return fmt.Sprintf("'%s'", t.Text)
	}
}

// is reports whether the next token is the given keyword or punctuation.
func (p *parser) is(text string) bool {
	t := p.peek()
	return (t.Kind == Keyword || t.Kind == Punct) && t.Text == text
}

// accept consumes the next token if it is the given keyword or punctuation.
func (p *parser) accept(text string) bool {
	if p.is(text) {
		p.next()
		return true
	}
	return false
}

// expect consumes the given keyword or punctuation or fails.
func (p *parser) expect(text string) Token {
	if !p.is(text) {
		if keywords[text] {
			p.failExpected(fmt.Sprintf("keyword '%s'", text))
		}
		p.failExpected(fmt.Sprintf("'%s'", text))
	}
	return p.next()
}

// name consumes an identifier or a quoted name.
func (p *parser) name() string {
	t := p.peek()
	if t.Kind != Ident && t.Kind != Name {
		p.failExpected("a name")
	}
	return p.next().Text
}

// qualifiedName consumes `A::B::C` and returns it as written.
func (p *parser) qualifiedName() string {
	parts := []string{p.name()}
	for p.is("::") && p.toks[p.pos+1].Kind != Punct {
		p.next()
		parts = append(parts, p.name())
	}
	return strings.Join(parts, "::")
}

// shortName consumes an optional `<'x'>` or `<x>`.
func (p *parser) shortName() string {
	if !p.accept("<") {
		return ""
	}
	s := p.name()
	p.expect(">")
	return s
}

// end returns the span from start to the end of the last consumed token.
func (p *parser) end(start int) Span { return Span{start, p.toks[p.pos-1].Span.End} }

// article is the indefinite article for a body name, so the refusal reads
// "an attribute body" rather than "a attribute body".
func article(body string) string {
	switch body[0] {
	case 'a', 'e', 'i', 'o', 'u':
		return "an"
	}
	return "a"
}

// unsupported refuses the next token as a member of the named body.
func (p *parser) unsupported(body string) {
	t := p.peek()
	if t.Kind == Keyword {
		p.fail(t.Span, fmt.Sprintf("keyword '%s' is not supported in %s %s body", t.Text, article(body), body))
	}
	p.failExpected(fmt.Sprintf("a member of %s %s body", article(body), body))
}

// once refuses a second occurrence of a singular member.
func (p *parser) once(present bool, what, body string) {
	if present {
		p.fail(p.peek().Span, fmt.Sprintf("only one %s is allowed in a %s body", what, body))
	}
}

// --- docs and comments ------------------------------------------------------

func (p *parser) doc() *Doc {
	start := p.expect("doc").Span.Start
	t := p.peek()
	if t.Kind != Comment {
		p.failExpected("a /* */ comment")
	}
	p.next()
	return &Doc{Text: docText(t.Text), Span: p.end(start)}
}

// docText normalises a comment body: each line loses surrounding whitespace
// and one leading asterisk, empty lines drop, and the rest join with a space.
func docText(raw string) string {
	var lines []string
	for _, l := range strings.Split(raw, "\n") {
		l = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "*"))
		if l != "" {
			lines = append(lines, l)
		}
	}
	return strings.Join(lines, " ")
}

// skipComment drops a bare /* */ comment used as a statement.
func (p *parser) skipComment() bool {
	if p.peek().Kind == Comment {
		p.next()
		return true
	}
	return false
}

// --- package ----------------------------------------------------------------

func (p *parser) parsePackage() Package {
	start := p.expect("package").Span.Start
	pkg := Package{ShortName: p.shortName(), Name: p.name()}
	p.expect("{")
	for !p.is("}") && p.peek().Kind != EOF {
		switch {
		case p.skipComment():
		case p.is("doc"):
			p.once(pkg.Doc != nil, "doc", "package")
			pkg.Doc = p.doc()
		case p.is("private") || p.is("public") || p.is("import"):
			pkg.Imports = append(pkg.Imports, p.parseImport())
		case p.is("attribute"):
			pkg.Attributes = append(pkg.Attributes, p.parseAttribute())
		case p.is("item"):
			pkg.ItemDefs = append(pkg.ItemDefs, p.parseItemDef())
		case p.is("port"):
			pkg.PortDefs = append(pkg.PortDefs, p.parsePortDef())
		case p.is("abstract") || (p.is("part") && p.toks[p.pos+1].Text == "def"):
			pkg.PartDefs = append(pkg.PartDefs, p.parsePartDef())
		case p.is("part"):
			pkg.Parts = append(pkg.Parts, p.parsePartUsage())
		case p.is("requirement") && p.toks[p.pos+1].Text == "def":
			pkg.RequirementDefs = append(pkg.RequirementDefs, p.parseRequirementDef())
		case p.is("requirement"):
			pkg.Requirements = append(pkg.Requirements, p.parseRequirementUsage())
		case p.is("verification") && p.toks[p.pos+1].Text == "def":
			pkg.VerificationDefs = append(pkg.VerificationDefs, p.parseVerificationDef())
		case p.is("verification"):
			pkg.Verifications = append(pkg.Verifications, p.parseVerificationUsage())
		case p.is("#"):
			pkg.Derivations = append(pkg.Derivations, p.parseDerivation())
		default:
			p.unsupported("package")
		}
	}
	p.expect("}")
	pkg.Span = p.end(start)
	return pkg
}

func (p *parser) parseImport() Import {
	start := p.peek().Span.Start
	if !p.accept("private") {
		p.accept("public")
	}
	p.expect("import")
	imp := Import{Path: p.qualifiedName()}
	if p.accept("::") {
		p.expect("*")
		imp.All = true
	}
	p.expect(";")
	imp.Span = p.end(start)
	return imp
}

func (p *parser) parseItemDef() ItemDef {
	start := p.expect("item").Span.Start
	p.expect("def")
	d := ItemDef{ShortName: p.shortName(), Name: p.name()}
	p.expect(";")
	d.Span = p.end(start)
	return d
}

// --- parts ------------------------------------------------------------------

func (p *parser) parsePartDef() PartDef {
	start := p.peek().Span.Start
	d := PartDef{Abstract: p.accept("abstract")}
	p.expect("part")
	p.expect("def")
	d.ShortName, d.Name = p.shortName(), p.name()
	if p.accept(":>") {
		d.Specializes = p.qualifiedName()
	}
	if p.accept("{") {
		for !p.is("}") && p.peek().Kind != EOF {
			switch {
			case p.skipComment():
			case p.is("doc"):
				p.once(d.Doc != nil, "doc", "part definition")
				d.Doc = p.doc()
			case p.is("attribute"):
				d.Attributes = append(d.Attributes, p.parseAttribute())
			case p.is("port"):
				d.Ports = append(d.Ports, p.parsePortUsage())
			case p.is("part"):
				d.Parts = append(d.Parts, p.parsePartUsage())
			default:
				p.unsupported("part definition")
			}
		}
		p.expect("}")
	} else {
		p.expect(";")
	}
	d.Span = p.end(start)
	return d
}

func (p *parser) parsePartUsage() PartUsage {
	start := p.expect("part").Span.Start
	u := PartUsage{ShortName: p.shortName(), Name: p.name()}
	if p.accept(":") {
		u.Type = p.qualifiedName()
	}
	if p.accept("{") {
		for !p.is("}") && p.peek().Kind != EOF {
			switch {
			case p.skipComment():
			case p.is("doc"):
				p.once(u.Doc != nil, "doc", "part")
				u.Doc = p.doc()
			case p.is("attribute"):
				u.Attributes = append(u.Attributes, p.parseAttribute())
			case p.is("port"):
				u.Ports = append(u.Ports, p.parsePortUsage())
			case p.is("part"):
				u.Parts = append(u.Parts, p.parsePartUsage())
			case p.is("connect"):
				u.Connects = append(u.Connects, p.parseConnect())
			case p.is("satisfy"):
				u.Satisfies = append(u.Satisfies, p.parseSatisfy())
			default:
				p.unsupported("part")
			}
		}
		p.expect("}")
	} else {
		p.expect(";")
	}
	u.Span = p.end(start)
	return u
}

// parseAttribute reads `attribute [<sn>] name [: T] [= expr] (; | { body })`
// or `attribute (:>> | redefines) name [: T] [= expr] (; | { body })`. Inside
// an attribute body the leading keyword may be omitted, so the body loop
// calls attributeRest directly.
func (p *parser) parseAttribute() AttributeUsage {
	start := p.expect("attribute").Span.Start
	return p.attributeRest(start)
}

func (p *parser) attributeRest(start int) AttributeUsage {
	var a AttributeUsage
	if p.accept(":>>") || p.accept("redefines") {
		a.Redefines = true
	} else {
		a.ShortName = p.shortName()
	}
	a.Name = p.name()
	if p.accept(":") {
		a.Type = p.qualifiedName()
	}
	if p.accept("=") {
		a.Value = p.parseExpr()
	}
	if p.accept("{") {
		for !p.is("}") && p.peek().Kind != EOF {
			if p.skipComment() {
				continue
			}
			inner := p.peek().Span.Start
			if !p.is(":>>") && !p.is("redefines") && !p.is("attribute") {
				p.unsupported("attribute")
			}
			p.accept("attribute")
			a.Body = append(a.Body, p.attributeRest(inner))
		}
		p.expect("}")
	} else {
		p.expect(";")
	}
	a.Span = p.end(start)
	return a
}

// --- ports and connections ----------------------------------------------------

func (p *parser) parsePortDef() PortDef {
	start := p.expect("port").Span.Start
	p.expect("def")
	d := PortDef{ShortName: p.shortName(), Name: p.name()}
	if !p.accept(";") {
		p.expect("{")
		for !p.is("}") {
			switch {
			case p.skipComment():
			case p.is("doc"):
				p.once(d.Doc != nil, "doc", "port definition")
				d.Doc = p.doc()
			case p.is("in") || p.is("out") || p.is("inout"):
				d.Items = append(d.Items, p.parseDirectedItem())
			default:
				p.unsupported("port definition")
			}
		}
		p.expect("}")
	}
	d.Span = p.end(start)
	return d
}

func (p *parser) parseDirectedItem() DirectedItem {
	t := p.next()
	dir := map[string]Direction{"in": DirectionIn, "out": DirectionOut, "inout": DirectionInOut}[t.Text]
	p.expect("item")
	it := DirectedItem{Direction: dir, Name: p.name()}
	if p.accept(":") {
		it.Type = p.qualifiedName()
	}
	p.expect(";")
	it.Span = p.end(t.Span.Start)
	return it
}

func (p *parser) parsePortUsage() PortUsage {
	start := p.expect("port").Span.Start
	u := PortUsage{ShortName: p.shortName(), Name: p.name()}
	p.expect(":")
	u.Type = p.qualifiedName()
	p.expect(";")
	u.Span = p.end(start)
	return u
}

func (p *parser) parseConnect() Connect {
	start := p.expect("connect").Span.Start
	from := p.parseChain()
	p.expect("to")
	to := p.parseChain()
	for _, c := range []FeatureChain{from, to} {
		if len(c.Names) != 2 {
			p.fail(c.Span, "connect ends must be written as part.port")
		}
	}
	p.expect(";")
	return Connect{From: from, To: to, Span: p.end(start)}
}

func (p *parser) parseSatisfy() Satisfy {
	start := p.expect("satisfy").Span.Start
	req := p.parseChain()
	p.expect("by")
	by := p.parseChain()
	p.expect(";")
	return Satisfy{Requirement: req, By: by, Span: p.end(start)}
}

// --- requirements -------------------------------------------------------------

func (p *parser) parseRequirementDef() RequirementDef {
	start := p.expect("requirement").Span.Start
	p.expect("def")
	d := RequirementDef{ShortName: p.shortName(), Name: p.name()}
	if p.accept(":>") {
		d.Specializes = p.qualifiedName()
	}
	d.RequirementBody = p.requirementBody()
	d.Span = p.end(start)
	return d
}

func (p *parser) parseRequirementUsage() RequirementUsage {
	start := p.expect("requirement").Span.Start
	u := RequirementUsage{ShortName: p.shortName(), Name: p.name()}
	if p.accept(":") {
		u.Type = p.qualifiedName()
	}
	u.RequirementBody = p.requirementBody()
	u.Span = p.end(start)
	return u
}

func (p *parser) requirementBody() RequirementBody {
	var b RequirementBody
	p.expect("{")
	for !p.is("}") {
		switch {
		case p.skipComment():
		case p.is("doc"):
			p.once(b.Doc != nil, "doc", "requirement")
			b.Doc = p.doc()
		case p.is("subject"):
			p.once(b.Subject != nil, "subject", "requirement")
			b.Subject = p.parseSubject()
		case p.is("attribute"):
			b.Attributes = append(b.Attributes, p.parseAttribute())
		case p.is("require"):
			p.once(b.Constraint != nil, "require constraint", "requirement")
			b.Constraint = p.parseConstraint()
		default:
			p.unsupported("requirement")
		}
	}
	p.expect("}")
	return b
}

// parseSubject reads `subject name : Type;`, `subject :>> name = chain;` or
// `subject name :> chain;`.
func (p *parser) parseSubject() *Subject {
	start := p.expect("subject").Span.Start
	s := &Subject{}
	if p.accept(":>>") {
		s.Name = p.name()
		p.expect("=")
		c := p.parseChain()
		s.Value = &c
	} else {
		s.Name = p.name()
		if p.accept(":>") {
			c := p.parseChain()
			s.Value = &c
		} else {
			p.expect(":")
			s.Type = p.qualifiedName()
		}
	}
	p.expect(";")
	s.Span = p.end(start)
	return s
}

func (p *parser) parseConstraint() *RequireConstraint {
	start := p.expect("require").Span.Start
	p.expect("constraint")
	p.expect("{")
	c := &RequireConstraint{Left: p.parseOperand()}
	c.Op = p.parseComparison()
	c.Right = p.parseOperand()
	p.expect("}")
	c.Span = p.end(start)
	return c
}

// --- verification -------------------------------------------------------------

func (p *parser) parseVerificationDef() VerificationDef {
	start := p.expect("verification").Span.Start
	p.expect("def")
	d := VerificationDef{ShortName: p.shortName(), Name: p.name()}
	d.VerificationBody = p.verificationBody()
	d.Span = p.end(start)
	return d
}

func (p *parser) parseVerificationUsage() VerificationUsage {
	start := p.expect("verification").Span.Start
	u := VerificationUsage{ShortName: p.shortName(), Name: p.name()}
	if p.accept(":") {
		u.Type = p.qualifiedName()
	}
	u.VerificationBody = p.verificationBody()
	u.Span = p.end(start)
	return u
}

func (p *parser) verificationBody() VerificationBody {
	var b VerificationBody
	p.expect("{")
	for !p.is("}") {
		switch {
		case p.skipComment():
		case p.is("doc"):
			p.once(b.Doc != nil, "doc", "verification")
			b.Doc = p.doc()
		case p.is("subject"):
			p.once(b.Subject != nil, "subject", "verification")
			b.Subject = p.parseSubject()
		case p.is("objective"):
			p.once(b.Objective != nil, "objective", "verification")
			b.Objective = p.parseObjective()
		default:
			p.unsupported("verification")
		}
	}
	p.expect("}")
	return b
}

func (p *parser) parseObjective() *Objective {
	start := p.expect("objective").Span.Start
	p.expect("{")
	p.expect("verify")
	o := &Objective{Verify: p.parseChain()}
	p.expect(";")
	p.expect("}")
	o.Span = p.end(start)
	return o
}

// --- derivation ---------------------------------------------------------------

func (p *parser) metadata() (Token, string) {
	hash := p.expect("#")
	t := p.peek()
	if t.Kind != Ident {
		p.failExpected("a metadata name after '#'")
	}
	p.next()
	return hash, t.Text
}

func (p *parser) parseDerivation() DerivationConnection {
	hash, kind := p.metadata()
	if kind != "derivation" {
		p.fail(Span{hash.Span.Start, p.toks[p.pos-1].Span.End}, "unknown metadata #"+kind)
	}
	p.expect("connection")
	p.expect("{")
	var d DerivationConnection
	originals := 0
	for !p.is("}") {
		if p.skipComment() {
			continue
		}
		p.expect("end")
		endHash, endKind := p.metadata()
		p.expect("::>")
		c := p.parseChain()
		p.expect(";")
		switch endKind {
		case "original":
			originals++
			d.Original = c
		case "derive":
			d.Derives = append(d.Derives, c)
		default:
			p.fail(Span{endHash.Span.Start, c.Span.End}, "unknown metadata #"+endKind)
		}
	}
	p.expect("}")
	d.Span = p.end(hash.Span.Start)
	if originals != 1 {
		p.fail(d.Span, "a derivation connection needs exactly one #original end")
	}
	if len(d.Derives) == 0 {
		p.fail(d.Span, "a derivation connection needs at least one #derive end")
	}
	return d
}
