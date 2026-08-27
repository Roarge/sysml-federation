package model_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

// warehousePath is the second fixture (SR-16): other names, a wiring with
// fan-in, a constraint written subject-last, a requirement without a short
// name, a literal limit inside a definition and an unbound attribute.
const warehousePath = "testdata/warehouse.sysml"

func loadWarehouse(t *testing.T) *model.Model {
	t.Helper()
	loaded, err := model.Load(warehousePath)
	return assert.Must(t, loaded, err)
}

// literalSpan returns the span of literal inside the first occurrence of stmt
// in the model text, so an expected Attribute can state its Span without a
// hand-counted byte offset.
func literalSpan(t *testing.T, m *model.Model, stmt, literal string) syntax.Span {
	t.Helper()
	i := strings.Index(m.Text, stmt)
	if i < 0 {
		t.Fatalf("%q is not in the model text", stmt)
	}
	start := i + strings.Index(stmt, literal)
	return syntax.Span{Start: start, End: start + len(literal)}
}

func TestSR16_WarehouseTextVersionAndIdentifiers(t *testing.T) {
	m := loadWarehouse(t)
	read, err := os.ReadFile(warehousePath)
	src := string(assert.Must(t, read, err))
	assert.Equal(t, m.Text, src)
	assert.Equal(t, m.Version, 1)

	var ids []string
	for _, p := range m.Parts {
		ids = append(ids, p.ID)
	}
	assert.SliceEqual(t, ids, []string{"WH-L1", "WH-A", "WH-B", "WH-C", "WH-D"})

	// packRate declares no short name, so its id is the qualified name (AD-0018).
	var rids []string
	for _, r := range m.Requirements {
		rids = append(rids, r.ID)
	}
	assert.SliceEqual(t, rids, []string{"WH-R1", "WH-R1.1", "Warehouse::packRate", "WH-R2"})
	found, err := req(m, "Warehouse::packRate")
	pack := assert.Must(t, found, err)
	assert.Equal(t, pack.ShortName, "")
	assert.Equal(t, pack.Name, "packRate")

	assert.Len(t, m.VerificationCases, 1)
	assert.Equal(t, m.VerificationCases[0].ID, "WH-VC1")
	assert.Equal(t, m.VerificationCases[0].ShortName, "WH-VC1")
	assert.Equal(t, m.VerificationCases[0].Name, "cycleTest")
}

func TestSR16_WarehousePartsAttributesPortsAndWiring(t *testing.T) {
	m := loadWarehouse(t)
	assert.Len(t, m.Roots, 1)
	line := m.Roots[0]
	assert.Equal(t, line.ID, "WH-L1")
	assert.Equal(t, line.ShortName, "WH-L1")
	assert.Equal(t, line.Name, "line")
	assert.Equal(t, line.Definition, "Line")
	assert.Equal(t, line.Doc, "pick and label are entries, ship is the exit.")
	// Station's capacity first, then Line's cycleTime and shifts. Only shifts
	// is bound, by the usage's `attribute :>> shifts = 2;`.
	assert.DeepEqual(t, line.Attributes, []model.Attribute{
		{Name: "capacity"},
		{Name: "cycleTime"},
		{Name: "shifts", Value: f64(2), Editable: true, Span: literalSpan(t, m, "shifts = 2;", "2")},
	})
	assert.Len(t, line.Ports, 0)

	robots := []struct {
		ID, Name, Stmt, Lit string
		Rate                float64
	}{
		{"WH-A", "pick", "part <'WH-A'> pick : Robot { attribute :>> rate = 40; }", "40", 40},
		{"WH-B", "pack", "part <'WH-B'> pack : Robot { attribute :>> rate = 30; }", "30", 30},
		{"WH-C", "label", "part <'WH-C'> label : Robot { attribute :>> rate = 25; }", "25", 25},
		{"WH-D", "ship", "part <'WH-D'> ship : Robot { attribute :>> rate = 50; }", "50", 50},
	}
	assert.Len(t, line.Parts, len(robots))
	for i, want := range robots {
		robot := line.Parts[i]
		assert.Equal(t, robot.ID, want.ID)
		assert.Equal(t, robot.Name, want.Name)
		assert.Equal(t, robot.Definition, "Robot")
		assert.Equal(t, robot.Doc, "")
		assert.DeepEqual(t, robot.Attributes, []model.Attribute{
			{Name: "capacity"},
			{Name: "rate", Value: f64(want.Rate), Editable: true, Span: literalSpan(t, m, want.Stmt, want.Lit)},
		})
		// Robot's ports in declaration order, directed by their definitions' items.
		assert.DeepEqual(t, robot.Ports, []model.Port{
			{Name: "take", Direction: syntax.DirectionIn},
			{Name: "give", Direction: syntax.DirectionOut},
		})
		assert.Len(t, robot.Parts, 0)
		assert.Len(t, robot.Connections, 0)
	}

	// Fan-in: pack and label both feed ship. Connections belong to the part
	// that declares them, in source order, with <from>.<port>-><to>.<port> ids.
	assert.DeepEqual(t, line.Connections, []model.Connection{
		{ID: "WH-A.give->WH-B.take", From: "WH-A", FromPort: "give", To: "WH-B", ToPort: "take"},
		{ID: "WH-B.give->WH-D.take", From: "WH-B", FromPort: "give", To: "WH-D", ToPort: "take"},
		{ID: "WH-C.give->WH-D.take", From: "WH-C", FromPort: "give", To: "WH-D", ToPort: "take"},
	})
}

func TestSR16_WarehouseRequirementsAndRelationships(t *testing.T) {
	m := loadWarehouse(t)
	type want struct {
		Text          string
		Subject       string
		Quantity      string
		Comparison    syntax.Comparison
		Limit         float64
		LimitUnit     string
		LimitEditable bool
		DerivedFrom   []string
		Derives       []string
		SatisfiedBy   []string
		VerifiedBy    []string
	}
	// Limits, from the fixture text:
	//   lineRate    required = 60                               literal, editable
	//   shipRate    required = (lineRate.required + 20) / 2     = (60 + 20) / 2 = 40
	//   packRate    required = lineRate.required - 10 * 3       = 60 - 30       = 30
	//   cycleLimit  station.cycleTime <= 30[s]                  literal in the definition
	// RateRequirement writes the subject on the right, `required <= station.capacity`,
	// so the projected comparison is the flipped GE on quantity capacity.
	tabletest.Run(t, []tabletest.Case[string, want]{
		{Name: "WH-R1", In: "WH-R1", Want: want{
			Text: "The line shall handle the required parcel rate", Subject: "WH-L1",
			Quantity: "capacity", Comparison: syntax.GE, Limit: 60, LimitEditable: true,
			Derives: []string{"WH-R1.1", "Warehouse::packRate"}, SatisfiedBy: []string{"WH-L1"}}},
		{Name: "WH-R1.1", In: "WH-R1.1", Want: want{
			// No doc on the usage, so the definition's doc is the text.
			Text: "The subject shall handle the required parcel rate.", Subject: "WH-D",
			Quantity: "capacity", Comparison: syntax.GE, Limit: 40,
			DerivedFrom: []string{"WH-R1"}, SatisfiedBy: []string{"WH-D"}}},
		{Name: "Warehouse::packRate", In: "Warehouse::packRate", Want: want{
			Text: "The subject shall handle the required parcel rate.", Subject: "WH-B",
			Quantity: "capacity", Comparison: syntax.GE, Limit: 30,
			DerivedFrom: []string{"WH-R1"}, SatisfiedBy: []string{"WH-B"}}},
		{Name: "WH-R2", In: "WH-R2", Want: want{
			Text: "A cycle shall finish within the limit", Subject: "WH-L1",
			Quantity: "cycleTime", Comparison: syntax.LE, Limit: 30, LimitUnit: "s", LimitEditable: true,
			SatisfiedBy: []string{"WH-L1"}, VerifiedBy: []string{"WH-VC1"}}},
	}, func(t *testing.T, id string) want {
		found, err := req(m, id)
		r := assert.Must(t, found, err)
		return want{Text: r.Text, Subject: r.Subject, Quantity: r.Quantity, Comparison: r.Comparison,
			Limit: r.Limit, LimitUnit: r.LimitUnit, LimitEditable: r.LimitEditable,
			DerivedFrom: r.DerivedFrom, Derives: r.Derives, SatisfiedBy: r.SatisfiedBy, VerifiedBy: r.VerifiedBy}
	})

	vc := m.VerificationCases[0]
	assert.SliceEqual(t, vc.Verifies, []string{"WH-R2"})
	line := m.Roots[0]
	assert.SliceEqual(t, line.Satisfies, []string{"WH-R1", "WH-R2"})
	assert.Len(t, line.Parts[0].Satisfies, 0) // pick
	assert.SliceEqual(t, line.Parts[1].Satisfies, []string{"Warehouse::packRate"})
	assert.Len(t, line.Parts[2].Satisfies, 0) // label
	assert.SliceEqual(t, line.Parts[3].Satisfies, []string{"WH-R1.1"})
}

func TestSR16_WarehouseMutations(t *testing.T) {
	m := loadWarehouse(t)

	// Raising the line's rate re-evaluates both derived limits:
	//   shipRate = (80 + 20) / 2 = 50 and packRate = 80 - 10 * 3 = 50.
	raised, err := m.SetLimit("WH-R1", 80)
	m2 := assert.Must(t, raised, err)
	assert.Equal(t, m2.Version, 2)
	assert.Equal(t, m2.Text, strings.Replace(m.Text, "required = 60;", "required = 80;", 1))
	for id, limit := range map[string]float64{"WH-R1": 80, "WH-R1.1": 50, "Warehouse::packRate": 50} {
		found, err := req(m2, id)
		r := assert.Must(t, found, err)
		assert.Equal(t, r.Limit, limit)
	}
	before, err := req(m, "WH-R1")
	old := assert.Must(t, before, err)
	assert.Equal(t, old.Limit, 60.0) // the old model is untouched

	// A limit that is a literal inside the definition patches there and keeps
	// its unit.
	relimited, err := m2.SetLimit("WH-R2", 45)
	m3 := assert.Must(t, relimited, err)
	assert.Equal(t, m3.Version, 3)
	limited, err := req(m3, "WH-R2")
	r2 := assert.Must(t, limited, err)
	assert.Equal(t, r2.Limit, 45.0)
	assert.Equal(t, r2.LimitUnit, "s")
	assert.Equal(t, r2.LimitEditable, true)
	assert.True(t, strings.Contains(m3.Text, "station.cycleTime <= 45[s]"), "the definition's literal is patched")

	// Derived limits are not editable, whether or not the requirement has a
	// short name, and an unknown id is not found.
	tabletest.RunErr(t, []tabletest.ErrCase[string, int]{
		{Name: "derived limit", In: "WH-R1.1", ErrIs: model.ErrNotEditable},
		{Name: "derived limit without a short name", In: "Warehouse::packRate", ErrIs: model.ErrNotEditable},
		{Name: "unknown requirement", In: "WH-R9", ErrIs: model.ErrNotFound},
		{Name: "literal limit", In: "WH-R1", Want: 2},
	}, func(t *testing.T, id string) (int, error) {
		next, err := m.SetLimit(id, 70)
		if err != nil {
			return 0, err
		}
		return next.Version, nil
	})

	// An unbound attribute has no literal to patch. A bound one patches.
	_, err = m.SetAttribute("WH-L1", "cycleTime", 1)
	assert.ErrorIs(t, err, model.ErrNotEditable)
	shifted, err := m.SetAttribute("WH-L1", "shifts", 3)
	m4 := assert.Must(t, shifted, err)
	assert.True(t, strings.Contains(m4.Text, "shifts = 3;"), "the part attribute is patched")
	patchedLine, err := part(m4, "WH-L1")
	l1 := assert.Must(t, patchedLine, err)
	assert.Equal(t, *l1.Attributes[2].Value, 3.0)
}
