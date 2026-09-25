package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestEXPSR02_NoKeyReachesAQuestion(t *testing.T) {
	c, err := LoadCorpus(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Tests) == 0 {
		t.Fatal("no tests read")
	}
	key := regexp.MustCompile(`\bS[RC]-?\d\d\b|TestS[RC]\d\d`)
	for _, tt := range c.Tests {
		_, user := baseQuestion(c, tt)
		if key.MatchString(user) {
			t.Errorf("the question about %s shows a key:\n%s", tt.ID, user)
		}
		if tt.Name != "" && !strings.Contains(user, tt.Name) {
			t.Errorf("the question about %s doesn't show its name %q", tt.ID, tt.Name)
		}
	}
}

func TestEXPSR04_BaseQuestionsShareOneSystemMessage(t *testing.T) {
	c, err := LoadCorpus(fixtureRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	var first string
	for i, tt := range c.Tests {
		system, _ := baseQuestion(c, tt)
		if i == 0 {
			first = system
			continue
		}
		if system != first {
			t.Fatalf("the system message for %s differs from the first", tt.ID)
		}
	}
	for _, r := range c.Requirements {
		if !strings.Contains(first, r.Key) || !strings.Contains(first, r.Statement) {
			t.Errorf("the system message lacks %s or its statement", r.Key)
		}
	}
}
