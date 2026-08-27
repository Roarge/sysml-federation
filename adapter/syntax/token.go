// Package syntax turns SysML v2 text into tokens and an abstract syntax tree in
// which every node carries its byte range in the source. It accepts exactly the
// subset of the notation the adapter projects and refuses anything else with
// the file, line and column of the first offending token. It has no opinion
// about meaning: name resolution and evaluation belong to the model package.
package syntax

import "fmt"

// Span is a half-open byte range [Start, End) in the source text.
type Span struct{ Start, End int }

// Kind classifies a token.
type Kind int

// The token kinds. EOF is the zero value so a missing token reads as the end.
const (
	EOF Kind = iota
	Ident
	Name
	Keyword
	Number
	String
	Comment
	Punct
)

// Token is one lexical unit with its position.
type Token struct {
	Kind Kind
	Text string
	Span Span
}

// Error is a construct the adapter refuses, with its position. Line and
// Column are 1-based, and Column counts bytes from the start of the line.
type Error struct {
	File    string
	Line    int
	Column  int
	Message string
}

// Error formats the refusal as "<file>:<line>:<column>: <message>".
func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
}

// Position converts a byte offset into a 1-based line and byte column.
func Position(src string, offset int) (line, column int) {
	line, lineStart := 1, 0
	for i := 0; i < offset && i < len(src); i++ {
		if src[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}
	return line, offset - lineStart + 1
}

// ErrorAt builds an Error for the start of span.
func ErrorAt(file, src string, at Span, message string) *Error {
	line, col := Position(src, at.Start)
	return &Error{File: file, Line: line, Column: col, Message: message}
}

// keywords is the reserved-word list of OMG SysML v2 Part 1 Language
// (formal/26-03-02, clause 8.2.2.1.2), plus the KerML words it inherits. Any of
// them in a name position is refused, which is what keeps a construct outside
// the subset from being taken for a name.
var keywords = map[string]bool{}

func init() {
	for _, w := range []string{
		"about", "abstract", "accept", "action", "actor", "after", "alias", "all", "allocate",
		"allocation", "analysis", "and", "as", "assert", "assign", "assoc", "assume", "at",
		"attribute", "bind", "binding", "by", "calc", "case", "comment", "concern", "connect",
		"connection", "const", "constant", "constraint", "crossing", "decide", "def", "default",
		"defined", "dependency", "derived", "do", "doc", "else", "end", "entry", "enum", "event",
		"exhibit", "exit", "expose", "false", "filter", "first", "flow", "for", "fork", "frame",
		"from", "hastype", "if", "implies", "import", "in", "include", "individual", "inout",
		"interface", "istype", "item", "join", "language", "library", "locale", "loop", "merge",
		"message", "meta", "metadata", "nonunique", "not", "null", "objective", "occurrence", "of",
		"or", "ordered", "out", "package", "parallel", "part", "perform", "port", "portion",
		"private", "protected", "public", "readonly", "redefines", "ref", "references", "render",
		"rendering", "rep", "require", "requirement", "return", "satisfy", "send", "snapshot",
		"specializes", "stakeholder", "standard", "state", "subject", "subsets", "succession",
		"terminate", "then", "timeslice", "to", "transition", "true", "until", "use", "variant",
		"variation", "verification", "verify", "via", "view", "viewpoint", "when", "while", "xor",
	} {
		keywords[w] = true
	}
}
