// module-checks.js: what the viewer's two pure modules are expected to do.
//
// The tokeniser and the layout are the only parts of either app that are
// functions of their arguments alone, with no DOM and no network under them,
// so they are the only parts a test can hold to account without a browser.
// Everything here is an assertion about a return value.
//
// This runs under node, from a Go test that copies the two modules beside it.
// It is not served to a browser and is not part of either app.

import { KEYWORDS, tokenise, esc, render } from "./tokeniser.js";
import { layers, positions, render as sketch } from "./sketch.js";

let checks = 0;
const failures = [];

function ok(condition, what) {
  checks += 1;
  if (!condition) failures.push(what);
}

function same(got, want, what) {
  checks += 1;
  const a = JSON.stringify(got);
  const b = JSON.stringify(want);
  if (a !== b) failures.push(`${what}: got ${a}, wanted ${b}`);
}

const kinds = (text) => tokenise(text).map((t) => t.type);
const texts = (text) => tokenise(text).map((t) => t.text);

// ------------------------------------------------------------- tokeniser --

// One line of the model, as the demonstration checklist reads it aloud.
const line = "part <'PIPE-S2'> parse : Server { attribute throughput : Real = 1200; }";
same(kinds(line).slice(0, 11),
  ["keyword", "space", "operator", "string", "operator", "space", "identifier",
    "space", "operator", "space", "identifier"],
  "the checklist's own line opens with the kinds it predicts");
same(tokenise(line).filter((t) => t.type === "string").map((t) => t.text), ["'PIPE-S2'"],
  "the quoted short name is one string token");
same(tokenise(line).filter((t) => t.type === "number").map((t) => t.text), ["1200"],
  "the literal is one number token");
ok(tokenise(line).map((t) => t.text).join("") === line,
  "the tokens of the checklist line concatenate back to the input");

// Boundaries: every token abuts the next, the first starts at 0 and the last
// ends at the length. A tokeniser that loses a byte loses it here.
const spans = tokenise(line);
ok(spans[0].start === 0, "the first token starts at byte 0");
ok(spans[spans.length - 1].end === line.length, "the last token ends at the input's length");
ok(spans.every((t, i) => i === 0 || t.start === spans[i - 1].end),
  "each token starts where the one before it ended");
ok(spans.every((t) => t.end > t.start), "no token is empty");
ok(spans.every((t) => t.type !== undefined), "no token comes back without a kind");

// A number carries its unit, spaced or not, as one token. Splitting it would
// put the unit through the keyword table and colour it as an identifier.
same(texts("200 [ms]"), ["200 [ms]"], "a spaced unit stays inside the number token");
same(texts("200[ms]"), ["200[ms]"], "an unspaced unit stays inside the number token");
same(kinds("200[ms]"), ["number"], "a number with a unit is one number token");
same(texts("1.5e3"), ["1.5e3"], "an exponent is part of the number");
same(texts("2.5"), ["2.5"], "a decimal point does not end the number");

// Comments. The multi-line note has to win over the line note, which shares
// its first two characters, and both have to win over the operator.
same(kinds("//* a\nb */"), ["comment"], "a multi-line note is one comment token");
same(kinds("// note"), ["comment"], "a line note is one comment token");
same(kinds("/* c */"), ["comment"], "a regular comment is one comment token");
same(texts("// note\nx"), ["// note", "\n", "x"], "a line note stops at the newline");
same(kinds("doc /* body */"), ["doc"], "doc and its body are one token");
same(kinds("doc d /* body */"), ["doc"], "a named doc and its body are one token");
same(kinds("doc"), ["keyword"], "doc on its own is a keyword");
same(kinds("docking"), ["identifier"], "docking is an identifier, not a doc");
ok(KEYWORDS.has("doc"), "the keyword table carries doc");
ok(KEYWORDS.has("requirement"), "the keyword table carries requirement");
ok(!KEYWORDS.has("Server"), "a type name is not a keyword");
same(kinds("part x"), ["keyword", "space", "identifier"], "part is a keyword and x is not");

// Operators are taken longest first, or a redefinition would read as two
// tokens and a qualified name as three.
same(texts(":>>"), [":>>"], "the longest operator wins over its prefixes");
same(texts("::>"), ["::>"], "a feature chaining operator is one token");
same(texts("a::b"), ["a", "::", "b"], "a qualified name keeps its separator whole");
same(texts(":>"), [":>"], "subsetting is one token");
same(texts("x = 1"), ["x", " ", "=", " ", "1"], "an assignment splits into five tokens");

// Strings, both spellings the language allows.
same(kinds("'a name'"), ["string"], "a quoted name is a string token");
same(kinds('"a literal"'), ["string"], "a double-quoted literal is a string token");
same(texts("'a\\'b'"), ["'a\\'b'"], "an escaped quote does not end the string");
same(kinds("'unclosed"), ["other", "identifier"], "an unclosed quote does not swallow the rest");

// Escaping. A name in the model file must not be able to close an attribute
// or open a tag in the pane the text is rendered into.
same(esc('<a href="x">&'), "&lt;a href=&quot;x&quot;&gt;&amp;", "esc covers the four characters");
// The angle brackets are operators to the tokeniser, so they come back as
// two escaped operator tokens with the word between them. What matters is
// that neither bracket survives as itself.
ok(!render("part <img> x").includes("<img>"), "render does not pass a tag through");
ok(render("part <img> x").includes("&lt;"), "render escapes an opening bracket");
ok(render("part <img> x").includes("&gt;"), "render escapes a closing bracket");
ok(render("a & b").includes("&amp;"), "render escapes an ampersand");
ok(render("part x").includes('class="tok-keyword"'), "render marks a keyword");
ok(render("part x").includes('class="tok-identifier"'), "render marks an identifier");
ok(!render(" ").includes("<span"), "whitespace is written out without a span");

// The regular expression is sticky, so a second call must not inherit the
// first call's position. This is the one piece of state in the module.
const first = tokenise("part alpha");
const second = tokenise("part alpha");
same(second.map((t) => t.text), first.map((t) => t.text), "a second call starts from the beginning");

// A longer excerpt, in the shape of the model file. Nothing may be lost.
const excerpt = `package Demo {
  // a note
  part def Server { attribute rate : Real; }
  part <'A1'> alpha : Server { attribute rate = 1200 [1/s]; }
}`;
ok(tokenise(excerpt).map((t) => t.text).join("") === excerpt, "a longer excerpt round-trips");
ok(tokenise(excerpt).every((t) => t.type !== undefined), "no token of the excerpt lacks a kind");
ok(tokenise("").length === 0, "empty text gives no tokens");
ok(tokenise("§").length === 1, "an unrecognised character is one token");
same(kinds("§"), ["other"], "an unrecognised character is kind other");

// ---------------------------------------------------------------- layout --

const part = (id, value) => ({ id, name: id, attributes: [{ name: "throughput", value }] });
const edge = (from, to) => ({ id: `${from}-${to}`, from, to });

// The checklist's second wiring: a and b feed nothing in common, c sits
// behind b, and d takes an arrow from a and from c at different depths.
const fanIn = {
  name: "t",
  capacity: 3,
  bottleneck: [{ id: "c", name: "c" }],
  parts: [part("a", 1), part("b", 2), part("c", 3), part("d", 4)],
  connections: [edge("a", "d"), edge("b", "c"), edge("c", "d")],
};
const fanInLayers = layers(fanIn.parts, fanIn.connections);
same(fanInLayers.get("a"), 0, "a part nobody feeds is in the first column");
same(fanInLayers.get("b"), 0, "a second unfed part shares the first column");
same(fanInLayers.get("c"), 1, "a part behind one other is in the second column");
same(fanInLayers.get("d"), 2, "fan-in puts a part behind its deepest feeder");

const placed = positions(fanIn.parts, fanInLayers);
ok(placed.pos.get("a").x === placed.pos.get("b").x, "one column shares one x");
ok(placed.pos.get("c").x > placed.pos.get("a").x, "the second column sits to the right of the first");
ok(placed.pos.get("d").x > placed.pos.get("c").x, "the third column sits to the right of the second");
ok(placed.pos.get("a").y !== placed.pos.get("b").y, "two parts in a column do not overlap");
ok(placed.pos.get("a").y < placed.pos.get("b").y, "a column keeps the order it was given");
ok(placed.pos.get("c").y > placed.pos.get("a").y && placed.pos.get("c").y < placed.pos.get("b").y,
  "a single-part column is centred against a two-part column");
ok(placed.width > 0 && placed.height > 0, "the frame has a size");

const drawing = sketch(fanIn);
ok((drawing.match(/<rect /g) ?? []).length === 4, "one box per part");
ok((drawing.match(/<line /g) ?? []).length === 3, "one arrow per connection");
ok((drawing.match(/class="box bottleneck"/g) ?? []).length === 1, "one box is marked the bottleneck");
ok(drawing.includes("capacity 3"), "the caption carries the capacity");
ok(drawing.includes("bottleneck c"), "the caption names the bottleneck");
ok(drawing.includes('viewBox="0 0 '), "the drawing carries a viewBox");

// A cycle has no longest path. The pass count is what stops it, and the
// drawing still has to come out.
const cycle = {
  name: "t", capacity: null, bottleneck: [],
  parts: [part("a", 1), part("b", 2)],
  connections: [edge("a", "b"), edge("b", "a")],
};
// Reaching this line at all is half the assertion: a cycle has no longest
// path, so only the pass count stops the loop. The numbers it settles on are
// meaningless, which is the honest thing to pin. They must be finite, and
// every part must still have one.
const cycleLayers = layers(cycle.parts, cycle.connections);
same(cycleLayers.size, 2, "a cycle still gives every part a layer");
ok(Number.isFinite(cycleLayers.get("a")), "a cycle leaves the first part a finite layer");
ok(Number.isFinite(cycleLayers.get("b")), "a cycle leaves the second part a finite layer");
ok(sketch(cycle).includes("<svg"), "a cycle still draws");
ok(sketch(cycle).includes("capacity not computed"), "a missing capacity is said rather than printed");

// An edge naming a part that is not in the drawing is dropped. Keeping it
// would ask for the position of something that was never placed.
const dangling = {
  name: "t", capacity: 1, bottleneck: [],
  parts: [part("a", 1), part("b", 2)],
  connections: [edge("a", "b"), edge("a", "ghost"), edge("ghost", "b")],
};
same(layers(dangling.parts, dangling.connections).size, 2, "an absent part gets no layer");
ok((sketch(dangling).match(/<line /g) ?? []).length === 1, "an edge to an absent part is not drawn");
ok(sketch(dangling).includes("<svg"), "an edge to an absent part does not stop the drawing");

// Escaping again, on the other side. Everything interpolated into the markup
// comes from the projection, which carries whatever the model file says.
const hostile = {
  name: '<script>x</script>',
  capacity: '<b>3</b>',
  bottleneck: [{ id: "a", name: "<em>a</em>" }],
  parts: [{ id: "a", name: "<img>", attributes: [{ name: "throughput", value: '"1"' }] }],
  connections: [],
};
const escaped = sketch(hostile);
ok(!escaped.includes("<script>"), "a part's own name cannot open a tag");
ok(!escaped.includes("<img>"), "a child part's name cannot open a tag");
ok(!escaped.includes("<b>3</b>"), "a capacity cannot open a tag");
ok(!escaped.includes("<em>a</em>"), "a bottleneck name cannot open a tag");
ok(escaped.includes("&lt;img&gt;"), "the child part's name is escaped rather than dropped");
ok(escaped.includes("&lt;b&gt;3&lt;/b&gt;"), "the capacity is escaped rather than dropped");

// A part with nothing inside it, and an attribute the analysis left unset.
const bare = { name: "empty", capacity: null, bottleneck: [], parts: [], connections: [] };
ok(!sketch(bare).includes("<svg"), "a part with no children draws no frame");
ok(sketch(bare).includes("has no child parts to draw"), "a part with no children says so");

const unset = {
  name: "t", capacity: 2, bottleneck: [],
  parts: [{ id: "a", name: "a", attributes: [{ name: "throughput", value: null }, { name: "rate", value: 7 }] }],
  connections: [],
};
ok(!sketch(unset).includes("throughput"), "an unset attribute is left out of the caption");
ok(sketch(unset).includes("rate 7"), "a set attribute is written into the caption");

// ------------------------------------------------------------------ done --

if (failures.length > 0) {
  for (const failure of failures) process.stdout.write(`FAIL ${failure}\n`);
  process.stdout.write(`${failures.length} of ${checks} checks failed\n`);
  process.exit(1);
}
process.stdout.write(`${checks} checks passed\n`);
