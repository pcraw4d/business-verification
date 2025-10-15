# Debugging Guide

## Overview

This guide provides comprehensive debugging techniques, tools, and strategies for troubleshooting issues with the Risk Assessment Service API. Learn how to identify, diagnose, and resolve problems effectively.

## Debugging Methodology

### 1. Systematic Approach

**Step 1: Reproduce the Issue**
- Identify the exact conditions that trigger the problem
- Document the steps to reproduce
- Note the expected vs actual behavior

**Step 2: Gather Information**
- Collect error messages and logs
- Capture request/response data
- Note system environment details

**Step 3: Isolate the Problem**
- Test with minimal data
- Try different endpoints
- Check external dependencies

**Step 4: Apply Solutions**
- Start with simple fixes
- Test each solution thoroughly
- Document what works

### 2. Debugging Tools

**API Testing Tools:**
- **Postman**: GUI-based API testing
- **curl**: Command-line HTTP client
- **Insomnia**: Lightweight API client
- **HTTPie**: User-friendly command-line tool

**Monitoring Tools:**
- **Browser DevTools**: Network tab for web requests
- **Wireshark**: Network packet analysis
- **tcpdump**: Command-line packet capture
- **ngrok**: Local tunnel for webhook testing

## Request/Response Debugging

### 1. Enable Debug Logging

**Node.js SDK:**
```javascript
const { KYBClient } = require('kyb-sdk');

const client = new KYBClient('YOUR_API_KEY', {
  debug: true,
  logLevel: 'debug'
});
```

**Python SDK:**
```python
from kyb_sdk import KYBClient
import logging

# Enable debug logging
logging.basicConfig(level=logging.DEBUG)

client = KYBClient(
    api_key="YOUR_API_KEY",
    debug=True
)
```

**Go SDK:**
```go
package main

import (
    "log"
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client, err := kyb.NewClient(&kyb.Config{
        APIKey:  "YOUR_API_KEY",
        Debug:   true,
        LogLevel: "debug",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2. Log Request/Response Data

**JavaScript:**
```javascript
function logAPICall(url, options, response) {
  console.log('=== API Request ===');
  console.log('URL:', url);
  console.log('Method:', options.method);
  console.log('Headers:', options.headers);
  console.log('Body:', options.body);
  
  console.log('=== API Response ===');
  console.log('Status:', response.status);
  console.log('Headers:', response.headers);
  console.log('Body:', response.body);
}

// Usage
const response = await fetch(url, options);
logAPICall(url, options, response);
```

**Python:**
```python
import logging
import json

def log_api_call(url, method, headers, body, response):
    logger = logging.getLogger(__name__)
    
    logger.debug("=== API Request ===")
    logger.debug(f"URL: {url}")
    logger.debug(f"Method: {method}")
    logger.debug(f"Headers: {headers}")
    logger.debug(f"Body: {body}")
    
    logger.debug("=== API Response ===")
    logger.debug(f"Status: {response.status_code}")
    logger.debug(f"Headers: {response.headers}")
    logger.debug(f"Body: {response.text}")
```

**Go:**
```go
package main

import (
    "log"
    "net/http"
)

func logAPICall(req *http.Request, resp *http.Response) {
    log.Println("=== API Request ===")
    log.Printf("URL: %s", req.URL)
    log.Printf("Method: %s", req.Method)
    log.Printf("Headers: %v", req.Header)
    
    log.Println("=== API Response ===")
    log.Printf("Status: %s", resp.Status)
    log.Printf("Headers: %v", resp.Header)
}
```

### 3. Capture Network Traffic

**Using curl with verbose output:**
```bash
curl -v -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/assess" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "business_address": "123 Test St, Test City, TC 12345",
    "industry": "Technology",
    "country": "US"
  }'
```

**Using HTTPie with debug output:**
```bash
http --debug POST https://risk-assessment-service-production.up.railway.app/api/v1/assess \
  Authorization:"Bearer YOUR_API_KEY" \
  business_name="Test Company" \
  business_address="123 Test St, Test City, TC 12345" \
  industry="Technology" \
  country="US"
```

**Using tcpdump to capture packets:**
```bash
# Capture HTTP traffic on port 443
sudo tcpdump -i any -s 0 -w api_traffic.pcap port 443

# Analyze captured traffic
tcpdump -r api_traffic.pcap -A
```

## Error Analysis

### 1. HTTP Status Code Analysis

**4xx Client Errors:**
```javascript
function analyzeClientError(response) {
  switch (response.status) {
    case 400:
      console.log('Bad Request - Check request format and required fields');
      break;
    case 401:
      console.log('Unauthorized - Check API key and authentication');
      break;
    case 403:
      console.log('Forbidden - Check permissions and IP whitelist');
      break;
    case 404:
      console.log('Not Found - Check endpoint URL and resource ID');
      break;
    case 429:
      console.log('Rate Limited - Implement backoff and retry logic');
      break;
    default:
      console.log(`Client Error: ${response.status}`);
  }
}
```

**5xx Server Errors:**
```javascript
function analyzeServerError(response) {
  switch (response.status) {
    case 500:
      console.log('Internal Server Error - Service issue, retry later');
      break;
    case 502:
      console.log('Bad Gateway - Upstream service issue');
      break;
    case 503:
      console.log('Service Unavailable - Maintenance or overload');
      break;
    case 504:
      console.log('Gateway Timeout - Request took too long');
      break;
    default:
      console.log(`Server Error: ${response.status}`);
  }
}
```

### 2. Error Response Analysis

**Parse Error Details:**
```javascript
async function analyzeErrorResponse(response) {
  try {
    const errorData = await response.json();
    
    console.log('Error Code:', errorData.code);
    console.log('Error Message:', errorData.message);
    console.log('Error Details:', errorData.details);
    console.log('Request ID:', errorData.request_id);
    console.log('Timestamp:', errorData.timestamp);
    
    // Check for validation errors
    if (errorData.validation) {
      console.log('Validation Errors:');
      errorData.validation.forEach(error => {
        console.log(`- ${error.field}: ${error.message}`);
      });
    }
    
  } catch (parseError) {
    console.log('Failed to parse error response:', parseError);
  }
}
```

**Python Error Analysis:**
```python
import json

def analyze_error_response(response):
    try:
        error_data = response.json()
        
        print(f"Error Code: {error_data.get('code')}")
        print(f"Error Message: {error_data.get('message')}")
        print(f"Error Details: {error_data.get('details')}")
        print(f"Request ID: {error_data.get('request_id')}")
        print(f"Timestamp: {error_data.get('timestamp')}")
        
        # Check for validation errors
        if 'validation' in error_data:
            print("Validation Errors:")
            for error in error_data['validation']:
                print(f"- {error['field']}: {error['message']}")
                
    except json.JSONDecodeError:
        print("Failed to parse error response")
```

### 3. Common Error Patterns

**Authentication Errors:**
```javascript
function debugAuthenticationError(error) {
  if (error.message.includes('Invalid API key')) {
    console.log('Check API key format and validity');
    console.log('Test key format: sk_test_...');
    console.log('Live key format: sk_live_...');
  }
  
  if (error.message.includes('Token expired')) {
    console.log('JWT token has expired, refresh required');
  }
  
  if (error.message.includes('IP address not allowed')) {
    console.log('Check IP whitelist configuration');
    console.log('Current IP:', await getCurrentIP());
  }
}
```

**Validation Errors:**
```javascript
function debugValidationError(error) {
  if (error.details && error.details.field) {
    console.log(`Field '${error.details.field}' validation failed`);
    console.log(`Message: ${error.details.message}`);
    
    // Check field requirements
    const requiredFields = ['business_name', 'business_address', 'industry', 'country'];
    if (requiredFields.includes(error.details.field)) {
      console.log('This is a required field');
    }
  }
}
```

## Performance Debugging

### 1. Response Time Analysis

**Measure API Response Times:**
```javascript
async function measureResponseTime(apiCall) {
  const startTime = performance.now();
  
  try {
    const result = await apiCall();
    const endTime = performance.now();
    const duration = endTime - startTime;
    
    console.log(`API call completed in ${duration.toFixed(2)}ms`);
    
    if (duration > 5000) {
      console.warn('API call took longer than expected (>5s)');
    }
    
    return result;
  } catch (error) {
    const endTime = performance.now();
    const duration = endTime - startTime;
    
    console.error(`API call failed after ${duration.toFixed(2)}ms:`, error);
    throw error;
  }
}
```

**Python Performance Monitoring:**
```python
import time
import functools

def measure_response_time(func):
    @functools.wraps(func)
    async def wrapper(*args, **kwargs):
        start_time = time.time()
        
        try:
            result = await func(*args, **kwargs)
            end_time = time.time()
            duration = (end_time - start_time) * 1000  # Convert to milliseconds
            
            print(f"API call completed in {duration:.2f}ms")
            
            if duration > 5000:
                print("Warning: API call took longer than expected (>5s)")
            
            return result
        except Exception as error:
            end_time = time.time()
            duration = (end_time - start_time) * 1000
            
            print(f"API call failed after {duration:.2f}ms: {error}")
            raise
    
    return wrapper
```

### 2. Memory Usage Debugging

**Monitor Memory Usage:**
```javascript
function monitorMemoryUsage() {
  if (performance.memory) {
    const memory = performance.memory;
    console.log('Memory Usage:');
    console.log(`- Used: ${(memory.usedJSHeapSize / 1024 / 1024).toFixed(2)} MB`);
    console.log(`- Total: ${(memory.totalJSHeapSize / 1024 / 1024).toFixed(2)} MB`);
    console.log(`- Limit: ${(memory.jsHeapSizeLimit / 1024 / 1024).toFixed(2)} MB`);
  }
}

// Monitor memory before and after API calls
monitorMemoryUsage();
const result = await apiCall();
monitorMemoryUsage();
```

**Node.js Memory Monitoring:**
```javascript
const v8 = require('v8');

function logMemoryUsage() {
  const heapStats = v8.getHeapStatistics();
  console.log('Heap Statistics:');
  console.log(`- Total Heap Size: ${(heapStats.total_heap_size / 1024 / 1024).toFixed(2)} MB`);
  console.log(`- Used Heap Size: ${(heapStats.used_heap_size / 1024 / 1024).toFixed(2)} MB`);
  console.log(`- Heap Size Limit: ${(heapStats.heap_size_limit / 1024 / 1024).toFixed(2)} MB`);
}
```

### 3. Network Performance

**Analyze Network Performance:**
```javascript
function analyzeNetworkPerformance(response) {
  const timing = response.timing;
  
  if (timing) {
    console.log('Network Timing:');
    console.log(`- DNS Lookup: ${timing.domainLookupEnd - timing.domainLookupStart}ms`);
    console.log(`- TCP Connect: ${timing.connectEnd - timing.connectStart}ms`);
    console.log(`- SSL Handshake: ${timing.secureConnectionStart ? timing.connectEnd - timing.secureConnectionStart : 'N/A'}ms`);
    console.log(`- Request: ${timing.responseStart - timing.requestStart}ms`);
    console.log(`- Response: ${timing.responseEnd - timing.responseStart}ms`);
    console.log(`- Total: ${timing.responseEnd - timing.navigationStart}ms`);
  }
}
```

## Webhook Debugging

### 1. Webhook Endpoint Testing

**Test Webhook Endpoint:**
```bash
# Test webhook endpoint locally
curl -X POST "http://localhost:3000/webhooks/risk-assessment" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Signature: sha256=test_signature" \
  -H "X-Webhook-Timestamp: 1640995200" \
  -H "X-Webhook-Id: wh_test_123" \
  -d '{
    "event": "assessment.completed",
    "data": {
      "id": "risk_test_123",
      "business_id": "biz_test_123",
      "risk_score": 0.75,
      "risk_level": "medium"
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "webhook_id": "wh_test_123"
  }'
```

**Use ngrok for Local Testing:**
```bash
# Install ngrok
npm install -g ngrok

# Start your local webhook server
node webhook-server.js

# In another terminal, expose local server
ngrok http 3000

# Use the ngrok URL for webhook configuration
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/webhooks" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://abc123.ngrok.io/webhooks/risk-assessment",
    "events": ["assessment.completed"],
    "secret": "test_secret"
  }'
```

### 2. Webhook Signature Debugging

**Debug Signature Verification:**
```javascript
const crypto = require('crypto');

function debugWebhookSignature(payload, signature, secret) {
  console.log('=== Webhook Signature Debug ===');
  console.log('Payload:', payload);
  console.log('Received Signature:', signature);
  console.log('Secret Length:', secret.length);
  
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  console.log('Expected Signature:', expectedSignature);
  console.log('Signatures Match:', expectedSignature === signature);
  
  // Test with different payload formats
  const payloadVariations = [
    payload,
    JSON.stringify(JSON.parse(payload)), // Re-stringify
    payload.trim() // Remove whitespace
  ];
  
  payloadVariations.forEach((variation, index) => {
    const sig = crypto
      .createHmac('sha256', secret)
      .update(variation)
      .digest('hex');
    console.log(`Variation ${index + 1}:`, sig);
  });
}
```

### 3. Webhook Delivery Monitoring

**Monitor Webhook Delivery:**
```javascript
async function monitorWebhookDelivery(webhookId) {
  const response = await fetch(`https://risk-assessment-service-production.up.railway.app/api/v1/webhooks/${webhookId}`, {
    headers: {
      'Authorization': `Bearer ${API_KEY}`
    }
  });
  
  const webhook = await response.json();
  
  console.log('Webhook Status:');
  console.log(`- Active: ${webhook.active}`);
  console.log(`- Total Deliveries: ${webhook.delivery_stats.total_deliveries}`);
  console.log(`- Successful: ${webhook.delivery_stats.successful_deliveries}`);
  console.log(`- Failed: ${webhook.delivery_stats.failed_deliveries}`);
  console.log(`- Success Rate: ${(webhook.delivery_stats.success_rate * 100).toFixed(2)}%`);
  
  if (webhook.last_delivery) {
    console.log('Last Delivery:');
    console.log(`- Timestamp: ${webhook.last_delivery.timestamp}`);
    console.log(`- Status: ${webhook.last_delivery.status}`);
    console.log(`- Response Time: ${webhook.last_delivery.response_time}ms`);
  }
}
```

## Database Debugging

### 1. Query Performance Analysis

**Analyze Database Queries:**
```sql
-- Enable query logging
SET log_statement = 'all';
SET log_duration = on;
SET log_min_duration_statement = 100; -- Log queries > 100ms

-- Analyze slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
WHERE mean_time > 1000
ORDER BY mean_time DESC;

-- Check index usage
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE tablename = 'risk_assessments';
```

### 2. Connection Pool Monitoring

**Monitor Connection Pool:**
```javascript
const { Pool } = require('pg');

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});

// Monitor pool statistics
setInterval(() => {
  console.log('Connection Pool Stats:');
  console.log(`- Total Connections: ${pool.totalCount}`);
  console.log(`- Idle Connections: ${pool.idleCount}`);
  console.log(`- Waiting Clients: ${pool.waitingCount}`);
}, 30000);

// Handle pool errors
pool.on('error', (err) => {
  console.error('Unexpected error on idle client', err);
});
```

## External API Debugging

### 1. External API Health Monitoring

**Check External API Status:**
```javascript
async function checkExternalAPIHealth() {
  try {
    const response = await fetch('https://risk-assessment-service-production.up.railway.app/api/v1/external/health', {
      headers: {
        'Authorization': `Bearer ${API_KEY}`
      }
    });
    
    const health = await response.json();
    
    console.log('External API Health:');
    console.log(`- Overall Status: ${health.status}`);
    console.log(`- Health Score: ${health.overall_health_score}`);
    
    Object.entries(health.services).forEach(([service, status]) => {
      console.log(`- ${service}: ${status.status} (${status.response_time}ms)`);
    });
    
  } catch (error) {
    console.error('Failed to check external API health:', error);
  }
}
```

### 2. External API Error Handling

**Handle External API Errors:**
```javascript
async function handleExternalAPIError(error) {
  console.log('External API Error:');
  console.log(`- Code: ${error.code}`);
  console.log(`- Message: ${error.message}`);
  
  if (error.code === 'EXTERNAL_API_TIMEOUT') {
    console.log('Solution: Implement retry with exponential backoff');
  } else if (error.code === 'EXTERNAL_API_RATE_LIMIT') {
    console.log('Solution: Implement request queuing');
  } else if (error.code === 'EXTERNAL_API_UNAVAILABLE') {
    console.log('Solution: Use fallback assessment without external data');
  }
}
```

## Debugging Tools and Utilities

### 1. API Testing Script

**Comprehensive API Test Script:**
```javascript
#!/usr/bin/env node

const { KYBClient } = require('kyb-sdk');

async function runAPITests() {
  const client = new KYBClient(process.env.KYB_API_KEY, {
    debug: true,
    logLevel: 'debug'
  });
  
  console.log('=== API Health Check ===');
  try {
    const health = await client.getHealth();
    console.log('Health Status:', health.status);
  } catch (error) {
    console.error('Health check failed:', error);
  }
  
  console.log('\n=== Basic Risk Assessment ===');
  try {
    const assessment = await client.assessRisk({
      business_name: 'Test Company',
      business_address: '123 Test St, Test City, TC 12345',
      industry: 'Technology',
      country: 'US'
    });
    console.log('Assessment ID:', assessment.id);
    console.log('Risk Score:', assessment.risk_score);
  } catch (error) {
    console.error('Assessment failed:', error);
  }
  
  console.log('\n=== Performance Stats ===');
  try {
    const stats = await client.getPerformanceStats();
    console.log('Response Times:', stats.response_times);
    console.log('Throughput:', stats.throughput);
  } catch (error) {
    console.error('Performance stats failed:', error);
  }
}

runAPITests().catch(console.error);
```

### 2. Error Simulation Script

**Simulate Common Errors:**
```javascript
async function simulateErrors() {
  const client = new KYBClient('invalid_key');
  
  console.log('=== Testing Invalid API Key ===');
  try {
    await client.assessRisk({});
  } catch (error) {
    console.log('Expected error:', error.message);
  }
  
  console.log('\n=== Testing Invalid Data ===');
  try {
    await client.assessRisk({
      business_name: '', // Invalid: empty name
      business_address: '123 Test St',
      industry: 'Technology',
      country: 'US'
    });
  } catch (error) {
    console.log('Expected error:', error.message);
  }
  
  console.log('\n=== Testing Rate Limiting ===');
  const promises = [];
  for (let i = 0; i < 150; i++) {
    promises.push(client.assessRisk({
      business_name: `Test Company ${i}`,
      business_address: '123 Test St, Test City, TC 12345',
      industry: 'Technology',
      country: 'US'
    }));
  }
  
  try {
    await Promise.all(promises);
  } catch (error) {
    console.log('Expected rate limit error:', error.message);
  }
}

simulateErrors().catch(console.error);
```

### 3. Performance Benchmark Script

**Benchmark API Performance:**
```javascript
async function benchmarkAPI() {
  const client = new KYBClient(process.env.KYB_API_KEY);
  const iterations = 100;
  const results = [];
  
  console.log(`Running ${iterations} API calls...`);
  
  for (let i = 0; i < iterations; i++) {
    const startTime = performance.now();
    
    try {
      await client.assessRisk({
        business_name: `Benchmark Company ${i}`,
        business_address: '123 Test St, Test City, TC 12345',
        industry: 'Technology',
        country: 'US'
      });
      
      const endTime = performance.now();
      results.push(endTime - startTime);
      
    } catch (error) {
      console.error(`Request ${i} failed:`, error.message);
    }
  }
  
  // Calculate statistics
  const sortedResults = results.sort((a, b) => a - b);
  const avg = results.reduce((a, b) => a + b, 0) / results.length;
  const p50 = sortedResults[Math.floor(sortedResults.length * 0.5)];
  const p95 = sortedResults[Math.floor(sortedResults.length * 0.95)];
  const p99 = sortedResults[Math.floor(sortedResults.length * 0.99)];
  
  console.log('\n=== Performance Results ===');
  console.log(`Average: ${avg.toFixed(2)}ms`);
  console.log(`P50: ${p50.toFixed(2)}ms`);
  console.log(`P95: ${p95.toFixed(2)}ms`);
  console.log(`P99: ${p99.toFixed(2)}ms`);
  console.log(`Success Rate: ${(results.length / iterations * 100).toFixed(2)}%`);
}

benchmarkAPI().catch(console.error);
```

## Debugging Checklist

### Pre-Debugging Checklist
- [ ] Check API status page
- [ ] Verify API key validity
- [ ] Confirm network connectivity
- [ ] Review recent changes
- [ ] Check rate limits

### During Debugging
- [ ] Enable debug logging
- [ ] Capture request/response data
- [ ] Test with minimal data
- [ ] Try different endpoints
- [ ] Check error details

### Post-Debugging
- [ ] Document the solution
- [ ] Update error handling
- [ ] Add monitoring
- [ ] Test edge cases
- [ ] Share learnings

## Common Debugging Scenarios

### Scenario 1: API Key Issues
**Symptoms:** 401 Unauthorized errors
**Debug Steps:**
1. Verify API key format
2. Check key permissions
3. Test with curl
4. Regenerate key if needed

### Scenario 2: Rate Limiting
**Symptoms:** 429 Rate Limit Exceeded
**Debug Steps:**
1. Check rate limit headers
2. Implement backoff logic
3. Monitor request patterns
4. Consider upgrading plan

### Scenario 3: Slow Responses
**Symptoms:** High response times
**Debug Steps:**
1. Measure response times
2. Check network performance
3. Analyze request size
4. Implement caching

### Scenario 4: Webhook Failures
**Symptoms:** Webhooks not received
**Debug Steps:**
1. Test webhook endpoint
2. Verify signature validation
3. Check delivery statistics
4. Monitor webhook logs

## Support and Resources

### Debugging Resources
- **[Troubleshooting Guide](TROUBLESHOOTING.md)**
- **[API Documentation](API_DOCUMENTATION.md)**
- **[FAQ](FAQ.md)**
- **[GitHub Issues](https://github.com/kyb-platform/risk-assessment-service/issues)**

### Support Channels
- **Email**: [api-support@kyb-platform.com](mailto:api-support@kyb-platform.com)
- **Chat**: Available in dashboard
- **Phone**: +1-555-KYB-HELP
- **Community**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

### Tools and Utilities
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **API Explorer**: [https://api-explorer.kyb-platform.com](https://api-explorer.kyb-platform.com)
- **Webhook Tester**: [https://webhook-tester.kyb-platform.com](https://webhook-tester.kyb-platform.com)

---

**Last Updated**: January 15, 2024  
**Version**: 2.0.0  
**Next Review**: April 15, 2024
