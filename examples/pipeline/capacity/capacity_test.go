package capacity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Roarge/sysml-federation/examples/pipeline/capacity/flow"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

var names = flow.Names{Quantity: "capacity", Attribute: "throughput"}

// pipelineRep is the representation the router builds for PIPE-P1 from the
// adapter's answer, with the five throughputs given.
func pipelineRep(ingest, parse, indexA, indexB, serve float64) string {
	child := func(id, name string, v float64) string {
		return fmt.Sprintf(`{"id":%q,"name":%q,"attributes":[{"name":"throughput","value":%v},{"name":"capacity","value":null}]}`, id, name, v)
	}
	return `{"__typename":"Part","id":"PIPE-P1","name":"pipeline","attributes":[{"name":"capacity","value":null}],"parts":[` +
		strings.Join([]string{child("PIPE-S1", "ingest", ingest), child("PIPE-S2", "parse", parse), child("PIPE-S3", "indexA", indexA),
			child("PIPE-S4", "indexB", indexB), child("PIPE-S5", "serve", serve)}, ",") +
		`],"connections":[{"from":"PIPE-S1","to":"PIPE-S2"},{"from":"PIPE-S2","to":"PIPE-S3"},{"from":"PIPE-S2","to":"PIPE-S4"},{"from":"PIPE-S3","to":"PIPE-S5"},{"from":"PIPE-S4","to":"PIPE-S5"}]}`
}

func requirementRep(quantity string, limit float64, subject, vc string) string {
	return fmt.Sprintf(`{"__typename":"Requirement","id":"PIPE-R1","quantity":%q,"comparison":"GE","limit":%v,"subject":%s,"verifiedBy":[%s]}`, quantity, limit, subject, vc)
}

func post(t *testing.T, h http.Handler, query, variables string) []byte {
	t.Helper()
	body := fmt.Sprintf(`{"query":%q,"variables":%s}`, query, variables)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	assert.Equal(t, rec.Code, http.StatusOK)
	return rec.Body.Bytes()
}

const entities = `query($reps: [_Any!]!) { _entities(representations: $reps) { ... on Part { capacity bottleneck { id } } ... on Requirement { verdict verdictReason } } }`

type entity struct {
	Capacity   *float64 `json:"capacity"`
	Bottleneck []struct {
		ID string `json:"id"`
	} `json:"bottleneck"`
	Verdict       string `json:"verdict"`
	VerdictReason string `json:"verdictReason"`
}

type answer struct {
	Data struct {
		Entities []entity `json:"_entities"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type state struct{ ingest, parse, indexA, indexB, serve float64 }

type expected struct {
	Capacity   float64
	Bottleneck []string
	Verdict    string
	Reason     string
}

func TestSR29_EntityResolversFollowTheWorkedExample(t *testing.T) {
	h := Handler(names)
	cases := []tabletest.Case[state, expected]{
		{Name: "shipped", In: state{2000, 1200, 700, 700, 1800}, Want: expected{1200, []string{"PIPE-S2"}, "FAIL", "capacity 1200 against 1500, limited by parse"}},
		{Name: "parse to 1700", In: state{2000, 1700, 700, 700, 1800}, Want: expected{1400, []string{"PIPE-S3", "PIPE-S4"}, "FAIL", "capacity 1400 against 1500, limited by indexA, indexB"}},
		{Name: "then indexA to 900", In: state{2000, 1700, 900, 700, 1800}, Want: expected{1600, []string{"PIPE-S3", "PIPE-S4"}, "PASS", "capacity 1600 against 1500, limited by indexA, indexB"}},
	}
	tabletest.Run(t, cases, func(t *testing.T, s state) expected {
		part := pipelineRep(s.ingest, s.parse, s.indexA, s.indexB, s.serve)
		raw := post(t, h, entities, `{"reps":[`+part+`,`+requirementRep("capacity", 1500, part, "")+`]}`)
		var a answer
		assert.NoError(t, json.Unmarshal(raw, &a))
		assert.Len(t, a.Errors, 0)
		assert.Len(t, a.Data.Entities, 2)
		p, r := a.Data.Entities[0], a.Data.Entities[1]
		var ids []string
		for _, b := range p.Bottleneck {
			ids = append(ids, b.ID)
		}
		return expected{*p.Capacity, ids, r.Verdict, r.VerdictReason}
	})
}

func TestSR30_LatencyIsInconclusiveThroughTheEntity(t *testing.T) {
	part := pipelineRep(2000, 1200, 700, 700, 1800)
	raw := post(t, Handler(names), entities, `{"reps":[`+requirementRep("latency", 200, part, `{"id":"PIPE-VC1","shortName":"PIPE-VC1"}`)+`]}`)
	var a answer
	assert.NoError(t, json.Unmarshal(raw, &a))
	assert.Equal(t, a.Data.Entities[0].Verdict, "INCONCLUSIVE")
	assert.Equal(t, a.Data.Entities[0].VerdictReason, "PIPE-VC1 is declared and no service runs it")
}

func TestLeafCapacityIsItsOwnAttribute(t *testing.T) {
	raw := post(t, Handler(names), entities, `{"reps":[{"__typename":"Part","id":"PIPE-S2","name":"parse","attributes":[{"name":"throughput","value":1200}],"parts":[],"connections":[]}]}`)
	var a answer
	assert.NoError(t, json.Unmarshal(raw, &a))
	assert.Equal(t, *a.Data.Entities[0].Capacity, 1200)
	assert.Len(t, a.Data.Entities[0].Bottleneck, 0)
}

// TestSR32_NoStore: two services built apart answer alike, because nothing
// is held between requests.
func TestSR32_NoStore(t *testing.T) {
	part := pipelineRep(2000, 1700, 700, 700, 1800)
	vars := `{"reps":[` + part + `]}`
	first := post(t, Handler(names), entities, vars)
	second := post(t, Handler(names), entities, vars)
	assert.Equal(t, string(first), string(second))
}

// selfImport is where this service keeps its pure code in this repository. A
// file that reads flow.Names has to import that path, which is an address
// rather than a word the service speaks, so the scan below removes the import
// line and nothing besides. The same path in a string constant or in a comment
// still counts.
const selfImport = "github.com/Roarge/sysml-federation/examples/pipeline/capacity/flow"

// TestSR31_NoExampleWordsInTheService: the hand-written sources of the
// service and its flow package carry the two configured names only through
// flow.Names, never as literals, and no other word of the example.
func TestSR31_NoExampleWordsInTheService(t *testing.T) {
	words := []string{"PIPE", "pipeline", "ingest", "indexA", "indexB", "throughput", "Server", "latency"}
	checked := 0
	for _, dir := range []string{".", "flow"} {
		listing, err := os.ReadDir(dir)
		for _, e := range assert.Must(t, listing, err) {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(dir, name))
			data := assert.Must(t, raw, err)
			if bytes.HasPrefix(data, []byte("// Code generated")) {
				continue
			}
			checked++
			data = bytes.ReplaceAll(data, []byte("\t\""+selfImport+"\"\n"), nil)
			for _, w := range words {
				if bytes.Contains(data, []byte(w)) {
					t.Errorf("%s/%s mentions %q", dir, name, w)
				}
			}
		}
	}
	// The six hand-written sources are the floor. A listing that yields fewer
	// is reading the wrong place and would pass for the wrong reason.
	assert.True(t, checked >= 6, "at least six hand-written service files were scanned")
}

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler(names).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, rec.Code, http.StatusOK)
}
