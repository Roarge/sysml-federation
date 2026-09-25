package main

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

// Element is one page of the wiki.
type Element struct {
	ID    string
	QName string
	Kind  string
	Name  string
	Short string
	Doc   string
	Attrs []Attr
	Meta  []Meta
	Owner string
	File  string
	Line  int
}

// Attr is not built yet.
func (e *Element) Attr(name string) string { return "" }

// Decisions is not built yet.
func (e *Element) Decisions() []Meta { return nil }

// LinkView is one link as an element shows it.
type LinkView struct {
	Rel   string
	Other string
	Via   string
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

// Wiki is the systems model read as pages and links.
type Wiki struct {
	Elements   []*Element
	Unresolved []string
}

// LoadWiki is not built yet.
func LoadWiki(root string) (*Wiki, error) { return &Wiki{}, nil }

// Get is not built yet.
func (w *Wiki) Get(ref string) (*Element, bool) { return nil, false }

// LinksOf is not built yet.
func (w *Wiki) LinksOf(id string) []LinkView { return nil }

// Hash is not built yet.
func (w *Wiki) Hash() string { return "" }

// Without is not built yet.
func (w *Wiki) Without(r Removal) *Wiki { return w }

// Reverse is not built yet.
func Reverse(rel string) string { return "" }
