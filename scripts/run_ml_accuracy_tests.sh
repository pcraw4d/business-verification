#!/bin/bash
# Run accuracy tests with ML (DistilBART) support
# This script starts the Python ML service and runs accuracy tests

set -e

echo "🚀 Starting ML-Enabled Accuracy Tests"
echo "======================================"
echo ""

# Check if Python ML service is already running
if curl -f http://localhost:8000/health > /dev/null 2>&1; then
    echo "✅ Python ML service is already running on port 8000"
    ML_SERVICE_RUNNING=true
else
    echo "⚠️  Python ML service is not running"
    ML_SERVICE_RUNNING=false
fi

# Start Python ML service if not running
if [ "$ML_SERVICE_RUNNING" = false ]; then
    echo ""
    echo "📦 Starting Python ML service..."
    echo "   (This will run in the background)"
    
    cd python_ml_service
    
    # Check if virtual environment exists
    if [ ! -d "venv" ]; then
        echo "⚠️  Virtual environment not found. Creating one..."
        python3 -m venv venv
    fi
    
    # Activate virtual environment
    source venv/bin/activate
    
    # Install dependencies if needed (check for torch, the most critical dependency)
    if ! python -c "import torch" 2>/dev/null; then
        echo "📥 Installing Python dependencies..."
        echo "   (This may take a few minutes - installing PyTorch and other ML libraries...)"
        pip install -q -r requirements.txt
    else
        echo "✅ Python dependencies already installed"
    fi
    
    # Start the service in background
    echo "🚀 Starting Python ML service on port 8000..."
    nohup python app.py > ../python_ml_service.log 2>&1 &
    ML_SERVICE_PID=$!
    echo "   Service started with PID: $ML_SERVICE_PID"
    
    cd ..
    
    # Wait for service to be ready
    echo "⏳ Waiting for service to be ready..."
    for i in {1..60}; do
        if curl -f http://localhost:8000/health > /dev/null 2>&1; then
            echo "✅ Python ML service is ready!"
            break
        fi
        if [ $i -eq 60 ]; then
            echo "❌ Service failed to start after 60 attempts"
            echo "   Check logs: tail -f python_ml_service.log"
            exit 1
        fi
        sleep 2
    done
fi

# Verify service is working
echo ""
echo "🔍 Verifying Python ML service..."
HEALTH_RESPONSE=$(curl -s http://localhost:8000/health)
if echo "$HEALTH_RESPONSE" | grep -q "healthy\|ok"; then
    echo "✅ Service health check passed"
else
    echo "⚠️  Service health check returned: $HEALTH_RESPONSE"
fi

# Check if quantization is enabled
echo ""
echo "🔍 Checking ML service configuration..."
MODEL_INFO=$(curl -s http://localhost:8000/model-info 2>/dev/null || echo "{}")
if echo "$MODEL_INFO" | grep -q "quantization_enabled.*true"; then
    echo "✅ Quantization is enabled"
else
    echo "ℹ️  Quantization status unknown (service may still work)"
fi

# Set environment variable for accuracy tests
export PYTHON_ML_SERVICE_URL="http://localhost:8000"
echo ""
echo "✅ Set PYTHON_ML_SERVICE_URL=http://localhost:8000"

# Set other required environment variables if not already set
if [ -z "$SUPABASE_URL" ]; then
    echo "⚠️  SUPABASE_URL not set. Please set it before running tests."
    echo "   Example: export SUPABASE_URL='https://your-project.supabase.co'"
fi

if [ -z "$SUPABASE_ANON_KEY" ]; then
    echo "⚠️  SUPABASE_ANON_KEY not set. Please set it before running tests."
fi

if [ -z "$DATABASE_URL" ]; then
    echo "⚠️  DATABASE_URL not set. Please set it before running tests."
fi

# Run accuracy tests
echo ""
echo "🧪 Running accuracy tests with ML support..."
echo "============================================"
echo ""

# Build the test binary if needed
if [ ! -f "bin/comprehensive_accuracy_test" ]; then
    echo "📦 Building accuracy test binary..."
    go build -o bin/comprehensive_accuracy_test ./cmd/comprehensive_accuracy_test
fi

# Run tests
OUTPUT_FILE="accuracy_report_ml_$(date +%Y%m%d_%H%M%S).json"
./bin/comprehensive_accuracy_test -verbose -output "$OUTPUT_FILE"

echo ""
echo "✅ Accuracy tests completed!"
echo "   Results saved to: $OUTPUT_FILE"
echo ""
echo "📊 To compare with keyword-based results:"
echo "   Run without PYTHON_ML_SERVICE_URL to get keyword-based results"
echo ""
echo "🛑 To stop the Python ML service:"
if [ -n "$ML_SERVICE_PID" ]; then
    echo "   kill $ML_SERVICE_PID"
else
    echo "   pkill -f 'python.*app.py'"
fi

