// Package trace reads the demo's own model and the repository it describes, so
// that a test can fail when the two disagree.
//
// The model under model/ names things that live elsewhere in the repository: a
// decision record, a Go test function, a check file, a published image, a
// service of the check session. Nothing in the modelling language makes those
// names true, and a register that names a file nobody wrote is worse than no
// register at all, because it reads as evidence. The functions here return what
// each side actually carries, and TestSR46_ModelAndRepositoryAgree beside them
// compares the two.
//
// Every function is pure and takes either the text it reads or a root
// directory. The root is what lets the test point a check at a fixture instead
// of at the repository, which is how each check is proved to fail for the
// reason it claims.
//
// The model is read with regular expressions rather than with a parser. The
// patterns sit together in one block below, each beside the form it matches,
// and nothing outside that block reads either side with a pattern of its own. A
// form the model could take and these patterns do not match is a hole in the
// check rather than a pass, which is why the registers are written in the
// shapes named here.
package trace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// The places the two sides live, as paths from the module root. A register the
// model has not grown yet is simply absent, and the test skips the agreement
// that would have read it.
const (
	// ModelDir holds the demo's own model.
	ModelDir = "model"
	// StakeholderStoriesFile holds the stories US-01 to US-19.
	StakeholderStoriesFile = "model/core/stories/stakeholder/stakeholder-stories.sysml"
	// SystemStoriesFile holds the system stories SR-nn and their derivations.
	SystemStoriesFile = "model/core/stories/system/system-stories.sysml"
	// DesignConstraintsFile holds the constraints SC-nn.
	DesignConstraintsFile = "model/core/constraints/design-constraints.sysml"
	// VerificationCasesFile holds one case per system story and constraint.
	VerificationCasesFile = "model/core/verification-validation/verification-cases/verification-cases.sysml"
	// ValidationCasesFile holds one case per stakeholder story.
	ValidationCasesFile = "model/core/verification-validation/validation-cases/validation-cases.sysml"
	// CheckCasesFile holds the inventory of live checks.
	CheckCasesFile = "model/core/verification-validation/check-cases/check-cases.sysml"
	// ViewsFile holds the views, which name the published images.
	ViewsFile = "model/core/views/views.sysml"
	// ComponentsFile holds the logical architecture, its allocations and the
	// session composite.
	ComponentsFile = "model/core/logical-architecture/components/components.sysml"
	// DecisionsDir holds the decision records.
	DecisionsDir = "docs/decisions"
	// ImagesDir holds the published boards and sheets.
	ImagesDir = "docs/img"
	// ChecksDir holds the check project's check files.
	ChecksDir = "checkly/__checks__"
	// ChecksManifestFile holds the inventory the check project reads.
	ChecksManifestFile = "checkly/__checks__/checks.json"
	// ComposeFile starts the check session beside the demo.
	ComposeFile = "checkly/compose.yml"
)

// StatusDone is the status a story carries once its acceptance is met, as
// @StoryMeta { status = StoryStatus::done; } writes it.
const StatusDone = "done"

// ErrNoModuleRoot reports that the walk upwards from the working directory
// found no go.mod, so there is no repository to read.
var ErrNoModuleRoot = errors.New("trace: no go.mod above the working directory")

// The forms the two sides are read for. Each pattern is written beside the
// syntax it matches. Short names are <'...'>, attributes are
// attribute name : Type = value; or attribute :>> name = value;, metadata
// applications are @Name { key = "value"; }, and a quoted string never contains
// a quote.
var (
	// A short name of the light scheme: <'US-04'>, <'SR-46'>, <'SC-07'>.
	shortNameRE = regexp.MustCompile(`<'((?:US|SR|SC)-\d\d)'>`)

	// The one place an identifier is introduced, a story or constraint
	// declaration: requirement <'SR-22'> SR_22_EditsPatchTheSource : UserStory.
	// The #derive a system story carries sits on the line above it.
	declarationRE = regexp.MustCompile(`(?m)^[\t ]*requirement\s+<'((?:US|SR|SC)-\d\d)'>\s+\w+\s*:\s*\w+`)

	// The same declaration with its opening brace, for reading what is inside:
	// requirement <'US-04'> US_04_RaiseTheBottleneck : UserStory {.
	storyRE = regexp.MustCompile(`(?m)^[\t ]*requirement\s+<'((?:US|SR)-\d\d)'>\s+(\w+)\s*:\s*UserStory\s*\{`)

	// The status of a story: @StoryMeta { status = StoryStatus::done; }.
	storyStatusRE = regexp.MustCompile(`@StoryMeta\s*\{\s*status\s*=\s*StoryStatus::(\w+)\s*;\s*\}`)

	// A decision record wherever the model names it, in a @DecisionRecord
	// application or in the prose of a doc: AD-0029.
	decisionIDRE = regexp.MustCompile(`AD-\d{4}`)

	// A decision record's file name: AD-0029-the-demos-own-model-in-sysml-v2.md.
	decisionFileRE = regexp.MustCompile(`^(AD-\d{4})-[\w-]+\.md$`)

	// The "Requirements affected" section of a record, up to the next
	// second-level heading or the end of the file.
	requirementsAffectedRE = regexp.MustCompile(`(?ms)^## Requirements affected[\t ]*$(.*?)(?:^## |\z)`)

	// A requirement as a record writes it in that section: SR-46, SC-07.
	requirementIDRE = regexp.MustCompile(`\b((?:SR|SC)-\d\d)\b`)

	// A Go test of the requirement scheme as the repository declares it:
	// func TestSR22_SetAttributePatchesTextAndProjectionTogether(t *testing.T).
	goTestFuncRE = regexp.MustCompile(`(?m)^func (TestS[RC]\d\d_\w+)\(`)

	// The same name written in prose, which is how the register carries the
	// second of the two functions that share a name.
	goTestNameRE = regexp.MustCompile(`TestS[RC]\d\d_\w+`)

	// An action usage and its opening brace, with a short name or without one:
	// action <'TestSR22_Foo'> patchesTogether { and action refusedInModel {.
	actionRE = regexp.MustCompile(`(?m)^[\t ]*action\s+(?:<'([^']*)'>\s+)?\w+\s*\{`)

	// What an action offers as evidence:
	// @Evidence { kind = "go-test"; location = "adapter/model/patch_test.go"; }.
	evidenceRE = regexp.MustCompile(`@Evidence\s*\{\s*kind\s*=\s*"([^"]*)"\s*;\s*location\s*=\s*"([^"]*)"\s*;\s*\}`)

	// A doc and its text, over as many lines as it takes: doc /* ... */.
	docRE = regexp.MustCompile(`(?s)doc\s*/\*(.*?)\*/`)

	// The objective of a verification case:
	// verify SR_22_EditsPatchTheSource.acceptance;.
	verifyRE = regexp.MustCompile(`verify\s+(SR_\d\d_\w+)\.acceptance`)

	// An allocation in the logical architecture:
	// satisfy SR_01_OneCommand by demo;.
	satisfyRE = regexp.MustCompile(`satisfy\s+(SR_\d\d_\w+)\s+by\b`)

	// A derived end of a derivation connection:
	// end #derive   ::> SR_12_SketchDrawnFromTheWiring;.
	deriveRE = regexp.MustCompile(`end\s+#derive\s+::>\s+(SR_\d\d_\w+)\s*;`)

	// A validation case, one per stakeholder story: verification def VAL_US_04.
	validationCaseRE = regexp.MustCompile(`(?m)^[\t ]*verification\s+def\s+(VAL_US_\d\d)\b`)

	// The same with its opening brace, for reading what one case names.
	validationCaseBlockRE = regexp.MustCompile(`(?m)^[\t ]*verification\s+def\s+VAL_US_\d\d\s*\{`)

	// A published image as a view's doc names it: docs/img/v1-context.png.
	imagePathRE = regexp.MustCompile(`docs/img/([A-Za-z0-9._-]+\.png)`)

	// A check file of the check project, wherever a name reaches the model:
	// stories.spec.ts, api.check.ts, checkly/__checks__/browser.check.ts.
	checkFileRE = regexp.MustCompile(`[A-Za-z0-9._/-]*[A-Za-z0-9._-]+\.(?:spec|check)\.ts`)

	// A quoted value, a string or a short name, which are the two ways the
	// model quotes a file name.
	quotedRE = regexp.MustCompile(`"([^"]*)"|<'([^']*)'>`)

	// A check case of the check register, with a short name or without one:
	// verification def CHK_StoriesSuite {.
	checkCaseRE = regexp.MustCompile(`(?m)^[\t ]*verification\s+def\s+(?:<'[^']*'>\s+)?CHK_\w+\s*\{`)

	// An attribute assignment, in both the forms the model writes:
	// attribute logicalId : String = "story-us01"; and
	// attribute :>> deployed = "whenHostnameSet";.
	attributeRE = regexp.MustCompile(`attribute\s+(?::>>\s+)?(\w+)\s*(?::\s*[\w:]+(?:\s*\[[^\]]*\])?)?\s*=\s*([^;]+);`)

	// A check the check project constructs: new BrowserCheck('story-us01', {.
	constructionRE = regexp.MustCompile(`new\s+\w+\s*\(\s*['"]([^'"]+)['"]`)

	// The session composite of the logical architecture, with its opening
	// brace: part session : CheckSession {.
	sessionRE = regexp.MustCompile(`(?m)^[\t ]*part\s+session\s*:\s*CheckSession\s*\{`)

	// The compose service a part of that composite maps to:
	// attribute composeService : String = "otel-collector";.
	composeServiceRE = regexp.MustCompile(`composeService\s*(?::\s*[\w:]+)?\s*=\s*"([^"]*)"`)

	// A key indented under the top-level services: of a compose file, which is
	// as much of that file as this package reads.
	composeKeyRE = regexp.MustCompile(`^([\t ]+)([A-Za-z0-9._-]+):[\t ]*(?:#.*)?$`)
)

// File is one file read from the repository: its path from the root in slash
// form, which is what a failure names, and its text.
type File struct {
	Path string
	Text string
}

// Block is a region of a model file: the capture groups of the header that
// opened it, and the text between its braces.
type Block struct {
	Header []string
	Body   string
}

// DecisionRecord is one record under docs/decisions: its identifier, its path
// from the root, and the requirements its "Requirements affected" section
// names, in order and without repeats.
type DecisionRecord struct {
	ID                   string
	Path                 string
	RequirementsAffected []string
}

// Story is a requirement usage typed UserStory: the short name that identifies
// it, the name the rest of the model refers to it by, and the status its
// @StoryMeta carries.
type Story struct {
	ShortName string
	Name      string
	Status    string
}

// TestFunc is one Go test of the requirement scheme: its function name and the
// file it is declared in. The file is part of it because two functions share a
// name across two packages, and a set of names alone would hide one of them.
type TestFunc struct {
	Name string
	File string
}

// CheckTuple is one row of the check inventory, as either side writes it. Every
// field is text so that the two sides can be compared as they are written:
// FrequencyMinutes is empty for a check that runs on no schedule, Deployed is
// true, false or whenHostnameSet, and Locations is the locations sorted and
// joined with commas, because the order they are written in means nothing.
type CheckTuple struct {
	LogicalID        string
	Kind             string
	FrequencyMinutes string
	Deployed         string
	Locations        string
}

// ManifestEntry is one entry of the check project's manifest. Two of its fields
// are read as raw JSON because they carry more than one type: deployed is a
// boolean or the string whenHostnameSet, and a check on no schedule carries no
// frequency at all.
type ManifestEntry struct {
	ID               string          `json:"id"`
	Kind             string          `json:"kind"`
	File             string          `json:"file"`
	FrequencyMinutes json.RawMessage `json:"frequencyMinutes"`
	Deployed         json.RawMessage `json:"deployed"`
	Locations        []string        `json:"locations"`
}

// ModuleRoot returns the directory holding the module's go.mod, by walking up
// from the working directory. A test binary runs in the directory of its own
// package, so the walk is what lets a test read files elsewhere in the
// repository without a path relative to itself.
func ModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoModuleRoot
		}
		dir = parent
	}
}

// Exists reports whether rel under root is there. An absent register is the
// ordinary state of a model still being written, not an error.
func Exists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

// Text returns the contents of rel under root, and the empty string when the
// file is absent: a register nobody has written yet names nothing.
func Text(root, rel string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// ModelFiles returns every .sysml file under model/, in path order. The checks
// that are about the model as a whole read all of it, because an identifier may
// be written in any register.
func ModelFiles(root string) ([]File, error) {
	return filesUnder(root, ModelDir, ".sysml")
}

// ShortNames returns every short name of the light scheme the model writes,
// once each and in order.
func ShortNames(files []File) []string {
	seen := make(map[string]bool)
	for _, file := range files {
		for _, id := range captures(file.Text, shortNameRE) {
			seen[id] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

// Declarations counts, per short name, the requirement declarations that
// introduce it. Anything other than one is a disagreement: none means a name
// used and never declared, and more than one means two elements answering to
// the same identifier.
func Declarations(files []File) map[string]int {
	times := make(map[string]int)
	for _, file := range files {
		for _, id := range captures(file.Text, declarationRE) {
			times[id]++
		}
	}
	return times
}

// DecisionIDs returns, per decision record the model names, the model files
// that name it, in path order. A record named several times in one file is
// listed once for that file.
func DecisionIDs(files []File) map[string][]string {
	named := make(map[string][]string)
	for _, file := range files {
		ids := decisionIDRE.FindAllString(file.Text, -1)
		for _, id := range slices.Compact(slices.Sorted(slices.Values(ids))) {
			named[id] = append(named[id], file.Path)
		}
	}
	return named
}

// DecisionRecords returns the records under docs/decisions, in file order.
func DecisionRecords(root string) ([]DecisionRecord, error) {
	dir := filepath.Join(root, filepath.FromSlash(DecisionsDir))
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var records []DecisionRecord
	for _, entry := range entries {
		name := decisionFileRE.FindStringSubmatch(entry.Name())
		if entry.IsDir() || name == nil {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		record := DecisionRecord{ID: name[1], Path: DecisionsDir + "/" + entry.Name()}
		if section := requirementsAffectedRE.FindStringSubmatch(string(raw)); section != nil {
			ids := captures(section[1], requirementIDRE)
			record.RequirementsAffected = slices.Compact(slices.Sorted(slices.Values(ids)))
		}
		records = append(records, record)
	}
	return records, nil
}

// GoTestFunctions returns every Go test of the requirement scheme the module
// declares, with the file it is in. The walk stays inside this module: a
// directory with a go.mod of its own is another module, a testdata directory is
// not compiled, and a dotted directory is not source.
func GoTestFunctions(root string) ([]TestFunc, error) {
	var found []TestFunc
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case entry.IsDir():
			return skipDirectory(root, name, entry)
		case !strings.HasSuffix(entry.Name(), "_test.go"):
			return nil
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		for _, function := range captures(string(raw), goTestFuncRE) {
			found = append(found, TestFunc{Name: function, File: filepath.ToSlash(rel)})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

// GoTestEvidence returns the Go tests the verification register names: one pair
// per action offering evidence of kind go-test. The name is the action's short
// name, or, where it has none, the test name written in its doc, which is how
// the register carries the second of two functions that share a name. A pair
// with an empty name is an action naming no test at all, which the test reports
// rather than passing over.
func GoTestEvidence(text string) []TestFunc {
	var found []TestFunc
	for _, block := range blocks(text, actionRE) {
		evidence := evidenceRE.FindStringSubmatch(block.Body)
		if evidence == nil || evidence[1] != "go-test" {
			continue
		}
		name := block.Header[1]
		if name == "" {
			if doc := docRE.FindStringSubmatch(block.Body); doc != nil {
				name = goTestNameRE.FindString(doc[1])
			}
		}
		found = append(found, TestFunc{Name: name, File: evidence[2]})
	}
	return found
}

// Stories returns the story usages of a register, in the order they are
// written, each with the status its @StoryMeta carries. A story with no
// @StoryMeta carries the empty status, which is not done and is reported as
// such.
func Stories(text string) []Story {
	var found []Story
	for _, block := range blocks(text, storyRE) {
		story := Story{ShortName: block.Header[1], Name: block.Header[2]}
		if status := storyStatusRE.FindStringSubmatch(block.Body); status != nil {
			story.Status = status[1]
		}
		found = append(found, story)
	}
	return found
}

// VerifiedStories returns the system stories whose acceptance a verification
// case verifies.
func VerifiedStories(text string) []string { return captures(text, verifyRE) }

// SatisfiedStories returns the system stories something in the logical
// architecture satisfies.
func SatisfiedStories(text string) []string { return captures(text, satisfyRE) }

// DerivedStories returns the system stories that are a derived end of a
// derivation connection.
func DerivedStories(text string) []string { return captures(text, deriveRE) }

// ValidationCases returns the names of the validation cases a register
// declares, one per stakeholder story.
func ValidationCases(text string) []string { return captures(text, validationCaseRE) }

// ImageNames returns the published images the views name, in the order the docs
// name them and with their repeats kept, because naming one image twice is one
// of the disagreements this package is for.
func ImageNames(text string) []string { return captures(text, imagePathRE) }

// PNGFiles returns the names of the published images, in order.
func PNGFiles(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(ImagesDir)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var found []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			found = append(found, entry.Name())
		}
	}
	return found, nil
}

// CheckFiles returns the names of the check project's check files, the
// *.spec.ts and *.check.ts under checkly/__checks__, in order. The lib
// directory the checks import from is not itself a check.
func CheckFiles(root string) ([]string, error) {
	base := filepath.Join(root, filepath.FromSlash(ChecksDir))
	var found []string
	err := filepath.WalkDir(base, func(name string, entry fs.DirEntry, err error) error {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return nil
		case err != nil:
			return err
		case entry.IsDir():
			if name != base && (entry.Name() == "lib" || entry.Name() == "node_modules") {
				return fs.SkipDir
			}
		case isCheckFile(entry.Name()):
			found = append(found, entry.Name())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(found)
	return found, nil
}

// CheckFileNames returns the check files a register names, read out of its
// quoted strings and short names so that a file named in prose is not mistaken
// for evidence. Only the file name is kept: the model may write a path and the
// project holds the files in one directory.
func CheckFileNames(text string) []string {
	var found []string
	for _, quoted := range quotedValues(text) {
		for _, name := range checkFileRE.FindAllString(quoted, -1) {
			found = append(found, path.Base(name))
		}
	}
	return found
}

// SpecFileCases counts, per suite file, the validation cases that name it. A
// suite belongs to one story, so anything other than one case is a
// disagreement.
func SpecFileCases(text string) map[string]int {
	times := make(map[string]int)
	for _, block := range blocks(text, validationCaseBlockRE) {
		named := CheckFileNames(block.Body)
		for _, name := range slices.Compact(slices.Sorted(slices.Values(named))) {
			if strings.HasSuffix(name, ".spec.ts") {
				times[name]++
			}
		}
	}
	return times
}

// ReadManifest reads the inventory the check project runs from. The file
// carries either the entries as a list or an object whose checks field is that
// list, and both are read, because the shape of the manifest is the check
// project's to choose. An absent file carries no entries.
func ReadManifest(root string) ([]ManifestEntry, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ChecksManifestFile)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []ManifestEntry
	if err := json.Unmarshal(raw, &entries); err == nil {
		return entries, nil
	}
	var wrapped struct {
		Checks []ManifestEntry `json:"checks"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("reading %s: %w", ChecksManifestFile, err)
	}
	return wrapped.Checks, nil
}

// ManifestTuples flattens the manifest entries to the rows the register is
// compared against.
func ManifestTuples(entries []ManifestEntry) []CheckTuple {
	tuples := make([]CheckTuple, 0, len(entries))
	for _, entry := range entries {
		tuples = append(tuples, CheckTuple{
			LogicalID:        entry.ID,
			Kind:             entry.Kind,
			FrequencyMinutes: jsonValue(entry.FrequencyMinutes),
			Deployed:         jsonValue(entry.Deployed),
			Locations:        joinSorted(entry.Locations),
		})
	}
	return tuples
}

// CheckCaseTuples reads the same rows out of the check register, one per CHK_
// case, from the attributes it declares.
func CheckCaseTuples(text string) []CheckTuple {
	var tuples []CheckTuple
	for _, block := range blocks(text, checkCaseRE) {
		values := attributeValues(block.Body)
		tuples = append(tuples, CheckTuple{
			LogicalID:        values["logicalId"],
			Kind:             values["kind"],
			FrequencyMinutes: values["frequencyMinutes"],
			Deployed:         values["deployed"],
			Locations:        values["locations"],
		})
	}
	return tuples
}

// Constructions counts, per logical identifier, the times the check project
// constructs it, as new BrowserCheck('story-us01', {.
func Constructions(root string) (map[string]int, error) {
	files, err := filesUnder(root, ChecksDir, ".check.ts")
	if err != nil {
		return nil, err
	}
	times := make(map[string]int)
	for _, file := range files {
		for _, id := range captures(file.Text, constructionRE) {
			times[id]++
		}
	}
	return times, nil
}

// ComposeServices returns the services a compose file declares. Only the names
// are wanted, so the file is read as text rather than as YAML: a service is a
// key at the first level of indentation under the top-level services:, and the
// list ends at the next key in the first column.
func ComposeServices(text string) []string {
	var (
		found  []string
		inside bool
		indent = -1
	)
	for line := range strings.Lines(text) {
		line = strings.TrimRight(line, "\r\n")
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "" || strings.HasPrefix(trimmed, "#"):
			continue
		case !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t"):
			inside = strings.HasPrefix(trimmed, "services:")
			indent = -1
			continue
		case !inside:
			continue
		}
		key := composeKeyRE.FindStringSubmatch(line)
		if key == nil {
			continue
		}
		width := len(strings.ReplaceAll(key[1], "\t", "    "))
		if indent == -1 {
			indent = width
		}
		if width == indent {
			found = append(found, key[2])
		}
	}
	return found
}

// SessionComposeServices returns the compose services the session composite of
// the logical architecture carries, one per part that maps to a service.
func SessionComposeServices(text string) []string {
	var found []string
	for _, block := range blocks(text, sessionRE) {
		found = append(found, captures(block.Body, composeServiceRE)...)
	}
	return found
}

// blocks returns one block per match of header in text. The header pattern ends
// at the block's opening brace, so the body runs from there to the brace that
// closes it. A group the header did not capture is the empty string, which is
// how an optional short name is read.
func blocks(text string, header *regexp.Regexp) []Block {
	var found []Block
	for _, match := range header.FindAllStringSubmatchIndex(text, -1) {
		groups := make([]string, 0, len(match)/2)
		for i := 0; i < len(match); i += 2 {
			if match[i] < 0 {
				groups = append(groups, "")
				continue
			}
			groups = append(groups, text[match[i]:match[i+1]])
		}
		found = append(found, Block{Header: groups, Body: blockBody(text, match[1])})
	}
	return found
}

// blockBody returns the text from start, which is just inside an opening brace,
// to the brace that closes it. Comments and quoted strings are stepped over, so
// that a brace in a doc or in a quoted value does not close the block. An
// unclosed block runs to the end of the file, which is what a reader of a
// half-written register would expect to see.
func blockBody(text string, start int) string {
	depth := 1
	for i := start; i < len(text); i++ {
		rest := text[i:]
		switch {
		case strings.HasPrefix(rest, "/*"):
			end := strings.Index(rest[2:], "*/")
			if end < 0 {
				return text[start:]
			}
			i += 2 + end + 1
		case strings.HasPrefix(rest, "//"):
			end := strings.IndexByte(rest, '\n')
			if end < 0 {
				return text[start:]
			}
			i += end
		case rest[0] == '"':
			end := strings.IndexByte(rest[1:], '"')
			if end < 0 {
				return text[start:]
			}
			i += 1 + end
		case rest[0] == '{':
			depth++
		case rest[0] == '}':
			depth--
			if depth == 0 {
				return text[start:i]
			}
		}
	}
	return text[start:]
}

// captures returns the first capture group of every match, in the order they
// are written.
func captures(text string, re *regexp.Regexp) []string {
	var found []string
	for _, match := range re.FindAllStringSubmatch(text, -1) {
		found = append(found, match[1])
	}
	return found
}

// quotedValues returns the contents of every quoted string and every short name
// in the text.
func quotedValues(text string) []string {
	var found []string
	for _, match := range quotedRE.FindAllStringSubmatch(text, -1) {
		found = append(found, match[1]+match[2])
	}
	return found
}

// attributeValues reads the attribute assignments of one block and flattens
// each value to the text the other side would write for it.
func attributeValues(body string) map[string]string {
	values := make(map[string]string)
	for _, match := range attributeRE.FindAllStringSubmatch(body, -1) {
		values[match[1]] = attributeValue(match[2])
	}
	return values
}

// attributeValue flattens one value: a quoted string loses its quotes, an
// enumeration literal loses its type, and a list is sorted and joined with
// commas so that the order it is written in does not count as a difference.
func attributeValue(raw string) string {
	value := strings.TrimSpace(raw)
	if strings.HasPrefix(value, "(") {
		items := strings.Split(strings.Trim(value, "()"), ",")
		for i, item := range items {
			items[i] = strings.Trim(strings.TrimSpace(item), `"`)
		}
		return joinSorted(items)
	}
	value = strings.Trim(value, `"`)
	if _, literal, qualified := strings.Cut(value, "::"); qualified {
		return literal
	}
	return value
}

// jsonValue flattens a manifest field to the text the model would write for it.
func jsonValue(raw json.RawMessage) string {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return ""
	}
	return strings.Trim(value, `"`)
}

// joinSorted puts a list in order and joins it with commas, which is how a list
// of locations is compared.
func joinSorted(items []string) string {
	sorted := slices.Sorted(slices.Values(items))
	return strings.Join(sorted, ",")
}

// isCheckFile reports whether a file name is one of the check project's checks.
func isCheckFile(name string) bool {
	return strings.HasSuffix(name, ".spec.ts") || strings.HasSuffix(name, ".check.ts")
}

// skipDirectory keeps a walk inside this module's own source.
func skipDirectory(root, name string, entry fs.DirEntry) error {
	if name == root {
		return nil
	}
	if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "testdata" || entry.Name() == "node_modules" {
		return fs.SkipDir
	}
	// A directory with a go.mod is a module of its own, and its tests are not
	// this module's tests.
	if _, err := os.Stat(filepath.Join(name, "go.mod")); err == nil {
		return fs.SkipDir
	}
	return nil
}

// filesUnder returns every file under dir whose name ends in suffix, in path
// order. A directory that is not there yields no files.
func filesUnder(root, dir, suffix string) ([]File, error) {
	base := filepath.Join(root, filepath.FromSlash(dir))
	var found []File
	err := filepath.WalkDir(base, func(name string, entry fs.DirEntry, err error) error {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return nil
		case err != nil:
			return err
		case entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix):
			return nil
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		found = append(found, File{Path: filepath.ToSlash(rel), Text: string(raw)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}
