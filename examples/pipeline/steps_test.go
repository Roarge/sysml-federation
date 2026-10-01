package pipeline

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/adapter/projection"
	"github.com/Roarge/sysml-federation/adapter/serve"
	"github.com/Roarge/sysml-federation/examples/pipeline/document"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the adapter and the document service as they ship$`, w.bothServices)
	sc.Given(`^parse's throughput has been set to 1700 and PIPE-R2 excluded from the document$`, w.bothAreEdited)
	sc.When(`^both services are started again$`, w.bothServices)
	sc.When(`^the adapter's resetModel and the document service's resetDocument are called$`, w.bothAreReset)
	sc.Then(`^parse's throughput is 1200 and PIPE-R2 is numbered 2$`, w.theShippedState)
	sc.Then(`^both versions are 1$`, func() { w.bothVersionsAre(1) })
	sc.Then(`^both versions are 3, since a reset moves a version on and never back$`, func() { w.bothVersionsAre(3) })
	sc.Given(`^the document service's sources, its tree package among them$`, func() {})
	sc.When(`^their imports are followed$`, w.theDocumentImportsAreFollowed)
	sc.Then(`^none leads into the adapter$`, w.noneIsFound)
	sc.Given(`^the three services behind the router$`, func() {})
	sc.When(`^their imports and hand-written sources are read$`, w.theServicesAreRead)
	sc.Then(`^none imports a package of another$`, w.noneImportsAnother)
	sc.Then(`^none holds an address, is handed one by a flag, reads the environment or opens a connection$`, w.noWayOut)
	sc.Given(`^the committed configuration and the three schema files$`, w.theConfiguration)
	sc.When(`^each schema file is compared with the schema the configuration embeds for it$`, w.eachSchemaIsCompared)
	sc.Then(`^all three match$`, w.noneIsFound)
	sc.Then(`^a schema changed by one character is reported as drift$`, w.aChangedSchemaDrifts)
	sc.Given(`^the composed configuration the router is started with$`, func() {})
	sc.When(`^the schema it serves is read from it$`, w.theServedSchemaIsRead)
	sc.Then(`^it carries types that only the model, only the capacity service and only the document service declare$`, w.itCarriesAllThree)
	sc.Given(`^the example's README$`, func() {})
	sc.When(`^its verification record is read$`, w.theRecordIsRead)
	sc.Then(`^its rows for the OMG pilot implementation name the release and the kernel$`, w.thePilotRowsNameTheRelease)
	sc.Then(`^its rows for OpenSysML name the version$`, w.theOpenSysMLRowsNameTheVersion)
}

// world is what one scenario of this package knows.
type world struct {
	t         *testing.T
	store     *projection.Store
	doc       *document.Service
	model     *client.Client
	documents *client.Client
	found     []string
	imports   []string
	waysOut   []string
	cfg       routerConfig
	served    string
	rows      [][]string
}

// bothServices starts the adapter on the shipped model and the document
// service on its shipped tree, which is all a restart of either does.
func (w *world) bothServices() {
	loaded, err := projection.Load("model.sysml")
	w.store = assert.Must(w.t, loaded, err)
	created, err := document.New()
	w.doc = assert.Must(w.t, created, err)
	w.model = client.New(serve.Handler(w.store), client.Path("/graphql"))
	w.documents = client.New(document.Handler(w.doc), client.Path("/graphql"))
}

func (w *world) bothAreEdited() {
	var out mutationOut
	w.model.MustPost(`mutation { setAttribute(partId: "PIPE-S2", name: "throughput", value: 1700) { id } }`, &map[string]struct{ ID string }{})
	w.documents.MustPost(`mutation { excludeRequirement(requirementId: "PIPE-R2") { version } }`, &out)
}

func (w *world) bothAreReset() {
	var out mutationOut
	w.model.MustPost(`mutation { resetModel { version } }`, &out)
	w.documents.MustPost(`mutation { resetDocument { version } }`, &out)
}

func (w *world) theShippedState() {
	var part struct {
		Part struct {
			Attributes []struct {
				Name  string
				Value *float64
			}
		}
	}
	w.model.MustPost(`{ part(id: "PIPE-S2") { attributes { name value } } }`, &part)
	throughput := "none"
	for _, a := range part.Part.Attributes {
		if a.Name == "throughput" && a.Value != nil {
			throughput = strconv.FormatFloat(*a.Value, 'f', -1, 64)
		}
	}
	assert.Equal(w.t, throughput, "1200")
	var doc struct {
		Document struct {
			Nodes []struct {
				Requirement *struct {
					ID             string
					DocumentNumber *string
				}
			}
		}
	}
	w.documents.MustPost(`{ document { nodes { requirement { id documentNumber } } } }`, &doc)
	number := "none"
	for _, n := range doc.Document.Nodes {
		if n.Requirement != nil && n.Requirement.ID == "PIPE-R2" && n.Requirement.DocumentNumber != nil {
			number = *n.Requirement.DocumentNumber
		}
	}
	assert.Equal(w.t, number, "2")
}

func (w *world) bothVersionsAre(version int) {
	assert.Equal(w.t, w.store.Version(), version)
	assert.Equal(w.t, w.doc.Version(), version)
}

func (w *world) theDocumentImportsAreFollowed() {
	walkSources(w.t, documentService, parser.ImportsOnly, func(path string, file *ast.File) {
		for _, imported := range importPaths(w.t, path, file) {
			if adapterService.owns(imported) {
				w.found = append(w.found, path+": imports "+imported)
			}
		}
	})
}

func (w *world) noneIsFound() {
	for _, found := range w.found {
		w.t.Error(found)
	}
}

func (w *world) theServicesAreRead() {
	w.imports, w.waysOut = crossImports(w.t), waysOut(w.t)
}

func (w *world) noneImportsAnother() {
	for _, found := range w.imports {
		w.t.Error(found)
	}
}

func (w *world) noWayOut() {
	for _, found := range w.waysOut {
		w.t.Error(found)
	}
}

func (w *world) theConfiguration() { w.cfg = loadConfig(w.t) }

// drifted reports whether a schema differs from the one the configuration
// embeds for its subgraph.
func (w *world) drifted(name, schema string) bool {
	return datasourceOf(w.t, w.cfg, name).CustomGraphql.Federation.ServiceSdl != schema
}

func (w *world) eachSchemaIsCompared() {
	for name, file := range schemaFiles {
		data, err := os.ReadFile(file)
		if w.drifted(name, string(assert.Must(w.t, data, err))) {
			w.found = append(w.found, file+" differs from the schema config.json embeds for "+name)
		}
	}
}

func (w *world) aChangedSchemaDrifts() {
	data, err := os.ReadFile(schemaFiles["capacity"])
	schema := string(assert.Must(w.t, data, err))
	changed := strings.Replace(schema, "Verdict", "Verdikt", 1)
	assert.True(w.t, changed != schema, "the schema was changed")
	assert.True(w.t, w.drifted("capacity", changed), "a changed schema is reported as drift")
}

// servedSchema is the one field of the configuration this scenario reads.
type servedSchema struct {
	EngineConfig struct {
		GraphqlSchema string `json:"graphqlSchema"`
	} `json:"engineConfig"`
}

func (w *world) theServedSchemaIsRead() {
	var cfg servedSchema
	data, err := os.ReadFile(configPath)
	assert.NoError(w.t, json.Unmarshal(assert.Must(w.t, data, err), &cfg))
	w.served = cfg.EngineConfig.GraphqlSchema
	assert.True(w.t, w.served != "", "the configuration carries the schema the router serves")
}

var declaredRE = regexp.MustCompile(`(?m)^(type|enum) (\w+)`)

// declared returns the types and enumerations a schema declares, as "type X"
// or "enum X".
func declared(schema string) []string {
	var found []string
	for _, m := range declaredRE.FindAllStringSubmatch(schema, -1) {
		found = append(found, m[1]+" "+m[2])
	}
	return found
}

func (w *world) itCarriesAllThree() {
	own := make(map[string][]string)
	for name, file := range schemaFiles {
		data, err := os.ReadFile(file)
		own[name] = declared(string(assert.Must(w.t, data, err)))
	}
	served := declared(w.served)
	for name, types := range own {
		var only []string
		for _, typ := range types {
			shared := false
			for other, theirs := range own {
				shared = shared || (other != name && slices.Contains(theirs, typ))
			}
			if !shared && !strings.HasSuffix(typ, " Query") && !strings.HasSuffix(typ, " Mutation") && !strings.HasSuffix(typ, " Subscription") {
				only = append(only, typ)
			}
		}
		assert.True(w.t, len(only) > 0, name+" declares a type of its own")
		for _, typ := range only {
			assert.Contains(w.t, served, typ)
		}
	}
}

var recordRowRE = regexp.MustCompile(`(?m)^\| (\d{4}-\d\d-\d\d) \| ([^|]+) \| ([^|]+) \|`)

func (w *world) theRecordIsRead() {
	data, err := os.ReadFile("README.md")
	readme := string(assert.Must(w.t, data, err))
	start := strings.Index(readme, "## Verification record")
	if start < 0 {
		w.t.Error("README.md has no verification record")
		return
	}
	for _, m := range recordRowRE.FindAllStringSubmatch(readme[start:], -1) {
		w.rows = append(w.rows, []string{m[1], strings.TrimSpace(m[2]), strings.TrimSpace(m[3])})
	}
}

// toolRows returns the versions column of the rows naming a tool.
func (w *world) toolRows(tool string) []string {
	var versions []string
	for _, row := range w.rows {
		if row[1] == tool {
			versions = append(versions, row[2])
		}
	}
	return versions
}

func (w *world) thePilotRowsNameTheRelease() {
	versions := w.toolRows("OMG pilot implementation")
	assert.True(w.t, len(versions) > 0, "the record has rows for the OMG pilot implementation")
	for _, v := range versions {
		assert.True(w.t, strings.Contains(v, "release ") && strings.Contains(v, "kernel "), "the row names the release and the kernel: "+v)
	}
}

var versionRE = regexp.MustCompile(`v\d+\.\d+\.\d+`)

func (w *world) theOpenSysMLRowsNameTheVersion() {
	versions := w.toolRows("OpenSysML")
	assert.True(w.t, len(versions) > 0, "the record has rows for OpenSysML")
	for _, v := range versions {
		assert.True(w.t, versionRE.MatchString(v), "the row names the version: "+v)
	}
}
