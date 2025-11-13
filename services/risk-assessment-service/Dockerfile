# Build stage - Go 1.24.0 for compatibility with go.mod (requires go >= 1.24.0)
FROM golang:1.24-alpine AS builder

# Force cache bust for Railway - updated timestamp and image
ARG CACHE_BUST=2025-11-10-04-58
ARG BUILD_DATE=2025-11-10T04:58:00Z

# Verify Go version and ensure it's 1.24+
RUN go version && go version | grep -qE "go1\.(24|2[5-9]|[3-9][0-9])" || (echo "ERROR: Go version must be 1.24 or higher" && exit 1)

# Set working directory
WORKDIR /app

# Disable Go workspace mode to avoid version conflicts
# This is necessary because both root and service have workspace files
ENV GOWORK=off

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

# Verify go.sum is up to date and download dependencies
RUN go mod verify && go mod download

# Copy source code
COPY . .

# Debug: List copied files including models
RUN echo "Files in /app after COPY . .:" && \
    ls -la /app/ && \
    echo "Model files:" && \
    ls -la /app/models/ 2>/dev/null || echo "No models directory found"

# Build the application with ONNX Runtime support
# Explicitly disable workspace mode in build command
RUN GOOS=linux GOARCH=amd64 GOWORK=off go build \
    -a -installsuffix cgo \
    -ldflags='-w -s' \
    -o risk-assessment-service \
    ./cmd/main.go ./cmd/cache_logger_wrapper.go

# Debug script removed - file does not exist

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

# Copy ONNX Runtime libraries from builder stage (robust copy)
COPY --from=builder /app/onnxruntime/ /app/onnxruntime/

# Debug: List copied libraries and create symlink if needed
RUN echo "Copied ONNX Runtime libraries:" && \
    ls -la /app/onnxruntime/lib/ && \
    echo "Library path contents:" && \
    find /app/onnxruntime -name "*.so*" -o -name "*.a" | head -10 && \
    echo "Creating symlink for ONNX Runtime library..." && \
    if [ -f /app/onnxruntime/lib/libonnxruntime.so ] && [ ! -f /app/onnxruntime/lib/onnxruntime.so ]; then \
        ln -s /app/onnxruntime/lib/libonnxruntime.so /app/onnxruntime/lib/onnxruntime.so && \
        echo "✅ Created symlink: onnxruntime.so -> libonnxruntime.so"; \
    elif [ -f /app/onnxruntime/lib/libonnxruntime.so.1 ] && [ ! -f /app/onnxruntime/lib/onnxruntime.so ]; then \
        ln -s /app/onnxruntime/lib/libonnxruntime.so.1 /app/onnxruntime/lib/onnxruntime.so && \
        echo "✅ Created symlink: onnxruntime.so -> libonnxruntime.so.1"; \
    else \
        echo "⚠️  No libonnxruntime.so found, checking for alternative names..."; \
        ls -la /app/onnxruntime/lib/*.so* 2>/dev/null || echo "No .so files found"; \
    fi

# Copy model files explicitly
COPY --from=builder /app/models/risk_lstm_v1.onnx /app/models/risk_lstm_v1.onnx
COPY --from=builder /app/models/xgb_model.json /app/models/xgb_model.json

# Debug: List copied model files
RUN echo "Copied model files:" && \
    ls -la /app/models/ && \
    echo "Model file sizes:" && \
    du -h /app/models/* 2>/dev/null || echo "No model files found"

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set environment variables for ONNX Runtime
ENV LD_LIBRARY_PATH=/app/onnxruntime/lib
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

# Run the application
CMD ["./risk-assessment-service"]
