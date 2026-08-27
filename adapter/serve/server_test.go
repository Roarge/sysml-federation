package serve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/client"

	"github.com/Roarge/sysml-federation/adapter/projection"
	"github.com/Roarge/sysml-federation/internal/assert"
)

const fixture = "../model/testdata/warehouse.sysml"

// mutationOut takes the answer of a mutation the test does not read. The
// client refuses to decode a key its target has no field for.
type mutationOut struct {
	SetAttribute struct{ ID string }   `json:"setAttribute"`
	SetLimit     struct{ ID string }   `json:"setLimit"`
	ResetModel   struct{ Version int } `json:"resetModel"`
}

func TestHandlerServesPostAndHealth(t *testing.T) {
	loaded, err := projection.Load(fixture)
	store := assert.Must(t, loaded, err)
	h := Handler(store)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, rec.Code, http.StatusOK)
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ model { version } }"}`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	assert.Equal(t, rec.Code, http.StatusOK)
	assert.Equal(t, strings.TrimSpace(rec.Body.String()), `{"data":{"model":{"version":1}}}`)
}

// TestSR26_ModelChangedReachesEverySubscriber: every accepted mutation,
// reset included, is pushed to every subscriber as the new version.
func TestSR26_ModelChangedReachesEverySubscriber(t *testing.T) {
	loaded, err := projection.Load(fixture)
	store := assert.Must(t, loaded, err)
	h := Handler(store)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	first := subscribe(t, srv.URL, "subscription { modelChanged }")
	second := subscribe(t, srv.URL, "subscription { modelChanged }")
	waitForSubscribers(t, store.Subscribers, 2)
	c := client.New(h, client.Path("/graphql"))
	mutations := []string{
		`mutation { setAttribute(partId: "WH-B", name: "rate", value: 35) { id } }`,
		`mutation { setLimit(requirementId: "WH-R1", value: 70) { id } }`,
		`mutation { resetModel { version } }`,
	}
	for i, m := range mutations {
		var out mutationOut
		c.MustPost(m, &out)
		for _, next := range []func() json.RawMessage{first, second} {
			var got struct {
				Data struct {
					ModelChanged int `json:"modelChanged"`
				} `json:"data"`
			}
			assert.NoError(t, json.Unmarshal(next(), &got))
			assert.Equal(t, got.Data.ModelChanged, i+2)
		}
	}
	var out mutationOut
	assert.Error(t, c.Post(`mutation { setLimit(requirementId: "WH-R1.1", value: 1) { id } }`, &out))
	assert.Equal(t, store.Version(), 4) // a refused mutation emits nothing
}
