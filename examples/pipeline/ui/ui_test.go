package ui_test

import (
	"io/fs"
	"testing"

	"github.com/Roarge/sysml-federation/examples/pipeline/ui"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// The apps are not written yet. Until they are, each app directory holds one
// page, so the UI server serves a document at /viewer/ and /document/.
func TestBothAppPagesAreEmbedded(t *testing.T) {
	for _, name := range []string{"viewer/index.html", "document/index.html"} {
		data, err := fs.ReadFile(ui.Files, name)
		got := assert.Must(t, data, err)
		assert.True(t, len(got) > 0, name+" is not empty")
	}
}
