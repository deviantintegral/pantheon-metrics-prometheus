# Stage 1: Build with GoReleaser
FROM golang:1.26-bookworm AS builder

# Install git (required by GoReleaser) and goreleaser
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    ca-certificates && \
    curl -sfL https://goreleaser.com/static/run | sh -s -- --version && \
    curl -sfL https://goreleaser.com/static/run > /usr/local/bin/goreleaser && \
    chmod +x /usr/local/bin/goreleaser && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy all source files
COPY . .

# Build using goreleaser snapshot for current architecture
RUN goreleaser build --snapshot --clean --single-target

# Stage 2: Create minimal runtime image
FROM debian:13-slim

# Install ca-certificates for HTTPS and wget for healthcheck
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    wget && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1000 exporter && \
    useradd -u 1000 -g exporter -s /bin/false -m exporter

# Copy the binary from builder
# GoReleaser places binaries in dist/<build-id>/
COPY --from=builder /app/dist/default_linux*/pantheon-metrics-exporter /usr/local/bin/pantheon-metrics-exporter

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
