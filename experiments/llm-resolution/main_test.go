package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEXPSR13_AMissingModelStopsTheRunBeforeAnyQuestion(t *testing.T) {
	f := newFakeServer(t)
	f.models = []string{"llama3:8b"}
	dir := t.TempDir()
	code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", fixtureRoot(t), "-out", dir)
	if code != 2 {
		t.Fatalf("exit %d, want 2 (stderr %q)", code, stderr)
	}
	if !strings.Contains(stderr, "qwen2.5-coder:14b") {
		t.Errorf("the message doesn't name the missing model: %q", stderr)
	}
	if n := len(f.chats()); n != 0 {
		t.Errorf("%d questions asked", n)
	}
	if files, _ := filepath.Glob(filepath.Join(dir, "*")); len(files) != 0 {
		t.Errorf("files left behind: %v", files)
	}
	if code, _, _ := runExperiment(t, "-repo", fixtureRoot(t), "-out", dir); code != 2 {
		t.Errorf("no -url: exit %d, want 2", code)
	}
}

func TestEXPSR15_TheReportIsRebuiltFromAFile(t *testing.T) {
	f := newFakeServer(t)
	dir := t.TempDir()
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", fixtureRoot(t), "-out", dir); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	results := resultsFile(t, dir)
	written, err := os.ReadFile(strings.TrimSuffix(results, ".jsonl") + ".md")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	code, stdout, stderr := runExperiment(t, "-report", results)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if stdout != string(written) || len(written) == 0 {
		t.Errorf("the rebuilt report differs from the one the run wrote:\n%s\n---\n%s", stdout, written)
	}
}
