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

// The fixture has eight tests, and the stand-in links seven of them, so a
// full run asks 8 base questions, 7 for each of the four probes that follow
// a link and ask whatever the evidence, 6 for the rare-shared control (the
// Health test shares no uncited word with its requirement), and 2 repeats:
// 44 in all.
const fixtureQuestions = 44

func TestEXPSR05_NothingToDeleteIsRecordedAndNotAsked(t *testing.T) {
	f := newFakeServer(t)
	f.reply = func(system, user string) string {
		a := Answer{Requirement: "none", Evidence: []string{}, Reason: "Nothing fits."}
		if m := nameLineRE.FindStringSubmatch(user); m != nil {
			if known, ok := fixtureAnswers[strings.TrimSpace(m[1])]; ok {
				a = known
				a.Evidence = []string{"zzzz"}
			}
		}
		data, _ := json.Marshal(a)
		return string(data)
	}
	dir := t.TempDir()
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", fixtureRoot(t), "-out", dir); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	c, err := ReadResults(resultsFile(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []string{"deletion", "control", "rare-shared"} {
		lines := callsWithProbe(c, probe)
		if len(lines) != 7 {
			t.Fatalf("%d %s lines, want 7", len(lines), probe)
		}
		for _, cl := range lines {
			if cl.Note == "" || cl.Reply != nil || cl.User != "" {
				t.Errorf("%s for %s: note %q, reply %v, question %q", probe, cl.Test, cl.Note, cl.Reply, cl.User)
			}
		}
	}
	if got, want := len(f.chats()), fixtureQuestions-20; got != want {
		t.Errorf("%d questions asked, want %d", got, want)
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

	f := newFakeServer(t)
	root := fixtureRoot(t)
	dir := t.TempDir()
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", root, "-out", dir); code != 0 {
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
	var got, want []string
	for _, cl := range callsWithProbe(c, "repeat") {
		got = append(got, cl.Test)
	}
	for _, tt := range repeatSample(corpus.Tests) {
		want = append(want, tt.ID)
	}
	if len(want) != 2 || !reflect.DeepEqual(got, want) {
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
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", fixtureRoot(t), "-out", dir); code != 0 {
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
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", root, "-out", dir, "-quick"); code != 0 {
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
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", root, "-out", first); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	full := resultsFile(t, first)
	stopped := filepath.Join(t.TempDir(), "stopped.jsonl")
	answered := truncate(t, full, stopped, 12, func(s string) string { return s })

	g := newFakeServer(t)
	second := t.TempDir()
	if code, _, stderr := runExperiment(t, "-url", g.URL, "-repo", root, "-out", second, "-resume", stopped); code != 0 {
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
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", root, "-out", first, "-quick"); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	stopped := filepath.Join(t.TempDir(), "other-corpus.jsonl")
	truncate(t, resultsFile(t, first), stopped, 3, func(s string) string {
		var h RunHeader
		_ = json.Unmarshal([]byte(s), &h)
		return strings.Replace(s, `"corpus_hash":"`+h.CorpusHash+`"`, `"corpus_hash":"0000000000000000"`, 1)
	})
	g := newFakeServer(t)
	code, _, stderr := runExperiment(t, "-url", g.URL, "-repo", root, "-out", t.TempDir(), "-quick", "-resume", stopped)
	if code != 2 || !strings.Contains(stderr, "different") {
		t.Errorf("another corpus: exit %d, stderr %q", code, stderr)
	}

	same := filepath.Join(t.TempDir(), "same.jsonl")
	truncate(t, resultsFile(t, first), same, 3, func(s string) string { return s })
	code, _, stderr = runExperiment(t, "-url", g.URL, "-repo", root, "-out", t.TempDir(), "-quick", "-seed", "7", "-resume", same)
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
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", fixtureRoot(t), "-out", t.TempDir()); code != 0 {
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

func TestEXPSR14_AQuickRunTakesTwelveTests(t *testing.T) {
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
	dir := t.TempDir()
	if code, _, stderr := runExperiment(t, "-url", f.URL, "-repo", fixtureRoot(t), "-out", dir, "-quick"); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	results, err := ReadResults(resultsFile(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	// The fixture has six keyed tests and two without, fewer than a quick run
	// takes, so it takes them all.
	if got := len(callsWithProbe(results, "base")); got != 8 {
		t.Errorf("%d base questions in a quick run of the fixture, want 8", got)
	}
}
