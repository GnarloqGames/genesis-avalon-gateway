FROM registry.0x42.in/library/docker/genesis-avalon-builder:bookworm-1.1.0 as builder

WORKDIR /build
COPY . .
RUN go build -o ./bin/gatewayd ./cmd/gateway/...

FROM debian:bookworm
COPY --from=builder /build/bin/gatewayd /usr/bin/gatewayd