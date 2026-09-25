package main

import (
	"regexp"
	"strings"
	"sync"
	"testing"
)

// answerIDs are the identifiers a list answer opens its lines with.
func answerIDs(answer string) []string {
	var ids []string
	for _, line := range strings.Split(answer, "\n") {
		if f := strings.Fields(line); len(f) > 1 && strings.HasPrefix(f[1], "(") {
			ids = append(ids, f[0])
		}
	}
	return ids
}

var (
	realCodeOnce sync.Once
	realCode     *CodeBase
	realCodeErr  error
)

// realKit gives the tools for one task over the checkout.
func realKit(t *testing.T, task string) *ToolKit {
	t.Helper()
	root := repoRoot(t)
	realCodeOnce.Do(func() { realCode, realCodeErr = LoadCode(root) })
	if realCodeErr != nil {
		t.Fatal(realCodeErr)
	}
	c, err := LoadCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	return NewToolKit(realWiki(t), realCode, c.Requirements, Scope{Task: task})
}

var keyedTestName = regexp.MustCompile(`TestS[RC]\d\d_\w+`)

func TestEXPSR02_TheRegisterIsHiddenInTheTestTask(t *testing.T) {
	// In the fixture, the register names TestSR02_ReturnsARankedPage.
	kit, _, _ := fixtureKit(t, TaskTest, Scope{})
	for _, call := range [][3]string{
		{"find", "ranked page", ""}, {"links", "VC_SR_02", ""}, {"doc", "VC_SR_02", ""},
		{"code", "VC_SR_02", ""}, {"grep", "ReturnsARankedPage", ""}, {"read", "adapter/serve/serve_test.go", "3"},
		{"path", "VC_SR_02", "SR-02"}, {"links", "SR-02", ""},
	} {
		if answer := kit.Call(call[0], call[1], call[2]); strings.Contains(answer, "TestSR02_ReturnsARankedPage") {
			t.Errorf("%s %s %s names a registered test:\n%s", call[0], call[1], call[2], answer)
		}
	}
	if answer := kit.Call("doc", "TestSR02_ReturnsARankedPage", ""); !strings.Contains(answer, "unknown") {
		t.Errorf("the registered test's action is still readable in the test task:\n%s", answer)
	}

	// In the checkout, no answer names a keyed test at all.
	real := realWiki(t)
	kit = realKit(t, TaskTest)
	var cases []string
	for _, e := range real.Elements {
		if e.Kind == "verification def" && strings.HasPrefix(e.ID, "VC_") {
			cases = append(cases, e.ID)
		}
	}
	if len(cases) < 30 {
		t.Fatalf("%d verification cases in the checkout", len(cases))
	}
	for _, id := range cases {
		for _, tool := range []string{"links", "doc", "code"} {
			if answer := kit.Call(tool, id, ""); keyedTestName.MatchString(answer) {
				t.Errorf("%s %s names a keyed test:\n%s", tool, id, answer)
			}
		}
	}
	for _, q := range []string{"func TestSR", "TestSR04", "four paths on one port"} {
		if answer := kit.Call("grep", q, ""); keyedTestName.MatchString(answer) {
			t.Errorf("grep %q names a keyed test:\n%s", q, answer)
		}
		if answer := kit.Call("find", q, ""); keyedTestName.MatchString(answer) {
			t.Errorf("find %q names a keyed test:\n%s", q, answer)
		}
	}
}

func TestEXPSR02_KeysAreMaskedInCodeAnswers(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskTest, Scope{})
	read := kit.Call("read", "adapter/serve/serve.go", "3")
	if strings.Contains(read, "SR-02") || !strings.Contains(read, "listens ([key])") {
		t.Errorf("read in the test task:\n%s", read)
	}
	grep := kit.Call("grep", "where the server listens", "")
	if strings.Contains(grep, "SR-02") || !strings.Contains(grep, "[key]") {
		t.Errorf("grep in the test task:\n%s", grep)
	}
	test := kit.Call("read", "adapter/parse/parse_test.go", "8")
	if strings.Contains(test, "TestSR01") || !strings.Contains(test, "func ParsesTokens(t *testing.T)") {
		t.Errorf("read of a test file in the test task:\n%s", test)
	}

	incident, _, _ := fixtureKit(t, TaskIncident, Scope{})
	if read := incident.Call("read", "adapter/serve/serve.go", "3"); !strings.Contains(read, "listens (SR-02)") {
		t.Errorf("read in the incident task:\n%s", read)
	}
}

func TestEXPSR03_SearchForRequirementsRanksAsTheBaseline(t *testing.T) {
	check := func(kit *ToolKit, c Corpus, n int) {
		b := NewBaseline(c.Requirements)
		for i, tt := range c.Tests {
			if i == n {
				break
			}
			var want []string
			for _, r := range b.Rank(tt) {
				if r.Score > 0 && len(want) < LimitFind {
					want = append(want, r.Key)
				}
			}
			answer := kit.Call("find", testText(tt), "requirement")
			if got := answerIDs(answer); strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("find for %s gives %v, the baseline %v:\n%s", tt.ID, got, want, answer)
			}
		}
	}
	kit, _, root := fixtureKit(t, TaskTest, Scope{})
	c, err := LoadCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	check(kit, c, len(c.Tests))

	real, err := LoadCorpus(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	check(realKit(t, TaskTest), real, 40)
}

func TestEXPSR17_LinksAreGroupedByRelationship(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	answer := kit.Call("links", "system::parser", "")
	for _, line := range []string{
		"typed by: Parser",
		"satisfies: SR-01",
		"bound to: Server::parserChild",
		"allocated from: AnswerAQuery::parse",
		"owned by: system",
	} {
		if !strings.Contains(answer, "\n"+line+"\n") && !strings.HasSuffix(answer, "\n"+line) {
			t.Errorf("links for system::parser lacks the line %q:\n%s", line, answer)
		}
	}
	if strings.Contains(answer, "quokka") || strings.Contains(answer, "The parser, in") {
		t.Errorf("links shows the description:\n%s", answer)
	}
	only := kit.Call("links", "SR-02", "satisfied by")
	if !strings.Contains(only, "satisfied by: system::server") || strings.Contains(only, "verified by") {
		t.Errorf("links for SR-02, satisfied by only:\n%s", only)
	}
	back := kit.Call("links", "SR-02", "")
	for _, line := range []string{"satisfied by: system::server", "verified by: VC_SR_02", "derived from: US-01 (via us01Derives)"} {
		if !strings.Contains(back, line) {
			t.Errorf("links for SR-02 lacks %q:\n%s", line, back)
		}
	}
}

func TestEXPSR17_PathsHaveAtMostThreeHops(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	answer := kit.Call("path", "SR-02", "ServerStates")
	want := "SR-02 -satisfied by-> system::server -typed by-> Server -exhibits-> ServerStates"
	lines := strings.Split(strings.TrimSpace(answer), "\n")
	if lines[0] != want {
		t.Errorf("the first path from SR-02 to ServerStates:\n%s\nwant %s", answer, want)
	}
	if far := kit.Call("path", "US-01", "ServerStates"); !strings.Contains(far, "no path") {
		t.Errorf("US-01 is five links from ServerStates, but path gives:\n%s", far)
	}
	real := realKit(t, TaskIncident)
	for _, pair := range [][2]string{
		{"CHK_MonitorViewer", "demo::router"}, {"demo::router", "Supervisor"}, {"SR-04", "SR-09"}, {"demo", "SupervisorStates"},
	} {
		answer := real.Call("path", pair[0], pair[1])
		lines := strings.Split(strings.TrimSpace(answer), "\n")
		if len(lines) > LimitPaths {
			t.Errorf("path %s %s gives %d paths:\n%s", pair[0], pair[1], len(lines), answer)
		}
		last := 0
		for _, line := range lines {
			hops := strings.Count(line, "->")
			if strings.Contains(line, "no path") {
				continue
			}
			if hops > 3 || hops < last || !strings.HasPrefix(line, pair[0]+" ") || !strings.HasSuffix(line, " "+pair[1]) {
				t.Errorf("path %s %s: %q", pair[0], pair[1], line)
			}
			last = hops
		}
	}
	if answer := real.Call("path", "CHK_MonitorViewer", "demo::router"); !strings.Contains(answer, "CHK_MonitorViewer -verifies-> SR-04 -satisfied by-> demo::router") {
		t.Errorf("path from the monitor to the router:\n%s", answer)
	}
}

func TestEXPSR17_EveryAnswerStaysWithinItsLimit(t *testing.T) {
	kit := realKit(t, TaskIncident)
	for _, c := range []struct {
		tool, arg, arg2 string
		more            bool
	}{
		{"find", "server", "", true},
		{"find", "the router", "part", false},
		{"links", "Demo", "", true},
		{"links", "Demo", "subject of", true},
		{"doc", "demo", "", false},
		{"code", "demo", "", true},
		{"grep", "func ", "", true},
		{"read", "cmd/sysml-federation/serve.go", "100", false},
		{"path", "demo", "SR-04", false},
	} {
		answer := kit.Call(c.tool, c.arg, c.arg2)
		if len(answer) > LimitAnswer {
			t.Errorf("%s %s is %d characters, over %d:\n%s", c.tool, c.arg, len(answer), LimitAnswer, answer)
		}
		if c.more && !strings.Contains(answer, "more") {
			t.Errorf("%s %s was cut without saying how many more:\n%s", c.tool, c.arg, answer)
		}
	}
	if n := len(answerIDs(kit.Call("find", "server", ""))); n != LimitFind {
		t.Errorf("find gives %d elements, want %d", n, LimitFind)
	}
	if answer := kit.Call("doc", "demo", ""); len(answer) > LimitDoc {
		t.Errorf("doc is %d characters, over %d", len(answer), LimitDoc)
	}
}

func TestEXPSR17_AnUnknownIdentifierIsAnsweredAsUnknown(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	for _, call := range [][3]string{
		{"links", "NoSuchElement", ""}, {"doc", "NoSuchElement", ""}, {"code", "NoSuchElement", ""},
		{"path", "NoSuchElement", "SR-01"}, {"path", "SR-01", "NoSuchElement"},
	} {
		answer := kit.Call(call[0], call[1], call[2])
		if !strings.Contains(answer, "unknown") || !strings.Contains(answer, "NoSuchElement") {
			t.Errorf("%s %s %s: %q", call[0], call[1], call[2], answer)
		}
	}
	if answer := kit.Call("read", "no/such/file.go", "1"); !strings.Contains(answer, "no file") {
		t.Errorf("read of a missing file: %q", answer)
	}
	if answer := kit.Call("find", "zzzzqqqq", ""); !strings.Contains(answer, "nothing") {
		t.Errorf("find with no match: %q", answer)
	}
	if answer := kit.Call("links", "Fixture_LogicalArchitecture::system::server", ""); !strings.Contains(answer, "satisfies: SR-02") {
		t.Errorf("a qualified name isn't accepted as an identifier:\n%s", answer)
	}
}

func TestEXPSR17_CodeGivesTheSystemsModelsOwnLocations(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	server := kit.Call("code", "Server", "")
	for _, want := range []string{"adapter/serve/serve.go", "127.0.0.1:4011 at adapter/serve/serve.go:4"} {
		if !strings.Contains(server, want) {
			t.Errorf("code for Server lacks %q:\n%s", want, server)
		}
	}
	if usage := kit.Call("code", "system::server", ""); !strings.Contains(usage, "adapter/serve/serve.go") {
		t.Errorf("code for system::server doesn't reach its definition's code:\n%s", usage)
	}
	cases := kit.Call("code", "VC_SR_02", "")
	for _, want := range []string{"adapter/serve/serve_test.go", "docs/notes.md"} {
		if !strings.Contains(cases, want) {
			t.Errorf("code for VC_SR_02 lacks the evidence location %q:\n%s", want, cases)
		}
	}
	if parser := kit.Call("code", "Parser", ""); !strings.Contains(parser, "adapter/parse") {
		t.Errorf("code for Parser lacks the folder its description names:\n%s", parser)
	}
}

func TestEXPSR18_GrepSearchesTheBuiltSystemOnly(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	answer := kit.Call("grep", "quokka", "")
	if !strings.Contains(answer, "adapter/serve/serve.go:3:") {
		t.Errorf("grep misses the built system's file:\n%s", answer)
	}
	for _, outside := range []string{"docs/", "model/", "testdata", ".hidden", "components.sysml"} {
		if strings.Contains(answer, outside) {
			t.Errorf("grep searched %s:\n%s", outside, answer)
		}
	}
	if n := strings.Count(answer, ".go:"); n != 1 {
		t.Errorf("grep found %d hits, want 1:\n%s", n, answer)
	}
	if none := kit.Call("grep", "zzzzqqqq", ""); !strings.Contains(none, "no match") {
		t.Errorf("grep with no match: %q", none)
	}
	if outside := kit.Call("read", "docs/notes.md", "1"); !strings.Contains(outside, "no file") {
		t.Errorf("read outside the built system: %q", outside)
	}
}

func TestEXPSR18_ReadShowsTheLinesAroundTheOneAsked(t *testing.T) {
	kit, _, _ := fixtureKit(t, TaskIncident, Scope{})
	answer := kit.Call("read", "adapter/serve/serve.go", "4")
	for _, want := range []string{"adapter/serve/serve.go", "1| package serve", "4| const Addr = \"127.0.0.1:4011\"", "12| }"} {
		if !strings.Contains(answer, want) {
			t.Errorf("read lacks %q:\n%s", want, answer)
		}
	}
	long := realKit(t, TaskIncident).Call("read", "cmd/sysml-federation/serve.go", "120")
	numbered := regexp.MustCompile(`(?m)^\s*(\d+)\| `).FindAllStringSubmatch(long, -1)
	if len(numbered) == 0 || len(numbered) > LimitReadLines || len(long) > LimitAnswer {
		t.Errorf("read of a long file shows %d lines in %d characters:\n%s", len(numbered), len(long), long)
	}
	found := false
	for _, m := range numbered {
		if m[1] == "120" {
			found = true
		}
	}
	if !found {
		t.Errorf("read of line 120 doesn't show line 120:\n%s", long)
	}
}
