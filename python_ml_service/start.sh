#!/bin/sh
# Start script for Railway deployment
# Railway sets PORT environment variable automatically

PORT=${PORT:-8000}
exec python -m uvicorn app:app --host 0.0.0.0 --port "$PORT" --workers 1

