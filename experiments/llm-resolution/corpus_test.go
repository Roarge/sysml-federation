package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestEXPSR01_ReadsEveryRequirementWithItsStatement(t *testing.T) {
	// The fixture: three stories and a constraint, one statement with an
	// escaped quotation.
	c, err := LoadCorpus(fixtureRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"SC-01", "SR-01", "SR-02", "SR-03"}
	if got := c.Keys(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("keys = %v, want %v", got, want)
	}
	for _, r := range c.Requirements {
		if r.Statement == "" || r.Name == "" {
			t.Errorf("%s has name %q and statement %q", r.Key, r.Name, r.Statement)
		}
		if r.Key == "SR-02" && !strings.Contains(r.Statement, `the "top" results`) {
			t.Errorf("SR-02's escaped quotes weren't unescaped: %q", r.Statement)
		}
		if r.Key == "SR-01" && r.Name != "Parse queries" {
			t.Errorf("SR-01's name = %q, want %q", r.Name, "Parse queries")
		}
	}

	// The checkout: as many requirements as the two registers declare.
	root := repoRoot(t)
	real, err := LoadCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	header := regexp.MustCompile(`requirement <'S[RC]-\d\d'>`)
	declared := 0
	for _, rel := range []string{"model/core/stories/system/system-stories.sysml", "model/core/constraints/design-constraints.sysml"} {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		declared += len(header.FindAll(data, -1))
	}
	if len(real.Requirements) != declared || declared == 0 {
		t.Fatalf("read %d requirements, the registers declare %d", len(real.Requirements), declared)
	}
	for _, r := range real.Requirements {
		if r.Statement == "" {
			t.Errorf("%s was read with no statement", r.Key)
		}
	}
}

func TestEXPSR01_ReadsEveryTopLevelTestOnce(t *testing.T) {
	c, err := LoadCorpus(fixtureRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, tt := range c.Tests {
		ids = append(ids, tt.ID)
	}
	want := []string{
		"adapter/parse/parse_test.go#TestHelperBuildsAQuery",
		"adapter/parse/parse_test.go#TestSR01_ParsesTokens",
		"adapter/parse/parse_test.go#TestSR01_RejectsAnEmptyQuery",
		"adapter/serve/serve_test.go#TestHealth",
		"adapter/serve/serve_test.go#TestSC01_ImportsOnlyTheStandardLibrary",
		"adapter/serve/serve_test.go#TestSR02_ReturnsARankedPage",
		"adapter/serve/serve_test.go#TestSR02_ServesTopResults",
		"adapter/serve/serve_test.go#TestSR03_ReportsLatency",
	}
	sort.Strings(ids)
	if strings.Join(ids, "\n") != strings.Join(want, "\n") {
		t.Fatalf("tests read:\n%s\nwant:\n%s", strings.Join(ids, "\n"), strings.Join(want, "\n"))
	}
	for _, tt := range c.Tests {
		if tt.ID == "adapter/serve/serve_test.go#TestHealth" && tt.Package != "serve_test" {
			t.Errorf("package = %q, want serve_test", tt.Package)
		}
		if tt.ID == "adapter/parse/parse_test.go#TestSR01_RejectsAnEmptyQuery" && tt.Gold != "SR-01" {
			t.Errorf("gold = %q, want SR-01", tt.Gold)
		}
	}
}

func TestEXPSR01_EveryKeyATestCarriesIsDeclared(t *testing.T) {
	real, err := LoadCorpus(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := real.Validate(); err != nil {
		t.Fatalf("the checkout doesn't validate: %v", err)
	}
	keyed := 0
	for _, tt := range real.Tests {
		if tt.Gold != "" {
			keyed++
		}
	}
	if keyed == 0 {
		t.Fatal("no test in the checkout carries a key")
	}

	root := fixtureRoot(t)
	extra := "package parse\n\nimport \"testing\"\n\nfunc TestSR09_NamesAnUndeclaredKey(t *testing.T) {}\n"
	if err := os.WriteFile(filepath.Join(root, "adapter", "parse", "extra_test.go"), []byte(extra), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "SR-09") {
		t.Fatalf("Validate() = %v, want an error naming SR-09", err)
	}
}

func TestEXPSR02_KeysAreRemovedFromNamesAndDocs(t *testing.T) {
	cases := []struct {
		funcName, doc      string
		wantName, wantGold string
		wantDoc            string
	}{
		{"TestSR22_SetAttributePatches", "TestSR22_SetAttributePatches: SR-22 says the text and SR22 agree.",
			"SetAttributePatches", "SR-22", "SetAttributePatches: [key] says the text and [key] agree."},
		{"TestSC04_TrackedPaths", "", "TrackedPaths", "SC-04", ""},
		{"TestHealth", "Health answers. See TestSR46_ModelAndRepositoryAgree.", "Health", "", "Health answers. See ModelAndRepositoryAgree."},
		{"Test_Underscore", "", "Underscore", "", ""},
	}
	for _, c := range cases {
		got := newTest(c.funcName, "pkg", "a/b_test.go", c.doc)
		if got.Name != c.wantName || got.Gold != c.wantGold || got.Doc != c.wantDoc {
			t.Errorf("newTest(%q) = name %q gold %q doc %q, want %q %q %q",
				c.funcName, got.Name, got.Gold, got.Doc, c.wantName, c.wantGold, c.wantDoc)
		}
		if got.FuncName != c.funcName || got.ID != "a/b_test.go#"+c.funcName {
			t.Errorf("newTest(%q) lost the original: func %q id %q", c.funcName, got.FuncName, got.ID)
		}
	}
}
