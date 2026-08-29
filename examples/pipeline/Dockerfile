FROM --platform=$BUILDPLATFORM golang:1.27 AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOTOOLCHAIN=local GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/sysml-federation ./cmd/sysml-federation

# Tag and digest together: the tag says which release this is, the digest is
# what actually gets pulled. The digest is the multi-platform index, so both
# architectures still resolve, and it cannot be re-pointed under the licence
# the next line certifies.
FROM ghcr.io/wundergraph/cosmo/router:0.343.1@sha256:cdbfd46b79b4e3fb309785f90162b8261dce75dc1b238acf7f9416351ad0fcf3 AS router

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=router /router /router
ADD --chmod=644 --checksum=sha256:c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4 https://raw.githubusercontent.com/wundergraph/cosmo/router@0.343.1/LICENSE /router.LICENSE
COPY --from=build /out/sysml-federation /sysml-federation
COPY examples/pipeline/config.json /app/config.json
COPY examples/pipeline/model.sysml /app/model.sysml
# Inert on the normal path, where the supervisor builds the router child's
# whole environment itself. They are here for anyone who runs /router
# directly out of the image, and they match what the supervisor sets.
ENV DO_NOT_TRACK=1 COSMO_TELEMETRY_DISABLED=true TRACING_ENABLED=false METRICS_OTLP_ENABLED=false PROMETHEUS_ENABLED=false
EXPOSE 8080
HEALTHCHECK --interval=10s --start-period=15s CMD ["/sysml-federation", "healthcheck"]
ENTRYPOINT ["/sysml-federation"]
CMD ["serve"]
