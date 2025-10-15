# API Authentication Guide

## Overview

The Risk Assessment Service API uses API key-based authentication for secure access to all endpoints. This guide covers authentication methods, security best practices, and troubleshooting authentication issues.

## Authentication Methods

### API Key Authentication

All API requests require an API key in the `Authorization` header:

```bash
Authorization: Bearer YOUR_API_KEY
```

**Example:**
```bash
curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/assess/risk_123" \
  -H "Authorization: Bearer sk_live_1234567890abcdef"
```

### JWT Token Authentication (Enterprise)

Enterprise customers can use JWT tokens for enhanced security:

```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Getting Your API Key

### 1. Dashboard Access

1. Log into your [KYB Platform Dashboard](https://dashboard.kyb-platform.com)
2. Navigate to **Settings** → **API Keys**
3. Click **Create New API Key**

### 2. API Key Types

#### Test Keys (Development)
- **Prefix**: `sk_test_`
- **Environment**: Sandbox/Staging
- **Rate Limit**: 100 requests/minute
- **Data**: Synthetic test data only

#### Live Keys (Production)
- **Prefix**: `sk_live_`
- **Environment**: Production
- **Rate Limit**: Based on your plan
- **Data**: Real business data

### 3. Key Permissions

API keys can be configured with specific permissions:

```json
{
  "permissions": {
    "risk_assessment": true,
    "compliance_check": true,
    "sanctions_screening": true,
    "media_monitoring": false,
    "analytics": true,
    "webhooks": true
  },
  "rate_limits": {
    "requests_per_minute": 100,
    "requests_per_day": 10000
  },
  "allowed_ips": [
    "192.168.1.0/24",
    "10.0.0.0/8"
  ]
}
```

## Security Best Practices

### 1. Key Management

#### Store Keys Securely
```bash
# ❌ Don't hardcode keys in source code
const API_KEY = "sk_live_1234567890abcdef";

# ✅ Use environment variables
const API_KEY = process.env.KYB_API_KEY;

# ✅ Use secure key management services
const API_KEY = await keyManager.getSecret('kyb-api-key');
```

#### Rotate Keys Regularly
- Rotate API keys every 90 days
- Use key versioning for zero-downtime rotation
- Monitor key usage for anomalies

#### Limit Key Scope
- Create separate keys for different environments
- Use minimal required permissions
- Restrict by IP address when possible

### 2. Request Security

#### Use HTTPS Only
```bash
# ❌ Never use HTTP
curl http://api.kyb-platform.com/v1/assess

# ✅ Always use HTTPS
curl https://api.kyb-platform.com/v1/assess
```

#### Validate SSL Certificates
```javascript
// Node.js example
const https = require('https');
const agent = new https.Agent({
  rejectUnauthorized: true // Validate SSL certificates
});

fetch('https://api.kyb-platform.com/v1/assess', {
  agent: agent
});
```

#### Implement Request Signing (Enterprise)
```javascript
// Request signing for enhanced security
const crypto = require('crypto');

function signRequest(method, path, body, timestamp, secret) {
  const message = `${method}${path}${body}${timestamp}`;
  return crypto
    .createHmac('sha256', secret)
    .update(message)
    .digest('hex');
}

const timestamp = Date.now();
const signature = signRequest('POST', '/api/v1/assess', body, timestamp, secret);

headers['X-Timestamp'] = timestamp;
headers['X-Signature'] = signature;
```

### 3. Error Handling

#### Don't Log API Keys
```javascript
// ❌ Don't log the full error with API key
console.error('API Error:', error);

// ✅ Log only safe information
console.error('API Error:', {
  status: error.status,
  message: error.message,
  code: error.code
});
```

#### Handle Authentication Errors
```javascript
async function makeAPICall() {
  try {
    const response = await fetch('/api/v1/assess', {
      headers: {
        'Authorization': `Bearer ${API_KEY}`
      }
    });
    
    if (response.status === 401) {
      // Handle authentication error
      throw new Error('Invalid API key or expired token');
    }
    
    return await response.json();
  } catch (error) {
    if (error.message.includes('401')) {
      // Refresh API key or notify admin
      await refreshAPIKey();
    }
    throw error;
  }
}
```

## Rate Limiting

### Rate Limit Headers

Every API response includes rate limit information:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
X-RateLimit-Window: 60
```

### Handling Rate Limits

#### Implement Exponential Backoff
```javascript
async function makeAPICallWithRetry(url, options, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const response = await fetch(url, options);
      
      if (response.status === 429) {
        const resetTime = response.headers.get('X-RateLimit-Reset');
        const waitTime = Math.pow(2, attempt) * 1000; // Exponential backoff
        
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

#### Monitor Rate Limit Usage
```javascript
function checkRateLimit(response) {
  const remaining = parseInt(response.headers.get('X-RateLimit-Remaining'));
  const limit = parseInt(response.headers.get('X-RateLimit-Limit'));
  
  if (remaining / limit < 0.1) { // Less than 10% remaining
    console.warn('Rate limit nearly exceeded. Consider implementing request queuing.');
  }
}
```

## IP Whitelisting

### Configure IP Restrictions

For enhanced security, restrict API key usage to specific IP addresses:

```bash
curl -X PUT "https://api.kyb-platform.com/v1/keys/sk_live_1234567890abcdef" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "allowed_ips": [
      "192.168.1.0/24",
      "10.0.0.0/8",
      "203.0.113.0/24"
    ]
  }'
```

### CIDR Notation Examples

```
192.168.1.0/24    # 192.168.1.0 - 192.168.1.255
10.0.0.0/8        # 10.0.0.0 - 10.255.255.255
203.0.113.0/24    # 203.0.113.0 - 203.0.113.255
```

## Webhook Authentication

### Webhook Signatures

Webhook payloads include a signature for verification:

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

// Express.js webhook handler
app.post('/webhooks/risk-assessment', (req, res) => {
  const signature = req.headers['x-webhook-signature'];
  const payload = JSON.stringify(req.body);
  
  if (!verifyWebhookSignature(payload, signature, WEBHOOK_SECRET)) {
    return res.status(401).send('Invalid signature');
  }
  
  // Process webhook
  res.status(200).send('OK');
});
```

### Webhook Security Headers

```http
X-Webhook-Signature: sha256=abc123...
X-Webhook-Timestamp: 1640995200
X-Webhook-Id: wh_1234567890
```

## Troubleshooting

### Common Authentication Issues

#### 1. Invalid API Key
```json
{
  "error": "Invalid API key",
  "code": "AUTHENTICATION_ERROR",
  "message": "The provided API key is invalid or has been revoked"
}
```

**Solutions:**
- Verify the API key is correct
- Check for typos or extra spaces
- Ensure you're using the correct key type (test vs live)
- Regenerate the key if necessary

#### 2. Expired Token
```json
{
  "error": "Token expired",
  "code": "AUTHENTICATION_ERROR",
  "message": "The provided JWT token has expired"
}
```

**Solutions:**
- Refresh your JWT token
- Implement automatic token renewal
- Check system clock synchronization

#### 3. IP Address Not Allowed
```json
{
  "error": "IP address not allowed",
  "code": "AUTHORIZATION_ERROR",
  "message": "Your IP address is not in the allowed list for this API key"
}
```

**Solutions:**
- Add your IP address to the whitelist
- Check if you're behind a proxy or load balancer
- Contact support for IP whitelist updates

#### 4. Rate Limit Exceeded
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "You have exceeded your rate limit of 100 requests per minute"
}
```

**Solutions:**
- Implement request queuing
- Use exponential backoff
- Upgrade your plan for higher limits
- Optimize your API usage

### Debugging Authentication

#### Enable Debug Logging
```javascript
// Enable detailed logging for debugging
const client = new KYBClient({
  apiKey: 'YOUR_API_KEY',
  debug: true,
  logLevel: 'debug'
});
```

#### Test API Key Validity
```bash
curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/health" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

#### Validate Webhook Signatures
```bash
# Test webhook signature verification
curl -X POST "https://your-app.com/webhooks/test" \
  -H "X-Webhook-Signature: sha256=abc123..." \
  -H "Content-Type: application/json" \
  -d '{"test": "payload"}'
```

## SDK Authentication Examples

### Go SDK
```go
package main

import (
    "context"
    "log"
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client, err := kyb.NewClient(&kyb.Config{
        BaseURL: "https://risk-assessment-service-production.up.railway.app/api/v1",
        APIKey:  "YOUR_API_KEY",
        Timeout: 30 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    assessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Assessment: %+v", assessment)
}
```

### Python SDK
```python
from kyb_sdk import KYBClient

# Initialize client with API key
client = KYBClient(
    api_key="YOUR_API_KEY",
    base_url="https://risk-assessment-service-production.up.railway.app/api/v1",
    timeout=30
)

# Make API call
try:
    assessment = client.assess_risk(
        business_name="Acme Corporation",
        business_address="123 Main St, Anytown, ST 12345",
        industry="Technology",
        country="US"
    )
    print(f"Risk Score: {assessment['risk_score']}")
except Exception as e:
    print(f"Error: {e}")
```

### Node.js SDK
```javascript
const { KYBClient } = require('kyb-sdk');

const client = new KYBClient('YOUR_API_KEY', {
  baseURL: 'https://risk-assessment-service-production.up.railway.app/api/v1',
  timeout: 30000
});

async function assessRisk() {
  try {
    const assessment = await client.assessRisk({
      businessName: 'Acme Corporation',
      businessAddress: '123 Main St, Anytown, ST 12345',
      industry: 'Technology',
      country: 'US'
    });
    
    console.log(`Risk Score: ${assessment.risk_score}`);
  } catch (error) {
    console.error('Error:', error.message);
  }
}

assessRisk();
```

## Security Checklist

- [ ] API keys stored securely (environment variables, key management service)
- [ ] HTTPS used for all API requests
- [ ] API keys rotated regularly (every 90 days)
- [ ] IP whitelisting configured where appropriate
- [ ] Rate limiting implemented with exponential backoff
- [ ] Webhook signatures verified
- [ ] Authentication errors handled gracefully
- [ ] API keys not logged or exposed in client-side code
- [ ] SSL certificate validation enabled
- [ ] Request signing implemented (enterprise)

## Support

For authentication-related issues:

- **Email**: [security@kyb-platform.com](mailto:security@kyb-platform.com)
- **Documentation**: [https://docs.kyb-platform.com/authentication](https://docs.kyb-platform.com/authentication)
- **Status Page**: [https://status.kyb-platform.com](https://status.kyb-platform.com)

## Changelog

### v2.0.0 (2024-01-15)
- **NEW**: JWT token authentication for enterprise customers
- **NEW**: Request signing for enhanced security
- **NEW**: IP whitelisting with CIDR notation support
- **NEW**: Webhook signature verification
- **ENHANCED**: Rate limiting with detailed headers
- **ENHANCED**: Authentication error messages
- **ENHANCED**: Security best practices documentation

### v1.0.0 (2024-01-15)
- Initial API key authentication
- Basic rate limiting
- HTTPS enforcement
