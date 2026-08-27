package syntax_test

import (
	"testing"

	"github.com/Roarge/sysml-federation/adapter/syntax"
	"github.com/Roarge/sysml-federation/internal/assert"
	"github.com/Roarge/sysml-federation/internal/tabletest"
)

// tok is a token without its span, for cases where only kind and text matter.
type tok struct {
	Kind syntax.Kind
	Text string
}

func strip(ts []syntax.Token) []tok {
	out := make([]tok, 0, len(ts))
	for _, t := range ts {
		out = append(out, tok{t.Kind, t.Text})
	}
	return out
}

func TestLexKindsAndText(t *testing.T) {
	tabletest.RunErr(t, []tabletest.ErrCase[string, []tok]{
		{Name: "empty", In: "", Want: []tok{{syntax.EOF, ""}}},
		{Name: "keywords and identifiers", In: "part def Server",
			Want: []tok{{syntax.Keyword, "part"}, {syntax.Keyword, "def"}, {syntax.Ident, "Server"}, {syntax.EOF, ""}}},
		{Name: "quoted name", In: "<'PIPE-R1.1'>",
			Want: []tok{{syntax.Punct, "<"}, {syntax.Name, "PIPE-R1.1"}, {syntax.Punct, ">"}, {syntax.EOF, ""}}},
		{Name: "number with unit", In: "200[ms]",
			Want: []tok{{syntax.Number, "200"}, {syntax.Punct, "["}, {syntax.Ident, "ms"}, {syntax.Punct, "]"}, {syntax.EOF, ""}}},
		{Name: "decimal and exponent", In: "0.5 1e3 2.5E-1",
			Want: []tok{{syntax.Number, "0.5"}, {syntax.Number, "1e3"}, {syntax.Number, "2.5E-1"}, {syntax.EOF, ""}}},
		{Name: "trailing dot is not part of the number", In: "1.x",
			Want: []tok{{syntax.Number, "1"}, {syntax.Punct, "."}, {syntax.Ident, "x"}, {syntax.EOF, ""}}},
		{Name: "string", In: `locale "en"`,
			Want: []tok{{syntax.Keyword, "locale"}, {syntax.String, "en"}, {syntax.EOF, ""}}},
		{Name: "doc comment is a token", In: "doc /* text */",
			Want: []tok{{syntax.Keyword, "doc"}, {syntax.Comment, " text "}, {syntax.EOF, ""}}},
		{Name: "line comment is dropped", In: "a // b\nc",
			Want: []tok{{syntax.Ident, "a"}, {syntax.Ident, "c"}, {syntax.EOF, ""}}},
		{Name: "multi-line note is dropped", In: "a //* b\n c */ d",
			Want: []tok{{syntax.Ident, "a"}, {syntax.Ident, "d"}, {syntax.EOF, ""}}},
		{Name: "longest operator wins", In: ":>> :> ::> :: >= <= == = > <",
			Want: []tok{{syntax.Punct, ":>>"}, {syntax.Punct, ":>"}, {syntax.Punct, "::>"}, {syntax.Punct, "::"},
				{syntax.Punct, ">="}, {syntax.Punct, "<="}, {syntax.Punct, "=="}, {syntax.Punct, "="},
				{syntax.Punct, ">"}, {syntax.Punct, "<"}, {syntax.EOF, ""}}},
		{Name: "every single-character token", In: "{}()[];:=.,+-*/#",
			Want: []tok{{syntax.Punct, "{"}, {syntax.Punct, "}"}, {syntax.Punct, "("}, {syntax.Punct, ")"},
				{syntax.Punct, "["}, {syntax.Punct, "]"}, {syntax.Punct, ";"}, {syntax.Punct, ":"},
				{syntax.Punct, "="}, {syntax.Punct, "."}, {syntax.Punct, ","}, {syntax.Punct, "+"},
				{syntax.Punct, "-"}, {syntax.Punct, "*"}, {syntax.Punct, "/"}, {syntax.Punct, "#"}, {syntax.EOF, ""}}},
		{Name: "unexpected character", In: "port p : ~Q;", WantErr: true},
		{Name: "unterminated comment", In: "doc /* open", WantErr: true},
		{Name: "unterminated quoted name", In: "<'PIPE", WantErr: true},
		{Name: "unterminated string", In: `"abc`, WantErr: true},
		{Name: "number out of range", In: "1e999", WantErr: true},
	}, func(t *testing.T, in string) ([]tok, error) {
		ts, err := syntax.Lex("m.sysml", in)
		if err != nil {
			return nil, err
		}
		return strip(ts), nil
	})
}

func TestLexSpansAreByteOffsets(t *testing.T) {
	src := "part <'S1'> a : B {\n  attribute :>> x = 12.5;\n}"
	lexed, err := syntax.Lex("m.sysml", src)
	ts := assert.Must(t, lexed, err)
	// The number sits on line 2; its span must select exactly "12.5".
	var num syntax.Token
	for _, tk := range ts {
		if tk.Kind == syntax.Number {
			num = tk
		}
	}
	assert.Equal(t, src[num.Span.Start:num.Span.End], "12.5")
	// The quoted name's span covers the quotes; its text does not.
	assert.Equal(t, src[ts[2].Span.Start:ts[2].Span.End], "'S1'")
	assert.Equal(t, ts[2].Text, "S1")
	// EOF sits at len(src).
	assert.Equal(t, ts[len(ts)-1].Span, syntax.Span{Start: len(src), End: len(src)})
}

func TestLexErrorsCarryFileLineAndColumn(t *testing.T) {
	tabletest.Run(t, []tabletest.Case[string, syntax.Error]{
		{Name: "first line", In: "part ~",
			Want: syntax.Error{File: "m.sysml", Line: 1, Column: 6, Message: "unexpected character '~'"}},
		{Name: "third line", In: "a\nb\n  '", // the quote opens on line 3, column 3
			Want: syntax.Error{File: "m.sysml", Line: 3, Column: 3, Message: "unterminated quoted name"}},
		{Name: "unreadable number", In: "x = 1e999;",
			Want: syntax.Error{File: "m.sysml", Line: 1, Column: 5, Message: `cannot read "1e999" as a number`}},
	}, func(t *testing.T, in string) syntax.Error {
		_, err := syntax.Lex("m.sysml", in)
		return *assert.ErrorAs[*syntax.Error](t, err)
	})
}

func TestPosition(t *testing.T) {
	src := "ab\ncd\n\nef"
	tabletest.Run(t, []tabletest.Case[int, [2]int]{
		{Name: "start", In: 0, Want: [2]int{1, 1}},
		{Name: "before newline", In: 2, Want: [2]int{1, 3}},
		{Name: "second line", In: 4, Want: [2]int{2, 2}},
		{Name: "after blank line", In: 7, Want: [2]int{4, 1}},
	}, func(t *testing.T, in int) [2]int {
		l, c := syntax.Position(src, in)
		return [2]int{l, c}
	})
}

func TestReservedWordsAreKeywords(t *testing.T) {
	// A sample across the OMG reserved-word list, including the ones the
	// parser never uses: they must still be keywords so a construct outside
	// the subset is refused as a keyword rather than taken for a name.
	for _, w := range []string{"package", "import", "private", "public", "abstract", "part", "def",
		"attribute", "port", "item", "in", "out", "inout", "connect", "to", "requirement", "require",
		"constraint", "subject", "doc", "satisfy", "by", "verification", "objective", "verify",
		"connection", "end", "redefines", "action", "state", "flow", "ref", "alias", "enum", "calc",
		"assume", "assert", "bind", "interface", "allocate", "view", "true", "false", "null", "locale"} {
		lexed, err := syntax.Lex("m.sysml", w)
		ts := assert.Must(t, lexed, err)
		assert.Equal(t, ts[0].Kind, syntax.Keyword)
	}
	// Metadata short names and the example's names are not reserved.
	for _, w := range []string{"derivation", "original", "derive", "target", "input", "output", "queries"} {
		lexed, err := syntax.Lex("m.sysml", w)
		ts := assert.Must(t, lexed, err)
		assert.Equal(t, ts[0].Kind, syntax.Ident)
	}
}
