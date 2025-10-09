# KYB Platform Risk Assessment Service Node.js SDK

[![npm version](https://badge.fury.io/js/kyb-sdk.svg)](https://badge.fury.io/js/kyb-sdk)
[![Node.js Support](https://img.shields.io/node/v/kyb-sdk.svg)](https://nodejs.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The official Node.js SDK for the KYB Platform Risk Assessment Service API. This SDK provides a simple and intuitive interface for performing business risk assessments, compliance checks, and analytics.

## Features

- **Risk Assessment**: Comprehensive business risk evaluation with ML-powered predictions
- **Compliance Checking**: Automated compliance verification and monitoring
- **Sanctions Screening**: Real-time sanctions and watchlist screening
- **Media Monitoring**: Adverse media monitoring and alerting
- **Analytics**: Risk trends and insights analysis
- **TypeScript Support**: Full TypeScript definitions included
- **Error Handling**: Comprehensive error handling with detailed error information
- **Retry Logic**: Automatic retry with exponential backoff
- **Promise-based**: Modern async/await support

## Installation

```bash
npm install kyb-sdk
```

## Quick Start

```javascript
const { KYBClient } = require('kyb-sdk');

// Initialize the client
const client = new KYBClient('your_api_key_here');

// Perform a risk assessment
async function assessRisk() {
    try {
        const assessment = await client.assessRisk({
            businessName: 'Acme Corporation',
            businessAddress: '123 Main St, Anytown, ST 12345',
            industry: 'Technology',
            country: 'US',
            phone: '+1-555-123-4567',
            email: 'contact@acme.com',
            website: 'https://www.acme.com'
        });

        console.log(`Risk Score: ${assessment.risk_score}`);
        console.log(`Risk Level: ${assessment.risk_level}`);
        console.log(`Confidence: ${assessment.confidence_score}`);
    } catch (error) {
        console.error('Error:', error.message);
    }
}

assessRisk();
```

## Usage Examples

### Risk Assessment

```javascript
const { KYBClient } = require('kyb-sdk');

const client = new KYBClient('your_api_key');

// Basic risk assessment
async function basicAssessment() {
    const assessment = await client.assessRisk({
        businessName: 'Acme Corporation',
        businessAddress: '123 Main St, Anytown, ST 12345',
        industry: 'Technology',
        country: 'US'
    });
    
    console.log(assessment);
}

// Risk assessment with metadata
async function assessmentWithMetadata() {
    const assessment = await client.assessRisk({
        businessName: 'Acme Corporation',
        businessAddress: '123 Main St, Anytown, ST 12345',
        industry: 'Technology',
        country: 'US',
        predictionHorizon: 6,
        metadata: {
            annual_revenue: 1000000,
            employee_count: 50,
            founded_year: 2020
        }
    });
    
    console.log(assessment);
}

// Get risk assessment by ID
async function getAssessment() {
    const assessment = await client.getRiskAssessment('risk_1234567890');
    console.log(assessment);
}
```

### Risk Prediction

```javascript
// Predict future risk
async function predictRisk() {
    const prediction = await client.predictRisk('risk_1234567890', {
        horizonMonths: 6,
        scenarios: ['optimistic', 'realistic', 'pessimistic']
    });
    
    console.log(`Predicted Score: ${prediction.predicted_score}`);
    console.log(`Predicted Level: ${prediction.predicted_level}`);
}
```

### Compliance Checking

```javascript
// Check compliance
async function checkCompliance() {
    const compliance = await client.checkCompliance({
        businessName: 'Acme Corporation',
        businessAddress: '123 Main St, Anytown, ST 12345',
        industry: 'Technology',
        country: 'US',
        complianceTypes: ['kyc', 'aml', 'sanctions']
    });
    
    console.log(`Compliance Status: ${compliance.compliance_status}`);
}
```

### Sanctions Screening

```javascript
// Screen for sanctions
async function screenSanctions() {
    const sanctions = await client.screenSanctions({
        businessName: 'Acme Corporation',
        businessAddress: '123 Main St, Anytown, ST 12345',
        country: 'US'
    });
    
    console.log(`Sanctions Status: ${sanctions.sanctions_status}`);
}
```

### Media Monitoring

```javascript
// Set up media monitoring
async function monitorMedia() {
    const monitoring = await client.monitorMedia({
        businessName: 'Acme Corporation',
        businessAddress: '123 Main St, Anytown, ST 12345',
        monitoringTypes: ['news', 'social_media', 'regulatory']
    });
    
    console.log(`Monitoring ID: ${monitoring.monitoring_id}`);
}
```

### Analytics

```javascript
// Get risk trends
async function getRiskTrends() {
    const trends = await client.getRiskTrends({
        industry: 'Technology',
        country: 'US',
        timeframe: '30d',
        limit: 100
    });
    
    console.log(`Average Risk Score: ${trends.summary.average_risk_score}`);
}

// Get risk insights
async function getRiskInsights() {
    const insights = await client.getRiskInsights({
        industry: 'Technology',
        country: 'US',
        riskLevel: 'high'
    });
    
    insights.insights.forEach(insight => {
        console.log(`Insight: ${insight.title}`);
    });
}
```

## Error Handling

The SDK provides comprehensive error handling with specific exception types:

```javascript
const { KYBClient, ValidationError, AuthenticationError, APIError } = require('kyb-sdk');

const client = new KYBClient('your_api_key');

async function handleErrors() {
    try {
        const assessment = await client.assessRisk({
            businessName: '', // This will cause a validation error
            businessAddress: '123 Main St',
            industry: 'Technology',
            country: 'US'
        });
    } catch (error) {
        if (error instanceof ValidationError) {
            console.log(`Validation Error: ${error.message}`);
            console.log(`Request ID: ${error.getRequestId()}`);
            error.getValidationErrors().forEach(err => {
                console.log(`Field: ${err.field}, Message: ${err.message}`);
            });
        } else if (error instanceof AuthenticationError) {
            console.log(`Authentication Error: ${error.message}`);
        } else if (error instanceof APIError) {
            console.log(`API Error: ${error.message}`);
            console.log(`Status Code: ${error.statusCode}`);
        } else {
            console.log(`Unexpected Error: ${error.message}`);
        }
    }
}
```

## Configuration

```javascript
const { KYBClient } = require('kyb-sdk');

// Custom configuration
const client = new KYBClient('your_api_key', {
    baseUrl: 'https://api.kyb-platform.com/v1',
    timeout: 60000,
    maxRetries: 5,
    userAgent: 'my-app/1.0.0'
});
```

## TypeScript Support

The SDK includes full TypeScript definitions:

```typescript
import { KYBClient, ValidationError, APIError } from 'kyb-sdk';

const client = new KYBClient('your_api_key');

interface RiskAssessment {
    id: string;
    risk_score: number;
    risk_level: string;
    confidence_score: number;
}

async function assessRisk(): Promise<RiskAssessment> {
    try {
        const assessment = await client.assessRisk({
            businessName: 'Acme Corporation',
            businessAddress: '123 Main St, Anytown, ST 12345',
            industry: 'Technology',
            country: 'US'
        });
        
        return assessment;
    } catch (error) {
        if (error instanceof ValidationError) {
            console.error('Validation failed:', error.message);
        } else if (error instanceof APIError) {
            console.error('API error:', error.message);
        }
        throw error;
    }
}
```

## Rate Limiting

The SDK automatically handles rate limiting with exponential backoff:

```javascript
const client = new KYBClient('your_api_key', {
    maxRetries: 5 // Maximum number of retries
});
```

## Logging

Enable logging to see request details:

```javascript
const { KYBClient } = require('kyb-sdk');

// Enable debug logging
process.env.DEBUG = 'kyb-sdk:*';

const client = new KYBClient('your_api_key');
```

## Testing

```javascript
const { KYBClient, ValidationError } = require('kyb-sdk');
const assert = require('assert');

describe('KYBClient', () => {
    let client;
    
    beforeEach(() => {
        client = new KYBClient('test_key');
    });
    
    it('should throw validation error for empty business name', async () => {
        try {
            await client.assessRisk({
                businessName: '', // Invalid: empty name
                businessAddress: '123 Main St',
                industry: 'Technology',
                country: 'US'
            });
            assert.fail('Should have thrown ValidationError');
        } catch (error) {
            assert(error instanceof ValidationError);
            assert(error.message.includes('businessName is required'));
        }
    });
    
    it('should successfully assess risk', async () => {
        // Mock the HTTP request
        const mockResponse = {
            data: {
                id: 'risk_123',
                risk_score: 0.75,
                risk_level: 'medium'
            }
        };
        
        // Mock axios
        const axios = require('axios');
        const originalPost = axios.post;
        axios.post = jest.fn().mockResolvedValue(mockResponse);
        
        const result = await client.assessRisk({
            businessName: 'Acme Corporation',
            businessAddress: '123 Main St',
            industry: 'Technology',
            country: 'US'
        });
        
        assert.equal(result.risk_score, 0.75);
        assert.equal(result.risk_level, 'medium');
        
        // Restore original function
        axios.post = originalPost;
    });
});
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **Issues**: [GitHub Issues](https://github.com/kyb-platform/nodejs-sdk/issues)
- **Email**: [support@kyb-platform.com](mailto:support@kyb-platform.com)

## Changelog

### v1.0.0 (2024-01-15)
- Initial release
- Risk assessment functionality
- Compliance checking
- Sanctions screening
- Media monitoring
- Analytics and insights
- Comprehensive error handling
- TypeScript support
- Promise-based API
