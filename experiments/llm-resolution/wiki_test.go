package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestEXPSR01_ElementsGetShortUniqueIDs(t *testing.T) {
	w, _ := fixtureWiki(t)
	cases := []struct{ qname, id, kind string }{
		{"Fixture_SystemStories::SR_02_ServeResults", "SR-02", "requirement"},
		{"Fixture_StakeholderStories::US_01_Search", "US-01", "requirement"},
		{"Fixture_LogicalArchitecture::Server", "Server", "part def"},
		{"Fixture_LogicalArchitecture::system", "system", "part"},
		{"Fixture_LogicalArchitecture::system::server", "system::server", "part"},
		{"Fixture_LogicalArchitecture::Server::parserChild", "Server::parserChild", "ref part"},
		{"Fixture_LogicalArchitecture::Server::lifecycle", "Server::lifecycle", "exhibit state"},
		{"Fixture_LogicalArchitecture::system::answer", "system::answer", "perform action"},
		{"Fixture_Behaviour::ServerStates", "ServerStates", "state def"},
		{"Fixture_Behaviour::ServerStates::serving", "ServerStates::serving", "state"},
		{"Fixture_Behaviour::ServerStates::onParserExit", "ServerStates::onParserExit", "transition"},
		{"Fixture_Behaviour::AnswerAQuery::parse", "AnswerAQuery::parse", "action"},
		{"Fixture_StakeholderStories::us01Derives", "us01Derives", "derivation connection"},
		{"Fixture_VerificationCases::VC_SR_02", "VC_SR_02", "verification def"},
		{"Fixture_VerificationCases::VC_SR_02::rankedPage", "TestSR02_ReturnsARankedPage", "action"},
		{"Fixture_Behaviour", "Fixture_Behaviour", "package"},
	}
	for _, c := range cases {
		e, ok := w.Get(c.qname)
		if !ok {
			t.Errorf("no element %s", c.qname)
			continue
		}
		if e.ID != c.id || e.Kind != c.kind || e.QName != c.qname {
			t.Errorf("%s: id %q kind %q qname %q, want %q %q", c.qname, e.ID, e.Kind, e.QName, c.id, c.kind)
		}
		if byID, ok := w.Get(c.id); !ok || byID != e {
			t.Errorf("%s is not found by its identifier %q", c.qname, c.id)
		}
	}
	if e, ok := w.Get("Server"); !ok || !strings.Contains(e.Doc, "serves one ranked page per query") {
		t.Errorf("Server's description = %q", e.Doc)
	}
	if e, ok := w.Get("SR-02"); !ok || e.Attr("statement") != `When a query is parsed, the server shall return the "top" results as a ranked page.` {
		t.Errorf("SR-02's statement = %q", e.Attr("statement"))
	}
	if e, ok := w.Get("Server"); !ok || e.Attr("listenAddress") != "127.0.0.1:4011" || len(e.Decisions()) != 1 {
		t.Errorf("Server's attributes and decisions: %+v", e)
	}

	real := realWiki(t)
	seen := map[string]string{}
	for _, e := range real.Elements {
		if e.ID == "" {
			t.Errorf("%s has no identifier", e.QName)
		}
		if other, dup := seen[e.ID]; dup {
			t.Errorf("%s and %s share the identifier %s", other, e.QName, e.ID)
		}
		seen[e.ID] = e.QName
	}
	if len(real.Elements) < 1000 {
		t.Errorf("read %d elements of the checkout's systems model", len(real.Elements))
	}
	for qname, id := range map[string]string{
		"Federation_LogicalArchitecture::demo::router":                                          "demo::router",
		"Federation_LogicalArchitecture::Supervisor":                                            "Supervisor",
		"Federation_FunctionalArchitecture::Supervision::SupervisorStates::onRouterExit":        "SupervisorStates::onRouterExit",
		"Federation_SystemStories::SR_04_FourPathsOnOnePort":                                    "SR-04",
		"Federation_CheckCases::CHK_MonitorViewer":                                              "CHK_MonitorViewer",
		"Federation_LogicalArchitecture::CheckSession::session::tunnelConnector::named":         "tunnelConnector::named",
		"Federation_FunctionalArchitecture::VerdictComputation::MaximumFlow":                    "MaximumFlow",
		"Federation_FunctionalArchitecture::EditPropagation::PropagateAnEdit::recomputeVerdict": "PropagateAnEdit::recomputeVerdict",
		"Federation_FunctionalArchitecture::Supervision::SupervisorStates":                      "SupervisorStates",
		"Federation_FunctionalArchitecture::Supervision::RouterExited":                          "RouterExited",
		"Federation_LogicalArchitecture::CheckSession::session::traceOneRequest":                "session::traceOneRequest",
		"Federation_FunctionalArchitecture::Session::RunASession::whichTunnel":                  "RunASession::whichTunnel",
		"Federation_FunctionalArchitecture::VerdictComputation::MeetsItsLimit":                  "MeetsItsLimit",
		"Federation_LogicalArchitecture::CheckSession::session::runner::runASession":            "runner::runASession",
		"Federation_FunctionalArchitecture::Supervision::SupervisorStates::stopping":            "SupervisorStates::stopping",
		"Federation_FunctionalArchitecture::EditPropagation::PropagateAnEdit":                   "PropagateAnEdit",
		"Federation_StakeholderStories::US_01_LaunchWithOneCommand":                             "US-01",
		"Federation_LogicalArchitecture::demo::supervisor":                                      "demo::supervisor",
		"Federation_LogicalArchitecture::demo::capacity":                                        "demo::capacity",
		"Federation_LogicalArchitecture::Router":                                                "Router",
		"Federation_FunctionalArchitecture::Session::RunASession::probeSubscriptions":           "RunASession::probeSubscriptions",
		"Federation_FunctionalArchitecture::Session::TraceOneCheckRequest::proxyToTheRouter":    "TraceOneCheckRequest::proxyToTheRouter",
		"Federation_LogicalArchitecture::demo::uiServer":                                        "demo::uiServer",
		"Federation_LogicalArchitecture::Supervisor::routerChild":                               "Supervisor::routerChild",
		"Federation_LogicalArchitecture::Supervisor::lifecycle":                                 "Supervisor::lifecycle",
		"Federation_SystemStories::SR_09_StateLivesInMemory":                                    "SR-09",
	} {
		e, ok := real.Get(qname)
		if !ok {
			t.Errorf("the checkout's systems model has no %s", qname)
			continue
		}
		if e.ID != id {
			t.Errorf("%s has the identifier %q, want %q", qname, e.ID, id)
		}
	}
}

// hasLink reports whether the element shows the link to other.
func hasLink(w *Wiki, id, rel, other string) bool {
	for _, l := range w.LinksOf(id) {
		if l.Rel == rel && l.Other == other {
			return true
		}
	}
	return false
}

func TestEXPSR01_EveryLinkShowsFromBothEnds(t *testing.T) {
	w, _ := fixtureWiki(t)
	cases := []struct{ from, rel, to, back string }{
		{"system::parser", "satisfies", "SR-01", "satisfied by"},
		{"system", "satisfies", "SC-01", "satisfied by"},
		{"VC_SR_02", "verifies", "SR-02", "verified by"},
		{"US-01", "derives", "SR-02", "derived from"},
		{"AnswerAQuery::serve", "allocated to", "system::server", "allocated from"},
		{"system", "performs", "AnswerAQuery", "performed by"},
		{"Server", "exhibits", "ServerStates", "exhibited by"},
		{"system::parser", "bound to", "Server::parserChild", "bound to"},
		{"ServerStates::onParserExit", "from", "ServerStates::serving", "left by"},
		{"ServerStates::onParserExit", "to", "ServerStates::stopped", "entered by"},
		{"ServerStates::onParserExit", "accepts", "ParserExited", "accepted by"},
		{"ServerStates", "starts in", "ServerStates::serving", "start of"},
		{"AnswerAQuery::parse", "then", "AnswerAQuery::serve", "after"},
		{"system::server", "typed by", "Server", "type of"},
		{"system::answer", "typed by", "AnswerAQuery", "type of"},
		{"Server", "owns", "Server::parserChild", "owned by"},
		{"Fixture_LogicalArchitecture", "owns", "Server", "owned by"},
		{"SR-01", "typed by", "UserStory", "type of"},
	}
	for _, c := range cases {
		if !hasLink(w, c.from, c.rel, c.to) {
			t.Errorf("%s doesn't show %s %s: %+v", c.from, c.rel, c.to, w.LinksOf(c.from))
		}
		if !hasLink(w, c.to, c.back, c.from) {
			t.Errorf("%s doesn't show %s %s: %+v", c.to, c.back, c.from, w.LinksOf(c.to))
		}
	}

	// Every link of the checkout's systems model shows on its other end too,
	// under the reverse name.
	real := realWiki(t)
	n := 0
	for _, e := range real.Elements {
		for _, l := range real.LinksOf(e.ID) {
			n++
			if !hasLink(real, l.Other, Reverse(l.Rel), e.ID) {
				t.Errorf("%s shows %s %s, but %s doesn't show %s %s", e.ID, l.Rel, l.Other, l.Other, Reverse(l.Rel), e.ID)
			}
		}
	}
	if n == 0 {
		t.Fatal("the checkout's systems model has no links")
	}
}

// statementsIn counts the statements a pattern opens in the systems model's
// files under root.
func statementsIn(t *testing.T, root string, re *regexp.Regexp) int {
	t.Helper()
	n := 0
	for _, dir := range []string{"model/library", "model/core"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".sysml") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			n += len(re.FindAll(data, -1))
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return n
}

// linksNamed counts the links of one relationship across a wiki.
func linksNamed(w *Wiki, rel string) int {
	n := 0
	for _, e := range w.Elements {
		for _, l := range w.LinksOf(e.ID) {
			if l.Rel == rel {
				n++
			}
		}
	}
	return n
}

func TestEXPSR01_EveryReferenceInTheSystemsModelResolves(t *testing.T) {
	w, _ := fixtureWiki(t)
	if len(w.Unresolved) != 0 {
		t.Errorf("the fixture's unresolved references: %v", w.Unresolved)
	}

	real := realWiki(t)
	for _, u := range real.Unresolved {
		t.Errorf("unresolved: %s", u)
	}
	root := repoRoot(t)
	for rel, re := range map[string]*regexp.Regexp{
		"satisfies":    regexp.MustCompile(`(?m)^\s*satisfy\s`),
		"allocated to": regexp.MustCompile(`(?m)^\s*allocate\s`),
		"bound to":     regexp.MustCompile(`(?m)^\s*bind\s`),
	} {
		want := statementsIn(t, root, re)
		if rel == "bound to" {
			want *= 2 // a binding shows on both of its ends under the same name
		}
		if got := linksNamed(real, rel); got != want || want == 0 {
			t.Errorf("%d %s links, and %d statements that make them", got, rel, want)
		}
	}
	for _, c := range []struct{ from, rel, to string }{
		{"demo::router", "satisfies", "SR-04"},
		{"CHK_MonitorViewer", "verifies", "SR-04"},
		{"US-01", "derives", "SR-04"},
		{"demo::router", "bound to", "Supervisor::routerChild"},
		{"Supervisor", "exhibits", "SupervisorStates"},
		{"SupervisorStates::onRouterExit", "accepts", "RouterExited"},
		{"SupervisorStates::onRouterExit", "to", "SupervisorStates::stopping"},
		{"PropagateAnEdit::recomputeVerdict", "allocated to", "demo::capacity"},
		{"TraceOneCheckRequest::proxyToTheRouter", "allocated to", "demo::uiServer"},
		{"RunASession::whichTunnel", "then", "RunASession::probeSubscriptions"},
		{"demo::uiServer", "connected to", "demo::router"},
		{"demo::capacity", "typed by", "CapacityService"},
		{"CHK_MonitorViewer", "subject", "Demo"},
	} {
		if !hasLink(real, c.from, c.rel, c.to) {
			t.Errorf("%s doesn't show %s %s", c.from, c.rel, c.to)
		}
	}
}
