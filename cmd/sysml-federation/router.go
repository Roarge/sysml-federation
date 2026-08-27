package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	defaultRouterBinary = "/router"
	defaultRouterConfig = "/app/config.json"
)

// process is what the supervisor needs from the router child. exec.Cmd
// satisfies it through execProcess, and the tests substitute an in-process
// fake.
type process interface {
	Start() error
	Wait() error
	Signal(sig os.Signal) error
}

// launcher builds the child from the environment it is to run with.
type launcher func(env []string) process

// routerConfig is where the binary and its configuration are, and where its
// output goes. Both paths come from the environment so a developer can point
// at a locally extracted router.
type routerConfig struct {
	Binary, Config, LogLevel string
	Stdout, Stderr           io.Writer
}

// routerFromEnv reads SYSML_FEDERATION_ROUTER, SYSML_FEDERATION_CONFIG and
// LOG_LEVEL, with the image's paths as defaults.
func routerFromEnv(stdout, stderr io.Writer) routerConfig {
	cfg := routerConfig{Binary: defaultRouterBinary, Config: defaultRouterConfig, LogLevel: os.Getenv("LOG_LEVEL"), Stdout: stdout, Stderr: stderr}
	if v := os.Getenv("SYSML_FEDERATION_ROUTER"); v != "" {
		cfg.Binary = v
	}
	if v := os.Getenv("SYSML_FEDERATION_CONFIG"); v != "" {
		cfg.Config = v
	}
	return cfg
}

// routerEnv is the child's whole environment (AD-0010, SR-03): the listen
// address, the static configuration, the playground path, the four
// variables that keep the router off the network, and LOG_LEVEL when set.
// Nothing of the supervisor's environment is inherited.
func routerEnv(listen, config, logLevel string) []string {
	env := []string{
		"LISTEN_ADDR=" + listen,
		"EXECUTION_CONFIG_FILE_PATH=" + config,
		"PLAYGROUND_PATH=/playground",
		"DO_NOT_TRACK=1",
		"COSMO_TELEMETRY_DISABLED=true",
		"TRACING_ENABLED=false",
		"METRICS_OTLP_ENABLED=false",
	}
	if logLevel != "" {
		env = append(env, "LOG_LEVEL="+logLevel)
	}
	return env
}

// command builds the exec.Cmd. The working directory is the configuration's
// directory, where no config.yaml exists, so the router runs from
// environment alone (C-20). Its output is forwarded line for line.
func (c routerConfig) command(env []string) *exec.Cmd {
	cmd := exec.Command(c.Binary)
	cmd.Env = env
	cmd.Dir = filepath.Dir(c.Config)
	cmd.Stdout, cmd.Stderr = c.Stdout, c.Stderr
	return cmd
}

func (c routerConfig) launch(env []string) process { return &execProcess{cmd: c.command(env)} }

// execProcess is a process over exec.Cmd.
type execProcess struct{ cmd *exec.Cmd }

func (p *execProcess) Start() error               { return p.cmd.Start() }
func (p *execProcess) Wait() error                { return p.cmd.Wait() }
func (p *execProcess) Signal(sig os.Signal) error { return p.cmd.Process.Signal(sig) }
