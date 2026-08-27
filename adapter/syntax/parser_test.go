package syntax_test

import (
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
