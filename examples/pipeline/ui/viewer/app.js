// app.js: the model viewer. It renders what the router serves and computes
// nothing (AD-0026): the text pane shows Model.text, the edit panel shows
// the literals the projection marks editable, the sketch shows the root
// part's children, connections, capacity and bottleneck, and each
// requirement shows the verdict and reason the capacity service returned.

import { query, subscribe } from "/shared/graphql.js";
import { render as renderText, esc } from "./tokeniser.js";
import { render as renderSketch } from "./sketch.js";

const VIEWER_QUERY = `query ViewerApp {
  model {
    version
    text
    roots {
      id shortName name capacity
      parts { id shortName name attributes { name value unit editable } }
      connections { id from to }
      bottleneck { id name }
    }
    requirements {
      id shortName name text quantity comparison limit limitUnit limitEditable
      subject { id name }
      derivedFrom { shortName } derives { shortName }
      satisfiedBy { name } verifiedBy { shortName }
      verdict verdictReason
    }
  }
}`;
const SET_ATTRIBUTE = `mutation SetAttribute($partId: ID!, $name: String!, $value: Float!) {
  setAttribute(partId: $partId, name: $name, value: $value) { id }
}`;
const SET_LIMIT = `mutation SetLimit($requirementId: ID!, $value: Float!) {
  setLimit(requirementId: $requirementId, value: $value) { id }
}`;
const RESET = `mutation Reset { resetModel { version } resetDocument { version } }`;
const MODEL_CHANGED = `subscription ModelChanged { modelChanged }`;

const COMPARISON = { GE: ">=", GT: ">", LE: "<=", LT: "<", EQ: "=" };
const el = (id) => document.getElementById(id);
let served = null; // the last data the router served, the only thing ever rendered

function status(message, isError = false) {
  el("status").textContent = message;
  el("status").classList.toggle("error", isError);
}

// evaluated says whether the analysis reached a verdict. A limit control is
// offered only on such requirements, which keeps PIPE-R2's literal limit
// read-only (SR-13) without the app naming PIPE-R2.
const evaluated = (r) => r.verdict === "PASS" || r.verdict === "FAIL";

function input(key, value, unit) {
  return `<input type="number" min="0" step="any" data-key="${esc(key)}" data-served="${esc(value)}" value="${esc(value)}">`
    + (unit ? ` <span class="unit">${esc(unit)}</span>` : "");
}

function renderInputs(model) {
  const rows = [];
  for (const root of model.roots) {
    for (const part of root.parts) {
      for (const a of part.attributes.filter((x) => x.editable)) {
        rows.push(`<label>${esc(part.name)} <span class="pencil">${esc(part.shortName ?? part.id)}</span> ${esc(a.name)} `
          + input(`attribute|${part.id}|${a.name}`, a.value ?? "", a.unit) + `</label>`);
      }
    }
  }
  for (const r of model.requirements.filter((x) => x.limitEditable && evaluated(x))) {
    rows.push(`<label><span class="pencil">${esc(r.shortName ?? r.id)}</span> limit `
      + input(`limit|${r.id}`, r.limit ?? "", r.limitUnit) + `</label>`);
  }
  return rows.join("");
}

function renderRequirement(r) {
  const names = (list) => list.map((x) => esc(x.shortName ?? x.name)).join(", ");
  const meta = [
    r.subject ? `subject ${esc(r.subject.name)}` : "",
    r.derivedFrom.length ? `derived from ${names(r.derivedFrom)}` : "",
    r.derives.length ? `derives ${names(r.derives)}` : "",
    r.satisfiedBy.length ? `satisfied by ${names(r.satisfiedBy)}` : "",
    r.verifiedBy.length ? `verified by ${names(r.verifiedBy)}` : "",
  ].filter(Boolean).join(" | ");
  const constraint = r.quantity
    ? `${esc(r.quantity)} ${COMPARISON[r.comparison] ?? ""} ${esc(r.limit ?? "")} ${esc(r.limitUnit ?? "")}`
    : "";
  return `<section class="req${r.verdict === "FAIL" ? " fail" : ""}" data-id="${esc(r.id)}">
    <h3><span class="short">${esc(r.shortName ?? r.id)}</span> ${esc(r.name)}</h3>
    ${r.text ? `<p class="reqtext">${esc(r.text)}</p>` : ""}
    <p class="constraint">${constraint}</p>
    <p class="meta pencil">${meta}</p>
    <p class="verdict"><b>${esc(r.verdict)}</b> ${esc(r.verdictReason)}</p>
  </section>`;
}

function render(data) {
  served = data;
  const { model } = data;
  const focused = document.activeElement?.dataset?.key ?? null;
  el("version").textContent = `version ${model.version}`;
  el("text").innerHTML = renderText(model.text);
  el("inputs").innerHTML = renderInputs(model);
  el("sketch").innerHTML = model.roots.map(renderSketch).join("");
  el("requirements").innerHTML = model.requirements.map(renderRequirement).join("");
  if (focused !== null) el("inputs").querySelector(`[data-key="${CSS.escape(focused)}"]`)?.focus();
}

async function refresh() {
  try {
    render(await query(VIEWER_QUERY));
    status("");
  } catch (err) {
    status(err.message, true);
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

async function onReset() {
  try {
    await query(RESET);
    status("");
  } catch (err) {
    status(err.message, true);
  }
}

el("inputs").addEventListener("change", onChange);
el("reset").addEventListener("click", onReset);
subscribe(MODEL_CHANGED, refresh, (err) => status(`live updates: ${err.message}`, true));
refresh();
