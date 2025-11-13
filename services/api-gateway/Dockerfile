# Build stage
FROM golang:1.24-alpine AS builder

# Force rebuild: Fixed API routing for risk benchmarks/predictions endpoints (preserve /risk in path)
# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Disable Go workspace mode to avoid "module appears multiple times" error
ENV GOWORK=off

# Copy the entire repository structure (needed for module resolution)
# Railway builds from repository root, so we copy everything
COPY . .

# Debug: Verify what was copied
RUN echo "=== Debug: Checking repository structure ===" && \
    ls -la /app/ | head -20 && \
    echo "=== Checking services directory ===" && \
    ls -la /app/services/ 2>/dev/null || echo "services directory not found" && \
    echo "=== Checking api-gateway directory ===" && \
    ls -la /app/services/api-gateway/ 2>/dev/null || echo "api-gateway directory not found" && \
    echo "=== Checking cmd directory ===" && \
    ls -la /app/services/api-gateway/cmd/ 2>/dev/null || echo "cmd directory not found"

# Change to service directory where go.mod is located
WORKDIR /app/services/api-gateway

# Verify we're in the right place and files exist
RUN pwd && ls -la && ls -la cmd/ 2>/dev/null || echo "cmd directory not found in current directory"

# Download dependencies for the api-gateway service module
RUN go mod download

# Build the application from the service directory
# This ensures Go can find the go.mod file and resolve imports correctly
RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -a -installsuffix cgo -o api-gateway ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates and wget for HTTPS requests and health checks
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Create app directory
RUN mkdir -p /app
WORKDIR /app

# Copy binary from builder stage (built in /app/services/api-gateway)
COPY --from=builder /app/services/api-gateway/api-gateway .

# Change ownership to non-root user
RUN chown appuser:appuser /app/api-gateway

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the application
CMD ["./api-gateway"]
