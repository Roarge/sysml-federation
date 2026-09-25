// Command llm-resolution asks a language model on the operator's own Ollama
// server which of the demo's requirements each Go test verifies, with the
// requirement keys hidden, and then puts the model's explanations to four
// tests. See README.md, and model/ for the requirements it meets.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// Exit statuses: 0 done, 1 the run failed part way, 2 nothing was asked
// because the setup is wrong (EXP-SR-13).
const (
	exitOK     = 0
	exitFailed = 1
	exitSetup  = 2
)

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("llm-resolution", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		serverURL = fs.String("url", "", "the Ollama server's address, such as http://192.168.1.20:11434")
		repo      = fs.String("repo", "", "the checkout to read (default: found from the working folder)")
		out       = fs.String("out", "results", "the folder the results file is written to")
		model     = fs.String("model", DefaultSettings().Model, "the model to ask")
		seed      = fs.Int("seed", DefaultSettings().Seed, "the seed for the model and for the probes' random choices")
		quick     = fs.Bool("quick", false, "run every step on a fixed sample of twelve tests")
		resume    = fs.String("resume", "", "a results file of a stopped run, whose replies are reused")
		report    = fs.String("report", "", "print the report for a results file and stop, contacting no server")
	)
	if err := fs.Parse(args); err != nil {
		return exitSetup
	}

	if *report != "" {
		c, err := ReadResults(*report)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitSetup
		}
		fmt.Fprint(stdout, Summarise(c).Markdown())
		return exitOK
	}

	if *serverURL == "" {
		fmt.Fprintln(stderr, "give the server's address with -url, or run run.sh with OLLAMA_URL set")
		return exitSetup
	}
	root := *repo
	if root == "" {
		var err error
		if root, err = FindRepoRoot("."); err != nil {
			fmt.Fprintln(stderr, err)
			return exitSetup
		}
	}
	corpus, err := LoadCorpus(root)
	if err == nil {
		err = corpus.Validate()
	}
	if err != nil {
		fmt.Fprintf(stderr, "reading %s: %v\n", root, err)
		return exitSetup
	}

	settings := DefaultSettings()
	settings.Model, settings.Seed = *model, *seed
	client := NewClient(*serverURL, settings, corpus.Keys())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	facts, err := client.Facts(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "the server isn't ready: %v\n", err)
		return exitSetup
	}
	var previous map[string]CallLine
	if *resume != "" {
		if previous, err = previousCalls(*resume, corpus.Hash(), settings); err != nil {
			fmt.Fprintf(stderr, "can't resume: %v\n", err)
			return exitSetup
		}
		fmt.Fprintf(stderr, "resuming: %d replies from %s will be reused\n", len(previous), *resume)
	}

	tests := corpus.Tests
	if *quick {
		tests = quickSample(tests)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return exitSetup
	}
	started := time.Now().UTC()
	path := filepath.Join(*out, "llm-resolution-"+started.Format("2006-01-02T150405Z")+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitSetup
	}
	w := lineWriter{f: f}

	system := systemPrompt(corpus.Requirements)
	header := RunHeader{
		Started: started.Format(time.RFC3339), Commit: commitOf(root), GoVersion: runtime.Version(),
		Settings: settings, Server: facts, CorpusHash: corpus.Hash(),
		Requirements: len(corpus.Requirements), Tests: len(tests), Keyed: countKeyed(tests),
		Quick: *quick, Probes: AllProbes, BaselineThreshold: BaselineThreshold,
		Instructions: instructions, RequirementList: strings.TrimPrefix(system, instructions),
	}
	err = w.write(header)
	baseline := NewBaseline(corpus.Requirements)
	for _, t := range tests {
		if err == nil {
			err = w.write(BaselineLine{BaselineResult: baseline.Resolve(t)})
		}
	}
	if err == nil {
		fmt.Fprintf(stderr, "asking %s at %s about %d tests; results in %s\n", settings.Model, *serverURL, len(tests), path)
		r := &Runner{Corpus: corpus, Tests: tests, Client: client, Out: w, Log: stderr, Seed: int64(settings.Seed),
			Probes: allProbes(), Previous: previous}
		err = r.Run(ctx)
	}
	if err != nil {
		f.Close()
		fmt.Fprintf(stderr, "the run stopped: %v\nevery reply so far is in %s; resume with -resume %s\n", err, path, path)
		return exitFailed
	}

	contents, err := ReadResults(path)
	if err == nil {
		summary := Summarise(contents)
		if err = w.write(SummaryLine{Summary: summary}); err == nil {
			err = os.WriteFile(strings.TrimSuffix(path, ".jsonl")+".md", []byte(summary.Markdown()), 0o644)
		}
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailed
	}
	fmt.Fprintf(stdout, "%s\n", path)
	fmt.Fprintf(stderr, "done. Upload %s\n", path)
	return exitOK
}

func allProbes() map[string]bool {
	m := map[string]bool{}
	for _, p := range AllProbes {
		m[p] = true
	}
	return m
}

func countKeyed(tests []Test) int {
	n := 0
	for _, t := range tests {
		if t.Gold != "" {
			n++
		}
	}
	return n
}

// commitOf names the commit the checkout is at, or "unknown".
func commitOf(root string) string {
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
