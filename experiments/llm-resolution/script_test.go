package main

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEXPSR13_AnUnreachableServerStopsTheScript(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("no bash")
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := "http://" + l.Addr().String()
	l.Close()

	out := t.TempDir()
	for _, url := range []string{closed, ""} {
		cmd := exec.Command("bash", "run.sh", "-out", out)
		cmd.Env = append(os.Environ(), "OLLAMA_URL="+url)
		stderr, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 2 {
			t.Errorf("OLLAMA_URL=%q: %v, want exit status 2\n%s", url, err, stderr)
			continue
		}
		if url == "" && !strings.Contains(string(stderr), "OLLAMA_URL") {
			t.Errorf("no address: the message doesn't say what to set:\n%s", stderr)
		}
	}
	if files, _ := filepath.Glob(filepath.Join(out, "*")); len(files) != 0 {
		t.Errorf("files left behind: %v", files)
	}
}
