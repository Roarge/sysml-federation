package model_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the second fixture, which shares no name with the example$`, func() { w.path = warehousePath })
	sc.Given(`^the example model$`, func() { w.path = examplePath })
	sc.Given(`^a package Plant whose part line owns a part a with the short name A and a part b with none$`, w.thePlantPackage)
	sc.Given(`^a fixture for each constraint shape the adapter does not read$`, func() {})
	sc.Given(`^parts a and b, each with an in port i, an out port o and an inout port b$`, func() {})
	sc.Given(`^a part u whose attribute b is 100, beside a part other whose attribute a is 10 ms$`, func() {})
	sc.When(`^it is projected$`, w.itIsProjected)
	sc.When(`^(PIPE-\S+) is projected$`, w.aRequirementIsProjected)
	sc.When(`^each is read$`, w.eachShapeIsRead)
	sc.When(`^a connection is declared from (\S+) to (\S+)$`, w.aConnectionIsDeclared)
	sc.When(`^(PIPE-\S+)'s limit is set to (\d+)$`, w.aLimitIsSet)
	sc.When(`^u's attribute x is bound to (.+)$`, w.xIsBoundTo)
	sc.Then(`^its parts are:$`, w.itsPartsAre)
	sc.Then(`^its connections are:$`, w.itsConnectionsAre)
	sc.Then(`^its requirements are:$`, w.itsRequirementsAre)
	sc.Then(`^(\S+) reads "([^"]*)"$`, w.aRequirementReads)
	sc.Then(`^its verification case (\S+), named (\w+), verifies (\S+)$`, w.itsVerificationCase)
	sc.Then(`^its quantity is (\w+), its comparison is (\w+) and its limit is (.+)$`, w.itsConstraintIs)
	sc.Then(`^each is refused at the line and column its fixture gives, as unsupported syntax is$`, w.eachShapeIsRefused)
	sc.Then(`^the adapter refuses it: (.+)$`, w.theAdapterRefusesIt)
	sc.Then(`^its parts are identified as (.+)$`, w.itsPartsAreIdentifiedAs)
	sc.Then(`^its requirements as (.+)$`, w.itsRequirementsAreIdentifiedAs)
	sc.Then(`^its verification case as (\S+)$`, w.itsVerificationCaseIsIdentifiedAs)
	sc.Then(`^(\w+) is identified as (\S+)$`, w.aPartIsIdentifiedAs)
	sc.Then(`^the limits are:$`, w.theLimitsAre)
	sc.Then(`^x reads (.+)$`, w.xReads)
}

// world is what one scenario of this package knows.
type world struct {
	t           *testing.T
	path        string
	src         string
	m           *model.Model
	requirement string
	refusals    []refusal
	x           model.Attribute
}

func (w *world) thePlantPackage() {
	w.src = "package Plant { part def S; part line : S { part <'A'> a : S; part b : S; } }"
}

func (w *world) itIsProjected() {
	if w.src != "" {
		w.m = parse(w.t, w.src)
		return
	}
	loaded, err := model.Load(w.path)
	w.m = assert.Must(w.t, loaded, err)
}

func (w *world) aRequirementIsProjected(id string) {
	w.itIsProjected()
	w.requirement = id
}

// number writes a value the way the scenarios do: as few digits as it takes,
// and the unit after a space.
func number(v float64, unit string) string {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	if unit != "" {
		s += " " + unit
	}
	return s
}

// cells is a table's rows below its header, each as its cells' text.
func cells(table *godog.Table) [][]string {
	var rows [][]string
	for _, row := range table.Rows[1:] {
		var values []string
		for _, cell := range row.Cells {
			values = append(values, cell.Value)
		}
		rows = append(rows, values)
	}
	return rows
}

func joined(items []string) string { return strings.Join(items, ", ") }

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func (w *world) itsPartsAre(table *godog.Table) {
	within := make(map[string]string)
	for _, p := range w.m.Parts {
		for _, child := range p.Parts {
			within[child.ID] = p.ID
		}
	}
	var got [][]string
	for _, p := range w.m.Parts {
		var attributes, ports []string
		for _, a := range p.Attributes {
			if a.Value == nil {
				attributes = append(attributes, a.Name)
				continue
			}
			attributes = append(attributes, a.Name+" = "+number(*a.Value, a.Unit))
		}
		for _, port := range p.Ports {
			ports = append(ports, port.Name+" "+strings.ToLower(port.Direction.String()))
		}
		got = append(got, []string{p.ID, p.Name, p.Definition, within[p.ID], joined(attributes), joined(ports)})
	}
	assert.DeepEqual(w.t, got, cells(table))
}

func (w *world) itsConnectionsAre(table *godog.Table) {
	var got [][]string
	for _, p := range w.m.Parts {
		for _, c := range p.Connections {
			got = append(got, []string{c.From, c.FromPort, c.To, c.ToPort})
		}
	}
	assert.DeepEqual(w.t, got, cells(table))
}

func (w *world) itsRequirementsAre(table *godog.Table) {
	var got [][]string
	for _, r := range w.m.Requirements {
		got = append(got, []string{r.ID, r.Subject, r.Quantity, r.Comparison.String(), number(r.Limit, r.LimitUnit),
			yesNo(r.LimitEditable), joined(r.DerivedFrom), joined(r.SatisfiedBy), joined(r.VerifiedBy)})
	}
	assert.DeepEqual(w.t, got, cells(table))
}

func (w *world) aRequirementReads(id, text string) {
	found, err := req(w.m, id)
	assert.Equal(w.t, assert.Must(w.t, found, err).Text, text)
}

func (w *world) itsVerificationCase(id, name, verifies string) {
	assert.Len(w.t, w.m.VerificationCases, 1)
	vc := w.m.VerificationCases[0]
	assert.Equal(w.t, vc.ID, id)
	assert.Equal(w.t, vc.Name, name)
	assert.SliceEqual(w.t, vc.Verifies, []string{verifies})
}

func (w *world) itsConstraintIs(quantity, comparison, limit string) {
	found, err := req(w.m, w.requirement)
	r := assert.Must(w.t, found, err)
	assert.Equal(w.t, r.Quantity, quantity)
	assert.Equal(w.t, r.Comparison.String(), comparison)
	assert.Equal(w.t, number(r.Limit, r.LimitUnit), limit)
}

func (w *world) eachShapeIsRead() {
	for _, shape := range otherShapes {
		w.refusals = append(w.refusals, refuse(w.t, shape.In))
	}
}

func (w *world) eachShapeIsRefused() {
	assert.Len(w.t, w.refusals, len(otherShapes))
	for i, shape := range otherShapes {
		assert.Equal(w.t, w.refusals[i], shape.Want)
	}
}

// portsHead declares ports in, out and inout, and a part definition carrying
// one of each.
const portsHead = "package P { port def I { in item q; } port def O { out item q; } port def B { inout item q; }\n" +
	"  part def S { port i : I; port o : O; port b : B; }\n"

func (w *world) aConnectionIsDeclared(first, second string) {
	src := portsHead + "  part box { part a : S; part b : S; connect " + first + " to " + second + "; } }"
	w.refusals = []refusal{refuse(w.t, src)}
}

func (w *world) theAdapterRefusesIt(message string) {
	assert.Len(w.t, w.refusals, 1)
	assert.Equal(w.t, w.refusals[0].Message, message)
}

// listed reads a list the way the scenarios write it: a, b and c.
var listedRE = regexp.MustCompile(`, | and `)

func listed(s string) []string { return listedRE.Split(s, -1) }

func (w *world) itsPartsAreIdentifiedAs(ids string) {
	var got []string
	for _, p := range w.m.Parts {
		got = append(got, p.ID)
	}
	assert.SliceEqual(w.t, got, listed(ids))
}

func (w *world) itsRequirementsAreIdentifiedAs(ids string) {
	var got []string
	for _, r := range w.m.Requirements {
		got = append(got, r.ID)
	}
	assert.SliceEqual(w.t, got, listed(ids))
}

func (w *world) itsVerificationCaseIsIdentifiedAs(id string) {
	assert.Len(w.t, w.m.VerificationCases, 1)
	assert.Equal(w.t, w.m.VerificationCases[0].ID, id)
}

func (w *world) aPartIsIdentifiedAs(name, id string) {
	for _, p := range w.m.Parts {
		if p.Name == name {
			assert.Equal(w.t, p.ID, id)
			return
		}
	}
	w.t.Errorf("no part is named %s", name)
}

func (w *world) aLimitIsSet(id string, limit int) {
	loaded, err := model.Load(examplePath)
	patched, err := assert.Must(w.t, loaded, err).SetLimit(id, float64(limit))
	w.m = assert.Must(w.t, patched, err)
}

func (w *world) theLimitsAre(table *godog.Table) {
	for _, row := range cells(table) {
		found, err := req(w.m, row[0])
		assert.Equal(w.t, number(assert.Must(w.t, found, err).Limit, ""), row[1])
	}
}

// operatorsHead declares the part other, whose attribute a is 10 ms, for the
// expressions below to reach.
const operatorsHead = "package P { part def D { attribute a : Real; attribute b : Real; attribute x : Real; }\n" +
	"  part <'other'> other : D { attribute :>> a = 10[ms]; }\n"

func (w *world) xIsBoundTo(expression string) {
	w.x = value(w.t, operatorsHead+"  part <'u'> u : D { attribute :>> b = 100; attribute :>> x = "+expression+"; } }", "x")
}

func (w *world) xReads(want string) {
	if w.x.Value == nil {
		w.t.Fatal("x has no value")
	}
	assert.Equal(w.t, number(*w.x.Value, w.x.Unit), want)
	assert.Equal(w.t, w.x.Editable, false)
}
