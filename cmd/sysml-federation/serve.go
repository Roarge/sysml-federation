package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/Roarge/sysml-federation/adapter/projection"
	"github.com/Roarge/sysml-federation/adapter/serve"
	"github.com/Roarge/sysml-federation/examples/pipeline/capacity"
	"github.com/Roarge/sysml-federation/examples/pipeline/capacity/flow"
	"github.com/Roarge/sysml-federation/examples/pipeline/document"
)

const (
	// readyTimeout bounds each wait for a component's health endpoint. The
	// whole start is SR-02's ten seconds, measured by the supervisor test.
	readyTimeout = 10 * time.Second
	// stopTimeout bounds a graceful stop of a server and of the router.
	stopTimeout  = 5 * time.Second
	pollInterval = 100 * time.Millisecond
)

// errExited reports that the component being waited for ended first.
var errExited = errors.New("exited before it was ready")

// addresses are the five listen addresses of the process tree.
type addresses struct{ adapter, capacity, document, router, ui string }

// supervisor is the serve subcommand: it starts the components in order,
// watches them, and stops them in reverse order (architecture V3, Startup).
type supervisor struct {
	model    string
	config   string
	logLevel string
	addrs    addresses
	assets   fs.FS
	names    flow.Names
	launch   launcher
	stdout   io.Writer
}

// run returns when ctx ends (nil) or when any component fails (its error),
// with everything stopped either way.
func (s *supervisor) run(ctx context.Context) error {
	store, err := projection.Load(s.model)
	if err != nil {
		return err
	}
	doc, err := document.New()
	if err != nil {
		return err
	}
	failed := make(chan error, 4)
	var subgraphs []*server
	defer func() {
		for i := len(subgraphs) - 1; i >= 0; i-- {
			subgraphs[i].stop()
		}
	}()
	for _, c := range []struct {
		name, addr string
		h          http.Handler
	}{
		{"adapter", s.addrs.adapter, serve.Handler(store)},
		{"capacity", s.addrs.capacity, capacity.Handler(s.names)},
		{"document", s.addrs.document, document.Handler(doc)},
	} {
		srv, err := listen(c.name, c.addr, c.h, failed)
		if err != nil {
			return err
		}
		subgraphs = append(subgraphs, srv)
		if err := waitFor(ctx, "http://"+srv.addr+"/health", failed); err != nil {
			return fmt.Errorf("%s: %w", c.name, err)
		}
		_, _ = fmt.Fprintf(s.stdout, "%s subgraph on http://%s/graphql\n", c.name, srv.addr)
	}

	router := s.launch(routerEnv(s.addrs.router, s.config, s.logLevel))
	if err := router.Start(); err != nil {
		return fmt.Errorf("router: %w", err)
	}
	exited := make(chan error, 1)
	go func() { exited <- router.Wait() }()
	if err := waitFor(ctx, "http://"+s.addrs.router+"/health/ready", exited); err != nil {
		if !errors.Is(err, errExited) {
			stopRouter(router, exited)
		}
		return fmt.Errorf("router: %w", err)
	}
	_, _ = fmt.Fprintf(s.stdout, "router on http://%s/graphql\n", s.addrs.router)

	target, err := url.Parse("http://" + s.addrs.router)
	if err != nil {
		stopRouter(router, exited)
		return err
	}
	h, err := uiHandler(s.assets, target)
	if err != nil {
		stopRouter(router, exited)
		return err
	}
	web, err := listen("ui", s.addrs.ui, h, failed)
	if err != nil {
		stopRouter(router, exited)
		return err
	}
	_, _ = fmt.Fprintf(s.stdout, "ready on http://%s/\n", web.addr)

	var failure error
	routerGone := false
	select {
	case <-ctx.Done():
		_, _ = fmt.Fprintln(s.stdout, "stopping")
	case err := <-failed:
		failure = err
	case err := <-exited:
		failure = fmt.Errorf("router exited: %w", exitStatus(err))
		routerGone = true
	}
	web.stop()
	if !routerGone {
		stopRouter(router, exited)
	}
	return failure // the deferred loop stops the subgraphs last
}

// server is one listening http.Server.
type server struct {
	addr string
	srv  *http.Server
}

// listen binds the address and serves h. A failure other than a stop is
// reported on failed, prefixed with the component's name.
func listen(name, addr string, h http.Handler, failed chan<- error) (*server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	s := &server{addr: ln.Addr().String(), srv: &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second}}
	go func() {
		if err := s.srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
			failed <- fmt.Errorf("%s: %w", name, err)
		}
	}()
	return s, nil
}

// stop drains the server within stopTimeout, then closes what is left. A
// subscription over SSE is a response that never ends, so the close is
// what actually ends it.
func (s *server) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		_ = s.srv.Close()
	}
}

// waitFor polls url until it answers 200. It gives up when ctx ends, when
// the component reports on failed, or after readyTimeout.
func waitFor(ctx context.Context, url string, failed <-chan error) error {
	client := &http.Client{Timeout: time.Second}
	deadline := time.NewTimer(readyTimeout)
	defer deadline.Stop()
	for {
		if answers(ctx, client, url) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-failed:
			if err != nil {
				return fmt.Errorf("%w: %w", errExited, err)
			}
			return errExited
		case <-deadline.C:
			return fmt.Errorf("%s did not answer within %s", url, readyTimeout)
		case <-time.After(pollInterval):
		}
	}
}

func answers(ctx context.Context, client *http.Client, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// stopRouter asks the child to stop and gives it stopTimeout before killing it.
func stopRouter(p process, exited <-chan error) {
	_ = p.Signal(syscall.SIGTERM)
	select {
	case <-exited:
	case <-time.After(stopTimeout):
		_ = p.Signal(os.Kill)
		<-exited
	}
}

// exitStatus names a clean exit, which is still a failure of a child that
// was meant to run forever.
func exitStatus(err error) error {
	if err == nil {
		return errors.New("exit status 0")
	}
	return err
}
