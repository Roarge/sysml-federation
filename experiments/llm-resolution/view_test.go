package main

import (
	"strings"
	"testing"
)

func TestEXPSR19_ExposeAndPruneChangeTheView(t *testing.T) {
	w, _ := fixtureWiki(t)
	var v View
	if err := v.Expose(w, "SR-02", "the requirement the page serves"); err != nil {
		t.Fatal(err)
	}
	if err := v.Expose(w, "Fixture_LogicalArchitecture::system::server", "serves the page"); err != nil {
		t.Fatal(err)
	}
	if err := v.Expose(w, "NoSuchElement", "made up"); err == nil || !strings.Contains(err.Error(), "NoSuchElement") {
		t.Errorf("exposing an unknown element gave %v", err)
	}
	if err := v.Expose(w, "SR-02", "the requirement, noted again"); err != nil {
		t.Fatal(err)
	}
	if len(v.Items) != 2 || v.Items[0] != (ViewItem{ID: "SR-02", Note: "the requirement, noted again"}) || v.Items[1].ID != "system::server" {
		t.Fatalf("view = %+v", v.Items)
	}
	lines := v.Lines()
	if !strings.Contains(lines, "SR-02: the requirement, noted again") || !strings.Contains(lines, "system::server: serves the page") {
		t.Errorf("the view as a question shows it:\n%s", lines)
	}
	if !v.Prune("SR-02") || v.Prune("SR-02") {
		t.Error("pruning SR-02 twice should succeed once")
	}
	if len(v.Items) != 1 || v.Items[0].ID != "system::server" {
		t.Errorf("after pruning, view = %+v", v.Items)
	}
	var empty View
	if !strings.Contains(empty.Lines(), "empty") {
		t.Errorf("an empty view shows %q", empty.Lines())
	}
}

func TestEXPSR19_TheViewIsWrittenOutAsSysML(t *testing.T) {
	w, _ := fixtureWiki(t)
	v := View{Viewpoint: "Why is the query page down?"}
	for _, id := range []string{"SR-02", "system::server", "ServerStates::onParserExit"} {
		if err := v.Expose(w, id, "a note"); err != nil {
			t.Fatal(err)
		}
	}
	text := v.SysML(w, "incident")
	for _, want := range []string{
		"viewpoint def WhyIsTheQueryPageDown {",
		"frame concern",
		"doc /* Why is the query page down? */",
		"view incident {",
		"viewpoint answers : WhyIsTheQueryPageDown;",
		"expose Fixture_SystemStories::SR_02_ServeResults;",
		"expose Fixture_LogicalArchitecture::system::server;",
		"expose Fixture_Behaviour::ServerStates::onParserExit;",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the SysML lacks %q:\n%s", want, text)
		}
	}
	if strings.Count(text, "{") != strings.Count(text, "}") {
		t.Errorf("unbalanced braces:\n%s", text)
	}
	odd := View{Viewpoint: "What */ breaks a comment?"}
	if out := odd.SysML(w, "v"); strings.Count(out, "*/") != 1 {
		t.Errorf("a viewpoint holding */ ends its doc early:\n%s", out)
	}
}
