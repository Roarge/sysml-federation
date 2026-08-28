// sketch.js: draws one part's wiring as a left-to-right graph (SR-12).
//
// Boxes are the part's child parts, arrows its connections, the caption its
// capacity, and a box in the bottleneck set is outlined red. All of it comes
// from the projection; the only thing computed here is where to draw it.
// The markup is built as a string and inserted with innerHTML, which puts
// the elements in the SVG namespace without naming its URL (SR-10).

import { esc } from "./tokeniser.js";

const BOX_W = 116;
const BOX_H = 52;
const COL_GAP = 56;
const ROW_GAP = 20;
const MARGIN = 16;
const CAPTION_H = 30;

// layers assigns each part the length of the longest path leading to it: a
// part nobody feeds is in layer 0, every other part sits one layer to the
// right of the deepest part that feeds it. The pass count bounds the loop,
// so a cycle in the wiring stops after parts.length passes.
export function layers(parts, connections) {
  const layer = new Map(parts.map((p) => [p.id, 0]));
  const edges = connections.filter((c) => layer.has(c.from) && layer.has(c.to));
  for (let pass = 0; pass < parts.length; pass++) {
    let changed = false;
    for (const e of edges) {
      const want = layer.get(e.from) + 1;
      if (want > layer.get(e.to)) {
        layer.set(e.to, want);
        changed = true;
      }
    }
    if (!changed) break;
  }
  return layer;
}

export function positions(parts, layer) {
  const columns = new Map();
  for (const p of parts) {
    const l = layer.get(p.id);
    if (!columns.has(l)) columns.set(l, []);
    columns.get(l).push(p);
  }
  const depth = Math.max(...columns.keys()) + 1;
  const rows = Math.max(...[...columns.values()].map((c) => c.length));
  const pos = new Map();
  for (const [l, column] of columns) {
    const offset = ((rows - column.length) * (BOX_H + ROW_GAP)) / 2;
    column.forEach((p, i) => pos.set(p.id, {
      x: MARGIN + l * (BOX_W + COL_GAP),
      y: MARGIN + offset + i * (BOX_H + ROW_GAP),
    }));
  }
  return {
    pos,
    width: MARGIN * 2 + depth * BOX_W + (depth - 1) * COL_GAP,
    height: MARGIN * 2 + rows * BOX_H + (rows - 1) * ROW_GAP + CAPTION_H,
  };
}

export function render(part) {
  const parts = part.parts;
  if (parts.length === 0) return `<p class="pencil">${esc(part.name)} has no child parts to draw.</p>`;
  const { pos, width, height } = positions(parts, layers(parts, part.connections));
  const cut = new Set(part.bottleneck.map((b) => b.id));
  const out = [
    `<svg class="sketch" viewBox="0 0 ${width} ${height}" width="${width}" height="${height}" role="img" aria-label="wiring of ${esc(part.name)}">`,
    `<defs><marker id="arrow" viewBox="0 0 8 8" refX="8" refY="4" markerWidth="8" markerHeight="8" orient="auto"><path d="M0,0 L8,4 L0,8 z" class="arrowhead"/></marker></defs>`,
  ];
  for (const c of part.connections) {
    const a = pos.get(c.from);
    const b = pos.get(c.to);
    if (a === undefined || b === undefined) continue;
    out.push(`<line class="edge" x1="${a.x + BOX_W}" y1="${a.y + BOX_H / 2}" x2="${b.x}" y2="${b.y + BOX_H / 2}" marker-end="url(#arrow)"/>`);
  }
  for (const p of parts) {
    const { x, y } = pos.get(p.id);
    const values = p.attributes.filter((a) => a.value !== null).map((a) => `${a.name} ${a.value}`).join(", ");
    out.push(`<g class="box${cut.has(p.id) ? " bottleneck" : ""}"><rect x="${x}" y="${y}" width="${BOX_W}" height="${BOX_H}" rx="3"/>`
      + `<text class="label" x="${x + BOX_W / 2}" y="${y + 21}">${esc(p.name)}</text>`
      + `<text class="value" x="${x + BOX_W / 2}" y="${y + 39}">${esc(values)}</text></g>`);
  }
  const capacity = part.capacity === null ? "capacity not computed" : `capacity ${part.capacity}`;
  const bottleneck = part.bottleneck.length === 0 ? "" : `, bottleneck ${part.bottleneck.map((b) => esc(b.name)).join(", ")}`;
  out.push(`<text class="caption" x="${MARGIN}" y="${height - 10}">${capacity}${bottleneck}</text></svg>`);
  return out.join("");
}
