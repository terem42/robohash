FROM docker.io/golang:1.24.3-alpine AS builder

RUN apk add --no-cache build-base xz git aom-dev && \
    wget -O /tmp/upx.tar.xz https://github.com/upx/upx/releases/download/v5.0.0/upx-5.0.0-amd64_linux.tar.xz && \
    tar -xJf /tmp/upx.tar.xz -C /tmp && \
    mv /tmp/upx-5.0.0-amd64_linux/upx /bin/upx && \
    chmod a+x /bin/upx && \
    rm -rf /tmp/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILD_VERSION

RUN CGO_ENABLED=1 GOOS=linux \
    go build -a -ldflags="-X main.buildVersion=$BUILD_VERSION -linkmode external -extldflags '-static' -s -w" \
    -o /app/robohash ./cmd/server

RUN upx --best --lzma /app/robohash

FROM scratch

WORKDIR /app

COPY --from=builder /app/robohash .
COPY assets ./assets/

EXPOSE 8080

LABEL org.opencontainers.image.source=https://github.com/terem42/robohash

ENTRYPOINT ["/app/robohash"]