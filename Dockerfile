FROM docker.io/golang:1.24.3-alpine AS builder

RUN apk add --no-cache build-base xz git aom-dev && \
    wget -O /tmp/upx.tar.xz https://github.com/upx/upx/releases/download/v5.0.1/upx-5.0.1-amd64_linux.tar.xz && \
    tar -xJf /tmp/upx.tar.xz -C /tmp && \
    mv /tmp/upx-5.0.1-amd64_linux/upx /bin/upx && \
    chmod a+x /bin/upx && \
    rm -rf /tmp/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN set -x && \
    cd healthcheck-app && \
    make

ARG BUILD_VERSION

RUN CGO_ENABLED=1 GOOS=linux \
    go build -a -ldflags="-X main.buildVersion=$BUILD_VERSION -linkmode external -extldflags '-static' -s -w" \
    -o /app/robohash ./cmd/server

RUN upx --best --lzma /app/robohash

FROM scratch

WORKDIR /app

COPY --from=builder /src/healthcheck-app/healthcheck /bin/healthcheck

COPY --from=builder /app/robohash .

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["/bin/healthcheck", "http://localhost:8080/health"]

LABEL org.opencontainers.image.title="Robohash" \
      org.opencontainers.image.description="Robohash generator Golang implementation" \
      org.opencontainers.image.authors="terem42" \
      org.opencontainers.image.url="https://github.com/terem42/robohash" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="terem42"

ENTRYPOINT ["/app/robohash"]