package main

// SystemMessage is not built yet.
func SystemMessage(kind string) string { return "" }

// TestTask is not built yet.
func TestTask(t Test) Task { return Task{Kind: TaskTest, ID: t.ID} }

// StepQuestion is not built yet.
func StepQuestion(task Task, v View, left int, last *StepRecord) string { return "" }
