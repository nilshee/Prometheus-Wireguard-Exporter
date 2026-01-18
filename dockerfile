# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy only dependency files first for better layer caching
COPY src/go.mod src/go.sum ./

# Download dependencies (cached if go.mod/go.sum haven't changed)
RUN go mod download

# Copy source code
COPY src/ ./

# Build static binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o wireguard_exporter \
    ./cmd

# Runtime stage - use minimal base image
FROM alpine:3.21

# Install only WireGuard tools (much smaller than full Ubuntu)
RUN apk add --no-cache wireguard-tools

WORKDIR /opt

# Copy binary from builder
COPY --from=builder /app/wireguard_exporter .

# Create non-root user for security
RUN addgroup -S wireguard && adduser -S -G wireguard wireguard && \
    chown wireguard:wireguard /opt/wireguard_exporter

USER wireguard

# Expose default port
EXPOSE 9011

# Run the exporter
ENTRYPOINT ["/opt/wireguard_exporter"]
