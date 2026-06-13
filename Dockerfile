# ─── Stage 1: Builder ────────────────────────────────────────────────────────
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the full source tree
COPY . .

# Build the binary: static, stripped, no CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /build/server ./cmd/server/main.go

# ─── Stage 2: Runtime ────────────────────────────────────────────────────────
FROM scratch

# Import TLS certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled binary
COPY --from=builder /build/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]