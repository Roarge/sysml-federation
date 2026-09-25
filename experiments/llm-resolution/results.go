package main

// RunHeader opens the results file.
type RunHeader struct {
	Type                 string      `json:"type"`
	Commit               string      `json:"commit"`
	Settings             Settings    `json:"settings"`
	Server               ServerFacts `json:"server"`
	CorpusHash           string      `json:"corpus_hash"`
	ModelHash            string      `json:"model_hash"`
	TestInstructions     string      `json:"test_instructions"`
	IncidentInstructions string      `json:"incident_instructions"`
	Key                  Key         `json:"incident_key"`
	BudgetTest           int         `json:"budget_test"`
	BudgetIncident       int         `json:"budget_incident"`
}

// LinkFact is one link of a test's answer, placed in the systems model.
type LinkFact struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Requirement bool   `json:"requirement,omitempty"`
	Near        bool   `json:"near,omitempty"`
}

// CallLine is one question and reply, or a browse's final line.
type CallLine struct {
	Type       string          `json:"type"`
	Key        string          `json:"key"`
	Probe      string          `json:"probe"`
	Variant    string          `json:"variant,omitempty"`
	Test       string          `json:"test"`
	Gold       string          `json:"gold,omitempty"`
	BasePick   string          `json:"base_pick,omitempty"`
	Step       int             `json:"step,omitempty"`
	Final      bool            `json:"final,omitempty"`
	Task       string          `json:"task,omitempty"`
	User       string          `json:"user,omitempty"`
	Reply      *Reply          `json:"reply,omitempty"`
	Error      string          `json:"error,omitempty"`
	Note       string          `json:"note,omitempty"`
	Tool       string          `json:"tool,omitempty"`
	ToolAnswer string          `json:"tool_answer,omitempty"`
	Pick       string          `json:"pick,omitempty"`
	Answer     *TestAnswer     `json:"test_answer,omitempty"`
	Account    *IncidentAnswer `json:"incident_answer,omitempty"`
	Facts      []LinkFact      `json:"link_facts,omitempty"`
	BaseRate   float64         `json:"base_rate,omitempty"`
	Citations  []Citation      `json:"citations,omitempty"`
	Items      []KeyResult     `json:"key_items,omitempty"`
	SysML      string          `json:"view_sysml,omitempty"`
}

// Contents is a results file read back.
type Contents struct {
	Header   RunHeader
	Baseline []BaselineResult
	Calls    []CallLine
}

// ReadResults is not built yet.
func ReadResults(path string) (Contents, error) { return Contents{}, errNotBuilt }
