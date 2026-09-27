package document

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the document service at version 1$`, w.theDocumentService)
	sc.Given(`^a subscriber on the document service's subscription$`, w.aSubscriber)
	sc.Given(`^(PIPE-\S+) has been excluded$`, w.isExcluded)
	sc.When(`^the document is read$`, func() {})
	sc.When(`^prose reading "([^"]*)" is added at the top$`, w.proseAtTheTop)
	sc.When(`^prose reading "([^"]*)" is added as (PIPE-\S+)'s first child$`, w.proseAsFirstChild)
	sc.When(`^(PIPE-\S+) is moved under (PIPE-\S+) at position (\d+)$`, w.isMovedUnder)
	sc.When(`^a heading "([^"]*)" is inserted above (PIPE-\S+)$`, w.aHeadingIsInserted)
	sc.When(`^the opening prose is rewritten as "([^"]*)"$`, w.theOpeningProseIsRewritten)
	sc.When(`^(PIPE-\S+) is excluded$`, w.isExcluded)
	sc.When(`^(PIPE-\S+) is restored$`, w.isRestored)
	sc.Then(`^the subscriber receives document version (\d+)$`, w.theSubscriberReceives)
	sc.Then(`^the document version is (\d+)$`, w.theDocumentVersionIs)
	sc.Then(`^the numbers are (.+)$`, w.theNumbersAre)
	sc.Then(`^the document's two prose nodes carry no number$`, w.twoProseNodesCarryNoNumber)
	sc.Then(`^(\S+) is numbered (\S+)$`, w.isNumbered)
	sc.Then(`^(PIPE-\S+)'s first child is prose reading "([^"]*)"$`, w.firstChildIsProse)
	sc.Then(`^the opening prose reads "([^"]*)"$`, w.theOpeningProseReads)
	sc.Then(`^(PIPE-\S+) is out of the document$`, w.isOutOfTheDocument)
	sc.Then(`^(PIPE-\S+) is (PIPE-\S+)'s last child, numbered (\S+)$`, w.isTheLastChild)
}

// world is what one scenario of this package knows.
type world struct {
	t    *testing.T
	s    *Service
	c    *client.Client
	next func() json.RawMessage
}

func (w *world) theDocumentService() {
	w.c, w.s = newClient(w.t)
	assert.Equal(w.t, w.s.Version(), 1)
}

func (w *world) aSubscriber() {
	srv := httptest.NewServer(Handler(w.s))
	w.t.Cleanup(srv.Close)
	w.next = subscribe(w.t, srv.URL, "subscription { documentChanged }")
	waitForSubscribers(w.t, w.s.Subscribers, 1)
}

func (w *world) mutate(mutation string) {
	var out mutationOut
	w.c.MustPost(mutation, &out)
}

func (w *world) proseAtTheTop(text string) {
	w.mutate(fmt.Sprintf(`mutation { addProse(parentId: null, index: 0, text: %q) { version } }`, text))
}

func (w *world) proseAsFirstChild(text, parent string) {
	w.mutate(fmt.Sprintf(`mutation { addProse(parentId: %q, index: 0, text: %q) { version } }`, parent, text))
}

func (w *world) isMovedUnder(id, parent string, position int) {
	w.mutate(fmt.Sprintf(`mutation { moveNode(id: %q, parentId: %q, index: %d) { version } }`, id, parent, position-1))
}

func (w *world) aHeadingIsInserted(text, above string) {
	w.mutate(fmt.Sprintf(`mutation { insertHeading(aboveId: %q, text: %q) { version } }`, above, text))
}

func (w *world) theOpeningProseIsRewritten(text string) {
	w.mutate(fmt.Sprintf(`mutation { editText(id: "intro", text: %q) { version } }`, text))
}

func (w *world) isExcluded(id string) {
	w.mutate(fmt.Sprintf(`mutation { excludeRequirement(requirementId: %q) { version } }`, id))
}

func (w *world) isRestored(id string) {
	w.mutate(fmt.Sprintf(`mutation { includeRequirement(requirementId: %q) { version } }`, id))
}

func (w *world) theSubscriberReceives(version int) {
	var got struct {
		Data struct {
			DocumentChanged int `json:"documentChanged"`
		} `json:"data"`
	}
	assert.NoError(w.t, json.Unmarshal(w.next(), &got))
	assert.Equal(w.t, got.Data.DocumentChanged, version)
}

func (w *world) read() docOut {
	var resp docOut
	w.c.MustPost(docQuery, &resp)
	return resp
}

func (w *world) theDocumentVersionIs(version int) {
	assert.Equal(w.t, w.read().Document.Version, version)
	assert.Equal(w.t, w.s.Version(), version)
}

// theNumbersAre reads "1 PIPE-R1, 1.1 PIPE-R1.1" as every number the
// document carries, and nothing else.
func (w *world) theNumbersAre(list string) {
	want := make(map[string]string)
	for _, item := range strings.Split(list, ", ") {
		number, id, _ := strings.Cut(item, " ")
		want[id] = number
	}
	assert.MapEqual(w.t, numbers(w.read().Document.Nodes, map[string]string{}), want)
}

// proseNodes returns every prose node of the tree.
func proseNodes(nodes []nodeOut) []nodeOut {
	var found []nodeOut
	for _, n := range nodes {
		if n.Kind == "PROSE" {
			found = append(found, n)
		}
		found = append(found, proseNodes(n.Children)...)
	}
	return found
}

func (w *world) twoProseNodesCarryNoNumber() {
	var resp struct {
		Document struct {
			Nodes []nodeOut `json:"nodes"`
		} `json:"document"`
	}
	w.c.MustPost(`{ document { nodes { id kind number children { id kind number children { id kind number } } } } }`, &resp)
	prose := proseNodes(resp.Document.Nodes)
	assert.Len(w.t, prose, 2)
	for _, n := range prose {
		assert.True(w.t, n.Number == nil, n.ID+" carries no number")
	}
}

func (w *world) isNumbered(id, number string) {
	assert.Equal(w.t, numbers(w.read().Document.Nodes, map[string]string{})[id], number)
}

// childrenOf returns the children of the node with the given id, as the
// document query reaches them.
func (w *world) childrenOf(id string) []nodeOut {
	var resp struct {
		Document struct {
			Nodes []nodeOut `json:"nodes"`
		} `json:"document"`
	}
	w.c.MustPost(`{ document { nodes { id kind text children { id kind text number } } } }`, &resp)
	for _, n := range resp.Document.Nodes {
		if n.ID == id {
			return n.Children
		}
	}
	w.t.Fatalf("no top-level node %s", id)
	return nil
}

func (w *world) firstChildIsProse(parent, text string) {
	children := w.childrenOf(parent)
	assert.True(w.t, len(children) > 0, parent+" has children")
	assert.Equal(w.t, children[0].Kind, "PROSE")
	assert.True(w.t, children[0].Text != nil && *children[0].Text == text, "the first child reads "+text)
}

func (w *world) theOpeningProseReads(text string) {
	nodes := w.read().Document.Nodes
	assert.Equal(w.t, nodes[0].Kind, "PROSE")
	assert.True(w.t, nodes[0].Text != nil && *nodes[0].Text == text, "the opening prose reads "+text)
}

func (w *world) isOutOfTheDocument(id string) {
	_, numbered := numbers(w.read().Document.Nodes, map[string]string{})[id]
	assert.True(w.t, !numbered, id+" carries no number in the document")
}

func (w *world) isTheLastChild(id, parent, number string) {
	children := w.childrenOf(parent)
	assert.True(w.t, len(children) > 0, parent+" has children")
	last := children[len(children)-1]
	assert.Equal(w.t, last.ID, id)
	assert.True(w.t, last.Number != nil && *last.Number == number, id+" is numbered "+number)
}
