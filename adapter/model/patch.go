package model

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Roarge/sysml-federation/adapter/syntax"
)

// Patch replaces the literal at a span with new text and re-parses, returning
// a new model with the version incremented. The span must be one that exactly
// one projected value carries, and the text must be a plain decimal number.
func (m *Model) Patch(at syntax.Span, literal string) (*Model, error) {
	if m.literals[at] != 1 {
		return nil, fmt.Errorf("%w: no literal at %d..%d", ErrNotEditable, at.Start, at.End)
	}
	if _, err := strconv.ParseFloat(literal, 64); err != nil ||
		strings.Trim(literal, "-0123456789.") != "" || strings.Count(literal, "-") > 1 {
		return nil, fmt.Errorf("%w: %q is not a plain number", ErrInvalidValue, literal)
	}
	text := m.Text[:at.Start] + literal + m.Text[at.End:]
	next, err := Parse(m.file.Name, text)
	if err != nil {
		return nil, err
	}
	next.Version = m.Version + 1
	return next, nil
}

// SetAttribute sets a part's attribute to a value that is a literal in the
// source. It refuses an expression-bound or unbound attribute (SR-24) and a
// value that is not a finite, non-negative number (SR-25).
func (m *Model) SetAttribute(partID, name string, value float64) (*Model, error) {
	p, ok := m.parts[partID]
	if !ok {
		return nil, fmt.Errorf("%w: part %q", ErrNotFound, partID)
	}
	for _, a := range p.Attributes {
		if a.Name != name {
			continue
		}
		if !a.Editable {
			return nil, fmt.Errorf("%w: attribute %q of part %q", ErrNotEditable, name, partID)
		}
		return m.setLiteral(a.Span, value)
	}
	return nil, fmt.Errorf("%w: attribute %q of part %q", ErrNotFound, name, partID)
}

// SetLimit sets a requirement's limit where it is a literal in the source.
func (m *Model) SetLimit(requirementID string, value float64) (*Model, error) {
	r, ok := m.reqs[requirementID]
	if !ok {
		return nil, fmt.Errorf("%w: requirement %q", ErrNotFound, requirementID)
	}
	if !r.LimitEditable {
		return nil, fmt.Errorf("%w: limit of requirement %q", ErrNotEditable, requirementID)
	}
	return m.setLiteral(r.limitSpan, value)
}

func (m *Model) setLiteral(at syntax.Span, value float64) (*Model, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return nil, fmt.Errorf("%w: got %v", ErrInvalidValue, value)
	}
	value += 0 // a negative zero becomes zero
	return m.Patch(at, strconv.FormatFloat(value, 'f', -1, 64))
}
