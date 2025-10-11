# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies including build tools for CGO
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for ONNX Runtime
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags='-w -s' \
    -o risk-assessment-service \
    ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates, timezone data, and runtime libraries for CGO
RUN apk --no-cache add ca-certificates tzdata libc6-compat

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Create app directory and logs directory
RUN mkdir -p /app/logs
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/risk-assessment-service .

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (Railway will set PORT environment variable)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the application
CMD ["./risk-assessment-service"]
