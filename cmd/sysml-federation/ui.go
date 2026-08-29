package main

import (
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// uiHandler is the UI server of AD-0011: the two apps and the shared module
// from the embedded files, /graphql and /playground proxied to the router,
// and / redirected to /viewer/ (SR-04). The router's health paths are not
// proxied, so the published port shows those paths and nothing else.
func uiHandler(assets fs.FS, router *url.URL) (http.Handler, error) {
	proxy := httputil.NewSingleHostReverseProxy(router)
	// Flush every write: a subscription over SSE is one response that never
	// ends, and a buffering proxy would hold its events back (C-62).
	proxy.FlushInterval = -1
	mux := http.NewServeMux()
	mux.Handle("/graphql", proxy)
	mux.Handle("/playground", proxy)
	mux.Handle("/playground/", proxy)
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/viewer/", http.StatusFound)
	})
	for _, dir := range []string{"viewer", "document", "shared"} {
		sub, err := fs.Sub(assets, dir)
		if err != nil {
			return nil, err
		}
		mux.Handle("/"+dir+"/", static("/"+dir+"/", sub))
	}
	return mux, nil
}

// static serves one app directory with no caching, since embedded files
// carry no modification time (C-63). A directory path other than the app's
// root would be listed, which the apps never need, so it is a 404.
func static(prefix string, files fs.FS) http.Handler {
	server := http.StripPrefix(prefix, http.FileServerFS(files))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != prefix {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		server.ServeHTTP(w, r)
	})
}
