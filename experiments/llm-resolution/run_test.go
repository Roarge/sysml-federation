package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// The fixture has eight tests, and the stand-in browses every task in two
// tool calls and a final question, three questions a browse. A full run
// browses the eight tests, the incident five ways, the four probes on the
// two tests of the probe sample, and two repeats: 8 + 5 + 8 + 2 = 23
// browses, 69 questions.
const fixtureQuestions = 69

func TestEXPSR05_NothingToDeleteIsRecordedAndNotAsked(t *testing.T) {
	f := newFakeServer(t)
	f.reply = func(q capturedRequest) string {
		if q.asks("links") {
			if known, ok := fixtureLinks[taskName(q.user())]; ok {
				return testAnswerJSON(TestAnswer{Links: []LinkAnswer{{ID: known.id, Relation: "verifies"}}, Evidence: []string{"zzzz"}, Mismatches: []Mismatch{}})
			}
		}
		return fixturePolicy(q)
	}
	c := readRun(t, f)
	for _, probe := range []string{ProbeDeletion, ProbeControl, ProbeRareShared} {
		lines := finals(c, probe)
		if len(lines) != 2 {
			t.Fatalf("%d %s lines, want one for each of the two sampled tests", len(lines), probe)
		}
		for _, cl := range lines {
			if cl.Note == "" || cl.Reply != nil || cl.User != "" {
				t.Errorf("%s for %s: note %q, reply %v, question %q", probe, cl.Test, cl.Note, cl.Reply, cl.User)
			}
		}
	}
	for _, cl := range c.Calls {
		if (cl.Probe == ProbeDeletion || cl.Probe == ProbeControl || cl.Probe == ProbeRareShared) && !cl.Final {
			t.Errorf("a %s question was asked: %s", cl.Probe, cl.Key)
		}
	}
	if got, want := len(f.chats()), fixtureQuestions-18; got != want {
		t.Errorf("%d questions asked, want %d", got, want)
	}
}

func TestEXPSR05_TheProbeSampleIsEveryFourthLinkedTest(t *testing.T) {
	var tests []Test
	picks := map[string]string{}
	for i := 0; i < 12; i++ {
		tt := Test{ID: fmt.Sprint(i)}
		tests = append(tests, tt)
		if i%3 != 2 { // every third test is linked to nothing
			picks[tt.ID] = "SR-01"
		} else {
			picks[tt.ID] = "none"
		}
	}
	var ids []string
	for _, tt := range probeSample(tests, picks) {
		ids = append(ids, tt.ID)
	}
	// Linked, in order: 0 1 3 4 6 7 9 10. The first and every fourth after it.
	if strings.Join(ids, ",") != "0,6" {
		t.Fatalf("probe sample = %v, want 0,6", ids)
	}

	c := readRun(t, newFakeServer(t))
	var probed []string
	for _, cl := range finals(c, ProbeReconstruction) {
		probed = append(probed, cl.Test)
	}
	want := []string{"adapter/parse/parse_test.go#TestSR01_ParsesTokens", "adapter/serve/serve_test.go#TestSR02_ReturnsARankedPage"}
	if !reflect.DeepEqual(probed, want) {
		t.Errorf("the fixture run probed %v, want %v", probed, want)
	}
}

func TestEXPSR08_EveryFourthTestIsAskedTwice(t *testing.T) {
	var tests []Test
	for i := 0; i < 9; i++ {
		tests = append(tests, Test{ID: fmt.Sprint(i)})
	}
	var ids []string
	for _, tt := range repeatSample(tests) {
		ids = append(ids, tt.ID)
	}
	if strings.Join(ids, ",") != "0,4,8" {
		t.Fatalf("repeat sample = %v, want 0,4,8", ids)
	}

	c := readRun(t, newFakeServer(t))
	var got []string
	for _, cl := range finals(c, ProbeRepeat) {
		got = append(got, cl.Test)
	}
	want := []string{"adapter/parse/parse_test.go#TestHelperBuildsAQuery", "adapter/serve/serve_test.go#TestHealth"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("repeated %v, want %v", got, want)
	}
}

// replyLines counts the call lines with a reply in a results file.
func replyLines(t *testing.T, path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4<<20), 4<<20)
	for sc.Scan() {
		var cl CallLine
		if json.Unmarshal(sc.Bytes(), &cl) == nil && cl.Type == "call" && cl.Reply != nil {
			n++
		}
	}
	return n
}

func TestEXPSR09_EachReplyIsWrittenAsItArrives(t *testing.T) {
	f := newFakeServer(t)
	dir := t.TempDir()
	var mu sync.Mutex
	f.before = func(n int) {
		mu.Lock()
		defer mu.Unlock()
		files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		if len(files) != 1 {
			t.Errorf("before question %d there are %d results files", n+1, len(files))
			return
		}
		if got := replyLines(t, files[0]); got != n {
			t.Errorf("before question %d the file holds %d replies, want %d", n+1, got, n)
		}
	}
	if code, _, stderr := runFixture(t, f, fixtureRoot(t), dir); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if got := replyLines(t, resultsFile(t, dir)); got != fixtureQuestions {
		t.Errorf("%d replies in the file, want %d", got, fixtureQuestions)
	}
}

func TestEXPSR09_TheHeaderRecordsWhatARepeatNeeds(t *testing.T) {
	root := fixtureRoot(t)
	commit := "unknown"
	if _, err := exec.LookPath("git"); err == nil {
		for _, args := range [][]string{
			{"init", "-q"}, {"add", "-A"},
			{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "fixture"},
		} {
			if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v %s", args, err, out)
			}
		}
		out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatal(err)
		}
		commit = strings.TrimSpace(string(out))
	}
	f := newFakeServer(t)
	dir := t.TempDir()
	if code, _, stderr := runFixture(t, f, root, dir, "-quick"); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	c, err := ReadResults(resultsFile(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := LoadCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	w, err := LoadWiki(root)
	if err != nil {
		t.Fatal(err)
	}
	h := c.Header
	if h.Type != "run" || h.Commit != commit || h.Settings != DefaultSettings() {
		t.Errorf("header type %q commit %q settings %+v", h.Type, h.Commit, h.Settings)
	}
	if h.Server.Version != "0.12.99-fake" || h.Server.Digest != "sha256:fakedigest" || h.Server.Model != "qwen2.5-coder:14b" {
		t.Errorf("server facts = %+v", h.Server)
	}
	if h.CorpusHash == "" || h.CorpusHash != corpus.Hash() {
		t.Errorf("corpus hash %q, want %q", h.CorpusHash, corpus.Hash())
	}
	if h.ModelHash == "" || h.ModelHash != w.Hash() {
		t.Errorf("systems model hash %q, want %q", h.ModelHash, w.Hash())
	}
	if h.TestInstructions != SystemMessage(TaskTest) || h.IncidentInstructions != SystemMessage(TaskIncident) || h.Key.Report == "" {
		t.Errorf("header instructions or key missing: %+v", h.Key)
	}
	if h.BudgetTest != BudgetTest || h.BudgetIncident != BudgetIncident {
		t.Errorf("budgets %d and %d", h.BudgetTest, h.BudgetIncident)
	}
}

// truncate keeps the header, the baseline lines and the first n call lines of
// a results file, as a run stopped part way would leave it.
func truncate(t *testing.T, src, dst string, n int, edit func(string) string) int {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	var kept []string
	calls, replies := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var cl CallLine
		_ = json.Unmarshal([]byte(line), &cl)
		switch cl.Type {
		case "summary":
			continue
		case "call":
			if calls == n {
				continue
			}
			calls++
			if cl.Reply != nil {
				replies++
			}
		case "run":
			line = edit(line)
		}
		kept = append(kept, line)
	}
	if err := os.WriteFile(dst, []byte(strings.Join(kept, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return replies
}

func callKeys(t *testing.T, path string) []string {
	c, err := ReadResults(path)
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, cl := range c.Calls {
		keys = append(keys, cl.Key)
	}
	return keys
}

func TestEXPSR10_ResumingAsksOnlyWhatIsMissing(t *testing.T) {
	root := fixtureRoot(t)
	f := newFakeServer(t)
	first := t.TempDir()
	if code, _, stderr := runFixture(t, f, root, first); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	full := resultsFile(t, first)
	stopped := filepath.Join(t.TempDir(), "stopped.jsonl")
	answered := truncate(t, full, stopped, 13, func(s string) string { return s })

	g := newFakeServer(t)
	second := t.TempDir()
	if code, _, stderr := runFixture(t, g, root, second, "-resume", stopped); code != 0 {
		t.Fatalf("resumed run: exit %d: %s", code, stderr)
	}
	if got, want := len(g.chats()), fixtureQuestions-answered; got != want {
		t.Errorf("the resumed run asked %d questions, want %d", got, want)
	}
	if !reflect.DeepEqual(callKeys(t, resultsFile(t, second)), callKeys(t, full)) {
		t.Error("the resumed run's file doesn't hold the same calls as the full run's")
	}
}

func TestEXPSR10_ResumingRefusesADifferentCorpusOrSettings(t *testing.T) {
	root := fixtureRoot(t)
	f := newFakeServer(t)
	first := t.TempDir()
	if code, _, stderr := runFixture(t, f, root, first, "-quick"); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	g := newFakeServer(t)
	for name, field := range map[string]string{"corpus": "corpus_hash", "systems model": "model_hash"} {
		stopped := filepath.Join(t.TempDir(), "other.jsonl")
		truncate(t, resultsFile(t, first), stopped, 3, func(s string) string {
			i := strings.Index(s, `"`+field+`":"`)
			if i < 0 {
				t.Fatalf("no %s in the header", field)
			}
			j := i + len(field) + 4
			return s[:j] + "0000" + s[j+4:]
		})
		code, _, stderr := runFixture(t, g, root, t.TempDir(), "-quick", "-resume", stopped)
		if code != 2 || !strings.Contains(stderr, "different") {
			t.Errorf("another %s: exit %d, stderr %q", name, code, stderr)
		}
	}

	same := filepath.Join(t.TempDir(), "same.jsonl")
	truncate(t, resultsFile(t, first), same, 3, func(s string) string { return s })
	code, _, stderr := runFixture(t, g, root, t.TempDir(), "-quick", "-seed", "7", "-resume", same)
	if code != 2 || !strings.Contains(stderr, "different") {
		t.Errorf("other settings: exit %d, stderr %q", code, stderr)
	}
	if n := len(g.chats()); n != 0 {
		t.Errorf("%d questions asked by refused runs", n)
	}
}

type recordingTransport struct {
	mu   sync.Mutex
	urls []*url.URL
	next http.RoundTripper
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.urls = append(r.urls, req.URL)
	r.mu.Unlock()
	return r.next.RoundTrip(req)
}

func TestEXPSR12_OnlyTheGivenServerIsContacted(t *testing.T) {
	rec := &recordingTransport{next: http.DefaultTransport}
	saved := transport
	transport = rec
	t.Cleanup(func() { transport = saved })

	f := newFakeServer(t)
	if code, _, stderr := runFixture(t, f, fixtureRoot(t), t.TempDir()); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	given, _ := url.Parse(f.URL)
	if len(rec.urls) < fixtureQuestions {
		t.Fatalf("%d requests seen, fewer than the %d questions", len(rec.urls), fixtureQuestions)
	}
	for _, u := range rec.urls {
		if u.Host != given.Host {
			t.Errorf("a request went to %s, not %s", u, given.Host)
		}
	}
}

func TestEXPSR14_AQuickRunTakesTwelveTestsAndOneIncident(t *testing.T) {
	c, err := LoadCorpus(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	s := quickSample(c.Tests)
	keyed := 0
	for _, tt := range s {
		if tt.Gold != "" {
			keyed++
		}
	}
	if len(s) != 12 || keyed != 8 {
		t.Fatalf("quick sample has %d tests, %d keyed, want 12 and 8", len(s), keyed)
	}
	if !reflect.DeepEqual(s, quickSample(c.Tests)) {
		t.Error("the quick sample changed between calls")
	}

	f := newFakeServer(t)
	results := readRun(t, f, "-quick")
	// The fixture has six keyed tests and two without, fewer than a quick run
	// takes, so it takes them all.
	if got := len(finals(results, ProbeBase)); got != 8 {
		t.Errorf("%d tests browsed in a quick run of the fixture, want 8", got)
	}
	incidents := finals(results, ProbeIncident)
	if len(incidents) != 1 || incidents[0].Variant != VariantReported {
		t.Errorf("incident browses in a quick run: %+v", incidents)
	}
	if got := len(f.chats()); got != 57 {
		t.Errorf("a quick run of the fixture asked %d questions, want 57", got)
	}
}

func TestEXPSR22_TheIncidentIsBrowsedInFiveWays(t *testing.T) {
	f := newFakeServer(t)
	c := readRun(t, f)
	lines := finals(c, ProbeIncident)
	var variants []string
	for _, cl := range lines {
		variants = append(variants, cl.Variant)
		if cl.Account == nil || len(cl.Items) == 0 || len(cl.Citations) == 0 || cl.SysML == "" {
			t.Errorf("%s: account %v, %d key items, %d citations, view %q", cl.Variant, cl.Account, len(cl.Items), len(cl.Citations), cl.SysML)
		}
	}
	want := []string{VariantReported, VariantAgain, VariantAlertOnly, VariantRemoved, VariantControl}
	if !reflect.DeepEqual(variants, want) {
		t.Fatalf("incident variants %v, want %v", variants, want)
	}
	key, err := LoadKey(fixtureKey)
	if err != nil {
		t.Fatal(err)
	}
	for _, cl := range c.Calls {
		if cl.Probe != ProbeIncident || cl.Step != 1 {
			continue
		}
		task := cl.User
		switch cl.Variant {
		case VariantAlertOnly:
			if !strings.Contains(task, key.AlertOnly) || strings.Contains(task, "parser exited") {
				t.Errorf("the alert-only task:\n%s", task)
			}
		default:
			if !strings.Contains(task, key.Report) {
				t.Errorf("%s task:\n%s", cl.Variant, task)
			}
		}
	}
	for _, cl := range c.Calls {
		if cl.Probe == ProbeIncident && cl.Tool == "links" {
			removed := !strings.Contains(cl.ToolAnswer, "onParserExit")
			if removed != (cl.Variant == VariantRemoved) {
				t.Errorf("%s: links of ServerStates:\n%s", cl.Variant, cl.ToolAnswer)
			}
		}
	}
	for _, cl := range lines {
		cited, _ := citation(cl.Citations, "mechanism", "ServerStates::onParserExit")
		if cited.Found == (cl.Variant == VariantRemoved) {
			t.Errorf("%s: the citation of the transition is %+v", cl.Variant, cited)
		}
	}
}
