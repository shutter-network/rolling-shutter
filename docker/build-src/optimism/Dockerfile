FROM golang:1.21 as builder
ENV GOMODCACHE=/root/.cache/mod

# allow private repos
ENV GOPRIVATE="github.com/shutter-network/*"
ENV GONOPROXY=localhost
ARG GITHUB_ACCESS_TOKEN
RUN git config --global url."https://$GITHUB_ACCESS_TOKEN@github.com/".insteadOf "https://github.com/"

# Fetch go modules separately to improve cache usage
RUN mkdir /gomod
COPY /rolling-shutter/go.* /gomod/
WORKDIR /gomod
RUN --mount=type=cache,target=/root/.cache go mod download

# Build binary
COPY / /src
WORKDIR /src/rolling-shutter


RUN go env
RUN --mount=type=cache,target=/root/.cache CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-v  make build

FROM scratch as runner

COPY --from=builder /src/rolling-shutter/bin/rolling-shutter /rolling-shutter

# Use 'uclibc' flavor to avoid https://github.com/docker-library/busybox/issues/155#issuecomment-1344375664
RUN --mount=from=busybox:uclibc,src=/bin,dst=/bin mkdir -p /etc/ssl
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/rolling-shutter"]
