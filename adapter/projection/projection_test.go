package projection

import (
	"bytes"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/Roarge/sysml-federation/adapter/model"
	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

// fixture is the model package's second fixture (SR-16). No word of the
// example appears in this package (SR-17).
const fixture = "../model/testdata/warehouse.sysml"

func f(v float64) *float64 { return &v }

// literal is a small model built by hand, so the query and entity tests
// depend on no parser. The lookups in projection.go scan the slices, which
// is what lets a model without the parser's indexes serve.
func literal() *model.Model {
	child := &model.Part{ID: "C1", ShortName: "C1", Name: "child", Definition: "Box",
		Attributes: []model.Attribute{{Name: "rate", Value: f(5), Unit: "Hz", Editable: true}, {Name: "size"}},
		Ports:      []model.Port{{Name: "feed", Direction: syntax.DirectionIn}, {Name: "drain", Direction: syntax.DirectionOut}},
		Satisfies:  []string{"R2"}}
	root := &model.Part{ID: "P1", ShortName: "P1", Name: "root", Definition: "Box", Doc: "the root",
		Attributes:  []model.Attribute{{Name: "total", Value: f(10), Expression: "child.rate * 2"}},
		Parts:       []*model.Part{child},
		Connections: []model.Connection{{ID: "C1.drain->C1.feed", From: "C1", FromPort: "drain", To: "C1", ToPort: "feed"}},
		Satisfies:   []string{"R1"}}
	r1 := &model.Requirement{ID: "R1", ShortName: "R1", Name: "total", Text: "the root shall keep up", Subject: "P1",
		Quantity: "total", Comparison: syntax.GE, Limit: 8, LimitUnit: "Hz", LimitEditable: true,
		Derives: []string{"R2"}, SatisfiedBy: []string{"P1"}, VerifiedBy: []string{"V1"}}
	r2 := &model.Requirement{ID: "R2", Name: "rate", Subject: "C1", Quantity: "rate", Comparison: syntax.LT, Limit: 9,
		DerivedFrom: []string{"R1"}, SatisfiedBy: []string{"C1"}}
	r3 := &model.Requirement{ID: "R3", Name: "free", Quantity: "x", Comparison: syntax.EQ, Limit: 1}
	vc := &model.VerificationCase{ID: "V1", ShortName: "V1", Name: "check", Verifies: []string{"R1"}}
	return &model.Model{Version: 1, Text: "part root", Roots: []*model.Part{root}, Parts: []*model.Part{root, child},
		Requirements: []*model.Requirement{r1, r2, r3}, VerificationCases: []*model.VerificationCase{vc}}
}

// tap is the handler the client posts through. It keeps the bytes of the
// answer, because a response's fields come back in the order the query asked
// for them and the client's decoded form no longer has that order.
type tap struct {
	handler http.Handler
	body    []byte
}

func (t *tap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := httptest.NewRecorder()
	t.handler.ServeHTTP(rec, r)
	t.body = rec.Body.Bytes()
	maps.Copy(w.Header(), rec.Header())
	w.WriteHeader(rec.Code)
	_, _ = w.Write(t.body)
}

// tapped is the test client, with the last answer's bytes beside it.
type tapped struct {
	*client.Client
	tap *tap
}

func newClient(t *testing.T, m *model.Model) (*tapped, *Store) {
	t.Helper()
	store := NewStore(m)
	srv := NewServer(store)
	srv.AddTransport(transport.POST{})
	through := &tap{handler: srv}
	return &tapped{Client: client.New(through, client.Path("/graphql")), tap: through}, store
}

func loadClient(t *testing.T) (*tapped, *Store) {
	t.Helper()
	m, err := model.Load(fixture)
	return newClient(t, assert.Must(t, m, err))
}

// mutationOut takes the answer of a mutation whose fields the test reads
// elsewhere. The client refuses to decode a key its target has no field for,
// so an empty struct serves only where the mutation is expected to fail.
type mutationOut struct {
	SetAttribute struct{ ID string }   `json:"setAttribute"`
	SetLimit     struct{ ID string }   `json:"setLimit"`
	ResetModel   struct{ Version int } `json:"resetModel"`
}

// raw runs a query and returns the compact JSON of its data.
func raw(t *testing.T, c *tapped, query string, opts ...client.Option) string {
	t.Helper()
	_, err := c.RawPost(query, opts...)
	assert.NoError(t, err)
	var answer struct {
		Data   json.RawMessage `json:"data"`
		Errors json.RawMessage `json:"errors"`
	}
	assert.NoError(t, json.Unmarshal(c.tap.body, &answer))
	assert.True(t, answer.Errors == nil, "the operation answered without errors, got "+string(answer.Errors))
	return string(answer.Data)
}

const partsQuery = `{ model { version text roots { id shortName name definition doc
  attributes { name value unit editable expression } ports { name direction }
  connections { id from fromPort to toPort } satisfies { id }
  parts { id shortName name definition doc attributes { name value unit editable expression }
    ports { name direction } parts { id } connections { id } satisfies { id } } } } }`

// escaped spells an expected answer the way the JSON encoder writes it. The
// encoder replaces the greater-than sign of a connection identifier with its
// escape, and the answers below are written with the character itself.
func escaped(s string) string {
	var buf bytes.Buffer
	json.HTMLEscape(&buf, []byte(s))
	return buf.String()
}

const partsAnswer = `{"model":{"version":1,"text":"part root","roots":[{"id":"P1","shortName":"P1","name":"root","definition":"Box","doc":"the root",` +
	`"attributes":[{"name":"total","value":10,"unit":"","editable":false,"expression":"child.rate * 2"}],"ports":[],` +
	`"connections":[{"id":"C1.drain->C1.feed","from":"C1","fromPort":"drain","to":"C1","toPort":"feed"}],"satisfies":[{"id":"R1"}],` +
	`"parts":[{"id":"C1","shortName":"C1","name":"child","definition":"Box","doc":"",` +
	`"attributes":[{"name":"rate","value":5,"unit":"Hz","editable":true,"expression":""},{"name":"size","value":null,"unit":"","editable":false,"expression":""}],` +
	`"ports":[{"name":"feed","direction":"IN"},{"name":"drain","direction":"OUT"}],"parts":[],"connections":[],"satisfies":[{"id":"R2"}]}]}]}}`

const requirementsQuery = `{ model { requirements { id shortName name text subject { id } quantity comparison limit limitUnit limitEditable
  derivedFrom { id } derives { id } satisfiedBy { id } verifiedBy { id } }
  verificationCases { id shortName name verifies { id } } }
  part(id: "C1") { name } requirement(id: "R2") { subject { name } } }`

const requirementsAnswer = `{"model":{"requirements":[` +
	`{"id":"R1","shortName":"R1","name":"total","text":"the root shall keep up","subject":{"id":"P1"},"quantity":"total","comparison":"GE","limit":8,"limitUnit":"Hz","limitEditable":true,"derivedFrom":[],"derives":[{"id":"R2"}],"satisfiedBy":[{"id":"P1"}],"verifiedBy":[{"id":"V1"}]},` +
	`{"id":"R2","shortName":"","name":"rate","text":"","subject":{"id":"C1"},"quantity":"rate","comparison":"LT","limit":9,"limitUnit":"","limitEditable":false,"derivedFrom":[{"id":"R1"}],"derives":[],"satisfiedBy":[{"id":"C1"}],"verifiedBy":[]},` +
	`{"id":"R3","shortName":"","name":"free","text":"","subject":null,"quantity":"x","comparison":"EQ","limit":1,"limitUnit":"","limitEditable":false,"derivedFrom":[],"derives":[],"satisfiedBy":[],"verifiedBy":[]}],` +
	`"verificationCases":[{"id":"V1","shortName":"V1","name":"check","verifies":[{"id":"R1"}]}]},"part":{"name":"child"},"requirement":{"subject":{"name":"child"}}}`

func TestQueriesServeEveryFieldOfTheProjection(t *testing.T) {
	c, _ := newClient(t, literal())
	assert.Equal(t, raw(t, c, partsQuery), escaped(partsAnswer))
	assert.Equal(t, raw(t, c, requirementsQuery), requirementsAnswer)
	assert.Equal(t, raw(t, c, `{ part(id: "nope") { id } requirement(id: "nope") { id } }`), `{"part":null,"requirement":null}`)
}

const entities = `query($reps: [_Any!]!) { _entities(representations: $reps) {
  ... on Part { id name } ... on Requirement { id limit } ... on VerificationCase { id name } } }`

func TestEntitiesResolveTheThreeKeyedTypes(t *testing.T) {
	c, _ := newClient(t, literal())
	reps := client.Var("reps", json.RawMessage(`[{"__typename":"Part","id":"C1"},{"__typename":"Requirement","id":"R2"},{"__typename":"VerificationCase","id":"V1"}]`))
	assert.Equal(t, raw(t, c, entities, reps), `{"_entities":[{"id":"C1","name":"child"},{"id":"R2","limit":9},{"id":"V1","name":"check"}]}`)
	var out struct{}
	err := c.Post(entities, &out, client.Var("reps", json.RawMessage(`[{"__typename":"Part","id":"nope"}]`)))
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), model.ErrNotFound.Error()), "an unknown entity id is reported as not found")
}

func TestSR22_SetAttributePatchesTheTextAndBumpsTheVersion(t *testing.T) {
	c, store := loadClient(t)
	got := raw(t, c, `mutation { setAttribute(partId: "WH-A", name: "rate", value: 45) { id attributes { name value } } }`)
	assert.Equal(t, got, `{"setAttribute":{"id":"WH-A","attributes":[{"name":"capacity","value":null},{"name":"rate","value":45}]}}`)
	assert.Equal(t, store.Version(), 2)
	var resp struct {
		Model struct {
			Version int
			Text    string
		}
	}
	c.MustPost(`{ model { version text } }`, &resp)
	assert.Equal(t, resp.Model.Version, 2)
	assert.True(t, strings.Contains(resp.Model.Text, "attribute :>> rate = 45;"), "the served text carries the new literal")
	assert.True(t, !strings.Contains(resp.Model.Text, "rate = 40;"), "the old literal is gone")
}

func TestSR23_SetLimitMovesTheDerivedLimits(t *testing.T) {
	c, store := loadClient(t)
	got := raw(t, c, `mutation { setLimit(requirementId: "WH-R1", value: 80) { id limit derives { limit } } }`)
	assert.Equal(t, got, `{"setLimit":{"id":"WH-R1","limit":80,"derives":[{"limit":50},{"limit":50}]}}`)
	assert.Equal(t, store.Version(), 2)
}

func TestSR24_ExpressionBoundAndUnboundValuesAreRefused(t *testing.T) {
	c, store := loadClient(t)
	var out struct{}
	for _, m := range []string{
		`mutation { setLimit(requirementId: "WH-R1.1", value: 1) { id } }`,
		`mutation { setAttribute(partId: "WH-L1", name: "capacity", value: 1) { id } }`,
	} {
		err := c.Post(m, &out)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), model.ErrNotEditable.Error()), "the refusal names the reason: "+m)
	}
	assert.Equal(t, store.Version(), 1)
}

func TestSR25_InvalidValuesAreRefused(t *testing.T) {
	c, store := loadClient(t)
	cases := []tabletest.Case[string, string]{
		{Name: "negative", In: `mutation { setAttribute(partId: "WH-A", name: "rate", value: -1) { id } }`, Want: model.ErrInvalidValue.Error()},
		{Name: "text", In: `mutation { setAttribute(partId: "WH-A", name: "rate", value: "abc") { id } }`, Want: "Float"},
		{Name: "empty", In: `mutation { setAttribute(partId: "WH-A", name: "rate", value: "") { id } }`, Want: "Float"},
	}
	tabletest.Run(t, cases, func(t *testing.T, m string) string {
		var out struct{}
		err := c.Post(m, &out)
		assert.Error(t, err)
		for _, want := range []string{model.ErrInvalidValue.Error(), "Float"} {
			if strings.Contains(err.Error(), want) {
				return want
			}
		}
		return err.Error()
	})
	var out struct{}
	assert.Error(t, c.Post(`mutation($v: Float!) { setAttribute(partId: "WH-A", name: "rate", value: $v) { id } }`, &out, client.Var("v", "NaN")))
	assert.Equal(t, store.Version(), 1)
	assert.Equal(t, raw(t, c, `{ part(id: "WH-A") { attributes { name value } } }`), `{"part":{"attributes":[{"name":"capacity","value":null},{"name":"rate","value":40}]}}`)
}

func TestUnknownIDsAreErrors(t *testing.T) {
	c, store := loadClient(t)
	var out struct{}
	for _, m := range []string{
		`mutation { setAttribute(partId: "nope", name: "rate", value: 1) { id } }`,
		`mutation { setAttribute(partId: "WH-A", name: "nope", value: 1) { id } }`,
		`mutation { setLimit(requirementId: "nope", value: 1) { id } }`,
	} {
		err := c.Post(m, &out)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), model.ErrNotFound.Error()), "the refusal names the reason: "+m)
	}
	assert.Equal(t, store.Version(), 1)
}

func TestSR44_ResetRestoresTheShippedValuesAndGrowsTheVersion(t *testing.T) {
	c, store := loadClient(t)
	var out mutationOut
	c.MustPost(`mutation { setAttribute(partId: "WH-A", name: "rate", value: 45) { id } }`, &out)
	assert.Equal(t, raw(t, c, `mutation { resetModel { version } }`), `{"resetModel":{"version":3}}`)
	assert.Equal(t, raw(t, c, `{ model { version } part(id: "WH-A") { attributes { name value } } }`),
		`{"model":{"version":3},"part":{"attributes":[{"name":"capacity","value":null},{"name":"rate","value":40}]}}`)
	assert.Equal(t, raw(t, c, `mutation { resetModel { version } }`), `{"resetModel":{"version":4}}`) // C-92: the counter never goes back
	assert.Equal(t, store.Version(), 4)
	assert.True(t, strings.Contains(store.Current().Text, "rate = 40;"), "the shipped text is served again")
}

// TestSR09_AFreshStoreIsTheShippedState: nothing an edit did survives a new store.
func TestSR09_AFreshStoreIsTheShippedState(t *testing.T) {
	loaded, err := Load(fixture)
	first := assert.Must(t, loaded, err)
	_, err = first.SetAttribute("WH-A", "rate", 45)
	assert.NoError(t, err)
	reloaded, err := Load(fixture)
	second := assert.Must(t, reloaded, err)
	assert.Equal(t, second.Version(), 1)
	assert.True(t, strings.Contains(second.Current().Text, "rate = 40;"), "the shipped literal")
}

// TestSR22_TextAndProjectionAgreeUnderConcurrentEdits: four writers patch
// the global limit while a reader queries text and limit in one operation.
// The two must always come from the same model. The writers count their
// own failures rather than calling into testing.T from another goroutine.
func TestSR22_TextAndProjectionAgreeUnderConcurrentEdits(t *testing.T) {
	c, store := loadClient(t)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	var failures atomic.Int32
	for i := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; ; j++ {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := store.SetLimit("WH-R1", float64(100+i*1000+j)); err != nil {
					failures.Add(1)
					return
				}
			}
		}()
	}
	for range 50 {
		var resp struct {
			Model       struct{ Text string }
			Requirement struct{ Limit float64 }
		}
		c.MustPost(`{ model { text } requirement(id: "WH-R1") { limit } }`, &resp)
		want := "attribute :>> required = " + strconv.FormatFloat(resp.Requirement.Limit, 'f', -1, 64) + ";"
		assert.True(t, strings.Contains(resp.Model.Text, want), "text and projection agree on "+want)
	}
	close(stop)
	wg.Wait()
	assert.Equal(t, failures.Load(), int32(0))
	assert.True(t, store.Version() > 1, "edits landed while reading")
}
