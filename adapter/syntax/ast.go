package syntax

// File is one parsed source file. Name is what errors report and Src is the
// text every Span indexes into.
type File struct {
	Name    string
	Src     string
	Package Package
}

// ErrorAt builds a positioned refusal for a span of this file. The model
// package uses it so every refusal has the same shape.
func (f *File) ErrorAt(at Span, message string) *Error {
	return ErrorAt(f.Name, f.Src, at, message)
}

// Text returns the source text a span covers.
func (f *File) Text(at Span) string { return f.Src[at.Start:at.End] }

// Package is the one package a file declares.
type Package struct {
	ShortName, Name  string
	Doc              *Doc
	Imports          []Import
	Attributes       []AttributeUsage // package-level attributes, such as a unit declaration
	ItemDefs         []ItemDef
	PortDefs         []PortDef
	PartDefs         []PartDef
	RequirementDefs  []RequirementDef
	VerificationDefs []VerificationDef
	Parts            []PartUsage
	Requirements     []RequirementUsage
	Verifications    []VerificationUsage
	Derivations      []DerivationConnection
	Span             Span
}

// Import is `private import A::B;` or `private import A::*;`.
type Import struct {
	Path string // the qualified name as written, without a trailing ::*
	All  bool   // true for ::*
	Span Span
}

// Doc is a `doc /* ... */` element. Text is the comment body with its
// delimiters, leading asterisks and line breaks normalised to single spaces.
type Doc struct {
	Text string
	Span Span
}

// ItemDef is `item def Name;`.
type ItemDef struct {
	ShortName, Name string
	Span            Span
}

// Direction is a feature direction.
type Direction int

// The three directions. The String form matches the projection's enum.
const (
	DirectionIn Direction = iota + 1
	DirectionOut
	DirectionInOut
)

// String returns the projection's spelling of the direction: "IN", "OUT" or
// "INOUT".
func (d Direction) String() string {
	switch d {
	case DirectionIn:
		return "IN"
	case DirectionOut:
		return "OUT"
	default:
		return "INOUT"
	}
}

// DirectedItem is `in item queries : Query;` inside a port definition.
type DirectedItem struct {
	Direction  Direction
	Name, Type string
	Span       Span
}

// PortDef is `port def Name { ... }` or `port def Name;`.
type PortDef struct {
	ShortName, Name string
	Doc             *Doc
	Items           []DirectedItem
	Span            Span
}

// PartDef is `part def Name :> Base { ... }`.
type PartDef struct {
	Abstract        bool
	ShortName, Name string
	Specializes     string // the name after :>, or ""
	Doc             *Doc
	Attributes      []AttributeUsage
	Ports           []PortUsage
	Parts           []PartUsage
	Span            Span
}

// PartUsage is `part <'S1'> name : Type { ... }`.
type PartUsage struct {
	ShortName, Name, Type string
	Doc                   *Doc
	Attributes            []AttributeUsage
	Ports                 []PortUsage
	Parts                 []PartUsage
	Connects              []Connect
	Satisfies             []Satisfy
	Span                  Span
}

// AttributeUsage is a declaration `attribute x : T;`, a binding
// `attribute :>> x = expr;` or either with a nested body.
type AttributeUsage struct {
	ShortName, Name, Type string
	Redefines             bool
	Value                 Expr // nil when unbound
	Body                  []AttributeUsage
	Span                  Span
}

// PortUsage is `port name : Type;`.
type PortUsage struct {
	ShortName, Name, Type string
	Span                  Span
}

// Connect is `connect a.x to b.y;`.
type Connect struct {
	From, To FeatureChain
	Span     Span
}

// Satisfy is `satisfy req by part;`.
type Satisfy struct {
	Requirement, By FeatureChain
	Span            Span
}

// Subject is `subject name : Type;`, `subject :>> name = chain;` or
// `subject name :> chain;`. Value is nil for the declaration form.
type Subject struct {
	Name, Type string
	Value      *FeatureChain
	Span       Span
}

// Comparison is a constraint operator. The String form matches the
// projection's enum.
type Comparison int

// The comparison operators.
const (
	GE Comparison = iota + 1
	GT
	LE
	LT
	EQ
)

// String returns the projection's spelling of the operator: "GE", "GT", "LE",
// "LT" or "EQ".
func (c Comparison) String() string {
	return [...]string{"", "GE", "GT", "LE", "LT", "EQ"}[c]
}

// RequireConstraint is `require constraint { left op right }`. Each operand is
// a FeatureChain or a Literal, and the parser refuses anything else.
type RequireConstraint struct {
	Left  Expr
	Op    Comparison
	Right Expr
	Span  Span
}

// RequirementBody is what a requirement definition and usage share.
type RequirementBody struct {
	Doc        *Doc
	Subject    *Subject
	Attributes []AttributeUsage
	Constraint *RequireConstraint
}

// RequirementDef is `requirement def Name :> Base { ... }`.
type RequirementDef struct {
	ShortName, Name, Specializes string
	RequirementBody
	Span Span
}

// RequirementUsage is `requirement <'R1'> name : Def { ... }`.
type RequirementUsage struct {
	ShortName, Name, Type string
	RequirementBody
	Span Span
}

// Objective is `objective { verify req; }`.
type Objective struct {
	Verify FeatureChain
	Span   Span
}

// VerificationBody is what a verification definition and usage share.
type VerificationBody struct {
	Doc       *Doc
	Subject   *Subject
	Objective *Objective
}

// VerificationDef is `verification def Name { ... }`.
type VerificationDef struct {
	ShortName, Name string
	VerificationBody
	Span Span
}

// VerificationUsage is `verification <'VC1'> name : Def { ... }`.
type VerificationUsage struct {
	ShortName, Name, Type string
	VerificationBody
	Span Span
}

// DerivationConnection is `#derivation connection { end #original ::> a; end
// #derive ::> b; }` with exactly one original end.
type DerivationConnection struct {
	Original FeatureChain
	Derives  []FeatureChain
	Span     Span
}

// Expr is an expression tree: a Literal, a FeatureChain or a Binary.
type Expr interface {
	expr()
}

// Literal is a number with an optional bracketed unit. Span covers the number
// and a leading minus sign only, never the unit, so a patch replaces exactly
// the digits.
type Literal struct {
	Number float64
	Unit   string
	Span   Span
}

// FeatureChain is `a.b.c`.
type FeatureChain struct {
	Names []string
	Span  Span
}

// Operator is a binary arithmetic operator.
type Operator int

// The four operators.
const (
	Add Operator = iota + 1
	Sub
	Mul
	Div
)

// Binary is `left op right`.
type Binary struct {
	Op          Operator
	Left, Right Expr
	Span        Span
}

func (Literal) expr()      {}
func (FeatureChain) expr() {}
func (Binary) expr()       {}
