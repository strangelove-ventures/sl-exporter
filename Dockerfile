# Build the exporter binary
FROM golang:1.23 as builder

ARG VERSION=0.0.0
ARG REVISION=abcdef

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY *.go .
COPY cmd/ cmd/
COPY cosmos/ cosmos/
COPY metrics/ metrics/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "\
    -X github.com/prometheus/common/version.Version=${VERSION} \
    -X github.com/prometheus/common/version.Revision=${REVISION} \
    " -a -o exporter .

# Use distroless as minimal base image to package the exporter binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.source=https://github.com/strangelove-ventures/sl-exporter

WORKDIR /
COPY --from=builder /workspace/exporter .
USER 65532:65532

ENTRYPOINT ["/exporter"]
