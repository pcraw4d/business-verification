FROM golang:1.22-alpine

WORKDIR /app

# Install git, ca-certificates, and wget (needed for health check)
RUN apk add --no-cache git ca-certificates wget

# Copy go.mod first for better caching
COPY go.mod ./

# Initialize go.sum if it doesn't exist and download dependencies
RUN go mod tidy && go mod download

# Copy go.sum if it exists
COPY go.sum* ./

# Copy source code
COPY . .

# Verify static directory exists before build
RUN test -d static || (echo "ERROR: static directory not found" && exit 1)

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o frontend-service main.go

# Verify binary was created
RUN test -f frontend-service || (echo "ERROR: frontend-service binary not created" && exit 1)

# Expose port
EXPOSE 8086

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8086}/health || exit 1

# Run the application
CMD ["./frontend-service"]
