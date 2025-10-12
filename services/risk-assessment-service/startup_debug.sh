#!/bin/sh

echo "🔍 ONNX Runtime Debug Information:"
echo "=================================="

echo "Environment Variables:"
echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"
echo "CGO_ENABLED: $CGO_ENABLED"
echo "LSTM_MODEL_PATH: $LSTM_MODEL_PATH"
echo "XGBOOST_MODEL_PATH: $XGBOOST_MODEL_PATH"

echo ""
echo "File System Check:"
echo "ONNX Runtime lib directory:"
ls -la /app/onnxruntime/lib/ 2>/dev/null || echo "❌ ONNX Runtime lib directory not found"

echo ""
echo "Model files:"
ls -la /app/models/ 2>/dev/null || echo "❌ Models directory not found"

echo ""
echo "Shared libraries:"
find /app/onnxruntime -name "*.so*" 2>/dev/null || echo "❌ No shared libraries found"

echo ""
echo "Starting application..."
exec ./risk-assessment-service
