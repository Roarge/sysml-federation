package model_test

import (
	"testing"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

func TestSR19_ExampleConstraints(t *testing.T) {
	m := loadExample(t)
	for _, id := range []string{"PIPE-R1", "PIPE-R1.1", "PIPE-R1.2", "PIPE-R1.3", "PIPE-R1.4", "PIPE-R1.5"} {
		got, err := req(m, id)
		r := assert.Must(t, got, err)
		assert.Equal(t, r.Quantity, "capacity")
		assert.Equal(t, r.Comparison, syntax.GE)
		assert.Equal(t, r.LimitUnit, "")
	}
	got, err := req(m, "PIPE-R2")
	r2 := assert.Must(t, got, err)
	assert.Equal(t, r2.Quantity, "latency")
	assert.Equal(t, r2.Comparison, syntax.LE)
	assert.Equal(t, r2.Limit, 200.0)
	assert.Equal(t, r2.LimitUnit, "ms")
	assert.Equal(t, r2.LimitEditable, true)
}

func TestSR23_DerivedLimitsOfTheExample(t *testing.T) {
	m := loadExample(t)
	for id, limit := range map[string]float64{"PIPE-R1": 1500, "PIPE-R1.1": 1500, "PIPE-R1.2": 1500,
		"PIPE-R1.3": 750, "PIPE-R1.4": 750, "PIPE-R1.5": 1500} {
		got, err := req(m, id)
		r := assert.Must(t, got, err)
		assert.Equal(t, r.Limit, limit)
		assert.Equal(t, r.LimitEditable, id == "PIPE-R1")
	}
}

func TestSR19_ConstraintShapes(t *testing.T) {
	const head = "package P { part def D { attribute q : Real; }\n" +
		"  part <'p'> p : D { part c : D; }\n"
	type shape struct {
		Quantity string
		Op       syntax.Comparison
		Limit    float64
		Unit     string
		Editable bool
	}
	tabletest.Run(t, []tabletest.Case[string, shape]{
		{Name: "subject on the left", In: "subject s : D; attribute l : Real; require constraint { s.q >= l } } requirement <'r'> r : R { subject :>> s = p; attribute :>> l = 5; }",
			Want: shape{"q", syntax.GE, 5, "", true}},
		{Name: "subject on the right flips", In: "subject s : D; attribute l : Real; require constraint { l <= s.q } } requirement <'r'> r : R { subject :>> s = p; attribute :>> l = 5; }",
			Want: shape{"q", syntax.GE, 5, "", true}},
		{Name: "strict flips", In: "subject s : D; attribute l : Real; require constraint { l > s.q } } requirement <'r'> r : R { subject :>> s = p; attribute :>> l = 5; }",
			Want: shape{"q", syntax.LT, 5, "", true}},
		{Name: "equality stays", In: "subject s : D; attribute l : Real; require constraint { 7 == s.q } } requirement <'r'> r : R { subject :>> s = p; }",
			Want: shape{"q", syntax.EQ, 7, "", true}},
		{Name: "literal with unit in the definition", In: "subject s : D; require constraint { s.q <= 30[s] } } requirement <'r'> r : R { subject :>> s = p; }",
			Want: shape{"q", syntax.LE, 30, "s", true}},
		{Name: "expression-bound limit is not editable", In: "subject s : D; attribute l : Real; require constraint { s.q >= l } } requirement <'r'> r : R { subject :>> s = p; attribute :>> l = 2 * 3; }",
			Want: shape{"q", syntax.GE, 6, "", false}},
		{Name: "constraint on the usage", In: "subject s : D; } requirement <'r'> r : R { subject :>> s = p; require constraint { s.q < 1 } }",
			Want: shape{"q", syntax.LT, 1, "", true}},
		{Name: "quantity through a child part", In: "subject s : D; attribute l : Real; require constraint { s.c.q >= l } } requirement <'r'> r : R { subject :>> s = p; attribute :>> l = 1; }",
			Want: shape{"q", syntax.GE, 1, "", true}},
	}, func(t *testing.T, in string) shape {
		m := parse(t, head+"  requirement def R { "+in+" }")
		r := m.Requirements[0]
		return shape{r.Quantity, r.Comparison, r.Limit, r.LimitUnit, r.LimitEditable}
	})
}

func TestSR19_AWalkedChainProjectsThePartThatOwnsTheAttribute(t *testing.T) {
	m := parse(t, "package P { part def D { attribute q : Real; }\n"+
		"  part <'p'> p : D { part <'c'> c : D; }\n"+
		"  requirement def R { subject s : D; attribute l : Real; require constraint { s.c.q >= l } }\n"+
		"  requirement <'r'> r : R { subject :>> s = p; attribute :>> l = 1; } }")
	got, err := req(m, "r")
	r := assert.Must(t, got, err)
	assert.Equal(t, r.Quantity, "q")
	assert.Equal(t, r.Subject, "c")
}

func TestSR19_ARequirementThatBindsNoSubjectIsRefused(t *testing.T) {
	_, err := model.Parse("t.sysml", "package P { part def D { attribute q : Real; }\n"+
		"  part <'p'> p : D;\n"+
		"  requirement def R { subject s : D; attribute l : Real; require constraint { s.q >= l } }\n"+
		"  requirement <'r'> r : R { attribute :>> l = 1; } }")
	se := assert.ErrorAs[*syntax.Error](t, err)
	assert.Equal(t, se.Message, `requirement "r" binds no subject`)
}

func TestSR19_OtherShapesAreRefused(t *testing.T) {
	const head = "package P { part def D { attribute q : Real; }\n  part <'p'> p : D;\n"
	tabletest.Run(t, []tabletest.Case[string, refusal]{
		{Name: "two subject chains", In: head + "  requirement def R { subject s : D; require constraint { s.q >= s.q } }\n  requirement <'r'> r : R { subject :>> s = p; } }",
			Want: refusal{3, 38, "the constraint compares two operands rooted at the subject"}},
		{Name: "no subject chain", In: head + "  requirement def R { subject s : D; attribute l : Real; require constraint { l >= 1 } }\n  requirement <'r'> r : R { subject :>> s = p; attribute :>> l = 1; } }",
			Want: refusal{3, 58, `no operand is a feature chain rooted at the subject "s"`}},
		{Name: "other operand is a chain into a part", In: head + "  requirement def R { subject s : D; require constraint { s.q >= p.q } }\n  requirement <'r'> r : R { subject :>> s = p; } }",
			Want: refusal{3, 66, `"p.q" is neither an attribute of the requirement nor a literal`}},
		{Name: "no constraint", In: head + "  requirement def R { subject s : D; }\n  requirement <'r'> r : R { subject :>> s = p; } }",
			Want: refusal{4, 3, `requirement "r" has no require constraint`}},
		{Name: "no subject", In: head + "  requirement def R { require constraint { s.q >= 1 } }\n  requirement <'r'> r : R { } }",
			Want: refusal{4, 3, `requirement "r" has no subject`}},
		{Name: "limit attribute unbound", In: head + "  requirement def R { subject s : D; attribute l : Real; require constraint { s.q >= l } }\n  requirement <'r'> r : R { subject :>> s = p; } }",
			Want: refusal{4, 3, `limit attribute "l" of requirement "r" has no value`}},
		{Name: "subject does not declare the quantity", In: head + "  requirement def R { subject s : D; require constraint { s.zz >= 1 } }\n  requirement <'r'> r : R { subject :>> s = p; } }",
			Want: refusal{3, 59, `subject "p" declares no attribute "zz"`}},
	}, refuse)
}
