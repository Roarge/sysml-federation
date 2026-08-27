package model_test

import (
	"testing"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

// value returns the projected attribute called name on the last root part,
// which is the part every source below binds the attribute under test on.
func value(t *testing.T, src, name string) model.Attribute {
	t.Helper()
	m := parse(t, src)
	for _, a := range m.Roots[len(m.Roots)-1].Attributes {
		if a.Name == name {
			return a
		}
	}
	t.Fatalf("no attribute %q", name)
	return model.Attribute{}
}

func TestSR23_OperatorsParenthesesChainsAndUnits(t *testing.T) {
	const head = "package P { part def D { attribute a : Real; attribute b : Real; attribute x : Real; }\n" +
		"  part <'other'> other : D { attribute :>> a = 10[ms]; }\n"
	type want struct {
		Value      float64
		Unit       string
		Expression string
	}
	tabletest.Run(t, []tabletest.Case[string, want]{
		{Name: "plus", In: "1 + 2", Want: want{3, "", "1 + 2"}},
		{Name: "minus", In: "5 - 7", Want: want{-2, "", "5 - 7"}},
		{Name: "times", In: "3 * 4", Want: want{12, "", "3 * 4"}},
		{Name: "divide", In: "9 / 2", Want: want{4.5, "", "9 / 2"}},
		{Name: "precedence", In: "1 + 2 * 3", Want: want{7, "", "1 + 2 * 3"}},
		{Name: "parentheses", In: "(1 + 2) * 3", Want: want{9, "", "(1 + 2) * 3"}},
		{Name: "chain to another part", In: "other.a / 2", Want: want{5, "ms", "other.a / 2"}},
		{Name: "own attribute", In: "b + 1", Want: want{101, "", "b + 1"}},
		{Name: "unit from the right", In: "2 * other.a", Want: want{20, "ms", "2 * other.a"}},
		{Name: "unit literal on the left", In: "3[s] + 1", Want: want{4, "s", "3[s] + 1"}},
	}, func(t *testing.T, in string) want {
		a := value(t, head+"  part <'u'> u : D { attribute :>> b = 100; attribute :>> x = "+in+"; } }", "x")
		return want{*a.Value, a.Unit, a.Expression}
	})
}

func TestSR23_ExpressionBoundValuesAreNotEditable(t *testing.T) {
	a := value(t, "package P { part def D { attribute x : Real; } part <'u'> u : D { attribute :>> x = 1 + 1; } }", "x")
	assert.Equal(t, a.Editable, false)
	assert.Equal(t, a.Span.End, 0)
	lit := value(t, "package P { part def D { attribute x : Real; } part <'u'> u : D { attribute :>> x = 2; } }", "x")
	assert.Equal(t, lit.Editable, true)
	assert.Equal(t, lit.Expression, "")
}

func TestSR23_RedefinitionWithoutAValue(t *testing.T) {
	inherited := value(t, "package P { part def D { attribute x = 5; } part u : D { attribute :>> x; } }", "x")
	if inherited.Value == nil {
		t.Fatal("x lost the value its definition binds")
	}
	assert.Equal(t, *inherited.Value, 5)

	unbound := value(t, "package P { part def D { attribute x : Real; } part u : D { attribute :>> x; } }", "x")
	assert.True(t, unbound.Value == nil, "an attribute no declaration binds reports no value")
}

func TestEvalRefusals(t *testing.T) {
	const head = "package P { part def D { attribute a : Real; attribute b : Real; }\n"
	tabletest.Run(t, []tabletest.Case[string, refusal]{
		{Name: "cycle", In: head + "  part <'u'> u : D { attribute :>> a = b + 1; attribute :>> b = a; } }",
			Want: refusal{2, 40, `cyclic binding of "a"`}},
		{Name: "self cycle", In: head + "  part <'u'> u : D { attribute :>> a = a; } }",
			Want: refusal{2, 40, `cyclic binding of "a"`}},
		{Name: "unbound reference", In: head + "  part <'u'> u : D { attribute :>> a = b; } }",
			Want: refusal{2, 40, `"b" has no value`}},
		{Name: "unresolved chain", In: head + "  part <'u'> u : D { attribute :>> a = zz.q; } }",
			Want: refusal{2, 40, `unresolved name "zz"`}},
		{Name: "chain ends on a part", In: head + "  part <'u'> u : D { attribute :>> a = u; } }",
			Want: refusal{2, 40, `"u" is not an attribute`}},
		{Name: "division by zero", In: head + "  part <'u'> u : D { attribute :>> a = 1 / 0; } }",
			Want: refusal{2, 40, "division by zero"}},
		{Name: "units differ", In: head + "  part <'u'> u : D { attribute :>> a = 1[s] + 1[kg]; } }",
			Want: refusal{2, 40, `units differ: "s" and "kg"`}},
	}, refuse)
}
