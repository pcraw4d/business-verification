#!/bin/sh
# Start script for Railway deployment
# Railway sets PORT environment variable automatically

# Use Railway's PORT, but log it for debugging
PORT=${PORT:-8000}
echo "🌐 Starting Python ML Service on port $PORT"

# Start uvicorn with Railway's PORT
exec python -m uvicorn app:app --host 0.0.0.0 --port "$PORT" --workers 1

