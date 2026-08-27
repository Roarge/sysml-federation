package model_test

import (
	"testing"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

const examplePath = "../../examples/pipeline/model.sysml"

func loadExample(t *testing.T) *model.Model {
	t.Helper()
	m, err := model.Load(examplePath)
	return assert.Must(t, m, err)
}

func parse(t *testing.T, src string) *model.Model {
	t.Helper()
	m, err := model.Parse("m.sysml", src)
	return assert.Must(t, m, err)
}

func f64(v float64) *float64 { return &v }

// refusal is what a model-level rejection must report.
type refusal struct {
	Line, Column int
	Message      string
}

func refuse(t *testing.T, src string) refusal {
	t.Helper()
	_, err := model.Parse("m.sysml", src)
	e := assert.ErrorAs[*syntax.Error](t, err)
	assert.Equal(t, e.File, "m.sysml")
	return refusal{e.Line, e.Column, e.Message}
}

func TestSR21_PartIdentifiersAreShortNamesWithQualifiedFallback(t *testing.T) {
	m := loadExample(t)
	var ids []string
	for _, p := range m.Parts {
		ids = append(ids, p.ID)
	}
	assert.SliceEqual(t, ids, []string{"PIPE-P1", "PIPE-S1", "PIPE-S2", "PIPE-S3", "PIPE-S4", "PIPE-S5"})

	// Without a short name the id is the qualified name.
	src := "package Plant { part def S; part line : S { part <'A'> a : S; part b : S; } }"
	n := parse(t, src)
	assert.Equal(t, n.Roots[0].ID, "Plant::line")
	assert.Equal(t, n.Roots[0].Parts[0].ID, "A")
	assert.Equal(t, n.Roots[0].Parts[1].ID, "Plant::line::b")
	assert.Equal(t, n.Roots[0].Parts[1].ShortName, "")
	assert.Equal(t, n.Roots[0].Parts[1].Name, "b")
}

func TestPartsTreeAttributesAndPorts(t *testing.T) {
	m := loadExample(t)
	assert.Equal(t, m.Version, 1)
	assert.Len(t, m.Roots, 1)
	pipe := m.Roots[0]
	assert.Equal(t, pipe.Name, "pipeline")
	assert.Equal(t, pipe.Definition, "Pipeline")
	assert.Equal(t, pipe.Doc, "Queries enter at ingest, are parsed, indexed on two servers in parallel and served.")
	assert.Len(t, pipe.Parts, 5)
	// Attributes come from the definition chain, base first, unbound where
	// nothing binds them.
	assert.DeepEqual(t, pipe.Attributes, []model.Attribute{
		{Name: "capacity"},
		{Name: "latency"},
	})
	ingest := pipe.Parts[0]
	assert.Equal(t, ingest.Definition, "Server")
	assert.Len(t, ingest.Attributes, 2)
	assert.Equal(t, ingest.Attributes[0].Name, "capacity")
	th := ingest.Attributes[1]
	assert.Equal(t, th.Name, "throughput")
	assert.Equal(t, *th.Value, 2000.0)
	assert.Equal(t, th.Editable, true)
	assert.Equal(t, th.Expression, "")
	assert.Equal(t, m.Text[th.Span.Start:th.Span.End], "2000")
	assert.DeepEqual(t, ingest.Ports, []model.Port{
		{Name: "input", Direction: syntax.DirectionIn},
		{Name: "output", Direction: syntax.DirectionOut},
	})
	got, ok := m.Part("PIPE-S4")
	assert.True(t, ok, "PIPE-S4 is found")
	assert.Equal(t, *got.Attributes[1].Value, 700.0)
	_, ok = m.Part("nope")
	assert.Equal(t, ok, false)
}

func TestAttributeShapes(t *testing.T) {
	src := "package P {\n" +
		"  part def D { attribute a : Real; attribute b : Real = 3[kg]; }\n" +
		"  port def Mixed { in item x; out item y; }\n" +
		"  port def None;\n" +
		"  part <'u'> u : D { attribute :>> a = 1.5; attribute own : Real = 2; attribute unbound : Real; port m : Mixed; port n : None; }\n" +
		"}\n"
	m := parse(t, src)
	u := m.Roots[0]
	assert.DeepEqual(t, u.Attributes, []model.Attribute{
		{Name: "a", Value: f64(1.5), Editable: true, Span: syntax.Span{Start: 177, End: 180}},
		{Name: "b", Value: f64(3), Unit: "kg", Editable: true, Span: syntax.Span{Start: 68, End: 69}},
		{Name: "own", Value: f64(2), Editable: true, Span: syntax.Span{Start: 205, End: 206}},
		{Name: "unbound"},
	})
	// A mixed or empty port definition projects INOUT.
	assert.DeepEqual(t, u.Ports, []model.Port{
		{Name: "m", Direction: syntax.DirectionInOut},
		{Name: "n", Direction: syntax.DirectionInOut},
	})
}

func TestBuildRefusals(t *testing.T) {
	tabletest.Run(t, []tabletest.Case[string, refusal]{
		{Name: "duplicate id", In: "package P { part def S; part <'A'> a : S; part <'A'> b : S; }",
			Want: refusal{1, 43, `duplicate id "A"`}},
		{Name: "unknown part definition", In: "package P { part a : Nope; }",
			Want: refusal{1, 13, `unresolved definition "Nope"`}},
		{Name: "unknown base definition", In: "package P { part def S :> Nope; part a : S; }",
			Want: refusal{1, 33, `unresolved definition "Nope"`}},
		{Name: "cyclic specialization", In: "package P { part def A :> B; part def B :> A; part a : A; }",
			Want: refusal{1, 47, `cyclic specialization of "A"`}},
		{Name: "redefining an undeclared attribute", In: "package P { part def S; part a : S { attribute :>> x = 1; } }",
			Want: refusal{1, 38, `redefined attribute "x" is not declared by "S"`}},
		{Name: "unknown port definition", In: "package P { part def S { port p : Nope; } part a : S; }",
			Want: refusal{1, 26, `unresolved definition "Nope"`}},
	}, refuse)
}
