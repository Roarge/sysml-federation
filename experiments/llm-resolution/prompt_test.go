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
		task := TestTask(tt)
		if key.MatchString(task.Text) {
			t.Errorf("the task for %s shows a key:\n%s", tt.ID, task.Text)
		}
		if tt.Name != "" && !strings.Contains(task.Text, tt.Name) {
			t.Errorf("the task for %s doesn't show its name %q", tt.ID, tt.Name)
		}
		user := StepQuestion(task, View{}, BudgetTest, nil)
		if !strings.HasPrefix(user, "Task:\n"+task.Text+"\n") {
			t.Errorf("the question for %s doesn't open with its task:\n%s", tt.ID, user)
		}
	}
}

func TestEXPSR04_QuestionsOfOneTaskShareOneSystemMessage(t *testing.T) {
	f := newFakeServer(t)
	readRun(t, f)
	systems := map[string]int{}
	var testSystem, incidentSystem string
	for _, q := range f.chats() {
		systems[q.system()]++
		switch {
		case taskName(q.user()) != "":
			if testSystem == "" {
				testSystem = q.system()
			} else if q.system() != testSystem {
				t.Fatalf("two test questions have different system messages")
			}
		case strings.Contains(q.user(), "query page"):
			if incidentSystem == "" {
				incidentSystem = q.system()
			} else if q.system() != incidentSystem {
				t.Fatalf("two incident questions have different system messages")
			}
		}
	}
	if len(systems) != 2 || testSystem == "" || incidentSystem == "" || testSystem == incidentSystem {
		t.Fatalf("%d system messages in a run, want one for the tests and another for the incident", len(systems))
	}
	if testSystem != SystemMessage(TaskTest) || incidentSystem != SystemMessage(TaskIncident) {
		t.Error("the system messages sent aren't the task's fixed instructions")
	}
	for _, tool := range []string{"find", "links", "doc", "path", "code", "grep", "read", "done"} {
		if !strings.Contains(testSystem, tool) || !strings.Contains(incidentSystem, tool) {
			t.Errorf("the instructions don't describe %s", tool)
		}
	}
}
