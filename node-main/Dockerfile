FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG SINGBOX_VERSION=v1.10.7

RUN apk update && apk add --no-cache make curl tar

WORKDIR /src

COPY go* .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} make NAME=main build

# Fetch a pre-built sing-box binary (release tarballs from SagerNet).
RUN set -eux; \
    case "${TARGETARCH}" in \
        amd64) SB_ARCH="amd64" ;; \
        arm64) SB_ARCH="arm64" ;; \
        arm)   SB_ARCH="armv7" ;; \
        *)     SB_ARCH="${TARGETARCH}" ;; \
    esac; \
    VER="${SINGBOX_VERSION#v}"; \
    TARBALL="sing-box-${VER}-linux-${SB_ARCH}.tar.gz"; \
    URL="https://github.com/SagerNet/sing-box/releases/download/${SINGBOX_VERSION}/${TARBALL}"; \
    curl -fsSL -o /tmp/sb.tgz "${URL}"; \
    tar -xzf /tmp/sb.tgz -C /tmp; \
    install -m 0755 /tmp/sing-box-${VER}-linux-${SB_ARCH}/sing-box /usr/local/bin/sing-box

FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/PasarGuard/node"

RUN apk update && apk add --no-cache wireguard-tools nftables iproute2 procps ca-certificates

WORKDIR /app
COPY --from=builder /src/main /app/main
COPY --from=builder /usr/local/bin/sing-box /usr/local/bin/sing-box

ENV SINGBOX_EXECUTABLE_PATH=/usr/local/bin/sing-box

ENTRYPOINT ["./main"]
