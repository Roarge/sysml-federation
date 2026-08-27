package syntax_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

// at returns the span of the first occurrence of sub in src.
func at(src, sub string) syntax.Span {
	i := strings.Index(src, sub)
	if i < 0 {
		panic("substring not in source: " + sub)
	}
	return syntax.Span{Start: i, End: i + len(sub)}
}

// within returns the span of name inside the first occurrence of context in
// src, for names whose first occurrence in src is somewhere else.
func within(src, context, name string) syntax.Span {
	c := at(src, context)
	i := strings.Index(src[c.Start:c.End], name)
	if i < 0 {
		panic("name not in context: " + name)
	}
	return syntax.Span{Start: c.Start + i, End: c.Start + i + len(name)}
}

func parse(t *testing.T, src string) *syntax.File {
	t.Helper()
	f, err := syntax.Parse("m.sysml", src)
	return assert.Must(t, f, err)
}

func TestParsePackageHeaderAndImports(t *testing.T) {
	src := "package <'PIPE'> QueryPipeline {\n" +
		"  doc /* A pipeline.\n   * Two lines. */\n" +
		"  private import ScalarValues::Real;\n" +
		"  public import ISQ::*;\n" +
		"  /* a bare comment is allowed and ignored */\n" +
		"}\n"
	f := parse(t, src)
	assert.Equal(t, f.Name, "m.sysml")
	assert.Equal(t, f.Package.ShortName, "PIPE")
	assert.Equal(t, f.Package.Name, "QueryPipeline")
	assert.Equal(t, f.Package.Doc.Text, "A pipeline. Two lines.")
	assert.DeepEqual(t, f.Package.Imports, []syntax.Import{
		{Path: "ScalarValues::Real", Span: at(src, "private import ScalarValues::Real;")},
		{Path: "ISQ", All: true, Span: at(src, "public import ISQ::*;")},
	})
	assert.Equal(t, f.Package.Span, syntax.Span{Start: 0, End: len(src) - 1})
}

func TestParseDefinitionsAndParts(t *testing.T) {
	src := "package P {\n" +
		"  abstract part def Component { attribute capacity : Real; }\n" +
		"  part def Server :> Component { doc /* s */ attribute throughput : Real; }\n" +
		"  item def Query;\n" +
		"  part <'S1'> ingest : Server { attribute :>> throughput = 2000; }\n" +
		"  part box { part <'X'> inner : Server { attribute redefines throughput = 1.5[kg]; } }\n" +
		"}\n"
	f := parse(t, src)
	p := f.Package
	assert.Len(t, p.PartDefs, 2)
	assert.Equal(t, p.PartDefs[0].Abstract, true)
	assert.Equal(t, p.PartDefs[0].Name, "Component")
	assert.DeepEqual(t, p.PartDefs[0].Attributes, []syntax.AttributeUsage{
		{Name: "capacity", Type: "Real", Span: at(src, "attribute capacity : Real;")},
	})
	assert.Equal(t, p.PartDefs[1].Specializes, "Component")
	assert.Equal(t, p.PartDefs[1].Doc.Text, "s")
	assert.DeepEqual(t, p.ItemDefs, []syntax.ItemDef{{Name: "Query", Span: at(src, "item def Query;")}})
	assert.Len(t, p.Parts, 2)
	assert.Equal(t, p.Parts[0].ShortName, "S1")
	assert.Equal(t, p.Parts[0].Name, "ingest")
	assert.Equal(t, p.Parts[0].Type, "Server")
	assert.DeepEqual(t, p.Parts[0].Attributes, []syntax.AttributeUsage{
		{Name: "throughput", Redefines: true,
			Value: syntax.Literal{Number: 2000, Span: at(src, "2000")},
			Span:  at(src, "attribute :>> throughput = 2000;")},
	})
	inner := p.Parts[1].Parts[0]
	assert.Equal(t, inner.ShortName, "X")
	assert.DeepEqual(t, inner.Attributes[0].Value, syntax.Expr(syntax.Literal{Number: 1.5, Unit: "kg", Span: at(src, "1.5")}))
	assert.Equal(t, inner.Attributes[0].Redefines, true)
}

func TestParseAttributeWithBody(t *testing.T) {
	src := "package P {\n" +
		"  attribute <ms> millisecond : DurationUnit {\n" +
		"    :>> unitConversion : ConversionByPrefix { :>> prefix = milli; :>> referenceUnit = s; }\n" +
		"  }\n" +
		"}\n"
	f := parse(t, src)
	a := f.Package.Attributes[0]
	assert.Equal(t, a.ShortName, "ms")
	assert.Equal(t, a.Name, "millisecond")
	assert.Equal(t, a.Type, "DurationUnit")
	assert.Len(t, a.Body, 1)
	conv := a.Body[0]
	assert.Equal(t, conv.Redefines, true)
	assert.Equal(t, conv.Name, "unitConversion")
	assert.Equal(t, conv.Type, "ConversionByPrefix")
	assert.Len(t, conv.Body, 2)
	// "milli" first occurs inside "millisecond" on the line above, so the span
	// is taken from the binding itself.
	assert.DeepEqual(t, conv.Body[0].Value,
		syntax.Expr(syntax.FeatureChain{Names: []string{"milli"}, Span: within(src, "prefix = milli", "milli")}))
}

func TestParseUnitIsAQualifiedName(t *testing.T) {
	f := parse(t, "package P { part a { attribute :>> m = 2500[SI::kg]; } }")
	lit, ok := f.Package.Parts[0].Attributes[0].Value.(syntax.Literal)
	assert.True(t, ok, "a literal value")
	assert.Equal(t, lit.Unit, "SI::kg")
}

// rejection is one row of the SR-18 table: a source and the position and
// message the parser must report.
type rejection struct {
	Line, Column int
	Message      string
}

func TestSR18_StructuralRejections(t *testing.T) {
	tabletest.Run(t, []tabletest.Case[string, rejection]{
		{Name: "no package", In: "part def A;",
			Want: rejection{1, 1, "expected keyword 'package', found keyword 'part'"}},
		{Name: "trailing text", In: "package P {} part def A;",
			Want: rejection{1, 14, "expected end of file, found keyword 'part'"}},
		{Name: "keyword as a name", In: "package P { part def in; }",
			Want: rejection{1, 22, "expected a name, found keyword 'in'"}},
		{Name: "missing semicolon", In: "package P {\n  private import A::B\n}",
			Want: rejection{3, 1, "expected ';', found '}'"}},
		{Name: "unsupported member in a package", In: "package P { action a; }",
			Want: rejection{1, 13, "keyword 'action' is not supported in a package body"}},
		{Name: "unsupported member in a part definition", In: "package P { part def A { ref b; } }",
			Want: rejection{1, 26, "keyword 'ref' is not supported in a part definition body"}},
		{Name: "unsupported member in a part", In: "package P { part a { requirement r; } }",
			Want: rejection{1, 22, "keyword 'requirement' is not supported in a part body"}},
		{Name: "multiplicity", In: "package P { part def A { part b : B[4..6]; } }",
			Want: rejection{1, 36, "expected ';', found '['"}},
		{Name: "second doc", In: "package P { doc /* a */ doc /* b */ }",
			Want: rejection{1, 25, "only one doc is allowed in a package body"}},
		{Name: "unexpected end of file", In: "package P {",
			Want: rejection{1, 12, "expected '}', found end of file"}},
		{Name: "lexical error surfaces from Parse", In: "package P { part a : ~B; }",
			Want: rejection{1, 22, "unexpected character '~'"}},
	}, func(t *testing.T, in string) rejection {
		_, err := syntax.Parse("m.sysml", in)
		e := assert.ErrorAs[*syntax.Error](t, err)
		assert.Equal(t, e.File, "m.sysml")
		return rejection{e.Line, e.Column, e.Message}
	})
}

func TestParseExpressionPrecedenceAndSpans(t *testing.T) {
	src := "package P { part a { attribute :>> x = (g.r + 1) * 2 - 3 / y; } }"
	f := parse(t, src)
	one := syntax.Literal{Number: 1, Span: at(src, "1")}
	two := syntax.Literal{Number: 2, Span: at(src, "2")}
	three := syntax.Literal{Number: 3, Span: at(src, "3")}
	gr := syntax.FeatureChain{Names: []string{"g", "r"}, Span: at(src, "g.r")}
	y := syntax.FeatureChain{Names: []string{"y"}, Span: within(src, "3 / y", "y")}
	want := syntax.Expr(syntax.Binary{Op: syntax.Sub,
		Left: syntax.Binary{Op: syntax.Mul,
			Left:  syntax.Binary{Op: syntax.Add, Left: gr, Right: one, Span: at(src, "g.r + 1")},
			Right: two, Span: at(src, "(g.r + 1) * 2")},
		Right: syntax.Binary{Op: syntax.Div, Left: three, Right: y, Span: at(src, "3 / y")},
		Span:  at(src, "(g.r + 1) * 2 - 3 / y")})
	assert.DeepEqual(t, f.Package.Parts[0].Attributes[0].Value, want)
}

func TestParseNegativeLiteralSpanIncludesTheSign(t *testing.T) {
	src := "package P { part a { attribute :>> x = -5[ms]; } }"
	f := parse(t, src)
	assert.DeepEqual(t, f.Package.Parts[0].Attributes[0].Value,
		syntax.Expr(syntax.Literal{Number: -5, Unit: "ms", Span: at(src, "-5")}))
}

func TestParsePortsConnectAndSatisfy(t *testing.T) {
	src := "package P {\n" +
		"  port def In { doc /* d */ in item q : Query; }\n" +
		"  port def Out { out item q : Query; inout item r; }\n" +
		"  port def Bare;\n" +
		"  part def S { port input : In; port <'o'> output : Out; }\n" +
		"  part box { part a : S; part b : S; connect a.output to b.input; satisfy req1 by a; }\n" +
		"}\n"
	f := parse(t, src)
	p := f.Package
	assert.Len(t, p.PortDefs, 3)
	assert.Equal(t, p.PortDefs[0].Doc.Text, "d")
	assert.DeepEqual(t, p.PortDefs[0].Items, []syntax.DirectedItem{
		{Direction: syntax.DirectionIn, Name: "q", Type: "Query", Span: at(src, "in item q : Query;")}})
	assert.DeepEqual(t, p.PortDefs[1].Items, []syntax.DirectedItem{
		{Direction: syntax.DirectionOut, Name: "q", Type: "Query", Span: at(src, "out item q : Query;")},
		{Direction: syntax.DirectionInOut, Name: "r", Span: at(src, "inout item r;")}})
	assert.Len(t, p.PortDefs[2].Items, 0)
	assert.DeepEqual(t, p.PartDefs[0].Ports, []syntax.PortUsage{
		{Name: "input", Type: "In", Span: at(src, "port input : In;")},
		{ShortName: "o", Name: "output", Type: "Out", Span: at(src, "port <'o'> output : Out;")}})
	box := p.Parts[0]
	assert.DeepEqual(t, box.Connects, []syntax.Connect{{
		From: syntax.FeatureChain{Names: []string{"a", "output"}, Span: at(src, "a.output")},
		To:   syntax.FeatureChain{Names: []string{"b", "input"}, Span: at(src, "b.input")},
		Span: at(src, "connect a.output to b.input;")}})
	assert.DeepEqual(t, box.Satisfies, []syntax.Satisfy{{
		Requirement: syntax.FeatureChain{Names: []string{"req1"}, Span: at(src, "req1")},
		By:          syntax.FeatureChain{Names: []string{"a"}, Span: within(src, "by a;", "a")},
		Span:        at(src, "satisfy req1 by a;")}})
}

func TestSR18_ExpressionAndPortRejections(t *testing.T) {
	tabletest.Run(t, []tabletest.Case[string, rejection]{
		{Name: "minus before a chain", In: "package P { part a { attribute :>> x = -y; } }",
			Want: rejection{1, 41, "expected a number after '-', found identifier 'y'"}},
		{Name: "function call", In: "package P { part a { attribute :>> x = sum(y); } }",
			Want: rejection{1, 43, "expected ';', found '('"}},
		{Name: "reduce idiom", In: "package P { part a { attribute :>> x = y->reduce min; } }",
			Want: rejection{1, 41, "expected ';', found '-'"}},
		{Name: "unit expression", In: "package P { part a { attribute :>> x = 10[km / L]; } }",
			Want: rejection{1, 46, "expected ']', found '/'"}},
		{Name: "unbalanced parenthesis", In: "package P { part a { attribute :>> x = (1 + 2; } }",
			Want: rejection{1, 46, "expected ')', found ';'"}},
		{Name: "port with a body", In: "package P { part def S { port p : In { attribute t : Real; } } }",
			Want: rejection{1, 38, "expected ';', found '{'"}},
		{Name: "untyped port", In: "package P { part def S { port p; } }",
			Want: rejection{1, 32, "expected ':', found ';'"}},
		{Name: "conjugated port", In: "package P { part def S { port p : ~In; } }",
			Want: rejection{1, 35, "unexpected character '~'"}},
		{Name: "attribute in a port definition", In: "package P { port def In { attribute t : Real; } }",
			Want: rejection{1, 27, "keyword 'attribute' is not supported in a port definition body"}},
		{Name: "unsupported member in an attribute body", In: "package P { part a { attribute x : Real { part b; } } }",
			Want: rejection{1, 43, "keyword 'part' is not supported in an attribute body"}},
		{Name: "non-member in an attribute body", In: "package P { part a { attribute x : Real { 5 } } }",
			Want: rejection{1, 43, "expected a member of an attribute body, found number '5'"}},
		{Name: "connect with a bare part", In: "package P { part box { connect a to b; } }",
			Want: rejection{1, 32, "connect ends must be written as part.port"}},
		{Name: "satisfy without by", In: "package P { part box { satisfy r; } }",
			Want: rejection{1, 33, "expected keyword 'by', found ';'"}},
	}, func(t *testing.T, in string) rejection {
		_, err := syntax.Parse("m.sysml", in)
		e := assert.ErrorAs[*syntax.Error](t, err)
		return rejection{e.Line, e.Column, e.Message}
	})
}

func TestParseRequirements(t *testing.T) {
	src := "package P {\n" +
		"  requirement def R :> Base {\n" +
		"    doc /* d */\n" +
		"    subject target : Component;\n" +
		"    attribute requiredRate : Real;\n" +
		"    require constraint { target.capacity >= requiredRate }\n" +
		"  }\n" +
		"  requirement <'R1'> global : R {\n" +
		"    subject :>> target = pipeline;\n" +
		"    attribute :>> requiredRate = 1500;\n" +
		"    require constraint { 3 == target.x }\n" +
		"  }\n" +
		"}\n"
	f := parse(t, src)
	d := f.Package.RequirementDefs[0]
	assert.Equal(t, d.Name, "R")
	assert.Equal(t, d.Specializes, "Base")
	assert.Equal(t, d.Doc.Text, "d")
	assert.DeepEqual(t, d.Subject, &syntax.Subject{Name: "target", Type: "Component", Span: at(src, "subject target : Component;")})
	assert.Len(t, d.Attributes, 1)
	assert.DeepEqual(t, d.Constraint, &syntax.RequireConstraint{
		Left:  syntax.FeatureChain{Names: []string{"target", "capacity"}, Span: at(src, "target.capacity")},
		Op:    syntax.GE,
		Right: syntax.FeatureChain{Names: []string{"requiredRate"}, Span: within(src, ">= requiredRate }", "requiredRate")},
		Span:  at(src, "require constraint { target.capacity >= requiredRate }")})
	u := f.Package.Requirements[0]
	assert.Equal(t, u.ShortName, "R1")
	assert.Equal(t, u.Type, "R")
	assert.DeepEqual(t, u.Subject, &syntax.Subject{Name: "target",
		Value: &syntax.FeatureChain{Names: []string{"pipeline"}, Span: at(src, "pipeline")},
		Span:  at(src, "subject :>> target = pipeline;")})
	assert.Equal(t, u.Constraint.Op, syntax.EQ)
	_, isLit := u.Constraint.Left.(syntax.Literal)
	assert.True(t, isLit, "a literal on the left")
}

func TestParseDerivationAndVerification(t *testing.T) {
	src := "package P {\n" +
		"  #derivation connection {\n" +
		"    end #original ::> r0;\n" +
		"    end #derive ::> r1;\n" +
		"    end #derive ::> r2;\n" +
		"  }\n" +
		"  verification def T { doc /* v */ subject target : Pipeline; objective { verify r0; } }\n" +
		"  verification <'VC1'> t : T { subject target :> pipeline; }\n" +
		"}\n"
	f := parse(t, src)
	d := f.Package.Derivations[0]
	assert.SliceEqual(t, d.Original.Names, []string{"r0"})
	assert.Len(t, d.Derives, 2)
	assert.SliceEqual(t, d.Derives[1].Names, []string{"r2"})
	assert.Equal(t, d.Span, at(src, "#derivation connection {\n    end #original ::> r0;\n    end #derive ::> r1;\n    end #derive ::> r2;\n  }"))
	vd := f.Package.VerificationDefs[0]
	assert.Equal(t, vd.Doc.Text, "v")
	assert.DeepEqual(t, vd.Objective, &syntax.Objective{
		Verify: syntax.FeatureChain{Names: []string{"r0"}, Span: within(src, "verify r0;", "r0")}, Span: at(src, "objective { verify r0; }")})
	vu := f.Package.Verifications[0]
	assert.Equal(t, vu.ShortName, "VC1")
	assert.Equal(t, vu.Type, "T")
	assert.DeepEqual(t, vu.Subject.Value, &syntax.FeatureChain{Names: []string{"pipeline"}, Span: within(src, ":> pipeline;", "pipeline")})
}

func TestSR18_RequirementRejections(t *testing.T) {
	tabletest.Run(t, []tabletest.Case[string, rejection]{
		{Name: "compound constraint", In: "package P { requirement def R { require constraint { a.b >= c + 1 } } }",
			Want: rejection{1, 63, "expected '}', found '+'"}},
		{Name: "constraint without an operator", In: "package P { requirement def R { require constraint { a.b } } }",
			Want: rejection{1, 58, "expected a comparison operator, found '}'"}},
		{Name: "assume constraint", In: "package P { requirement def R { assume constraint { a > 0 } } }",
			Want: rejection{1, 33, "keyword 'assume' is not supported in a requirement body"}},
		{Name: "require shorthand", In: "package P { requirement r { require other; } }",
			Want: rejection{1, 37, "expected keyword 'constraint', found identifier 'other'"}},
		{Name: "second constraint", In: "package P { requirement def R { require constraint { a > 1 } require constraint { a > 2 } } }",
			Want: rejection{1, 62, "only one require constraint is allowed in a requirement body"}},
		{Name: "unknown metadata", In: "package P { #trace connection { } }",
			Want: rejection{1, 13, "unknown metadata #trace"}},
		{Name: "derivation without original", In: "package P { #derivation connection { end #derive ::> r1; } }",
			Want: rejection{1, 13, "a derivation connection needs exactly one #original end"}},
		{Name: "derivation with two originals", In: "package P { #derivation connection { end #original ::> a; end #original ::> b; end #derive ::> c; } }",
			Want: rejection{1, 13, "a derivation connection needs exactly one #original end"}},
		{Name: "derivation without derive", In: "package P { #derivation connection { end #original ::> a; } }",
			Want: rejection{1, 13, "a derivation connection needs at least one #derive end"}},
		{Name: "typed derivation", In: "package P { #derivation connection : D { end a ::> x; } }",
			Want: rejection{1, 36, "expected '{', found ':'"}},
		{Name: "named objective", In: "package P { verification def T { objective o { verify r; } } }",
			Want: rejection{1, 44, "expected '{', found identifier 'o'"}},
		{Name: "return in a verification", In: "package P { verification def T { return verdict : VerdictKind; } }",
			Want: rejection{1, 34, "keyword 'return' is not supported in a verification body"}},
	}, func(t *testing.T, in string) rejection {
		_, err := syntax.Parse("m.sysml", in)
		e := assert.ErrorAs[*syntax.Error](t, err)
		return rejection{e.Line, e.Column, e.Message}
	})
}

func TestExampleModelParses(t *testing.T) {
	raw, err := os.ReadFile("../../examples/pipeline/model.sysml")
	src := string(assert.Must(t, raw, err))
	parsed, err := syntax.Parse("model.sysml", src)
	f := assert.Must(t, parsed, err)
	p := f.Package
	assert.Equal(t, p.ShortName, "PIPE")
	assert.Len(t, p.Imports, 6)
	assert.Len(t, p.Attributes, 1) // the millisecond
	assert.Len(t, p.PartDefs, 3)
	assert.Len(t, p.PortDefs, 2)
	assert.Len(t, p.RequirementDefs, 2)
	assert.Len(t, p.Parts, 1)
	assert.Len(t, p.Parts[0].Parts, 5)
	assert.Len(t, p.Parts[0].Connects, 5)
	assert.Len(t, p.Parts[0].Satisfies, 7)
	assert.Len(t, p.Requirements, 7)
	assert.Len(t, p.Derivations, 1)
	assert.Len(t, p.Derivations[0].Derives, 5)
	assert.Len(t, p.VerificationDefs, 1)
	assert.Len(t, p.Verifications, 1)
	// Every span indexes real text: the literal 1500 is where it says it is.
	lit := p.Requirements[0].Attributes[0].Value.(syntax.Literal)
	assert.Equal(t, f.Text(lit.Span), "1500")
}
