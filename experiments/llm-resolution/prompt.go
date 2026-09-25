package main

import (
	"fmt"
	"strings"
)

// instructions open the system message. The requirements follow them, and
// the two together are the same for every base call, so the server can keep
// them in its prompt cache and read only the test each time.
const instructions = `You recover trace links between a systems model and the code that verifies it.

The requirements below come from a SysML v2 systems model. Each question gives one Go test function from the same repository, with any requirement keys removed from it. Decide which single requirement the test verifies, or answer "none" if no requirement fits.

Reply in JSON with three fields:
- requirement: the key of the requirement, or "none"
- evidence: up to five words or short phrases, copied exactly from the test or the requirement, that decided your answer
- reason: one sentence explaining the link, or why there is none

Requirements:
`

// TestView is a test as one prompt shows it. The probes alter these fields.
type TestView struct {
	Name    string `json:"name"`
	Package string `json:"package"`
	File    string `json:"file"`
	Doc     string `json:"doc"`
}

func viewOf(t Test) TestView {
	return TestView{Name: t.Name, Package: t.Package, File: t.File, Doc: t.Doc}
}

func systemPrompt(reqs []Requirement) string {
	var b strings.Builder
	b.WriteString(instructions)
	for _, r := range reqs {
		fmt.Fprintf(&b, "%s %s: %s\n", r.Key, r.Name, r.Statement)
	}
	return b.String()
}

func userPrompt(v TestView) string {
	doc := v.Doc
	if strings.TrimSpace(doc) == "" {
		doc = "(none)"
	}
	return fmt.Sprintf("Test function:\nname: %s\npackage: %s\nfile: %s\ndoc: %s\n\nWhich requirement does this test verify?",
		orNotShown(v.Name), orNotShown(v.Package), orNotShown(v.File), doc)
}

func orNotShown(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(not shown)"
	}
	return s
}

// baseQuestion is the unaltered question about one test: every requirement
// in the registers' order, and the test as read.
func baseQuestion(c Corpus, t Test) (system, user string) {
	return systemPrompt(c.Requirements), userPrompt(viewOf(t))
}
