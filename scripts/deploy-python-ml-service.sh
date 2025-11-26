#!/bin/bash
# Deploy Python ML Service with Quantization
# This script deploys the Python ML service with DistilBART quantization enabled

set -e

echo "🚀 Deploying Python ML Service with Quantization..."
echo ""

cd "$(dirname "$0")/../python_ml_service"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Build the production image
echo "📦 Building production Docker image..."
docker-compose build --target production python-ml-service

# Start the service
echo "🚀 Starting Python ML Service..."
docker-compose up -d python-ml-service

# Wait for service to be ready
echo "⏳ Waiting for service to start..."
sleep 10

# Check service status
echo ""
echo "📊 Service Status:"
docker-compose ps python-ml-service

# Check health
echo ""
echo "🏥 Checking service health..."
for i in {1..30}; do
    if curl -f http://localhost:8000/health > /dev/null 2>&1; then
        echo "✅ Service is healthy!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ Service health check failed after 30 attempts"
        exit 1
    fi
    sleep 2
done

# Check quantization status
echo ""
echo "🔍 Checking quantization status..."
QUANTIZATION_STATUS=$(curl -s http://localhost:8000/model-info | grep -o '"quantization_enabled":[^,]*' | cut -d: -f2 || echo "unknown")

if [ "$QUANTIZATION_STATUS" = "true" ]; then
    echo "✅ Quantization is ENABLED"
else
    echo "⚠️  Quantization status: $QUANTIZATION_STATUS"
    echo "   Check logs for details: docker-compose logs python-ml-service"
fi

# Show logs
echo ""
echo "📋 Recent logs (showing last 20 lines):"
echo "   (Use 'docker-compose logs -f python-ml-service' to follow logs)"
docker-compose logs --tail=20 python-ml-service | grep -E "(quantization|DistilBART|ERROR|✅)" || docker-compose logs --tail=20 python-ml-service

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📝 Useful commands:"
echo "   View logs:        docker-compose logs -f python-ml-service"
echo "   Check status:     docker-compose ps"
echo "   Stop service:     docker-compose down"
echo "   Restart service:  docker-compose restart python-ml-service"
echo "   Health check:     curl http://localhost:8000/health"
echo "   Model info:       curl http://localhost:8000/model-info | jq"

