// Command llm-resolution browses the demo's systems model with a language
// model on the operator's own Ollama server. See README.md.
package main

import (
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int { return 1 }
