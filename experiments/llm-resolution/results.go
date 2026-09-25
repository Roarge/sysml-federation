package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// The results file is JSON Lines: a run header, one baseline line per test,
// one line per call as it completes, and the summary last. Every line carries
// a "type". A run that stops part way leaves every completed call on disk, so
// it can be resumed, and the file can be uploaded as it stands.

// RunHeader opens the file.
type RunHeader struct {
	Type              string      `json:"type"` // "run"
	Started           string      `json:"started"`
	Commit            string      `json:"commit"`
	GoVersion         string      `json:"go_version"`
	Settings          Settings    `json:"settings"`
	Server            ServerFacts `json:"server"`
	CorpusHash        string      `json:"corpus_hash"`
	Requirements      int         `json:"requirements"`
	Tests             int         `json:"tests"`
	Keyed             int         `json:"keyed_tests"`
	Quick             bool        `json:"quick"`
	Probes            []string    `json:"probes"`
	BaselineThreshold float64     `json:"baseline_threshold"`
	Instructions      string      `json:"instructions"`
	RequirementList   string      `json:"requirement_list"` // as the base prompt lists them
}

// BaselineLine is the baseline's answer for one test.
type BaselineLine struct {
	Type string `json:"type"` // "baseline"
	BaselineResult
}

// CallLine is one question put to the model, or one the probe had no reason
// to ask (Note says why).
type CallLine struct {
	Type      string       `json:"type"` // "call"
	Key       string       `json:"key"`
	Probe     string       `json:"probe"`
	Test      string       `json:"test"`
	Gold      string       `json:"gold,omitempty"`
	BasePick  string       `json:"base_pick,omitempty"`
	User      string       `json:"user,omitempty"` // the user message, which shows the test
	Removed   []string     `json:"removed,omitempty"`
	EditedReq *Requirement `json:"edited_requirement,omitempty"`
	Shuffled  bool         `json:"shuffled_requirements,omitempty"`
	Reply     *Reply       `json:"reply,omitempty"`
	Error     string       `json:"error,omitempty"`
	Note      string       `json:"note,omitempty"`
	Seconds   float64      `json:"seconds,omitempty"`
	Reused    bool         `json:"reused,omitempty"`
}

// SummaryLine closes the file.
type SummaryLine struct {
	Type string `json:"type"` // "summary"
	Summary
}

type lineType struct {
	Type string `json:"type"`
}

// Contents is a results file read back.
type Contents struct {
	Header   RunHeader
	Baseline []BaselineResult
	Calls    []CallLine
}

// ReadResults reads a results file, skipping the summary it may end with.
func ReadResults(path string) (Contents, error) {
	var c Contents
	f, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for n := 1; ; n++ {
		line, err := r.ReadString('\n')
		if strings.TrimSpace(line) != "" {
			if perr := parseLine(line, &c); perr != nil {
				return c, fmt.Errorf("%s line %d: %w", path, n, perr)
			}
		}
		if err == io.EOF {
			return c, nil
		}
		if err != nil {
			return c, err
		}
	}
}

func parseLine(line string, c *Contents) error {
	var t lineType
	if err := json.Unmarshal([]byte(line), &t); err != nil {
		return err
	}
	switch t.Type {
	case "run":
		return json.Unmarshal([]byte(line), &c.Header)
	case "baseline":
		var b BaselineLine
		if err := json.Unmarshal([]byte(line), &b); err != nil {
			return err
		}
		c.Baseline = append(c.Baseline, b.BaselineResult)
	case "call":
		var cl CallLine
		if err := json.Unmarshal([]byte(line), &cl); err != nil {
			return err
		}
		c.Calls = append(c.Calls, cl)
	}
	return nil
}

// lineWriter writes one JSON value per line and flushes each, so nothing is
// lost if the run is stopped.
type lineWriter struct {
	f *os.File
}

func (w lineWriter) write(v json.Marshaler) error {
	data, err := v.MarshalJSON()
	if err != nil {
		return err
	}
	if _, err := w.f.Write(append(data, '\n')); err != nil {
		return err
	}
	return w.f.Sync()
}

// The four line types marshal themselves, so lineWriter needs no empty
// interface.

func (h RunHeader) MarshalJSON() ([]byte, error) {
	type plain RunHeader
	h.Type = "run"
	return json.Marshal(plain(h))
}

func (b BaselineLine) MarshalJSON() ([]byte, error) {
	type plain BaselineLine
	b.Type = "baseline"
	return json.Marshal(plain(b))
}

func (c CallLine) MarshalJSON() ([]byte, error) {
	type plain CallLine
	c.Type = "call"
	return json.Marshal(plain(c))
}

func (s SummaryLine) MarshalJSON() ([]byte, error) {
	type plain SummaryLine
	s.Type = "summary"
	return json.Marshal(plain(s))
}
