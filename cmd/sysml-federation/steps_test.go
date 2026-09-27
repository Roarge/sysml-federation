package main

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/examples/pipeline/ui"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^a stand-in for the router$`, w.aStandInForTheRouter)
	sc.Given(`^the demo is running with a stand-in for the router$`, w.theDemoIsRunning)
	sc.Given(`^the router's configuration file is not named$`, func() {})
	sc.Given(`^the router's configuration file is named as (\S+)$`, w.theConfigurationFileIsNamed)
	sc.Given(`^no router answers behind the UI server$`, w.noRouterAnswers)
	sc.When(`^the demo starts$`, w.theDemoStarts)
	sc.When(`^the four paths are requested on the published port$`, w.theFourPathsAreRequested)
	sc.When(`^the root path is requested$`, w.theRootPathIsRequested)
	sc.When(`^each app's page and the GraphQL endpoint are requested$`, w.eachPageAndTheEndpointAreRequested)
	sc.Then(`^the viewer answers within ten seconds of the start$`, w.theViewerAnswersWithinTenSeconds)
	sc.Then(`^the three subgraphs, the router and the document all answer$`, w.theRestAnswer)
	sc.Then(`^the router's environment sets:$`, w.theRoutersEnvironmentSets)
	sc.Then(`^every address the composed configuration names is on loopback$`, w.everyComposedAddressIsOnLoopback)
	sc.Then(`^the router's environment carries the nine variables the supervisor sets$`, w.theNineVariables)
	sc.Then(`^its CONFIG_PATH is absent$`, w.itsConfigPathIsAbsent)
	sc.Then(`^its CONFIG_PATH is (/\S+)$`, w.itsConfigPathIs)
	sc.Then(`^each is served by the app or the endpoint it names:$`, w.eachIsServedBy)
	sc.Then(`^it redirects to (\S+)$`, w.itRedirectsTo)
	sc.Then(`^both pages are served$`, w.bothPagesAreServed)
	sc.Then(`^every request to /graphql fails with 502 Bad Gateway$`, w.everyGraphQLRequestFails)
	sc.Then(`^neither app's files name the address of a subgraph$`, w.noSubgraphAddress)
}

// nineVariables are the names of the router's environment as the supervisor
// sets it (AD-0010).
var nineVariables = []string{
	"LISTEN_ADDR", "EXECUTION_CONFIG_FILE_PATH", "PLAYGROUND_PATH", "DO_NOT_TRACK", "COSMO_TELEMETRY_DISABLED",
	"TRACING_ENABLED", "METRICS_OTLP_ENABLED", "SUBGRAPH_ERROR_PROPAGATION_MODE", "PROMETHEUS_ENABLED",
}

// world is what one scenario of this package knows.
type world struct {
	t       *testing.T
	s       *supervisor
	fake    *fakeRouter
	base    string
	ready   time.Duration
	pages   map[string]page
	graphql page
}

// page is one answer of the UI server with its body.
type page struct {
	status   int
	location string
	body     string
}

func (w *world) fetch(method, path string) page {
	w.t.Helper()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequest(method, w.base+path, strings.NewReader(`{"query":"{ __typename }"}`))
	assert.NoError(w.t, err)
	req.Header.Set("Content-Type", "application/json")
	got, err := client.Do(req) //nolint:bodyclose // closed below, once assert.Must has proved there is one
	resp := assert.Must(w.t, got, err)
	defer resp.Body.Close()
	read, err := io.ReadAll(resp.Body)
	return page{status: resp.StatusCode, location: resp.Header.Get("Location"), body: string(assert.Must(w.t, read, err))}
}

func (w *world) aStandInForTheRouter() { w.s, w.fake = newSupervisor(w.t) }

func (w *world) theConfigurationFileIsNamed(path string) { w.s.routerConfigFile = path }

func (w *world) theDemoStarts() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	started := time.Now()
	go func() { done <- w.s.run(ctx) }()
	w.t.Cleanup(func() {
		cancel()
		assert.NoError(w.t, <-done)
	})
	w.base = "http://" + w.s.addrs.ui
	waitHTTP(w.t, w.base+"/viewer/")
	w.ready = time.Since(started)
}

func (w *world) theDemoIsRunning() {
	w.aStandInForTheRouter()
	w.theDemoStarts()
}

func (w *world) theViewerAnswersWithinTenSeconds() {
	assert.True(w.t, w.ready < 10*time.Second, "the viewer answered within ten seconds, after "+w.ready.String())
}

func (w *world) theRestAnswer() {
	for _, u := range []string{
		"http://" + w.s.addrs.adapter + "/health", "http://" + w.s.addrs.capacity + "/health",
		"http://" + w.s.addrs.document + "/health", "http://" + w.s.addrs.router + "/health/ready", w.base + "/document/",
	} {
		assert.Equal(w.t, get(w.t, u).StatusCode, http.StatusOK)
	}
}

func (w *world) theRoutersEnvironmentSets(table *godog.Table) {
	for _, row := range table.Rows[1:] {
		name, value := row.Cells[0].Value, row.Cells[1].Value
		assert.Equal(w.t, envValue(w.fake.env, name), value)
	}
}

func (w *world) everyComposedAddressIsOnLoopback() {
	read, err := os.ReadFile("../../examples/pipeline/config.json")
	raw := assert.Must(w.t, read, err)
	var cfg composedConfig
	assert.NoError(w.t, json.Unmarshal(raw, &cfg))
	var named []string
	for _, s := range cfg.Subgraphs {
		named = append(named, s.RoutingURL)
	}
	for _, d := range cfg.EngineConfig.Datasources {
		named = append(named, d.CustomGraphql.Fetch.URL.Static)
	}
	assert.True(w.t, len(named) == 6, "three subgraphs, each with a routing URL and a fetch URL")
	for _, address := range named {
		parsed, err := url.Parse(address)
		u := assert.Must(w.t, parsed, err)
		ip := net.ParseIP(u.Hostname())
		assert.True(w.t, ip != nil && ip.IsLoopback(), address+" is on loopback")
	}
}

func (w *world) theNineVariables() {
	for _, name := range nineVariables {
		assert.True(w.t, strings.Contains(strings.Join(w.fake.env, "\n")+"\n", name+"="), name+" is set")
	}
	for _, kv := range w.fake.env {
		name, _, _ := strings.Cut(kv, "=")
		known := name == "CONFIG_PATH"
		for _, nine := range nineVariables {
			known = known || name == nine
		}
		assert.True(w.t, known, kv+" is one of the nine, or the configuration file named")
	}
}

func (w *world) itsConfigPathIsAbsent() {
	for _, kv := range w.fake.env {
		assert.True(w.t, !strings.HasPrefix(kv, "CONFIG_PATH="), "no configuration file is handed over: "+kv)
	}
}

func (w *world) itsConfigPathIs(path string) {
	assert.Equal(w.t, envValue(w.fake.env, "CONFIG_PATH"), path)
}

func (w *world) theFourPathsAreRequested() {
	w.pages = map[string]page{
		"/viewer/":    w.fetch(http.MethodGet, "/viewer/"),
		"/document/":  w.fetch(http.MethodGet, "/document/"),
		"/graphql":    w.fetch(http.MethodPost, "/graphql"),
		"/playground": w.fetch(http.MethodGet, "/playground"),
	}
}

// servedBy tells from an answer which of the four served it: a page by its
// title, and the router by what the stand-in answers.
func servedBy(p page) string {
	switch {
	case strings.Contains(p.body, "<title>Model viewer</title>"):
		return "the Model viewer page"
	case strings.Contains(p.body, "<title>Requirements document</title>"):
		return "the Requirements document"
	case p.body == `{"data":{"__typename":"Query"}}` || p.body == "<html>playground</html>":
		return "the router"
	}
	return "nothing the scenario names"
}

func (w *world) eachIsServedBy(table *godog.Table) {
	for _, row := range table.Rows[1:] {
		path, by := row.Cells[0].Value, row.Cells[1].Value
		got := w.pages[path]
		assert.Equal(w.t, got.status, http.StatusOK)
		assert.Equal(w.t, servedBy(got), by)
	}
}

func (w *world) theRootPathIsRequested() {
	w.pages = map[string]page{"/": w.fetch(http.MethodGet, "/")}
}

func (w *world) itRedirectsTo(location string) {
	assert.Equal(w.t, w.pages["/"].status, http.StatusFound)
	assert.Equal(w.t, w.pages["/"].location, location)
}

func (w *world) noRouterAnswers() {
	parsed, err := url.Parse("http://" + freeAddr(w.t))
	target := assert.Must(w.t, parsed, err)
	built, err := uiHandler(ui.Files, target)
	srv := httptest.NewServer(assert.Must(w.t, built, err))
	w.t.Cleanup(srv.Close)
	w.base = srv.URL
}

func (w *world) eachPageAndTheEndpointAreRequested() {
	w.pages = map[string]page{
		"/viewer/":   w.fetch(http.MethodGet, "/viewer/"),
		"/document/": w.fetch(http.MethodGet, "/document/"),
	}
	w.graphql = w.fetch(http.MethodPost, "/graphql")
}

func (w *world) bothPagesAreServed() {
	assert.Equal(w.t, servedBy(w.pages["/viewer/"]), "the Model viewer page")
	assert.Equal(w.t, servedBy(w.pages["/document/"]), "the Requirements document")
}

func (w *world) everyGraphQLRequestFails() {
	assert.Equal(w.t, w.graphql.status, http.StatusBadGateway)
}

func (w *world) noSubgraphAddress() {
	var ports []string
	for _, addr := range []string{adapterAddr, capacityAddr, documentAddr} {
		_, port, err := net.SplitHostPort(addr)
		ports = append(ports, ":"+assert.Must(w.t, port, err))
	}
	err := fs.WalkDir(ui.Files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		body, err := fs.ReadFile(ui.Files, name)
		if err != nil {
			return err
		}
		for _, port := range ports {
			assert.True(w.t, !strings.Contains(string(body), port), name+" names no subgraph's port "+port)
		}
		return nil
	})
	assert.NoError(w.t, err)
}
