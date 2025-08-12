# Beta Testing Quick Start Guide

## 🚀 Getting Started with KYB Platform Beta Testing

This guide will help you quickly set up and start beta testing the KYB Platform's website classification MVP.

## Prerequisites

- Docker installed and running
- Go 1.22+ installed
- Git access to the repository
- Basic understanding of API testing

## Quick Setup (5 minutes)

### 1. Set Up Beta Environment

```bash
# Run the beta environment setup script
./scripts/setup-beta-environment.sh
```

This script will:
- Create beta environment configuration
- Set up beta database
- Configure monitoring and analytics
- Create feedback collection system
- Set up beta user management

### 2. Start the Beta Environment

```bash
# Start the beta environment
./scripts/dev.sh beta
```

The beta environment will be available at:
- **API Endpoint**: http://localhost:8081
- **Monitoring Dashboard**: http://localhost:3000 (Grafana)
- **Metrics**: http://localhost:9090 (Prometheus)

### 3. Test the Classification API

```bash
# Test a simple classification
curl -X POST http://localhost:8081/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "website_url": "https://www.acme.com",
    "description": "Global manufacturing and technology company"
  }'
```

Expected response:
```json
{
  "success": true,
  "data": {
    "primary_industry": {
      "industry_code": "332996",
      "industry_name": "Manufacturing",
      "confidence_score": 0.92
    },
    "secondary_industries": [
      {
        "industry_code": "541330",
        "industry_name": "Engineering Services",
        "confidence_score": 0.85
      },
      {
        "industry_code": "541511",
        "industry_name": "Custom Computer Programming Services",
        "confidence_score": 0.78
      }
    ],
    "classification_metadata": {
      "method_used": "url_based",
      "processing_time_ms": 1200,
      "accuracy_validation": {
        "is_accurate": true,
        "accuracy_score": 0.89
      }
    }
  }
}
```

## Beta Testing Features

### 1. Website Classification

#### URL-Based Classification
- Direct website analysis and scraping
- Industry classification from website content
- Business information extraction

#### Web Search Classification
- Business name search when no URL provided
- Search result analysis and classification
- Fallback mechanism for unknown businesses

#### Dual Flow Selection
- Automatic routing between URL and search-based flows
- Intelligent fallback mechanisms
- Unified result format

### 2. Classification Results

#### Top-3 Industry Results
- Primary industry with highest confidence
- Secondary industries for comprehensive coverage
- Confidence scoring for each classification

#### Accuracy Validation
- Real-time accuracy assessment
- Benchmark comparison
- Quality indicators

#### Result Presentation
- User-friendly output format
- Detailed confidence breakdown
- Processing metadata

### 3. API Endpoints

#### Classification API
```
POST /api/v1/classify
POST /api/v1/classify/batch
GET /api/v1/classify/{id}
```

#### Accuracy Validation API
```
POST /api/v1/accuracy/validate
GET /api/v1/accuracy/metrics
POST /api/v1/accuracy/feedback
```

#### Health and Monitoring
```
GET /health
GET /metrics
GET /api/v1/status
```

## Testing Scenarios

### Scenario 1: Financial Institution Testing

**Use Case**: Verify business classification for loan applications

**Test Cases**:
1. Classify 50+ business websites
2. Validate industry accuracy against known data
3. Test batch processing capabilities
4. Assess integration with existing systems

**Sample Test Data**:
```json
{
  "business_name": "Financial Services Inc",
  "website_url": "https://financialservices.com",
  "description": "Investment banking and wealth management"
}
```

### Scenario 2: Risk Assessment Testing

**Use Case**: Assess business risk based on industry classification

**Test Cases**:
1. Classify high-risk industries
2. Validate risk scoring accuracy
3. Test real-time classification
4. Assess reporting capabilities

### Scenario 3: Business Research Testing

**Use Case**: Market research and competitive analysis

**Test Cases**:
1. Classify competitors and market players
2. Validate classification consistency
3. Test data export functionality
4. Assess API integration

## Performance Testing

### Automated Performance Tests

```bash
# Run performance tests
./scripts/run-beta-tests.sh
```

This will test:
- Response time under load
- Throughput capabilities
- Error rates
- Resource utilization

### Manual Performance Testing

```bash
# Test with Apache Bench
ab -n 100 -c 10 -p test/beta/data/test_businesses.json \
  -T application/json \
  http://localhost:8081/api/v1/classify
```

## Feedback Collection

### In-App Feedback

The beta environment includes built-in feedback collection:
- Feature usage tracking
- Satisfaction ratings
- Error reporting
- Performance monitoring

### Survey Collection

Pre-built surveys are available:
- **Onboarding Survey**: `test/beta/feedback-surveys/onboarding-survey.json`
- **Feature Usage Survey**: `test/beta/feedback-surveys/feature-usage-survey.json`
- **Overall Experience Survey**: `test/beta/feedback-surveys/overall-experience-survey.json`

### Feedback Submission

```bash
# Submit feedback via API
curl -X POST http://localhost:8081/api/v1/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "beta_user_001",
    "feature": "website_classification",
    "rating": 9,
    "comments": "Excellent accuracy and fast response times",
    "category": "feature_satisfaction"
  }'
```

## Monitoring and Analytics

### Real-Time Monitoring

Access the monitoring dashboard at http://localhost:3000:
- **Classification Accuracy**: Real-time accuracy metrics
- **Performance Metrics**: Response times and throughput
- **User Activity**: Beta user engagement
- **Error Tracking**: System errors and issues

### Key Metrics to Monitor

1. **Classification Accuracy**: Target >90%
2. **Response Time**: Target <5 seconds
3. **Error Rate**: Target <5%
4. **User Satisfaction**: Target >8/10
5. **Feature Adoption**: Target >70%

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues
```bash
# Check database status
docker ps | grep kyb_beta_db

# Restart database if needed
docker restart kyb_beta_db
```

#### 2. API Not Responding
```bash
# Check API health
curl http://localhost:8081/health

# Check logs
docker logs kyb_beta_api
```

#### 3. Classification Accuracy Issues
```bash
# Run accuracy validation
curl -X POST http://localhost:8081/api/v1/accuracy/validate \
  -H "Content-Type: application/json" \
  -d @test/beta/data/test_businesses.json
```

### Getting Help

1. **Check the logs**: `docker logs kyb_beta_api`
2. **Monitor metrics**: http://localhost:9090
3. **Review documentation**: `docs/beta-testing-plan.md`
4. **Submit feedback**: Use the feedback API or surveys

## Next Steps

### Week 1: Initial Testing
- [ ] Complete setup and configuration
- [ ] Test basic classification functionality
- [ ] Validate accuracy with test data
- [ ] Submit initial feedback

### Week 2: Extended Testing
- [ ] Test with real business data
- [ ] Evaluate performance under load
- [ ] Assess user experience
- [ ] Identify feature gaps

### Week 3: Integration Testing
- [ ] Test API integration
- [ ] Validate batch processing
- [ ] Assess error handling
- [ ] Review monitoring data

### Week 4: Feedback and Planning
- [ ] Complete feedback surveys
- [ ] Analyze performance data
- [ ] Identify improvement areas
- [ ] Plan next development phase

## Success Criteria

### Quantitative Metrics
- **Classification Accuracy**: >90% on real business data
- **User Satisfaction**: >8/10 average rating
- **Feature Adoption**: >70% active usage
- **Performance**: <5 second average response time
- **Retention**: >80% user return rate

### Qualitative Metrics
- **User Feedback**: Positive sentiment in qualitative responses
- **Feature Requests**: Alignment with product roadmap
- **Competitive Positioning**: Favorable comparison to alternatives
- **Market Validation**: Confirmed product-market fit

## Contact and Support

For beta testing support:
- **Documentation**: Check `docs/` directory
- **Issues**: Use the feedback API or create GitHub issues
- **Questions**: Review the beta testing plan in `docs/beta-testing-plan.md`

---

**Happy Beta Testing! 🎉**

The KYB Platform team is excited to get your feedback and make the website classification MVP even better.
