# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Disable Go workspace mode to avoid "module appears multiple times" error
# This is necessary because we're using a replace directive to reference the root module
ENV GOWORK=off

# Copy root go.mod and go.sum to establish module context
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy pkg/ directory (needed for kyb-platform/pkg/errors)
RUN mkdir -p ./pkg
COPY pkg ./pkg

# Copy service-specific source code
COPY services/merchant-service/cmd/ ./services/merchant-service/cmd/
COPY services/merchant-service/internal/ ./services/merchant-service/internal/
COPY services/merchant-service/go.mod ./services/merchant-service/go.mod

# Build the application
# Explicitly disable workspace mode in build command
RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -a -installsuffix cgo -o merchant-service ./services/merchant-service/cmd/main.go

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
COPY --from=builder /app/merchant-service .

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
