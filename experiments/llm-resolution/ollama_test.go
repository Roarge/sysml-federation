package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestEXPSR04_EveryQuestionCarriesTheFixedSettings(t *testing.T) {
	s := DefaultSettings()
	if s.Model != "qwen2.5-coder:14b" || s.Temperature != 0 || s.Seed != 42 || s.NumCtx != 8192 || s.KeepAlive != "24h" {
		t.Fatalf("default settings = %+v", s)
	}
	f := newFakeServer(t)
	c := NewClient(f.URL, s, []string{"SR-01", "SR-02"})
	if _, err := c.Chat(context.Background(), "the system message", "the user message"); err != nil {
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
	if r.Options.Temperature != 0 || r.Options.Seed != 42 || r.Options.NumCtx != 8192 {
		t.Errorf("options = %+v", r.Options)
	}
	if len(r.Messages) != 2 || r.Messages[0].Role != "system" || r.Messages[0].Content != "the system message" ||
		r.Messages[1].Role != "user" || r.Messages[1].Content != "the user message" {
		t.Errorf("messages = %+v", r.Messages)
	}
}

func TestEXPSR04_TheReplySchemaAdmitsOnlyKeysOrNone(t *testing.T) {
	keys := []string{"SC-01", "SR-01", "SR-02"}
	s := newAnswerSchema(keys)
	if got := strings.Join(s.Properties.Requirement.Enum, ","); got != "SC-01,SR-01,SR-02,none" {
		t.Fatalf("enum = %s", got)
	}
	if strings.Join(s.Required, ",") != "requirement,evidence,reason" || s.Properties.Evidence.MaxItems != 5 {
		t.Errorf("required %v, evidence %+v", s.Required, s.Properties.Evidence)
	}
	f := newFakeServer(t)
	if _, err := NewClient(f.URL, DefaultSettings(), keys).Chat(context.Background(), "s", "u"); err != nil {
		t.Fatal(err)
	}
	sent := f.chats()[0].Format
	if strings.Join(sent.Properties.Requirement.Enum, ",") != "SC-01,SR-01,SR-02,none" || sent.Type != "object" {
		t.Errorf("sent format = %+v", sent)
	}
}

func TestEXPSR04_AReplyThatIsNotTheAskedJSONIsAnError(t *testing.T) {
	f := newFakeServer(t)
	f.reply = func(string, string) string { return "I think it's SR-01." }
	reply, err := NewClient(f.URL, DefaultSettings(), []string{"SR-01"}).Chat(context.Background(), "s", "u")
	if err == nil {
		t.Fatal("no error for a reply that isn't JSON")
	}
	if reply.Raw != "I think it's SR-01." {
		t.Errorf("raw = %q, want the text kept", reply.Raw)
	}

	f.reply = fixturePolicy
	reply, err = NewClient(f.URL, DefaultSettings(), []string{"SR-01"}).Chat(context.Background(), "s", "name: ParsesTokens")
	if err != nil || reply.Answer.Requirement != "SR-01" || len(reply.Answer.Evidence) != 1 {
		t.Fatalf("a well-formed reply gave %+v, %v", reply.Answer, err)
	}
}

func TestEXPSR04_TheSchemaAsksForTheAnswerBeforeItsEvidence(t *testing.T) {
	data, err := json.Marshal(newAnswerSchema([]string{"SR-01"}))
	if err != nil {
		t.Fatal(err)
	}
	props := string(data)[strings.Index(string(data), `"properties"`):]
	r, e, why := strings.Index(props, `"requirement"`), strings.Index(props, `"evidence"`), strings.Index(props, `"reason"`)
	if r < 0 || e < 0 || why < 0 || !(r < e && e < why) {
		t.Fatalf("the schema doesn't list requirement, evidence and reason in that order: %s", data)
	}
}
