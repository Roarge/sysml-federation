package main

import (
	"fmt"
	"strings"
)

// The instructions are the system message of every question of one task, the
// same text each time, so the server can keep them in its prompt cache.

const toolInstructions = `Each question gives you the task, your viewpoint (the question your view answers), your view, the calls you have left, and the answer to your last call. Earlier answers are not shown again, so keep what you need in the notes of your view.

Reply with one step in JSON:
- viewpoint: the question your view answers, in one sentence. Keep it, or sharpen it.
- expose: up to three elements to add to the view, each with its id and a short note saying why it matters.
- prune: up to three ids to take out of the view.
- tool, arg, arg2: your next call, or the tool "done" when your view answers the task.

The tools:
- find <words> [kind]: elements whose text shares the words, best first. arg2 may name a kind, such as requirement (the system requirements), part, part def, action, state def, transition or verification def.
- links <id> [relationship]: the element's links, grouped by relationship, both ways round. arg2 may name one relationship, such as "satisfied by".
- doc <id>: the element's description, attributes, decisions and evidence.
- path <id> <id>: up to three shortest link paths between two elements, of three links at most.
- code <id>: where the systems model puts the element in the code: evidence, paths its description names, and where its distinctive values occur.
- grep <text>: lines of the built system's code that hold the text.
- read <file> <line>: thirty lines of a file around the line.
- done: stop and give your final answer.

Use identifiers as find, links and path give them. An identifier the systems model lacks is answered as unknown.`

const testInstructions = `You link a Go test to the elements of a SysML v2 systems model that it verifies or exercises. The systems model describes the system the test belongs to, and a systems engineer linked its elements to one another: requirements to the parts that satisfy them and the cases that verify them, parts to their definitions, actions to the parts they are allocated to. You read it the way you would read a wiki, one call at a time, and you keep a view of the elements that matter to your task.

` + toolInstructions + ` Requirement keys are hidden in the test and in the code, shown as [key].

When you are done, or your calls run out, you give your final answer:
- links: up to five element ids the test verifies or exercises, the most likely first, each with a relation: verifies, exercises or other. If the test verifies a system requirement, put that requirement first.
- evidence: up to five words or short phrases, copied exactly from the test, that decided your answer.
- reason: one sentence.
- mismatches: up to three places where you suspect the system was not built as the systems model says, each with the element's id, what the systems model says, what the system shows, and why. Leave it empty if you suspect nothing.`

const incidentInstructions = `You explain an incident in a running system from its SysML v2 systems model and its code, for the engineer on call. The engineer tells you what is down. A systems engineer linked the systems model's elements to one another: checks to the requirements they verify, requirements to the parts that satisfy them, parts to their definitions and to the behaviour they exhibit. You read it the way you would read a wiki, one call at a time, and you keep a view of the elements that explain the incident.

` + toolInstructions + `

When you are done, or your calls run out, you give your final answer:
- cause: the ids of the elements where the fault lies.
- mechanism: the ids of the elements that describe how the fault spread.
- code: the files to read first, as path or path:line.
- consequences: the ids of the requirements, checks and other elements the incident affects.
- path: the chain you followed from the report to the cause, as ids and files in order, each joined to the next by a link or by code the element names.
- why: two or three sentences for the engineer.
- mismatches: up to three places where you suspect the system was not built as the systems model says, each with the element's id, what the systems model says, what the system shows, and why. Leave it empty if you suspect nothing.`

// SystemMessage gives the fixed instructions for a task.
func SystemMessage(kind string) string {
	if kind == TaskIncident {
		return incidentInstructions
	}
	return testInstructions
}

// TestView is a test as a task shows it. The probes alter these fields.
type TestView struct {
	Name    string `json:"name"`
	Package string `json:"package"`
	File    string `json:"file"`
	Doc     string `json:"doc"`
}

func viewOf(t Test) TestView {
	return TestView{Name: t.Name, Package: t.Package, File: t.File, Doc: t.Doc}
}

// taskText is the task for a test: its fields, with any key removed.
func taskText(v TestView) string {
	doc := v.Doc
	if strings.TrimSpace(doc) == "" {
		doc = "(none)"
	}
	return fmt.Sprintf("Link this Go test to what it verifies.\nname: %s\npackage: %s\nfile: %s\ndoc: %s",
		orNotShown(v.Name), orNotShown(v.Package), orNotShown(v.File), doc)
}

func orNotShown(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(not shown)"
	}
	return s
}

// TestTask is the task for one test as read.
func TestTask(t Test) Task {
	return Task{Kind: TaskTest, ID: t.ID, Text: taskText(viewOf(t))}
}

// taskFields are the values of a test task's fields, without their labels.
func taskFields(task string) []string {
	var out []string
	for _, line := range strings.Split(task, "\n") {
		for _, field := range []string{"name: ", "package: ", "file: ", "doc: "} {
			if v, ok := strings.CutPrefix(line, field); ok && v != "(none)" && v != "(not shown)" {
				out = append(out, v)
			}
		}
	}
	return out
}

// StepQuestion is the user message of one step: the task, the viewpoint, the
// view, the calls left and the last call with its answer, and nothing older.
func StepQuestion(task Task, v View, left int, last *StepRecord) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task:\n%s\n\n", task.Text)
	vp := v.Viewpoint
	if vp == "" {
		vp = "(not framed yet)"
	}
	fmt.Fprintf(&b, "Viewpoint: %s\n\nView:\n%s\n\nCalls left: %d\n\n", vp, v.Lines(), left)
	if last == nil {
		b.WriteString("Last call: none")
		return b.String()
	}
	call := strings.TrimSpace(strings.Join([]string{last.Step.Tool, last.Step.Arg, last.Step.Arg2}, " "))
	if last.Error != "" {
		call = "(not understood)"
	}
	fmt.Fprintf(&b, "Last call: %s\nAnswer:\n%s", call, last.Answer)
	return b.String()
}

// FinalQuestion is the user message that asks for the final answer.
func FinalQuestion(task Task, v View, last *StepRecord) string {
	return StepQuestion(task, v, 0, last) + "\n\nGive your final answer now."
}
