package model

import (
	"fmt"

	"github.com/Roarge/sysml-federation/adapter/syntax"
)

// attrSlot is one attribute of a part or requirement: where it is declared and
// which usage, if any, binds its value.
type attrSlot struct {
	name string
	bind *syntax.AttributeUsage // carries Value, or nil when unbound
}

// partNode is a part usage during building.
type partNode struct {
	ast      *syntax.PartUsage
	defs     []*syntax.PartDef // most specific first
	parent   *partNode
	children []*partNode
	attrs    []attrSlot
	out      *Part
}

// reqNode is a requirement usage during building.
type reqNode struct {
	ast     *syntax.RequirementUsage
	def     *syntax.RequirementDef
	attrs   []attrSlot
	subject *partNode
	out     *Requirement
}

// vcNode is a verification usage during building.
type vcNode struct {
	ast *syntax.VerificationUsage
	def *syntax.VerificationDef
	out *VerificationCase
}

type builder struct {
	f         *syntax.File
	m         *Model
	partDefs  map[string]*syntax.PartDef
	portDefs  map[string]*syntax.PortDef
	reqDefs   map[string]*syntax.RequirementDef
	vcDefs    map[string]*syntax.VerificationDef
	ids       map[string]bool
	parts     []*partNode // every part, depth first
	roots     []*partNode
	byName    map[string]*partNode // package-level parts
	reqs      []*reqNode
	reqByName map[string]*reqNode
	vcs       []*vcNode
	values    map[*syntax.AttributeUsage]quantity
	busy      map[*syntax.AttributeUsage]bool
}

// fail refuses the model at a span. Callers format the message themselves,
// so no variadic empty interface appears in a declaration of this package.
func (b *builder) fail(at syntax.Span, msg string) {
	panic(b.f.ErrorAt(at, msg))
}

func build(f *syntax.File) *Model {
	b := &builder{f: f, m: &Model{Version: 1, Text: f.Src, file: f,
		parts: map[string]*Part{}, reqs: map[string]*Requirement{}, vcs: map[string]*VerificationCase{},
		literals: map[syntax.Span]int{}},
		partDefs: map[string]*syntax.PartDef{}, portDefs: map[string]*syntax.PortDef{},
		reqDefs: map[string]*syntax.RequirementDef{}, vcDefs: map[string]*syntax.VerificationDef{},
		ids: map[string]bool{}, byName: map[string]*partNode{}, reqByName: map[string]*reqNode{},
		values: map[*syntax.AttributeUsage]quantity{}, busy: map[*syntax.AttributeUsage]bool{}}
	pkg := &f.Package
	for i := range pkg.PartDefs {
		b.partDefs[pkg.PartDefs[i].Name] = &pkg.PartDefs[i]
	}
	for i := range pkg.PortDefs {
		b.portDefs[pkg.PortDefs[i].Name] = &pkg.PortDefs[i]
	}
	for i := range pkg.RequirementDefs {
		b.reqDefs[pkg.RequirementDefs[i].Name] = &pkg.RequirementDefs[i]
	}
	for i := range pkg.VerificationDefs {
		b.vcDefs[pkg.VerificationDefs[i].Name] = &pkg.VerificationDefs[i]
	}
	for i := range pkg.Parts {
		n := b.part(&pkg.Parts[i], nil, pkg.Name)
		b.roots = append(b.roots, n)
		if _, dup := b.byName[n.ast.Name]; dup {
			b.fail(n.ast.Span, fmt.Sprintf("part %q is declared twice", n.ast.Name))
		}
		b.byName[n.ast.Name] = n
		b.m.Roots = append(b.m.Roots, n.out)
	}
	b.requirements(pkg)
	b.verifications(pkg)
	b.link(pkg)     // subjects, connections, satisfies, derivations, verifications
	b.constraints() // quantity, comparison and limit
	b.project()     // attribute values, literal index
	b.exclusive()   // an edit may not move a literal a second value carries
	return b.m
}

// id applies AD-0018: the short name, else the qualified name.
func (b *builder) id(short, qualified string, at syntax.Span) string {
	id := short
	if id == "" {
		id = qualified
	}
	if b.ids[id] {
		b.fail(at, fmt.Sprintf("duplicate id %q", id))
	}
	b.ids[id] = true
	return id
}

// defChain resolves a part definition and its bases, most specific first.
func (b *builder) defChain(name string, at syntax.Span) []*syntax.PartDef {
	var chain []*syntax.PartDef
	seen := map[string]bool{}
	for name != "" {
		d, ok := b.partDefs[name]
		if !ok {
			b.fail(at, fmt.Sprintf("unresolved definition %q", name))
		}
		if seen[name] {
			b.fail(at, fmt.Sprintf("cyclic specialization of %q", chain[0].Name))
		}
		seen[name] = true
		chain = append(chain, d)
		name = d.Specializes
	}
	return chain
}

func (b *builder) part(u *syntax.PartUsage, parent *partNode, owner string) *partNode {
	n := &partNode{ast: u, parent: parent}
	qualified := owner + "::" + u.Name
	n.out = &Part{ID: b.id(u.ShortName, qualified, u.Span), ShortName: u.ShortName, Name: u.Name,
		Definition: u.Type, Doc: docText(u.Doc)}
	if u.Type != "" {
		n.defs = b.defChain(u.Type, u.Span)
	}
	// Attributes: declarations from the base definition down, then the usage's
	// own. A redefinition binds a slot that must already exist.
	for i := len(n.defs) - 1; i >= 0; i-- {
		for j := range n.defs[i].Attributes {
			b.slot(&n.attrs, &n.defs[i].Attributes[j], n.defs[i].Name)
		}
	}
	for j := range u.Attributes {
		b.slot(&n.attrs, &u.Attributes[j], u.Type)
	}
	// Ports: from the definitions, base first, then the usage's own.
	for i := len(n.defs) - 1; i >= 0; i-- {
		for _, p := range n.defs[i].Ports {
			n.out.Ports = append(n.out.Ports, Port{Name: p.Name, Direction: b.portDirection(p)})
		}
	}
	for _, p := range u.Ports {
		n.out.Ports = append(n.out.Ports, Port{Name: p.Name, Direction: b.portDirection(p)})
	}
	b.parts = append(b.parts, n)
	b.m.Parts = append(b.m.Parts, n.out)
	b.m.parts[n.out.ID] = n.out
	for i := range u.Parts {
		c := b.part(&u.Parts[i], n, qualified)
		n.children = append(n.children, c)
		n.out.Parts = append(n.out.Parts, c.out)
	}
	return n
}

// slot adds a declaration or binds a redefinition. owner names the definition
// a redefinition must find its declaration in, for the error message.
func (b *builder) slot(slots *[]attrSlot, a *syntax.AttributeUsage, owner string) {
	if a.Redefines {
		for i := range *slots {
			if (*slots)[i].name == a.Name {
				if a.Value != nil {
					(*slots)[i].bind = a
				}
				return
			}
		}
		b.fail(a.Span, fmt.Sprintf("redefined attribute %q is not declared by %q", a.Name, owner))
	}
	s := attrSlot{name: a.Name}
	if a.Value != nil {
		s.bind = a
	}
	*slots = append(*slots, s)
}

// portDirection derives a port's direction from its definition's items.
func (b *builder) portDirection(p syntax.PortUsage) Direction {
	d, ok := b.portDefs[p.Type]
	if !ok {
		b.fail(p.Span, fmt.Sprintf("unresolved definition %q", p.Type))
	}
	if len(d.Items) == 0 {
		return syntax.DirectionInOut
	}
	dir := d.Items[0].Direction
	for _, it := range d.Items[1:] {
		if it.Direction != dir {
			return syntax.DirectionInOut
		}
	}
	return dir
}

func docText(d *syntax.Doc) string {
	if d == nil {
		return ""
	}
	return d.Text
}

// exprSpan is the source range of an expression.
func exprSpan(e syntax.Expr) syntax.Span {
	switch x := e.(type) {
	case syntax.Literal:
		return x.Span
	case syntax.FeatureChain:
		return x.Span
	case syntax.Binary:
		return x.Span
	}
	return syntax.Span{}
}

// project fills every part's attributes from its slots and records every
// numeric literal.
func (b *builder) project() {
	for _, n := range b.parts {
		for _, s := range n.attrs {
			n.out.Attributes = append(n.out.Attributes, b.attribute(n, s))
		}
	}
}

func (b *builder) attribute(ctx *partNode, s attrSlot) Attribute {
	a := Attribute{Name: s.name}
	if s.bind == nil {
		return a
	}
	q := b.eval(ctx, nil, s.bind)
	a.Value, a.Unit = &q.num, q.unit
	if lit, ok := s.bind.Value.(syntax.Literal); ok {
		a.Editable, a.Span = true, lit.Span
		b.m.literals[lit.Span]++
	} else {
		a.Expression = b.f.Text(exprSpan(s.bind.Value))
	}
	return a
}

// exclusive clears the editable flag of every value whose literal another
// projected value shares. A definition may bind an attribute or hold a
// requirement's limit, and each of its usages then points at the same bytes,
// so an edit through one would silently move the others.
func (b *builder) exclusive() {
	for _, n := range b.parts {
		for i := range n.out.Attributes {
			if a := &n.out.Attributes[i]; a.Editable && b.m.literals[a.Span] > 1 {
				a.Editable, a.Span = false, syntax.Span{}
			}
		}
	}
	for _, n := range b.reqs {
		if n.out.LimitEditable && b.m.literals[n.out.limitSpan] > 1 {
			n.out.LimitEditable, n.out.limitSpan = false, syntax.Span{}
		}
	}
}

func (b *builder) requirements(pkg *syntax.Package) {
	for i := range pkg.Requirements {
		u := &pkg.Requirements[i]
		n := &reqNode{ast: u}
		if u.Type != "" {
			d, ok := b.reqDefs[u.Type]
			if !ok {
				b.fail(u.Span, fmt.Sprintf("unresolved definition %q", u.Type))
			}
			n.def = d
			for j := range d.Attributes {
				b.slot(&n.attrs, &d.Attributes[j], d.Name)
			}
		}
		for j := range u.Attributes {
			b.slot(&n.attrs, &u.Attributes[j], u.Type)
		}
		text := docText(u.Doc)
		if text == "" && n.def != nil {
			text = docText(n.def.Doc)
		}
		n.out = &Requirement{ID: b.id(u.ShortName, pkg.Name+"::"+u.Name, u.Span), ShortName: u.ShortName,
			Name: u.Name, Text: text}
		b.reqs = append(b.reqs, n)
		if _, dup := b.reqByName[u.Name]; dup {
			b.fail(u.Span, fmt.Sprintf("requirement %q is declared twice", u.Name))
		}
		b.reqByName[u.Name] = n
		b.m.Requirements = append(b.m.Requirements, n.out)
		b.m.reqs[n.out.ID] = n.out
	}
}

func (b *builder) verifications(pkg *syntax.Package) {
	for i := range pkg.Verifications {
		u := &pkg.Verifications[i]
		n := &vcNode{ast: u}
		if u.Type != "" {
			d, ok := b.vcDefs[u.Type]
			if !ok {
				b.fail(u.Span, fmt.Sprintf("unresolved definition %q", u.Type))
			}
			n.def = d
		}
		n.out = &VerificationCase{ID: b.id(u.ShortName, pkg.Name+"::"+u.Name, u.Span), ShortName: u.ShortName, Name: u.Name}
		b.vcs = append(b.vcs, n)
		b.m.VerificationCases = append(b.m.VerificationCases, n.out)
		b.m.vcs[n.out.ID] = n.out
	}
}

// link wires subjects, connections, satisfies, derivations and verifications
// once every element exists.
func (b *builder) link(pkg *syntax.Package) {
	for _, r := range b.reqs {
		if s := r.ast.Subject; s != nil && s.Value != nil {
			r.subject = b.findPart(*s.Value, nil)
			r.out.Subject = r.subject.out.ID
		}
	}
	for _, n := range b.parts {
		for _, c := range n.ast.Connects {
			b.connect(n, c)
		}
		for _, s := range n.ast.Satisfies {
			r := b.findReq(s.Requirement)
			by := b.findPart(s.By, n)
			r.out.SatisfiedBy = append(r.out.SatisfiedBy, by.out.ID)
			by.out.Satisfies = append(by.out.Satisfies, r.out.ID)
		}
	}
	for _, d := range pkg.Derivations {
		orig := b.findReq(d.Original)
		for _, c := range d.Derives {
			der := b.findReq(c)
			orig.out.Derives = append(orig.out.Derives, der.out.ID)
			der.out.DerivedFrom = append(der.out.DerivedFrom, orig.out.ID)
		}
	}
	for _, v := range b.vcs {
		var objectives []*syntax.Objective
		if v.def != nil && v.def.Objective != nil {
			objectives = append(objectives, v.def.Objective)
		}
		if v.ast.Objective != nil {
			objectives = append(objectives, v.ast.Objective)
		}
		for _, o := range objectives {
			r := b.findReq(o.Verify)
			v.out.Verifies = append(v.out.Verifies, r.out.ID)
			r.out.VerifiedBy = append(r.out.VerifiedBy, v.out.ID)
		}
	}
}
