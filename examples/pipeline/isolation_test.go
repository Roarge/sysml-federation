package pipeline

import (
	"testing"

	"github.com/99designs/gqlgen/client"

	"github.com/Roarge/sysml-federation/adapter/projection"
	"github.com/Roarge/sysml-federation/examples/pipeline/document"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// mutationOut absorbs a mutation's answer under whatever field name it
// carries. The test client refuses a key its target has no place for, so a
// mutation cannot be posted into an empty struct.
type mutationOut map[string]struct {
	Version int `json:"version"`
}

// TestSR36_DocumentOperationsLeaveTheModelVersion runs every editorial
// operation against the document service beside a live adapter store and
// reads the model version after each one.
func TestSR36_DocumentOperationsLeaveTheModelVersion(t *testing.T) {
	loaded, err := projection.Load("model.sysml")
	store := assert.Must(t, loaded, err)
	created, err := document.New()
	svc := assert.Must(t, created, err)
	c := client.New(document.Handler(svc), client.Path("/graphql"))
	operations := []string{
		`mutation { moveNode(id: "PIPE-R1.5", parentId: "PIPE-R1", index: 0) { version } }`,
		`mutation { moveNode(id: "PIPE-R2", parentId: "PIPE-R1", index: 5) { version } }`,
		`mutation { insertHeading(aboveId: "PIPE-R1", text: "Performance") { version } }`,
		`mutation { addProse(parentId: null, index: 0, text: "Why allocated limits fail.") { version } }`,
		`mutation { editText(id: "intro", text: "Rewritten.") { version } }`,
		`mutation { excludeRequirement(requirementId: "PIPE-R1.4") { version } }`,
		`mutation { includeRequirement(requirementId: "PIPE-R1.4") { version } }`,
		`mutation { resetDocument { version } }`,
	}
	for i, op := range operations {
		var out mutationOut
		c.MustPost(op, &out)
		assert.Equal(t, svc.Version(), i+2)
		assert.Equal(t, store.Version(), 1)
	}
}
