# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Disable Go workspace mode to avoid "module appears multiple times" error
ENV GOWORK=off

# Copy the entire repository structure (needed for module resolution)
# Railway builds from repository root, so we copy everything
COPY . .

# Change to service directory where go.mod is located
WORKDIR /app/services/merchant-service

# Download dependencies for the merchant service module
RUN go mod download

# Build the application from the service directory
# This ensures Go can find the go.mod file and resolve imports correctly
RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -a -installsuffix cgo -o merchant-service ./cmd/main.go

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
COPY --from=builder /app/services/merchant-service/merchant-service .

# Change ownership to non-root user
RUN chown appuser:appuser /app/merchant-service

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Run the application
CMD ["./merchant-service"]
