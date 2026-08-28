FROM --platform=$BUILDPLATFORM golang:1.27 AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/sysml-federation ./cmd/sysml-federation

FROM ghcr.io/wundergraph/cosmo/router:0.343.1 AS router

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=router /router /router
ADD --checksum=sha256:c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4 https://raw.githubusercontent.com/wundergraph/cosmo/router@0.343.1/LICENSE /router.LICENSE
COPY --from=build /out/sysml-federation /sysml-federation
COPY examples/pipeline/config.json /app/config.json
COPY examples/pipeline/model.sysml /app/model.sysml
ENV DO_NOT_TRACK=1 COSMO_TELEMETRY_DISABLED=true TRACING_ENABLED=false METRICS_OTLP_ENABLED=false
EXPOSE 8080
HEALTHCHECK --interval=10s --start-period=15s CMD ["/sysml-federation", "healthcheck"]
ENTRYPOINT ["/sysml-federation"]
CMD ["serve"]
