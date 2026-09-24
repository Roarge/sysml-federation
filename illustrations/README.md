# Illustrations

The architecture views, the two A3 sheets, the use-case storyboard and 28 of the
images in the articles are all built from the HTML files in this directory. Each
file is one board, a standalone page at its printed size, and
[`build.py`](build.py) prints it with headless Chrome as it stands. A correction
to any of those drawings is an edit to one of these files followed by a rebuild.

## Which file makes which output

| Output under `docs/` | Page | Source |
|---|---|---|
| [`architecture/architecture-views.pdf`](../docs/architecture/architecture-views.pdf) | 1 | [`architecture/overview.html`](architecture/overview.html) |
| | 2 | [`architecture/v1-context.html`](architecture/v1-context.html) |
| | 3 | [`architecture/v2-composition.html`](architecture/v2-composition.html) |
| | 4 | [`architecture/v3-runtime.html`](architecture/v3-runtime.html) |
| | 5 | [`architecture/v4-deployment.html`](architecture/v4-deployment.html) |
| | 6 | [`architecture/v5-adapter.html`](architecture/v5-adapter.html) |
| [`a3/L0-federating-a-systems-model.pdf`](../docs/a3/L0-federating-a-systems-model.pdf) | 1 | [`a3/L0-federating-a-systems-model.html`](a3/L0-federating-a-systems-model.html) |
| [`a3/L2b-pipeline-example-capacity-and-verdicts.pdf`](../docs/a3/L2b-pipeline-example-capacity-and-verdicts.pdf) | 1 | [`a3/L2b-pipeline-example-capacity-and-verdicts.html`](a3/L2b-pipeline-example-capacity-and-verdicts.html) |
| [`stories/use-cases.pdf`](../docs/stories/use-cases.pdf) | 1 | [`stories/overview.html`](stories/overview.html) |
| | 2 | [`stories/us01-launch.html`](stories/us01-launch.html) |
| | 3 | [`stories/us02-read-the-model.html`](stories/us02-read-the-model.html) |
| | 4 | [`stories/us03-raise-a-non-bottleneck.html`](stories/us03-raise-a-non-bottleneck.html) |
| | 5 | [`stories/us04-raise-the-bottleneck.html`](stories/us04-raise-the-bottleneck.html) |
| | 6 | [`stories/us05-tighten-the-limit.html`](stories/us05-tighten-the-limit.html) |
| | 7 | [`stories/us06-read-the-document.html`](stories/us06-read-the-document.html) |
| | 8 | [`stories/us07-reorder-and-nest.html`](stories/us07-reorder-and-nest.html) |
| | 9 | [`stories/us08-shape-the-document.html`](stories/us08-shape-the-document.html) |
| | 10 | [`stories/us09-change-from-the-document.html`](stories/us09-change-from-the-document.html) |
| | 11 | [`stories/us10-change-from-the-viewer.html`](stories/us10-change-from-the-viewer.html) |
| | 12 | [`stories/us11-query-the-graph.html`](stories/us11-query-the-graph.html) |
| | 13 | [`stories/us12-reset.html`](stories/us12-reset.html) |

Every file in `a3/` becomes a one-page PDF of the same name. The pages of the
other two PDFs, and their order, are listed in `build.py`.

The images are cut from the same boards. A whole board is a screenshot of the
page. A drawing is the board's first SVG, or a region of it, rendered on its
own. The rest are the box of one element of a use-case board, named by a CSS
selector. On a single use case that element is `div.sketch`, which holds the
sketch and its caption.

| Image under `docs/img/` | Source | Taken as |
|---|---|---|
| `architecture-five-views.png` | `architecture/overview.html` | whole board |
| `v1-context.png` | `architecture/v1-context.html` | drawing, region |
| `v2-composition.png` | `architecture/v2-composition.html` | drawing, whole |
| `v2-merged-requirement.png` | `architecture/v2-composition.html` | drawing, region |
| `v2-subgraph-schemas.png` | `architecture/v2-composition.html` | drawing, region |
| `v3-runtime.png` | `architecture/v3-runtime.html` | drawing, whole |
| `v3-nothing-moves.png` | `architecture/v3-runtime.html` | drawing, region |
| `v4-deployment.png` | `architecture/v4-deployment.html` | drawing, whole |
| `v4-container.png` | `architecture/v4-deployment.html` | drawing, region |
| `v4-image-layers.png` | `architecture/v4-deployment.html` | drawing, region |
| `v5-adapter.png` | `architecture/v5-adapter.html` | drawing, whole |
| `a3-l0-model-side.png` | `a3/L0-federating-a-systems-model.html` | whole board |
| `a3-l0-legend.png` | `a3/L0-federating-a-systems-model.html` | drawing, region |
| `a3-l2b-model-side.png` | `a3/L2b-pipeline-example-capacity-and-verdicts.html` | whole board |
| `a3-l2b-arithmetic.png` | `a3/L2b-pipeline-example-capacity-and-verdicts.html` | drawing, region |
| `a3-l2b-wiring-states.png` | `a3/L2b-pipeline-example-capacity-and-verdicts.html` | drawing, region |
| `a3-l2b-physical.png` | `a3/L2b-pipeline-example-capacity-and-verdicts.html` | drawing, region |
| `overview-sketch.png` | `stories/overview.html` | drawing, region |
| `stories-personas.png` | `stories/overview.html` | the element `div.personas` |
| `stories-journey.png` | `stories/overview.html` | the element `div.journey` |
| `us01-launch.png` | `stories/us01-launch.html` | the element `div.sketch` |
| `us03-nothing-moves.png` | `stories/us03-raise-a-non-bottleneck.html` | the element `div.sketch` |
| `us04-bottleneck-moves.png` | `stories/us04-raise-the-bottleneck.html` | the element `div.sketch` |
| `us06-document.png` | `stories/us06-read-the-document.html` | the element `div.sketch` |
| `us09-edit-from-document.png` | `stories/us09-change-from-the-document.html` | the element `div.sketch` |
| `us10-viewer-to-document.png` | `stories/us10-change-from-the-viewer.html` | the element `div.sketch` |
| `us11-query.png` | `stories/us11-query-the-graph.html` | the element `div.sketch` |
| `us12-reset.png` | `stories/us12-reset.html` | the element `div.sketch` |

Images are rendered at twice their size in CSS pixels, and one that would exceed
1.5 MB is rendered again at 1.5 times. The images in `docs/img` whose names
begin with `app-` are screenshots of the running apps and are not built here.

## Previewing a board

Open the file in Chrome or Chromium. The board is a block of fixed size at the
top left of the page, so the browser shows it at its true size, and printing it
to PDF from the browser gives a page of the published size, since each file
states its page size in an `@page` rule. The type comes from Google Fonts, so a
preview needs a network connection. Without one the browser substitutes other
faces and the text reflows.

## What to keep when editing

The build reads each board's page size from the `@page` rule in the head of its
file. A `.board` rule in the file's style gives the board the same width and
height, and the two must agree. Where a file has more than one `.board` rule the
last one sets the size, as on the L2b sheet, whose second style block overrides
a first rule of 1440 by 900 px. Content that grows taller than the page runs
onto a second page, and a build of the PDFs then stops with a message naming the
file. A build with `--only png` makes no such check and gives no warning. An
image of the whole board is cut off at the page edge. The image of an element
that reaches past the page keeps the element's full size, and everything beyond
the page edge comes out solid black. The sizes in use are 1440 by 560 px for
the architecture overview, 1440 by 900 px for the five views, 1440 by 800 px for
the use-case overview, 960 by 640 px for each use case and 1587 by 1123 px for
the A3 sheets, which is A3 at 96 px to the inch.

Inside the SVG drawings every colour is written as a literal `oklch()` value,
and an edit uses the same values exactly. Every file but the L0 sheet also
declares colours from the table below as custom properties at the top of its
style block (`--ink`, `--pencil` and so on), and the HTML text and legends of
the architecture and use-case boards mostly take their colours from those. Pale
fills, such as the tint of a router box, mix one of these colours into paper
with `color-mix()`.

| Name | Value | Use |
|---|---|---|
| paper | `oklch(98% 0.005 90)` | the page, and the fill of plain boxes |
| ink | `oklch(25% 0.01 260)` | lines and text, and on the A3 sheets a known value |
| pencil | `oklch(58% 0.01 260)` | secondary text and lines, such as scope lines, captions and notes, and on the L0 sheet the question mark of a value not yet measured |
| shade | `oklch(93% 0.006 90)` | a shaded fill for what the example owns rather than the adapter on the architecture views and the A3 sheets (on L0, also what an adopter replaces), for a model owner story and, with a dashed outline, a document owner story on the use-case overview and in the role tag at the top of each use case, and for a panel of a screen or terminal in the use-case sketches |
| red | `oklch(55% 0.19 25)` | a failing requirement or verdict and the bottleneck, and in the adapter view the refusal path |
| blue | `oklch(52% 0.12 250)` | the router and what passes through it |
| amber | `oklch(70% 0.14 80)` | an estimated value, on the A3 sheets |

The six architecture boards, the use-case overview and both A3 sheets each carry
a legend of their own, and an edit keeps to the legend of its board. The marks
differ from board to board. A dashed outline, for instance, stands for something
outside the demo on the architecture overview and for a document owner story on
the use-case overview. The twelve use cases have no legend, and the role tag at
the top of each is drawn with the marks of the overview's. The legend of the
architecture views is set out in [Five views and twenty-six decisions](../docs/articles/06-five-views-and-twenty-six-decisions.md),
and the layout, type and colour code of the A3 sheets in
[An A3 sheet for a fifteen-minute reader](../docs/articles/07-an-a3-sheet-for-a-fifteen-minute-reader.md).

Two faces are loaded from Google Fonts. Source Sans 3, at weights 400 and 600,
carries the running text, the legends, most labels in the architecture drawings
and everything on the A3 sheets. Patrick Hand, a handwriting face, carries the
titles and headings of the architecture and use-case boards, a few labels and
notes in the architecture drawings, and the sketches on the use-case boards.
Inside the sketches of the twelve use cases, Source Sans 3 sets most of what a
sketched screen or terminal shows, such as model text, commands and document
rows. A face other than these two is not loaded, so the browser would draw it in
whatever font the machine has.

No text on an A3 sheet is smaller than 19 px, the equivalent of 14 points when
the sheet prints at A3. Text that no longer fits is shortened rather than set
smaller.

The article images are cut at places fixed in `build.py`. The table
`SVG_PANELS` gives each region in the drawing's own coordinates, the numbers of
its `viewBox`, and the comments beside some regions say why an edge sits where
it does. Moving or enlarging something inside or near a region can cut it off
in the image, so after editing a board look at the images taken from it and
adjust the region when needed. The table `HTML_CROPS` takes the box around
named elements and follows them wherever the layout puts them. Those crops
depend on the class names `sketch`, `personas` and `journey`. The image of a
single use case is the box of `div.sketch`, and its caption is in the image
only because the caption sits inside that element, so a caption moved out of it
drops out of the image without a warning. The build stops when the selector of
an image matches nothing.

## Building

The build needs:

- Python 3.10 or later
- Chrome or Chromium
- `pdfunite` and `pdfinfo` from poppler, for the PDFs
- Pillow, for the images
- a network connection to `fonts.googleapis.com` and `fonts.gstatic.com`

On Debian or Ubuntu:

```
sudo apt install poppler-utils python3-pil
sudo apt install chromium        # Debian
sudo snap install chromium       # Ubuntu
```

On macOS with Homebrew:

```
brew install python poppler
brew install --cask google-chrome
python3 -m venv .venv
. .venv/bin/activate
pip install Pillow
```

The build looks for `google-chrome`, `google-chrome-stable`, `chromium` and
`chromium-browser` on the `PATH`, in that order, and then for Google Chrome in
the macOS Applications folder. From the root of a clone:

```
python3 illustrations/build.py
```

Before rendering, it reads the page size of every board it needs and checks that
the tools are installed. It then opens one of those boards in Chrome, to make
sure Chrome can read the files where they are, and loads the three font faces
(Source Sans 3 at 400 and 600, Patrick Hand at 400) on a small test page it
writes to the work directory. If Chrome cannot open either page, the build stops
and shows what Chrome reported. If the test page opens and a face fails to load,
it stops with a message that the fonts could not be fetched. Both checks come
before any output is written. Chrome that cannot read a file renders its own
error page in place of the board and still reports success, and Chrome that
cannot load a face falls back to another without a warning and reflows the
page. After that Chrome prints the boards and renders the images, pdfinfo
confirms that each board printed as one page, and pdfunite joins the pages of
the two multi-page PDFs. The build writes one line with the path of each file
as it goes, and any failure stops it with a message and a non-zero exit status.

The options:

- `SOURCE ...`, one or more of the HTML files, as paths from the current
  directory. Only what they feed is rebuilt: every PDF that contains one of
  them, whole, and every image taken from one of them. Every other file under
  the output directory keeps its bytes. A path outside `illustrations/`, or a
  file that feeds no output, is an error.
- `--only pdf` or `--only png` builds only the PDFs or only the images. The
  default is `--only all`.
- `--out DIR` writes the outputs under `DIR` instead of `docs/`, in the same
  layout.
- `--work DIR` holds the intermediate files and Chrome's profile. The default is
  `.cache/illustrations` inside the clone, a hidden directory that git does not
  track. Chrome has to read the pages the build writes there, and a Chromium
  installed as a snap package can read only some directories. It reads the
  home directory, but not the hidden directories at its top level, such as
  `~/.cache`. With that browser, keep the clone and any `--work` elsewhere in
  the home directory, or pass `--chrome` a browser that is not a snap.
- `--chrome PATH` names the browser to run.

A correction to the deployment view, for example, is rebuilt with

```
python3 illustrations/build.py illustrations/architecture/v4-deployment.html
```

which rewrites `docs/architecture/architecture-views.pdf` and the three images
taken from that board, `v4-deployment.png`, `v4-container.png` and
`v4-image-layers.png`.

`make illustrations` runs the full build. The Makefile looks for a Go toolchain
as it is read and stops without one, so the make target needs Go installed and
the Python command does not.

## Checking a change

A rebuilt PDF is never byte-identical to the one before it. Each A3 sheet
carries the time Chrome printed it, and pdfunite gives each of the two joined
PDFs a new file identifier on every run. Compare the pages as pictures instead,
from two builds on the same machine, one made before the edit and one after:

```
python3 illustrations/build.py illustrations/stories/us05-tighten-the-limit.html --out .cache/before
# edit illustrations/stories/us05-tighten-the-limit.html
python3 illustrations/build.py illustrations/stories/us05-tighten-the-limit.html --out .cache/after
pdftoppm -r 96 -png .cache/before/stories/use-cases.pdf .cache/before/use-cases
pdftoppm -r 96 -png .cache/after/stories/use-cases.pdf .cache/after/use-cases
```

`pdftoppm` comes with poppler and writes one PNG per page, numbered from 01 for
this PDF. A few lines of Pillow show whether two pages differ and where:

```python
from PIL import Image, ImageChops

a = Image.open(".cache/before/use-cases-06.png").convert("RGB")
b = Image.open(".cache/after/use-cases-06.png").convert("RGB")
print(ImageChops.difference(a, b).getbbox())  # None when the pages are identical
```

The images under `img/` of the two output directories compare the same way.
Only the pages and images the edit was meant to change should differ. Chrome of
another version, or on another system, can draw text a fraction of a pixel away
from where it sits in the committed files, which is why both builds come from
the same machine.

## Sending a correction

A correction goes in one pull request, with the edited source and the files
rebuilt from it in the same commit, so that the published PDF or image and its
source agree at every commit. Naming the edited source on the command line
keeps the pull request to the outputs that source feeds. The documentation site
is built from `docs/` alone, so a change reaches it only through the rebuilt
files.

Adding, renaming or removing an image under `docs/img` also means a change to
the model of the demo, because a unit test in `internal/trace` checks that every
image there is named by exactly one view and that every image a view names
exists. [The model's README](../model/README.md) describes the views.
