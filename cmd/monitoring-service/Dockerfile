FROM golang:1.22-alpine

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go.mod first for better caching
COPY go.mod ./

# Download dependencies (this will create go.sum if missing)
RUN go mod download

# Copy source code (go.sum will be included if it exists)
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-monitoring main.go

# Expose port
EXPOSE 8084

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8084/health || exit 1

# Run the application
CMD ["./kyb-monitoring"]
