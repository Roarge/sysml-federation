// Package serve mounts the adapter's subgraph over HTTP: /graphql for POST
// and for subscriptions over WebSocket, which is how the router reaches it
// (C-11), and /health for the supervisor.
package serve

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/Roarge/sysml-federation/adapter/projection"
)

// Handler serves the store on /graphql and answers /health.
func Handler(store *projection.Store) http.Handler {
	srv := projection.NewServer(store)
	srv.AddTransport(transport.POST{})
	// The bare implementation keeps the library's same-origin check, so a page
	// a browser happens to visit cannot open a socket to the subgraph. A
	// server-side client such as the router sends no origin at all, which the
	// library allows.
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Implementation:        transport.CoderWebsocketImplementation{},
	})
	srv.Use(extension.Introspection{})
	mux := http.NewServeMux()
	mux.Handle("/graphql", srv)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return mux
}
