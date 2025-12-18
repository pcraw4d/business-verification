#!/bin/bash
# Startup script for embedding-service on Railway
# Handles PORT environment variable expansion

set -e

# Use Railway's PORT if set, otherwise default to 8000
PORT=${PORT:-8000}

echo "Starting embedding-service on port $PORT"

# Start uvicorn with the resolved port
exec python3 -m uvicorn app:app --host 0.0.0.0 --port "$PORT"

