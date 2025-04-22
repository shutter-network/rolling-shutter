FROM golang:1.23.6 AS builder
ENV GOMODCACHE=/root/.cache/mod
# Fetch go modules separately to improve cache usage
RUN mkdir /gomod
COPY /rolling-shutter/go.* /gomod/
WORKDIR /gomod
RUN --mount=type=cache,id=go-mod-$TARGETARCH$TARGETVARIANT,sharing=locked,target=/root/.cache go mod download

# Build binary
COPY / /src
WORKDIR /src/rolling-shutter
RUN --mount=type=cache,id=go-mod-$TARGETARCH$TARGETVARIANT,sharing=locked,target=/root/.cache CGO_ENABLED=1 GOOS=linux GOFLAGS=-v make build

FROM golang:1.21 AS runner

COPY --from=builder /src/rolling-shutter/bin/rolling-shutter /rolling-shutter

# Use 'uclibc' flavor to avoid https://github.com/docker-library/busybox/issues/155#issuecomment-1344375664
RUN --mount=from=busybox:uclibc,src=/bin,dst=/bin mkdir -p /etc/ssl
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/rolling-shutter"]
