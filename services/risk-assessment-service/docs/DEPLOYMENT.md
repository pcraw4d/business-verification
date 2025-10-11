# Risk Assessment Service Deployment Guide

This guide provides comprehensive instructions for deploying the Risk Assessment Service to Railway with ONNX Runtime support and ML model capabilities.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Railway Setup](#railway-setup)
3. [Environment Configuration](#environment-configuration)
4. [Model Deployment](#model-deployment)
5. [Deployment Process](#deployment-process)
6. [Validation and Testing](#validation-and-testing)
7. [Monitoring and Maintenance](#monitoring-and-maintenance)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

- **Railway CLI**: Install from [railway.app](https://railway.app)
- **Docker**: For local testing and model preparation
- **Python 3.12+**: For ML model training and ONNX conversion
- **Go 1.22+**: For building the service
- **Git**: For version control

### Required Accounts

- Railway account with billing enabled (for custom domains and scaling)
- GitHub account (for repository access)

### System Requirements

- **Memory**: Minimum 2GB RAM (recommended 4GB+)
- **Storage**: 1GB for models and dependencies
- **CPU**: 2+ cores recommended for ML inference

## Railway Setup

### 1. Install Railway CLI

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login
```

### 2. Create New Project

```bash
# Navigate to service directory
cd services/risk-assessment-service

# Initialize Railway project
railway init

# Link to existing project (if applicable)
railway link [PROJECT_ID]
```

### 3. Configure Railway Settings

```bash
# Set project name
railway variables set PROJECT_NAME="risk-assessment-service"

# Set environment
railway variables set ENVIRONMENT="production"
```

## Environment Configuration

### Required Environment Variables

Create a `.env` file or set Railway variables:

```bash
# Service Configuration
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info

# Database Configuration
DATABASE_URL=postgresql://user:password@host:port/database
REDIS_URL=redis://host:port

# ML Model Configuration
LSTM_MODEL_PATH=/app/models/risk_lstm_v1.onnx
XGBOOST_MODEL_PATH=/app/models/xgb_model.json
ENABLE_ENSEMBLE=true
DEFAULT_PREDICTION_HORIZON=3

# ONNX Runtime Configuration
LD_LIBRARY_PATH=/app/onnxruntime/lib
CGO_ENABLED=1
CGO_LDFLAGS=-L/app/onnxruntime/lib
CGO_CFLAGS=-I/app/onnxruntime/include

# Security
JWT_SECRET=your-jwt-secret-key
API_KEY=your-api-key

# External Services
GOVDATA_API_KEY=your-govdata-api-key
CREDITBUREAU_API_KEY=your-creditbureau-api-key
```

### Set Railway Variables

```bash
# Set all environment variables
railway variables set PORT=8080
railway variables set ENVIRONMENT=production
railway variables set LOG_LEVEL=info
railway variables set ENABLE_ENSEMBLE=true
railway variables set DEFAULT_PREDICTION_HORIZON=3

# Set security variables (use Railway secrets)
railway variables set JWT_SECRET=your-jwt-secret-key --secret
railway variables set API_KEY=your-api-key --secret
```

## Model Deployment

### 1. Prepare ML Models

The service requires pre-trained models in ONNX format:

```bash
# Train and export models (if not already done)
cd ml-training
python train_lstm_model.py
python train_xgboost_model.py
python export_models_to_onnx.py
```

### 2. Model Files Required

Ensure these files are present in the `models/` directory:

```
models/
├── risk_lstm_v1.onnx          # LSTM model for time-series prediction
├── xgb_model.json             # XGBoost model for risk assessment
├── model_metadata.json        # Model metadata and configuration
└── feature_scaler.pkl         # Feature scaling parameters
```

### 3. Upload Models to Railway

```bash
# Copy models to Railway (models will be included in Docker build)
railway up --detach
```

## Deployment Process

### 1. Automated Deployment

Use the provided deployment script:

```bash
# Run deployment script
./scripts/deploy-railway.sh

# Or with custom configuration
./scripts/deploy-railway.sh --environment production --scale 2
```

### 2. Manual Deployment

```bash
# Build and deploy
railway up

# Deploy with specific configuration
railway up --detach --scale 2
```

### 3. Custom Domain Setup

```bash
# Add custom domain
railway domain add your-domain.com

# Configure SSL (automatic with Railway)
railway domain ssl enable your-domain.com
```

## Validation and Testing

### 1. Automated Validation

Run the comprehensive validation script:

```bash
# Validate deployment
./scripts/validate-deployment.sh

# Validate with specific URL
./scripts/validate-deployment.sh https://your-service.railway.app
```

### 2. Manual Testing

Test key endpoints:

```bash
# Health check
curl https://your-service.railway.app/health

# Risk assessment
curl -X POST https://your-service.railway.app/api/v1/assess \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "business_address": "123 Test St, Test City, TC 12345",
    "industry": "technology",
    "country": "US",
    "prediction_horizon": 3
  }'

# Advanced prediction
curl -X POST https://your-service.railway.app/api/v1/risk/predict-advanced \
  -H "Content-Type: application/json" \
  -d '{
    "business": {
      "business_name": "Test Company",
      "business_address": "123 Test St, Test City, TC 12345",
      "industry": "technology",
      "country": "US"
    },
    "prediction_horizons": [3, 6, 12],
    "model_preference": "auto"
  }'
```

### 3. Performance Testing

```bash
# Run performance benchmarks
go run cmd/benchmark_lstm.go

# Load testing with Apache Bench
ab -n 1000 -c 10 https://your-service.railway.app/api/v1/assess
```

## Monitoring and Maintenance

### 1. Railway Dashboard

Monitor your service through the Railway dashboard:

- **Metrics**: CPU, memory, network usage
- **Logs**: Real-time application logs
- **Deployments**: Deployment history and status
- **Variables**: Environment configuration

### 2. Application Metrics

Access service metrics:

```bash
# Get performance metrics
curl https://your-service.railway.app/metrics

# Get model performance
curl https://your-service.railway.app/api/v1/models/performance
```

### 3. Health Monitoring

Set up health checks:

```bash
# Railway health check (automatic)
railway health check

# Custom health monitoring
curl https://your-service.railway.app/health
```

### 4. Log Monitoring

```bash
# View real-time logs
railway logs

# View logs with filtering
railway logs --filter "ERROR"
```

## Troubleshooting

### Common Issues

#### 1. ONNX Runtime Issues

**Problem**: ONNX Runtime library not found
```
Error: could not load ONNX Runtime library
```

**Solution**:
```bash
# Check ONNX Runtime installation
ls -la /app/onnxruntime/lib/

# Verify environment variables
echo $LD_LIBRARY_PATH
echo $CGO_LDFLAGS
```

#### 2. Model Loading Issues

**Problem**: Models not found or corrupted
```
Error: model file not found: /app/models/risk_lstm_v1.onnx
```

**Solution**:
```bash
# Check model files
ls -la /app/models/

# Verify model integrity
file /app/models/risk_lstm_v1.onnx
```

#### 3. Memory Issues

**Problem**: Out of memory errors
```
Error: runtime: out of memory
```

**Solution**:
```bash
# Increase Railway memory allocation
railway variables set RAILWAY_MEMORY_LIMIT=4GB

# Scale service
railway scale 2
```

#### 4. Performance Issues

**Problem**: High latency or timeouts

**Solution**:
```bash
# Check service metrics
curl https://your-service.railway.app/metrics

# Scale horizontally
railway scale 3

# Check database connections
railway variables set DATABASE_POOL_SIZE=20
```

### Debugging Commands

```bash
# Check service status
railway status

# View deployment logs
railway logs --deployment [DEPLOYMENT_ID]

# Connect to service shell
railway shell

# Check environment variables
railway variables

# Restart service
railway restart
```

### Performance Optimization

#### 1. Database Optimization

```bash
# Set connection pool size
railway variables set DATABASE_POOL_SIZE=20
railway variables set DATABASE_MAX_IDLE_CONNS=10
railway variables set DATABASE_MAX_OPEN_CONNS=25
```

#### 2. Caching Configuration

```bash
# Enable Redis caching
railway variables set ENABLE_CACHE=true
railway variables set CACHE_TTL=3600
```

#### 3. ML Model Optimization

```bash
# Set model cache size
railway variables set MODEL_CACHE_SIZE=100
railway variables set ENABLE_MODEL_CACHING=true
```

## Security Considerations

### 1. API Security

- Use HTTPS for all communications
- Implement API key authentication
- Set up rate limiting
- Validate all input data

### 2. Data Protection

- Encrypt sensitive data in transit and at rest
- Use secure environment variables for secrets
- Implement data retention policies
- Regular security audits

### 3. Access Control

- Use Railway's built-in access controls
- Implement proper IAM policies
- Monitor access logs
- Regular access reviews

## Scaling and Performance

### 1. Horizontal Scaling

```bash
# Scale to multiple instances
railway scale 3

# Auto-scaling configuration
railway variables set AUTO_SCALE_MIN=2
railway variables set AUTO_SCALE_MAX=10
```

### 2. Resource Allocation

```bash
# Increase memory allocation
railway variables set RAILWAY_MEMORY_LIMIT=4GB

# Increase CPU allocation
railway variables set RAILWAY_CPU_LIMIT=2
```

### 3. Database Scaling

- Use Railway's managed PostgreSQL
- Configure read replicas for read-heavy workloads
- Implement connection pooling
- Monitor database performance

## Backup and Recovery

### 1. Database Backups

Railway provides automatic database backups:

```bash
# Check backup status
railway database backup list

# Create manual backup
railway database backup create
```

### 2. Model Backups

```bash
# Backup model files
tar -czf models-backup.tar.gz models/

# Store in Railway volumes
railway volume create model-backups
```

### 3. Configuration Backups

```bash
# Export environment variables
railway variables > environment-backup.txt

# Export Railway configuration
railway config > railway-config.json
```

## Cost Optimization

### 1. Resource Monitoring

- Monitor CPU and memory usage
- Set up cost alerts
- Review usage patterns
- Optimize resource allocation

### 2. Scaling Strategies

- Use auto-scaling for variable workloads
- Implement efficient caching
- Optimize database queries
- Use CDN for static assets

### 3. Cost Tracking

```bash
# Check current usage
railway usage

# Set spending limits
railway billing set-limit 100
```

## Support and Resources

### 1. Documentation

- [Railway Documentation](https://docs.railway.app)
- [Go ONNX Runtime](https://github.com/yalue/onnxruntime_go)
- [Service API Documentation](./API.md)

### 2. Community Support

- Railway Discord community
- GitHub Issues for bug reports
- Stack Overflow for technical questions

### 3. Professional Support

- Railway Pro support
- Custom development services
- Performance optimization consulting

---

**Last Updated**: December 2024  
**Version**: 1.0.0  
**Maintainer**: KYB Platform Team
