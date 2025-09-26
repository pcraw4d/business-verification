# Multi-stage build for production
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go workspace file
COPY go.work ./

# Copy go mod files for all modules
COPY go.mod go.sum ./
COPY cmd/railway-server/go.mod cmd/railway-server/go.sum ./cmd/railway-server/
COPY pkg/*/go.mod pkg/*/go.sum ./pkg/*/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application from the railway-server directory
WORKDIR /app/cmd/railway-server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-platform main.go

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/cmd/railway-server/kyb-platform .

# Copy configuration files
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /app/monitoring ./monitoring
COPY --from=builder /app/web ./web

# Create necessary directories
RUN mkdir -p logs backups build && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./kyb-platform"]
