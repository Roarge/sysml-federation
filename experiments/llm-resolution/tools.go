package main

// The two tasks.
const (
	TaskTest     = "test"
	TaskIncident = "incident"
)

// The tools' limits.
const (
	LimitFind      = 8
	LimitPaths     = 3
	LimitDoc       = 800
	LimitReadLines = 30
	LimitAnswer    = 1500
)

// Scope is what one browse's tools hide.
type Scope struct {
	Task     string
	Redact   map[string]bool
	HideTest string
}

// CodeBase is the built system's files.
type CodeBase struct {
	Root string
}

// LoadCode is not built yet.
func LoadCode(root string) (*CodeBase, error) { return &CodeBase{Root: root}, nil }

// ToolKit answers one browse's tool calls.
type ToolKit struct {
	scope Scope
}

// NewToolKit is not built yet.
func NewToolKit(w *Wiki, code *CodeBase, reqs []Requirement, sc Scope) *ToolKit {
	return &ToolKit{scope: sc}
}

// Call is not built yet.
func (k *ToolKit) Call(tool, arg, arg2 string) string { return "" }

// redact is not built yet.
func redact(text string, drop map[string]bool) string { return text }
