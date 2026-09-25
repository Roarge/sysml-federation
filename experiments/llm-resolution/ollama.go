package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
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

// DefaultSettings match the server described for this experiment: the 14B
// coder model at 4-bit, with the context capped at 8,192 tokens so it stays in
// the GPU's 12 GB, and room for a final answer of 512 tokens.
func DefaultSettings() Settings {
	return Settings{Model: "qwen2.5-coder:14b", NumCtx: 8192, NumPredict: 512, Temperature: 0, Seed: 42, KeepAlive: "24h", Logprobs: true}
}

// ---- the replies asked for ------------------------------------------------------

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

// TestAnswer is the final answer about a test. The links come first, so the
// evidence and the reason are written after the answer they explain.
type TestAnswer struct {
	Links      []LinkAnswer `json:"links"`
	Evidence   []string     `json:"evidence"`
	Reason     string       `json:"reason"`
	Mismatches []Mismatch   `json:"mismatches"`
}

// IncidentAnswer is the final account of an incident. The cited elements,
// files and path come before the sentence that explains them.
type IncidentAnswer struct {
	Cause        []string   `json:"cause"`
	Mechanism    []string   `json:"mechanism"`
	Code         []string   `json:"code"`
	Consequences []string   `json:"consequences"`
	Path         []string   `json:"path"`
	Why          string     `json:"why"`
	Mismatches   []Mismatch `json:"mismatches"`
}

// ---- the JSON schemas, as typed values -------------------------------------------

type strSchema struct {
	Type string `json:"type"`
}

type enumSchema struct {
	Type string   `json:"type"`
	Enum []string `json:"enum"`
}

type strListSchema struct {
	Type     string    `json:"type"`
	Items    strSchema `json:"items"`
	MaxItems int       `json:"maxItems,omitempty"`
}

type exposureSchema struct {
	Type       string `json:"type"`
	Properties struct {
		ID   strSchema `json:"id"`
		Note strSchema `json:"note"`
	} `json:"properties"`
	Required []string `json:"required"`
}

type exposureListSchema struct {
	Type     string         `json:"type"`
	Items    exposureSchema `json:"items"`
	MaxItems int            `json:"maxItems"`
}

type stepSchema struct {
	Type       string `json:"type"`
	Properties struct {
		Viewpoint strSchema          `json:"viewpoint"`
		Expose    exposureListSchema `json:"expose"`
		Prune     strListSchema      `json:"prune"`
		Tool      enumSchema         `json:"tool"`
		Arg       strSchema          `json:"arg"`
		Arg2      strSchema          `json:"arg2"`
	} `json:"properties"`
	Required []string `json:"required"`
}

type linkSchema struct {
	Type       string `json:"type"`
	Properties struct {
		ID       strSchema  `json:"id"`
		Relation enumSchema `json:"relation"`
	} `json:"properties"`
	Required []string `json:"required"`
}

type linkListSchema struct {
	Type     string     `json:"type"`
	Items    linkSchema `json:"items"`
	MaxItems int        `json:"maxItems"`
}

type mismatchSchema struct {
	Type       string `json:"type"`
	Properties struct {
		ID          strSchema `json:"id"`
		ModelSays   strSchema `json:"model_says"`
		SystemShows strSchema `json:"system_shows"`
		Why         strSchema `json:"why"`
	} `json:"properties"`
	Required []string `json:"required"`
}

type mismatchListSchema struct {
	Type     string         `json:"type"`
	Items    mismatchSchema `json:"items"`
	MaxItems int            `json:"maxItems"`
}

type testFinalSchema struct {
	Type       string `json:"type"`
	Properties struct {
		Links      linkListSchema     `json:"links"`
		Evidence   strListSchema      `json:"evidence"`
		Reason     strSchema          `json:"reason"`
		Mismatches mismatchListSchema `json:"mismatches"`
	} `json:"properties"`
	Required []string `json:"required"`
}

type incidentFinalSchema struct {
	Type       string `json:"type"`
	Properties struct {
		Cause        strListSchema      `json:"cause"`
		Mechanism    strListSchema      `json:"mechanism"`
		Code         strListSchema      `json:"code"`
		Consequences strListSchema      `json:"consequences"`
		Path         strListSchema      `json:"path"`
		Why          strSchema          `json:"why"`
		Mismatches   mismatchListSchema `json:"mismatches"`
	} `json:"properties"`
	Required []string `json:"required"`
}

var (
	str     = strSchema{Type: "string"}
	strList = func(n int) strListSchema { return strListSchema{Type: "array", Items: str, MaxItems: n} }
)

func mismatchList() mismatchListSchema {
	var m mismatchSchema
	m.Type = "object"
	m.Properties.ID, m.Properties.ModelSays, m.Properties.SystemShows, m.Properties.Why = str, str, str, str
	m.Required = []string{"id", "model_says", "system_shows", "why"}
	return mismatchListSchema{Type: "array", Items: m, MaxItems: 3}
}

func mustMarshal(data []byte, err error) json.RawMessage {
	if err != nil {
		panic(err)
	}
	return data
}

// StepSchema admits one tool call and up to three exposures and prunings.
func StepSchema() json.RawMessage {
	var s stepSchema
	s.Type = "object"
	s.Properties.Viewpoint = str
	var ex exposureSchema
	ex.Type = "object"
	ex.Properties.ID, ex.Properties.Note = str, str
	ex.Required = []string{"id", "note"}
	s.Properties.Expose = exposureListSchema{Type: "array", Items: ex, MaxItems: 3}
	s.Properties.Prune = strList(3)
	s.Properties.Tool = enumSchema{Type: "string", Enum: ToolNames}
	s.Properties.Arg, s.Properties.Arg2 = str, str
	s.Required = []string{"viewpoint", "expose", "prune", "tool", "arg", "arg2"}
	return mustMarshal(json.Marshal(s))
}

// TestFinalSchema asks for the links, then the evidence, the reason and any
// suspected mismatches.
func TestFinalSchema() json.RawMessage {
	var s testFinalSchema
	s.Type = "object"
	var l linkSchema
	l.Type = "object"
	l.Properties.ID = str
	l.Properties.Relation = enumSchema{Type: "string", Enum: []string{"verifies", "exercises", "other"}}
	l.Required = []string{"id", "relation"}
	s.Properties.Links = linkListSchema{Type: "array", Items: l, MaxItems: 5}
	s.Properties.Evidence = strList(5)
	s.Properties.Reason = str
	s.Properties.Mismatches = mismatchList()
	s.Required = []string{"links", "evidence", "reason", "mismatches"}
	return mustMarshal(json.Marshal(s))
}

// IncidentFinalSchema asks for the account of an incident.
func IncidentFinalSchema() json.RawMessage {
	var s incidentFinalSchema
	s.Type = "object"
	s.Properties.Cause, s.Properties.Mechanism = strList(5), strList(5)
	s.Properties.Code, s.Properties.Consequences = strList(5), strList(10)
	s.Properties.Path = strList(10)
	s.Properties.Why = str
	s.Properties.Mismatches = mismatchList()
	s.Required = []string{"cause", "mechanism", "code", "consequences", "path", "why", "mismatches"}
	return mustMarshal(json.Marshal(s))
}

// ParseStep reads one step reply.
func ParseStep(raw string) (Step, error) {
	var s Step
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return s, fmt.Errorf("the reply isn't the JSON asked for: %w", err)
	}
	s.Tool = strings.TrimSpace(s.Tool)
	if !slices.Contains(ToolNames, s.Tool) {
		return s, fmt.Errorf("there is no tool %q", s.Tool)
	}
	return s, nil
}

// ParseTestAnswer reads the final answer about a test.
func ParseTestAnswer(raw string) (TestAnswer, error) {
	var a TestAnswer
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		return a, fmt.Errorf("the reply isn't the JSON asked for: %w", err)
	}
	return a, nil
}

// ParseIncidentAnswer reads the final account of an incident.
func ParseIncidentAnswer(raw string) (IncidentAnswer, error) {
	var a IncidentAnswer
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		return a, fmt.Errorf("the reply isn't the JSON asked for: %w", err)
	}
	return a, nil
}

// ---- the chat call ----------------------------------------------------------------

// Timing is what Ollama reports about one call.
type Timing struct {
	PromptTokens int   `json:"prompt_tokens"`
	OutputTokens int   `json:"output_tokens"`
	TotalNS      int64 `json:"total_ns"`
	PromptNS     int64 `json:"prompt_ns"`
	OutputNS     int64 `json:"output_ns"`
}

// Logprob is one generated token with its log probability, kept for the
// first tokens of each reply only.
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

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatOptions struct {
	NumCtx      int     `json:"num_ctx"`
	NumPredict  int     `json:"num_predict"`
	Temperature float64 `json:"temperature"`
	Seed        int     `json:"seed"`
}

type chatRequest struct {
	Model       string          `json:"model"`
	Messages    []chatMessage   `json:"messages"`
	Stream      bool            `json:"stream"`
	Format      json.RawMessage `json:"format"`
	Options     chatOptions     `json:"options"`
	KeepAlive   string          `json:"keep_alive"`
	Logprobs    bool            `json:"logprobs,omitempty"`
	TopLogprobs int             `json:"top_logprobs,omitempty"`
}

type chatResponse struct {
	Model   string      `json:"model"`
	Message chatMessage `json:"message"`
	Done    bool        `json:"done"`
	Error   string      `json:"error"`

	TotalDuration      int64     `json:"total_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int64     `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
	Logprobs           []Logprob `json:"logprobs"`
}

// transport carries every request the experiment makes. Tests replace it to
// see where requests go.
var transport http.RoundTripper = http.DefaultTransport

// Client talks to one Ollama server.
type Client struct {
	BaseURL  string
	HTTP     *http.Client
	Settings Settings
}

// NewClient returns a client for the server at baseURL. It never talks to
// any other address (EXP-SR-12).
func NewClient(baseURL string, s Settings) *Client {
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		HTTP:     &http.Client{Transport: transport, Timeout: 10 * time.Minute},
		Settings: s,
	}
}

// Chat sends one system and one user message with the reply schema given,
// and returns the reply's text as it came. It retries twice on a network
// error, never on a reply the server gave. Reading the text is the caller's
// business, so a reply that isn't the JSON asked for is still a reply.
func (c *Client) Chat(ctx context.Context, system, user string, format json.RawMessage) (Reply, error) {
	req := chatRequest{
		Model:     c.Settings.Model,
		Messages:  []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
		Format:    format,
		KeepAlive: c.Settings.KeepAlive,
		Options: chatOptions{
			NumCtx: c.Settings.NumCtx, NumPredict: c.Settings.NumPredict,
			Temperature: c.Settings.Temperature, Seed: c.Settings.Seed,
		},
		Logprobs: c.Settings.Logprobs,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return Reply{}, err
	}
	var (
		resp chatResponse
		last error
	)
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return Reply{}, ctx.Err()
			case <-time.After(time.Duration(attempt*5) * time.Second):
			}
		}
		resp, last = c.post(ctx, "/api/chat", body)
		var netErr *transportError
		if last == nil || !errors.As(last, &netErr) {
			break
		}
	}
	if last != nil {
		return Reply{}, last
	}
	r := Reply{
		Raw: resp.Message.Content,
		Timing: Timing{
			PromptTokens: resp.PromptEvalCount, OutputTokens: resp.EvalCount,
			TotalNS: resp.TotalDuration, PromptNS: resp.PromptEvalDuration, OutputNS: resp.EvalDuration,
		},
	}
	if len(resp.Logprobs) > 0 {
		r.Logprobs = resp.Logprobs[:min(16, len(resp.Logprobs))]
	}
	return r, nil
}

type transportError struct{ err error }

func (e *transportError) Error() string { return e.err.Error() }
func (e *transportError) Unwrap() error { return e.err }

func (c *Client) post(ctx context.Context, path string, body []byte) (chatResponse, error) {
	var out chatResponse
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return out, &transportError{err}
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return out, &transportError{err}
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, fmt.Errorf("%s answered %s with something that isn't JSON: %.200s", path, res.Status, data)
	}
	if res.StatusCode != http.StatusOK || out.Error != "" {
		return out, fmt.Errorf("%s answered %s: %s", path, res.Status, out.Error)
	}
	return out, nil
}

// ---- server facts recorded with a run ------------------------------------------

type versionResponse struct {
	Version string `json:"version"`
}

type tagsResponse struct {
	Models []tagModel `json:"models"`
}

type tagModel struct {
	Name    string     `json:"name"`
	Digest  string     `json:"digest"`
	Size    int64      `json:"size"`
	Details tagDetails `json:"details"`
}

type tagDetails struct {
	Family            string `json:"family"`
	ParameterSize     string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

// ServerFacts is what the run records about the server and the model.
type ServerFacts struct {
	Version       string `json:"ollama_version"`
	Model         string `json:"model"`
	Digest        string `json:"model_digest"`
	ParameterSize string `json:"parameter_size"`
	Quantization  string `json:"quantization"`
}

// Facts asks the server for its version and the model's digest, and fails if
// the model isn't there.
func (c *Client) Facts(ctx context.Context) (ServerFacts, error) {
	f := ServerFacts{Model: c.Settings.Model}
	var v versionResponse
	if err := c.getJSON(ctx, "/api/version", func(d []byte) error { return json.Unmarshal(d, &v) }); err != nil {
		return f, err
	}
	f.Version = v.Version
	var tags tagsResponse
	if err := c.getJSON(ctx, "/api/tags", func(d []byte) error { return json.Unmarshal(d, &tags) }); err != nil {
		return f, err
	}
	for _, m := range tags.Models {
		if m.Name == c.Settings.Model || strings.TrimSuffix(m.Name, ":latest") == c.Settings.Model {
			f.Digest = m.Digest
			f.ParameterSize = m.Details.ParameterSize
			f.Quantization = m.Details.QuantizationLevel
			return f, nil
		}
	}
	var names []string
	for _, m := range tags.Models {
		names = append(names, m.Name)
	}
	return f, fmt.Errorf("the server has no model %q (it has: %s)", c.Settings.Model, strings.Join(names, ", "))
}

func (c *Client) getJSON(ctx context.Context, path string, decode func([]byte) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("can't reach %s: %w", c.BaseURL, err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", path, res.Status)
	}
	return decode(data)
}
