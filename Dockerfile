# Stage 1: Build with GoReleaser
FROM golang:1.24-alpine AS builder

# Install git (required by GoReleaser) and goreleaser
RUN apk add --no-cache git curl && \
    curl -sfL https://goreleaser.com/static/run | sh -s -- --version && \
    curl -sfL https://goreleaser.com/static/run > /usr/local/bin/goreleaser && \
    chmod +x /usr/local/bin/goreleaser

WORKDIR /app

# Copy all source files
COPY . .

# Build using goreleaser snapshot for current architecture
RUN goreleaser build --snapshot --clean --single-target

# Stage 2: Create minimal runtime image
FROM alpine:3.23

# Install ca-certificates for HTTPS and wget for healthcheck
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN addgroup -g 1000 exporter && \
    adduser -D -u 1000 -G exporter exporter

# Copy the binary from builder
# GoReleaser places binaries in dist/<build-id>/
COPY --from=builder /app/dist/pantheon-metrics-exporter_linux_*/pantheon-metrics-exporter /usr/local/bin/pantheon-metrics-exporter

# Switch to non-root user
USER exporter

WORKDIR /home/exporter

# Expose the metrics port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the exporter
ENTRYPOINT ["/usr/local/bin/pantheon-metrics-exporter"]
