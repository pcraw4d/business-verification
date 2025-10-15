# Frequently Asked Questions (FAQ)

## General Questions

### What is the Risk Assessment Service?

The Risk Assessment Service is a comprehensive API that provides real-time business risk assessment, compliance screening, and predictive analytics. It uses advanced machine learning models to analyze business data and predict future risk trends with sub-1-second response times.

### What makes this service different from competitors?

Our service offers several unique advantages:
- **Predictive Analytics**: 3-12 month risk forecasting (unique in market)
- **Sub-1-Second Response**: 5x faster than competitors
- **Developer-First Experience**: Best-in-class API and SDK experience
- **Real-time Processing**: Live risk monitoring vs. batch processing
- **Modular Pricing**: Pay-for-what-you-use model
- **Global Coverage**: Multi-country support from day one

### What industries do you support?

We support 9 specialized industry models:
- **FinTech**: Payment processing, digital banking, cryptocurrency
- **Healthcare**: Medical devices, pharmaceuticals, telemedicine
- **Technology**: Software, hardware, cloud services
- **E-commerce**: Online retail, marketplaces, logistics
- **Real Estate**: Property management, construction, development
- **Manufacturing**: Industrial, automotive, aerospace
- **Energy**: Oil & gas, renewable energy, utilities
- **Transportation**: Logistics, shipping, ride-sharing
- **Professional Services**: Consulting, legal, accounting

### What countries are supported?

We currently support risk assessment for businesses in:
- **North America**: United States, Canada, Mexico
- **Europe**: United Kingdom, Germany, France, Netherlands, Sweden, Norway
- **Asia-Pacific**: Australia, Singapore, Japan
- **Additional countries**: Contact support for availability

## API Usage

### How do I get started with the API?

1. **Sign up** for a KYB Platform account
2. **Get your API key** from the dashboard
3. **Read the [Quick Start Guide](API_QUICK_START.md)**
4. **Make your first API call** using our examples
5. **Explore advanced features** like predictions and compliance

### What is the rate limit?

- **Test Environment**: 100 requests per minute
- **Production Environment**: Based on your plan
  - **Starter**: 100 requests/minute
  - **Professional**: 500 requests/minute
  - **Enterprise**: Unlimited

Rate limit information is included in response headers:
- `X-RateLimit-Limit`: Maximum requests per minute
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the rate limit resets

### How do I handle rate limits?

Implement exponential backoff and request queuing:

```javascript
async function makeAPICallWithRetry(url, options, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const response = await fetch(url, options);
      
      if (response.status === 429) {
        const waitTime = Math.pow(2, attempt) * 1000;
        console.log(`Rate limited. Waiting ${waitTime}ms before retry ${attempt + 1}`);
        await new Promise(resolve => setTimeout(resolve, waitTime));
        continue;
      }
      
      return response;
    } catch (error) {
      if (attempt === maxRetries - 1) throw error;
    }
  }
}
```

### What data do I need to provide for a risk assessment?

**Required fields:**
- `business_name`: Company name
- `business_address`: Full business address
- `industry`: Industry category
- `country`: Country code (ISO 3166-1 alpha-2)

**Optional fields:**
- `phone`: Business phone number
- `email`: Business email address
- `website`: Company website
- `metadata`: Additional business information

### How accurate are the risk predictions?

Our ML models achieve high accuracy:
- **XGBoost Model**: 92% accuracy for 1-3 month predictions
- **LSTM Model**: 89% accuracy for 6-12 month predictions
- **Ensemble Model**: 94% accuracy for 3-6 month predictions

Accuracy varies by industry and data quality. We provide confidence scores with each prediction.

### What is the difference between risk score and risk level?

- **Risk Score**: Numerical value (0.0 - 1.0) indicating risk probability
- **Risk Level**: Categorical classification (low, medium, high, critical)

**Risk Level Mapping:**
- **Low**: 0.0 - 0.3 (Green)
- **Medium**: 0.3 - 0.6 (Yellow)
- **High**: 0.6 - 0.8 (Orange)
- **Critical**: 0.8 - 1.0 (Red)

## Authentication & Security

### How do I authenticate with the API?

Use your API key in the `Authorization` header:

```bash
Authorization: Bearer YOUR_API_KEY
```

**Example:**
```bash
curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/assess/risk_123" \
  -H "Authorization: Bearer sk_live_1234567890abcdef"
```

### How do I get an API key?

1. Log into your [KYB Platform Dashboard](https://dashboard.kyb-platform.com)
2. Navigate to **Settings** → **API Keys**
3. Click **Create New API Key**
4. Copy the generated key (you won't see it again!)

### What's the difference between test and live API keys?

- **Test Keys** (`sk_test_`): Use synthetic data, perfect for development
- **Live Keys** (`sk_live_`): Use real business data, for production use

### How do I secure my API key?

**Best Practices:**
- Store in environment variables, not source code
- Use secure key management services
- Rotate keys regularly (every 90 days)
- Restrict by IP address when possible
- Monitor key usage for anomalies

**Example:**
```bash
# ❌ Don't hardcode keys
const API_KEY = "sk_live_1234567890abcdef";

# ✅ Use environment variables
const API_KEY = process.env.KYB_API_KEY;
```

### Can I restrict API key usage by IP address?

Yes, you can whitelist specific IP addresses:

```bash
curl -X PUT "https://api.kyb-platform.com/v1/keys/YOUR_API_KEY" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "allowed_ips": ["192.168.1.0/24", "10.0.0.0/8"]
  }'
```

## Risk Assessment

### How long does a risk assessment take?

- **Basic Assessment**: < 1 second
- **Comprehensive Assessment**: 2-5 seconds
- **Industry-Specific Assessment**: 3-7 seconds
- **Batch Assessment**: Varies by batch size

### What factors are considered in risk assessment?

**Financial Factors:**
- Credit score and history
- Revenue and profitability
- Debt-to-equity ratio
- Cash flow analysis

**Operational Factors:**
- Business age and stability
- Management experience
- Operational efficiency
- Market position

**Compliance Factors:**
- Regulatory compliance
- Sanctions screening
- Adverse media monitoring
- Industry-specific regulations

**External Factors:**
- Economic conditions
- Industry trends
- Geographic risk
- Market volatility

### Can I get explanations for risk factors?

Yes, use the explainability endpoint:

```bash
curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/explain/risk_1234567890" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

This returns SHAP-like explanations showing how each factor contributed to the overall risk score.

### How often should I reassess risk?

**Recommended Frequency:**
- **High-risk businesses**: Monthly
- **Medium-risk businesses**: Quarterly
- **Low-risk businesses**: Annually
- **After significant events**: Immediately

**Triggers for Reassessment:**
- Regulatory changes
- Adverse media coverage
- Financial changes
- Operational changes
- Market conditions

## Predictions & Analytics

### What prediction horizons are supported?

- **Short-term**: 1-3 months (XGBoost model)
- **Medium-term**: 3-6 months (Ensemble model)
- **Long-term**: 6-12 months (LSTM model)
- **Custom**: Up to 24 months

### How do I generate risk predictions?

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/assess/risk_1234567890/predict" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "horizon_months": 6,
    "scenarios": ["optimistic", "realistic", "pessimistic"]
  }'
```

### What is scenario analysis?

Scenario analysis uses Monte Carlo simulations to model different risk scenarios:

- **Optimistic**: Best-case scenario with favorable conditions
- **Realistic**: Most likely scenario based on current trends
- **Pessimistic**: Worst-case scenario with adverse conditions

### How accurate are long-term predictions?

**Accuracy by Horizon:**
- **1-3 months**: 92% accuracy
- **3-6 months**: 89% accuracy
- **6-12 months**: 85% accuracy
- **12+ months**: 78% accuracy

Longer-term predictions have lower accuracy due to increased uncertainty.

## Compliance & Screening

### What compliance checks are performed?

**Sanctions Screening:**
- OFAC (Office of Foreign Assets Control)
- UN Security Council
- EU Sanctions List
- UK HM Treasury

**Adverse Media Monitoring:**
- News articles and reports
- Regulatory announcements
- Legal proceedings
- Reputation analysis

**Regulatory Compliance:**
- Industry-specific regulations
- Licensing requirements
- Reporting obligations
- Compliance history

### How do I perform sanctions screening?

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/sanctions/screen" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "business_address": "123 Main St, Anytown, ST 12345",
    "country": "US"
  }'
```

### What happens if a sanctions match is found?

If a match is found:
1. **Immediate Alert**: Webhook notification sent
2. **Risk Score**: Automatically set to critical (1.0)
3. **Compliance Report**: Detailed match information
4. **Recommendation**: Block or investigate further

### How do I monitor adverse media?

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/media/monitor" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_name": "Acme Corporation",
    "keywords": ["fraud", "scandal", "investigation"],
    "days_back": 30
  }'
```

## Webhooks & Notifications

### What webhook events are available?

- `assessment.completed`: Risk assessment finished
- `assessment.failed`: Risk assessment failed
- `prediction.updated`: Risk prediction generated
- `compliance.alert`: Compliance issue detected
- `sanctions.match`: Sanctions list match found
- `media.alert`: Adverse media detected

### How do I set up webhooks?

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/webhooks" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/risk-assessment",
    "events": ["assessment.completed", "prediction.updated"],
    "secret": "your_webhook_secret_here"
  }'
```

### How do I verify webhook signatures?

```javascript
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature, 'hex'),
    Buffer.from(expectedSignature, 'hex')
  );
}
```

### What is the webhook retry policy?

- **Initial retry**: 1 second
- **Maximum retries**: 5 attempts
- **Backoff multiplier**: 2x
- **Maximum delay**: 5 minutes

## Billing & Pricing

### How is billing calculated?

**Pricing Model:**
- **Assessment**: $0.10 per assessment
- **Prediction**: $0.05 per prediction
- **Compliance Check**: $0.15 per check
- **Webhook**: $0.01 per webhook delivery

**Volume Discounts:**
- **1,000+ assessments/month**: 10% discount
- **10,000+ assessments/month**: 20% discount
- **100,000+ assessments/month**: 30% discount

### What is included in the free tier?

**Free Tier (Test Environment):**
- 100 assessments per month
- 50 predictions per month
- 25 compliance checks per month
- Basic webhook support
- Test data only

### How do I upgrade my plan?

1. Log into your [KYB Platform Dashboard](https://dashboard.kyb-platform.com)
2. Navigate to **Billing** → **Plans**
3. Select your desired plan
4. Complete the upgrade process

### Can I get a custom enterprise plan?

Yes, contact our sales team for custom enterprise plans with:
- Unlimited API calls
- Dedicated support
- Custom SLA
- On-premise deployment options
- Custom integrations

## Technical Support

### What programming languages are supported?

**Official SDKs:**
- **Go**: `go get github.com/kyb-platform/go-sdk`
- **Python**: `pip install kyb-sdk`
- **Node.js**: `npm install kyb-sdk`

**Community SDKs:**
- **Ruby**: `gem install kyb-sdk`
- **Java**: Available on Maven Central
- **PHP**: Available on Packagist

### How do I get help with integration?

**Support Channels:**
- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **GitHub Issues**: [https://github.com/kyb-platform/risk-assessment-service/issues](https://github.com/kyb-platform/risk-assessment-service/issues)
- **Email Support**: [api-support@kyb-platform.com](mailto:api-support@kyb-platform.com)
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

### What is your SLA?

**Service Level Agreement:**
- **Uptime**: 99.9% availability
- **Response Time**: < 1 second for 95% of requests
- **Support Response**: 4 hours for critical issues
- **Data Retention**: 7 years for compliance

### How do I report a bug?

1. **Check existing issues** on GitHub
2. **Create a new issue** with:
   - Detailed description
   - Steps to reproduce
   - Expected vs actual behavior
   - API key (test key only)
   - Request/response examples

### How do I request a feature?

1. **Check existing feature requests** on GitHub
2. **Create a new issue** with:
   - Feature description
   - Use case and benefits
   - Implementation suggestions
   - Priority level

## Data & Privacy

### What data do you store?

**Stored Data:**
- Business information (name, address, industry)
- Risk assessment results
- Prediction history
- Compliance check results
- API usage logs

**Not Stored:**
- Personal identifiable information (PII)
- Financial account details
- Sensitive business data
- Raw external API responses

### How long is data retained?

- **Assessment Data**: 7 years (compliance requirement)
- **API Logs**: 1 year
- **Usage Analytics**: 2 years
- **Error Logs**: 90 days

### Is my data encrypted?

**Encryption:**
- **In Transit**: TLS 1.3 encryption
- **At Rest**: AES-256 encryption
- **Database**: Encrypted with rotating keys
- **Backups**: Encrypted and geographically distributed

### Do you comply with GDPR?

Yes, we are GDPR compliant:
- **Data Processing**: Lawful basis for processing
- **Data Subject Rights**: Access, rectification, erasure
- **Data Protection**: Technical and organizational measures
- **Data Breach**: Notification within 72 hours
- **Privacy by Design**: Built into our systems

### Can I export my data?

Yes, you can export your data:
- **API**: Use the export endpoints
- **Dashboard**: Download CSV/JSON files
- **Support**: Request full data export
- **Format**: JSON, CSV, or custom format

## Performance & Scaling

### What is the maximum request size?

- **Single Assessment**: 1MB
- **Batch Assessment**: 10MB (up to 1,000 businesses)
- **Webhook Payload**: 1MB
- **File Upload**: 5MB

### How many concurrent requests can I make?

**Concurrent Limits:**
- **Starter Plan**: 10 concurrent requests
- **Professional Plan**: 50 concurrent requests
- **Enterprise Plan**: 200 concurrent requests

### What is the maximum batch size?

- **Batch Assessment**: 1,000 businesses per batch
- **Batch Prediction**: 500 businesses per batch
- **Batch Compliance**: 100 businesses per batch

### How do I optimize API performance?

**Best Practices:**
- Use batch endpoints for multiple requests
- Implement caching for repeated requests
- Use webhooks instead of polling
- Optimize request payload size
- Implement connection pooling

## Migration & Updates

### How do I migrate from v1 to v2?

**Migration Steps:**
1. **Review changelog** for breaking changes
2. **Update SDKs** to latest version
3. **Test in staging** environment
4. **Update API endpoints** if needed
5. **Deploy to production**

### What are the breaking changes in v2?

**Breaking Changes:**
- New response format for predictions
- Updated webhook event structure
- Changed error response format
- New authentication requirements

### How do I handle API versioning?

**Versioning Strategy:**
- **URL Versioning**: `/api/v1/`, `/api/v2/`
- **Header Versioning**: `API-Version: v2`
- **Backward Compatibility**: 12 months support
- **Deprecation Notice**: 6 months advance notice

### When will v1 be deprecated?

- **Deprecation Notice**: January 15, 2024
- **End of Support**: July 15, 2024
- **Migration Deadline**: July 15, 2024

## Enterprise Features

### What enterprise features are available?

**Enterprise Features:**
- **Custom Risk Models**: Industry-specific models
- **Dedicated Support**: 24/7 phone support
- **SLA**: 99.99% uptime guarantee
- **On-premise Deployment**: Self-hosted option
- **Custom Integrations**: Tailored solutions
- **Advanced Analytics**: Custom dashboards
- **White-label**: Branded solutions

### How do I get enterprise support?

**Enterprise Support:**
- **Email**: [enterprise@kyb-platform.com](mailto:enterprise@kyb-platform.com)
- **Phone**: +1-555-KYB-ENTERPRISE
- **Slack**: Dedicated enterprise channel
- **On-site**: Available for large deployments

### Can I get a custom SLA?

Yes, enterprise customers can get custom SLAs:
- **Uptime**: Up to 99.99%
- **Response Time**: Custom performance targets
- **Support**: Dedicated support team
- **Penalties**: Service credits for SLA breaches

### Do you offer on-premise deployment?

Yes, we offer on-premise deployment for enterprise customers:
- **Docker Containers**: Easy deployment
- **Kubernetes**: Scalable orchestration
- **Air-gapped**: Offline deployment
- **Custom Integration**: Tailored solutions

## Still Have Questions?

### Contact Support

- **Email**: [api-support@kyb-platform.com](mailto:api-support@kyb-platform.com)
- **Phone**: +1-555-KYB-HELP
- **Chat**: Available in dashboard
- **GitHub**: [Create an issue](https://github.com/kyb-platform/risk-assessment-service/issues)

### Community Resources

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)
- **Stack Overflow**: [kyb-platform tag](https://stackoverflow.com/questions/tagged/kyb-platform)
- **Slack**: [kyb-platform.slack.com](https://kyb-platform.slack.com)

### Status & Updates

- **Status Page**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Changelog**: [https://docs.kyb-platform.com/changelog](https://docs.kyb-platform.com/changelog)
- **Blog**: [https://blog.kyb-platform.com](https://blog.kyb-platform.com)
- **Twitter**: [@kybplatform](https://twitter.com/kybplatform)

---

**Last Updated**: January 15, 2024  
**Version**: 2.0.0  
**Next Review**: April 15, 2024
