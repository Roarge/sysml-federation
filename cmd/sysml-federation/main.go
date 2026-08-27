// Command sysml-federation is the demo's one binary. Its default subcommand,
// serve, is the supervisor of architecture V4: the three subgraphs as
// goroutines on loopback, the Cosmo router as a child process and the UI
// server on the published port. The other subcommands run one component
// each, and healthcheck is the image's HEALTHCHECK (AD-0011).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Roarge/sysml-federation/adapter/projection"
	"github.com/Roarge/sysml-federation/adapter/serve"
	"github.com/Roarge/sysml-federation/examples/pipeline/capacity"
	"github.com/Roarge/sysml-federation/examples/pipeline/capacity/flow"
	"github.com/Roarge/sysml-federation/examples/pipeline/document"
	"github.com/Roarge/sysml-federation/examples/pipeline/ui"
)

// The subgraphs and the router listen on loopback: only the router talks to
// the subgraphs and only the UI server talks to the router, so nothing but
// the published port is reachable from outside the container (SR-03,
// SR-04). The subgraph ports are the ones the committed config.json routes
// to (AD-0012) and the router port is the one healthcheck probes, so they
// are constants rather than flags.
const (
	adapterAddr  = "127.0.0.1:3011"
	capacityAddr = "127.0.0.1:3012"
	documentAddr = "127.0.0.1:3013"
	routerAddr   = "127.0.0.1:3002"
	uiAddr       = "0.0.0.0:8080"
	uiProbeAddr  = "127.0.0.1:8080"

	defaultModelPath = "/app/model.sysml"
)

// exampleNames are the capacity service's two configured names (SR-31).
var exampleNames = flow.Names{Quantity: "capacity", Attribute: "throughput"}

// defaultAddresses is the process tree of architecture V4.
var defaultAddresses = addresses{adapter: adapterAddr, capacity: capacityAddr, document: documentAddr, router: routerAddr, ui: uiAddr}

// command is one subcommand. It runs until ctx ends or it fails.
type command func(ctx context.Context, args []string, stdout io.Writer) error

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	// Released here rather than by defer, because os.Exit runs no deferred call.
	stop()
	os.Exit(code)
}

// run dispatches to a subcommand and maps its outcome to an exit status:
// 0 for success or help, 1 for a failure, 2 for an unknown command.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	commands := map[string]command{
		"serve": runServe, "adapter": runAdapter, "capacity": runCapacity,
		"document": runDocument, "ui": runUI, "healthcheck": runHealthcheck,
	}
	name, rest := "serve", args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name, rest = args[0], args[1:]
	}
	if name == "help" {
		usage(stdout)
		return 0
	}
	cmd, ok := commands[name]
	if !ok {
		_, _ = fmt.Fprintf(stderr, "unknown command %q\n", name)
		usage(stderr)
		return 2
	}
	err := cmd(ctx, rest, stdout)
	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "sysml-federation %s: %v\n", name, err)
		return 1
	}
}

func usage(w io.Writer) {
	_, _ = fmt.Fprint(w, `usage: sysml-federation [command] [flags]

  serve        the whole demo: three subgraphs, the router, the UI server (default; -model)
  adapter      the adapter subgraph alone (-model, -addr)
  capacity     the capacity subgraph alone (-addr)
  document     the document subgraph alone (-addr)
  ui           the UI server alone (-addr, -router)
  healthcheck  probe the router and the UI server, exit 1 if either fails
  help         print this text

The router child is /router, or SYSML_FEDERATION_ROUTER. Its configuration
is /app/config.json, or SYSML_FEDERATION_CONFIG. LOG_LEVEL is passed on.
`)
}

// flags builds a subcommand's flag set, printing help to stdout.
func flags(name string, stdout io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stdout)
	return fs
}

func runServe(ctx context.Context, args []string, stdout io.Writer) error {
	fs := flags("serve", stdout)
	modelPath := fs.String("model", defaultModelPath, "the SysML v2 model file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg := routerFromEnv(stdout, os.Stderr)
	s := &supervisor{model: *modelPath, config: cfg.Config, logLevel: cfg.LogLevel, addrs: defaultAddresses,
		assets: ui.Files, names: exampleNames, launch: cfg.launch, stdout: stdout}
	return s.run(ctx)
}

func runAdapter(ctx context.Context, args []string, stdout io.Writer) error {
	fs := flags("adapter", stdout)
	modelPath := fs.String("model", defaultModelPath, "the SysML v2 model file")
	addr := fs.String("addr", adapterAddr, "address to listen on")
	if err := fs.Parse(args); err != nil {
		return err
	}
	store, err := projection.Load(*modelPath)
	if err != nil {
		return err
	}
	return serveOne(ctx, *addr, serve.Handler(store), stdout)
}

func runCapacity(ctx context.Context, args []string, stdout io.Writer) error {
	fs := flags("capacity", stdout)
	addr := fs.String("addr", capacityAddr, "address to listen on")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return serveOne(ctx, *addr, capacity.Handler(exampleNames), stdout)
}

func runDocument(ctx context.Context, args []string, stdout io.Writer) error {
	fs := flags("document", stdout)
	addr := fs.String("addr", documentAddr, "address to listen on")
	if err := fs.Parse(args); err != nil {
		return err
	}
	svc, err := document.New()
	if err != nil {
		return err
	}
	return serveOne(ctx, *addr, document.Handler(svc), stdout)
}

func runUI(ctx context.Context, args []string, stdout io.Writer) error {
	fs := flags("ui", stdout)
	addr := fs.String("addr", uiAddr, "address to listen on")
	router := fs.String("router", "http://"+routerAddr, "the router's base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	target, err := url.Parse(*router)
	if err != nil {
		return err
	}
	h, err := uiHandler(ui.Files, target)
	if err != nil {
		return err
	}
	return serveOne(ctx, *addr, h, stdout)
}

func runHealthcheck(ctx context.Context, args []string, _ io.Writer) error {
	if len(args) > 0 {
		return errors.New("healthcheck takes no arguments")
	}
	return healthcheck(ctx, "http://"+routerAddr+"/health/ready", "http://"+uiProbeAddr+"/viewer/")
}

// serveOne runs one handler on one address until ctx ends or it fails.
func serveOne(ctx context.Context, addr string, h http.Handler, stdout io.Writer) error {
	failed := make(chan error, 1)
	s, err := listen("server", addr, h, failed)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "listening on http://%s/\n", s.addr)
	select {
	case <-ctx.Done():
		s.stop()
		return nil
	case err := <-failed:
		return err
	}
}
