# Multi-stage build for Enhanced Railway Server
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

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

# Build the enhanced railway-server
WORKDIR /app/cmd/railway-server
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o railway-server main.go

# Final stage with minimal footprint
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/cmd/railway-server/railway-server .

# Create non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./railway-server"]
