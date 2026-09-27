package projection

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the adapter serves the second fixture at version 1$`, w.theAdapterServesTheSecondFixture)
	sc.Given(`^four writers setting (WH-\S+)'s limit over and over$`, w.fourWriters)
	sc.Given(`^(WH-\S+)'s limit is bound by an expression$`, w.theLimitIsBound)
	sc.When(`^(WH-\S+)'s (\w+) is set to (\d+)$`, w.aValueIsSet)
	sc.When(`^fifty reads each take the served text and (WH-\S+)'s limit together$`, w.fiftyReads)
	sc.When(`^(.+) is submitted as (WH-\S+)'s (\w+)$`, w.anInputIsSubmitted)
	sc.Step(`^(WH-\S+)'s (\w+) is (\d+)$`, w.aValueIs)
	sc.Then(`^(WH-\S+)'s (\w+) is still (\d+)$`, w.aValueIs)
	sc.Then(`^the served text carries "([^"]*)" and no longer "([^"]*)"$`, w.theServedTextCarries)
	sc.Then(`^the model version is (\d+)$`, w.theVersionIs)
	sc.Then(`^the model version is still (\d+)$`, w.theVersionIs)
	sc.Then(`^every read finds the limit it was given in the text it was given$`, w.everyReadAgrees)
	sc.Then(`^the writers' patches landed while the reads ran$`, w.thePatchesLanded)
	sc.Then(`^the mutation is refused as not editable$`, w.refusedAsNotEditable)
	sc.Then(`^the mutation is rejected$`, w.theMutationIsRejected)
}

// world is what one scenario of this package knows.
type world struct {
	t         *testing.T
	c         *tapped
	store     *Store
	err       error
	stop      chan struct{}
	writers   sync.WaitGroup
	failures  atomic.Int32
	disagreed []string
}

func (w *world) theAdapterServesTheSecondFixture() {
	w.c, w.store = loadClient(w.t)
	assert.Equal(w.t, w.store.Version(), 1)
}

// mutation is the GraphQL that sets one value: a requirement's limit, or an
// attribute of a part.
func mutation(id, name, value string) string {
	if name == "limit" {
		return fmt.Sprintf(`mutation { setLimit(requirementId: %q, value: %s) { id } }`, id, value)
	}
	return fmt.Sprintf(`mutation { setAttribute(partId: %q, name: %q, value: %s) { id } }`, id, name, value)
}

func (w *world) aValueIsSet(id, name, value string) {
	var out mutationOut
	w.err = w.c.Post(mutation(id, name, value), &out)
}

// read returns a part's attribute or a requirement's limit, as the graph
// serves it.
func (w *world) read(id, name string) string {
	var answer struct {
		Part *struct {
			Attributes []struct {
				Name  string
				Value *float64
			}
		}
		Requirement *struct{ Limit float64 }
	}
	if name == "limit" {
		w.c.MustPost(fmt.Sprintf(`{ requirement(id: %q) { limit } }`, id), &answer)
		return strconv.FormatFloat(answer.Requirement.Limit, 'f', -1, 64)
	}
	w.c.MustPost(fmt.Sprintf(`{ part(id: %q) { attributes { name value } } }`, id), &answer)
	for _, a := range answer.Part.Attributes {
		if a.Name == name && a.Value != nil {
			return strconv.FormatFloat(*a.Value, 'f', -1, 64)
		}
	}
	return "no value"
}

func (w *world) aValueIs(id, name, want string) { assert.Equal(w.t, w.read(id, name), want) }

func (w *world) theServedTextCarries(now, before string) {
	var answer struct{ Model struct{ Text string } }
	w.c.MustPost(`{ model { text } }`, &answer)
	assert.True(w.t, strings.Contains(answer.Model.Text, now), "the served text carries "+now)
	assert.True(w.t, !strings.Contains(answer.Model.Text, before), "the served text no longer carries "+before)
}

func (w *world) theVersionIs(version int) { assert.Equal(w.t, w.store.Version(), version) }

// fourWriters patch the limit in a loop until the reads are done. They count
// their own failures rather than calling into testing.T from another
// goroutine.
func (w *world) fourWriters(id string) {
	w.stop = make(chan struct{})
	for i := range 4 {
		w.writers.Add(1)
		go func() {
			defer w.writers.Done()
			for j := 0; ; j++ {
				select {
				case <-w.stop:
					return
				default:
				}
				if _, _, err := w.store.SetLimit(id, float64(100+i*1000+j)); err != nil {
					w.failures.Add(1)
					return
				}
			}
		}()
	}
	w.t.Cleanup(w.stopWriters)
}

func (w *world) stopWriters() {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
	w.writers.Wait()
}

func (w *world) fiftyReads(id string) {
	for range 50 {
		var resp struct {
			Model       struct{ Text string }
			Requirement struct{ Limit float64 }
		}
		w.c.MustPost(fmt.Sprintf(`{ model { text } requirement(id: %q) { limit } }`, id), &resp)
		want := "attribute :>> required = " + strconv.FormatFloat(resp.Requirement.Limit, 'f', -1, 64) + ";"
		if !strings.Contains(resp.Model.Text, want) {
			w.disagreed = append(w.disagreed, want)
		}
	}
}

func (w *world) everyReadAgrees() {
	for _, want := range w.disagreed {
		w.t.Errorf("a read was given a limit its text does not carry: %s", want)
	}
}

func (w *world) thePatchesLanded() {
	w.stopWriters()
	assert.Equal(w.t, w.failures.Load(), int32(0))
	assert.True(w.t, w.store.Version() > 1, "patches landed while the reads ran")
}

func (w *world) theLimitIsBound(id string) {
	var answer struct{ Requirement struct{ LimitEditable bool } }
	w.c.MustPost(fmt.Sprintf(`{ requirement(id: %q) { limitEditable } }`, id), &answer)
	assert.Equal(w.t, answer.Requirement.LimitEditable, false)
}

func (w *world) refusedAsNotEditable() {
	assert.Error(w.t, w.err)
	if w.err != nil {
		assert.True(w.t, strings.Contains(w.err.Error(), model.ErrNotEditable.Error()), "the refusal names the reason: "+w.err.Error())
	}
}

// anInputIsSubmitted sends what a visitor typed. An empty value and text
// reach the graph as strings, and infinity and not a number as the variable a
// client would send, since JSON has no number for either.
func (w *world) anInputIsSubmitted(input, id, name string) {
	var out mutationOut
	switch input {
	case "an empty value":
		w.err = w.c.Post(mutation(id, name, `""`), &out)
	case "the text abc":
		w.err = w.c.Post(mutation(id, name, `"abc"`), &out)
	case "infinity", "not a number":
		value := map[string]string{"infinity": "Infinity", "not a number": "NaN"}[input]
		query := fmt.Sprintf(`mutation($v: Float!) { setAttribute(partId: %q, name: %q, value: $v) { id } }`, id, name)
		w.err = w.c.Post(query, &out, client.Var("v", value))
	default:
		w.err = w.c.Post(mutation(id, name, input), &out)
	}
}

func (w *world) theMutationIsRejected() { assert.Error(w.t, w.err) }
