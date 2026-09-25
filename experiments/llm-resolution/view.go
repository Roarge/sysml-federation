package main

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

// Expose is not built yet.
func (v *View) Expose(w *Wiki, id, note string) error { return nil }

// Prune is not built yet.
func (v *View) Prune(id string) bool { return false }

// Lines is not built yet.
func (v View) Lines() string { return "" }

// SysML is not built yet.
func (v View) SysML(w *Wiki, name string) string { return "" }
