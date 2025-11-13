# Simple Go-based frontend server - Force rebuild for 404 fix
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Disable Go workspace mode to avoid "module appears multiple times" error
ENV GOWORK=off

# Copy the entire repository structure (needed for module resolution)
COPY . .

# Change to service directory where go.mod is located
WORKDIR /app/services/frontend

# Download dependencies for the frontend service module
RUN go mod download

# Debug: List what was copied (check if public directory exists)
RUN ls -la && (test -d public && ls -la public/ || echo "public directory not found")

# Build the frontend server with debug info from the service directory
RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -a -installsuffix cgo -ldflags="-X main.version=5.0.1-MERCHANT-CENTRIC-UI-FORCE-DEPLOY" -o frontend-server ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage (built in /app/services/frontend)
COPY --from=builder /app/services/frontend/frontend-server .

# Copy public files (from service directory)
COPY --from=builder /app/services/frontend/public ./public

# Expose port
EXPOSE 8080

# Run the frontend server
CMD ["./frontend-server"]
