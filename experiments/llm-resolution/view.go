package main

import (
	"fmt"
	"strings"
	"unicode"
)

// The view is the resolver's working state. It starts empty. The language
// model frames a viewpoint, the question the view answers, and exposes the
// elements that matter with a note saying why, and prunes those that turn out
// not to. The view and the last tool's answer are all a question carries, so
// the notes are how the language model keeps what it has learnt.

// ViewItem is one exposed element with the note saying why.
type ViewItem struct {
	ID   string `json:"id"`
	Note string `json:"note"`
}

// View is the resolver's working view.
type View struct {
	Viewpoint string     `json:"viewpoint"`
	Items     []ViewItem `json:"items"`
}

// Expose adds an element to the view, or renews its note. An element the
// wiki doesn't hold is refused.
func (v *View) Expose(w *Wiki, id, note string) error {
	e, ok := w.Get(id)
	if !ok {
		return fmt.Errorf("%s is unknown, so it isn't in the view", strings.TrimSpace(id))
	}
	note = strings.TrimSpace(note)
	for i := range v.Items {
		if v.Items[i].ID == e.ID {
			v.Items[i].Note = note
			return nil
		}
	}
	v.Items = append(v.Items, ViewItem{ID: e.ID, Note: note})
	return nil
}

// Prune takes an element out of the view, and reports whether it was there.
func (v *View) Prune(id string) bool {
	id = strings.TrimSpace(id)
	for i, it := range v.Items {
		if it.ID == id {
			v.Items = append(v.Items[:i], v.Items[i+1:]...)
			return true
		}
	}
	return false
}

// Lines is the view as a question shows it.
func (v View) Lines() string {
	if len(v.Items) == 0 {
		return "(empty)"
	}
	var lines []string
	for _, it := range v.Items {
		lines = append(lines, "- "+it.ID+": "+it.Note)
	}
	return strings.Join(lines, "\n")
}

// SysML writes the view out in SysML v2, as a package of its own: a
// viewpoint definition that frames the viewpoint's question as its concern,
// and a view that exposes each element by its qualified name. The notes go
// in as comments beside the exposures.
func (v View) SysML(w *Wiki, name string) string {
	vp := viewpointName(v.Viewpoint)
	question := strings.ReplaceAll(v.Viewpoint, "*/", "* /")
	if question == "" {
		question = "(no viewpoint was framed)"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "package %s {\n", viewpointName(name+" view"))
	fmt.Fprintf(&b, "    viewpoint def %s {\n", vp)
	fmt.Fprintf(&b, "        frame concern question {\n            doc /* %s */\n        }\n    }\n", question)
	fmt.Fprintf(&b, "    view %s {\n", name)
	fmt.Fprintf(&b, "        viewpoint answers : %s;\n", vp)
	for _, it := range v.Items {
		q := it.ID
		if e, ok := w.Get(it.ID); ok {
			q = e.QName
		}
		if note := strings.ReplaceAll(it.Note, "*/", "* /"); note != "" {
			fmt.Fprintf(&b, "        // %s\n", strings.ReplaceAll(note, "\n", " "))
		}
		fmt.Fprintf(&b, "        expose %s;\n", sysmlName(q))
	}
	b.WriteString("    }\n}\n")
	return b.String()
}

// viewpointName turns the viewpoint's question into a name: "Why is the
// query page down?" gives WhyIsTheQueryPageDown.
func viewpointName(q string) string {
	var b strings.Builder
	for _, word := range strings.FieldsFunc(q, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		r := []rune(word)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	name := b.String()
	if name == "" || unicode.IsDigit([]rune(name)[0]) {
		name = "Viewpoint" + name
	}
	if len(name) > 60 {
		name = name[:60]
	}
	return name
}

// sysmlName quotes the segments of a qualified name that aren't plain names.
func sysmlName(q string) string {
	segs := strings.Split(q, "::")
	for i, s := range segs {
		plain := s != ""
		for j, r := range s {
			if !(r == '_' || unicode.IsLetter(r) || j > 0 && unicode.IsDigit(r)) {
				plain = false
			}
		}
		if !plain {
			segs[i] = "'" + strings.ReplaceAll(s, "'", "\\'") + "'"
		}
	}
	return strings.Join(segs, "::")
}
