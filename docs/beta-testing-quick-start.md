# Beta Testing Quick Start Guide

## 🚀 Getting Started with KYB Platform Beta Testing

This guide will help you quickly set up and start beta testing the KYB Platform MVP, a comprehensive enterprise-grade Know Your Business platform.

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
- **Web Interface**: http://localhost:8080 (User-friendly dashboard)
- **API Endpoint**: http://localhost:8081
- **API Documentation**: http://localhost:8081/docs (Interactive Swagger docs)
- **Monitoring Dashboard**: http://localhost:3000 (Grafana)
- **Metrics**: http://localhost:9090 (Prometheus)

### 3. Test the Complete Platform

#### Option A: Web Interface (Recommended for Non-Technical Users)
1. **Open your browser** and navigate to: http://localhost:8080
2. **Register/Login** using the web interface
3. **Use the dashboard** to:
   - Classify businesses using the simple form
   - View risk assessments and compliance status
   - Generate reports and export data
   - Manage your account and preferences

#### Option B: API Integration (For Technical Users)

```bash
# Test authentication
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Test business classification
curl -X POST http://localhost:8081/api/v1/classify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "business_name": "Acme Corporation",
    "website_url": "https://www.acme.com",
    "description": "Global manufacturing and technology company"
  }'

# Test risk assessment
curl -X POST http://localhost:8081/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "business_id": "business_123",
    "risk_factors": ["industry_risk", "financial_risk", "compliance_risk"]
  }'

# Test compliance checking
curl -X POST http://localhost:8081/api/v1/compliance/check \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "business_id": "business_123",
    "compliance_frameworks": ["SOC2", "PCI_DSS", "GDPR"]
  }'
```

Expected classification response:
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
      "method_used": "multi_method",
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

### 1. Business Classification Engine

#### Multi-Method Classification
- Keyword, business type, industry, and name-based classification
- NAICS code mapping with comprehensive industry names
- Confidence scoring for all classification methods
- Batch processing for multiple businesses
- Result caching for performance optimization

#### Classification Results
- Primary industry with highest confidence
- Secondary industries for comprehensive coverage
- Confidence scoring for each classification
- Real-time accuracy assessment and validation

### 2. Risk Assessment Engine

#### Multi-Factor Risk Scoring
- Comprehensive risk calculation algorithms
- Industry-specific risk models
- Risk trend analysis and prediction
- Automated risk alerts and monitoring
- Detailed risk assessment reports

### 3. Compliance Framework

#### Regulatory Compliance
- SOC 2, PCI DSS, GDPR compliance tracking
- Automated compliance requirement checking
- Compliance gap analysis and recommendations
- Complete audit trail generation
- Automated compliance report generation

### 4. Authentication & Authorization System

#### JWT-based Authentication
- Secure token-based authentication
- Role-Based Access Control (RBAC)
- API key management for integrations
- Complete user lifecycle management
- Security hardening with rate limiting and audit logging

### 5. API Gateway & Ecosystem

#### Complete API Ecosystem
- RESTful API with versioning
- Comprehensive middleware stack
- Health monitoring and metrics
- Interactive API documentation
- Consistent error handling

### 6. Web User Interface

#### User-Friendly Dashboard
- Login/Registration: Secure user authentication
- Business Classification Form: Simple form for business information input
- Results Dashboard: Visual display of classification results
- Risk Assessment View: Interactive risk scoring and analysis
- Compliance Status: Real-time compliance tracking and reporting
- User Profile: Account management and preferences
- Help & Support: Built-in documentation and support chat

### 3. API Endpoints

#### Authentication API
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
GET /api/v1/auth/profile
```

#### Classification API
```
POST /api/v1/classify
POST /api/v1/classify/batch
GET /api/v1/classify/{id}
GET /api/v1/classify/history
```

#### Risk Assessment API
```
POST /api/v1/risk/assess
GET /api/v1/risk/{id}
GET /api/v1/risk/history
POST /api/v1/risk/alerts
GET /api/v1/risk/reports
```

#### Compliance API
```
POST /api/v1/compliance/check
GET /api/v1/compliance/{id}
GET /api/v1/compliance/status
POST /api/v1/compliance/reports
GET /api/v1/compliance/audit
```

#### Health and Monitoring
```
GET /health
GET /metrics
GET /api/v1/status
GET /api/v1/docs
```

## Interface Options for Beta Testing

### Dual Interface Approach

The beta testing program provides two interface options to accommodate different user types:

#### Web Interface (Recommended for Non-Technical Users)
- **Access**: http://localhost:8080
- **No Technical Knowledge Required**: Point-and-click interface
- **Guided Workflows**: Step-by-step processes
- **Visual Reports**: Interactive charts and graphs
- **User Management**: Account creation and profile management

#### API Integration (For Technical Users)
- **Access**: http://localhost:8081
- **Documentation**: http://localhost:8081/docs
- **Programmatic Testing**: Direct API calls
- **SDK Support**: Client libraries available
- **Webhook Support**: Real-time notifications

### Choosing Your Interface

- **Business Users** (Compliance Officers, Risk Managers): Use Web Interface
- **Developers** (Integration Specialists): Use API Integration
- **Analysts** (Business Analysts, Researchers): Use Both Interfaces

## Testing Scenarios

### Scenario 1: Financial Institution End-to-End Testing

**Use Case**: Complete KYB workflow for loan applications
**Recommended Interface**: Web Interface (Primary), API Integration (Secondary)

**Test Cases**:
1. **Authentication**: User registration, login, and role assignment via web interface
2. **Business Classification**: Classify 50+ businesses using web form and API
3. **Risk Assessment**: Generate comprehensive risk scores and reports via dashboard
4. **Compliance Checking**: Verify regulatory compliance requirements through web interface
5. **Integration**: Test complete workflow from classification to decision using both interfaces

**Sample Test Workflow**:
```json
{
  "workflow": {
    "step1": "User authentication and authorization",
    "step2": "Business classification with confidence scoring",
    "step3": "Risk assessment with industry-specific models",
    "step4": "Compliance verification and gap analysis",
    "step5": "Report generation and decision support"
  }
}
```

### Scenario 2: Risk Management & Assessment Testing

**Use Case**: Comprehensive risk assessment and monitoring

**Test Cases**:
1. **Risk Scoring**: Test multi-factor risk calculation algorithms
2. **Industry Models**: Validate industry-specific risk assessments
3. **Trend Analysis**: Analyze historical risk trends and predictions
4. **Alerting**: Test automated risk alerts and notifications
5. **Reporting**: Generate detailed risk assessment reports

### Scenario 3: Compliance & Regulatory Testing

**Use Case**: Regulatory compliance and audit preparation

**Test Cases**:
1. **Compliance Framework**: Test SOC 2, PCI DSS, GDPR compliance tracking
2. **Audit Trails**: Verify complete audit logging and reporting
3. **Gap Analysis**: Identify compliance gaps and generate recommendations
4. **Reporting**: Generate compliance reports for regulatory submissions
5. **Monitoring**: Set up compliance alerts and monitoring

### Scenario 4: API Integration & Development Testing

**Use Case**: Third-party system integration

**Test Cases**:
1. **API Authentication**: Test JWT tokens and API key management
2. **Rate Limiting**: Verify fair usage policies and limits
3. **Batch Processing**: Test efficient processing of multiple requests
4. **Error Handling**: Validate consistent error responses
5. **Performance**: Test API response times and throughput

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
