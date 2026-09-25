package main

import (
	"strings"
	"unicode"
)

// splitIdentifier breaks an identifier or a run of text into words at
// punctuation, at lower-to-upper case changes, before the last capital of an
// acronym (HTTPServer becomes HTTP, Server) and between letters and digits.
// The words keep their original case.
func splitIdentifier(s string) []string {
	var out []string
	for _, span := range wordSpans(s) {
		out = append(out, s[span[0]:span[1]])
	}
	return out
}

// wordSpans returns the byte ranges of the words splitIdentifier would give.
func wordSpans(s string) [][2]int {
	var spans [][2]int
	rs := []rune(s)
	offsets := make([]int, len(rs)+1)
	pos := 0
	for i, r := range rs {
		offsets[i] = pos
		pos += len(string(r))
	}
	offsets[len(rs)] = pos
	start := -1
	flush := func(end int) {
		if start >= 0 && end > start {
			spans = append(spans, [2]int{offsets[start], offsets[end]})
		}
		start = -1
	}
	for i, r := range rs {
		alnum := unicode.IsLetter(r) || unicode.IsDigit(r)
		if !alnum {
			flush(i)
			continue
		}
		if start < 0 {
			start = i
			continue
		}
		prev := rs[i-1]
		switch {
		case unicode.IsLower(prev) && unicode.IsUpper(r):
			flush(i)
			start = i
		case unicode.IsUpper(prev) && unicode.IsUpper(r) && i+1 < len(rs) && unicode.IsLower(rs[i+1]):
			flush(i)
			start = i
		case unicode.IsLetter(prev) != unicode.IsLetter(r):
			flush(i)
			start = i
		}
	}
	flush(len(rs))
	return spans
}

var stopWords = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`a an the and or of to in on for with by at as is are be been
		it its that this these those when shall which from into than then not no nor each every
		one any all if so but only same has have had was were do does did their there what who
		whose while where how can may must should will would test t go file s`) {
		stopWords[w] = true
	}
}

// normalise lowercases a word and strips a few English endings, so that
// "patches", "patched" and "patch" compare equal. It is deliberately crude
// and the same on both sides of every comparison.
func normalise(w string) string {
	w = strings.ToLower(w)
	switch {
	case len(w) > 4 && strings.HasSuffix(w, "ies"):
		return w[:len(w)-3] + "y"
	case len(w) > 5 && strings.HasSuffix(w, "ing"):
		return w[:len(w)-3]
	case len(w) > 4 && strings.HasSuffix(w, "ed"):
		return w[:len(w)-2]
	case len(w) > 4 && (strings.HasSuffix(w, "ches") || strings.HasSuffix(w, "shes") || strings.HasSuffix(w, "xes") || strings.HasSuffix(w, "sses")):
		return w[:len(w)-2]
	case len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss"):
		return w[:len(w)-1]
	}
	return w
}

// terms returns the normalised content words of a text, stop words dropped.
func terms(s string) []string {
	var out []string
	for _, w := range splitIdentifier(s) {
		n := normalise(w)
		if len(n) < 2 || stopWords[n] || isNumber(n) {
			continue
		}
		out = append(out, n)
	}
	return out
}

func isNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}

// removeTerms deletes from s every word whose normalised form is in drop,
// keeping everything else as written. An identifier stays one identifier:
// SetAttributePatchesText without "patch" becomes SetAttributeText.
func removeTerms(s string, drop map[string]bool) (string, int) {
	var b strings.Builder
	last, removed := 0, 0
	for _, sp := range wordSpans(s) {
		if drop[normalise(s[sp[0]:sp[1]])] {
			b.WriteString(s[last:sp[0]])
			last = sp[1]
			removed++
		}
	}
	b.WriteString(s[last:])
	return collapseSpaces(b.String()), removed
}

// removeOccurrences deletes the words at the given positions (indexes into
// the content words of s, as terms would list them).
func removeOccurrences(s string, positions map[int]bool) string {
	var b strings.Builder
	last, idx := 0, 0
	for _, sp := range wordSpans(s) {
		n := normalise(s[sp[0]:sp[1]])
		if len(n) < 2 || stopWords[n] || isNumber(n) {
			continue
		}
		if positions[idx] {
			b.WriteString(s[last:sp[0]])
			last = sp[1]
		}
		idx++
	}
	b.WriteString(s[last:])
	return collapseSpaces(b.String())
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
