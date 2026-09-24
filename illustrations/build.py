#!/usr/bin/env python3
"""Build the published PDFs and article images from the sources beside this file.

Each source is a standalone HTML page that states its own page size in an
@page rule. Headless Chrome prints it as it is, pdfunite joins the pages of a
set into one PDF, and Pillow cuts the article images out of screenshots.

Usage:
    python3 illustrations/build.py [SOURCE ...] [--only pdf|png|all]
                                   [--out DIR] [--work DIR] [--chrome PATH]

With no SOURCE every output is built. Naming sources rebuilds every PDF that
contains one of them, whole, and every image taken from one of them, and leaves
every other file under the output directory as it was:

    python3 illustrations/build.py illustrations/architecture/v4-deployment.html

Requires Python 3.10 or later, Chrome or Chromium, pdfunite and pdfinfo
(poppler) for the PDFs, Pillow for the images, and a network connection for
the fonts.
"""

from __future__ import annotations

import argparse
import html
import json
import re
import shutil
import subprocess
import sys
from collections.abc import Callable
from pathlib import Path

try:
    from PIL import Image
except ImportError:  # only the images need Pillow, and main() says so
    Image = None

REPO = Path(__file__).resolve().parent.parent
SRC = REPO / "illustrations"

PAPER = "oklch(98% 0.005 90)"

# The Google Fonts stylesheet the boards link (the L0 sheet links its Source Sans 3
# part alone), and the faces it has to deliver.
FONTS_CSS = (
    "https://fonts.googleapis.com/css2?family=Patrick+Hand"
    "&family=Source+Sans+3:wght@400;600&display=swap"
)
FACES = [("Source Sans 3", 400), ("Source Sans 3", 600), ("Patrick Hand", 400)]

CHROME_NAMES = ("google-chrome", "google-chrome-stable", "chromium", "chromium-browser")
MAC_CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

# Set by main(): the browser, and the directory for intermediates and its profile.
CHROME = ""
WORK = REPO / ".cache" / "illustrations"

# ------------------------------------------------------------------- outputs

ARCH_PAGES = [
    "overview",
    "v1-context",
    "v2-composition",
    "v3-runtime",
    "v4-deployment",
    "v5-adapter",
]

STORY_PAGES = [
    "overview",
    "us01-launch",
    "us02-read-the-model",
    "us03-raise-a-non-bottleneck",
    "us04-raise-the-bottleneck",
    "us05-tighten-the-limit",
    "us06-read-the-document",
    "us07-reorder-and-nest",
    "us08-shape-the-document",
    "us09-change-from-the-document",
    "us10-change-from-the-viewer",
    "us11-query-the-graph",
    "us12-reset",
]


def pdf_sets() -> dict[str, list[str]]:
    """Each PDF under the output directory, with its pages in order."""
    sets = {
        "architecture/architecture-views.pdf": [
            f"architecture/{n}.html" for n in ARCH_PAGES
        ]
    }
    for page in sorted((SRC / "a3").glob("*.html")):
        sets[f"a3/{page.stem}.pdf"] = [f"a3/{page.name}"]
    sets["stories/use-cases.pdf"] = [f"stories/{n}.html" for n in STORY_PAGES]
    return sets


L0 = "a3/L0-federating-a-systems-model.html"
L2B = "a3/L2b-pipeline-example-capacity-and-verdicts.html"

# (source, image under img/): the whole board.
WHOLE_BOARDS = [
    ("architecture/overview.html", "architecture-five-views.png"),
    (L0, "a3-l0-model-side.png"),
    (L2B, "a3-l2b-model-side.png"),
]

# (source, region of its first <svg> in viewBox units as x, y, w, h, or None for
# the whole viewBox, image under img/)
SVG_PANELS = [
    ("architecture/v1-context.html", (0, 0, 1000, 720), "v1-context.png"),
    ("architecture/v2-composition.html", None, "v2-composition.png"),
    # y 22 (not 16) clears the descenders of the heading above the panel.
    (
        "architecture/v2-composition.html",
        (670, 22, 320, 582),
        "v2-merged-requirement.png",
    ),
    # y 186 (not 218) takes the three fetch arrows whole rather than cutting
    # their heads; it starts just below the router box's bottom edge at 182.75.
    (
        "architecture/v2-composition.html",
        (10, 186, 664, 530),
        "v2-subgraph-schemas.png",
    ),
    ("architecture/v3-runtime.html", None, "v3-runtime.png"),
    ("architecture/v3-runtime.html", (10, 582, 335, 138), "v3-nothing-moves.png"),
    ("architecture/v4-deployment.html", None, "v4-deployment.png"),
    ("architecture/v4-deployment.html", (150, 52, 520, 448), "v4-container.png"),
    # The fifth layer row is 642..667, so 502..652 cut it in half.
    ("architecture/v4-deployment.html", (150, 500, 520, 172), "v4-image-layers.png"),
    ("architecture/v5-adapter.html", None, "v5-adapter.png"),
    (L0, (0, 847, 1507, 196), "a3-l0-legend.png"),
    (L2B, (888, 144, 619, 444), "a3-l2b-arithmetic.png"),
    # h 426 (not 450) ends at 606: clear of the third state's boxes (to 603)
    # and above the cap height of "verdicts and reasons" (baseline 621), which
    # belongs to the next block.
    (L2B, (414, 180, 470, 426), "a3-l2b-wiring-states.png"),
    (L2B, (888, 592, 619, 296), "a3-l2b-physical.png"),
    ("stories/overview.html", (0, 0, 1000, 250), "overview-sketch.png"),
]

# Each use case's caption is a <span class="caption"> inside div.sketch, so the
# box of div.sketch takes the sketch and its caption together.
SKETCH = ["div.sketch"]

# (source, CSS selectors whose union box is cut from the board, image under img/).
# A selector that matches nothing is skipped, and the build stops only when
# every selector of an image misses.
HTML_CROPS = [
    ("stories/overview.html", ["div.personas"], "stories-personas.png"),
    ("stories/overview.html", ["div.journey"], "stories-journey.png"),
    ("stories/us01-launch.html", SKETCH, "us01-launch.png"),
    ("stories/us03-raise-a-non-bottleneck.html", SKETCH, "us03-nothing-moves.png"),
    ("stories/us04-raise-the-bottleneck.html", SKETCH, "us04-bottleneck-moves.png"),
    ("stories/us06-read-the-document.html", SKETCH, "us06-document.png"),
    (
        "stories/us09-change-from-the-document.html",
        SKETCH,
        "us09-edit-from-document.png",
    ),
    (
        "stories/us10-change-from-the-viewer.html",
        SKETCH,
        "us10-viewer-to-document.png",
    ),
    ("stories/us11-query-the-graph.html", SKETCH, "us11-query.png"),
    ("stories/us12-reset.html", SKETCH, "us12-reset.png"),
]


def image_sources() -> set[str]:
    return {row[0] for row in WHOLE_BOARDS + SVG_PANELS + HTML_CROPS}


def all_sources() -> set[str]:
    return {p for pages in pdf_sets().values() for p in pages} | image_sources()


# ------------------------------------------------------------------- sources

PAGE_RULE = re.compile(r"@page\s*\{\s*size:\s*(\d+)px\s+(\d+)px")


def page_size(rel: str) -> tuple[int, int]:
    """The page size a source states in its @page rule, in CSS px."""
    src = SRC / rel
    if not src.is_file():
        raise SystemExit(f"illustrations/{rel}: source missing")
    m = PAGE_RULE.search(src.read_text(encoding="utf-8"))
    if not m:
        raise SystemExit(
            f"illustrations/{rel}: no '@page {{ size: Wpx Hpx' rule, "
            "so the page size is unknown"
        )
    return int(m.group(1)), int(m.group(2))


def slug(rel: str) -> str:
    return rel.removesuffix(".html").replace("/", "-")


def body_span(text: str) -> tuple[int, int]:
    """(index just after <body>, index of </body>)."""
    return text.index("<body>") + len("<body>"), text.index("</body>")


# -------------------------------------------------------------------- chrome


def run(cmd: list[str]) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, capture_output=True, text=True, check=False)


def chrome_pdf(page: Path, out: Path) -> None:
    out.parent.mkdir(parents=True, exist_ok=True)
    if out.exists():
        out.unlink()
    res = run(
        [
            CHROME,
            "--headless=new",
            "--no-sandbox",
            "--disable-gpu",
            "--no-pdf-header-footer",
            "--run-all-compositor-stages-before-draw",
            "--virtual-time-budget=10000",
            f"--user-data-dir={WORK / 'chrome-profile'}",
            f"--print-to-pdf={out}",
            page.as_uri(),
        ]
    )
    if not out.exists():
        sys.stderr.write(res.stderr)
        raise SystemExit(f"chrome produced no pdf for {page}")


def chrome_png(page: Path, out: Path, w: int, h: int, scale: float) -> None:
    out.parent.mkdir(parents=True, exist_ok=True)
    if out.exists():
        out.unlink()
    res = run(
        [
            CHROME,
            "--headless=new",
            "--no-sandbox",
            "--disable-gpu",
            "--hide-scrollbars",
            f"--force-device-scale-factor={scale:g}",
            f"--window-size={w},{h}",
            "--virtual-time-budget=10000",
            f"--user-data-dir={WORK / 'chrome-profile'}",
            f"--screenshot={out}",
            page.as_uri(),
        ]
    )
    if not out.exists():
        sys.stderr.write(res.stderr)
        raise SystemExit(f"chrome produced no png for {page}")


def chrome_dump_dom(page: Path, w: int, h: int, want: str) -> str:
    """Group 1 of the pattern `want` in the page's DOM once loaded, unescaped.

    Chrome exits 0 even when it cannot open the file (it dumps its own error
    page instead), so no match is reported as Chrome's failure, with the
    error code from that page when there is one, and Chrome's messages."""
    res = run(
        [
            CHROME,
            "--headless=new",
            "--no-sandbox",
            "--disable-gpu",
            "--hide-scrollbars",
            f"--window-size={w},{h}",
            "--virtual-time-budget=10000",
            f"--user-data-dir={WORK / 'chrome-profile'}",
            "--dump-dom",
            page.as_uri(),
        ]
    )
    m = re.search(want, res.stdout, re.S)
    if m:
        return html.unescape(m.group(1))
    code = re.search(r'"errorCode":"([A-Z_]+)"', res.stdout)
    sys.stderr.write(res.stderr)
    raise SystemExit(
        f"Chrome did not load {page} (exit {res.returncode}"
        + (f", {code.group(1)}" if code else "")
        + "). Anything Chrome printed is above. A Chromium installed as a snap "
        "package reads only some directories, and not the hidden ones at the "
        "top of the home directory: keep the clone and --work elsewhere in the "
        "home directory, or pass --chrome a browser that is not a snap."
    )


# The <pre> that the probe pages' scripts add once loaded, and its text.
PRE = r'<pre id="%s">(.*?)</pre>'


def check_sources(rel: str) -> None:
    """Stop unless Chrome can read the sources where they are.

    The PDFs, the whole boards and the element crops are rendered from the
    files in illustrations/ themselves, and Chrome that cannot read one prints
    or screenshots its own error page and exits 0. main() has read every needed
    source by then, so what can still keep Chrome out is a confinement by
    directory (a snap), which keeps it out of all of them alike: opening one
    and finding its @page rule in the DOM is the test."""
    w, h = page_size(rel)
    chrome_dump_dom(SRC / rel, w, h, PAGE_RULE.pattern)


# --------------------------------------------------------------------- fonts

# Text in each face and weight, laid out before document.fonts.ready is awaited
# so that every face has started loading. document.fonts.check() is no test
# here: it answers true for a family with no @font-face at all.
FONT_PROBE = """<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<link rel="stylesheet" href="{css}">
</head>
<body>
<p style="font-family: 'Source Sans 3'; font-weight: 400">Source Sans 3 regular</p>
<p style="font-family: 'Source Sans 3'; font-weight: 600">Source Sans 3 semibold</p>
<p style="font-family: 'Patrick Hand'; font-weight: 400">Patrick Hand</p>
<script>
window.addEventListener('load', function () {
  document.body.getBoundingClientRect();
  document.fonts.ready.then(function () {
    var faces = [];
    document.fonts.forEach(function (f) {
      faces.push([f.family.replace(/["']/g, ''), String(f.weight), f.status]);
    });
    var pre = document.createElement('pre');
    pre.id = 'fonts';
    pre.textContent = JSON.stringify(faces);
    document.body.appendChild(pre);
  });
});
</script>
</body>
</html>
"""


def covers(weight: str, wanted: int) -> bool:
    """Whether a FontFace weight ("400", "normal", "300 700") includes `wanted`."""
    names = {"normal": "400", "bold": "700"}
    ends = [int(float(names.get(p, p))) for p in weight.split()]
    return min(ends) <= wanted <= max(ends)


def check_fonts() -> None:
    """Stop unless Chrome can load every face the boards are drawn in.

    A face that fails to load is replaced silently by a fallback, which reflows
    the page, so this runs before anything is written."""
    probe = WORK / "fonts.html"
    probe.write_text(FONT_PROBE.replace("{css}", FONTS_CSS), encoding="utf-8")
    faces = json.loads(chrome_dump_dom(probe, 800, 600, PRE % "fonts"))
    missing = [
        f"{family} {weight}"
        for family, weight in FACES
        if not any(
            f == family and s == "loaded" and covers(w, weight) for f, w, s in faces
        )
    ]
    if missing:
        raise SystemExit(
            "the fonts could not be fetched from Google Fonts ("
            + ", ".join(missing)
            + " did not load). The build needs a connection to "
            "fonts.googleapis.com and fonts.gstatic.com. Nothing was written."
        )


# ---------------------------------------------------------------------- pdfs


def page_count(pdf: Path) -> int:
    res = run(["pdfinfo", str(pdf)])
    m = re.search(r"^Pages:\s+(\d+)", res.stdout, re.M)
    if not m:
        sys.stderr.write(res.stderr)
        raise SystemExit(f"pdfinfo could not read {pdf}")
    return int(m.group(1))


def build_pdfs(out_root: Path, chosen: set[str]) -> None:
    """Every PDF with a page among `chosen`, rebuilt whole."""
    for out_rel, pages in pdf_sets().items():
        if not chosen.intersection(pages):
            continue
        parts = []
        for rel in pages:
            part = WORK / "pdf" / f"{slug(rel)}.pdf"
            chrome_pdf(SRC / rel, part)
            # Content taller than the @page size spills onto a second page,
            # which would shift every later page of the set.
            count = page_count(part)
            if count != 1:
                w, h = page_size(rel)
                raise SystemExit(
                    f"illustrations/{rel} prints as {count} pages: its content "
                    f"does not fit the {w} by {h} px page in its @page rule"
                )
            parts.append(part)
        if len(parts) == 1:
            joined = parts[0]
        else:
            joined = WORK / "pdf" / Path(out_rel).name
            res = run(["pdfunite", *map(str, parts), str(joined)])
            if res.returncode != 0:
                sys.stderr.write(res.stderr)
                raise SystemExit(f"pdfunite failed for {out_rel}")
        out = out_root / out_rel
        out.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(joined, out)
        print(shown(out))


# ---------------------------------------------------------------------- pngs

RECT_SCRIPT = """<script>
window.addEventListener('load', function () {
  var sels = %s;
  var box = null;
  sels.forEach(function (s) {
    var e = document.querySelector(s);
    if (!e) { return; }
    var r = e.getBoundingClientRect();
    var b = { l: r.left, t: r.top, r: r.right, b: r.bottom };
    box = box === null ? b : { l: Math.min(box.l, b.l), t: Math.min(box.t, b.t),
                               r: Math.max(box.r, b.r), b: Math.max(box.b, b.b) };
  });
  var pre = document.createElement('pre');
  pre.id = 'rect';
  pre.textContent = JSON.stringify(box);
  document.body.appendChild(pre);
});
</script>
"""


def html_rect(rel: str, selectors: list[str]) -> dict:
    """Union bounding box of the selectors on the board, in CSS px.

    Measured on a copy of the source with a script added to its head; the
    copy is never printed."""
    w, h = page_size(rel)
    text = (SRC / rel).read_text(encoding="utf-8")
    probe = WORK / "pages" / f"{slug(rel)}.rect.html"
    probe.parent.mkdir(parents=True, exist_ok=True)
    script = RECT_SCRIPT % json.dumps(selectors)
    probe.write_text(text.replace("</head>", script + "</head>", 1), encoding="utf-8")
    box = json.loads(chrome_dump_dom(probe, w, h, PRE % "rect"))
    if box is None:
        raise SystemExit(f"selectors matched nothing on {rel}: {selectors}")
    return box


def svg_span(text: str) -> tuple[str, int, int]:
    """(root open tag, index after it, index of its matching </svg>)."""
    i = text.index("<svg")
    j = text.index(">", i) + 1
    depth, k = 1, j
    while depth:
        nxt_open = text.find("<svg", k)
        nxt_close = text.find("</svg>", k)
        if nxt_close == -1:
            raise SystemExit("unbalanced <svg>")
        if nxt_open != -1 and nxt_open < nxt_close:
            depth += 1
            k = nxt_open + 4
        else:
            depth -= 1
            k = nxt_close + 6
    return text[i:j], j, k - 6


def whole_svg_region(rel: str) -> tuple[float, float, float, float]:
    text = (SRC / rel).read_text(encoding="utf-8")
    b0, b1 = body_span(text)
    open_tag, _, _ = svg_span(text[b0:b1])
    vb = re.search(r'viewBox="([^"]*)"', open_tag).group(1).split()
    return (float(vb[0]), float(vb[1]), float(vb[2]), float(vb[3]))


def svg_panel_page(
    rel: str, region: tuple[float, float, float, float]
) -> tuple[Path, int, int]:
    """A page with the source's head whose body is the board's first <svg>,
    narrowed to `region`."""
    text = (SRC / rel).read_text(encoding="utf-8")
    b0, b1 = body_span(text)
    body = text[b0:b1]
    open_tag, j, close = svg_span(body)
    inner = body[j:close]
    x, y, w, h = region
    if "viewBox=" not in open_tag:
        raise SystemExit(f"no viewBox on the first svg of {rel}")
    tag = re.sub(r'viewBox="[^"]*"', f'viewBox="{x} {y} {w} {h}"', open_tag)
    tag = re.sub(r'\swidth="[^"]*"', f' width="{w}"', tag)
    tag = re.sub(r'\sheight="[^"]*"', f' height="{h}"', tag)
    iw, ih = int(round(w)), int(round(h))
    board = (
        f'<div class="board" style="width:{w}px;height:{h}px;padding:0;margin:0;'
        f'background:{PAPER};display:block;overflow:hidden">{tag}{inner}</svg></div>'
    )
    out = WORK / "pages" / f"panel-{slug(rel)}-{int(x)}-{int(y)}-{iw}x{ih}.html"
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(text[:b0] + "\n" + board + "\n" + text[b1:], encoding="utf-8")
    return out, iw, ih


def save_optimized(img: Image.Image, out: Path) -> None:
    out.parent.mkdir(parents=True, exist_ok=True)
    img.save(out, optimize=True)


def render_whole_board(rel: str, out: Path, scale: float) -> None:
    w, h = page_size(rel)
    shot = WORK / "png" / f"{slug(rel)}.board.png"
    chrome_png(SRC / rel, shot, w, h, scale)
    save_optimized(Image.open(shot), out)


def crop_from_board(rel: str, box: dict, out: Path, scale: float) -> None:
    w, h = page_size(rel)
    shot = WORK / "png" / f"{slug(rel)}.board.png"
    chrome_png(SRC / rel, shot, w, h, scale)
    img = Image.open(shot)
    save_optimized(
        img.crop(
            (
                int(round(box["l"] * scale)),
                int(round(box["t"] * scale)),
                int(round(box["r"] * scale)),
                int(round(box["b"] * scale)),
            )
        ),
        out,
    )


def render_svg_panel(
    rel: str, region: tuple[float, float, float, float], out: Path, scale: float
) -> None:
    page, iw, ih = svg_panel_page(rel, region)
    # Chrome clips the paint of a very small headless window (a 335x138 window
    # rendered only the top 50 CSS px of the panel). Render into a comfortable
    # window and crop back to the panel box, which is anchored at 0,0.
    ww, wh = max(iw, 800), max(ih, 600)
    shot = WORK / "png" / f"{out.stem}.raw.png"
    chrome_png(page, shot, ww, wh, scale)
    img = Image.open(shot)
    img = img.crop(
        (0, 0, min(img.width, int(iw * scale)), min(img.height, int(ih * scale)))
    )
    save_optimized(img, out)


def fit(render: Callable[[float], None], out: Path) -> None:
    """Render at 2x, and re-export at 1.5x anything over 1.5 MB."""
    render(2)
    if out.stat().st_size > 1_500_000:
        render(1.5)
    print(shown(out))


def build_pngs(out_root: Path, chosen: set[str]) -> None:
    """Every image taken from a source among `chosen`."""
    img = out_root / "img"

    for rel, name in WHOLE_BOARDS:
        if rel in chosen:
            out = img / name
            fit(lambda s: render_whole_board(rel, out, s), out)

    for rel, region, name in SVG_PANELS:
        if rel in chosen:
            out = img / name
            reg = region if region else whole_svg_region(rel)
            fit(lambda s: render_svg_panel(rel, reg, out, s), out)

    for rel, selectors, name in HTML_CROPS:
        if rel in chosen:
            out = img / name
            box = html_rect(rel, selectors)
            fit(lambda s: crop_from_board(rel, box, out, s), out)


# ---------------------------------------------------------------------- main


def shown(path: Path) -> str:
    """`path` relative to the current directory when it lies beneath it."""
    try:
        return str(path.relative_to(Path.cwd()))
    except ValueError:
        return str(path)


def find_chrome(given: str | None) -> str:
    if given:
        path = shutil.which(given)
        if not path:
            raise SystemExit(f"--chrome {given}: not an executable program")
        return path
    for name in CHROME_NAMES:
        path = shutil.which(name)
        if path:
            return path
    if Path(MAC_CHROME).is_file():
        return MAC_CHROME
    raise SystemExit(
        "Chrome or Chromium not found. Install one (Debian: sudo apt install "
        "chromium, Ubuntu: sudo snap install chromium, macOS: brew install "
        "--cask google-chrome) or pass --chrome PATH."
    )


def chosen_sources(names: list[str]) -> set[str]:
    """The SOURCE arguments as paths relative to illustrations/."""
    known = all_sources()
    if not names:
        return known
    chosen = set()
    for name in names:
        path = Path(name).resolve()
        if not path.is_file():
            raise SystemExit(f"{name}: no such file")
        if not path.is_relative_to(SRC):
            raise SystemExit(f"{name}: not under illustrations/")
        rel = path.relative_to(SRC).as_posix()
        if rel not in known:
            raise SystemExit(f"{name}: not a source of any PDF or image")
        chosen.add(rel)
    return chosen


def main() -> None:
    global CHROME, WORK
    ap = argparse.ArgumentParser(
        description="Build the PDFs and article images from illustrations/."
    )
    ap.add_argument(
        "sources",
        nargs="*",
        metavar="SOURCE",
        help="rebuild only the outputs taken from these .html files "
        "(default: every output)",
    )
    ap.add_argument(
        "--only",
        choices=["pdf", "png", "all"],
        default="all",
        help="build only the PDFs or only the images (default: all)",
    )
    ap.add_argument(
        "--out",
        metavar="DIR",
        default=str(REPO / "docs"),
        help="output directory (default: docs/ of this repository)",
    )
    ap.add_argument(
        "--work",
        metavar="DIR",
        default=str(REPO / ".cache" / "illustrations"),
        help="directory for intermediates and the Chrome profile "
        "(default: .cache/illustrations in this repository)",
    )
    ap.add_argument(
        "--chrome", metavar="PATH", help="the Chrome or Chromium to run"
    )
    args = ap.parse_args()

    pdf = args.only in ("pdf", "all")
    png = args.only in ("png", "all")
    chosen = chosen_sources(args.sources)

    # A PDF is rebuilt whole, so every page of it is read, named or not, and
    # each must state its size before anything is rendered.
    needed: set[str] = set()
    if pdf:
        for pages in pdf_sets().values():
            if chosen.intersection(pages):
                needed.update(pages)
    if png:
        needed.update(chosen & image_sources())
    if not needed:
        print(
            "nothing to build: no requested output is taken from the named sources",
            file=sys.stderr,
        )
        return
    for rel in sorted(needed):
        page_size(rel)

    CHROME = find_chrome(args.chrome)
    missing = [t for t in ("pdfunite", "pdfinfo") if not shutil.which(t)]
    if pdf and missing:
        raise SystemExit(
            f"{' and '.join(missing)} not found. Install poppler (Debian or "
            "Ubuntu: sudo apt install poppler-utils, macOS: brew install poppler)."
        )
    if png and Image is None:
        raise SystemExit(
            "Pillow not found. Install it (Debian or Ubuntu: sudo apt install "
            "python3-pil, elsewhere: pip install Pillow in a virtual environment)."
        )

    WORK = Path(args.work).resolve()
    WORK.mkdir(parents=True, exist_ok=True)
    check_sources(min(needed))
    check_fonts()

    out_root = Path(args.out).resolve()
    if pdf:
        build_pdfs(out_root, chosen)
    if png:
        build_pngs(out_root, chosen)


if __name__ == "__main__":
    main()
