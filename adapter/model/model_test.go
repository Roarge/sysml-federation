package model_test

import (
	"fmt"
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

func TestSR21_RequirementAndVerificationIdentifiers(t *testing.T) {
	m := loadExample(t)
	var rids []string
	for _, r := range m.Requirements {
		rids = append(rids, r.ID)
	}
	assert.SliceEqual(t, rids, []string{"PIPE-R1", "PIPE-R1.1", "PIPE-R1.2", "PIPE-R1.3", "PIPE-R1.4", "PIPE-R1.5", "PIPE-R2"})
	assert.Equal(t, m.VerificationCases[0].ID, "PIPE-VC1")
}

func TestSR20_ConnectionsFollowEndOrderAndPortDirections(t *testing.T) {
	m := loadExample(t)
	pipe := m.Roots[0]
	assert.DeepEqual(t, pipe.Connections, []model.Connection{
		{ID: "PIPE-S1.output->PIPE-S2.input", From: "PIPE-S1", FromPort: "output", To: "PIPE-S2", ToPort: "input"},
		{ID: "PIPE-S2.output->PIPE-S3.input", From: "PIPE-S2", FromPort: "output", To: "PIPE-S3", ToPort: "input"},
		{ID: "PIPE-S2.output->PIPE-S4.input", From: "PIPE-S2", FromPort: "output", To: "PIPE-S4", ToPort: "input"},
		{ID: "PIPE-S3.output->PIPE-S5.input", From: "PIPE-S3", FromPort: "output", To: "PIPE-S5", ToPort: "input"},
		{ID: "PIPE-S4.output->PIPE-S5.input", From: "PIPE-S4", FromPort: "output", To: "PIPE-S5", ToPort: "input"},
	})
	for _, child := range pipe.Parts {
		assert.Len(t, child.Connections, 0)
	}

	const head = "package P { port def I { in item q; } port def O { out item q; } port def B { inout item q; }\n" +
		"  part def S { port i : I; port o : O; port b : B; }\n"
	tabletest.Run(t, []tabletest.Case[string, refusal]{
		{Name: "reversed", In: head + "  part box { part a : S; part b : S; connect b.i to a.o; } }",
			Want: refusal{3, 46, `first end "b.i" is not an out port`}},
		{Name: "second end not in", In: head + "  part box { part a : S; part b : S; connect a.o to b.o; } }",
			Want: refusal{3, 53, `second end "b.o" is not an in port`}},
		{Name: "inout at an end", In: head + "  part box { part a : S; part b : S; connect a.b to b.i; } }",
			Want: refusal{3, 46, `first end "a.b" is not an out port`}},
		{Name: "unknown child", In: head + "  part box { part a : S; connect a.o to c.i; } }",
			Want: refusal{3, 41, `unresolved name "c"`}},
		{Name: "unknown port", In: head + "  part box { part a : S; part b : S; connect a.o to b.x; } }",
			Want: refusal{3, 53, `"b" has no port "x"`}},
		{Name: "duplicate connection", In: head + "  part box { part a : S; part b : S; connect a.o to b.i; connect a.o to b.i; } }",
			Want: refusal{3, 58, `duplicate connection "P::box::a.o->P::box::b.i"`}},
	}, refuse)
}

func TestRelationships(t *testing.T) {
	m := loadExample(t)
	got, err := req(m, "PIPE-R1")
	r1 := assert.Must(t, got, err)
	assert.Equal(t, r1.Name, "globalThroughput")
	assert.Equal(t, r1.ShortName, "PIPE-R1")
	assert.Equal(t, r1.Text, "The pipeline shall sustain the required query rate")
	assert.Equal(t, r1.Subject, "PIPE-P1")
	assert.SliceEqual(t, r1.Derives, []string{"PIPE-R1.1", "PIPE-R1.2", "PIPE-R1.3", "PIPE-R1.4", "PIPE-R1.5"})
	assert.Len(t, r1.DerivedFrom, 0)
	assert.SliceEqual(t, r1.SatisfiedBy, []string{"PIPE-P1"})
	assert.Len(t, r1.VerifiedBy, 0)
	got, err = req(m, "PIPE-R1.3")
	r13 := assert.Must(t, got, err)
	assert.Equal(t, r13.Subject, "PIPE-S3")
	assert.SliceEqual(t, r13.DerivedFrom, []string{"PIPE-R1"})
	assert.SliceEqual(t, r13.SatisfiedBy, []string{"PIPE-S3"})
	got, err = req(m, "PIPE-R2")
	r2 := assert.Must(t, got, err)
	assert.SliceEqual(t, r2.VerifiedBy, []string{"PIPE-VC1"})
	assert.SliceEqual(t, r2.SatisfiedBy, []string{"PIPE-P1"})
	vc := m.VerificationCases[0]
	assert.Equal(t, vc.Name, "latencyTest")
	assert.SliceEqual(t, vc.Verifies, []string{"PIPE-R2"})
	pipe := m.Roots[0]
	assert.SliceEqual(t, pipe.Satisfies, []string{"PIPE-R1", "PIPE-R2"})
	assert.SliceEqual(t, pipe.Parts[0].Satisfies, []string{"PIPE-R1.1"})

	// A requirement without a bound subject projects an empty subject and a
	// usage's doc falls back to its definition's.
	n := parse(t, "package P { requirement def R { doc /* from def */ subject s : X; attribute l : Real; require constraint { s.q <= l } }\n"+
		"  requirement <'r'> r : R { attribute :>> l = 1; } }")
	assert.Equal(t, n.Requirements[0].Subject, "")
	assert.Equal(t, n.Requirements[0].Text, "from def")
}

func req(m *model.Model, id string) (*model.Requirement, error) {
	r, ok := m.Requirement(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", model.ErrNotFound, id)
	}
	return r, nil
}

func TestLinkRefusals(t *testing.T) {
	const head = "package P { part def S; part <'a'> a : S; requirement def R { subject s : S; attribute l : Real; require constraint { s.x >= l } }\n" +
		"  requirement <'r'> r : R { subject :>> s = a; attribute :>> l = 1; }\n"
	tabletest.Run(t, []tabletest.Case[string, refusal]{
		{Name: "subject is not a part", In: head + "  requirement <'q'> q : R { subject :>> s = r; attribute :>> l = 1; } }",
			Want: refusal{3, 45, `"r" is not a part`}},
		{Name: "unknown requirement definition", In: head + "  requirement <'q'> q : Nope { } }",
			Want: refusal{3, 3, `unresolved definition "Nope"`}},
		{Name: "satisfy names a part", In: head + "  part box { satisfy a by a; } }",
			Want: refusal{3, 22, `"a" is not a requirement`}},
		{Name: "satisfy by an unknown part", In: head + "  part box { satisfy r by zz; } }",
			Want: refusal{3, 27, `unresolved name "zz"`}},
		{Name: "derivation end is not a requirement", In: head + "  #derivation connection { end #original ::> r; end #derive ::> a; } }",
			Want: refusal{3, 65, `"a" is not a requirement`}},
		{Name: "verify names a part", In: head + "  verification def T { objective { verify a; } } verification <'v'> v : T { } }",
			Want: refusal{3, 43, `"a" is not a requirement`}},
	}, refuse)
}
