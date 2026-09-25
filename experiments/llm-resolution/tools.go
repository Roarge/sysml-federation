package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The tools answer the language model's calls from the systems model and
// the code, with the smallest text that answers each. The program does
// everything a program can do, searching, following links, finding paths and
// reading files, so that the language model only chooses and judges.

// The two tasks.
const (
	TaskTest     = "test"
	TaskIncident = "incident"
)

// The tools' limits.
const (
	LimitFind      = 8    // elements a search lists
	LimitPaths     = 3    // paths a path call gives
	LimitHops      = 3    // links in one path, and the reach of "near"
	LimitDoc       = 800  // characters of a doc answer
	LimitCode      = 8    // lines of a code answer
	LimitGrep      = 8    // hits of a search of the code
	LimitGrepFile  = 3    // hits from any one file
	LimitReadLines = 30   // lines a read shows
	LimitGroup     = 600  // characters of one relationship's list
	LimitAnswer    = 1500 // characters of any answer
	LimitLine      = 140  // characters of one line of code shown
)

// ToolNames are the tools, in the order the instructions list them.
var ToolNames = []string{"find", "links", "doc", "path", "code", "grep", "read", "done"}

// Scope is what one browse's tools hide.
type Scope struct {
	Task     string          // TaskTest or TaskIncident
	Redact   map[string]bool // normalised words deleted from every answer
	HideTest string          // file#FuncName, a test whose own code the code tools hide
}

// ---- the built system's code --------------------------------------------------

// CodeFile is one file of the built system.
type CodeFile struct {
	Path  string
	Lines []string
}

// CodeBase is the built system's files in the checkout: its Go, JavaScript,
// TypeScript, YAML and JSON files, Dockerfiles and shell scripts, outside
// model, docs, experiments and test data.
type CodeBase struct {
	Root   string
	Files  []*CodeFile
	byPath map[string]*CodeFile
	masked map[string][]string
}

var codeExtensions = map[string]bool{".go": true, ".js": true, ".mjs": true, ".cjs": true, ".ts": true, ".tsx": true,
	".yml": true, ".yaml": true, ".json": true, ".sh": true}

var skippedDirs = map[string]bool{"model": true, "docs": true, "experiments": true, "testdata": true, "vendor": true,
	"node_modules": true, "results": true}

const maxCodeFile = 512 << 10

// LoadCode reads the built system's files under root.
func LoadCode(root string) (*CodeBase, error) {
	c := &CodeBase{Root: root, byPath: map[string]*CodeFile{}, masked: map[string][]string{}}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (strings.HasPrefix(name, ".") || skippedDirs[name]) {
				return filepath.SkipDir
			}
			return nil
		}
		if !codeExtensions[filepath.Ext(name)] && !strings.HasPrefix(name, "Dockerfile") {
			return nil
		}
		if info, err := d.Info(); err != nil || info.Size() > maxCodeFile || strings.HasSuffix(name, "-lock.json") {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		cf := &CodeFile{Path: filepath.ToSlash(rel)}
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			cf.Lines = append(cf.Lines, sc.Text())
		}
		if err := sc.Err(); err != nil {
			return err
		}
		c.Files = append(c.Files, cf)
		c.byPath[cf.Path] = cf
		return nil
	})
	sort.Slice(c.Files, func(i, j int) bool { return c.Files[i].Path < c.Files[j].Path })
	return c, err
}

// linesOf gives a file's lines as a task sees them: with every requirement
// key masked in the test task.
func (c *CodeBase) linesOf(f *CodeFile, task string) []string {
	if task != TaskTest {
		return f.Lines
	}
	if m, ok := c.masked[f.Path]; ok {
		return m
	}
	m := make([]string, len(f.Lines))
	for i, l := range f.Lines {
		m[i] = maskKeys(l)
	}
	c.masked[f.Path] = m
	return m
}

// exists reports whether a file or folder is in the checkout at all.
func (c *CodeBase) exists(rel string) (bool, bool) {
	info, err := os.Stat(filepath.Join(c.Root, filepath.FromSlash(rel)))
	if err != nil {
		return false, false
	}
	return true, info.IsDir()
}

// ---- the tool kit -------------------------------------------------------------

// ToolKit answers one browse's tool calls.
type ToolKit struct {
	w      *Wiki
	code   *CodeBase
	scope  Scope
	keys   map[string]bool // the system requirements' keys
	reqs   *Baseline       // the search among the system requirements, the baseline itself
	all    *Baseline
	byKind map[string]*Baseline
	hidden map[string][2]int // file: first and last line hidden, 1-based
	looked map[string]bool   // files read, or found by a search, in this browse
}

// NewToolKit gives the tools for one browse. In the test task the
// verification register's Go tests are out of the wiki, and requirement keys
// are masked in the code and in test names.
func NewToolKit(w *Wiki, code *CodeBase, reqs []Requirement, sc Scope) *ToolKit {
	if sc.Task == TaskTest {
		w = w.withoutGoTests()
	}
	k := &ToolKit{w: w, code: code, scope: sc, keys: map[string]bool{}, reqs: NewBaseline(reqs),
		byKind: map[string]*Baseline{}, hidden: map[string][2]int{}, looked: map[string]bool{}}
	for _, r := range reqs {
		k.keys[r.Key] = true
	}
	var ids, texts []string
	for _, e := range w.Elements {
		ids = append(ids, e.ID)
		texts = append(texts, pageText(e))
	}
	k.all = newIndex(ids, texts)
	if sc.HideTest != "" {
		k.hideTest(sc.HideTest)
	}
	return k
}

// Wiki is the systems model as this browse sees it.
func (k *ToolKit) Wiki() *Wiki { return k.w }

// pageText is what a search reads of an element.
func pageText(e *Element) string {
	parts := []string{nameWords(e.Name), e.Short, e.Doc}
	for _, a := range e.Attrs {
		parts = append(parts, a.Value)
	}
	for _, m := range e.Evidence() {
		parts = append(parts, m.Field("location"))
	}
	return strings.Join(parts, " ")
}

// hideTest hides a test function and the comment above it from the code
// tools, for the rebuild probe.
func (k *ToolKit) hideTest(id string) {
	file, fn, ok := strings.Cut(id, "#")
	if !ok {
		return
	}
	f := k.code.byPath[file]
	if f == nil {
		return
	}
	start := -1
	for i, l := range f.Lines {
		if strings.HasPrefix(l, "func "+fn+"(") {
			start = i
			break
		}
	}
	if start < 0 {
		return
	}
	end := start
	if !strings.HasSuffix(strings.TrimSpace(f.Lines[start]), "}") {
		for end = start + 1; end < len(f.Lines) && f.Lines[end] != "}"; end++ {
		}
	}
	for start > 0 && strings.HasPrefix(f.Lines[start-1], "//") {
		start--
	}
	k.hidden[file] = [2]int{start + 1, end + 1}
}

func (k *ToolKit) isHidden(file string, line int) bool {
	h, ok := k.hidden[file]
	return ok && line >= h[0] && line <= h[1]
}

// Call answers one tool call. Every answer is cut to its limit, and in the
// deletion probes the deleted words go from it.
func (k *ToolKit) Call(tool, arg, arg2 string) string {
	arg, arg2 = strings.TrimSpace(arg), strings.TrimSpace(arg2)
	var answer string
	switch tool {
	case "find":
		answer = k.find(arg, arg2)
	case "links":
		answer = k.links(arg, arg2)
	case "doc":
		answer = k.doc(arg)
	case "path":
		answer = k.path(arg, arg2)
	case "code":
		answer = k.codeOf(arg)
	case "grep":
		answer = k.grep(arg)
	case "read":
		answer = k.read(arg, arg2)
	default:
		answer = fmt.Sprintf("there is no tool %q: use one of %s", tool, strings.Join(ToolNames, ", "))
	}
	if k.scope.Task == TaskTest {
		answer = keyInTextRE.ReplaceAllString(answer, "$1")
	}
	answer = capText(answer, LimitAnswer)
	if len(k.scope.Redact) > 0 {
		answer = redact(answer, k.scope.Redact)
	}
	return answer
}

func capText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := strings.LastIndex(s[:n-12], "\n")
	if cut < n/2 {
		cut = n - 12
	}
	return s[:cut] + "\n(cut short)"
}

func (k *ToolKit) unknown(id string) string {
	return fmt.Sprintf("unknown identifier %q: use an identifier as find, links or path give it", id)
}

// get finds an element of this browse's wiki.
func (k *ToolKit) get(id string) (*Element, bool) { return k.w.Get(id) }

// ---- find ---------------------------------------------------------------------

func (k *ToolKit) find(text, kind string) string {
	words := k.queryTerms(text)
	if len(words) == 0 {
		return fmt.Sprintf("nothing found for %q", text)
	}
	var index *Baseline
	switch kind {
	case "":
		index = k.all
	case "requirement", "requirements":
		index = k.reqs
	default:
		index = k.kindIndex(kind)
		if index == nil {
			return fmt.Sprintf("no elements of kind %q; kinds include %s", kind, strings.Join(k.kinds(), ", "))
		}
	}
	var lines []string
	more := 0
	for _, r := range index.rankTerms(words) {
		if r.Score <= 0 {
			break
		}
		if len(lines) == LimitFind {
			more++
			continue
		}
		e, ok := k.get(r.Key)
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s (%s): %s", e.ID, e.Kind, gist(e)))
	}
	if len(lines) == 0 {
		return fmt.Sprintf("nothing found for %q", text)
	}
	if more > 0 {
		lines = append(lines, fmt.Sprintf("(and %d more)", more))
	}
	return strings.Join(lines, "\n")
}

// queryTerms are the search's words, less any the probe deleted.
func (k *ToolKit) queryTerms(text string) []string {
	var out []string
	for _, t := range terms(text) {
		if !k.scope.Redact[t] {
			out = append(out, t)
		}
	}
	return out
}

func (k *ToolKit) kindIndex(kind string) *Baseline {
	if b, ok := k.byKind[kind]; ok {
		return b
	}
	var ids, texts []string
	for _, e := range k.w.Elements {
		if e.Kind == kind {
			ids = append(ids, e.ID)
			texts = append(texts, pageText(e))
		}
	}
	if len(ids) == 0 {
		return nil
	}
	b := newIndex(ids, texts)
	k.byKind[kind] = b
	return b
}

func (k *ToolKit) kinds() []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range k.w.Elements {
		if !seen[e.Kind] {
			seen[e.Kind] = true
			out = append(out, e.Kind)
		}
	}
	sort.Strings(out)
	return out
}

// gist is one short line about an element.
func gist(e *Element) string {
	if e.Short != "" && strings.Contains(e.Kind, "requirement") {
		return nameWords(e.Name)
	}
	if e.Doc != "" {
		return clip(firstSentence(e.Doc), 80)
	}
	for _, t := range e.types {
		article := "a "
		if strings.ContainsRune("aeiou", rune(e.Kind[0])) {
			article = "an "
		}
		return article + e.Kind + " of " + t.ID
	}
	return nameWords(e.Name)
}

func firstSentence(s string) string {
	if i := strings.Index(s, ". "); i > 0 {
		return s[:i+1]
	}
	return s
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := strings.LastIndex(s[:n], " ")
	if cut < n/2 {
		cut = n
	}
	return s[:cut] + "…"
}

// ---- links --------------------------------------------------------------------

func (k *ToolKit) links(id, rel string) string {
	e, ok := k.get(id)
	if !ok {
		return k.unknown(id)
	}
	type group struct {
		rel   string
		items []string
	}
	var groups []*group
	byRel := map[string]*group{}
	for _, l := range k.w.LinksOf(e.ID) {
		if rel != "" && l.Rel != rel {
			continue
		}
		g, ok := byRel[l.Rel]
		if !ok {
			g = &group{rel: l.Rel}
			byRel[l.Rel] = g
			groups = append(groups, g)
		}
		item := l.Other
		if l.Via != "" {
			item += " (via " + l.Via + ")"
		}
		g.items = append(g.items, item)
	}
	lines := []string{fmt.Sprintf("%s (%s)", e.ID, e.Kind)}
	if len(groups) == 0 {
		if rel != "" {
			return lines[0] + fmt.Sprintf("\nno %q links", rel)
		}
		return lines[0] + "\nno links"
	}
	room := LimitGroup
	if rel != "" {
		room = LimitAnswer - 100
	}
	size := len(lines[0])
	for n, g := range groups {
		if size+len(g.rel)+40 > LimitAnswer {
			lines = append(lines, fmt.Sprintf("(and %d more relationships)", len(groups)-n))
			break
		}
		room := min(room, LimitAnswer-size-40)
		line := g.rel + ": "
		shown := 0
		for i, item := range g.items {
			sep := ""
			if i > 0 {
				sep = ", "
			}
			if len(line)+len(sep)+len(item) > room {
				break
			}
			line += sep + item
			shown++
		}
		if shown < len(g.items) {
			line += fmt.Sprintf(" (and %d more)", len(g.items)-shown)
		}
		lines = append(lines, line)
		size += len(line) + 1
	}
	return strings.Join(lines, "\n")
}

// ---- doc ----------------------------------------------------------------------

func (k *ToolKit) doc(id string) string {
	e, ok := k.get(id)
	if !ok {
		return k.unknown(id)
	}
	head := fmt.Sprintf("%s (%s)", e.ID, e.Kind)
	if o, ok := k.w.byQName[e.Owner]; ok {
		head += ", owned by " + o.ID
	}
	lines := []string{head}
	var types []string
	for _, l := range k.w.LinksOf(e.ID) {
		if l.Rel == "typed by" || l.Rel == "specialises" {
			types = append(types, l.Other)
		}
	}
	if len(types) > 0 {
		lines = append(lines, "typed by: "+strings.Join(types, ", "))
	}
	switch {
	case e.Doc != "":
		lines = append(lines, "doc: "+e.Doc)
	case len(types) > 0:
		if t, ok := k.get(types[0]); ok && t.Doc != "" {
			lines = append(lines, "doc of "+t.ID+": "+t.Doc)
		}
	}
	var attrs []string
	for _, a := range e.Attrs {
		attrs = append(attrs, a.Name+" = "+a.Value)
	}
	if len(attrs) > 0 {
		lines = append(lines, "attributes: "+strings.Join(attrs, "; "))
	}
	var decisions, evidence []string
	for _, m := range e.Decisions() {
		decisions = append(decisions, m.Field("id")+" "+m.Field("title"))
	}
	for _, m := range e.Evidence() {
		evidence = append(evidence, m.Field("kind")+" "+m.Field("location"))
	}
	if len(decisions) > 0 {
		lines = append(lines, "decisions: "+strings.Join(decisions, "; "))
	}
	if len(evidence) > 0 {
		lines = append(lines, "evidence: "+strings.Join(evidence, "; "))
	}
	return capText(strings.Join(lines, "\n"), LimitDoc)
}

// ---- path ---------------------------------------------------------------------

type hop struct {
	rel string
	to  *Element
}

// paths finds the shortest link paths of at most three links between two
// elements, never passing through a package, since everything a package owns
// would otherwise be two links from everything else it owns.
func (k *ToolKit) paths(a, b *Element, most int) [][]hop {
	var found [][]hop
	add := func(p []hop) {
		seen := map[*Element]bool{a: true}
		for _, h := range p {
			if seen[h.to] {
				return
			}
			seen[h.to] = true
		}
		found = append(found, p)
	}
	through := func(e *Element) bool { return e != a && e != b && !e.isPackage() }
	next := func(e *Element) []hop {
		var out []hop
		for _, l := range k.w.links[e] {
			out = append(out, hop{l.rel, l.other})
		}
		return out
	}
	into := map[*Element][]hop{} // the last hop into b, by the element it leaves
	for _, l := range k.w.links[b] {
		into[l.other] = append(into[l.other], hop{Reverse(l.rel), b})
	}
	for length := 1; length <= LimitHops && len(found) < most; length++ {
		var batch [][]hop
		keep := func(p []hop) { batch = append(batch, append([]hop{}, p...)) }
		for _, h1 := range next(a) {
			switch length {
			case 1:
				if h1.to == b {
					keep([]hop{h1})
				}
			case 2:
				if through(h1.to) {
					for _, h2 := range into[h1.to] {
						keep([]hop{h1, h2})
					}
				}
			case 3:
				if !through(h1.to) {
					continue
				}
				for _, h2 := range next(h1.to) {
					if through(h2.to) && h2.to != h1.to {
						for _, h3 := range into[h2.to] {
							keep([]hop{h1, h2, h3})
						}
					}
				}
			}
		}
		sort.SliceStable(batch, func(i, j int) bool { return pathText(a, batch[i]) < pathText(a, batch[j]) })
		for _, p := range batch {
			if len(found) == most {
				break
			}
			add(p)
		}
	}
	return found
}

func pathText(a *Element, p []hop) string {
	var b strings.Builder
	b.WriteString(a.ID)
	for _, h := range p {
		b.WriteString(" -" + h.rel + "-> " + h.to.ID)
	}
	return b.String()
}

func (k *ToolKit) path(from, to string) string {
	a, ok := k.get(from)
	if !ok {
		return k.unknown(from)
	}
	b, ok := k.get(to)
	if !ok {
		return k.unknown(to)
	}
	if a == b {
		return "the two are the same element, " + a.ID
	}
	ps := k.paths(a, b, LimitPaths)
	if len(ps) == 0 {
		return fmt.Sprintf("no path of %d links or fewer between %s and %s", LimitHops, a.ID, b.ID)
	}
	var lines []string
	for _, p := range ps {
		lines = append(lines, pathText(a, p))
	}
	return strings.Join(lines, "\n")
}

// near gives every element within three links of e, not through a package.
func (w *Wiki) near(e *Element) map[*Element]bool {
	seen := map[*Element]bool{e: true}
	frontier := []*Element{e}
	for d := 0; d < LimitHops; d++ {
		var next []*Element
		for _, x := range frontier {
			if x != e && x.isPackage() {
				continue
			}
			for _, l := range w.links[x] {
				if !seen[l.other] {
					seen[l.other] = true
					next = append(next, l.other)
				}
			}
		}
		frontier = next
	}
	delete(seen, e)
	return seen
}

// ---- code ---------------------------------------------------------------------

var pathInText = regexp.MustCompile(`[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)+`)

// codeLines gives where the systems model puts an element in the code, as
// (path, line text) pairs: evidence locations of the element and everything
// it owns, paths its description or its type's names, and where its
// distinctive values occur in the code.
func (k *ToolKit) codeLines(e *Element) []codeRef {
	var refs []codeRef
	seen := map[string]bool{}
	add := func(path, text string) {
		if key := path + "\x00" + text; !seen[key] {
			seen[key] = true
			refs = append(refs, codeRef{path, text})
		}
	}
	// The element, what it owns, and the definitions of both: a usage's code
	// is mostly its definition's.
	subjects := []*Element{e}
	for _, x := range k.w.Elements {
		if strings.HasPrefix(x.QName, e.QName+"::") {
			subjects = append(subjects, x)
		}
	}
	have := map[*Element]bool{}
	for _, x := range subjects {
		have[x] = true
	}
	for _, x := range append([]*Element{}, subjects...) {
		for _, l := range k.w.links[x] {
			if (l.rel == "typed by" || l.rel == "specialises" || l.rel == "exhibits" || l.rel == "performs") && !have[l.other] {
				have[l.other] = true
				subjects = append(subjects, l.other)
			}
		}
	}
	for _, x := range subjects {
		for _, m := range x.Evidence() {
			loc := m.Field("location")
			add(loc, fmt.Sprintf("evidence (%s) of %s: %s%s", m.Field("kind"), x.ID, loc, k.missing(loc)))
		}
	}
	for _, x := range subjects {
		for _, p := range pathInText.FindAllString(x.Doc, -1) {
			p = strings.TrimRight(p, ".,;:")
			if first, _, _ := strings.Cut(p, "/"); strings.Contains(first, ".") {
				continue // a host, as in ghcr.io/owner/image, not a file
			}
			if ok, dir := k.code.exists(p); ok || strings.Contains(p, ".") {
				if dir {
					add(p, fmt.Sprintf("named in the description of %s: %s/ (a folder)", x.ID, p))
				} else {
					add(p, fmt.Sprintf("named in the description of %s: %s%s", x.ID, p, k.missing(p)))
				}
			}
		}
	}
	for _, x := range subjects {
		for _, a := range x.Attrs {
			if !distinctive(a.Value) {
				continue
			}
			hits := k.search(a.Value, 2)
			for _, h := range hits {
				add(h.path, fmt.Sprintf("%s at %s:%d", a.Value, h.path, h.line))
			}
		}
	}
	return refs
}

type codeRef struct {
	path string
	text string
}

func (k *ToolKit) missing(path string) string {
	if ok, _ := k.code.exists(path); ok {
		return ""
	}
	return " (not in the checkout)"
}

// distinctive values are worth searching the code for: addresses, service
// names and paths, and not plain words, numbers or sentences.
func distinctive(v string) bool {
	if len(v) < 4 || strings.ContainsAny(v, " \t") || !strings.ContainsAny(v, "0123456789.:/-_") {
		return false
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return false
	}
	return !strings.Contains(v, "::")
}

func (k *ToolKit) codeOf(id string) string {
	e, ok := k.get(id)
	if !ok {
		return k.unknown(id)
	}
	refs := k.codeLines(e)
	if len(refs) == 0 {
		return fmt.Sprintf("the systems model gives no code for %s", e.ID)
	}
	var lines []string
	size := 0
	for i, r := range refs {
		text := clip(r.text, 2*LimitLine)
		if i == LimitCode || size+len(text) > LimitAnswer-40 {
			lines = append(lines, fmt.Sprintf("(and %d more)", len(refs)-i))
			break
		}
		k.looked[r.path] = true
		lines = append(lines, text)
		size += len(text) + 1
	}
	return strings.Join(lines, "\n")
}

// codePaths are the files and folders the code tool gives for an element.
func (k *ToolKit) codePaths(e *Element) map[string]bool {
	out := map[string]bool{}
	for _, r := range k.codeLines(e) {
		out[r.path] = true
	}
	return out
}

// ---- grep and read ------------------------------------------------------------

type hit struct {
	path string
	line int
	text string
}

// search finds up to n lines holding the text, apart from case, in the built
// system as this task sees it.
func (k *ToolKit) search(text string, n int) []hit {
	needle := strings.ToLower(text)
	var out []hit
	for _, f := range k.code.Files {
		for i, l := range k.code.linesOf(f, k.scope.Task) {
			if k.isHidden(f.Path, i+1) {
				continue
			}
			if strings.Contains(strings.ToLower(l), needle) {
				out = append(out, hit{f.Path, i + 1, l})
				if n > 0 && len(out) == n {
					return out
				}
			}
		}
	}
	return out
}

func (k *ToolKit) grep(text string) string {
	pattern := text
	if len(k.scope.Redact) > 0 {
		pattern, _ = removeTerms(pattern, k.scope.Redact)
	}
	if strings.TrimSpace(pattern) == "" {
		return fmt.Sprintf("no match for %q in the built system", text)
	}
	hits := k.search(pattern, 0)
	if len(hits) == 0 {
		return fmt.Sprintf("no match for %q in the built system", text)
	}
	var lines []string
	perFile := map[string]int{}
	shown := 0
	for _, h := range hits {
		if shown == LimitGrep || perFile[h.path] == LimitGrepFile {
			continue
		}
		perFile[h.path]++
		shown++
		k.looked[h.path] = true
		lines = append(lines, fmt.Sprintf("%s:%d: %s", h.path, h.line, clip(strings.TrimSpace(h.text), LimitLine)))
	}
	if shown < len(hits) {
		files := map[string]bool{}
		for _, h := range hits {
			files[h.path] = true
		}
		lines = append(lines, fmt.Sprintf("(and %d more matches, in %d files in all)", len(hits)-shown, len(files)))
	}
	return strings.Join(lines, "\n")
}

var pathWithLine = regexp.MustCompile(`^(.*?):(\d+)$`)

func (k *ToolKit) read(path, lineArg string) string {
	if m := pathWithLine.FindStringSubmatch(path); m != nil && lineArg == "" {
		path, lineArg = m[1], m[2]
	}
	path = strings.TrimPrefix(strings.TrimSpace(path), "./")
	f := k.code.byPath[path]
	if f == nil {
		return fmt.Sprintf("no file %q in the built system", path)
	}
	line, err := strconv.Atoi(strings.TrimSpace(lineArg))
	if err != nil || line < 1 {
		line = 1
	}
	lines := k.code.linesOf(f, k.scope.Task)
	if line > len(lines) {
		line = len(lines)
	}
	start := max(1, line-10)
	end := min(len(lines), start+LimitReadLines-1)
	start = max(1, min(start, end-LimitReadLines+1))
	render := func(a, b int) string {
		var out []string
		out = append(out, fmt.Sprintf("%s, lines %d to %d of %d", path, a, b, len(lines)))
		for n := a; n <= b; n++ {
			if k.isHidden(path, n) {
				continue
			}
			out = append(out, fmt.Sprintf("%d| %s", n, clip(lines[n-1], LimitLine)))
		}
		return strings.Join(out, "\n")
	}
	text := render(start, end)
	for len(text) > LimitAnswer && end > start {
		if line-start > end-line {
			start++
		} else {
			end--
		}
		text = render(start, end)
	}
	k.looked[path] = true
	return text
}

// ---- redaction ----------------------------------------------------------------

// redact deletes every word whose normalised form is in drop, line by line,
// keeping each line's indentation.
func redact(text string, drop map[string]bool) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		indent := l[:len(l)-len(strings.TrimLeft(l, " \t"))]
		out, n := removeTerms(l, drop)
		if n > 0 {
			lines[i] = indent + out
		}
	}
	return strings.Join(lines, "\n")
}

// shuffled is a fixed shuffle of words for a probe's control.
func shuffled(words []string, seed int64) []string {
	out := append([]string{}, words...)
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}
