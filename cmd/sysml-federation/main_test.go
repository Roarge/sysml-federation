package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Roarge/sysml-federation/examples/pipeline/capacity/flow"
	"github.com/Roarge/sysml-federation/examples/pipeline/ui"
	"github.com/Roarge/sysml-federation/internal/assert"
)

const testModel = "../../examples/pipeline/model.sysml"

// TestHelperRouter is not a test. It is the body of the child process that
// TestRouterRunsAsAChildProcess starts from this binary in place of
// /router: it listens where LISTEN_ADDR says, answers /health/ready and
// leaves on SIGTERM.
func TestHelperRouter(t *testing.T) {
	if os.Getenv("SYSML_FEDERATION_HELPER") != "1" {
		t.Skip("runs only as the helper process")
	}
	ln, err := net.Listen("tcp", envValue(os.Environ(), "LISTEN_ADDR"))
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	fmt.Println("helper router listening")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	go func() { _ = http.Serve(ln, mux) }()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop
	fmt.Println("helper router stopping")
	os.Exit(0)
}

func envValue(env []string, key string) string {
	for _, kv := range env {
		if v, ok := strings.CutPrefix(kv, key+"="); ok {
			return v
		}
	}
	return ""
}

func freeAddr(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	ln := assert.Must(t, listener, err)
	addr := ln.Addr().String()
	assert.NoError(t, ln.Close())
	return addr
}

// response is what a probe of the UI server is asserted on. The body is
// closed before the helper returns, so no open response outlives the call.
type response struct {
	StatusCode int
	Header     http.Header
}

func get(t *testing.T, url string) response {
	t.Helper()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	got, err := client.Get(url) //nolint:bodyclose // closed below, once assert.Must has proved there is one
	resp := assert.Must(t, got, err)
	defer resp.Body.Close()
	return response{StatusCode: resp.StatusCode, Header: resp.Header}
}

func post(t *testing.T, url, body string) string {
	t.Helper()
	got, err := http.Post(url, "application/json", strings.NewReader(body)) //nolint:bodyclose // closed below, once assert.Must has proved there is one
	resp := assert.Must(t, got, err)
	defer resp.Body.Close()
	read, err := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(assert.Must(t, read, err)))
}

func waitHTTP(t *testing.T, url string) {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s never answered 200", url)
}

// fakeRouter stands in for /router inside the test process. It listens on
// the LISTEN_ADDR of the environment it is launched with, answers what the
// UI server proxies, and ends on the first signal or when told to die.
type fakeRouter struct {
	env     []string
	srv     *http.Server
	exited  chan error
	signals chan os.Signal
}

func newFakeRouter() *fakeRouter {
	return &fakeRouter{exited: make(chan error, 1), signals: make(chan os.Signal, 1)}
}

func (f *fakeRouter) launch(env []string) process {
	f.env = env
	return f
}

func (f *fakeRouter) Start() error {
	ln, err := net.Listen("tcp", envValue(f.env, "LISTEN_ADDR"))
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("POST /graphql", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"__typename":"Query"}}`)
	})
	mux.HandleFunc("GET /playground", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html>playground</html>")
	})
	f.srv = &http.Server{Handler: mux, ReadHeaderTimeout: time.Second}
	go func() { _ = f.srv.Serve(ln) }()
	return nil
}

func (f *fakeRouter) Wait() error { return <-f.exited }

func (f *fakeRouter) Signal(sig os.Signal) error {
	f.signals <- sig
	_ = f.srv.Close()
	f.exited <- nil
	return nil
}

func (f *fakeRouter) die() {
	_ = f.srv.Close()
	f.exited <- errors.New("exit status 2")
}

func newSupervisor(t *testing.T) (*supervisor, *fakeRouter) {
	t.Helper()
	fake := newFakeRouter()
	s := &supervisor{
		model:  testModel,
		config: "config.json",
		addrs:  addresses{adapter: freeAddr(t), capacity: freeAddr(t), document: freeAddr(t), router: freeAddr(t), ui: freeAddr(t)},
		assets: ui.Files,
		names:  flow.Names{Quantity: "capacity", Attribute: "throughput"},
		launch: fake.launch,
		stdout: io.Discard,
	}
	return s, fake
}

// TestSR02_ReadyWithinTenSeconds starts the whole tree and times it.
func TestSR02_ReadyWithinTenSeconds(t *testing.T) {
	s, fake := newSupervisor(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	started := time.Now()
	go func() { done <- s.run(ctx) }()
	base := "http://" + s.addrs.ui
	waitHTTP(t, base+"/viewer/")
	assert.True(t, time.Since(started) < 10*time.Second, "ready within ten seconds")
	for _, u := range []string{
		"http://" + s.addrs.adapter + "/health", "http://" + s.addrs.capacity + "/health",
		"http://" + s.addrs.document + "/health", "http://" + s.addrs.router + "/health/ready", base + "/document/",
	} {
		assert.Equal(t, get(t, u).StatusCode, http.StatusOK)
	}
	assert.Equal(t, post(t, base+"/graphql", `{"query":"{ __typename }"}`), `{"data":{"__typename":"Query"}}`)
	assert.Equal(t, post(t, "http://"+s.addrs.adapter+"/graphql", `{"query":"{ model { version } }"}`), `{"data":{"model":{"version":1}}}`)
	assert.Equal(t, envValue(fake.env, "EXECUTION_CONFIG_FILE_PATH"), "config.json")
	cancel()
	assert.NoError(t, <-done)
	assert.Equal(t, <-fake.signals, os.Signal(syscall.SIGTERM))
	for _, addr := range []string{s.addrs.ui, s.addrs.adapter, s.addrs.capacity, s.addrs.document} {
		_, err := net.Dial("tcp", addr)
		assert.Error(t, err)
	}
}

func TestServeFailsWhenTheRouterDies(t *testing.T) {
	s, fake := newSupervisor(t)
	done := make(chan error, 1)
	go func() { done <- s.run(context.Background()) }()
	waitHTTP(t, "http://"+s.addrs.ui+"/viewer/")
	fake.die()
	err := <-done
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "router exited"), "the error names the router: "+err.Error())
	_, err = net.Dial("tcp", s.addrs.adapter)
	assert.Error(t, err)
}

func TestServeRefusesAModelItCannotRead(t *testing.T) {
	s, _ := newSupervisor(t)
	s.model = "missing.sysml"
	assert.Error(t, s.run(context.Background()))
}

// TestSR03_RouterEnvironmentDisablesTelemetry: the child's environment is
// the seven variables of AD-0010 and nothing inherited.
func TestSR03_RouterEnvironmentDisablesTelemetry(t *testing.T) {
	env := routerEnv("127.0.0.1:3002", "/app/config.json", "")
	assert.SliceEqual(t, env, []string{
		"LISTEN_ADDR=127.0.0.1:3002", "EXECUTION_CONFIG_FILE_PATH=/app/config.json", "PLAYGROUND_PATH=/playground",
		"DO_NOT_TRACK=1", "COSMO_TELEMETRY_DISABLED=true", "TRACING_ENABLED=false", "METRICS_OTLP_ENABLED=false",
	})
	assert.Contains(t, routerEnv("127.0.0.1:3002", "/app/config.json", "debug"), "LOG_LEVEL=debug")
	cfg := routerConfig{Binary: "/router", Config: "/app/config.json", Stdout: io.Discard, Stderr: io.Discard}
	cmd := cfg.command(env)
	assert.Equal(t, cmd.Path, "/router")
	assert.Equal(t, cmd.Dir, "/app")
	assert.SliceEqual(t, cmd.Env, env)
}

// TestRouterRunsAsAChildProcess exercises execProcess against the helper
// above: start, readiness, forwarded stdout, SIGTERM, clean exit.
func TestRouterRunsAsAChildProcess(t *testing.T) {
	var out bytes.Buffer
	cfg := routerConfig{Binary: os.Args[0], Config: filepath.Join(t.TempDir(), "config.json"), Stdout: &out, Stderr: io.Discard}
	addr := freeAddr(t)
	cmd := cfg.command(routerEnv(addr, cfg.Config, ""))
	cmd.Args = append(cmd.Args, "-test.run=^TestHelperRouter$")
	cmd.Env = append(cmd.Env, "SYSML_FEDERATION_HELPER=1")
	var p process = &execProcess{cmd: cmd}
	assert.NoError(t, p.Start())
	exited := make(chan error, 1)
	go func() { exited <- p.Wait() }()
	assert.NoError(t, waitFor(context.Background(), "http://"+addr+"/health/ready", exited))
	assert.NoError(t, p.Signal(syscall.SIGTERM))
	assert.NoError(t, <-exited)
	assert.True(t, strings.Contains(out.String(), "helper router listening"), "stdout is forwarded")
	assert.True(t, strings.Contains(out.String(), "helper router stopping"), "SIGTERM reached the child")
}

func newUI(t *testing.T, upstream http.Handler) string {
	t.Helper()
	router := httptest.NewServer(upstream)
	t.Cleanup(router.Close)
	parsed, err := url.Parse(router.URL)
	target := assert.Must(t, parsed, err)
	handler, err := uiHandler(ui.Files, target)
	h := assert.Must(t, handler, err)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv.URL
}

// TestSR04_FourPathsOnOnePort: the two apps, /graphql and /playground on
// one origin, / redirected, nothing else (SR-40: the router's other paths
// are not reachable).
func TestSR04_FourPathsOnOnePort(t *testing.T) {
	base := newUI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/graphql":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":{"__typename":"Query"}}`)
		case "/playground":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, "<html>playground</html>")
		default:
			http.NotFound(w, r)
		}
	}))
	resp := get(t, base+"/")
	assert.Equal(t, resp.StatusCode, http.StatusFound)
	assert.Equal(t, resp.Header.Get("Location"), "/viewer/")
	// The slashless form is ServeMux's own trailing-slash redirect, which is
	// a temporary one as of Go 1.27, and it lands on the app's root.
	slashless := get(t, base+"/viewer")
	assert.Equal(t, slashless.StatusCode, http.StatusTemporaryRedirect)
	assert.Equal(t, slashless.Header.Get("Location"), "/viewer/")
	for _, p := range []string{"/viewer/", "/document/"} {
		resp := get(t, base+p)
		assert.Equal(t, resp.StatusCode, http.StatusOK)
		assert.True(t, strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html"), p+" is HTML")
		assert.Equal(t, resp.Header.Get("Cache-Control"), "no-cache")
	}
	assert.Equal(t, post(t, base+"/graphql", `{"query":"{ __typename }"}`), `{"data":{"__typename":"Query"}}`)
	assert.Equal(t, get(t, base+"/playground").StatusCode, http.StatusOK)
	for _, p := range []string{"/health/ready", "/health", "/other", "/viewer/sub/"} {
		assert.Equal(t, get(t, base+p).StatusCode, http.StatusNotFound)
	}
}

// TestSR04_ProxyStreamsServerSentEvents: the first event reaches the client
// while the upstream is still holding the response open.
func TestSR04_ProxyStreamsServerSentEvents(t *testing.T) {
	release := make(chan struct{})
	base := newUI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		rc := http.NewResponseController(w)
		fmt.Fprint(w, "event: next\ndata: {\"data\":{\"modelChanged\":2}}\n\n")
		assert.NoError(t, rc.Flush())
		<-release
		fmt.Fprint(w, "event: next\ndata: {\"data\":{\"modelChanged\":3}}\n\n")
		assert.NoError(t, rc.Flush())
	}))
	built, err := http.NewRequest(http.MethodPost, base+"/graphql", strings.NewReader(`{"query":"subscription { modelChanged }"}`))
	req := assert.Must(t, built, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	sent, err := http.DefaultClient.Do(req) //nolint:bodyclose // closed below, once assert.Must has proved there is one
	resp := assert.Must(t, sent, err)       //nolint:bodyclose // the deferred close on the next line is this one
	defer resp.Body.Close()
	frames := make(chan string, 2)
	go func() {
		r := bufio.NewReader(resp.Body)
		for range 2 {
			var frame strings.Builder
			for {
				line, err := r.ReadString('\n')
				if err != nil {
					frames <- "read error: " + err.Error()
					return
				}
				if line == "\n" {
					break
				}
				frame.WriteString(line)
			}
			frames <- frame.String()
		}
	}()
	next := func() string {
		select {
		case f := <-frames:
			return f
		case <-time.After(5 * time.Second):
			t.Fatal("no frame within five seconds: the proxy is buffering")
			return ""
		}
	}
	assert.Equal(t, next(), "event: next\ndata: {\"data\":{\"modelChanged\":2}}\n")
	close(release)
	assert.Equal(t, next(), "event: next\ndata: {\"data\":{\"modelChanged\":3}}\n")
}

func TestHealthcheckFailsWhenEitherProbeFails(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	t.Cleanup(ok.Close)
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	t.Cleanup(bad.Close)
	ctx := context.Background()
	assert.NoError(t, healthcheck(ctx, ok.URL+"/health/ready", ok.URL+"/viewer/"))
	assert.Error(t, healthcheck(ctx, ok.URL+"/health/ready", bad.URL+"/viewer/"))
	assert.Error(t, healthcheck(ctx, bad.URL+"/health/ready", ok.URL+"/viewer/"))
	assert.Error(t, healthcheck(ctx, "http://127.0.0.1:1/health/ready"))
}

func TestRunDispatchesSubcommandsAndExitCodes(t *testing.T) {
	ctx := context.Background()
	var out, errw bytes.Buffer
	assert.Equal(t, run(ctx, []string{"nope"}, &out, &errw), 2)
	assert.True(t, strings.Contains(errw.String(), "usage"), "an unknown command prints the usage")
	assert.Equal(t, run(ctx, []string{"help"}, &out, &errw), 0)
	assert.Equal(t, run(ctx, []string{"adapter", "-h"}, &out, &errw), 0)
	assert.Equal(t, run(ctx, []string{"adapter", "-model", "missing.sysml", "-addr", "127.0.0.1:0"}, &out, &errw), 1)
	assert.Equal(t, run(ctx, []string{"healthcheck", "extra"}, &out, &errw), 1)
}

func TestSubcommandsServeOneComponent(t *testing.T) {
	for _, name := range []string{"adapter", "capacity", "document"} {
		t.Run(name, func(t *testing.T) {
			addr := freeAddr(t)
			args := []string{name, "-addr", addr}
			if name == "adapter" {
				args = append(args, "-model", testModel)
			}
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan int, 1)
			go func() { done <- run(ctx, args, io.Discard, io.Discard) }()
			waitHTTP(t, "http://"+addr+"/health")
			cancel()
			assert.Equal(t, <-done, 0)
		})
	}
}
