# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory to root of project
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy root go.mod and go.sum first (needed for replace directive)
COPY go.mod go.sum* ./

# Copy merchant-service go.mod
COPY services/merchant-service/go.mod services/merchant-service/go.sum* ./services/merchant-service/

# Download root dependencies
RUN go mod download

# Download merchant-service dependencies
WORKDIR /app/services/merchant-service
RUN go mod download

# Copy root internal packages (needed for replace directive)
WORKDIR /app
COPY internal ./internal

# Copy merchant-service source code
COPY services/merchant-service ./services/merchant-service

# Build the application
WORKDIR /app/services/merchant-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o merchant-service ./cmd/main.go

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
