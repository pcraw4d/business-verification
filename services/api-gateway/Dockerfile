# Build stage
FROM golang:1.22-alpine AS builder

# Force rebuild: Fixed CORS duplicate headers and API routing issues
# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files (go.sum is optional)
COPY go.mod ./

# Download dependencies (this will create go.sum if missing)
RUN go mod download

# Copy go.sum if it exists (optional)
COPY go.sum* ./

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway ./cmd/main.go

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
COPY --from=builder /app/api-gateway .

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
