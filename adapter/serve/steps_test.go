package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/adapter/projection"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the adapter serves the second fixture at version 1$`, w.theAdapterServesTheSecondFixture)
	sc.Given(`^a subscriber on the adapter's subscription$`, w.aSubscriber)
	sc.When(`^(WH-\S+)'s (\w+) is set to (\d+)$`, w.aValueIsSet)
	sc.Then(`^the subscriber receives model version (\d+)$`, w.theSubscriberReceives)
}

// world is what one scenario of this package knows.
type world struct {
	t     *testing.T
	store *projection.Store
	h     http.Handler
	url   string
	next  func() json.RawMessage
}

func (w *world) theAdapterServesTheSecondFixture() {
	loaded, err := projection.Load(fixture)
	w.store = assert.Must(w.t, loaded, err)
	assert.Equal(w.t, w.store.Version(), 1)
	w.h = Handler(w.store)
	srv := httptest.NewServer(w.h)
	w.t.Cleanup(srv.Close)
	w.url = srv.URL
}

func (w *world) aSubscriber() {
	w.next = subscribe(w.t, w.url, "subscription { modelChanged }")
	waitForSubscribers(w.t, w.store.Subscribers, 1)
}

func (w *world) aValueIsSet(id, name, value string) {
	m := fmt.Sprintf(`mutation { setAttribute(partId: %q, name: %q, value: %s) { id } }`, id, name, value)
	if name == "limit" {
		m = fmt.Sprintf(`mutation { setLimit(requirementId: %q, value: %s) { id } }`, id, value)
	}
	var out mutationOut
	client.New(w.h, client.Path("/graphql")).MustPost(m, &out)
}

func (w *world) theSubscriberReceives(version int) {
	var got struct {
		Data struct {
			ModelChanged int `json:"modelChanged"`
		} `json:"data"`
	}
	assert.NoError(w.t, json.Unmarshal(w.next(), &got))
	assert.Equal(w.t, got.Data.ModelChanged, version)
}
