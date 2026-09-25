package main

// The declarations the tests are written against, before any behaviour
// exists. Each returns the zero value or errNotBuilt, so every test compiles
// and fails until the file that owns it is written.

import (
	"context"
	"errors"
	"io"
	"net/http"
)

var errNotBuilt = errors.New("not built yet")

// ---- corpus.go

type Requirement struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Statement string `json:"statement"`
}

type Test struct {
	ID       string `json:"id"`
	FuncName string `json:"func_name"`
	Name     string `json:"name"`
	Package  string `json:"package"`
	File     string `json:"file"`
	Doc      string `json:"doc"`
	Gold     string `json:"gold,omitempty"`
}

type Corpus struct {
	Requirements []Requirement
	Tests        []Test
}

func LoadCorpus(root string) (Corpus, error)       { return Corpus{}, errNotBuilt }
func FindRepoRoot(start string) (string, error)    { return "", errNotBuilt }
func (c Corpus) Validate() error                   { return errNotBuilt }
func (c Corpus) Keys() []string                    { return nil }
func (c Corpus) Hash() string                      { return "" }
func newTest(funcName, pkg, file, doc string) Test { return Test{} }

// ---- words.go

func terms(s string) []string { return nil }

// ---- baseline.go

const BaselineThreshold = 0.15

type Baseline struct{}

type Ranked struct {
	Key    string   `json:"key"`
	Score  float64  `json:"score"`
	Shared []string `json:"shared"`
}

type BaselineResult struct {
	Test   string   `json:"test"`
	Gold   string   `json:"gold,omitempty"`
	Pick   string   `json:"pick"`
	Top    []Ranked `json:"top"`
	Rank   int      `json:"gold_rank,omitempty"`
	Reason []string `json:"reason"`
}

func NewBaseline(reqs []Requirement) *Baseline    { return &Baseline{} }
func (b *Baseline) Resolve(t Test) BaselineResult { return BaselineResult{} }

// ---- prompt.go

type TestView struct {
	Name    string `json:"name"`
	Package string `json:"package"`
	File    string `json:"file"`
	Doc     string `json:"doc"`
}

func viewOf(t Test) TestView                         { return TestView{} }
func systemPrompt(reqs []Requirement) string         { return "" }
func userPrompt(v TestView) string                   { return "" }
func baseQuestion(c Corpus, t Test) (string, string) { return "", "" }

// ---- ollama.go

type Settings struct {
	Model       string  `json:"model"`
	NumCtx      int     `json:"num_ctx"`
	NumPredict  int     `json:"num_predict"`
	Temperature float64 `json:"temperature"`
	Seed        int     `json:"seed"`
	KeepAlive   string  `json:"keep_alive"`
	Logprobs    bool    `json:"logprobs"`
}

func DefaultSettings() Settings { return Settings{} }

type Answer struct {
	Requirement string   `json:"requirement"`
	Evidence    []string `json:"evidence"`
	Reason      string   `json:"reason"`
}

type Reply struct {
	Answer Answer `json:"answer"`
	Raw    string `json:"raw"`
}

type ServerFacts struct {
	Version string `json:"ollama_version"`
	Model   string `json:"model"`
	Digest  string `json:"model_digest"`
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

type answerSchema struct {
	Type       string      `json:"type"`
	Properties answerProps `json:"properties"`
	Required   []string    `json:"required"`
}

func newAnswerSchema(keys []string) answerSchema { return answerSchema{} }

// transport carries every request the experiment makes. Tests replace it to
// see where requests go.
var transport http.RoundTripper = http.DefaultTransport

type Client struct{}

func NewClient(baseURL string, s Settings, keys []string) *Client { return &Client{} }
func (c *Client) Chat(ctx context.Context, system, user string) (Reply, error) {
	return Reply{}, errNotBuilt
}
func (c *Client) Facts(ctx context.Context) (ServerFacts, error) { return ServerFacts{}, errNotBuilt }

// ---- probes.go

type Variant struct {
	Requirements []Requirement
	View         TestView
	Removed      []string
	Note         string
}

func deletionVariant(reqs []Requirement, t Test, base Answer) Variant { return Variant{} }
func controlVariant(reqs []Requirement, t Test, base Answer, seed int64) Variant {
	return Variant{}
}
func reconstructionVariant(reqs []Requirement, base Answer) Variant { return Variant{} }
func orderVariant(reqs []Requirement, seed int64) Variant           { return Variant{} }

// ---- run.go

func quickSample(tests []Test) []Test  { return nil }
func repeatSample(tests []Test) []Test { return nil }

// ---- results.go

type RunHeader struct {
	Type       string      `json:"type"`
	Commit     string      `json:"commit"`
	Settings   Settings    `json:"settings"`
	Server     ServerFacts `json:"server"`
	CorpusHash string      `json:"corpus_hash"`
}

type CallLine struct {
	Type     string `json:"type"`
	Key      string `json:"key"`
	Probe    string `json:"probe"`
	Test     string `json:"test"`
	Gold     string `json:"gold,omitempty"`
	BasePick string `json:"base_pick,omitempty"`
	User     string `json:"user,omitempty"`
	Reply    *Reply `json:"reply,omitempty"`
	Error    string `json:"error,omitempty"`
	Note     string `json:"note,omitempty"`
}

type Contents struct {
	Header   RunHeader
	Baseline []BaselineResult
	Calls    []CallLine
}

func ReadResults(path string) (Contents, error) { return Contents{}, errNotBuilt }

// ---- report.go

type Rate struct {
	K  int     `json:"k"`
	N  int     `json:"n"`
	Lo float64 `json:"lo"`
	Hi float64 `json:"hi"`
}

type ResolverScore struct {
	Resolver  string `json:"resolver"`
	Proposed  int    `json:"proposed"`
	Correct   int    `json:"correct"`
	Precision Rate   `json:"precision"`
	Recall    Rate   `json:"recall"`
}

type ProbeRate struct {
	Probe   string `json:"probe"`
	Changed Rate   `json:"changed"`
	Skipped int    `json:"skipped"`
}

type Proposal struct {
	Test     string `json:"test"`
	Resolver string `json:"resolver"`
	Pick     string `json:"pick"`
	Reason   string `json:"reason"`
}

type Summary struct {
	Resolvers []ResolverScore `json:"resolvers"`
	Probes    []ProbeRate     `json:"probes"`
	Proposals []Proposal      `json:"proposals"`
}

func wilson(k, n int) (float64, float64) { return 0, 0 }
func Summarise(c Contents) Summary       { return Summary{} }
func (s Summary) Markdown() string       { return "" }

// ---- main.go

func main() {}

func run(args []string, stdout, stderr io.Writer) int { return 1 }
