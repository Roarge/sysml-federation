package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// fixtureRoot copies testdata/repo to a temporary folder, adds the hidden
// folder git can't carry, and returns the copy's root.
func fixtureRoot(t *testing.T) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.WalkDir("testdata/repo", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("testdata/repo", path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	hidden := filepath.Join(dst, ".hidden")
	if err := os.MkdirAll(hidden, 0o755); err != nil {
		t.Fatal(err)
	}
	src := "package hidden\n\nimport \"testing\"\n\nfunc TestSR01_InAHiddenFolderIsSkipped(t *testing.T) {}\n"
	if err := os.WriteFile(filepath.Join(hidden, "hidden_test.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return dst
}

// repoRoot is the checkout this module sits in.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "model", "core")); err != nil {
		t.Skip("not inside the sysml-federation checkout")
	}
	return root
}

// ---- a stand-in for the Ollama server ------------------------------------------

type capturedMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type capturedOptions struct {
	NumCtx      int     `json:"num_ctx"`
	NumPredict  int     `json:"num_predict"`
	Temperature float64 `json:"temperature"`
	Seed        int     `json:"seed"`
}

type capturedRequest struct {
	Model     string            `json:"model"`
	Messages  []capturedMessage `json:"messages"`
	Stream    bool              `json:"stream"`
	Format    answerSchema      `json:"format"`
	Options   capturedOptions   `json:"options"`
	KeepAlive string            `json:"keep_alive"`
}

type fakeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type fakeChatReply struct {
	Model           string      `json:"model"`
	Message         fakeMessage `json:"message"`
	Done            bool        `json:"done"`
	TotalDuration   int64       `json:"total_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
	EvalCount       int         `json:"eval_count"`
}

type fakeTagDetails struct {
	ParameterSize     string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

type fakeTag struct {
	Name    string         `json:"name"`
	Digest  string         `json:"digest"`
	Details fakeTagDetails `json:"details"`
}

type fakeTags struct {
	Models []fakeTag `json:"models"`
}

type fakeVersion struct {
	Version string `json:"version"`
}

type fakeServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []capturedRequest
	models   []string
	// reply returns the message content for one question.
	reply func(system, user string) string
	// before runs before each chat reply, with the number of chats so far.
	before func(n int)
}

func newFakeServer(t *testing.T) *fakeServer {
	t.Helper()
	f := &fakeServer{models: []string{"qwen2.5-coder:14b"}, reply: fixturePolicy}
	f.Server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.Close)
	return f
}

func (f *fakeServer) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.URL.Path {
	case "/api/version":
		_ = json.NewEncoder(w).Encode(fakeVersion{Version: "0.12.99-fake"})
	case "/api/tags":
		var tags fakeTags
		for _, m := range f.models {
			tags.Models = append(tags.Models, fakeTag{Name: m, Digest: "sha256:fakedigest", Details: fakeTagDetails{ParameterSize: "14.8B", QuantizationLevel: "Q4_K_M"}})
		}
		_ = json.NewEncoder(w).Encode(tags)
	case "/api/chat":
		var req capturedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		n := len(f.requests)
		f.requests = append(f.requests, req)
		f.mu.Unlock()
		if f.before != nil {
			f.before(n)
		}
		system, user := "", ""
		for _, m := range req.Messages {
			switch m.Role {
			case "system":
				system = m.Content
			case "user":
				user = m.Content
			}
		}
		_ = json.NewEncoder(w).Encode(fakeChatReply{
			Model: req.Model, Message: fakeMessage{Role: "assistant", Content: f.reply(system, user)},
			Done: true, TotalDuration: 1_000_000, PromptEvalCount: 100, EvalCount: 20,
		})
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeServer) chats() []capturedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]capturedRequest{}, f.requests...)
}

var nameLineRE = regexp.MustCompile(`(?m)^name: (.*)$`)

// fixtureAnswers are what the stand-in answers for the fixture's tests,
// looked up by the name the question shows. A name it doesn't know, such as
// one with words deleted, gets "none".
var fixtureAnswers = map[string]Answer{
	"RejectsAnEmptyQuery":           {Requirement: "SR-01", Evidence: []string{"empty query"}, Reason: "It rejects an empty query."},
	"ParsesTokens":                  {Requirement: "SR-01", Evidence: []string{"tokens"}, Reason: "It parses tokens."},
	"ReturnsARankedPage":            {Requirement: "SR-02", Evidence: []string{"ranked page"}, Reason: "It returns a ranked page."},
	"ReportsLatency":                {Requirement: "SR-03", Evidence: []string{"latency"}, Reason: "It reports latency."},
	"ImportsOnlyTheStandardLibrary": {Requirement: "SC-01", Evidence: []string{"standard library"}, Reason: "It checks the imports."},
	"ServesTopResults":              {Requirement: "SR-02", Evidence: []string{"top results"}, Reason: "It serves the top results."},
	"Health":                        {Requirement: "SR-03", Evidence: []string{"page"}, Reason: "A health check reads a page."},
}

func fixturePolicy(system, user string) string {
	a := Answer{Requirement: "none", Evidence: []string{}, Reason: "Nothing fits."}
	if m := nameLineRE.FindStringSubmatch(user); m != nil {
		if known, ok := fixtureAnswers[strings.TrimSpace(m[1])]; ok {
			a = known
		}
	}
	data, _ := json.Marshal(a)
	return string(data)
}

// runExperiment runs the program with the given arguments and returns its
// exit status and what it printed.
func runExperiment(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// resultsFile finds the one results file a run wrote under dir.
func resultsFile(t *testing.T, dir string) string {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if len(files) != 1 {
		t.Fatalf("want one results file under %s, found %v", dir, files)
	}
	return files[0]
}

// callsWithProbe returns the call lines of one probe.
func callsWithProbe(c Contents, probe string) []CallLine {
	var out []CallLine
	for _, cl := range c.Calls {
		if cl.Probe == probe {
			out = append(out, cl)
		}
	}
	return out
}
