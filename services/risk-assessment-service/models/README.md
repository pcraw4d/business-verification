# Model Files Directory

This directory contains the trained machine learning models used by the Risk Assessment Service.

## Model Files

### LSTM Model
- **File**: `risk_lstm_v1.onnx`
- **Type**: ONNX format LSTM model
- **Purpose**: Long-term risk prediction (6-24 months)
- **Size**: ~5-10MB
- **Features**: 25+ temporal features
- **Accuracy**: 85-90% for 6-12 month predictions

### XGBoost Model
- **File**: `xgb_model.json`
- **Type**: XGBoost JSON format
- **Purpose**: Short-term risk prediction (1-3 months)
- **Size**: ~1-2MB
- **Features**: 50+ business features
- **Accuracy**: 88-92% for 1-3 month predictions

## Model Loading

Models are automatically loaded by the ModelManager during service initialization:

```go
// LSTM model path
LSTM_MODEL_PATH=/app/models/risk_lstm_v1.onnx

// XGBoost model path
XGBOOST_MODEL_PATH=/app/models/xgb_model.json
```

## Model Updates

To update models:

1. Train new models using the Python training environment
2. Export to appropriate formats (ONNX for LSTM, JSON for XGBoost)
3. Replace model files in this directory
4. Restart the service

## Model Validation

Models are validated during service startup:

- File existence checks
- Format validation
- ONNX model loading tests
- XGBoost model loading tests
- Performance benchmarks

## Security

- Model files are read-only in production
- Models are loaded into memory at startup
- No model files are exposed via HTTP endpoints
- Model paths are configurable via environment variables

## Performance

- LSTM model: ~60-120ms inference time
- XGBoost model: ~20-40ms inference time
- Ensemble: ~80-150ms inference time
- Memory usage: ~1-1.5GB total

## Troubleshooting

### Model Loading Errors
- Check file permissions
- Verify model file integrity
- Check ONNX Runtime library path
- Validate model format compatibility

### Performance Issues
- Monitor memory usage
- Check CPU utilization
- Verify model optimization
- Review inference latency metrics
