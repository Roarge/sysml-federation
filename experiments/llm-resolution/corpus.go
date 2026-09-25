package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Requirement is one system story (SR) or design constraint (SC) from the
// demo's own systems model, as the resolvers see it.
type Requirement struct {
	Key       string `json:"key"`       // SR-22
	Name      string `json:"name"`      // Edits patch the source
	Statement string `json:"statement"` // the EARS statement or constraint text
}

// Test is one top-level Go test function. The fields a resolver sees have
// the requirement key removed. Gold is the key the name carried, or "" for a
// test whose name carries none.
type Test struct {
	ID       string `json:"id"`        // file#FuncName, unique
	FuncName string `json:"func_name"` // the name as written, key and all
	Name     string `json:"name"`      // the name with Test and any key removed
	Package  string `json:"package"`
	File     string `json:"file"` // slash-separated, relative to the repository root
	Doc      string `json:"doc"`  // the doc comment, keys masked
	Gold     string `json:"gold,omitempty"`
}

// Corpus is everything read from the repository.
type Corpus struct {
	Requirements []Requirement
	Tests        []Test
}

var (
	reqHeaderRE = regexp.MustCompile(`requirement <'(S[RC]-\d\d)'> S[RC]_\d\d_(\w+) : (?:UserStory|DesignConstraint) \{`)
	statementRE = regexp.MustCompile(`attribute\s+(?::>>\s*)?statement\s*(?::\s*String\s*)?=\s*"((?:[^"\\]|\\.)*)"`)
	testFuncRE  = regexp.MustCompile(`^func (Test\w*)\(\s*\w+\s+\*testing\.T\s*\)`)
	keyedNameRE = regexp.MustCompile(`^Test(S[RC])(\d\d)_(\w+)$`)
	packageRE   = regexp.MustCompile(`^package (\w+)`)
	keyInTextRE = regexp.MustCompile(`TestS[RC]\d\d_(\w+)`)
	bareKeyRE   = regexp.MustCompile(`\bS[RC]-?\d\d\b`)
)

// requirementFiles are the two registers the requirements come from, relative
// to the repository root.
var requirementFiles = []string{
	"model/core/stories/system/system-stories.sysml",
	"model/core/constraints/design-constraints.sysml",
}

// LoadCorpus reads the requirements and the Go tests under root.
func LoadCorpus(root string) (Corpus, error) {
	var c Corpus
	for _, rel := range requirementFiles {
		text, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return c, err
		}
		c.Requirements = append(c.Requirements, parseRequirements(string(text))...)
	}
	sort.Slice(c.Requirements, func(i, j int) bool { return c.Requirements[i].Key < c.Requirements[j].Key })
	tests, err := readTests(root)
	if err != nil {
		return c, err
	}
	c.Tests = tests
	return c, nil
}

func parseRequirements(text string) []Requirement {
	heads := reqHeaderRE.FindAllStringSubmatchIndex(text, -1)
	var out []Requirement
	for i, h := range heads {
		end := len(text)
		if i+1 < len(heads) {
			end = heads[i+1][0]
		}
		body := text[h[1]:end]
		st := ""
		if m := statementRE.FindStringSubmatch(body); m != nil {
			st = strings.ReplaceAll(m[1], `\"`, `"`)
		}
		out = append(out, Requirement{
			Key:       text[h[2]:h[3]],
			Name:      identifierToWords(text[h[4]:h[5]]),
			Statement: st,
		})
	}
	return out
}

// identifierToWords turns EditsPatchTheSource into "Edits patch the source".
func identifierToWords(id string) string {
	parts := splitIdentifier(id)
	for i := range parts {
		if i > 0 && !isAllUpper(parts[i]) {
			parts[i] = strings.ToLower(parts[i])
		}
	}
	return strings.Join(parts, " ")
}

func isAllUpper(s string) bool {
	return strings.ToUpper(s) == s && strings.ToLower(s) != s
}

func readTests(root string) ([]Test, error) {
	var tests []Test
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "testdata" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				return filepath.SkipDir // another module, this experiment's own among them
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		found, err := testsInFile(path, filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		tests = append(tests, found...)
		return nil
	})
	sort.Slice(tests, func(i, j int) bool { return tests[i].ID < tests[j].ID })
	return tests, err
}

func testsInFile(path, rel string) ([]Test, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var (
		out     []Test
		pkg     string
		comment []string
	)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if m := packageRE.FindStringSubmatch(line); m != nil && pkg == "" {
			pkg = m[1]
		}
		if strings.HasPrefix(line, "//") {
			if !strings.HasPrefix(line, "//go:") {
				comment = append(comment, strings.TrimSpace(strings.TrimPrefix(line, "//")))
			}
			continue
		}
		if m := testFuncRE.FindStringSubmatch(line); m != nil {
			out = append(out, newTest(m[1], pkg, rel, strings.Join(comment, " ")))
		}
		comment = nil
	}
	return out, sc.Err()
}

func newTest(funcName, pkg, file, doc string) Test {
	t := Test{ID: file + "#" + funcName, FuncName: funcName, Package: pkg, File: file}
	if m := keyedNameRE.FindStringSubmatch(funcName); m != nil {
		t.Gold = m[1] + "-" + m[2]
		t.Name = m[3]
	} else {
		t.Name = strings.TrimPrefix(strings.TrimPrefix(funcName, "Test"), "_")
	}
	t.Doc = maskKeys(doc)
	return t
}

// maskKeys removes requirement keys from free text, so a doc comment can't
// give the answer away.
func maskKeys(s string) string {
	s = keyInTextRE.ReplaceAllString(s, "$1")
	return bareKeyRE.ReplaceAllString(s, "[key]")
}

// Keys returns the requirement keys in order.
func (c Corpus) Keys() []string {
	keys := make([]string, len(c.Requirements))
	for i, r := range c.Requirements {
		keys[i] = r.Key
	}
	return keys
}

// Hash identifies the corpus, so two runs can be seen to have read the same
// repository state.
func (c Corpus) Hash() string {
	h := sha256.New()
	for _, r := range c.Requirements {
		fmt.Fprintf(h, "%s\x00%s\x00%s\n", r.Key, r.Name, r.Statement)
	}
	for _, t := range c.Tests {
		fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s\n", t.ID, t.Name, t.Doc, t.Gold)
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Validate checks the corpus against the shape the repository's own trace
// test enforces, so a run on a broken checkout stops before it starts.
func (c Corpus) Validate() error {
	if len(c.Requirements) == 0 || len(c.Tests) == 0 {
		return fmt.Errorf("read %d requirements and %d tests: is -repo the repository root?", len(c.Requirements), len(c.Tests))
	}
	known := map[string]bool{}
	for _, r := range c.Requirements {
		if r.Statement == "" {
			return fmt.Errorf("%s has no statement", r.Key)
		}
		known[r.Key] = true
	}
	for _, t := range c.Tests {
		if t.Gold != "" && !known[t.Gold] {
			return fmt.Errorf("%s names %s, which the systems model doesn't declare", t.ID, t.Gold)
		}
	}
	return nil
}

// mainModule is the module line of the repository the experiment reads.
const mainModule = "module github.com/Roarge/sysml-federation\n"

// FindRepoRoot walks up from start to the folder whose go.mod declares the
// repository's main module, passing this experiment's own module on the way.
func FindRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.HasPrefix(string(data), mainModule) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no checkout of github.com/Roarge/sysml-federation above " + start + ": give -repo")
		}
		dir = parent
	}
}
