# Railway Deployment Guide

This guide covers deploying the LSTM-enhanced Risk Assessment Service to Railway.

## Overview

The Risk Assessment Service is deployed to Railway with the following features:
- **LSTM Model**: ONNX-based time-series prediction for 6-24 month forecasts
- **XGBoost Model**: Traditional ML model for 1-3 month predictions
- **Ensemble Routing**: Smart model selection based on prediction horizon
- **Performance Monitoring**: Real-time metrics and health checks
- **Auto-scaling**: Automatic scaling based on CPU and memory usage

## Prerequisites

### Required Tools
- [Railway CLI](https://docs.railway.app/develop/cli) installed
- [Docker](https://www.docker.com/) running locally
- Railway account with project access

### Installation
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login
```

## Deployment Process

### 1. Automated Deployment

Use the provided deployment script:

```bash
# Deploy to production
./scripts/deploy-railway.sh

# Deploy to staging
./scripts/deploy-railway.sh staging

# Deploy to development
./scripts/deploy-railway.sh development
```

### 2. Manual Deployment

```bash
# Initialize Railway project (if not already done)
railway init

# Deploy the service
railway up --detach

# Check deployment status
railway status

# View logs
railway logs
```

## Configuration

### Environment Variables

The service uses the following environment variables:

#### Core Configuration
- `PORT`: Service port (default: 8080)
- `ENVIRONMENT`: Deployment environment (production/staging/development)
- `LOG_LEVEL`: Logging level (debug/info/warn/error)

#### Model Configuration
- `LSTM_MODEL_PATH`: Path to LSTM ONNX model file
- `XGBOOST_MODEL_PATH`: Path to XGBoost model file
- `ONNX_RUNTIME_PATH`: Path to ONNX Runtime libraries
- `ENABLE_ENSEMBLE`: Enable ensemble routing (true/false)
- `DEFAULT_PREDICTION_HORIZON`: Default prediction horizon in months

#### Performance Configuration
- `MAX_CONCURRENT_REQUESTS`: Maximum concurrent requests
- `REQUEST_TIMEOUT`: Request timeout duration
- `MODEL_LOAD_TIMEOUT`: Model loading timeout

#### Monitoring Configuration
- `METRICS_ENABLED`: Enable metrics collection
- `METRICS_INTERVAL`: Metrics collection interval
- `HEALTH_CHECK_INTERVAL`: Health check interval
- `ALERT_THRESHOLDS`: JSON string with alert thresholds

### Railway Configuration

The `railway.json` file contains Railway-specific configuration:

```json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "numReplicas": 1,
    "restartPolicyType": "ON_FAILURE",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 10
  },
  "resources": {
    "memory": "2GB",
    "cpu": "1.0",
    "disk": "10GB"
  }
}
```

## Service Architecture

### Container Structure
```
/app/
├── risk-assessment-service          # Main application binary
├── models/                          # ML model files
│   ├── risk_lstm_v1.onnx          # LSTM model
│   └── xgb_model.json             # XGBoost model
├── onnxruntime/                    # ONNX Runtime libraries
│   └── lib/                       # C libraries
└── logs/                          # Application logs
```

### Model Loading
1. **Startup**: Models are loaded during service initialization
2. **LSTM Model**: ONNX Runtime loads the LSTM model from `/app/models/risk_lstm_v1.onnx`
3. **XGBoost Model**: XGBoost loads the model from `/app/models/xgb_model.json`
4. **Ensemble Router**: Initialized if both models are available

### Health Checks
- **Endpoint**: `GET /health`
- **Interval**: 30 seconds
- **Timeout**: 10 seconds
- **Checks**: Model availability, memory usage, error rates

## Monitoring and Observability

### Metrics Endpoints
- **Overall Metrics**: `GET /metrics`
- **Model Performance**: `GET /metrics/model/{model_type}`
- **Health Status**: `GET /health`
- **Performance Snapshot**: `GET /metrics/snapshot`

### Key Metrics
- **Latency**: P50, P95, P99 for each model
- **Throughput**: Requests per minute
- **Error Rate**: Percentage of failed requests
- **Memory Usage**: Current memory consumption
- **Model Distribution**: Usage by model type
- **Horizon Distribution**: Usage by prediction horizon

### Alerts
- **High Error Rate**: >5% error rate
- **High Latency**: P95 >200ms
- **High Memory Usage**: >1GB memory usage
- **Model Failures**: Model loading or inference failures

## Scaling Configuration

### Auto-scaling
- **Min Replicas**: 1
- **Max Replicas**: 3
- **Target CPU**: 70%
- **Target Memory**: 80%
- **Scale Up Cooldown**: 5 minutes
- **Scale Down Cooldown**: 10 minutes

### Resource Limits
- **Memory**: 2GB per replica
- **CPU**: 1.0 cores per replica
- **Disk**: 10GB per replica

## Security Configuration

### HTTPS
- **Enabled**: Automatic HTTPS with Railway
- **Certificates**: Managed by Railway
- **Redirect**: HTTP to HTTPS redirect

### CORS
- **Enabled**: Cross-origin resource sharing
- **Origins**: Configurable (default: *)
- **Methods**: GET, POST, PUT, DELETE, OPTIONS
- **Headers**: Content-Type, Authorization, X-Requested-With

### Rate Limiting
- **Enabled**: Request rate limiting
- **Limit**: 1000 requests per hour
- **Burst**: 100 requests
- **Window**: 1 hour

## API Endpoints

### Core Endpoints
- `POST /api/v1/assess` - Risk assessment
- `GET /api/v1/assess/{id}` - Get assessment by ID
- `POST /api/v1/assess/{id}/predict` - Risk prediction

### Advanced Endpoints
- `POST /api/v1/risk/predict-advanced` - Multi-horizon prediction
- `GET /api/v1/models/info` - Model information
- `GET /api/v1/models/performance` - Model performance metrics

### Monitoring Endpoints
- `GET /health` - Health check
- `GET /metrics` - System metrics
- `GET /metrics/model/{type}` - Model-specific metrics

## Troubleshooting

### Common Issues

#### Model Loading Failures
```bash
# Check model files
railway shell
ls -la /app/models/

# Check ONNX Runtime
ldd /app/risk-assessment-service

# Check environment variables
railway variables
```

#### High Memory Usage
```bash
# Check memory usage
railway logs | grep "memory"

# Check model sizes
railway shell
du -sh /app/models/*
```

#### Performance Issues
```bash
# Check latency metrics
curl https://your-service.railway.app/metrics

# Check error rates
railway logs | grep "error"
```

### Debug Commands

```bash
# View service logs
railway logs

# Access service shell
railway shell

# Check service status
railway status

# View environment variables
railway variables

# Redeploy service
railway redeploy
```

### Performance Optimization

#### Model Optimization
- **ONNX Optimization**: Models are optimized for inference
- **Memory Management**: Models are loaded once at startup
- **Caching**: Prediction results are cached when appropriate

#### Request Optimization
- **Connection Pooling**: HTTP connections are pooled
- **Request Batching**: Multiple predictions can be batched
- **Async Processing**: Non-blocking request handling

## Maintenance

### Regular Tasks
- **Log Rotation**: Automatic log rotation and cleanup
- **Model Updates**: Deploy new model versions
- **Security Updates**: Keep dependencies updated
- **Performance Monitoring**: Monitor key metrics

### Backup Strategy
- **Model Files**: Backed up to version control
- **Configuration**: Environment variables backed up
- **Logs**: Retained for 7 days
- **Metrics**: Retained for 24 hours

### Update Process
1. **Test Locally**: Test changes locally with Docker
2. **Deploy to Staging**: Deploy to staging environment
3. **Run Tests**: Execute smoke tests and validation
4. **Deploy to Production**: Deploy to production environment
5. **Monitor**: Monitor metrics and alerts

## Cost Optimization

### Resource Usage
- **Memory**: Optimized model loading and caching
- **CPU**: Efficient inference with ONNX Runtime
- **Storage**: Minimal disk usage with optimized models
- **Network**: Efficient API responses

### Scaling Strategy
- **Horizontal Scaling**: Add replicas based on load
- **Vertical Scaling**: Increase resources if needed
- **Auto-scaling**: Automatic scaling based on metrics

### Cost Monitoring
- **Resource Usage**: Monitor CPU, memory, and storage
- **Request Volume**: Track API usage and costs
- **Performance**: Optimize for cost-effectiveness

## Support

### Documentation
- **API Documentation**: Available at `/docs` endpoint
- **Model Documentation**: See `docs/ML_MODELS.md`
- **Deployment Guide**: This document

### Monitoring
- **Health Checks**: Automatic health monitoring
- **Alerts**: Email and Slack notifications
- **Metrics**: Real-time performance metrics

### Troubleshooting
- **Logs**: Comprehensive logging for debugging
- **Metrics**: Detailed performance metrics
- **Health Endpoints**: Service health information