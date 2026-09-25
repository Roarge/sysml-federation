package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestEXPSR04_EveryQuestionCarriesTheFixedSettings(t *testing.T) {
	s := DefaultSettings()
	if s.Model != "qwen2.5-coder:14b" || s.Temperature != 0 || s.Seed != 42 || s.NumCtx != 8192 || s.KeepAlive != "24h" || s.NumPredict != 512 {
		t.Fatalf("default settings = %+v", s)
	}
	f := newFakeServer(t)
	c := NewClient(f.URL, s)
	if _, err := c.Chat(context.Background(), "the system message", "the user message", StepSchema()); err != nil {
		t.Fatal(err)
	}
	reqs := f.chats()
	if len(reqs) != 1 {
		t.Fatalf("%d chats, want 1", len(reqs))
	}
	r := reqs[0]
	if r.Model != "qwen2.5-coder:14b" || r.Stream || r.KeepAlive != "24h" {
		t.Errorf("model %q stream %v keep_alive %q", r.Model, r.Stream, r.KeepAlive)
	}
	if r.Options.Temperature != 0 || r.Options.Seed != 42 || r.Options.NumCtx != 8192 || r.Options.NumPredict != 512 {
		t.Errorf("options = %+v", r.Options)
	}
	if r.system() != "the system message" || r.user() != "the user message" || len(r.Messages) != 2 || r.Messages[0].Role != "system" {
		t.Errorf("messages = %+v", r.Messages)
	}
	if !r.asks("tool") {
		t.Errorf("the step schema wasn't sent: %s", r.Format)
	}
}

// A JSON schema as far as these tests read it.
type schemaNode struct {
	Type       string                `json:"type"`
	Enum       []string              `json:"enum"`
	MaxItems   int                   `json:"maxItems"`
	Items      *schemaNode           `json:"items"`
	Properties map[string]schemaNode `json:"properties"`
	Required   []string              `json:"required"`
}

func parseSchema(t *testing.T, raw json.RawMessage) schemaNode {
	t.Helper()
	var s schemaNode
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("%v: %s", err, raw)
	}
	return s
}

func TestEXPSR04_TheStepSchemaAdmitsOneToolAndViewEdits(t *testing.T) {
	s := parseSchema(t, StepSchema())
	tool := s.Properties["tool"]
	if got := strings.Join(tool.Enum, ","); got != "find,links,doc,path,code,grep,read,done" {
		t.Errorf("tools = %s", got)
	}
	for _, p := range []string{"viewpoint", "tool", "arg", "arg2", "expose", "prune"} {
		if _, ok := s.Properties[p]; !ok {
			t.Errorf("the step schema lacks %s", p)
		}
	}
	expose, prune := s.Properties["expose"], s.Properties["prune"]
	if expose.MaxItems != 3 || prune.MaxItems != 3 || expose.Items == nil || prune.Items == nil || prune.Items.Type != "string" {
		t.Errorf("expose %+v, prune %+v", expose, prune)
	}
	if _, ok := expose.Items.Properties["id"]; !ok {
		t.Errorf("an exposure has no id: %+v", expose.Items)
	}
	if _, ok := expose.Items.Properties["note"]; !ok {
		t.Errorf("an exposure has no note: %+v", expose.Items)
	}
	if len(s.Required) != len(s.Properties) {
		t.Errorf("required %v of %d properties", s.Required, len(s.Properties))
	}
	step, err := ParseStep(`{"viewpoint":"v","tool":"links","arg":"SR-01","arg2":"","expose":[{"id":"SR-01","note":"n"}],"prune":["x"]}`)
	if err != nil || step.Tool != "links" || step.Expose[0].ID != "SR-01" || step.Prune[0] != "x" {
		t.Errorf("parsed %+v, %v", step, err)
	}
}

func TestEXPSR04_AReplyThatIsNotTheAskedJSONIsAnError(t *testing.T) {
	f := newFakeServer(t)
	f.reply = func(capturedRequest) string { return "I think it's SR-01." }
	c := NewClient(f.URL, DefaultSettings())
	reply, err := c.Chat(context.Background(), "s", "u", StepSchema())
	if err != nil {
		t.Fatalf("the server answered, so Chat shouldn't fail: %v", err)
	}
	if reply.Raw != "I think it's SR-01." {
		t.Errorf("raw = %q, want the text kept", reply.Raw)
	}
	if _, err := ParseStep(reply.Raw); err == nil {
		t.Error("no error for a step that isn't JSON")
	}
	if _, err := ParseTestAnswer(reply.Raw); err == nil {
		t.Error("no error for an answer that isn't JSON")
	}
	if _, err := ParseIncidentAnswer(reply.Raw); err == nil {
		t.Error("no error for an account that isn't JSON")
	}
	if _, err := ParseStep(`{"tool":"shout","arg":"x"}`); err == nil {
		t.Error("no error for a tool that doesn't exist")
	}

	// In a browse, the error is recorded with the text and the browse goes on.
	kit, w, _ := fixtureKit(t, TaskTest, Scope{})
	n := 0
	f.reply = func(q capturedRequest) string {
		n++
		if n == 1 {
			return "not JSON"
		}
		return fixturePolicy(q)
	}
	result, err := Browse(context.Background(), Task{Kind: TaskTest, ID: "x", Text: "name: ParsesTokens\npackage: parse\nfile: a_test.go\ndoc: (none)"}, kit, w, directAsker(c))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) == 0 || result.Steps[0].Error == "" || result.Steps[0].Raw != "not JSON" {
		t.Fatalf("the first step = %+v", result.Steps)
	}
	if result.Final.Test == nil {
		t.Errorf("the browse stopped at the bad reply: %+v", result)
	}
}

// propertyOrder gives the top-level property names of a schema in the order
// they are written.
func propertyOrder(t *testing.T, raw json.RawMessage) []string {
	t.Helper()
	text := string(raw)
	props := text[strings.Index(text, `"properties"`):]
	s := parseSchema(t, raw)
	type at struct {
		name string
		pos  int
	}
	var found []at
	for name := range s.Properties {
		found = append(found, at{name, strings.Index(props, `"`+name+`":`)})
	}
	for i := range found {
		for j := i + 1; j < len(found); j++ {
			if found[j].pos < found[i].pos {
				found[i], found[j] = found[j], found[i]
			}
		}
	}
	var out []string
	for _, f := range found {
		out = append(out, f.name)
	}
	return out
}

func TestEXPSR04_TheFinalSchemaAsksForTheLinksBeforeTheEvidence(t *testing.T) {
	order := strings.Join(propertyOrder(t, TestFinalSchema()), ",")
	if !strings.HasPrefix(order, "links,evidence,reason") {
		t.Fatalf("the final schema for a test asks in the order %s", order)
	}
	s := parseSchema(t, TestFinalSchema())
	links := s.Properties["links"]
	if links.MaxItems != 5 || links.Items == nil || strings.Join(links.Items.Properties["relation"].Enum, ",") != "verifies,exercises,other" {
		t.Errorf("links = %+v", links)
	}
	if s.Properties["evidence"].MaxItems != 5 {
		t.Errorf("evidence = %+v", s.Properties["evidence"])
	}
	a, err := ParseTestAnswer(`{"links":[{"id":"SR-01","relation":"verifies"}],"evidence":["x"],"reason":"r","mismatches":[]}`)
	if err != nil || a.Links[0].ID != "SR-01" {
		t.Errorf("parsed %+v, %v", a, err)
	}
}

func TestEXPSR21_EveryFinalSchemaAsksForSuspectedMismatches(t *testing.T) {
	for name, raw := range map[string]json.RawMessage{"test": TestFinalSchema(), "incident": IncidentFinalSchema()} {
		s := parseSchema(t, raw)
		m, ok := s.Properties["mismatches"]
		if !ok || m.Items == nil || m.MaxItems != 3 {
			t.Errorf("%s: mismatches = %+v", name, m)
			continue
		}
		for _, p := range []string{"id", "model_says", "system_shows", "why"} {
			if _, ok := m.Items.Properties[p]; !ok {
				t.Errorf("%s: a mismatch lacks %s", name, p)
			}
		}
	}
	s := parseSchema(t, IncidentFinalSchema())
	for _, p := range []string{"why", "cause", "mechanism", "code", "consequences", "path"} {
		if _, ok := s.Properties[p]; !ok {
			t.Errorf("the incident's final schema lacks %s", p)
		}
	}
}
