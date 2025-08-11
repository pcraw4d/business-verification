# KYB Platform - Multi-Stage Dockerfile
# Supports development, testing, and production environments

# Base stage for shared dependencies
FROM golang:1.22-alpine AS base

# Install common dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Development stage
FROM base AS development

# Install development tools
RUN apk add --no-cache \
    gcc \
    musl-dev \
    && go install github.com/cosmtrek/air@latest

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Run with hot reload
CMD ["air", "-c", ".air.toml"]

# Testing stage
FROM base AS testing

# Install testing dependencies
RUN apk add --no-cache \
    gcc \
    musl-dev

# Copy source code
COPY . .

# Run tests
RUN go test -v -race -coverprofile=coverage.out ./...

# Build stage for production
FROM base AS builder

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty) -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
    -o /app/kyb-platform \
    ./cmd/api

# Build stage for debugging
FROM base AS builder-debug

# Copy source code
COPY . .

# Build with debug information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X main.version=$(git describe --tags --always --dirty) -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
    -gcflags="all=-N -l" \
    -o /app/kyb-platform-debug \
    ./cmd/api

# Production stage
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates \
    && addgroup -g 1001 -S kyb \
    && adduser -u 1001 -S kyb -G kyb

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/kyb-platform .

# Copy configuration files
COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/docs/api/openapi.yaml ./docs/api/

# Create necessary directories
RUN mkdir -p /app/logs /app/data /app/tmp \
    && chown -R kyb:kyb /app

# Switch to non-root user
USER kyb

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release \
    GO_ENV=production \
    PORT=8080

# Run the application
CMD ["./kyb-platform"]

# Debug stage
FROM alpine:3.19 AS debug

# Install runtime dependencies and debugging tools
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    gdb \
    && update-ca-certificates \
    && addgroup -g 1001 -S kyb \
    && adduser -u 1001 -S kyb -G kyb

# Set working directory
WORKDIR /app

# Copy debug binary from builder-debug stage
COPY --from=builder-debug /app/kyb-platform-debug ./kyb-platform

# Copy configuration files
COPY --from=builder-debug /app/configs/ ./configs/
COPY --from=builder-debug /app/docs/api/openapi.yaml ./docs/api/

# Create necessary directories
RUN mkdir -p /app/logs /app/data /app/tmp \
    && chown -R kyb:kyb /app

# Switch to non-root user
USER kyb

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables for debug
ENV GIN_MODE=debug \
    GO_ENV=debug \
    PORT=8080

# Run the application
CMD ["./kyb-platform"]

# Minimal production stage (smallest image)
FROM scratch AS production-minimal

# Copy binary from builder stage
COPY --from=builder /app/kyb-platform .

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release \
    GO_ENV=production \
    PORT=8080

# Run the application
CMD ["./kyb-platform"]

# Security scanning stage
FROM production AS security-scan

# Install security scanning tools
RUN apk add --no-cache \
    clamav \
    clamav-libunrar \
    && freshclam

# Switch back to root for scanning
USER root

# Scan the application
RUN clamscan --recursive --infected /app || true

# Switch back to kyb user
USER kyb

# Run the application
CMD ["./kyb-platform"]
