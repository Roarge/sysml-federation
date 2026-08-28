// app.js: the requirements document. The document service owns the tree and
// its numbering (AD-0025). The text, relationships, current values and
// verdicts arrive through the same query from the other two services. The
// app renders what is served, sends one mutation per editorial action and
// re-renders only when a version event arrives (AD-0014).

import { query, subscribe } from "/shared/graphql.js";

const Sortable = window.Sortable;

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
              children { ...NodeFields } } } } }
    }
  }
}
fragment NodeFields on Node {
  id kind number text
  requirement {
    id shortName name text quantity comparison limit limitUnit limitEditable
    derivedFrom { shortName } derives { shortName }
    satisfiedBy { name } verifiedBy { shortName }
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

function esc(s) {
  return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}

function status(message, isError = false) {
  el("status").textContent = message;
  el("status").classList.toggle("error", isError);
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

function renderNode(node) {
  const req = node.requirement;
  const failed = req !== null && req !== undefined && req.verdict === "FAIL";
  const number = `<span class="number">${esc(node.number ?? "")}</span>`;
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
    node.kind === "PROSE" ? "" : `<button type="button" data-act="prose" data-id="${esc(node.id)}" data-count="${node.children.length}">Add prose</button>`,
    req ? `<button type="button" data-act="exclude" data-req="${esc(req.id)}">Exclude</button>` : "",
  ].join("");
  const children = node.kind === "PROSE" ? "" : `<ol class="nodes" data-parent="${esc(node.id)}">${node.children.map(renderNode).join("")}</ol>`;
  return `<li class="node kind-${node.kind.toLowerCase()}${failed ? " fail" : ""}" data-id="${esc(node.id)}">
    <div class="row"><span class="grip" title="Drag to move">::</span>${number}<div class="content">${content}</div><span class="tools">${tools}</span></div>${children}
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
  const focused = document.activeElement?.dataset?.key ?? null;
  el("versions").textContent = `document version ${data.document.version}, model version ${data.model.version}`;
  el("tree").innerHTML = data.document.nodes.map(renderNode).join("");
  el("add-prose").dataset.count = data.document.nodes.length;
  el("excluded").innerHTML = renderExcluded(data.model.requirements);
  for (const list of [el("tree"), ...el("tree").querySelectorAll("ol.nodes")]) {
    new Sortable(list, { group: "nodes", handle: ".grip", animation: 150, fallbackOnBody: true, swapThreshold: 0.65, onEnd });
  }
  if (focused !== null) el("tree").querySelector(`[data-key="${CSS.escape(focused)}"]`)?.focus();
}

async function refresh() {
  try {
    render(await query(DOCUMENT_QUERY));
    status("");
  } catch (err) {
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

// onEnd maps one drop to one moveNode: the item's id, the list it landed in
// (its data-parent, empty at the root) and its final index in that list.
function onEnd(evt) {
  if (evt.from === evt.to && evt.oldIndex === evt.newIndex) return;
  const parent = evt.to.dataset.parent;
  mutate(MOVE_NODE, { id: evt.item.dataset.id, parentId: parent === "" ? null : parent, index: evt.newIndex });
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
  const value = Number(field.value);
  if (field.value.trim() === "" || !Number.isFinite(value) || value < 0) {
    status(`"${field.value}" is not a finite, non-negative number, the served value stands`, true);
    field.value = field.dataset.served;
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
