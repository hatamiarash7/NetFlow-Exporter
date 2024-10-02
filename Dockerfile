FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23.2 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG APP_VERSION

WORKDIR /app/
ADD . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w -X 'main.version=${APP_VERSION}'" -o exporter main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch

ARG DATE_CREATED
ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION

LABEL org.opencontainers.image.created=$DATE_CREATED
LABEL org.opencontainers.version="$APP_VERSION"
LABEL org.opencontainers.image.authors="Arash Hatami <info@arash-hatami.ir>"
LABEL org.opencontainers.image.vendor="Arash Hatami"
LABEL org.opencontainers.image.title="netflow-exporter"
LABEL org.opencontainers.image.description="Prometheus exporter for NetFlow"
LABEL org.opencontainers.image.source="https://github.com/hatamiarash7/netflow-exporter"
LABEL org.opencontainers.image.url="https://github.com/hatamiarash7/netflow-exporter"
LABEL org.opencontainers.image.documentation="https://github.com/hatamiarash7/netflow-exporter"

WORKDIR /app/

COPY --from=builder /app/exporter /app/exporter

ENTRYPOINT ["/app/exporter"]
