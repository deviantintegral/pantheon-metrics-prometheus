# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./
COPY testdata ./testdata

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pantheon-metrics-exporter .

# Stage 2: Create final image with Terminus CLI
FROM php:8.3-cli-alpine

# Install dependencies for Terminus
RUN apk add --no-cache \
    bash \
    git \
    openssh-client \
    ca-certificates \
    curl \
    composer

# Install Terminus CLI
RUN mkdir -p /usr/local/share/terminus && \
    cd /usr/local/share/terminus && \
    curl -L https://github.com/pantheon-systems/terminus/releases/download/3.5.1/terminus.phar -o terminus && \
    chmod +x terminus && \
    ln -s /usr/local/share/terminus/terminus /usr/local/bin/terminus

# Create non-root user
RUN addgroup -g 1000 exporter && \
    adduser -D -u 1000 -G exporter exporter

# Copy the binary from builder
COPY --from=builder /app/pantheon-metrics-exporter /usr/local/bin/pantheon-metrics-exporter

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
