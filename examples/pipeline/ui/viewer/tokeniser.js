// tokeniser.js: splits SysML v2 text into tokens for display (SR-11).
//
// This is a tokeniser and not a parser. One sticky regular expression is
// tried at each position, its alternatives ordered by priority, and the
// first alternative that matches decides the token. The keyword table is
// the reserved-word list of OMG SysML v2 Part 1: Language, formal/26-03-02,
// clause 8.2.2.1.2 "Lexical Structure" (C-82). Nothing is copied from any
// implementation's grammar.

export const KEYWORDS = new Set(`about abstract accept action actor after alias all allocate allocation analysis
and as assert assign assume at attribute bind binding by calc case comment
concern connect connection constant constraint crosses decide def default
defined dependency derived do doc else end entry enum event exhibit exit expose
false filter first flow for fork frame from hastype if implies import in include
individual inout interface istype item join language library locale loop merge
message meta metadata nonunique not null objective occurrence of or ordered out
package parallel part perform port private protected public redefines ref
references render rendering rep require requirement return satisfy send snapshot
specializes stakeholder standard state subject subsets succession terminate then
timeslice to transition true until use variant variation verification verify via
view viewpoint when while xor`.split(/\s+/));

// Alternatives in priority order, each capturing group is one token type.
const TOKEN = new RegExp(
  [
    /\/\/\*[\s\S]*?\*\/|\/\/[^\n]*|\/\*[\s\S]*?\*\//.source, // 1 comment: multi-line note, line note, regular comment
    /doc(?!\w)(?:\s+[A-Za-z_]\w*)?\s*\/\*[\s\S]*?\*\//.source, // 2 doc, optional name, and its body
    /'(?:[^'\\\n]|\\.)*'|"(?:[^"\\\n]|\\.)*"/.source,     // 3 string: quoted name or string literal
    /\d+(?:\.\d+)?(?:[eE][+-]?\d+)?(?:\s*\[[^\]\n]*\])?/.source, // 4 number with optional [unit]
    /[A-Za-z_]\w*/.source,                                // 5 word: keyword or identifier, decided below
    /:>>|::>|:>|::|=>|:=|\.\.|->|[{}()[\];:=<>+\-*/.,#@~^|&!?%]/.source, // 6 operator
    /\s+/.source,                                         // 7 whitespace
    /[\s\S]/.source,                                      // 8 anything else, one character
  ].map((s) => `(${s})`).join("|"),
  "y",
);
const TYPES = ["comment", "doc", "string", "number", "word", "operator", "space", "other"];

export function tokenise(text) {
  const tokens = [];
  TOKEN.lastIndex = 0;
  let match;
  while (TOKEN.lastIndex < text.length && (match = TOKEN.exec(text)) !== null) {
    const group = match.findIndex((g, i) => i > 0 && g !== undefined);
    let type = TYPES[group - 1];
    if (type === "word") type = KEYWORDS.has(match[0]) ? "keyword" : "identifier";
    tokens.push({ type, text: match[0], start: match.index, end: TOKEN.lastIndex });
  }
  return tokens;
}

export function esc(s) {
  return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}

// render returns HTML for a <pre>: one span per token worth colouring and
// escaped text for whitespace and anything unrecognised.
export function render(text) {
  return tokenise(text)
    .map((t) => (t.type === "space" || t.type === "other"
      ? esc(t.text)
      : `<span class="tok-${t.type}">${esc(t.text)}</span>`))
    .join("");
}
