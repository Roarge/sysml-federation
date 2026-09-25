package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
// the GPU's 12 GB.
func DefaultSettings() Settings {
	return Settings{Model: "qwen2.5-coder:14b", NumCtx: 8192, NumPredict: 256, Temperature: 0, Seed: 42, KeepAlive: "24h", Logprobs: true}
}

// Answer is the structured reply the model is asked for.
type Answer struct {
	Requirement string   `json:"requirement"`
	Evidence    []string `json:"evidence"`
	Reason      string   `json:"reason"`
}

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

// --- request and response shapes of POST /api/chat --------------------------

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

type stringProp struct {
	Type string `json:"type"`
}

type enumProp struct {
	Type string   `json:"type"`
	Enum []string `json:"enum"`
}

type arrayProp struct {
	Type     string     `json:"type"`
	Items    stringProp `json:"items"`
	MaxItems int        `json:"maxItems"`
}

type answerProps struct {
	Requirement enumProp   `json:"requirement"`
	Evidence    arrayProp  `json:"evidence"`
	Reason      stringProp `json:"reason"`
}

// answerSchema is the JSON schema passed as `format`, so the reply must be an
// Answer whose requirement is one of the keys or "none".
type answerSchema struct {
	Type       string      `json:"type"`
	Properties answerProps `json:"properties"`
	Required   []string    `json:"required"`
}

func newAnswerSchema(keys []string) answerSchema {
	enum := append(append([]string{}, keys...), "none")
	return answerSchema{
		Type: "object",
		Properties: answerProps{
			Requirement: enumProp{Type: "string", Enum: enum},
			Evidence:    arrayProp{Type: "array", Items: stringProp{Type: "string"}, MaxItems: 5},
			Reason:      stringProp{Type: "string"},
		},
		Required: []string{"requirement", "evidence", "reason"},
	}
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Format      answerSchema  `json:"format"`
	Options     chatOptions   `json:"options"`
	KeepAlive   string        `json:"keep_alive"`
	Logprobs    bool          `json:"logprobs,omitempty"`
	TopLogprobs int           `json:"top_logprobs,omitempty"`
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

// NewClient returns a client for the server at baseURL, asking with the given
// settings and a reply schema built from the requirement keys. It never
// talks to any other address (EXP-SR-12).
func NewClient(baseURL string, s Settings, keys []string) *Client {
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		HTTP:     &http.Client{Transport: transport, Timeout: 10 * time.Minute},
		Settings: s,
		Schema:   newAnswerSchema(keys),
	}
}

// Client talks to one Ollama server.
type Client struct {
	BaseURL  string
	HTTP     *http.Client
	Settings Settings
	Schema   answerSchema
}

// Reply is one call's outcome.
type Reply struct {
	Answer   Answer    `json:"answer"`
	Raw      string    `json:"raw"`
	Timing   Timing    `json:"timing"`
	Logprobs []Logprob `json:"logprobs,omitempty"`
}

// Chat sends one system and one user message and parses the answer. It
// retries twice on a network error, never on a reply the server gave.
func (c *Client) Chat(ctx context.Context, system, user string) (Reply, error) {
	req := chatRequest{
		Model:     c.Settings.Model,
		Messages:  []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
		Format:    c.Schema,
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
	if err := json.Unmarshal([]byte(resp.Message.Content), &r.Answer); err != nil {
		return r, fmt.Errorf("the reply isn't the JSON asked for: %w", err)
	}
	r.Answer.Requirement = strings.TrimSpace(r.Answer.Requirement)
	return r, nil
}

type transportError struct{ err error }

func (e *transportError) Error() string { return e.err.Error() }
func (e *transportError) Unwrap() error { return e.err }

func (c *Client) post(ctx context.Context, path string, body []byte) (chatResponse, error) {
	var out chatResponse
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+path, bytes.NewReader(body))
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

// --- server facts recorded with a run ------------------------------------------

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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+path, nil)
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
