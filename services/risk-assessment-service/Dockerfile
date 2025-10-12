# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies including build tools for ONNX Runtime
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev wget tar

# Download and install ONNX Runtime C libraries
RUN wget -q https://github.com/microsoft/onnxruntime/releases/download/v1.16.0/onnxruntime-linux-x64-1.16.0.tgz && \
    tar -xzf onnxruntime-linux-x64-1.16.0.tgz && \
    mv onnxruntime-linux-x64-1.16.0 onnxruntime && \
    rm onnxruntime-linux-x64-1.16.0.tgz && \
    echo "ONNX Runtime installed:" && \
    ls -la /app/onnxruntime/ && \
    echo "Library files:" && \
    ls -la /app/onnxruntime/lib/ && \
    echo "Include files:" && \
    ls -la /app/onnxruntime/include/

# Set environment variables for ONNX Runtime
ENV CGO_ENABLED=1
ENV CGO_LDFLAGS="-L/app/onnxruntime/lib"
ENV CGO_CFLAGS="-I/app/onnxruntime/include"

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with ONNX Runtime support
RUN GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags='-w -s' \
    -o risk-assessment-service \
    ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates, timezone data, and wget for health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Create app directory and required subdirectories
RUN mkdir -p /app/logs /app/models /app/onnxruntime/lib
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/risk-assessment-service .

# Copy startup debug script
COPY startup_debug.sh .
RUN chmod +x startup_debug.sh

# Copy ONNX Runtime libraries from builder stage
COPY --from=builder /app/onnxruntime/lib/* /app/onnxruntime/lib/

# Debug: List copied libraries
RUN echo "Copied ONNX Runtime libraries:" && \
    ls -la /app/onnxruntime/lib/ && \
    echo "Library path contents:" && \
    find /app/onnxruntime -name "*.so*" -o -name "*.a" | head -10

# Copy model files (if they exist)
COPY --from=builder /app/models/*.onnx /app/models/ 2>/dev/null || true
COPY --from=builder /app/models/*.json /app/models/ 2>/dev/null || true

# Debug: List copied model files
RUN echo "Copied model files:" && \
    ls -la /app/models/ && \
    echo "Model file sizes:" && \
    du -h /app/models/* 2>/dev/null || echo "No model files found"

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set environment variables for ONNX Runtime
ENV LD_LIBRARY_PATH=/app/onnxruntime/lib:$LD_LIBRARY_PATH
ENV LSTM_MODEL_PATH=/app/models/risk_lstm_v1.onnx
ENV XGBOOST_MODEL_PATH=/app/models/xgb_model.json
ENV ENABLE_ENSEMBLE=true
ENV DEFAULT_PREDICTION_HORIZON=3

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (Railway will set PORT environment variable)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the application with debug information
CMD ["./startup_debug.sh"]
