package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	caseRE     = regexp.MustCompile(`verification def VC_EXP_(S[RC])_(\d\d) \{`)
	evidenceRE = regexp.MustCompile(`action <'(Test\w+)'> \w+ \{\s*@Evidence \{ kind = "go-test"; location = "experiments/llm-resolution/([\w.]+)"; \}`)
	expFuncRE  = regexp.MustCompile(`(?m)^func (TestEXP(S[RC])(\d\d)_\w+)\(t \*testing\.T\)`)
)

func TestEXPSR16_TheModelAndTheTestsAgree(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("model", "verification-cases.sysml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	type named struct{ caseKey, file string }
	register := map[string]named{}
	heads := caseRE.FindAllStringSubmatchIndex(text, -1)
	for i, h := range heads {
		end := len(text)
		if i+1 < len(heads) {
			end = heads[i+1][0]
		}
		key := text[h[2]:h[3]] + text[h[4]:h[5]]
		for _, m := range evidenceRE.FindAllStringSubmatch(text[h[1]:end], -1) {
			if _, dup := register[m[1]]; dup {
				t.Errorf("%s is named twice in the register", m[1])
			}
			register[m[1]] = named{key, m[2]}
		}
	}
	if len(register) == 0 {
		t.Fatal("the register names no test")
	}

	found := map[string]string{}
	files, _ := filepath.Glob("*_test.go")
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range expFuncRE.FindAllStringSubmatch(string(src), -1) {
			found[m[1]] = f
			entry, ok := register[m[1]]
			switch {
			case !ok:
				t.Errorf("%s in %s is missing from the register", m[1], f)
			case entry.caseKey != m[2]+m[3]:
				t.Errorf("%s is named in VC_EXP_%s, not its own requirement's case", m[1], entry.caseKey)
			}
		}
	}
	for name, entry := range register {
		if !strings.HasPrefix(name, "TestEXP") {
			t.Errorf("%s in the register doesn't follow the TestEXP naming", name)
			continue
		}
		if file, ok := found[name]; !ok {
			t.Errorf("the register names %s, which doesn't exist", name)
		} else if file != entry.file {
			t.Errorf("the register puts %s in %s, but it is in %s", name, entry.file, file)
		}
	}
}
