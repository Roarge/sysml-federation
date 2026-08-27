# AD-0020 The image published by a tag-triggered workflow to GHCR

Status: accepted. Date: 2026-08-27.

## Context

The launch shape is fixed by D5: one image, one command, `docker run --rm
-p 8080:8080 ghcr.io/roarge/sysml-federation`, on Linux, macOS and Windows
with nothing but Docker installed. The README says nothing of how a visitor
obtains the demo, and D5 says where the image lives and not how it gets
there or who builds it. The repository's CI policy said that it "runs the
unit tests and nothing else", and `test.yml` runs on pull requests and
pushes to `main` with `permissions: contents: read` (C-74).

The registry has rules of its own. A package first published under a
personal account is private by default even from a public repository, is
made public once by hand in the package settings, and cannot be made
private again (C-46). Publishing from a workflow needs `contents: read`
and `packages: write`, a login to `ghcr.io` with the workflow token, and
the repository that contains the workflow is linked to the package
automatically, which gives the package the repository's access permissions
though not its visibility (C-47). `docker/metadata-action` writes the OCI
labels, turns `refs/tags/v1.2.3` into the tag `1.2.3` and adds `latest`
for semver tags (C-48).

Every process in the demo is Go, and the one foreign binary, the Cosmo
router, is already published for linux/amd64 and linux/arm64. Docker's own
page says emulation with QEMU "can be much slower than native builds,
especially for compute-heavy tasks like compilation" and gives the
cross-compilation form with `FROM --platform=$BUILDPLATFORM`, `TARGETOS`
and `TARGETARCH` (C-49). The base is `gcr.io/distroless/static-debian13`
at about 2 MiB, the same base the router uses (C-50). `COPY --from`
accepts an image reference, the router is statically linked and released
under Apache 2.0 with its LICENSE attached, and every router bump means a
rebuild (C-51). `docker/build-push-action` adds provenance attestations by
default, and with them present the GHCR package page shows an
unknown/unknown platform row (C-52). Whether `COPY --from` resolves the
target platform's variant during a multi-platform build was not confirmed
in the reference text (C-53).

The choice was put to the owner during planning and recorded as D10.

## Decision

We will publish the image from a second GitHub Actions workflow,
`.github/workflows/publish.yml`, that runs on `push: tags: ['v*']` with
`permissions: contents: read, packages: write`, logs in to GHCR with the
workflow token, builds for linux/amd64 and linux/arm64 by Go
cross-compilation with `CGO_ENABLED=0` and no QEMU, tags with
`type=semver,pattern={{version}}` and `latest`, sets `provenance: false`
and `sbom: false`, pushes, and then reads the manifest back and fails the
job if either platform is missing or exceeds the size budget. The package
is made public once by hand. `test.yml` keeps its read-only permissions
and runs the unit tests only, and the CI policy line becomes
"unit tests, plus image publishing on tags" (C-90, SC-07).

## Alternatives considered

A manual push from the maintainer's machine, the one alternative the
engineering log records for D10. It would have left the CI policy line
untouched. It lost because a manual push has neither the automatic link
between package and repository that a workflow publish gives (C-47) nor the
tag alignment of C-48.

Multi-platform build by QEMU emulation, the form Docker's own GitHub
Actions example uses (C-49, research on packaging). It lost because
emulation is slow for compilation and unnecessary when every stage is Go
and the foreign binary is already published for both architectures. Native
arm64 runners, free for public repositories, are named by the research
only for the case where a native build is ever wanted.

Keeping provenance and SBOM attestations (C-52). They give SLSA
provenance, which some readers regard as good hygiene for a public image.
They lost because the package page would show a third platform called
unknown/unknown, which the research says reads as a broken build unless
the README explains it, and the design removes the row rather than
explaining it.

Publishing on pushes to `main` as well, with `sha` and `edge` tags, which
the packaging research recommends beside the tag trigger. The design takes
the tag trigger only, and SC-07 names two jobs and nothing else.

## Consequences

The one-liner works for nobody until the package has been flipped to
public by hand, and the flip cannot be undone (C-46). SR-01's demonstration
from a host not authenticated to GHCR is what proves the flip was made.

Two workflow files with different permissions. `test.yml` stays read-only
and `publish.yml` alone holds `packages: write`. The CI policy line is a
local convention, so the second workflow file is the public evidence of the
change (C-90). The allowlist tracks it (SC-04).

The workflow is also the test bench for SR-05 and SR-06. After the push it
reads the manifest back and fails if either platform is absent or if any
platform's compressed layers exceed 80 MB. The router alone is about 40 MB
compressed and sets the floor, which leaves about 40 MB for the demo.

C-53 is the spike this record depends on. Before the first tag is pushed,
the router binary inside the built arm64 layer is inspected to confirm
that `COPY --from` picked the arm64 variant. If it did not, the arm64
image would carry an amd64 router and fail on Apple Silicon.

Provenance is off, so the image ships without SLSA attestations, and the
package page lists exactly two platforms (C-52). Every router bump is a
rebuild of the image and a new tag, because the binary is copied rather than
pulled at runtime (C-51). The root `NOTICE` names the router at a pinned
version that moves with each bump (SR-08). Its licence text is fetched at
build time at the pinned tag and sits beside `/router` (SR-07).

`latest` follows the newest tag, so it can drift from what the README
describes if the README is not updated in the same tagged commit, a cost
the packaging research names and the design accepts.

## Requirements affected

SR-05, SR-06, SR-08, SC-07

## Sources

The design brief D5, D10. The constraints list C-46 to C-53, C-74, C-90. The engineering log, design phase planning entry, D10. The architecture description V4 Publishing and The image. The requirements list SR-01, SR-05, SR-06, SR-07, SR-08, SC-04, SC-07. The research notes on packaging, options. The design-phase plan, D10.
