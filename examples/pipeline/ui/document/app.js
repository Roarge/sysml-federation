// app.js: the requirements document. The document service owns the tree and
// its numbering (AD-0025). The text, relationships, current values and
// verdicts arrive through the same query from the other two services. The
// app renders what is served, sends one mutation per editorial action and
// re-renders only when a version event arrives (AD-0014).

import { query, subscribe } from "/shared/graphql.js";

const Sortable = window.Sortable;

// MAX_DEPTH must match the number of node levels the query below selects, and
// the sentence renderNode shows at that level spells the same number out.
const MAX_DEPTH = 6;

// The tree is fetched MAX_DEPTH levels deep, deeper than any of the demo's
// own material nests. The deepest level asks for its children's ids and
// nothing else, which costs one field and is what lets the app tell a node
// with nothing under it from one whose children were not fetched. Without
// that the app would have to treat both the same and would either throw on
// the missing field or quietly drop a person's work into a level it never
// shows again.
const DOCUMENT_QUERY = `query DocumentApp {
  model { version requirements { id shortName name included } }
  document {
    version
    nodes {
      ...NodeFields
      children { ...NodeFields
        children { ...NodeFields
          children { ...NodeFields
            children { ...NodeFields
              children { ...NodeFields children { id } } } } } }
    }
  }
}
fragment NodeFields on Node {
  id kind number text
  requirement {
    id shortName name text quantity comparison limit limitUnit limitEditable
    derivedFrom { shortName name } derives { shortName name }
    satisfiedBy { name } verifiedBy { shortName name }
    verdict verdictReason
    subject { id name capacity attributes { name value unit editable } }
  }
}`;
const MOVE_NODE = `mutation MoveNode($id: ID!, $parentId: ID, $index: Int!) {
  moveNode(id: $id, parentId: $parentId, index: $index) { version }
}`;
const INSERT_HEADING = `mutation InsertHeading($aboveId: ID!, $text: String!) {
  insertHeading(aboveId: $aboveId, text: $text) { version }
}`;
const ADD_PROSE = `mutation AddProse($parentId: ID, $index: Int!, $text: String!) {
  addProse(parentId: $parentId, index: $index, text: $text) { version }
}`;
const EDIT_TEXT = `mutation EditText($id: ID!, $text: String!) {
  editText(id: $id, text: $text) { version }
}`;
const EXCLUDE = `mutation ExcludeRequirement($requirementId: ID!) {
  excludeRequirement(requirementId: $requirementId) { version }
}`;
const INCLUDE = `mutation IncludeRequirement($requirementId: ID!) {
  includeRequirement(requirementId: $requirementId) { version }
}`;
const SET_ATTRIBUTE = `mutation SetAttribute($partId: ID!, $name: String!, $value: Float!) {
  setAttribute(partId: $partId, name: $name, value: $value) { id }
}`;
const SET_LIMIT = `mutation SetLimit($requirementId: ID!, $value: Float!) {
  setLimit(requirementId: $requirementId, value: $value) { id }
}`;
const RESET = `mutation Reset { resetModel { version } resetDocument { version } }`;
const MODEL_CHANGED = `subscription ModelChanged { modelChanged }`;
const DOCUMENT_CHANGED = `subscription DocumentChanged { documentChanged }`;

const COMPARISON = { GE: ">=", GT: ">", LE: "<=", LT: "<", EQ: "=" };
const el = (id) => document.getElementById(id);
let served = null; // the last data the router served, the only thing ever rendered

// generation counts the refreshes. Both subscriptions call refresh, so two
// answers can be in flight at once and the router is free to return them in
// either order. The counter says which answer is still the latest, so a slow
// one is dropped rather than painted over a newer one.
let generation = 0;

// dragging and pending keep a refresh out of a drag. Replacing the tree while
// the drag library holds an element leaves that element behind, so the drop
// puts a node the person can no longer see back into the new markup and the
// move that follows carries an index computed against a tree that is gone.
// A refresh that lands mid-drag is held here and rendered when the drop is
// over. sortables holds the live drag instances, each destroyed before the
// markup it watches is replaced.
let dragging = false;
let pending = null;
let sortables = [];

function esc(s) {
  return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}

function status(message, isError = false) {
  el("status").textContent = message;
  el("status").classList.toggle("error", isError);
}

// refuse puts the served value back and says why the entry was not sent.
function refuse(field, message) {
  status(message, true);
  field.value = field.dataset.served;
}

// evaluated says whether the analysis reached a verdict. The current value
// and the limit control appear only on such rows, which keeps PIPE-R2's
// literal limit read-only (SR-38) without the app naming PIPE-R2.
const evaluated = (r) => r.verdict === "PASS" || r.verdict === "FAIL";

function numberInput(key, value, unit) {
  return `<input type="number" min="0" step="any" data-key="${esc(key)}" data-served="${esc(value)}" value="${esc(value)}">`
    + (unit ? ` <span class="unit">${esc(unit)}</span>` : "");
}

function renderRequirement(r) {
  const names = (list) => list.map((x) => esc(x.shortName ?? x.name)).join(", ");
  const unit = r.limitUnit ? ` ${esc(r.limitUnit)}` : "";
  const rows = [];
  if (r.limitEditable && evaluated(r)) rows.push(["limit", numberInput(`limit|${r.id}`, r.limit ?? "", r.limitUnit)]);
  else rows.push(["limit", `${esc(r.quantity ?? "")} ${COMPARISON[r.comparison] ?? ""} ${esc(r.limit ?? "")}${unit}`]);
  if (r.derivedFrom.length) rows.push(["derived from", names(r.derivedFrom)]);
  if (r.derives.length) rows.push(["derives", names(r.derives)]);
  if (r.satisfiedBy.length) rows.push(["satisfied by", names(r.satisfiedBy)]);
  if (r.verifiedBy.length) rows.push(["verified by", names(r.verifiedBy)]);
  if (evaluated(r) && r.subject && r.subject.capacity !== null) {
    rows.push(["current value", `${esc(r.quantity)} ${esc(r.subject.capacity)}`]);
  }
  for (const a of r.subject ? r.subject.attributes.filter((x) => x.editable) : []) {
    rows.push([`${esc(r.subject.name)} ${esc(a.name)}`, numberInput(`attribute|${r.subject.id}|${a.name}`, a.value ?? "", a.unit)]);
  }
  return `<div class="requirement">
    <p class="head"><span class="short">${esc(r.shortName ?? r.id)}</span> ${esc(r.name)}</p>
    ${r.text ? `<p class="reqtext">${esc(r.text)}</p>` : ""}
    <dl>${rows.map(([k, v]) => `<dt>${k}</dt><dd>${v}</dd>`).join("")}</dl>
    <p class="verdict"><b>${esc(r.verdict)}</b> ${esc(r.verdictReason)}</p>
  </div>`;
}

// renderNode draws one node and, above the deepest fetched level, its
// children. At that level there is no child list to draw and no place to put
// a new paragraph, so "Add prose" is withheld and the row says what it is not
// showing. Prose is a leaf wherever it sits (phase decision 6).
function renderNode(node, depth) {
  const req = node.requirement;
  const failed = req !== null && req !== undefined && req.verdict === "FAIL";
  const number = `<span class="number">${esc(node.number ?? "")}</span>`;
  const kids = node.children ?? [];
  const deepest = depth >= MAX_DEPTH;
  const leaf = node.kind === "PROSE" || deepest;
  let content;
  if (node.kind === "REQUIREMENT") {
    content = req ? renderRequirement(req) : `<p class="pencil">This requirement is not in the model.</p>`;
  } else if (node.kind === "HEADING") {
    content = `<h2 class="text" contenteditable="true" data-id="${esc(node.id)}" data-original="${esc(node.text ?? "")}">${esc(node.text ?? "")}</h2>`;
  } else {
    content = `<p class="text prose" contenteditable="true" data-id="${esc(node.id)}" data-original="${esc(node.text ?? "")}">${esc(node.text ?? "")}</p>`;
  }
  const tools = [
    `<button type="button" data-act="heading" data-id="${esc(node.id)}">Heading above</button>`,
    leaf ? "" : `<button type="button" data-act="prose" data-id="${esc(node.id)}" data-count="${kids.length}">Add prose</button>`,
    req ? `<button type="button" data-act="exclude" data-req="${esc(req.id)}">Exclude</button>` : "",
  ].join("");
  const children = leaf ? "" : `<ol class="nodes" data-parent="${esc(node.id)}">${kids.map((c) => renderNode(c, depth + 1)).join("")}</ol>`;
  const truncated = deepest && node.kind !== "PROSE" ? `<p class="pencil deeper">${kids.length > 0
    ? "More sits below this item than the document shows. The document is shown six levels deep."
    : "The document is shown six levels deep, so nothing can be added below this item."}</p>` : "";
  return `<li class="node kind-${node.kind.toLowerCase()}${failed ? " fail" : ""}" data-id="${esc(node.id)}">
    <div class="row"><span class="grip" title="Drag to move">::</span>${number}<div class="content">${content}</div><span class="tools">${tools}</span></div>${children}${truncated}
  </li>`;
}

function renderExcluded(requirements) {
  const hidden = requirements.filter((r) => !r.included);
  if (hidden.length === 0) return `<p class="pencil">No requirement is excluded.</p>`;
  return hidden.map((r) => `<p class="excluded-item"><span class="short">${esc(r.shortName ?? r.id)}</span> ${esc(r.name)} `
    + `<button type="button" data-act="include" data-req="${esc(r.id)}">Restore</button></p>`).join("");
}

function render(data) {
  served = data;
  // A drag owns the markup until the drop. Rendering under it would strand
  // the element the library is holding, so the data waits.
  if (dragging) {
    pending = data;
    return;
  }
  pending = null;
  const focused = document.activeElement?.dataset?.key ?? null;
  el("versions").textContent = `document version ${data.document.version}, model version ${data.model.version}`;
  for (const sortable of sortables) sortable.destroy();
  sortables = [];
  el("tree").innerHTML = data.document.nodes.map((node) => renderNode(node, 1)).join("");
  el("add-prose").dataset.count = data.document.nodes.length;
  el("excluded").innerHTML = renderExcluded(data.model.requirements);
  for (const list of [el("tree"), ...el("tree").querySelectorAll("ol.nodes")]) {
    sortables.push(new Sortable(list, {
      group: "nodes", handle: ".grip", animation: 150, fallbackOnBody: true, swapThreshold: 0.65, onStart, onEnd,
    }));
  }
  if (focused !== null) el("tree").querySelector(`[data-key="${CSS.escape(focused)}"]`)?.focus();
}

// flush renders the answer that arrived during a drag. It runs from a timer
// rather than from onEnd itself, so the library has finished with the element
// before the markup holding it is replaced.
function flush() {
  if (pending !== null) render(pending);
}

async function refresh() {
  const mine = ++generation;
  try {
    const data = await query(DOCUMENT_QUERY);
    if (mine !== generation) return;
    render(data);
    status("");
  } catch (err) {
    if (mine !== generation) return;
    status(err.message, true);
  }
}

// mutate sends one mutation and renders nothing from its answer. The version
// event that follows triggers the refetch. On refusal the served state is
// re-rendered, so a drag the service refused is visibly undone.
async function mutate(operation, variables) {
  try {
    await query(operation, variables);
    status("");
  } catch (err) {
    status(err.message, true);
    if (served !== null) render(served);
  }
}

// onStart claims the markup for the drag, so that a refresh arriving before
// the drop waits rather than pulling the dragged element out from under it.
function onStart() {
  dragging = true;
}

// onEnd maps one drop to one moveNode: the item's id, the list it landed in
// (its data-parent, empty at the root) and its final index in that list. The
// index is read from the tree the person was looking at, which is why nothing
// was allowed to re-render underneath them.
function onEnd(evt) {
  dragging = false;
  const parent = evt.to.dataset.parent;
  const moved = evt.from !== evt.to || evt.oldIndex !== evt.newIndex;
  if (moved) {
    mutate(MOVE_NODE, { id: evt.item.dataset.id, parentId: parent === "" ? null : parent, index: evt.newIndex });
  }
  setTimeout(flush, 0);
}

async function onClick(event) {
  const button = event.target.closest("button[data-act]");
  if (button === null) return;
  const { act, id, req, count } = button.dataset;
  if (act === "heading") {
    const text = window.prompt("Heading text");
    if (text) await mutate(INSERT_HEADING, { aboveId: id, text: text.trim() });
  } else if (act === "prose") {
    const text = window.prompt("Paragraph text");
    if (text) await mutate(ADD_PROSE, { parentId: id ?? null, index: Number(count), text: text.trim() });
  } else if (act === "exclude") {
    await mutate(EXCLUDE, { requirementId: req });
  } else if (act === "include") {
    await mutate(INCLUDE, { requirementId: req });
  } else if (act === "reset") {
    await mutate(RESET, {});
  }
}

function onFocusOut(event) {
  const field = event.target;
  if (!(field instanceof HTMLElement) || !field.matches("[contenteditable]")) return;
  const text = field.textContent.trim();
  if (text !== "" && text !== field.dataset.original) mutate(EDIT_TEXT, { id: field.dataset.id, text });
  else field.textContent = field.dataset.original;
}

function onKeyDown(event) {
  if (event.key === "Enter" && !event.shiftKey && event.target.matches("[contenteditable]")) {
    event.preventDefault();
    event.target.blur();
  }
}

async function onChange(event) {
  const field = event.target;
  const key = field.dataset.key;
  if (key === undefined) return;
  // A number input reports an entry it could not parse as an empty value and
  // sets badInput, so the text is no longer there to be quoted. The field's
  // own validity is what tells that apart from a field a person cleared.
  if (field.validity.badInput) {
    refuse(field, "the entry is not a number, the served value stands");
    return;
  }
  const value = Number(field.value);
  if (field.value.trim() === "" || !Number.isFinite(value) || value < 0) {
    refuse(field, `"${field.value}" is not a finite, non-negative number, the served value stands`);
    return;
  }
  const [kind, target, name] = key.split("|");
  try {
    if (kind === "attribute") await query(SET_ATTRIBUTE, { partId: target, name, value });
    else await query(SET_LIMIT, { requirementId: target, value });
    status("");
  } catch (err) {
    field.value = field.dataset.served;
    status(err.message, true);
  }
}

if (typeof Sortable !== "function") status("Sortable.min.js did not load, drag and drop is unavailable", true);
document.addEventListener("click", onClick);
el("tree").addEventListener("change", onChange);
el("tree").addEventListener("focusout", onFocusOut);
el("tree").addEventListener("keydown", onKeyDown);
subscribe(MODEL_CHANGED, refresh, (err) => status(`live updates: ${err.message}`, true));
subscribe(DOCUMENT_CHANGED, refresh, (err) => status(`live updates: ${err.message}`, true));
refresh();
