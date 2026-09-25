package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// Settings are the generation settings every call uses. They are recorded in
// the results file with the model's digest, so a run can be repeated.
type Settings struct {
	Model       string  `json:"model"`
	NumCtx      int     `json:"num_ctx"`
	NumPredict  int     `json:"num_predict"`
	Temperature float64 `json:"temperature"`
	Seed        int     `json:"seed"`
	KeepAlive   string  `json:"keep_alive"`
	Logprobs    bool    `json:"logprobs"`
}

// DefaultSettings are not built yet.
func DefaultSettings() Settings { return Settings{} }

// Step is one tool call with the edits to the view that ride on it.
type Step struct {
	Viewpoint string     `json:"viewpoint"`
	Expose    []Exposure `json:"expose"`
	Prune     []string   `json:"prune"`
	Tool      string     `json:"tool"`
	Arg       string     `json:"arg"`
	Arg2      string     `json:"arg2"`
}

// Exposure adds an element to the view with a note saying why.
type Exposure struct {
	ID   string `json:"id"`
	Note string `json:"note"`
}

// LinkAnswer is one link of a test's final answer.
type LinkAnswer struct {
	ID       string `json:"id"`
	Relation string `json:"relation"`
}

// Mismatch is a suspicion that the system was not built as modelled.
type Mismatch struct {
	ID          string `json:"id"`
	ModelSays   string `json:"model_says"`
	SystemShows string `json:"system_shows"`
	Why         string `json:"why"`
}

// TestAnswer is the final answer about a test.
type TestAnswer struct {
	Links      []LinkAnswer `json:"links"`
	Evidence   []string     `json:"evidence"`
	Reason     string       `json:"reason"`
	Mismatches []Mismatch   `json:"mismatches"`
}

// IncidentAnswer is the final account of an incident.
type IncidentAnswer struct {
	Cause        []string   `json:"cause"`
	Mechanism    []string   `json:"mechanism"`
	Code         []string   `json:"code"`
	Consequences []string   `json:"consequences"`
	Path         []string   `json:"path"`
	Why          string     `json:"why"`
	Mismatches   []Mismatch `json:"mismatches"`
}

var errNotBuilt = errors.New("not built yet")

// StepSchema is not built yet.
func StepSchema() json.RawMessage { return json.RawMessage(`{}`) }

// TestFinalSchema is not built yet.
func TestFinalSchema() json.RawMessage { return json.RawMessage(`{}`) }

// IncidentFinalSchema is not built yet.
func IncidentFinalSchema() json.RawMessage { return json.RawMessage(`{}`) }

// ParseStep is not built yet.
func ParseStep(raw string) (Step, error) { return Step{}, errNotBuilt }

// ParseTestAnswer is not built yet.
func ParseTestAnswer(raw string) (TestAnswer, error) { return TestAnswer{}, errNotBuilt }

// ParseIncidentAnswer is not built yet.
func ParseIncidentAnswer(raw string) (IncidentAnswer, error) { return IncidentAnswer{}, errNotBuilt }

// Timing is what Ollama reports about one call.
type Timing struct {
	PromptTokens int   `json:"prompt_tokens"`
	OutputTokens int   `json:"output_tokens"`
	TotalNS      int64 `json:"total_ns"`
	PromptNS     int64 `json:"prompt_ns"`
	OutputNS     int64 `json:"output_ns"`
}

// Logprob is one generated token with its log probability.
type Logprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
}

// Reply is one call's outcome.
type Reply struct {
	Raw      string    `json:"raw"`
	Timing   Timing    `json:"timing"`
	Logprobs []Logprob `json:"logprobs,omitempty"`
}

// transport carries every request the experiment makes.
var transport http.RoundTripper = http.DefaultTransport

// Client talks to one Ollama server.
type Client struct {
	BaseURL  string
	Settings Settings
}

// NewClient is not built yet.
func NewClient(baseURL string, s Settings) *Client { return &Client{BaseURL: baseURL, Settings: s} }

// Chat is not built yet.
func (c *Client) Chat(ctx context.Context, system, user string, format json.RawMessage) (Reply, error) {
	return Reply{}, errNotBuilt
}

// ServerFacts is what the run records about the server and the model.
type ServerFacts struct {
	Version       string `json:"ollama_version"`
	Model         string `json:"model"`
	Digest        string `json:"model_digest"`
	ParameterSize string `json:"parameter_size"`
	Quantization  string `json:"quantization"`
}

// Facts is not built yet.
func (c *Client) Facts(ctx context.Context) (ServerFacts, error) { return ServerFacts{}, errNotBuilt }
