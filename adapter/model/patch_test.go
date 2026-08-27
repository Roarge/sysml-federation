package model_test

import (
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

func TestSR22_SetAttributePatchesTextAndProjectionTogether(t *testing.T) {
	m := loadExample(t)
	patched, err := m.SetAttribute("PIPE-S2", "throughput", 1700)
	m2 := assert.Must(t, patched, err)
	assert.Equal(t, m2.Version, 2)
	got, err := part(m2, "PIPE-S2")
	p := assert.Must(t, got, err)
	th := p.Attributes[1]
	assert.Equal(t, *th.Value, 1700.0)
	assert.Equal(t, m2.Text[th.Span.Start:th.Span.End], "1700")
	assert.True(t, strings.Contains(m2.Text, "attribute :>> throughput = 1700;"), "the served text carries the new literal")
	// The previous model is untouched.
	before, err := part(m, "PIPE-S2")
	old := assert.Must(t, before, err)
	assert.Equal(t, *old.Attributes[1].Value, 1200.0)
	assert.Equal(t, m.Version, 1)
	// Every other literal keeps its value, and the text differs only there.
	assert.Equal(t, len(m2.Text), len(m.Text))
	assert.Equal(t, strings.Replace(m.Text, "= 1200;", "= 1700;", 1), m2.Text)
}

func part(m *model.Model, id string) (*model.Part, error) {
	p, ok := m.Part(id)
	if !ok {
		return nil, model.ErrNotFound
	}
	return p, nil
}

// TestSR22_ConcurrentReadsDuringPatches: readers of one model never see a
// mixture, because a patch produces a new model and leaves the old one alone.
// The goroutines count disagreements rather than calling into testing.T,
// which may not fail a test from another goroutine.
func TestSR22_ConcurrentReadsDuringPatches(t *testing.T) {
	m := loadExample(t)
	var wg sync.WaitGroup
	var wrong atomic.Int32
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				p, ok := m.Part("PIPE-S1")
				if !ok {
					wrong.Add(1)
					continue
				}
				th := p.Attributes[1]
				if *th.Value != 2000 || m.Text[th.Span.Start:th.Span.End] != "2000" {
					wrong.Add(1)
				}
			}
		}()
	}
	cur := m
	for j := range 20 {
		patched, err := cur.SetAttribute("PIPE-S1", "throughput", float64(3000+j))
		cur = assert.Must(t, patched, err)
	}
	wg.Wait()
	assert.Equal(t, wrong.Load(), int32(0))
	assert.Equal(t, cur.Version, 21)
}

func TestSR22_ASharedLiteralIsNotEditable(t *testing.T) {
	attrs := parse(t, "package P { part def D { attribute x = 5; }\n"+
		"  part <'u1'> u1 : D;\n"+
		"  part <'u2'> u2 : D; }")
	for _, id := range []string{"u1", "u2"} {
		found, ok := attrs.Part(id)
		assert.True(t, ok, "part "+id)
		assert.Equal(t, found.Attributes[0].Editable, false)
	}
	_, err := attrs.SetAttribute("u1", "x", 9)
	assert.ErrorIs(t, err, model.ErrNotEditable)

	alone := parse(t, "package P { part def D { attribute x = 5; } part <'u1'> u1 : D; }")
	one, ok := alone.Part("u1")
	assert.True(t, ok, "part u1")
	assert.Equal(t, one.Attributes[0].Editable, true)
	edited, err := alone.SetAttribute("u1", "x", 9)
	m2 := assert.Must(t, edited, err)
	after, ok := m2.Part("u1")
	assert.True(t, ok, "part u1 after the edit")
	assert.Equal(t, *after.Attributes[0].Value, 9)

	limits := parse(t, "package P { part def D { attribute q : Real; }\n"+
		"  part <'p'> p : D { attribute :>> q = 1; }\n"+
		"  requirement def R { subject s : D; require constraint { s.q <= 30[s] } }\n"+
		"  requirement <'r1'> r1 : R { subject :>> s = p; }\n"+
		"  requirement <'r2'> r2 : R { subject :>> s = p; } }")
	for _, id := range []string{"r1", "r2"} {
		found, ok := limits.Requirement(id)
		assert.True(t, ok, "requirement "+id)
		assert.Equal(t, found.LimitEditable, false)
	}
	_, err = limits.SetLimit("r1", 45)
	assert.ErrorIs(t, err, model.ErrNotEditable)
}

func TestSR23_DerivedLimitsFollowTheGlobalLimit(t *testing.T) {
	m := loadExample(t)
	patched, err := m.SetLimit("PIPE-R1", 2000)
	m2 := assert.Must(t, patched, err)
	for id, limit := range map[string]float64{"PIPE-R1": 2000, "PIPE-R1.1": 2000, "PIPE-R1.2": 2000,
		"PIPE-R1.3": 1000, "PIPE-R1.4": 1000, "PIPE-R1.5": 2000} {
		found, err := req(m2, id)
		r := assert.Must(t, found, err)
		assert.Equal(t, r.Limit, limit)
	}
	assert.True(t, strings.Contains(m2.Text, "requiredRate = 2000;"), "the global literal is patched")
	// The latency limit, a literal with a unit, patches too and keeps its unit.
	relimited, err := m2.SetLimit("PIPE-R2", 250)
	m3 := assert.Must(t, relimited, err)
	found, err := req(m3, "PIPE-R2")
	r2 := assert.Must(t, found, err)
	assert.Equal(t, r2.Limit, 250.0)
	assert.Equal(t, r2.LimitUnit, "ms")
	assert.True(t, strings.Contains(m3.Text, "maxLatency = 250[ms];"), "the unit survives")
}

func TestSR24_ExpressionBoundValuesAreRefused(t *testing.T) {
	m := loadExample(t)
	_, err := m.SetLimit("PIPE-R1.3", 900)
	assert.ErrorIs(t, err, model.ErrNotEditable)
	_, err = m.SetAttribute("PIPE-P1", "capacity", 1)
	assert.ErrorIs(t, err, model.ErrNotEditable) // unbound: no literal to patch
	src := "package P { part def D { attribute x : Real; } part <'u'> u : D { attribute :>> x = 1 + 1; } }"
	n := parse(t, src)
	_, err = n.SetAttribute("u", "x", 5)
	assert.ErrorIs(t, err, model.ErrNotEditable)
	assert.Equal(t, n.Text, src)
}

func TestSR25_InvalidValuesAreRefused(t *testing.T) {
	m := loadExample(t)
	tabletest.RunErr(t, []tabletest.ErrCase[float64, int]{
		{Name: "negative", In: -1, ErrIs: model.ErrInvalidValue},
		{Name: "positive infinity", In: math.Inf(1), ErrIs: model.ErrInvalidValue},
		{Name: "negative infinity", In: math.Inf(-1), ErrIs: model.ErrInvalidValue},
		{Name: "not a number", In: math.NaN(), ErrIs: model.ErrInvalidValue},
		{Name: "zero is allowed", In: 0, Want: 2},
		{Name: "negative zero is written as zero", In: math.Copysign(0, -1), Want: 2},
		{Name: "fraction", In: 0.5, Want: 2},
	}, func(t *testing.T, in float64) (int, error) {
		m2, err := m.SetAttribute("PIPE-S3", "throughput", in)
		if err != nil {
			return 0, err
		}
		got, err := part(m2, "PIPE-S3")
		p := assert.Must(t, got, err)
		assert.True(t, !strings.Contains(m2.Text, "= -0;"), "no negative zero in the text")
		assert.Equal(t, *p.Attributes[1].Value, in+0) // +0 clears a negative zero
		return m2.Version, nil
	})
	_, err := m.SetAttribute("nope", "throughput", 1)
	assert.ErrorIs(t, err, model.ErrNotFound)
	_, err = m.SetAttribute("PIPE-S3", "nope", 1)
	assert.ErrorIs(t, err, model.ErrNotFound)
	_, err = m.SetLimit("nope", 1)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestPatchGuardsItsInputs(t *testing.T) {
	m := loadExample(t)
	_, err := m.Patch(syntax.Span{Start: 0, End: 7}, "1") // "package" is not a literal
	assert.ErrorIs(t, err, model.ErrNotEditable)
	found, err := part(m, "PIPE-S1")
	p := assert.Must(t, found, err)
	span := p.Attributes[1].Span
	_, err = m.Patch(span, "12 + 3")
	assert.ErrorIs(t, err, model.ErrInvalidValue)
	_, err = m.Patch(span, "abc")
	assert.ErrorIs(t, err, model.ErrInvalidValue)
	patched, err := m.Patch(span, "-5")
	m2 := assert.Must(t, patched, err)
	negative, err := part(m2, "PIPE-S1")
	np := assert.Must(t, negative, err)
	assert.Equal(t, *np.Attributes[1].Value, -5.0)
	assert.Equal(t, m2.Version, 2)
}
