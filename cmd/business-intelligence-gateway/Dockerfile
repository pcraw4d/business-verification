FROM golang:1.22-alpine

# Force rebuild for deployment
WORKDIR /app

# Install git, ca-certificates, and wget (needed for health checks)
RUN apk add --no-cache git ca-certificates wget

# Copy go.mod first for better caching
COPY go.mod ./

# Download dependencies (creates go.sum if missing)
RUN go mod download

# Copy source code (includes go.sum if it exists)
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-business-intelligence-gateway main.go

# Verify binary was created
RUN test -f kyb-business-intelligence-gateway || (echo "ERROR: binary not created" && exit 1)

# Expose port (Railway will set PORT environment variable)
EXPOSE 8087

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8087}/health || exit 1

# Run the application
CMD ["./kyb-business-intelligence-gateway"]
# Force rebuild - 2025-12-09
# BI-SERVICE DOCKERFILE - DO NOT USE ROOT DOCKERFILE
