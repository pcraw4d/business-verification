# Simple Go-based frontend server - Force rebuild for 404 fix
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Debug: List what was copied
RUN ls -la && ls -la public/

# Build the frontend server with debug info
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-X main.version=5.0.1-MERCHANT-CENTRIC-UI-FORCE-DEPLOY" -o frontend-server ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/frontend-server .

# Copy public files
COPY --from=builder /app/public ./public

# Expose port
EXPOSE 8080

# Run the frontend server
CMD ["./frontend-server"]
