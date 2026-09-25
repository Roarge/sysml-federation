package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// The wiki is the demo's systems model read as pages and links, the way a
// person reads a wiki: every element is a page, and every relationship the
// systems engineer wrote is a link that shows on both of its ends. It reads
// the SysML v2 text itself, with a reader that knows the forms the demo's
// systems model uses, and records every reference it cannot resolve.

// Attr is one attribute value an element carries.
type Attr struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Meta is one metadata annotation, such as @Evidence or @DecisionRecord.
type Meta struct {
	Type   string `json:"type"`
	Fields []Attr `json:"fields"`
}

// Field is the value of one of the annotation's fields.
func (m Meta) Field(name string) string {
	for _, f := range m.Fields {
		if f.Name == name {
			return f.Value
		}
	}
	return ""
}

// Element is one page of the wiki.
type Element struct {
	ID    string // short and unique: the short name, or the shortest telling suffix of QName
	QName string // qualified name, segments joined by ::
	Kind  string // the declaring keywords, such as "part def" or "exhibit state"
	Name  string
	Short string // the short name, such as SR-04
	Doc   string // the first doc comment
	Attrs []Attr
	Meta  []Meta
	Owner string // the owner's qualified name
	File  string // the model file, relative to the checkout
	Line  int

	owner    *Element
	children map[string]*Element
	order    []*Element
	imports  []importSpec
	types    []*Element // typed by, specialises and subsets, once resolved
}

type importSpec struct {
	target string
	star   bool
}

// Attr gives the value of one attribute, or "".
func (e *Element) Attr(name string) string {
	for _, a := range e.Attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

// Annotations gives the element's metadata of one type.
func (e *Element) Annotations(typ string) []Meta {
	var out []Meta
	for _, m := range e.Meta {
		if m.Type == typ {
			out = append(out, m)
		}
	}
	return out
}

// Decisions are the decision records the element names.
func (e *Element) Decisions() []Meta { return e.Annotations("DecisionRecord") }

// Evidence are the evidence locations the element names.
func (e *Element) Evidence() []Meta { return e.Annotations("Evidence") }

func (e *Element) isPackage() bool { return e.Kind == "package" }

// LinkView is one link as an element shows it.
type LinkView struct {
	Rel   string
	Other string // the other end's identifier
	Via   string // the connector or derivation the link runs through, if any
}

// LinkRef names one link by the qualified names of its ends.
type LinkRef struct {
	From string `json:"from"`
	Rel  string `json:"rel"`
	To   string `json:"to"`
}

// Removal is what an incident probe takes out of the wiki.
type Removal struct {
	Elements []string
	Links    []LinkRef
}

type edge struct {
	rel   string
	other *Element
	via   *Element
}

// Wiki is the systems model read as pages and links.
type Wiki struct {
	Elements   []*Element
	Unresolved []string // every reference no element answers, as file:line: text

	byID    map[string]*Element
	byQName map[string]*Element
	links   map[*Element][]edge
}

// relationships lists every link name with its reverse, in the order an
// element's links are shown.
var relationships = [][2]string{
	{"owned by", "owns"},
	{"typed by", "type of"},
	{"specialises", "specialised by"},
	{"satisfies", "satisfied by"},
	{"verifies", "verified by"},
	{"derives", "derived from"},
	{"performs", "performed by"},
	{"exhibits", "exhibited by"},
	{"includes", "included by"},
	{"allocated to", "allocated from"},
	{"bound to", "bound to"},
	{"connected to", "connected to"},
	{"flows to", "flows from"},
	{"carries", "carried by"},
	{"from", "left by"},
	{"to", "entered by"},
	{"accepts", "accepted by"},
	{"starts in", "start of"},
	{"then", "after"},
	{"subject", "subject of"},
	{"actor", "actor in"},
	{"stakeholder", "stakeholder in"},
	{"frames", "framed by"},
	{"exposes", "exposed by"},
}

var (
	reverseOf = map[string]string{}
	relOrder  = map[string]int{}
)

func init() {
	n := 0
	for _, r := range relationships {
		reverseOf[r[0]], reverseOf[r[1]] = r[1], r[0]
		relOrder[r[0]] = n
		n++
		if r[1] != r[0] {
			relOrder[r[1]] = n
			n++
		}
	}
	// owns goes last, since a package owns a long list
	relOrder["owns"] = n + 1
}

// Reverse gives the name a link shows on its other end.
func Reverse(rel string) string { return reverseOf[rel] }

// LoadWiki reads model/library and model/core under root.
func LoadWiki(root string) (*Wiki, error) {
	files := map[string]string{}
	var names []string
	for _, dir := range []string{"model/library", "model/core"} {
		base := filepath.Join(root, filepath.FromSlash(dir))
		if _, err := os.Stat(base); err != nil {
			continue
		}
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".sysml") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			files[rel] = string(data)
			names = append(names, rel)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no systems model under %s/model", root)
	}
	sort.Strings(names)
	b := newBuilder()
	for _, name := range names {
		if err := b.parseFile(name, files[name]); err != nil {
			return nil, err
		}
	}
	return b.finish(), nil
}

// Get finds an element by its identifier or its qualified name. It forgives
// the usual slips: a dot where the model has ::, quotes around the name, or a
// suffix of a qualified name that only one element has.
func (w *Wiki) Get(ref string) (*Element, bool) {
	ref = strings.Trim(strings.TrimSpace(ref), "`'\"")
	if ref == "" {
		return nil, false
	}
	if e, ok := w.byID[ref]; ok {
		return e, true
	}
	if e, ok := w.byQName[ref]; ok {
		return e, true
	}
	if dotted := strings.ReplaceAll(ref, ".", "::"); dotted != ref {
		if e, ok := w.Get(dotted); ok {
			return e, true
		}
	}
	var found *Element
	for _, e := range w.Elements {
		if strings.HasSuffix(e.QName, "::"+ref) {
			if found != nil {
				return nil, false
			}
			found = e
		}
	}
	return found, found != nil
}

// LinksOf gives an element's links, in the order they are shown.
func (w *Wiki) LinksOf(id string) []LinkView {
	e, ok := w.Get(id)
	if !ok {
		return nil
	}
	var out []LinkView
	for _, l := range w.links[e] {
		v := LinkView{Rel: l.rel, Other: l.other.ID}
		if l.via != nil {
			v.Via = l.via.ID
		}
		out = append(out, v)
	}
	return out
}

// linked reports whether any link joins a and b, in either direction.
func (w *Wiki) linked(a, b *Element) bool {
	for _, l := range w.links[a] {
		if l.other == b {
			return true
		}
	}
	return false
}

// Hash identifies what the wiki holds, so two runs can be seen to have read
// the same systems model.
func (w *Wiki) Hash() string {
	h := sha256.New()
	for _, e := range w.Elements {
		fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s\x00%s\n", e.ID, e.QName, e.Kind, e.Short, e.Doc)
		for _, a := range e.Attrs {
			fmt.Fprintf(h, "a\x00%s\x00%s\n", a.Name, a.Value)
		}
		for _, m := range e.Meta {
			fmt.Fprintf(h, "m\x00%s", m.Type)
			for _, f := range m.Fields {
				fmt.Fprintf(h, "\x00%s=%s", f.Name, f.Value)
			}
			fmt.Fprintln(h)
		}
		for _, l := range w.LinksOf(e.ID) {
			fmt.Fprintf(h, "l\x00%s\x00%s\x00%s\n", l.Rel, l.Other, l.Via)
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Without gives a copy of the wiki with elements or links taken out. An
// element goes with everything it owns and every link that touches them.
// Every other element keeps its identifier, so answers from the copy can be
// compared with answers from the original.
func (w *Wiki) Without(r Removal) *Wiki {
	gone := map[*Element]bool{}
	for _, ref := range r.Elements {
		if e, ok := w.Get(ref); ok {
			for _, x := range w.Elements {
				if x == e || strings.HasPrefix(x.QName, e.QName+"::") {
					gone[x] = true
				}
			}
		}
	}
	type pair struct {
		a, b *Element
		rel  string
	}
	cut := map[pair]bool{}
	for _, l := range r.Links {
		a, okA := w.Get(l.From)
		b, okB := w.Get(l.To)
		if okA && okB {
			cut[pair{a, b, l.Rel}] = true
			cut[pair{b, a, Reverse(l.Rel)}] = true
		}
	}
	out := &Wiki{byID: map[string]*Element{}, byQName: map[string]*Element{}, links: map[*Element][]edge{}}
	for _, e := range w.Elements {
		if gone[e] {
			continue
		}
		out.Elements = append(out.Elements, e)
		out.byID[e.ID], out.byQName[e.QName] = e, e
		for _, l := range w.links[e] {
			if gone[l.other] || (l.via != nil && gone[l.via]) || cut[pair{e, l.other, l.rel}] {
				continue
			}
			out.links[e] = append(out.links[e], l)
		}
	}
	return out
}

// withoutGoTests hides the verification register's Go tests, for the test
// task: an action whose evidence is a go-test names the test it stands for.
func (w *Wiki) withoutGoTests() *Wiki {
	var hide []string
	for _, e := range w.Elements {
		for _, m := range e.Evidence() {
			if m.Field("kind") == "go-test" {
				hide = append(hide, e.QName)
				break
			}
		}
	}
	return w.Without(Removal{Elements: hide})
}

// ---- the reader ---------------------------------------------------------------

type tokKind int

const (
	tIdent tokKind = iota
	tName          // a 'quoted name'
	tString
	tNumber
	tComment // the text of a /* */ comment
	tPunct
)

type sysTok struct {
	kind tokKind
	text string
	line int
}

var puncts = []string{"::>", ":>>", "::", ":>", ":=", "..", "**", "->", ">=", "<=", "==", "!=", "=>"}

func tokenize(src string) ([]sysTok, error) {
	var out []sysTok
	line := 1
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == '\n':
			line++
			i++
		case c == ' ' || c == '\t' || c == '\r':
			i++
		case strings.HasPrefix(src[i:], "//"):
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case strings.HasPrefix(src[i:], "/*"):
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				return nil, fmt.Errorf("line %d: a comment that never ends", line)
			}
			text := src[i+2 : i+2+end]
			out = append(out, sysTok{tComment, text, line})
			line += strings.Count(text, "\n")
			i += end + 4
		case c == '\'' || c == '"':
			j := i + 1
			var b strings.Builder
			for j < len(src) && src[j] != c {
				if src[j] == '\\' && j+1 < len(src) {
					j++
				}
				if src[j] == '\n' {
					line++
				}
				b.WriteByte(src[j])
				j++
			}
			if j >= len(src) {
				return nil, fmt.Errorf("line %d: a quotation that never ends", line)
			}
			kind := tString
			if c == '\'' {
				kind = tName
			}
			out = append(out, sysTok{kind, b.String(), line})
			i = j + 1
		case c >= '0' && c <= '9':
			j := i
			for j < len(src) && (src[j] >= '0' && src[j] <= '9') {
				j++
			}
			if j+1 < len(src) && src[j] == '.' && src[j+1] >= '0' && src[j+1] <= '9' {
				j++
				for j < len(src) && (src[j] >= '0' && src[j] <= '9' || src[j] == 'e' || src[j] == 'E') {
					j++
				}
			}
			out = append(out, sysTok{tNumber, src[i:j], line})
			i = j
		case c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
			j := i
			for j < len(src) && (src[j] == '_' || src[j] >= 'a' && src[j] <= 'z' || src[j] >= 'A' && src[j] <= 'Z' || src[j] >= '0' && src[j] <= '9') {
				j++
			}
			out = append(out, sysTok{tIdent, src[i:j], line})
			i = j
		default:
			p := string(c)
			for _, cand := range puncts {
				if strings.HasPrefix(src[i:], cand) {
					p = cand
					break
				}
			}
			out = append(out, sysTok{tPunct, p, line})
			i += len(p)
		}
	}
	return out, nil
}

// declKeywords may open a declaration, before its name.
var declKeywords = map[string]bool{}

func init() {
	for _, k := range strings.Fields(`abstract variation variant ref readonly derived constant in out inout end
		individual snapshot timeslice part def attribute item port action state transition requirement
		constraint calc verification use case concern viewpoint view enum interface connection flow
		allocation metadata occurrence analysis rendering stakeholder actor subject objective return
		exhibit perform include decide merge fork join package library standard event message succession
		assert assume require frame exit do alias dependency`) {
		declKeywords[k] = true
	}
}

// dropped are keywords that say nothing about what an element is.
var dropped = map[string]bool{"abstract": true, "readonly": true, "derived": true, "constant": true, "standard": true, "library": true}

// statement is one statement of a body, and the body it opens, if any.
type statement struct {
	toks    []sysTok
	body    []statement
	hasBody bool
}

type reader struct {
	toks []sysTok
	pos  int
}

// statements reads statements until the closing brace of the body or the end.
func (r *reader) statements() ([]statement, error) {
	var out []statement
	var cur []sysTok
	depth := 0 // parentheses and brackets
	flush := func() {
		if len(cur) > 0 {
			out = append(out, statement{toks: cur})
			cur = nil
		}
	}
	for r.pos < len(r.toks) {
		t := r.toks[r.pos]
		r.pos++
		switch {
		case t.kind == tComment:
			if len(cur) == 0 {
				out = append(out, statement{toks: []sysTok{t}})
				continue
			}
			if first := cur[0]; first.kind == tIdent && (first.text == "doc" || first.text == "comment") {
				cur = append(cur, t)
				flush()
			}
			// a comment inside any other statement is only a remark
		case t.kind == tPunct && (t.text == "(" || t.text == "["):
			depth++
			cur = append(cur, t)
		case t.kind == tPunct && (t.text == ")" || t.text == "]"):
			depth--
			cur = append(cur, t)
		case t.kind == tPunct && t.text == ";" && depth == 0:
			flush()
		case t.kind == tPunct && t.text == "{" && depth == 0:
			body, err := r.statements()
			if err != nil {
				return nil, err
			}
			out = append(out, statement{toks: cur, body: body, hasBody: true})
			cur = nil
		case t.kind == tPunct && t.text == "}" && depth == 0:
			flush()
			return out, nil
		default:
			cur = append(cur, t)
		}
	}
	flush()
	return out, nil
}

// ---- building the wiki ---------------------------------------------------------

type pending struct {
	scope *Element
	file  string
	line  int
	do    func(b *builder) bool // false when a reference didn't resolve
	text  string
	types bool // resolved in the first pass, before the references that need types
}

type builder struct {
	root     *Element // the unnamed namespace that owns every top-level package
	elements []*Element
	byQName  map[string]*Element
	links    map[*Element][]edge
	pend     []pending
	unres    []string
	file     string
}

func newBuilder() *builder {
	return &builder{
		root:    &Element{children: map[string]*Element{}},
		byQName: map[string]*Element{},
		links:   map[*Element][]edge{},
	}
}

func (b *builder) parseFile(name, src string) error {
	toks, err := tokenize(src)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	r := &reader{toks: toks}
	stmts, err := r.statements()
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	b.file = name
	b.body(b.root, b.root, stmts, nil)
	return nil
}

// ctx is what a body's statements need from the statements before them.
type ctx struct {
	lastDecl   *Element // the last element declared in this body
	lastDecide *Element // the last decide node, for if and else
	afterEntry bool     // the last statement was entry
	enumBody   bool
	derivation *derivation
}

type derivation struct {
	originals []refAt
	derives   []refAt
}

type refAt struct {
	text string
	line int
}

// body reads one body. owner is the element that owns what is declared in
// it, and scope the element whose names its references see, the same element
// unless the body belongs to an unnamed declaration.
func (b *builder) body(owner, scope *Element, stmts []statement, c *ctx) {
	if c == nil {
		c = &ctx{}
	}
	for _, s := range stmts {
		b.statement(owner, scope, s, c)
	}
}

func text(ts []sysTok) string {
	var parts []string
	for _, t := range ts {
		switch t.kind {
		case tString:
			parts = append(parts, `"`+t.text+`"`)
		case tName:
			parts = append(parts, "'"+t.text+"'")
		default:
			parts = append(parts, t.text)
		}
	}
	return strings.Join(parts, " ")
}

// qref reads a qualified name with an optional feature chain from toks[i:],
// and returns it with the index after it.
func qref(ts []sysTok, i int) (string, int) {
	var b strings.Builder
	for i < len(ts) {
		t := ts[i]
		if t.kind == tIdent || t.kind == tName {
			b.WriteString(t.text)
			i++
			if i < len(ts) && ts[i].kind == tPunct && (ts[i].text == "::" || ts[i].text == ".") {
				if i+1 < len(ts) && ts[i+1].kind == tPunct && (ts[i+1].text == "*" || ts[i+1].text == "**") {
					b.WriteString("::" + ts[i+1].text)
					return b.String(), i + 2
				}
				b.WriteString(ts[i].text)
				i++
				continue
			}
			return b.String(), i
		}
		break
	}
	return b.String(), i
}

func cleanDoc(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		l = strings.TrimSpace(l)
		l = strings.TrimPrefix(l, "*")
		lines[i] = strings.TrimSpace(l)
	}
	return strings.Join(strings.Fields(strings.Join(lines, " ")), " ")
}

func (b *builder) statement(owner, scope *Element, s statement, c *ctx) {
	ts := s.toks
	if len(ts) == 0 {
		if s.hasBody {
			b.body(owner, scope, s.body, &ctx{})
		}
		return
	}
	for len(ts) > 0 && ts[0].kind == tIdent && (ts[0].text == "private" || ts[0].text == "public" || ts[0].text == "protected") {
		ts = ts[1:]
	}
	if len(ts) == 0 {
		return
	}
	first := ts[0]
	line := first.line
	wasEntry := c.afterEntry
	c.afterEntry = false
	switch {
	case first.kind == tComment:
		return
	case first.kind == tIdent && (first.text == "doc" || first.text == "comment"):
		if first.text == "doc" && owner.Doc == "" && owner != b.root {
			for _, t := range ts {
				if t.kind == tComment {
					owner.Doc = cleanDoc(t.text)
				}
			}
		}
		return
	case first.kind == tPunct && first.text == "@":
		m := Meta{}
		m.Type, _ = qref(ts, 1)
		if i := strings.LastIndex(m.Type, "::"); i >= 0 {
			m.Type = m.Type[i+2:]
		}
		for _, f := range s.body {
			if len(f.toks) >= 3 && (f.toks[1].text == "=" || f.toks[1].text == ":=") {
				m.Fields = append(m.Fields, Attr{Name: f.toks[0].text, Value: valueText(f.toks[2:])})
			}
		}
		if owner != b.root {
			owner.Meta = append(owner.Meta, m)
		}
		return
	case first.kind == tPunct && first.text == ":>>":
		name, i := qref(ts, 1)
		if i < len(ts) && (ts[i].text == "=" || ts[i].text == ":=") {
			owner.Attrs = append(owner.Attrs, Attr{Name: lastSegment(name), Value: valueText(ts[i+1:])})
		}
		if s.hasBody {
			b.body(owner, scope, s.body, &ctx{})
		}
		return
	case first.kind == tIdent && first.text == "import":
		ref, _ := qref(ts, 1)
		spec := importSpec{target: ref}
		if strings.HasSuffix(ref, "::*") || strings.HasSuffix(ref, "::**") {
			spec.star = true
			spec.target = strings.TrimSuffix(strings.TrimSuffix(ref, "::**"), "::*")
		}
		scope.imports = append(scope.imports, spec)
		return
	case first.kind == tIdent && first.text == "satisfy":
		i := 1
		if i < len(ts) && ts[i].text == "requirement" {
			i++
		}
		req, j := qref(ts, i)
		by := ""
		if j < len(ts) && ts[j].text == "by" {
			by, _ = qref(ts, j+1)
		}
		holder := owner
		b.defer1(scope, line, "satisfy "+req+" by "+by, func(b *builder) bool {
			r := b.resolve(req, scope)
			p := holder
			if by != "" {
				p = b.resolve(by, scope)
			}
			if r == nil || p == nil {
				return false
			}
			b.link(p, "satisfies", r, nil)
			return true
		})
		return
	case first.kind == tIdent && first.text == "verify":
		i := 1
		if i < len(ts) && ts[i].text == "requirement" {
			i++
		}
		req, _ := qref(ts, i)
		holder := owner
		b.defer1(scope, line, "verify "+req, func(b *builder) bool {
			r := b.resolve(req, scope)
			if r == nil {
				return false
			}
			b.link(holder, "verifies", outermostRequirement(r), nil)
			return true
		})
		return
	case first.kind == tIdent && first.text == "allocate":
		from, i := qref(ts, 1)
		if i < len(ts) && ts[i].text == "to" {
			to, _ := qref(ts, i+1)
			b.defer1(scope, line, "allocate "+from+" to "+to, func(b *builder) bool {
				x, y := b.resolve(from, scope), b.resolve(to, scope)
				if x == nil || y == nil {
					return false
				}
				b.link(x, "allocated to", y, nil)
				return true
			})
		}
		return
	case first.kind == tIdent && first.text == "bind":
		x, i := qref(ts, 1)
		if i < len(ts) && ts[i].text == "=" {
			y, _ := qref(ts, i+1)
			b.defer1(scope, line, "bind "+x+" = "+y, func(b *builder) bool {
				p, q := b.resolve(x, scope), b.resolve(y, scope)
				if p == nil || q == nil {
					return false
				}
				b.link(p, "bound to", q, nil)
				return true
			})
		}
		return
	case first.kind == tIdent && first.text == "connect":
		x, i := qref(ts, 1)
		if i < len(ts) && ts[i].text == "to" {
			y, _ := qref(ts, i+1)
			b.connect(scope, line, x, y, nil)
		}
		return
	case first.kind == tIdent && first.text == "first":
		b.succession(scope, ts, line)
		return
	case first.kind == tIdent && first.text == "then":
		target, _ := qref(ts, 1)
		if target == "" {
			return
		}
		switch {
		case wasEntry:
			holder := owner
			b.defer1(scope, line, "then "+target, func(b *builder) bool {
				x := b.resolve(target, scope)
				if x == nil {
					return false
				}
				b.link(holder, "starts in", x, nil)
				return true
			})
		case c.lastDecl != nil:
			prev := c.lastDecl
			b.defer1(scope, line, "then "+target, func(b *builder) bool {
				x := b.resolve(target, scope)
				if x == nil {
					return false
				}
				b.link(prev, "then", x, nil)
				return true
			})
		}
		return
	case first.kind == tIdent && (first.text == "if" || first.text == "else"):
		if c.lastDecide == nil {
			return
		}
		target := ""
		for i := 1; i < len(ts); i++ {
			if first.text == "else" && i == 1 || ts[i-1].text == "then" {
				target, _ = qref(ts, i)
				if target != "" {
					break
				}
			}
		}
		if target == "" {
			return
		}
		d := c.lastDecide
		b.defer1(scope, line, first.text+" "+target, func(b *builder) bool {
			x := b.resolve(target, scope)
			if x == nil {
				return false
			}
			b.link(d, "then", x, nil)
			return true
		})
		return
	case first.kind == tIdent && first.text == "expose":
		target, _ := qref(ts, 1)
		target = strings.TrimSuffix(strings.TrimSuffix(target, "::**"), "::*")
		holder := owner
		b.defer1(scope, line, "expose "+target, func(b *builder) bool {
			x := b.resolve(target, scope)
			if x == nil {
				return false
			}
			b.link(holder, "exposes", x, nil)
			return true
		})
		return
	case first.kind == tIdent && first.text == "entry":
		if len(ts) == 1 {
			c.afterEntry = true
			return
		}
	case first.kind == tIdent && first.text == "end" && c.derivation != nil:
		for i := 1; i < len(ts); i++ {
			if ts[i].text == "::>" {
				ref, _ := qref(ts, i+1)
				role := ""
				for _, t := range ts[1:i] {
					if t.kind == tIdent {
						role = t.text
					}
				}
				switch role {
				case "original":
					c.derivation.originals = append(c.derivation.originals, refAt{ref, line})
				case "derive":
					c.derivation.derives = append(c.derivation.derives, refAt{ref, line})
				}
			}
		}
		return
	case first.kind == tIdent && first.text == "frame":
		b.frame(owner, scope, s, ts, line)
		return
	case c.enumBody && len(ts) == 1 && ts[0].kind == tIdent:
		b.declare(owner, "enum", ts[0].text, "", line)
		return
	}
	b.declaration(owner, scope, s, ts, c)
}

func lastSegment(q string) string {
	if i := strings.LastIndexAny(q, ":."); i >= 0 {
		return q[i+1:]
	}
	return q
}

// valueText is an attribute's value as written, with a lone string unquoted.
func valueText(ts []sysTok) string {
	if len(ts) == 1 && ts[0].kind == tString {
		return ts[0].text
	}
	return text(ts)
}

// outermostRequirement lifts a reference to part of a requirement, such as its
// acceptance criteria, to the requirement itself.
func outermostRequirement(e *Element) *Element {
	top := e
	for x := e; x != nil && x.QName != ""; x = x.owner {
		if strings.Contains(" "+x.Kind+" ", " requirement ") && !strings.HasSuffix(x.Kind, "def") {
			top = x
		}
	}
	return top
}

func (b *builder) defer1(scope *Element, line int, txt string, do func(b *builder) bool) {
	b.pend = append(b.pend, pending{scope: scope, file: b.file, line: line, do: do, text: txt})
}

func (b *builder) deferTypes(scope *Element, line int, txt string, do func(b *builder) bool) {
	b.pend = append(b.pend, pending{scope: scope, file: b.file, line: line, do: do, text: txt, types: true})
}

// succession reads first a then b then c.
func (b *builder) succession(scope *Element, ts []sysTok, line int) {
	var chain []string
	for i := 1; i < len(ts); {
		ref, j := qref(ts, i)
		if ref == "" {
			i++
			continue
		}
		chain = append(chain, ref)
		i = j
		if i < len(ts) && ts[i].text == "then" {
			i++
		}
	}
	for k := 0; k+1 < len(chain); k++ {
		x, y := chain[k], chain[k+1]
		b.defer1(scope, line, "first "+x+" then "+y, func(b *builder) bool {
			p, q := b.resolve(x, scope), b.resolve(y, scope)
			if p == nil || q == nil {
				return false
			}
			b.link(p, "then", q, nil)
			return true
		})
	}
}

// connect links the parts at two connection ends, each end being the part
// that owns the port or item the end names.
func (b *builder) connect(scope *Element, line int, x, y string, via *Element) {
	rel := "connected to"
	if via != nil && strings.Contains(via.Kind, "flow") {
		rel = "flows to"
	}
	b.defer1(scope, line, "connect "+x+" to "+y, func(b *builder) bool {
		p, q := b.endPart(x, scope), b.endPart(y, scope)
		if p == nil || q == nil {
			return false
		}
		b.link(p, rel, q, via)
		return true
	})
}

// endPart resolves a connection end and walks back past ports and items to
// the part that holds them.
func (b *builder) endPart(ref string, scope *Element) *Element {
	parts := splitChain(ref)
	for n := len(parts); n > 0; n-- {
		e := b.resolve(joinChain(ref, n), scope)
		if e == nil {
			if n == len(parts) {
				return nil
			}
			continue
		}
		if !isPortOrItem(e) {
			return e
		}
	}
	return nil
}

func isPortOrItem(e *Element) bool {
	k := " " + e.Kind + " "
	return strings.Contains(k, " port ") || strings.Contains(k, " item ") || strings.HasPrefix(k, " in ") || strings.HasPrefix(k, " out ") || strings.HasPrefix(k, " inout ")
}

// splitChain splits a reference into its qualified head and its feature
// chain: A::B.c.d gives A::B, c, d.
func splitChain(ref string) []string { return strings.Split(ref, ".") }

func joinChain(ref string, n int) string { return strings.Join(splitChain(ref)[:n], ".") }

func (b *builder) frame(owner, scope *Element, s statement, ts []sysTok, line int) {
	// frame T;  frame concern : T;  frame concern name { ... }
	holder := owner
	if len(ts) < 2 {
		return
	}
	if ts[1].text != "concern" {
		ref, _ := qref(ts, 1)
		b.defer1(scope, line, "frame "+ref, func(b *builder) bool {
			x := b.resolve(ref, scope)
			if x == nil {
				return false
			}
			b.link(holder, "frames", x, nil)
			return true
		})
		return
	}
	i := 2
	name := ""
	if i < len(ts) && (ts[i].kind == tIdent || ts[i].kind == tName) {
		name = ts[i].text
		i++
	}
	typ := ""
	if i < len(ts) && ts[i].text == ":" {
		typ, _ = qref(ts, i+1)
	}
	switch {
	case name != "" && s.hasBody:
		e := b.declare(owner, "concern", name, "", line)
		b.body(e, e, s.body, &ctx{})
		b.link(holder, "frames", e, nil)
	case s.hasBody:
		b.body(owner, scope, s.body, &ctx{})
	}
	if typ != "" {
		b.defer1(scope, line, "frame concern : "+typ, func(b *builder) bool {
			x := b.resolve(typ, scope)
			if x == nil {
				return b.external(typ)
			}
			b.link(holder, "frames", x, nil)
			return true
		})
	}
}

// declare adds an element to its owner.
func (b *builder) declare(owner *Element, kind, name, short string, line int) *Element {
	q := name
	if owner.QName != "" {
		q = owner.QName + "::" + name
	}
	if e, ok := b.byQName[q]; ok && e.Kind == kind && kind == "package" {
		return e // a package declared again in another file
	}
	if _, dup := b.byQName[q]; dup {
		b.unres = append(b.unres, fmt.Sprintf("%s:%d: %s is declared twice", b.file, line, q))
		q = fmt.Sprintf("%s#%d", q, line)
	}
	e := &Element{QName: q, Kind: kind, Name: name, Short: short, Owner: owner.QName, File: b.file, Line: line,
		owner: owner, children: map[string]*Element{}}
	owner.children[name] = e
	if short != "" {
		if _, taken := owner.children[short]; !taken {
			owner.children[short] = e
		}
	}
	owner.order = append(owner.order, e)
	b.byQName[q] = e
	b.elements = append(b.elements, e)
	if owner != b.root {
		b.link(e, "owned by", owner, nil)
	}
	return e
}

// noPage are the kinds whose declarations make no page of their own: what
// they say is lifted onto the element that holds them.
var noPage = map[string]bool{"subject": true, "actor": true, "stakeholder": true}

// lifted are the usages whose type becomes a link from the element that
// holds them.
var lifted = map[string]string{"perform action": "performs", "exhibit state": "exhibits", "include use case": "includes",
	"subject": "subject", "actor": "actor", "stakeholder": "stakeholder"}

func (b *builder) declaration(owner, scope *Element, s statement, ts []sysTok, c *ctx) {
	line := ts[0].line
	var kinds, prefixes []string
	i := 0
	for i < len(ts) {
		t := ts[i]
		if t.kind == tPunct && t.text == "#" && i+1 < len(ts) && ts[i+1].kind == tIdent {
			prefixes = append(prefixes, ts[i+1].text)
			i += 2
			continue
		}
		if t.kind == tIdent && declKeywords[t.text] {
			if !dropped[t.text] {
				kinds = append(kinds, t.text)
			}
			i++
			continue
		}
		break
	}
	kind := strings.Join(kinds, " ")
	var marks []Meta
	for _, p := range prefixes {
		if p == "derivation" && strings.Contains(kind, "connection") {
			kind = "derivation " + kind
			continue
		}
		marks = append(marks, Meta{Type: p})
	}
	if kind == "" {
		// an expression, such as a constraint's last line, says nothing to link
		if s.hasBody {
			b.body(owner, scope, s.body, &ctx{})
		}
		return
	}
	short := ""
	if i+2 < len(ts) && ts[i].text == "<" && (ts[i+1].kind == tName || ts[i+1].kind == tIdent) && ts[i+2].text == ">" {
		short = ts[i+1].text
		i += 3
	}
	name := ""
	if i < len(ts) && (ts[i].kind == tIdent || ts[i].kind == tName) && !isClauseWord(ts[i].text) {
		name = ts[i].text
		i++
	}
	var types, specs, redefs, values []string
	var valueToks []sysTok
	var connectEnds, flowEnds []string
	var flowItem string
	var trFirst, trAccept, trThen string
	for i < len(ts) {
		t := ts[i]
		switch {
		case t.text == "[":
			depth := 0
			for i < len(ts) {
				if ts[i].text == "[" {
					depth++
				}
				if ts[i].text == "]" {
					depth--
					if depth == 0 {
						break
					}
				}
				i++
			}
			i++
		case t.text == ":" || t.text == "typed":
			i++
			if i < len(ts) && ts[i].text == "by" {
				i++
			}
			for {
				ref, j := qref(ts, i)
				if ref == "" {
					break
				}
				types = append(types, ref)
				i = j
				for i < len(ts) && ts[i].text == "[" {
					for i < len(ts) && ts[i].text != "]" {
						i++
					}
					i++
				}
				if i < len(ts) && ts[i].text == "," {
					i++
					continue
				}
				break
			}
		case t.text == ":>" || t.text == "subsets" || t.text == "specializes":
			for {
				ref, j := qref(ts, i+1)
				if ref == "" {
					i++
					break
				}
				specs = append(specs, ref)
				i = j
				if i < len(ts) && ts[i].text == "," {
					continue
				}
				break
			}
		case t.text == ":>>" || t.text == "redefines":
			ref, j := qref(ts, i+1)
			redefs = append(redefs, ref)
			i = max(j, i+1)
		case t.text == "::>" || t.text == "references":
			ref, j := qref(ts, i+1)
			specs = append(specs, ref)
			i = max(j, i+1)
		case t.text == "=" || t.text == ":=" || t.text == "default":
			valueToks = ts[i+1:]
			for _, v := range valueToks {
				if v.kind == tIdent || v.kind == tName {
					ref, _ := qref(valueToks, 0)
					values = append(values, ref)
					break
				}
			}
			i = len(ts)
		case t.text == "connect":
			x, j := qref(ts, i+1)
			if j < len(ts) && ts[j].text == "to" {
				y, k := qref(ts, j+1)
				connectEnds = []string{x, y}
				i = k
			} else {
				i = j
			}
		case t.text == "of" && strings.Contains(kind, "flow"):
			flowItem, i = qref(ts, i+1)
		case t.text == "from" && strings.Contains(kind, "flow"):
			x, j := qref(ts, i+1)
			if j < len(ts) && ts[j].text == "to" {
				y, k := qref(ts, j+1)
				flowEnds = []string{x, y}
				i = k
			} else {
				i = j
			}
		case t.text == "first" && strings.Contains(kind, "transition"):
			trFirst, i = qref(ts, i+1)
		case t.text == "accept" && strings.Contains(kind, "transition"):
			trAccept, i = qref(ts, i+1)
			if i < len(ts) && ts[i].text == ":" {
				trAccept, i = qref(ts, i+1)
			}
		case t.text == "then" && strings.Contains(kind, "transition"):
			trThen, i = qref(ts, i+1)
		default:
			i++
		}
	}
	if name == "" && len(redefs) > 0 {
		name = lastSegment(redefs[0])
	}

	// Attributes are values on their owner, not pages.
	if (kind == "attribute" || strings.HasSuffix(kind, " attribute")) && name != "" {
		owner.Attrs = append(owner.Attrs, Attr{Name: name, Value: valueText(valueToks)})
		return
	}

	holder := owner
	liftRel, isLifted := lifted[kind]
	if noPage[kind] || name == "" {
		// Nothing of its own to show: its type, value and body go to the owner.
		for _, typ := range types {
			if !isLifted {
				continue
			}
			ref := typ
			b.defer1(scope, line, kind+" : "+ref, func(b *builder) bool {
				x := b.resolve(ref, scope)
				if x == nil {
					return b.external(ref)
				}
				b.link(holder, liftRel, x, nil)
				return true
			})
		}
		if isLifted && len(types) == 0 {
			for _, v := range values {
				ref := v
				b.defer1(scope, line, kind+" = "+ref, func(b *builder) bool {
					x := b.resolve(ref, scope)
					if x == nil {
						return false
					}
					b.link(holder, liftRel, x, nil)
					return true
				})
			}
		}
		if len(connectEnds) == 2 {
			b.connect(scope, line, connectEnds[0], connectEnds[1], nil)
		}
		if s.hasBody {
			b.body(owner, scope, s.body, &ctx{})
		}
		return
	}

	e := b.declare(owner, kind, name, short, line)
	e.Meta = append(e.Meta, marks...)
	c.lastDecl = e
	if kind == "decide" {
		c.lastDecide = e
	}
	for _, typ := range types {
		ref := typ
		b.deferTypes(scope, line, name+" : "+ref, func(b *builder) bool {
			x := b.resolve(ref, scope)
			if x == nil {
				return b.external(ref)
			}
			e.types = append(e.types, x)
			b.link(e, "typed by", x, nil)
			if isLifted {
				b.link(holder, liftRel, x, nil)
			}
			return true
		})
	}
	for _, sp := range specs {
		ref := sp
		b.deferTypes(scope, line, name+" :> "+ref, func(b *builder) bool {
			x := b.resolve(ref, scope)
			if x == nil {
				return b.external(ref)
			}
			e.types = append(e.types, x)
			b.link(e, "specialises", x, nil)
			return true
		})
	}
	if len(connectEnds) == 2 {
		b.connect(scope, line, connectEnds[0], connectEnds[1], e)
	}
	if len(flowEnds) == 2 {
		b.connect(scope, line, flowEnds[0], flowEnds[1], e)
		if flowItem != "" {
			item := flowItem
			b.defer1(scope, line, "flow of "+item, func(b *builder) bool {
				x := b.resolve(item, scope)
				if x == nil {
					return b.external(item)
				}
				b.link(e, "carries", x, nil)
				return true
			})
		}
	}
	for _, tr := range [][2]string{{trFirst, "from"}, {trThen, "to"}, {trAccept, "accepts"}} {
		if tr[0] == "" {
			continue
		}
		ref, rel := tr[0], tr[1]
		b.defer1(scope, line, "transition "+rel+" "+ref, func(b *builder) bool {
			x := b.resolve(ref, scope)
			if x == nil {
				return false
			}
			b.link(e, rel, x, nil)
			return true
		})
	}
	if s.hasBody {
		inner := &ctx{enumBody: strings.HasPrefix(kind, "enum")}
		if strings.Contains(kind, "derivation") {
			inner.derivation = &derivation{}
		}
		b.body(e, e, s.body, inner)
		if d := inner.derivation; d != nil {
			for _, o := range d.originals {
				for _, x := range d.derives {
					orig, der := o, x
					b.defer1(e, x.line, "derivation "+orig.text+" to "+der.text, func(b *builder) bool {
						p, q := b.resolve(orig.text, e), b.resolve(der.text, e)
						if p == nil || q == nil {
							return false
						}
						b.link(p, "derives", q, e)
						return true
					})
				}
			}
		}
	}
}

func isClauseWord(s string) bool {
	switch s {
	case "first", "then", "accept", "connect", "of", "from", "to", "by", "if", "do", "via", "redefines", "subsets", "specializes", "typed", "references":
		return true
	}
	return false
}

// externalTypes are the library types the demo's systems model uses from
// outside itself.
var externalTypes = map[string]bool{"String": true, "Real": true, "Integer": true, "Boolean": true, "Natural": true,
	"Positive": true, "Rational": true, "Complex": true, "Number": true, "Anything": true}

var externalPackages = []string{"ScalarValues", "ISQ", "SI", "Base", "Links", "RequirementDerivation", "Occurrences",
	"Items", "Parts", "Actions", "States", "Requirements", "Views", "VerdictKind", "VerificationCases", "Collections"}

// external reports whether an unresolved reference names the standard
// library, which the checkout doesn't hold and nothing needs to link to.
func (b *builder) external(ref string) bool {
	head := strings.Split(strings.Split(ref, ".")[0], "::")
	if externalTypes[head[len(head)-1]] && len(head) <= 2 {
		return true
	}
	for _, p := range externalPackages {
		if head[0] == p {
			return true
		}
	}
	return false
}

// link records a link and its reverse, once.
func (b *builder) link(from *Element, rel string, to, via *Element) {
	b.addLink(from, rel, to, via)
}

func (b *builder) addLink(from *Element, rel string, to *Element, via *Element) {
	if from == nil || to == nil || from == b.root || to == b.root {
		return
	}
	add := func(x *Element, r string, y *Element) {
		for _, l := range b.links[x] {
			if l.rel == r && l.other == y && l.via == via {
				return
			}
		}
		b.links[x] = append(b.links[x], edge{r, y, via})
	}
	add(from, rel, to)
	add(to, Reverse(rel), from)
}

// ---- resolution -----------------------------------------------------------------

// member finds a named member of e, its own or one it has through its types.
func (b *builder) member(e *Element, name string, seen map[*Element]bool) *Element {
	if e == nil || seen[e] {
		return nil
	}
	seen[e] = true
	if m, ok := e.children[name]; ok {
		return m
	}
	for _, t := range e.types {
		if m := b.member(t, name, seen); m != nil {
			return m
		}
	}
	return nil
}

// visible finds a name from scope: its members and imports, then those of
// every namespace around it.
func (b *builder) visible(name string, scope *Element) *Element {
	for s := scope; s != nil; s = s.owner {
		if m := b.member(s, name, map[*Element]bool{}); m != nil {
			return m
		}
		for _, imp := range s.imports {
			target := b.byQName[imp.target]
			if target == nil {
				continue
			}
			if imp.star {
				if m := b.member(target, name, map[*Element]bool{}); m != nil {
					return m
				}
			} else if lastSegment(imp.target) == name {
				return target
			}
		}
	}
	return nil
}

// resolve finds the element a reference names, seen from scope.
func (b *builder) resolve(ref string, scope *Element) *Element {
	if ref == "" {
		return nil
	}
	chain := splitChain(ref)
	segs := strings.Split(chain[0], "::")
	e := b.byQName[chain[0]]
	if e == nil {
		e = b.visible(segs[0], scope)
		if e == nil {
			// any package's own member, as the model's ::* imports allow
			var found *Element
			for _, x := range b.elements {
				if x.Name == segs[0] && x.owner != nil && (x.owner.isPackage() || x.owner == b.root) {
					if found != nil {
						found = nil
						break
					}
					found = x
				}
			}
			e = found
		}
		for _, s := range segs[1:] {
			if e == nil {
				break
			}
			e = b.member(e, s, map[*Element]bool{})
		}
	}
	for _, c := range chain[1:] {
		if e == nil {
			break
		}
		e = b.member(e, c, map[*Element]bool{})
	}
	return e
}

func (b *builder) finish() *Wiki {
	// Types first, since members are found through them, then everything else.
	for _, pass := range []bool{true, false} {
		for _, p := range b.pend {
			if p.types != pass {
				continue
			}
			if !p.do(b) {
				b.unres = append(b.unres, fmt.Sprintf("%s:%d: %s", p.file, p.line, p.text))
			}
		}
	}
	w := &Wiki{Unresolved: b.unres, byID: map[string]*Element{}, byQName: map[string]*Element{}, links: b.links}
	w.Elements = b.elements
	assignIDs(w.Elements)
	for _, e := range w.Elements {
		w.byID[e.ID] = e
		w.byQName[e.QName] = e
	}
	for e, ls := range w.links {
		sort.SliceStable(ls, func(i, j int) bool {
			if relOrder[ls[i].rel] != relOrder[ls[j].rel] {
				return relOrder[ls[i].rel] < relOrder[ls[j].rel]
			}
			if ls[i].other.ID != ls[j].other.ID {
				return ls[i].other.ID < ls[j].other.ID
			}
			return viaID(ls[i]) < viaID(ls[j])
		})
		w.links[e] = ls
	}
	return w
}

func viaID(l edge) string {
	if l.via == nil {
		return ""
	}
	return l.via.ID
}

var simpleShort = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// assignIDs gives every element its identifier: its short name when that is
// simple and unique, and otherwise the shortest suffix of its qualified name
// that no other element shares, at least two segments long for a feature so
// that it reads as demo::router and not router.
func assignIDs(es []*Element) {
	shorts := map[string]int{}
	for _, e := range es {
		if e.Short != "" && simpleShort.MatchString(e.Short) {
			shorts[e.Short]++
		}
	}
	least := func(e *Element) int {
		if e.owner != nil && e.owner.QName != "" && !e.owner.isPackage() {
			return 2
		}
		return 1
	}
	// A suffix counts only where an element could take it as its identifier,
	// so a feature's name doesn't crowd out a definition's.
	suffixes := map[string]int{}
	for _, e := range es {
		segs := strings.Split(e.QName, "::")
		for k := min(least(e), len(segs)); k <= len(segs); k++ {
			suffixes[strings.Join(segs[len(segs)-k:], "::")]++
		}
	}
	for _, e := range es {
		if e.Short != "" && shorts[e.Short] == 1 && suffixes[e.Short] == 0 {
			e.ID = e.Short
			continue
		}
		segs := strings.Split(e.QName, "::")
		least := least(e)
		e.ID = e.QName
		for k := min(least, len(segs)); k <= len(segs); k++ {
			cand := strings.Join(segs[len(segs)-k:], "::")
			if suffixes[cand] == 1 && shorts[cand] == 0 {
				e.ID = cand
				break
			}
		}
	}
}

// nameWords turns an element's name into words: SR_04_FourPathsOnOnePort
// gives "Four paths on one port".
func nameWords(name string) string {
	name = keyPrefix.ReplaceAllString(name, "")
	words := splitIdentifier(name)
	for i := range words {
		if i > 0 && !isAllUpper(words[i]) {
			words[i] = strings.ToLower(words[i])
		}
	}
	if len(words) > 0 {
		r := []rune(words[0])
		r[0] = unicode.ToUpper(r[0])
		words[0] = string(r)
	}
	return strings.Join(words, " ")
}

var keyPrefix = regexp.MustCompile(`^[A-Z]+_\d+_`)
