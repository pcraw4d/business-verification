# Simple, reliable Dockerfile for KYB Platform Enhanced v4.0.0
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go workspace files
COPY go.work go.mod go.sum ./

# Copy all go.mod files
COPY cmd/railway-server/go.mod cmd/railway-server/go.sum ./cmd/railway-server/
COPY pkg/cache/go.mod pkg/cache/go.sum ./pkg/cache/
COPY pkg/performance/go.mod pkg/performance/go.sum ./pkg/performance/
COPY pkg/monitoring/go.mod pkg/monitoring/go.sum ./pkg/monitoring/
COPY pkg/security/go.mod pkg/security/go.sum ./pkg/security/
COPY pkg/analytics/go.mod pkg/analytics/go.sum ./pkg/analytics/
COPY pkg/api/go.mod pkg/api/go.sum ./pkg/api/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
WORKDIR /app/cmd/railway-server
RUN go build -o railway-server main.go

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy binary
COPY --from=builder /app/cmd/railway-server/railway-server .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./railway-server"]