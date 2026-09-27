package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// The incident is the one outage the experiment puts to the language model:
// what the engineer on call reports, and the key, written and committed
// before any run, of what a good account names. An account is scored item by
// item against the key, and every element, link and file it cites is checked
// against what the run's tools could show.

// The incident probe and its variants.
const (
	ProbeIncident    = "incident"
	VariantReported  = "as reported"
	VariantAgain     = "again"
	VariantAlertOnly = "alert only"
	VariantRemoved   = "transition removed"
	VariantControl   = "control link removed"
)

// IncidentVariants are the five browses of the incident, in the order a run
// makes them.
var IncidentVariants = []string{VariantReported, VariantAgain, VariantAlertOnly, VariantRemoved, VariantControl}

// KeyItem is one thing a good account names, in one of its fields: any of
// the elements, by qualified name or short name, or any of the files.
type KeyItem struct {
	Item  string   `json:"item"`
	Field string   `json:"field"`
	Any   []string `json:"any"`
}

// KnownMismatch is a place the systems model is known to be wrong.
type KnownMismatch struct {
	ID   string `json:"id"`
	Note string `json:"note"`
}

// Key is the incident as reported and what a good account names.
type Key struct {
	Report          string          `json:"report"`
	AlertOnly       string          `json:"alert_only"`
	Items           []KeyItem       `json:"items"`
	RemoveElement   string          `json:"remove_element"`
	ControlLink     LinkRef         `json:"control_link"`
	KnownMismatches []KnownMismatch `json:"known_mismatches,omitempty"`
}

// KeyResult is one key item, found or missed.
type KeyResult struct {
	Item  string `json:"item"`
	Found bool   `json:"found"`
}

// Citation is one cited element, link or file, checked.
type Citation struct {
	Field string `json:"field"`
	Ref   string `json:"ref"`
	Found bool   `json:"found"`
	Why   string `json:"why,omitempty"`
}

// LoadKey reads the incident's key.
func LoadKey(path string) (Key, error) {
	var k Key
	data, err := os.ReadFile(path)
	if err != nil {
		return k, err
	}
	if err := json.Unmarshal(data, &k); err != nil {
		return k, fmt.Errorf("%s: %w", path, err)
	}
	return k, nil
}

// isFile tells a file's path from an element's name.
func isFile(ref string) bool { return strings.Contains(ref, "/") }

var lineSuffix = regexp.MustCompile(`:\d+(-\d+)?$`)

// filePart is a cited file without its line.
func filePart(ref string) string {
	return strings.TrimPrefix(lineSuffix.ReplaceAllString(strings.TrimSpace(ref), ""), "./")
}

// Check makes sure the key names only elements of the systems model and
// files of the checkout, and that the removals it asks for have something to
// remove.
func (k Key) Check(w *Wiki, code *CodeBase) error {
	var errs []error
	if strings.TrimSpace(k.Report) == "" || strings.TrimSpace(k.AlertOnly) == "" {
		errs = append(errs, errors.New("the key needs the report and the alert alone"))
	}
	fields := map[string]bool{"cause": true, "mechanism": true, "code": true, "consequences": true, "path": true}
	for _, it := range k.Items {
		if !fields[it.Field] {
			errs = append(errs, fmt.Errorf("%q is in no field of an account: %q", it.Item, it.Field))
		}
		for _, ref := range it.Any {
			if isFile(ref) {
				if ok, _ := code.exists(filePart(ref)); !ok {
					errs = append(errs, fmt.Errorf("%q names %s, which the checkout lacks", it.Item, ref))
				}
			} else if _, ok := w.Get(ref); !ok {
				errs = append(errs, fmt.Errorf("%q names %s, which the systems model lacks", it.Item, ref))
			}
		}
	}
	if _, ok := w.Get(k.RemoveElement); !ok {
		errs = append(errs, fmt.Errorf("the element to remove, %s, isn't in the systems model", k.RemoveElement))
	}
	from, okF := w.Get(k.ControlLink.From)
	to, okT := w.Get(k.ControlLink.To)
	if !okF || !okT || !hasLinkTo(w, from, k.ControlLink.Rel, to) {
		errs = append(errs, fmt.Errorf("the control link %s %s %s isn't in the systems model", k.ControlLink.From, k.ControlLink.Rel, k.ControlLink.To))
	}
	for _, m := range k.KnownMismatches {
		if _, ok := w.Get(m.ID); !ok {
			errs = append(errs, fmt.Errorf("the known mismatch names %s, which the systems model lacks", m.ID))
		}
	}
	return errors.Join(errs...)
}

func hasLinkTo(w *Wiki, from *Element, rel string, to *Element) bool {
	for _, l := range w.links[from] {
		if l.rel == rel && l.other == to {
			return true
		}
	}
	return false
}

// fieldOf gives an account's citations in one field.
func fieldOf(a IncidentAnswer, field string) []string {
	switch field {
	case "cause":
		return a.Cause
	case "mechanism":
		return a.Mechanism
	case "code":
		return a.Code
	case "consequences":
		return a.Consequences
	case "path":
		return a.Path
	}
	return nil
}

// ScoreIncident marks each item of the key found when the account cites any
// of its elements or files in the item's own field. Elements are looked up in
// the whole systems model, so an account that cites an element a probe took
// away is still scored on what it says, and the citation check says the rest.
func ScoreIncident(k Key, a IncidentAnswer, w *Wiki) []KeyResult {
	var out []KeyResult
	for _, it := range k.Items {
		found := false
		for _, cited := range fieldOf(a, it.Field) {
			for _, want := range it.Any {
				if sameRef(w, cited, want) {
					found = true
				}
			}
		}
		out = append(out, KeyResult{Item: it.Item, Found: found})
	}
	return out
}

func sameRef(w *Wiki, cited, want string) bool {
	if isFile(want) || isFile(cited) {
		return filePart(cited) == filePart(want)
	}
	x, ok1 := w.Get(cited)
	y, ok2 := w.Get(want)
	return ok1 && ok2 && x == y
}

// CheckCitations checks every element and file an account cites against what
// the browse's tools could show, and every step of its path against the link
// or the code that should join it to the step before.
func CheckCitations(a IncidentAnswer, kit *ToolKit) []Citation {
	var out []Citation
	for _, field := range []string{"cause", "mechanism", "consequences", "code"} {
		for _, ref := range fieldOf(a, field) {
			out = append(out, kit.cite(field, ref))
		}
	}
	var prev string
	for i, ref := range a.Path {
		c := kit.cite("path", ref)
		if i > 0 && c.Found {
			if why := kit.joined(prev, ref); why != "" {
				c.Found, c.Why = false, why
			}
		}
		out = append(out, c)
		prev = ref
	}
	return out
}

// CheckTestCitations checks the elements a test's final answer links to.
func CheckTestCitations(a TestAnswer, kit *ToolKit) []Citation {
	var out []Citation
	for _, l := range a.Links {
		out = append(out, kit.cite("links", l.ID))
	}
	return out
}

// cite checks that one cited element or file exists for this browse.
func (k *ToolKit) cite(field, ref string) Citation {
	c := Citation{Field: field, Ref: ref}
	if isFile(ref) {
		if ok, _ := k.code.exists(filePart(ref)); ok {
			c.Found = true
		} else {
			c.Why = "no such file in the checkout"
		}
		return c
	}
	if _, ok := k.get(ref); ok {
		c.Found = true
	} else {
		c.Why = "no such element in what the tools could show"
	}
	return c
}

// joined says why two consecutive steps of a path aren't joined, or "" when
// they are: two elements by a link, an element and a file by the code the
// element names, and two files by the second having been read or found by a
// search in the browse.
func (k *ToolKit) joined(a, b string) string {
	switch {
	case isFile(a) && isFile(b):
		if k.looked[filePart(b)] {
			return ""
		}
		return fmt.Sprintf("%s was neither read nor found by a search in this browse", b)
	case isFile(b):
		e, ok := k.get(a)
		if ok && k.codePaths(e)[filePart(b)] {
			return ""
		}
		return fmt.Sprintf("the code the systems model gives for %s doesn't include %s", a, b)
	case isFile(a):
		e, ok := k.get(b)
		if ok && k.codePaths(e)[filePart(a)] {
			return ""
		}
		return fmt.Sprintf("the code the systems model gives for %s doesn't include %s", b, a)
	}
	x, okX := k.get(a)
	y, okY := k.get(b)
	if !okX || !okY {
		return fmt.Sprintf("no link joins %s and %s", a, b)
	}
	if x == y || k.w.linked(x, y) {
		return ""
	}
	return fmt.Sprintf("no link joins %s and %s", a, b)
}
