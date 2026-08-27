package tree

import (
	"testing"

	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

const shipped = `{"nodes":[
 {"id":"intro","kind":"PROSE","text":"An idealised rollup."},
 {"id":"PIPE-R1","kind":"REQUIREMENT","requirement":"PIPE-R1","children":[
  {"id":"PIPE-R1.1","kind":"REQUIREMENT","requirement":"PIPE-R1.1"},
  {"id":"PIPE-R1.2","kind":"REQUIREMENT","requirement":"PIPE-R1.2"},
  {"id":"PIPE-R1.3","kind":"REQUIREMENT","requirement":"PIPE-R1.3"},
  {"id":"PIPE-R1.4","kind":"REQUIREMENT","requirement":"PIPE-R1.4"},
  {"id":"PIPE-R1.5","kind":"REQUIREMENT","requirement":"PIPE-R1.5"}]},
 {"id":"PIPE-R2","kind":"REQUIREMENT","requirement":"PIPE-R2"}]}`

func load(t *testing.T) *Tree {
	t.Helper()
	tr, err := Load([]byte(shipped))
	return assert.Must(t, tr, err)
}

func TestSR33_ShippedNumbering(t *testing.T) {
	assert.MapEqual(t, load(t).Numbers(), map[string]string{
		"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.4": "1.4", "PIPE-R1.5": "1.5", "PIPE-R2": "2",
	})
}

type op func(*Tree) error

func TestSR35_Operations(t *testing.T) {
	cases := []tabletest.Case[op, map[string]string]{
		{Name: "US-07 move PIPE-R1.5 above PIPE-R1.1", In: func(tr *Tree) error { return tr.Move("PIPE-R1.5", "PIPE-R1", 0) },
			Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.5": "1.1", "PIPE-R1.1": "1.2", "PIPE-R1.2": "1.3", "PIPE-R1.3": "1.4", "PIPE-R1.4": "1.5", "PIPE-R2": "2"}},
		{Name: "US-07 nest PIPE-R2 under PIPE-R1", In: func(tr *Tree) error { return tr.Move("PIPE-R2", "PIPE-R1", 5) },
			Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.4": "1.4", "PIPE-R1.5": "1.5", "PIPE-R2": "1.6"}},
		{Name: "an index past the end appends", In: func(tr *Tree) error { return tr.Move("PIPE-R2", "PIPE-R1", 99) },
			Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.4": "1.4", "PIPE-R1.5": "1.5", "PIPE-R2": "1.6"}},
		{Name: "move to the root", In: func(tr *Tree) error { return tr.Move("PIPE-R1.3", "", 0) },
			Want: map[string]string{"PIPE-R1.3": "1", "PIPE-R1": "2", "PIPE-R1.1": "2.1", "PIPE-R1.2": "2.2", "PIPE-R1.4": "2.3", "PIPE-R1.5": "2.4", "PIPE-R2": "3"}},
		{Name: "US-08 heading Performance as parent of PIPE-R1", In: func(tr *Tree) error { _, err := tr.InsertHeading("PIPE-R1", "Performance"); return err },
			Want: map[string]string{"n1": "1", "PIPE-R1": "1.1", "PIPE-R1.1": "1.1.1", "PIPE-R1.2": "1.1.2", "PIPE-R1.3": "1.1.3", "PIPE-R1.4": "1.1.4", "PIPE-R1.5": "1.1.5", "PIPE-R2": "2"}},
		{Name: "US-08 prose under the heading carries no number", In: func(tr *Tree) error {
			h, err := tr.InsertHeading("PIPE-R1", "Performance")
			if err != nil {
				return err
			}
			_, err = tr.AddProse(h, 0, "Why allocated limits fail.")
			return err
		}, Want: map[string]string{"n1": "1", "PIPE-R1": "1.1", "PIPE-R1.1": "1.1.1", "PIPE-R1.2": "1.1.2", "PIPE-R1.3": "1.1.3", "PIPE-R1.4": "1.1.4", "PIPE-R1.5": "1.1.5", "PIPE-R2": "2"}},
		{Name: "US-08 exclude PIPE-R1.4", In: func(tr *Tree) error { return tr.Exclude("PIPE-R1.4") },
			Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.5": "1.4", "PIPE-R2": "2"}},
		{Name: "US-08 restore returns as the last child", In: func(tr *Tree) error {
			if err := tr.Exclude("PIPE-R1.4"); err != nil {
				return err
			}
			return tr.Include("PIPE-R1.4")
		}, Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.5": "1.4", "PIPE-R1.4": "1.5", "PIPE-R2": "2"}},
		{Name: "excluding a parent promotes its children", In: func(tr *Tree) error { return tr.Exclude("PIPE-R1") },
			Want: map[string]string{"PIPE-R1.1": "1", "PIPE-R1.2": "2", "PIPE-R1.3": "3", "PIPE-R1.4": "4", "PIPE-R1.5": "5", "PIPE-R2": "6"}},
		{Name: "restoring a former parent that is gone lands at the root", In: func(tr *Tree) error {
			if err := tr.Exclude("PIPE-R1"); err != nil {
				return err
			}
			if err := tr.Exclude("PIPE-R1.4"); err != nil {
				return err
			}
			return tr.Include("PIPE-R1.4")
		}, Want: map[string]string{"PIPE-R1.1": "1", "PIPE-R1.2": "2", "PIPE-R1.3": "3", "PIPE-R1.5": "4", "PIPE-R2": "5", "PIPE-R1.4": "6"}},
	}
	tabletest.Run(t, cases, func(t *testing.T, apply op) map[string]string {
		tr := load(t)
		assert.NoError(t, apply(tr))
		return tr.Numbers()
	})
}

func TestEditText(t *testing.T) {
	tr := load(t)
	assert.NoError(t, tr.EditText("intro", "Changed."))
	assert.Equal(t, tr.Find("intro").Text, "Changed.")
	assert.ErrorIs(t, tr.EditText("PIPE-R1", "no"), ErrNotText)
	assert.ErrorIs(t, tr.EditText("nope", "no"), ErrUnknown)
}

func TestRefusals(t *testing.T) {
	tr := load(t)
	assert.ErrorIs(t, tr.Move("PIPE-R1", "PIPE-R1.2", 0), ErrCycle)
	assert.ErrorIs(t, tr.Move("PIPE-R1", "PIPE-R1", 0), ErrCycle)
	assert.ErrorIs(t, tr.Move("PIPE-R2", "intro", 0), ErrProseParent)
	assert.ErrorIs(t, tr.Move("nope", "", 0), ErrUnknown)
	assert.ErrorIs(t, tr.Include("PIPE-R2"), ErrNotExcluded)
	assert.ErrorIs(t, tr.Exclude("PIPE-R9"), ErrUnknown)
	assert.NoError(t, tr.Exclude("PIPE-R2"))
	assert.ErrorIs(t, tr.Exclude("PIPE-R2"), ErrUnknown)
	_, err := tr.AddProse("intro", 0, "x")
	assert.ErrorIs(t, err, ErrProseParent)
}

func TestRequirementLookup(t *testing.T) {
	tr := load(t)
	number, included := tr.Requirement("PIPE-R1.3")
	assert.Equal(t, number, "1.3")
	assert.True(t, included, "PIPE-R1.3 is in the shipped tree")
	_, included = tr.Requirement("PIPE-R9")
	assert.True(t, !included, "an unknown id is not included")
	assert.NoError(t, tr.Exclude("PIPE-R1.3"))
	_, included = tr.Requirement("PIPE-R1.3")
	assert.True(t, !included, "an excluded requirement is not included")
}

func TestLoadRefusesDuplicateIDs(t *testing.T) {
	_, err := Load([]byte(`{"nodes":[{"id":"a","kind":"PROSE"},{"id":"a","kind":"PROSE"}]}`))
	assert.Error(t, err)
	_, err = Load([]byte(`{"nodes":[{"id":"a","kind":"REQUIREMENT"}]}`))
	assert.Error(t, err)
}
