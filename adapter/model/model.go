// Package model resolves a parsed SysML v2 file into the elements the adapter
// projects. A part carries its attributes, ports, children and connections. A
// requirement carries the quantity, comparison and limit read from its
// constraint. Verification cases and the relationships between the elements
// are resolved as well. The package evaluates bound expressions, checks port
// directions against connection end order, and patches numeric literals in the
// source so the served text and the projection are rebuilt together.
package model

import (
	"errors"
	"os"

	"github.com/Roarge/sysml-federation/adapter/syntax"
)

// Direction and Comparison are the syntax package's enums. Their String forms
// are the projection's enum values.
type (
	// Direction is a feature direction.
	Direction = syntax.Direction

	// Comparison is a constraint operator.
	Comparison = syntax.Comparison
)

// Sentinel errors for mutations. Each is wrapped with the element's identity.
var (
	ErrNotFound     = errors.New("no such element")
	ErrNotEditable  = errors.New("the value is not a literal in the source")
	ErrInvalidValue = errors.New("the value must be a finite, non-negative number")
)

// Attribute is one attribute of a part, declared by its definition or by the
// usage itself. Value is nil when nothing binds it.
type Attribute struct {
	Name       string
	Value      *float64
	Unit       string
	Editable   bool        // true when the bound value is a literal no other value shares
	Expression string      // source text of a bound expression; "" otherwise
	Span       syntax.Span // the literal's span when Editable; zero otherwise
}

// Port is a port usage with the direction its definition's items give it.
type Port struct {
	Name      string
	Direction Direction
}

// Connection is one connect statement, directed from its first end to its
// second. ID is "<from>.<fromPort>-><to>.<toPort>".
type Connection struct {
	ID, From, FromPort, To, ToPort string
}

// Part is one part usage.
type Part struct {
	ID, ShortName, Name string
	Definition          string // the type as written in the source, "" when untyped
	Doc                 string
	Attributes          []Attribute
	Ports               []Port
	Parts               []*Part
	Connections         []Connection
	Satisfies           []string // requirement IDs this part satisfies
}

// Requirement is one requirement usage with its constraint read into a
// quantity, a comparison and a limit.
type Requirement struct {
	ID, ShortName, Name string
	Text                string
	Subject             string // part ID, "" when the usage binds no subject
	Quantity            string
	Comparison          Comparison
	Limit               float64
	LimitUnit           string
	LimitEditable       bool
	DerivedFrom         []string // requirement IDs
	Derives             []string // requirement IDs
	SatisfiedBy         []string // part IDs
	VerifiedBy          []string // verification case IDs

	limitSpan syntax.Span // the limit literal's span when LimitEditable
}

// VerificationCase is one verification usage.
type VerificationCase struct {
	ID, ShortName, Name string
	Verifies            []string // requirement IDs
}

// Model is the resolved projection of one source file. A Model is never
// modified after Parse returns, and every mutation returns a new one.
type Model struct {
	Version           int
	Text              string
	Roots             []*Part
	Parts             []*Part // every part usage, depth first in source order
	Requirements      []*Requirement
	VerificationCases []*VerificationCase

	file     *syntax.File
	parts    map[string]*Part
	reqs     map[string]*Requirement
	vcs      map[string]*VerificationCase
	literals map[syntax.Span]int // how many projected values each literal's span carries
}

// Load reads and parses a file. Errors carry the path as given.
func Load(path string) (*Model, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(path, string(src))
}

// Parse builds a model from source. The name appears in every error. The
// returned model has Version 1. Every refusal is a *syntax.Error.
func Parse(name, src string) (m *Model, err error) {
	f, err := syntax.Parse(name, src)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			se, ok := r.(*syntax.Error)
			if !ok {
				panic(r)
			}
			m, err = nil, se
		}
	}()
	return build(f), nil
}

// Part looks a part up by ID.
func (m *Model) Part(id string) (*Part, bool) { p, ok := m.parts[id]; return p, ok }

// Requirement looks a requirement up by ID.
func (m *Model) Requirement(id string) (*Requirement, bool) { r, ok := m.reqs[id]; return r, ok }

// VerificationCase looks a verification case up by ID.
func (m *Model) VerificationCase(id string) (*VerificationCase, bool) {
	v, ok := m.vcs[id]
	return v, ok
}
