FROM golang:1.22-alpine

WORKDIR /app

# Install git, ca-certificates, and wget (needed for health checks)
RUN apk add --no-cache git ca-certificates wget

# Copy go.mod first for better caching
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy go.sum if it exists
COPY go.sum* ./

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-service-discovery main.go

# Verify binary was created
RUN test -f kyb-service-discovery || (echo "ERROR: binary not created" && exit 1)

# Expose port (Railway will set PORT environment variable)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the application
CMD ["./kyb-service-discovery"]
