package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// A browse is one task put to the language model: one question per tool
// call, each carrying the view and the last answer only, until the language
// model says it is done or its calls run out, and then the final question.

// The budgets of tool calls.
const (
	BudgetTest     = 8
	BudgetIncident = 20
)

// Task is what a browse is asked about.
type Task struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Text string `json:"text"`
}

// Question is one chat request of a browse. Key names its place in the
// browse: 1 for the first step, and final for the last question.
type Question struct {
	Key    string
	System string
	User   string
	Format json.RawMessage
	Final  bool
}

// Asker puts one question to the server.
type Asker func(ctx context.Context, q Question) (Reply, error)

// StepRecord is one step of a browse.
type StepRecord struct {
	Step   Step   `json:"step"`
	Raw    string `json:"raw"`
	Error  string `json:"error,omitempty"`
	Answer string `json:"answer"` // the tool's answer, or what was wrong with the step
}

// FinalRecord is a browse's final answer.
type FinalRecord struct {
	Test     *TestAnswer
	Incident *IncidentAnswer
	Raw      string
	Error    string
}

// BrowseResult is a whole browse.
type BrowseResult struct {
	Steps []StepRecord
	View  View
	Final FinalRecord
}

// observer hears each step as it happens, with the question and the reply
// behind it, and then the whole browse with its final question and reply.
type observer func(q Question, reply Reply, step *StepRecord, whole *BrowseResult) error

// Browse runs one browse. It stops early only when the server can't be
// asked: a reply that isn't the JSON asked for is recorded and the browse goes
// on.
func Browse(ctx context.Context, task Task, kit *ToolKit, w *Wiki, ask Asker) (BrowseResult, error) {
	return browse(ctx, task, kit, ask, nil)
}

func browse(ctx context.Context, task Task, kit *ToolKit, ask Asker, observe observer) (BrowseResult, error) {
	var r BrowseResult
	budget, format := BudgetTest, TestFinalSchema()
	if task.Kind == TaskIncident {
		budget, format = BudgetIncident, IncidentFinalSchema()
	}
	system := SystemMessage(task.Kind)
	var last *StepRecord
	for n := 1; n <= budget; n++ {
		q := Question{Key: fmt.Sprint(n), System: system, User: StepQuestion(task, r.View, budget-n+1, last), Format: StepSchema()}
		reply, err := ask(ctx, q)
		if err != nil {
			return r, err
		}
		rec := StepRecord{Raw: reply.Raw}
		step, err := ParseStep(reply.Raw)
		done := false
		if err != nil {
			rec.Error = err.Error()
			rec.Answer = "Your last reply wasn't the JSON step asked for (" + err.Error() + "). Reply with one step."
		} else {
			rec.Step = step
			if vp := strings.TrimSpace(step.Viewpoint); vp != "" {
				r.View.Viewpoint = vp
			}
			var notes []string
			for _, id := range step.Prune {
				r.View.Prune(id)
			}
			for _, x := range step.Expose {
				if err := r.View.Expose(kit.Wiki(), x.ID, x.Note); err != nil {
					notes = append(notes, err.Error())
				}
			}
			if step.Tool == "done" {
				done = true
				rec.Answer = "done"
			} else {
				rec.Answer = kit.Call(step.Tool, step.Arg, step.Arg2)
			}
			if len(notes) > 0 {
				rec.Answer = strings.Join(notes, "\n") + "\n" + rec.Answer
			}
		}
		r.Steps = append(r.Steps, rec)
		last = &r.Steps[len(r.Steps)-1]
		if observe != nil {
			if err := observe(q, reply, last, nil); err != nil {
				return r, err
			}
		}
		if done {
			break
		}
	}
	q := Question{Key: "final", System: system, User: FinalQuestion(task, r.View, last), Format: format, Final: true}
	reply, err := ask(ctx, q)
	if err != nil {
		return r, err
	}
	r.Final.Raw = reply.Raw
	if task.Kind == TaskIncident {
		a, err := ParseIncidentAnswer(reply.Raw)
		if err != nil {
			r.Final.Error = err.Error()
		} else {
			r.Final.Incident = &a
		}
	} else {
		a, err := ParseTestAnswer(reply.Raw)
		if err != nil {
			r.Final.Error = err.Error()
		} else {
			r.Final.Test = &a
		}
	}
	if observe != nil {
		if err := observe(q, reply, nil, &r); err != nil {
			return r, err
		}
	}
	return r, nil
}
