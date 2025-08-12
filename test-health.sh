#!/bin/bash

# Health Check Test Script
# Usage: ./test-health.sh [URL]

URL="${1:-http://localhost:8080}"

echo "🏥 Testing health endpoint: $URL/health"

response=$(curl -s -o /dev/null -w "%{http_code}" "$URL/health")

if [ "$response" = "200" ]; then
    echo "✅ Health check passed!"
    echo "🌐 Application is running at: $URL"
else
    echo "❌ Health check failed (HTTP $response)"
    echo "🔍 Check the application logs"
fi
