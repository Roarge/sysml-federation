package main

import (
	"context"
	"encoding/json"
)

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

// Question is one chat request of a browse.
type Question struct {
	Key    string
	System string
	User   string
	Format json.RawMessage
	Final  bool
}

// Asker puts one question to the server.
type Asker func(ctx context.Context, q Question) (Reply, error)

// StepRecord is one tool call of a browse.
type StepRecord struct {
	Step   Step
	Raw    string
	Error  string
	Answer string
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

// Browse is not built yet.
func Browse(ctx context.Context, task Task, kit *ToolKit, w *Wiki, ask Asker) (BrowseResult, error) {
	return BrowseResult{}, errNotBuilt
}
