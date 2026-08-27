package document

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/client"

	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

const docQuery = `{ document { version nodes { id kind number text requirement { id documentNumber included } children { id number children { id number } } } } }`

type nodeOut struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	Number      *string `json:"number"`
	Text        *string `json:"text"`
	Requirement *struct {
		ID             string  `json:"id"`
		DocumentNumber *string `json:"documentNumber"`
		Included       bool    `json:"included"`
	} `json:"requirement"`
	Children []nodeOut `json:"children"`
}

type docOut struct {
	Document struct {
		Version int       `json:"version"`
		Nodes   []nodeOut `json:"nodes"`
	} `json:"document"`
}

// mutationOut absorbs a mutation's answer under whatever field name it
// carries. The test client refuses a key its target has no place for, so a
// mutation cannot be posted into an empty struct.
type mutationOut map[string]struct {
	Version int `json:"version"`
}

func newClient(t *testing.T) (*client.Client, *Service) {
	t.Helper()
	svc, err := New()
	s := assert.Must(t, svc, err)
	return client.New(Handler(s), client.Path("/graphql")), s
}

func numbers(nodes []nodeOut, out map[string]string) map[string]string {
	for _, n := range nodes {
		if n.Number != nil {
			out[n.ID] = *n.Number
		}
		numbers(n.Children, out)
	}
	return out
}

func TestSR33_ShippedDocument(t *testing.T) {
	c, _ := newClient(t)
	var resp docOut
	c.MustPost(docQuery, &resp)
	assert.Equal(t, resp.Document.Version, 1)
	assert.Equal(t, resp.Document.Nodes[0].Kind, "PROSE")
	assert.True(t, resp.Document.Nodes[0].Number == nil, "prose is unnumbered")
	assert.Equal(t, *resp.Document.Nodes[1].Requirement.DocumentNumber, "1")
	assert.MapEqual(t, numbers(resp.Document.Nodes, map[string]string{}), map[string]string{
		"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.4": "1.4", "PIPE-R1.5": "1.5", "PIPE-R2": "2"})
}

func TestSR35_MutationsRenumberAndBumpTheVersion(t *testing.T) {
	cases := []tabletest.Case[string, map[string]string]{
		{Name: "moveNode", In: `mutation { moveNode(id: "PIPE-R1.5", parentId: "PIPE-R1", index: 0) { version } }`,
			Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.5": "1.1", "PIPE-R1.1": "1.2", "PIPE-R1.2": "1.3", "PIPE-R1.3": "1.4", "PIPE-R1.4": "1.5", "PIPE-R2": "2"}},
		{Name: "nest", In: `mutation { moveNode(id: "PIPE-R2", parentId: "PIPE-R1", index: 5) { version } }`,
			Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.4": "1.4", "PIPE-R1.5": "1.5", "PIPE-R2": "1.6"}},
		{Name: "insertHeading", In: `mutation { insertHeading(aboveId: "PIPE-R1", text: "Performance") { version } }`,
			Want: map[string]string{"n1": "1", "PIPE-R1": "1.1", "PIPE-R1.1": "1.1.1", "PIPE-R1.2": "1.1.2", "PIPE-R1.3": "1.1.3", "PIPE-R1.4": "1.1.4", "PIPE-R1.5": "1.1.5", "PIPE-R2": "2"}},
		{Name: "addProse", In: `mutation { addProse(parentId: "PIPE-R1", index: 0, text: "Allocated.") { version } }`,
			Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.4": "1.4", "PIPE-R1.5": "1.5", "PIPE-R2": "2"}},
		{Name: "excludeRequirement", In: `mutation { excludeRequirement(requirementId: "PIPE-R1.4") { version } }`,
			Want: map[string]string{"PIPE-R1": "1", "PIPE-R1.1": "1.1", "PIPE-R1.2": "1.2", "PIPE-R1.3": "1.3", "PIPE-R1.5": "1.4", "PIPE-R2": "2"}},
	}
	tabletest.Run(t, cases, func(t *testing.T, mutation string) map[string]string {
		c, s := newClient(t)
		var out mutationOut
		c.MustPost(mutation, &out)
		assert.Equal(t, s.Version(), 2)
		var resp docOut
		c.MustPost(docQuery, &resp)
		assert.Equal(t, resp.Document.Version, 2)
		return numbers(resp.Document.Nodes, map[string]string{})
	})
}

func TestSR35_IncludeAndEditText(t *testing.T) {
	c, s := newClient(t)
	var out mutationOut
	c.MustPost(`mutation { excludeRequirement(requirementId: "PIPE-R1.4") { version } }`, &out)
	c.MustPost(`mutation { includeRequirement(requirementId: "PIPE-R1.4") { version } }`, &out)
	c.MustPost(`mutation { editText(id: "intro", text: "Rewritten.") { version } }`, &out)
	assert.Equal(t, s.Version(), 4)
	var resp docOut
	c.MustPost(docQuery, &resp)
	assert.Equal(t, numbers(resp.Document.Nodes, map[string]string{})["PIPE-R1.4"], "1.5")
	assert.Equal(t, *resp.Document.Nodes[0].Text, "Rewritten.")
}

func TestARefusedMutationIsAnErrorAndLeavesTheVersion(t *testing.T) {
	c, s := newClient(t)
	var out mutationOut
	assert.Error(t, c.Post(`mutation { moveNode(id: "PIPE-R1", parentId: "PIPE-R1.2", index: 0) { version } }`, &out))
	assert.Equal(t, s.Version(), 1)
}

func TestSR44_ResetRestoresTheShippedTreeAndGrowsTheVersion(t *testing.T) {
	c, s := newClient(t)
	var out mutationOut
	c.MustPost(`mutation { moveNode(id: "PIPE-R2", parentId: "PIPE-R1", index: 0) { version } }`, &out)
	c.MustPost(`mutation { resetDocument { version } }`, &out)
	assert.Equal(t, s.Version(), 3)
	var resp docOut
	c.MustPost(docQuery, &resp)
	assert.Equal(t, numbers(resp.Document.Nodes, map[string]string{})["PIPE-R2"], "2")
}

func TestEntityAnswersForUnknownIDs(t *testing.T) {
	c, _ := newClient(t)
	var resp struct {
		Entities []struct {
			DocumentNumber *string `json:"documentNumber"`
			Included       bool    `json:"included"`
		} `json:"_entities"`
	}
	c.MustPost(`query($reps: [_Any!]!) { _entities(representations: $reps) { ... on Requirement { documentNumber included } } }`, &resp,
		client.Var("reps", json.RawMessage(`[{"__typename":"Requirement","id":"PIPE-R1.2"},{"__typename":"Requirement","id":"PIPE-R9"}]`)))
	assert.Equal(t, *resp.Entities[0].DocumentNumber, "1.2")
	assert.True(t, resp.Entities[0].Included, "PIPE-R1.2 is in the shipped tree")
	assert.True(t, resp.Entities[1].DocumentNumber == nil && !resp.Entities[1].Included, "an unknown id is not included and has no number")
}

func TestSR27_DocumentChangedReachesEverySubscriber(t *testing.T) {
	svc, err := New()
	s := assert.Must(t, svc, err)
	srv := httptest.NewServer(Handler(s))
	t.Cleanup(srv.Close)
	first := subscribe(t, srv.URL, "subscription { documentChanged }")
	second := subscribe(t, srv.URL, "subscription { documentChanged }")
	waitForSubscribers(t, s.Subscribers, 2)
	c := client.New(Handler(s), client.Path("/graphql"))
	var out mutationOut
	c.MustPost(`mutation { addProse(parentId: null, index: 0, text: "x") { version } }`, &out)
	for _, next := range []func() json.RawMessage{first, second} {
		var got struct {
			Data struct {
				DocumentChanged int `json:"documentChanged"`
			} `json:"data"`
		}
		assert.NoError(t, json.Unmarshal(next(), &got))
		assert.Equal(t, got.Data.DocumentChanged, 2)
	}
}

// TestSR09_RestartReturnsToShipped: a fresh service is the shipped state.
func TestSR09_RestartReturnsToShipped(t *testing.T) {
	c, _ := newClient(t)
	var out mutationOut
	c.MustPost(`mutation { excludeRequirement(requirementId: "PIPE-R2") { version } }`, &out)
	c2, s2 := newClient(t)
	assert.Equal(t, s2.Version(), 1)
	var resp docOut
	c2.MustPost(docQuery, &resp)
	assert.Equal(t, numbers(resp.Document.Nodes, map[string]string{})["PIPE-R2"], "2")
}
