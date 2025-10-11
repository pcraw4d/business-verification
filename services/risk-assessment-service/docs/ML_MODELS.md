# Machine Learning Models Guide

## Overview

The Risk Assessment Service now includes advanced machine learning capabilities with LSTM time-series prediction and smart ensemble routing. This guide covers the available models, their capabilities, and when to use each one.

## Available Models

### 1. XGBoost Model
- **Type**: Gradient Boosting
- **Best For**: Short-term predictions (1-3 months)
- **Accuracy**: 88-92% for 1-3 month horizons
- **Latency**: 20-40ms
- **Use Cases**: 
  - Quick risk assessments
  - Real-time decision making
  - High-frequency predictions

### 2. LSTM Model
- **Type**: Long Short-Term Memory Neural Network
- **Best For**: Long-term predictions (6-24 months)
- **Accuracy**: 85-90% for 6-12 month horizons
- **Latency**: 60-120ms
- **Use Cases**:
  - Strategic planning
  - Long-term risk forecasting
  - Time-series pattern analysis

### 3. Ensemble Model
- **Type**: Smart combination of XGBoost and LSTM
- **Best For**: Medium-term predictions (3-6 months)
- **Accuracy**: 87-91% for 3-6 month horizons
- **Latency**: 80-150ms
- **Use Cases**:
  - Balanced accuracy and speed
  - When both short and long-term factors matter
  - Comprehensive risk analysis

## Smart Model Routing

The system automatically selects the best model based on prediction horizon:

| Horizon | Selected Model | Reasoning |
|---------|---------------|-----------|
| 1-3 months | XGBoost | Proven accuracy, low latency |
| 3-6 months | Ensemble | Weighted combination (70% XGBoost, 30% LSTM) |
| 6-12 months | LSTM | Temporal patterns, long-term forecasting |
| 12+ months | LSTM | Extended time-series analysis |

## LSTM Model Details

### Architecture
- **Input**: Time-series sequences (12 months of historical data)
- **Layers**: 2 LSTM layers with 128 hidden units each
- **Attention**: Attention mechanism for interpretability
- **Output**: Multi-horizon predictions (6, 9, 12 months)
- **Regularization**: Dropout (0.3) and L2 regularization

### Features
- **Temporal Analysis**: Captures seasonal patterns and trends
- **Uncertainty Estimation**: Monte Carlo dropout for confidence scoring
- **Multi-Horizon Output**: Single model, multiple prediction windows
- **Attention Weights**: Shows which time periods matter most

### Training Data
- **Synthetic Data**: 2-3 years of generated time-series data
- **Business Count**: 10,000+ businesses across industries
- **Features**: 25+ temporal features including:
  - Financial health trends
  - Compliance history
  - Market conditions
  - Risk volatility patterns

## API Usage

### Basic Risk Assessment
```bash
curl -X POST /api/v1/assess \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Tech Corp",
    "business_address": "123 Main St, City, ST 12345",
    "industry": "technology",
    "country": "US",
    "prediction_horizon": 6,
    "model_type": "auto"
  }'
```

### Advanced Multi-Horizon Prediction
```bash
curl -X POST /api/v1/risk/predict-advanced \
  -H "Content-Type: application/json" \
  -d '{
    "business": {
      "business_name": "Manufacturing Inc",
      "business_address": "456 Industrial Ave, Factory City, FC 67890",
      "industry": "manufacturing",
      "country": "US"
    },
    "prediction_horizons": [3, 6, 9, 12],
    "model_preference": "auto",
    "include_temporal_analysis": true,
    "include_scenario_analysis": true,
    "include_model_comparison": true
  }'
```

### Model-Specific Prediction
```bash
curl -X POST /api/v1/assess/{id}/predict \
  -H "Content-Type: application/json" \
  -d '{
    "horizon_months": 12,
    "model_type": "lstm",
    "include_temporal_analysis": true
  }'
```

## Response Examples

### LSTM Prediction Response
```json
{
  "id": "risk_1703123456789",
  "business_id": "biz_123456789",
  "risk_score": 0.72,
  "risk_level": "medium",
  "prediction_horizon": 12,
  "confidence_score": 0.85,
  "metadata": {
    "model_type": "lstm",
    "temporal_analysis": {
      "trend_analysis": {
        "overall_trend": "stable",
        "trend_strength": 0.3,
        "seasonal_patterns": ["q1_increase", "q4_decrease"]
      },
      "volatility_analysis": {
        "historical_volatility": 0.15,
        "volatility_trend": "decreasing",
        "volatility_forecast": 0.12
      }
    }
  }
}
```

### Advanced Prediction Response
```json
{
  "request_id": "adv_pred_1703123456789",
  "business_id": "biz_123456789",
  "predictions": {
    "3": {
      "horizon_months": 3,
      "model_type": "xgboost",
      "predicted_score": 0.65,
      "predicted_level": "medium",
      "confidence_score": 0.88
    },
    "6": {
      "horizon_months": 6,
      "model_type": "lstm",
      "predicted_score": 0.72,
      "predicted_level": "medium",
      "confidence_score": 0.82
    }
  },
  "model_comparison": {
    "best_model_per_horizon": {
      "3": "xgboost",
      "6": "lstm",
      "9": "lstm",
      "12": "lstm"
    },
    "agreement_analysis": {
      "overall_agreement": 0.85,
      "high_disagreement_horizons": [12]
    }
  },
  "temporal_analysis": {
    "trend_analysis": {
      "overall_trend": "stable",
      "trend_strength": 0.3
    }
  },
  "scenario_analysis": [
    {
      "scenario_name": "Optimistic",
      "probability": 0.2,
      "risk_score": 0.2,
      "risk_level": "low"
    }
  ],
  "confidence_analysis": {
    "overall_confidence": 0.85,
    "confidence_by_horizon": {
      "3": 0.88,
      "6": 0.82,
      "9": 0.78,
      "12": 0.75
    }
  }
}
```

## Performance Characteristics

### Latency (p95)
- **XGBoost**: 20-40ms
- **LSTM**: 60-120ms
- **Ensemble**: 80-150ms

### Memory Usage
- **XGBoost**: ~200MB
- **LSTM**: ~500MB
- **Total Service**: ~1.5GB

### Throughput
- **XGBoost**: 2000+ req/min
- **LSTM**: 1500+ req/min
- **Ensemble**: 1200+ req/min

## Accuracy Expectations

### By Horizon
| Horizon | XGBoost | LSTM | Ensemble |
|---------|---------|------|----------|
| 1 month | 92% | 85% | 91% |
| 3 months | 90% | 87% | 89% |
| 6 months | 85% | 88% | 87% |
| 12 months | 78% | 85% | 82% |

### By Industry
| Industry | XGBoost | LSTM | Ensemble |
|----------|---------|------|----------|
| Technology | 89% | 87% | 88% |
| Financial | 91% | 89% | 90% |
| Manufacturing | 87% | 90% | 88% |
| Healthcare | 88% | 86% | 87% |

## Best Practices

### Model Selection
1. **Use "auto"** for most cases - the system will choose the best model
2. **Specify model** only when you have specific requirements
3. **Consider latency** vs accuracy trade-offs
4. **Use ensemble** for balanced predictions

### Prediction Horizons
1. **1-3 months**: Use for operational decisions
2. **3-6 months**: Use for tactical planning
3. **6-12 months**: Use for strategic planning
4. **12+ months**: Use for long-term forecasting

### Confidence Thresholds
1. **High confidence (>0.8)**: Use for automated decisions
2. **Medium confidence (0.6-0.8)**: Use with human review
3. **Low confidence (<0.6)**: Use for exploratory analysis only

## Troubleshooting

### Common Issues

#### Low Confidence Scores
- **Cause**: Insufficient historical data
- **Solution**: Use ensemble model or synthetic data generation

#### High Latency
- **Cause**: Model loading or inference issues
- **Solution**: Check model files and ONNX runtime

#### Inaccurate Predictions
- **Cause**: Model drift or data quality issues
- **Solution**: Retrain models or validate input data

### Error Codes
- `MODEL_NOT_FOUND`: Model file missing or corrupted
- `INFERENCE_FAILED`: ONNX runtime error
- `INVALID_INPUT`: Input data validation failed
- `TIMEOUT`: Prediction timeout exceeded

## Model Updates

### Retraining Schedule
- **XGBoost**: Monthly retraining with new data
- **LSTM**: Quarterly retraining with expanded dataset
- **Ensemble**: Updated when component models change

### Version Management
- Models are versioned (e.g., `risk_lstm_v1.onnx`)
- Backward compatibility maintained for 2 versions
- Gradual rollout for new model versions

## Cost Implications

### Compute Costs
- **XGBoost**: $0.0001 per prediction
- **LSTM**: $0.0002 per prediction
- **Ensemble**: $0.0003 per prediction

### Storage Costs
- **Model files**: ~50MB total
- **Cache**: ~100MB for temporal features
- **Total**: ~150MB per service instance

## Monitoring and Alerts

### Key Metrics
- **Latency**: p50, p95, p99 by model
- **Accuracy**: Rolling 30-day accuracy
- **Throughput**: Requests per minute
- **Error Rate**: Failed predictions percentage

### Alerts
- **High Latency**: >200ms p95
- **Low Accuracy**: <80% for any model
- **High Error Rate**: >5% failed predictions
- **Memory Usage**: >2GB per instance

## Future Enhancements

### Planned Features
- **Transformer Models**: For even better long-term predictions
- **Real-time Learning**: Continuous model updates
- **Multi-modal Input**: Text and image analysis
- **Explainable AI**: Detailed prediction explanations

### Research Areas
- **Federated Learning**: Privacy-preserving model training
- **Causal Inference**: Understanding cause-effect relationships
- **Uncertainty Quantification**: Better confidence estimation
- **Adversarial Robustness**: Defense against model attacks
