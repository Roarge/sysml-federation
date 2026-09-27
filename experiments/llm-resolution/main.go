// Command llm-resolution puts a language model on the operator's own Ollama
// server to work on the demo's systems model, read as a wiki. It links each Go
// test to the elements it verifies, with the requirement keys hidden, and
// explains one incident, and then puts its answers and its explanations to
// the probes. See README.md, and model/ for the requirements it meets.
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
		keyPath   = fs.String("key", "incident-key.json", "the incident's key")
		model     = fs.String("model", DefaultSettings().Model, "the model to ask")
		seed      = fs.Int("seed", DefaultSettings().Seed, "the seed for the model and for the probes' random choices")
		quick     = fs.Bool("quick", false, "run every step on a fixed sample of twelve tests and browse the incident once")
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
	wiki, err := LoadWiki(root)
	if err != nil {
		fmt.Fprintf(stderr, "reading the systems model: %v\n", err)
		return exitSetup
	}
	if n := len(wiki.Unresolved); n > 0 {
		fmt.Fprintf(stderr, "%d references in the systems model don't resolve, the first: %s\n", n, wiki.Unresolved[0])
	}
	code, err := LoadCode(root)
	if err != nil {
		fmt.Fprintf(stderr, "reading the code: %v\n", err)
		return exitSetup
	}
	key, err := LoadKey(*keyPath)
	if err == nil {
		err = key.Check(wiki, code)
	}
	if err != nil {
		fmt.Fprintf(stderr, "the incident's key %s doesn't fit the checkout: %v\n", *keyPath, err)
		return exitSetup
	}

	settings := DefaultSettings()
	settings.Model, settings.Seed = *model, *seed
	client := NewClient(*serverURL, settings)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	facts, err := client.Facts(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "the server isn't ready: %v\n", err)
		return exitSetup
	}
	var previous map[string]CallLine
	if *resume != "" {
		if previous, err = previousCalls(*resume, corpus.Hash(), wiki.Hash(), settings); err != nil {
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

	header := RunHeader{
		Started: started.Format(time.RFC3339), Commit: commitOf(root), GoVersion: runtime.Version(),
		Settings: settings, Server: facts, CorpusHash: corpus.Hash(), ModelHash: wiki.Hash(),
		Requirements: len(corpus.Requirements), Elements: len(wiki.Elements), Tests: len(tests), Keyed: countKeyed(tests),
		Quick: *quick, BaselineThreshold: BaselineThreshold, BudgetTest: BudgetTest, BudgetIncident: BudgetIncident,
		TestInstructions: SystemMessage(TaskTest), IncidentInstructions: SystemMessage(TaskIncident), Key: key,
	}
	err = w.write(header)
	baseline := NewBaseline(corpus.Requirements)
	for _, t := range tests {
		if err == nil {
			err = w.write(BaselineLine{BaselineResult: baseline.Resolve(t)})
		}
	}
	if err == nil {
		fmt.Fprintf(stderr, "asking %s at %s about %d tests and the incident; results in %s\n", settings.Model, *serverURL, len(tests), path)
		r := &Runner{Corpus: corpus, Tests: tests, Wiki: wiki, Code: code, Key: key, Client: client, Out: w, Log: stderr,
			Seed: int64(settings.Seed), Quick: *quick, Previous: previous}
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
