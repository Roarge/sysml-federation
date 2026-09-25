package main

import (
	"context"
	"strings"
	"sync"
	"testing"
)

// directAsker asks the stand-in with no recording.
func directAsker(c *Client) Asker {
	return func(ctx context.Context, q Question) (Reply, error) {
		return c.Chat(ctx, q.System, q.User, q.Format)
	}
}

func TestEXPSR04_EachQuestionCarriesTheViewAndOnlyTheLastAnswer(t *testing.T) {
	kit, w, _ := fixtureKit(t, TaskTest, Scope{})
	f := newFakeServer(t)
	var mu sync.Mutex
	n := 0
	f.reply = func(q capturedRequest) string {
		mu.Lock()
		defer mu.Unlock()
		n++
		switch n {
		case 1:
			return stepJSON(Step{Viewpoint: "What does it check?", Tool: "grep", Arg: "quokka",
				Expose: []Exposure{{ID: "SR-02", Note: "serves the page"}}, Prune: []string{}})
		case 2:
			return stepJSON(Step{Viewpoint: "What does it check?", Tool: "links", Arg: "SR-02", Expose: []Exposure{}, Prune: []string{}})
		case 3:
			return stepJSON(Step{Viewpoint: "What does it check?", Tool: "done", Expose: []Exposure{}, Prune: []string{}})
		}
		return testAnswerJSON(TestAnswer{Links: []LinkAnswer{{ID: "SR-02", Relation: "verifies"}}, Evidence: []string{}, Mismatches: []Mismatch{}})
	}
	c := NewClient(f.URL, DefaultSettings())
	task := Task{Kind: TaskTest, ID: "t", Text: "name: SomethingElse\npackage: p\nfile: f_test.go\ndoc: (none)"}
	result, err := Browse(context.Background(), task, kit, w, directAsker(c))
	if err != nil {
		t.Fatal(err)
	}
	chats := f.chats()
	if len(chats) != 4 {
		t.Fatalf("%d questions, want three steps and the final", len(chats))
	}
	third := chats[2].user()
	for _, want := range []string{"name: SomethingElse", "What does it check?", "SR-02: serves the page", "satisfied by: system::server", "Calls left: 6"} {
		if !strings.Contains(third, want) {
			t.Errorf("the third question lacks %q:\n%s", want, third)
		}
	}
	for _, old := range []string{"serve.go:3", "quokka"} {
		if strings.Contains(third, old) {
			t.Errorf("the third question still holds the first answer's %q:\n%s", old, third)
		}
	}
	if !strings.Contains(chats[0].user(), "Last call: none") || !strings.Contains(chats[0].user(), "Calls left: 8") {
		t.Errorf("the first question:\n%s", chats[0].user())
	}
	if len(result.Steps) != 3 || result.Steps[0].Answer == "" || result.Final.Test == nil || len(result.View.Items) != 1 {
		t.Errorf("result = %+v", result)
	}
}

func TestEXPSR04_ABrowseStopsAtItsBudget(t *testing.T) {
	for _, c := range []struct {
		kind   string
		budget int
		final  string
	}{{TaskTest, 8, "links"}, {TaskIncident, 20, "cause"}} {
		kit, w, _ := fixtureKit(t, c.kind, Scope{})
		f := newFakeServer(t)
		f.reply = func(q capturedRequest) string {
			if q.asks("tool") {
				return stepJSON(Step{Tool: "find", Arg: "query", Expose: []Exposure{}, Prune: []string{}})
			}
			return fixturePolicy(q)
		}
		task := Task{Kind: c.kind, ID: "x", Text: "name: Anything\npackage: p\nfile: f_test.go\ndoc: (none)"}
		if c.kind == TaskIncident {
			task.Text = "The query page is down."
		}
		result, err := Browse(context.Background(), task, kit, w, directAsker(NewClient(f.URL, DefaultSettings())))
		if err != nil {
			t.Fatal(err)
		}
		chats := f.chats()
		if len(chats) != c.budget+1 || len(result.Steps) != c.budget {
			t.Errorf("%s: %d questions and %d steps, want %d tool calls and the final", c.kind, len(chats), len(result.Steps), c.budget)
			continue
		}
		last := chats[len(chats)-1]
		if !last.asks(c.final) || last.asks("tool") || !strings.Contains(last.user(), "Calls left: 0") {
			t.Errorf("%s: the last question isn't the final one:\n%s", c.kind, last.user())
		}
	}
}
