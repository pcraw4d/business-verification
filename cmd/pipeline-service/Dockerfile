FROM golang:1.22-alpine

WORKDIR /app

# Install git, ca-certificates, and wget (needed for health checks)
RUN apk add --no-cache git ca-certificates wget

# Copy go.mod first for better caching (go.sum is optional)
COPY go.mod ./

# Download dependencies (this will create go.sum if missing)
RUN go mod download

# Copy go.sum if it exists (optional)
COPY go.sum* ./

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pipeline-service main.go

# Expose port
EXPOSE 8085

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8085}/health || exit 1

# Run the application
CMD ["./pipeline-service"]
