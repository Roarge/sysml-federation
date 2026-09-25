package main

// The incident probe and its variants.
const (
	ProbeIncident    = "incident"
	VariantReported  = "as reported"
	VariantAgain     = "again"
	VariantAlertOnly = "alert only"
	VariantRemoved   = "transition removed"
	VariantControl   = "control link removed"
)

// KeyItem is one thing a good account names.
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

// LoadKey is not built yet.
func LoadKey(path string) (Key, error) { return Key{}, errNotBuilt }

// Check is not built yet.
func (k Key) Check(w *Wiki, code *CodeBase) error { return errNotBuilt }

// ScoreIncident is not built yet.
func ScoreIncident(k Key, a IncidentAnswer, w *Wiki) []KeyResult { return nil }

// CheckCitations is not built yet.
func CheckCitations(a IncidentAnswer, kit *ToolKit) []Citation { return nil }

// CheckTestCitations is not built yet.
func CheckTestCitations(a TestAnswer, kit *ToolKit) []Citation { return nil }
