package model

import (
	"os"
	"testing"

	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// TestPatchGuardsItsInputs exercises the internal write. The public setters
// reach it only through setLiteral, which refuses a negative value first.
func TestPatchGuardsItsInputs(t *testing.T) {
	src, err := readExample()
	assert.NoError(t, err)
	loaded, err := Parse("model.sysml", src)
	m := assert.Must(t, loaded, err)
	_, err = m.patch(syntax.Span{Start: 0, End: 7}, "1") // "package" is not a literal
	assert.ErrorIs(t, err, ErrNotEditable)
	p, ok := m.Part("PIPE-S1")
	assert.True(t, ok, "part PIPE-S1")
	span := p.Attributes[1].Span
	_, err = m.patch(span, "12 + 3")
	assert.ErrorIs(t, err, ErrInvalidValue)
	_, err = m.patch(span, "abc")
	assert.ErrorIs(t, err, ErrInvalidValue)
	patched, err := m.patch(span, "-5")
	m2 := assert.Must(t, patched, err)
	np, ok := m2.Part("PIPE-S1")
	assert.True(t, ok, "part PIPE-S1 after the write")
	assert.Equal(t, *np.Attributes[1].Value, -5.0)
	assert.Equal(t, m2.Version, 2)
}

// readExample reads the example model the way the external tests do.
func readExample() (string, error) {
	b, err := os.ReadFile("../../examples/pipeline/model.sysml")
	return string(b), err
}
