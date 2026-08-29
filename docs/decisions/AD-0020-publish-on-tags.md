# AD-0020 The image published by a tag-triggered workflow to GHCR

Status: accepted, amended once the publishing job was built. Date: 2026-08-27.

Amendment, 2026-08-29: the decision as first accepted put `latest` on the
image in the same push as the version tag and measured the result afterwards,
so an image over the size budget would already be what an untagged pull
returns before the gate could reject it. The version tag now goes up alone and
`latest` is moved only after both platforms have been measured and passed.

## Context

The launch shape is fixed: one image, one command, `docker run --rm
-p 8080:8080 ghcr.io/roarge/sysml-federation`, on Linux, macOS and Windows
with nothing but Docker installed. The README says nothing of how a visitor
obtains the demo, and the brief says where the image lives and not how it
gets there or who builds it. The repository's CI policy said that it "runs
the unit tests and nothing else", and `test.yml` runs on pull requests and
pushes to `main` with `permissions: contents: read`.

The registry has rules of its own. A package first published under a
personal account is private by default even from a public repository, is
made public once by hand in the package settings, and cannot be made
private again. Publishing from a workflow needs `contents: read`
and `packages: write`, a login to `ghcr.io` with the workflow token, and
the repository that contains the workflow is linked to the package
automatically, which gives the package the repository's access permissions
though not its visibility. `docker/metadata-action` writes the OCI
labels, turns `refs/tags/v1.2.3` into the tag `1.2.3` and adds `latest`
for semver tags.

Every process in the demo is Go, and the one foreign binary, the Cosmo
router, is already published for linux/amd64 and linux/arm64. Docker's own
page says emulation with QEMU "can be much slower than native builds,
especially for compute-heavy tasks like compilation" and gives the
cross-compilation form with `FROM --platform=$BUILDPLATFORM`, `TARGETOS`
and `TARGETARCH`. The base is `gcr.io/distroless/static-debian13`
at about 2 MiB, the same base the router uses. `COPY --from`
accepts an image reference, the router is statically linked and released
under Apache 2.0 with its LICENSE attached, and every router bump means a
rebuild. `docker/build-push-action` adds provenance attestations by
default, and with them present the GHCR package page shows an
unknown/unknown platform row. Whether `COPY --from` resolves the
target platform's variant during a multi-platform build was not confirmed
in the reference text.

## Decision

We will publish the image from a second GitHub Actions workflow,
`.github/workflows/publish.yml`, that runs on `push: tags: ['v*']` with
`permissions: contents: read, packages: write`, logs in to GHCR with the
workflow token, builds for linux/amd64 and linux/arm64 by Go
cross-compilation with `CGO_ENABLED=0` and no QEMU, tags with
`type=semver,pattern={{version}}` alone with the metadata action's `latest`
flavour turned off, sets `provenance: false`
and `sbom: false`, and pushes that one tag. It then reads the manifest back
and fails the job if either platform is missing or exceeds the size budget,
and only once both have passed does it put `latest` on the digest it
measured. A version with a pre-release suffix stops there and leaves `latest`
alone, so a release candidate is published under its own tag and is never what
an untagged pull returns. The package
is made public once by hand. `test.yml` keeps its read-only permissions
and runs the unit tests only, and the CI policy line becomes
"unit tests, plus image publishing on tags" (SC-07).

## Alternatives considered

A manual push from the maintainer's machine, the one alternative considered.
It would have left the CI policy line
untouched. It lost because a manual push has neither the automatic link
between package and repository that a workflow publish gives nor the
tag alignment the metadata action supplies.

Multi-platform build by QEMU emulation, the form Docker's own GitHub
Actions example uses. It lost because
emulation is slow for compilation and unnecessary when every stage is Go
and the foreign binary is already published for both architectures. Native
arm64 runners, free for public repositories, are named by the research
only for the case where a native build is ever wanted.

Keeping provenance and SBOM attestations. They give SLSA
provenance, which some readers regard as good hygiene for a public image.
They lost because the package page would show a third platform called
unknown/unknown, which reads as a broken build unless
the README explains it, and the design removes the row rather than
explaining it.

Publishing on pushes to `main` as well, with `sha` and `edge` tags, which
the packaging research recommends beside the tag trigger. The design takes
the tag trigger only, and SC-07 names two jobs and nothing else.

Moving `latest` in the same push as the version tag, which is what the
metadata action does for a semver tag unless it is told otherwise, and which
this record first accepted. Registry-counted sizes cannot be had without
pushing, so the gate can only run after the push, and with `latest` already
moved an over-budget image is what an untagged pull returns until somebody
notices. Turning `latest` off in the metadata step and putting it on the
measured digest afterwards costs one more step and one more registry call.

## Consequences

The launch line works for nobody until the package has been flipped to
public by hand, and the flip cannot be undone. SR-01's demonstration
from a host not authenticated to GHCR is what proves the flip was made.

Two workflow files with different permissions. `test.yml` stays read-only
and `publish.yml` alone holds `packages: write`, and the allowlist tracks
it (SC-04).

The workflow is also the test bench for SR-05 and SR-06. After the push it
reads the manifest back and fails if either platform is absent or if any
platform's compressed layers exceed 80 MB, and `latest` waits on that result.
The router alone is about 40 MB
compressed and sets the floor, which leaves about 40 MB for the demo.

One spike stands behind this record. Before the first tag is pushed,
the router binary inside the built arm64 layer is inspected to confirm
that `COPY --from` picked the arm64 variant. If it did not, the arm64
image would carry an amd64 router and fail on Apple Silicon. The outcome is
in [Five spikes before the first line](../articles/09-five-spikes-before-the-first-line.md).

Provenance is off, so the image ships without SLSA attestations, and the
package page lists exactly two platforms. Every router bump is a
rebuild of the image and a new tag, because the binary is copied rather than
pulled at runtime. The root `NOTICE` names the router at a pinned
version that moves with each bump (SR-08). Its licence text is fetched at
build time at the pinned tag and sits beside `/router` (SR-07).

`latest` is what the launch line pulls, since it names no tag, so what `latest`
points at decides what a stranger gets (SR-01). It follows the newest release
that passes the gate, so it can drift from what the README describes if the
README is not updated in the same tagged commit, a cost the packaging
research names and the design accepts. It can also
lag the newest tag in the registry instead of tracking it. A failed gate
leaves the version tag published and `latest` where it was, and a pre-release
moves nothing, so an untagged pull returns the last full release that was
measured and passed rather than whatever went up most recently. Cutting a
release means reading the run rather than assuming that a tag in the registry
is what `latest` names.

## Requirements affected

SR-01, SR-05, SR-06, SR-08, SC-07

## Sources

GitHub's documentation on package visibility for a package first published under a personal account, and on the permissions a publishing workflow needs. Docker's pages on multi-platform builds, QEMU emulation and cross-compilation, `docker/metadata-action` and `docker/build-push-action`. The distroless static base image. [The demo as it shipped](../articles/10-the-demo-being-built.md) for the Dockerfile and the publish job as they ship.
