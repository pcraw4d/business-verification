# Risk Assessment Service

## Overview
This is the risk assessment service for the KYB Platform, providing comprehensive business risk assessment and predictive analytics capabilities.

## Features
- **Real-time risk assessment** with sub-1-second response times
- **Advanced ML Models**:
  - XGBoost for short-term predictions (1-3 months)
  - LSTM for long-term time-series forecasting (6-24 months)
  - Smart ensemble routing for optimal accuracy
- **Multi-horizon predictions** with temporal analysis
- **Advanced scenario analysis** and model comparison
- **SHAP explainability** for risk factors
- **Integration with external data sources** (Thomson Reuters, OFAC, etc.)
- **Comprehensive compliance screening**
- **Multi-tenant architecture** for enterprise customers

## ML Model Capabilities

### Smart Model Routing
The system automatically selects the best model based on prediction horizon:
- **1-3 months**: XGBoost (88-92% accuracy, 20-40ms latency)
- **3-6 months**: Ensemble (87-91% accuracy, 80-150ms latency)
- **6-12 months**: LSTM (85-90% accuracy, 60-120ms latency)
- **12+ months**: LSTM (extended time-series analysis)

### LSTM Time-Series Prediction
- **Architecture**: 2-layer LSTM with attention mechanism
- **Features**: 25+ temporal features including financial trends, compliance history
- **Training**: 10,000+ businesses with 2-3 years of synthetic data
- **Output**: Multi-horizon predictions with uncertainty estimation

### Advanced Prediction API
- **Multi-horizon predictions** in a single request
- **Model comparison** across different approaches
- **Temporal analysis** with trend and volatility forecasting
- **Scenario analysis** with optimistic, base case, and pessimistic outcomes
- **Confidence analysis** with calibration scoring

## API Endpoints

### Core Risk Assessment
- `POST /api/v1/assess` - Perform risk assessment with LSTM/ensemble support
- `GET /api/v1/assess/{id}` - Get risk assessment by ID
- `POST /api/v1/assess/{id}/predict` - Generate risk prediction with model selection

### Advanced ML Features
- `POST /api/v1/risk/predict-advanced` - Multi-horizon prediction with model comparison
- `GET /api/v1/models/info` - Get model information and capabilities
- `GET /api/v1/models/performance` - Get model performance metrics

### Compliance & Monitoring
- `POST /api/v1/compliance/check` - Perform compliance screening
- `POST /api/v1/sanctions/screen` - Screen against sanctions lists
- `POST /api/v1/media/monitor` - Monitor adverse media
- `GET /api/v1/analytics/trends` - Get risk trends and insights

## Development

### Local Development
```bash
# From project root
make dev-risk-assessment

# Or directly
cd services/risk-assessment-service
go run cmd/main.go
```

### Testing
```bash
make test-risk-assessment
```

### ML Model Training
```bash
# Set up Python environment
cd ml-training
pip install -r requirements.txt

# Train LSTM model
python training/train_lstm.py

# Export to ONNX
python export/export_to_onnx.py
```

### Performance Testing
```bash
# Run benchmarks
go run cmd/benchmark_lstm.go

# Run validation
go run cmd/run_validation.go
```

### Deployment
```bash
make deploy-risk-assessment
```

## Configuration
The service is configured to integrate with:
- Supabase for data storage
- Redis for caching
- External APIs for risk data
- Prometheus for monitoring

## File Structure
- `cmd/` - Main application entry point and utilities
  - `main.go` - Application entry point
  - `benchmark_lstm.go` - Performance benchmarks
  - `validate_model.go` - Model validation
- `internal/` - Private application code
  - `config/` - Configuration management
  - `handlers/` - HTTP handlers (including advanced prediction)
  - `models/` - Data models and structures
  - `ml/` - Machine learning models and training
    - `models/` - Model implementations (XGBoost, LSTM, ensemble)
    - `ensemble/` - Ensemble routing and combination logic
    - `data/` - Data generation and processing (synthetic, hybrid)
    - `validation/` - Model validation and testing
    - `service/` - ML service orchestration
- `ml-training/` - Python ML training environment
  - `data/` - Data generation and preprocessing
  - `models/` - LSTM model architecture
  - `training/` - Training scripts and hyperparameter tuning
  - `export/` - ONNX export utilities
- `docs/` - Documentation
  - `ML_MODELS.md` - Comprehensive ML models guide
  - `API_DOCUMENTATION.md` - API documentation
- `api/` - OpenAPI specifications
- `pkg/client/` - Go client SDK
- `sdks/` - Multi-language SDKs (Python, Node.js)
- `Dockerfile` - Container configuration
