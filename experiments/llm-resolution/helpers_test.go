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
	src := "package hidden\n\nimport \"testing\"\n\n// quokka\nfunc TestSR01_InAHiddenFolderIsSkipped(t *testing.T) {}\n"
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

// fixtureKey is the incident key written for the fixture tree.
const fixtureKey = "testdata/fixture-incident-key.json"

// fixtureWiki reads the fixture tree's systems model.
func fixtureWiki(t *testing.T) (*Wiki, string) {
	t.Helper()
	root := fixtureRoot(t)
	w, err := LoadWiki(root)
	if err != nil {
		t.Fatal(err)
	}
	return w, root
}

// realWiki reads the checkout's systems model once per test binary.
var (
	realOnce  sync.Once
	realModel *Wiki
	realErr   error
)

func realWiki(t *testing.T) *Wiki {
	t.Helper()
	root := repoRoot(t)
	realOnce.Do(func() { realModel, realErr = LoadWiki(root) })
	if realErr != nil {
		t.Fatal(realErr)
	}
	return realModel
}

// fixtureKit gives the tools for one task over the fixture tree.
func fixtureKit(t *testing.T, task string, sc Scope) (*ToolKit, *Wiki, string) {
	t.Helper()
	w, root := fixtureWiki(t)
	code, err := LoadCode(root)
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := LoadCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	sc.Task = task
	return NewToolKit(w, code, corpus.Requirements, sc), w, root
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
	Format    json.RawMessage   `json:"format"`
	Options   capturedOptions   `json:"options"`
	KeepAlive string            `json:"keep_alive"`
}

func (r capturedRequest) system() string { return r.message("system") }
func (r capturedRequest) user() string   { return r.message("user") }

func (r capturedRequest) message(role string) string {
	for _, m := range r.Messages {
		if m.Role == role {
			return m.Content
		}
	}
	return ""
}

// schemaProperties are the property names of a reply schema as sent.
type schemaProperties struct {
	Properties map[string]json.RawMessage `json:"properties"`
}

func (r capturedRequest) asks(property string) bool {
	var s schemaProperties
	if json.Unmarshal(r.Format, &s) != nil {
		return false
	}
	_, ok := s.Properties[property]
	return ok
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
	reply func(q capturedRequest) string
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
		_ = json.NewEncoder(w).Encode(fakeChatReply{
			Model: req.Model, Message: fakeMessage{Role: "assistant", Content: f.reply(req)},
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

// fixtureLink is what the stand-in links a fixture test to, by the name the
// task shows. A name it doesn't know, such as one with words deleted, is
// linked to nothing.
type fixtureLink struct {
	id       string
	evidence []string
}

var fixtureLinks = map[string]fixtureLink{
	"RejectsAnEmptyQuery":           {"SR-01", []string{"empty query"}},
	"ParsesTokens":                  {"SR-01", []string{"tokens"}},
	"ReturnsARankedPage":            {"SR-02", []string{"ranked page"}},
	"ReportsLatency":                {"SR-03", []string{"latency"}},
	"ImportsOnlyTheStandardLibrary": {"SC-01", []string{"standard library"}},
	"ServesTopResults":              {"SR-02", []string{"top results"}},
	"Health":                        {"SR-03", []string{"page"}},
}

// taskName is the test name a question's task shows, or "" for the incident.
func taskName(user string) string {
	if m := nameLineRE.FindStringSubmatch(user); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func stepJSON(s Step) string {
	data, _ := json.Marshal(s)
	return string(data)
}

func testAnswerJSON(a TestAnswer) string {
	data, _ := json.Marshal(a)
	return string(data)
}

func incidentAnswerJSON(a IncidentAnswer) string {
	data, _ := json.Marshal(a)
	return string(data)
}

// fixturePolicy browses every task in two steps and then answers. A test's
// first step searches the requirements for its name and the second exposes
// the requirement it knows and stops. The incident's first step reads the
// links of the server's state machine and the second exposes it and stops.
func fixturePolicy(q capturedRequest) string {
	user := q.user()
	name := taskName(user)
	incident := name == "" && !strings.Contains(user, "name: ")
	known, ok := fixtureLinks[name]
	switch {
	case q.asks("tool"):
		s := Step{Viewpoint: "Which requirement does this test verify?", Expose: []Exposure{}, Prune: []string{}}
		if incident {
			s.Viewpoint = "Why is the query page down?"
		}
		switch {
		case strings.Contains(user, "Last call: none") && incident:
			s.Tool, s.Arg = "links", "ServerStates"
		case strings.Contains(user, "Last call: none"):
			s.Tool, s.Arg, s.Arg2 = "find", name, "requirement"
		case incident:
			s.Tool = "done"
			s.Expose = []Exposure{{ID: "ServerStates", Note: "stops the server when the parser exits"}}
		default:
			s.Tool = "done"
			if ok {
				s.Expose = []Exposure{{ID: known.id, Note: "the requirement it verifies"}}
			}
		}
		return stepJSON(s)
	case q.asks("links"):
		a := TestAnswer{Links: []LinkAnswer{}, Evidence: []string{}, Reason: "Nothing fits.", Mismatches: []Mismatch{}}
		if ok {
			a.Links = []LinkAnswer{{ID: known.id, Relation: "verifies"}}
			a.Evidence = known.evidence
			a.Reason = "The test checks what " + known.id + " asks for."
		}
		return testAnswerJSON(a)
	case q.asks("cause"):
		return incidentAnswerJSON(IncidentAnswer{
			Cause:        []string{"system::parser"},
			Mechanism:    []string{"ServerStates::onParserExit"},
			Code:         []string{"adapter/serve/serve.go"},
			Consequences: []string{"SR-01", "SR-02"},
			Path:         []string{"SR-02", "system::server", "Server", "ServerStates", "ServerStates::onParserExit"},
			Why:          "The parser exited and the server's state machine stops the server when it does.",
			Mismatches:   []Mismatch{},
		})
	}
	return "{}"
}

// runExperiment runs the program with the given arguments and returns its
// exit status and what it printed.
func runExperiment(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// runFixture runs the program against the stand-in on a fixture tree.
func runFixture(t *testing.T, f *fakeServer, root, out string, extra ...string) (int, string, string) {
	t.Helper()
	args := append([]string{"-url", f.URL, "-repo", root, "-out", out, "-key", fixtureKey}, extra...)
	return runExperiment(t, args...)
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

// readRun runs a full fixture run and reads its results file back.
func readRun(t *testing.T, f *fakeServer, extra ...string) Contents {
	t.Helper()
	dir := t.TempDir()
	if code, _, stderr := runFixture(t, f, fixtureRoot(t), dir, extra...); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	c, err := ReadResults(resultsFile(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// finals returns the final lines of one probe, one per browse or note.
func finals(c Contents, probe string) []CallLine {
	var out []CallLine
	for _, cl := range c.Calls {
		if cl.Probe == probe && cl.Final {
			out = append(out, cl)
		}
	}
	return out
}
