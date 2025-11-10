# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies including wget for health checks
RUN apk add --no-cache git ca-certificates wget

# Copy root go.mod and go.sum to establish module context
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the entire project structure (needed for internal packages)
# This assumes Railway is building from the repository root
# If Railway builds from services/classification-service, we need to configure
# the root directory in Railway dashboard to be the repo root
COPY internal/ ./internal/
COPY pkg/ ./pkg/ 2>/dev/null || true

# Copy service-specific source code
COPY services/classification-service/cmd/ ./services/classification-service/cmd/
COPY services/classification-service/internal/ ./services/classification-service/internal/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o classification-service ./services/classification-service/cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates and wget for HTTPS requests and health checks
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Create app directory
RUN mkdir -p /app
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/classification-service .

# Change ownership to non-root user
RUN chown appuser:appuser /app/classification-service

# Switch to non-root user
USER appuser

# Expose port (Railway will set PORT environment variable)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the application
CMD ["./classification-service"]
