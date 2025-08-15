# KYB Platform Beta Testing Guide

## 🎯 Overview

This guide provides comprehensive instructions for beta testing the enhanced KYB Platform classification service. The platform now includes advanced features like website analysis, ML model integration, geographic awareness, and real-time feedback collection.

## 🚀 Quick Start

### 1. Automated Setup

Run the automated setup script:

```bash
./scripts/beta-testing-setup.sh setup
```

This will:
- Check dependencies
- Set up environment variables
- Build the application
- Run database migrations
- Execute tests
- Build Docker image
- Generate a beta testing report

### 2. Local Development

Start the local development environment:

```bash
./scripts/beta-testing-setup.sh local
```

### 3. Deploy to Railway

Deploy to Railway for cloud testing:

```bash
./scripts/beta-testing-setup.sh deploy
```

## 📋 Prerequisites

### Required Tools
- **Go 1.22+**: [Download](https://golang.org/dl/)
- **Docker**: [Download](https://docker.com/)
- **Docker Compose**: [Download](https://docs.docker.com/compose/install/)

### Optional Tools
- **Railway CLI**: `npm install -g @railway/cli`
- **Supabase CLI**: `npm install -g supabase`

### Environment Setup
1. Copy `env.example` to `.env`
2. Configure your Supabase credentials
3. Set up Railway project (if using cloud deployment)

## 🔧 Manual Setup

### 1. Environment Configuration

```bash
# Copy environment template
cp env.example .env

# Generate secure secrets
JWT_SECRET=$(openssl rand -hex 32)
ENCRYPTION_KEY=$(openssl rand -hex 32)

# Update .env file with your values
sed -i "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env
sed -i "s/ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$ENCRYPTION_KEY/" .env
```

### 2. Database Setup

```bash
# Start Supabase locally
supabase start

# Run migrations
supabase db reset
```

### 3. Build and Test

```bash
# Build the application
go build -o kyb-platform ./cmd/api

# Run tests
go test ./... -v

# Build Docker image
docker build -f Dockerfile.beta -t kyb-platform:beta .
```

## 🧪 Testing the Enhanced Classification Service

### API Endpoints

#### 1. Single Business Classification

```bash
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "business_type": "Corporation",
    "industry": "Technology",
    "description": "Software development company",
    "geographic_region": "us",
    "enhanced_metadata": {
      "website": "https://acme.com",
      "registration_number": "123456789"
    }
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "business_id": "uuid",
  "classifications": [...],
  "primary_industry": {
    "industry": "Software Development",
    "confidence": 0.92,
    "method": "website_analysis"
  },
  "overall_confidence": 0.92,
  "classification_method": "website_analysis",
  "processing_time": "1.2s",
  "geographic_region": "us",
  "region_confidence_score": 0.92,
  "api_version": "v1"
}
```

#### 2. Batch Classification

```bash
curl -X POST http://localhost:8080/v1/classify/batch \
  -H "Content-Type: application/json" \
  -d '{
    "businesses": [
      {
        "business_name": "TechCorp Inc",
        "industry": "Technology",
        "geographic_region": "us"
      },
      {
        "business_name": "Global Retail Ltd",
        "industry": "Retail",
        "geographic_region": "uk"
      }
    ],
    "geographic_region": "global"
  }'
```

#### 3. Get Classification by ID

```bash
curl -X GET http://localhost:8080/v1/classify/{business_id}
```

### Enhanced Features Testing

#### 1. Geographic Region Support

Test different geographic regions:

```bash
# US business
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "American Tech",
    "geographic_region": "us"
  }'

# UK business
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "British Retail",
    "geographic_region": "uk"
  }'

# Japanese business
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Japanese Manufacturing",
    "geographic_region": "jp"
  }'
```

#### 2. Industry-Specific Testing

Test high-code-density industries:

```bash
# Agriculture
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Green Acres Farm",
    "industry": "Agriculture",
    "description": "Organic farming and crop production"
  }'

# Retail
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "MegaMart Retail",
    "industry": "Retail",
    "description": "Department store and consumer goods"
  }'

# Food & Beverage
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Fresh Foods Co",
    "industry": "Food & Beverage",
    "description": "Restaurant chain and food service"
  }'
```

#### 3. Confidence Score Testing

Verify confidence score ranges:

- **Website Analysis**: 0.85-0.95
- **Web Search**: 0.75-0.85
- **Keyword-based**: 0.60-0.75
- **Fuzzy Matching**: 0.50-0.70
- **Crosswalk Mapping**: 0.40-0.60

#### 4. Performance Testing

Test response times:

```bash
# Test single classification performance
time curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company"}'

# Test batch classification performance
time curl -X POST http://localhost:8080/v1/classify/batch \
  -H "Content-Type: application/json" \
  -d '{"businesses": [{"business_name": "Test1"}, {"business_name": "Test2"}]}'
```

### Monitoring and Observability

#### 1. Health Check

```bash
curl http://localhost:8080/health
```

#### 2. Metrics Endpoint

```bash
curl http://localhost:8080/v1/metrics
```

#### 3. API Status

```bash
curl http://localhost:8080/v1/status
```

## 📊 Test Scenarios

### Scenario 1: Basic Classification
**Objective**: Verify basic classification functionality

**Steps**:
1. Submit a simple business classification request
2. Verify response contains required fields
3. Check confidence score is within expected range
4. Validate processing time is under 2 seconds

**Expected Results**:
- Success: true
- Overall confidence: > 0.6
- Processing time: < 2s
- Valid business_id returned

### Scenario 2: Geographic Region Testing
**Objective**: Test geographic region awareness

**Steps**:
1. Submit classification with different geographic regions
2. Compare confidence scores across regions
3. Verify region_confidence_score is calculated correctly

**Expected Results**:
- US region: No confidence reduction
- Other regions: Appropriate confidence reduction
- Region-specific patterns applied

### Scenario 3: Industry-Specific Testing
**Objective**: Test industry-specific improvements

**Steps**:
1. Test high-code-density industries (Agriculture, Retail, Food & Beverage)
2. Verify industry-specific mappings are applied
3. Check for improved accuracy in these industries

**Expected Results**:
- Higher accuracy for industry-specific classifications
- Industry-specific data in response
- Appropriate confidence scoring

### Scenario 4: Performance Testing
**Objective**: Verify performance requirements

**Steps**:
1. Submit multiple concurrent requests
2. Test batch classification with various sizes
3. Monitor response times and resource usage

**Expected Results**:
- 95% of requests complete within 2 seconds
- Batch processing handles up to 100 businesses
- System remains stable under load

### Scenario 5: Error Handling
**Objective**: Test error scenarios

**Steps**:
1. Submit invalid requests (missing required fields)
2. Test with malformed JSON
3. Submit requests with invalid geographic regions

**Expected Results**:
- Appropriate error responses
- Clear error messages
- Proper HTTP status codes

## 🔍 Monitoring During Testing

### Key Metrics to Monitor

1. **Classification Accuracy**
   - Overall accuracy should be > 90%
   - Industry-specific accuracy > 85%

2. **Performance Metrics**
   - Response time < 2 seconds for 95% of requests
   - Throughput: 1000+ concurrent requests

3. **System Health**
   - CPU usage < 80%
   - Memory usage < 80%
   - Error rate < 5%

### Monitoring Commands

```bash
# Check application logs
docker logs kyb-platform

# Monitor system resources
docker stats kyb-platform

# Check database performance
supabase db logs
```

## 🐛 Troubleshooting

### Common Issues

#### 1. Build Failures
```bash
# Clean and rebuild
go clean -cache
go mod tidy
go build ./cmd/api
```

#### 2. Database Connection Issues
```bash
# Check Supabase status
supabase status

# Restart Supabase
supabase stop
supabase start
```

#### 3. Docker Issues
```bash
# Clean Docker
docker system prune -a

# Rebuild image
docker build -f Dockerfile.beta -t kyb-platform:beta .
```

#### 4. Railway Deployment Issues
```bash
# Check Railway status
railway status

# View logs
railway logs

# Redeploy
railway up
```

### Getting Help

1. **Check the logs**: Application logs contain detailed error information
2. **Review metrics**: The metrics endpoint provides system health data
3. **Contact support**: For persistent issues, contact the development team

## 📈 Success Criteria

### Functional Requirements
- ✅ All classification endpoints respond correctly
- ✅ Geographic region support works as expected
- ✅ Industry-specific improvements are applied
- ✅ Confidence scoring follows specified ranges
- ✅ Error handling is robust

### Performance Requirements
- ✅ Response time < 2 seconds for 95% of requests
- ✅ System handles 1000+ concurrent requests
- ✅ Memory and CPU usage remain within limits
- ✅ Database performance is acceptable

### Quality Requirements
- ✅ Classification accuracy > 90%
- ✅ Test coverage > 90%
- ✅ All tests pass
- ✅ No critical security vulnerabilities

## 📝 Reporting

### Beta Testing Report Template

After completing testing, generate a report including:

1. **Test Summary**
   - Number of test cases executed
   - Pass/fail statistics
   - Performance metrics

2. **Issues Found**
   - Bug descriptions
   - Severity levels
   - Steps to reproduce

3. **Feedback**
   - User experience feedback
   - Feature suggestions
   - Performance observations

4. **Recommendations**
   - Areas for improvement
   - Priority fixes
   - Future enhancements

### Generate Report

```bash
./scripts/beta-testing-setup.sh
```

This will generate a comprehensive beta testing report in `beta-testing-report.md`.

## 🎉 Next Steps

After successful beta testing:

1. **Deploy to Production**: Use the validated deployment process
2. **Monitor Performance**: Set up production monitoring
3. **Collect Feedback**: Implement user feedback collection
4. **Iterate**: Use feedback to improve the system

---

**Happy Testing! 🚀**

For questions or support, please contact the development team.
