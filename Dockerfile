# P2P 60-24 OneClick Evo - Production Docker Image
# Multi-stage build per PM directive (Uwaga 5): minimize final image size.

FROM golang:1.21 AS builder
WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X github.com/60-24/p2p-60-24/cmd/p2p/commands.Version=1.0.0" -o p2p ./cmd/p2p

# Minimal runtime image (scratch, not alpine) per PM directive
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/p2p /usr/local/bin/p2p
COPY --from=builder /build/config/default.yaml /etc/p2p/default.yaml

EXPOSE 6024/udp 9090/tcp

ENTRYPOINT ["/usr/local/bin/p2p"]
CMD ["run", "--config", "/etc/p2p/default.yaml"]
