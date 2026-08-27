package syntax

import (
	"fmt"
	"strconv"
	"strings"
)

// operators lists every multi-character token first so the longest match wins.
var operators = []string{"::>", ":>>", ":>", "::", ">=", "<=", "==",
	"{", "}", "(", ")", "[", "]", ";", ":", "=", ".", ",", "<", ">", "+", "-", "*", "/", "#"}

// Lex turns src into tokens. Line comments and //* */ notes are dropped; /* */
// comments are tokens because doc bodies are made of them.
func Lex(file, src string) ([]Token, error) {
	var toks []Token
	fail := func(at int, msg string) error {
		return ErrorAt(file, src, Span{Start: at, End: at}, msg)
	}
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			i++
		case strings.HasPrefix(src[i:], "//*"):
			end := strings.Index(src[i+3:], "*/")
			if end < 0 {
				return nil, fail(i, "unterminated note")
			}
			i += 3 + end + 2
		case strings.HasPrefix(src[i:], "//"):
			end := strings.IndexByte(src[i:], '\n')
			if end < 0 {
				end = len(src) - i
			}
			i += end
		case strings.HasPrefix(src[i:], "/*"):
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				return nil, fail(i, "unterminated comment")
			}
			toks = append(toks, Token{Comment, src[i+2 : i+2+end], Span{i, i + 2 + end + 2}})
			i += 2 + end + 2
		case isLetter(c):
			start := i
			for i < len(src) && (isLetter(src[i]) || isDigit(src[i])) {
				i++
			}
			word := src[start:i]
			kind := Ident
			if keywords[word] {
				kind = Keyword
			}
			toks = append(toks, Token{kind, word, Span{start, i}})
		case isDigit(c):
			start := i
			i = scanNumber(src, i)
			text := src[start:i]
			if _, err := strconv.ParseFloat(text, 64); err != nil {
				return nil, fail(start, fmt.Sprintf("cannot read %q as a number", text))
			}
			toks = append(toks, Token{Number, text, Span{start, i}})
		case c == '\'':
			end := strings.IndexAny(src[i+1:], "'\n")
			if end < 0 || src[i+1+end] == '\n' {
				return nil, fail(i, "unterminated quoted name")
			}
			toks = append(toks, Token{Name, src[i+1 : i+1+end], Span{i, i + 1 + end + 1}})
			i += 1 + end + 1
		case c == '"':
			start, j := i, i+1
			for j < len(src) && src[j] != '"' && src[j] != '\n' {
				if src[j] == '\\' {
					j++
				}
				j++
			}
			if j >= len(src) || src[j] != '"' {
				return nil, fail(start, "unterminated string")
			}
			toks = append(toks, Token{String, src[start+1 : j], Span{start, j + 1}})
			i = j + 1
		default:
			op := matchOperator(src[i:])
			if op == "" {
				return nil, fail(i, fmt.Sprintf("unexpected character %q", c))
			}
			toks = append(toks, Token{Punct, op, Span{i, i + len(op)}})
			i += len(op)
		}
	}
	return append(toks, Token{EOF, "", Span{len(src), len(src)}}), nil
}

func isLetter(c byte) bool { return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }

// scanNumber accepts digits, an optional fraction and an optional exponent,
// and returns the index after the last byte that belongs to the number.
func scanNumber(src string, i int) int {
	for i < len(src) && isDigit(src[i]) {
		i++
	}
	if i+1 < len(src) && src[i] == '.' && isDigit(src[i+1]) {
		i++
		for i < len(src) && isDigit(src[i]) {
			i++
		}
	}
	if i < len(src) && (src[i] == 'e' || src[i] == 'E') {
		j := i + 1
		if j < len(src) && (src[j] == '+' || src[j] == '-') {
			j++
		}
		if j < len(src) && isDigit(src[j]) {
			for j < len(src) && isDigit(src[j]) {
				j++
			}
			i = j
		}
	}
	return i
}

func matchOperator(rest string) string {
	for _, op := range operators {
		if strings.HasPrefix(rest, op) {
			return op
		}
	}
	return ""
}
