# Multi-stage build for Next.js + Go
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install frontend dependencies
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build Next.js app
RUN npm run build

# Go build stage
FROM golang:1.22-alpine AS go-builder

WORKDIR /app

# Install git, ca-certificates, and wget
RUN apk add --no-cache git ca-certificates wget

# Copy go.mod first for better caching
COPY go.mod ./

# Initialize go.sum if it doesn't exist and download dependencies
RUN go mod tidy && go mod download

# Copy go.sum if it exists
COPY go.sum* ./

# Copy source code
COPY . .

# Copy Next.js build from frontend-builder
COPY --from=frontend-builder /app/frontend/.next ./frontend/.next
COPY --from=frontend-builder /app/frontend/public ./frontend/public

# Verify static directory exists before build
RUN test -d static || (echo "ERROR: static directory not found" && exit 1)

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o frontend-service cmd/frontend-service/main.go

# Verify binary was created
RUN test -f frontend-service || (echo "ERROR: frontend-service binary not created" && exit 1)

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates and wget for health check
RUN apk add --no-cache ca-certificates wget

# Copy binary from builder
COPY --from=go-builder /app/frontend-service .

# Copy static files and Next.js build
COPY --from=go-builder /app/cmd/frontend-service/static ./static
COPY --from=frontend-builder /app/frontend/.next ./static/.next
COPY --from=frontend-builder /app/frontend/public ./static

# Expose port
EXPOSE 8086

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8086}/health || exit 1

# Run the application
CMD ["./frontend-service"]
