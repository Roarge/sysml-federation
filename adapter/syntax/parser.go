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

// unsupported refuses the next token as a member of the named body.
func (p *parser) unsupported(body string) {
	t := p.peek()
	if t.Kind == Keyword {
		p.fail(t.Span, fmt.Sprintf("keyword '%s' is not supported in a %s body", t.Text, body))
	}
	p.failExpected(fmt.Sprintf("a member of a %s body", body))
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

// --- stubs replaced in Tasks 1.4 and 1.5 ------------------------------------

func (p *parser) parsePortDef() PortDef     { p.unsupported("package"); return PortDef{} }
func (p *parser) parsePortUsage() PortUsage { p.unsupported("part definition"); return PortUsage{} }
func (p *parser) parseConnect() Connect     { p.unsupported("part"); return Connect{} }
func (p *parser) parseSatisfy() Satisfy     { p.unsupported("part"); return Satisfy{} }
func (p *parser) parseRequirementDef() RequirementDef {
	p.unsupported("package")
	return RequirementDef{}
}
func (p *parser) parseRequirementUsage() RequirementUsage {
	p.unsupported("package")
	return RequirementUsage{}
}
func (p *parser) parseVerificationDef() VerificationDef {
	p.unsupported("package")
	return VerificationDef{}
}
func (p *parser) parseVerificationUsage() VerificationUsage {
	p.unsupported("package")
	return VerificationUsage{}
}
func (p *parser) parseDerivation() DerivationConnection {
	p.unsupported("package")
	return DerivationConnection{}
}
