package pipeline

import (
	"encoding/json"
	"net/url"
	"os"
	"testing"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// configPath is the composed router configuration the image ships (AD-0012).
const configPath = "config.json"

// The parts of wgc's output the tests read. encoding/json ignores every
// other key, so the structs name only what is checked.
type routerConfig struct {
	Subgraphs    []subgraph   `json:"subgraphs"`
	EngineConfig engineConfig `json:"engineConfig"`
}

type subgraph struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RoutingURL string `json:"routingUrl"`
}

type engineConfig struct {
	DatasourceConfigurations []datasource `json:"datasourceConfigurations"`
}

type datasource struct {
	ID            string        `json:"id"`
	RootNodes     []rootNode    `json:"rootNodes"`
	CustomGraphql customGraphql `json:"customGraphql"`
}

// rootNode is one of the types a datasource answers for at the root. The
// router can only route a subscription to a datasource that lists
// Subscription here.
type rootNode struct {
	TypeName string `json:"typeName"`
}

type customGraphql struct {
	Federation   federation   `json:"federation"`
	Subscription subscription `json:"subscription"`
}

type federation struct {
	Enabled    bool   `json:"enabled"`
	ServiceSdl string `json:"serviceSdl"`
}

type subscription struct {
	Enabled              bool            `json:"enabled"`
	URL                  variableContent `json:"url"`
	Protocol             string          `json:"protocol"`
	WebsocketSubprotocol string          `json:"websocketSubprotocol"`
}

// variableContent is how the configuration wraps a value the router may
// template. The subscription URL is always static here.
type variableContent struct {
	Static string `json:"staticVariableContent"`
}

// schemaFiles maps each subgraph of graph.yaml to the schema file it was
// composed from.
var schemaFiles = map[string]string{
	"model":    "../../adapter/schema.graphql",
	"capacity": "capacity/schema.graphql",
	"document": "document/schema.graphql",
}

var routingURLs = map[string]string{
	"model":    "http://127.0.0.1:3011/graphql",
	"capacity": "http://127.0.0.1:3012/graphql",
	"document": "http://127.0.0.1:3013/graphql",
}

func loadConfig(t *testing.T) routerConfig {
	t.Helper()
	var cfg routerConfig
	data, err := os.ReadFile(configPath)
	raw := assert.Must(t, data, err)
	assert.NoError(t, json.Unmarshal(raw, &cfg))
	assert.Len(t, cfg.Subgraphs, 3)
	return cfg
}

// datasourceOf joins subgraphs[].name to datasourceConfigurations[].id.
func datasourceOf(t *testing.T, cfg routerConfig, name string) datasource {
	t.Helper()
	for _, sg := range cfg.Subgraphs {
		if sg.Name != name {
			continue
		}
		for _, ds := range cfg.EngineConfig.DatasourceConfigurations {
			if ds.ID == sg.ID {
				return ds
			}
		}
		t.Fatalf("subgraph %s (id %s) has no datasource configuration", name, sg.ID)
	}
	t.Fatalf("subgraph %s is not in config.json", name)
	return datasource{}
}

// TestSR42_ComposedSchemasMatchTheSchemaFiles fails when a schema file is
// edited without running make compose.
func TestSR42_ComposedSchemasMatchTheSchemaFiles(t *testing.T) {
	cfg := loadConfig(t)
	for name, file := range schemaFiles {
		ds := datasourceOf(t, cfg, name)
		data, err := os.ReadFile(file)
		want := string(assert.Must(t, data, err))
		assert.True(t, ds.CustomGraphql.Federation.Enabled, name+" is a federated datasource")
		assert.Equal(t, ds.CustomGraphql.Federation.ServiceSdl, want)
	}
}

// TestSR03_RoutingURLsAreLoopback: the router can only reach the three
// goroutines beside it.
func TestSR03_RoutingURLsAreLoopback(t *testing.T) {
	cfg := loadConfig(t)
	for _, sg := range cfg.Subgraphs {
		parsed, err := url.Parse(sg.RoutingURL)
		u := assert.Must(t, parsed, err)
		assert.Equal(t, u.Hostname(), "127.0.0.1")
		assert.Equal(t, sg.RoutingURL, routingURLs[sg.Name])
	}
}

// hasRootNode reports whether the datasource answers for the named root type.
func hasRootNode(ds datasource, typeName string) bool {
	for _, rn := range ds.RootNodes {
		if rn.TypeName == typeName {
			return true
		}
	}
	return false
}

// TestSR26_SubscriptionsAreComposedOverWebSocket: the two pushing subgraphs
// are reached over ws with graphql-transport-ws, the capacity service is not
// subscribed to at all.
func TestSR26_SubscriptionsAreComposedOverWebSocket(t *testing.T) {
	cfg := loadConfig(t)
	for _, name := range []string{"model", "document"} {
		ds := datasourceOf(t, cfg, name)
		assert.True(t, hasRootNode(ds, "Subscription"), name+" answers for Subscription")
		sub := ds.CustomGraphql.Subscription
		assert.True(t, sub.Enabled, name+" subscriptions are enabled")
		assert.Equal(t, sub.URL.Static, routingURLs[name])
		// proto3 JSON leaves an enum at its zero value out, and WS is value 0.
		assert.True(t, sub.Protocol == "" || sub.Protocol == "GRAPHQL_SUBSCRIPTION_PROTOCOL_WS", name+" uses ws, got "+sub.Protocol)
		assert.Equal(t, sub.WebsocketSubprotocol, "GRAPHQL_WEBSOCKET_SUBPROTOCOL_TRANSPORT_WS")
	}
	// wgc writes an enabled subscription block on every datasource, whether or
	// not the subgraph declares a Subscription type, so the enabled flag says
	// nothing. What decides where a subscription may go is the root nodes: the
	// capacity service answers for no Subscription field, so the router never
	// opens a socket to it.
	assert.True(t, !hasRootNode(datasourceOf(t, cfg, "capacity"), "Subscription"), "the capacity service has no subscription")
}
