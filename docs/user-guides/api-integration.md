# KYB Platform - API Integration Guide

This guide provides comprehensive information for integrating the KYB Platform API into your applications. Whether you're building a simple integration or a complex enterprise system, this guide covers everything you need to know.

## Table of Contents

1. [Authentication](#authentication)
2. [SDK Integration](#sdk-integration)
3. [REST API Reference](#rest-api-reference)
4. [Best Practices](#best-practices)
5. [Error Handling](#error-handling)
6. [Rate Limiting](#rate-limiting)
7. [Webhooks](#webhooks)
8. [Advanced Patterns](#advanced-patterns)
9. [Testing](#testing)
10. [Production Deployment](#production-deployment)

## Authentication

### API Keys vs JWT Tokens

The KYB Platform supports two authentication methods:

**API Keys** (Recommended for server-to-server):
- Long-lived credentials
- No token refresh needed
- Ideal for background jobs and services

**JWT Tokens** (Recommended for user sessions):
- Short-lived access tokens
- Refresh token mechanism
- Better for user-facing applications

### API Key Authentication

```bash
# Using API Key in headers
curl -X POST https://api.kybplatform.com/v1/classify \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Acme Corp"}'
```

```python
# Python SDK
import kyb_platform

client = kyb_platform.Client(api_key="kyb_live_1234567890abcdef")
result = client.classify(business_name="Acme Corp")
```

```javascript
// JavaScript SDK
const { KYBPlatform } = require('@kyb-platform/sdk');

const client = new KYBPlatform({
  apiKey: 'kyb_live_1234567890abcdef'
});

const result = await client.classify({ businessName: 'Acme Corp' });
```

### JWT Token Authentication

```bash
# Login to get tokens
curl -X POST https://api.kybplatform.com/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@company.com",
    "password": "password123"
  }'

# Use access token
curl -X POST https://api.kybplatform.com/v1/classify \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Acme Corp"}'

# Refresh token when expired
curl -X POST https://api.kybplatform.com/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

```python
# Python SDK with JWT
import kyb_platform

client = kyb_platform.Client(
    email="user@company.com",
    password="password123"
)

# SDK handles token refresh automatically
result = client.classify(business_name="Acme Corp")
```

## SDK Integration

### Python SDK

**Installation**:
```bash
pip install kyb-platform
```

**Basic Usage**:
```python
import kyb_platform
from kyb_platform import KYBPlatformError

# Initialize client
client = kyb_platform.Client(api_key="your_api_key")

try:
    # Classify a business
    classification = client.classify(
        business_name="Acme Corporation",
        address="123 Business St, New York, NY 10001"
    )
    print(f"NAICS Code: {classification.primary_classification.naics_code}")
    
    # Assess risk
    risk_assessment = client.assess_risk(
        business_id=classification.business_id,
        assessment_type="comprehensive"
    )
    print(f"Risk Level: {risk_assessment.risk_level}")
    
    # Check compliance
    compliance = client.check_compliance(
        business_id=classification.business_id,
        frameworks=["soc2", "pci_dss"]
    )
    print(f"Compliance Score: {compliance.overall_compliance_score}")
    
except KYBPlatformError as e:
    print(f"Error: {e.message}")
    print(f"Code: {e.code}")
```

**Advanced Usage**:
```python
# Batch processing
businesses = [
    {"name": "Business A", "address": "123 Main St"},
    {"name": "Business B", "address": "456 Oak Ave"},
    {"name": "Business C", "address": "789 Pine Rd"}
]

results = client.classify_batch(businesses)

# Custom retry logic
from kyb_platform import RetryConfig

client = kyb_platform.Client(
    api_key="your_api_key",
    retry_config=RetryConfig(
        max_retries=3,
        backoff_factor=2,
        retry_on_status_codes=[429, 500, 502, 503, 504]
    )
)

# Async operations
import asyncio

async def process_businesses():
    async with kyb_platform.AsyncClient(api_key="your_api_key") as client:
        tasks = []
        for business in businesses:
            task = client.classify(business_name=business["name"])
            tasks.append(task)
        
        results = await asyncio.gather(*tasks)
        return results

# Run async function
results = asyncio.run(process_businesses())
```

### JavaScript/Node.js SDK

**Installation**:
```bash
npm install @kyb-platform/sdk
```

**Basic Usage**:
```javascript
const { KYBPlatform, KYBPlatformError } = require('@kyb-platform/sdk');

// Initialize client
const client = new KYBPlatform({
  apiKey: 'your_api_key'
});

async function processBusiness() {
  try {
    // Classify a business
    const classification = await client.classify({
      businessName: 'Acme Corporation',
      address: '123 Business St, New York, NY 10001'
    });
    
    console.log(`NAICS Code: ${classification.primaryClassification.naicsCode}`);
    
    // Assess risk
    const riskAssessment = await client.assessRisk({
      businessId: classification.businessId,
      assessmentType: 'comprehensive'
    });
    
    console.log(`Risk Level: ${riskAssessment.riskLevel}`);
    
    // Check compliance
    const compliance = await client.checkCompliance({
      businessId: classification.businessId,
      frameworks: ['soc2', 'pci_dss']
    });
    
    console.log(`Compliance Score: ${compliance.overallComplianceScore}`);
    
  } catch (error) {
    if (error instanceof KYBPlatformError) {
      console.error(`Error: ${error.message}`);
      console.error(`Code: ${error.code}`);
    } else {
      console.error('Unexpected error:', error);
    }
  }
}

processBusiness();
```

**Advanced Usage**:
```javascript
// Batch processing
const businesses = [
  { name: 'Business A', address: '123 Main St' },
  { name: 'Business B', address: '456 Oak Ave' },
  { name: 'Business C', address: '789 Pine Rd' }
];

const results = await client.classifyBatch(businesses);

// Custom retry configuration
const client = new KYBPlatform({
  apiKey: 'your_api_key',
  retryConfig: {
    maxRetries: 3,
    backoffFactor: 2,
    retryOnStatusCodes: [429, 500, 502, 503, 504]
  }
});

// Webhook handling
const express = require('express');
const app = express();

app.post('/webhooks/kyb', express.raw({ type: 'application/json' }), (req, res) => {
  const signature = req.headers['x-kyb-signature'];
  const isValid = client.verifyWebhookSignature(req.body, signature, 'your_webhook_secret');
  
  if (!isValid) {
    return res.status(400).send('Invalid signature');
  }
  
  const event = JSON.parse(req.body);
  console.log('Received event:', event.type, event.data);
  
  res.status(200).send('OK');
});
```

### Go SDK

**Installation**:
```bash
go get github.com/kyb-platform/go-sdk
```

**Basic Usage**:
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kyb-platform/go-sdk"
)

func main() {
    // Initialize client
    client := kyb.NewClient("your_api_key")
    
    ctx := context.Background()
    
    // Classify a business
    classification, err := client.Classify(ctx, &kyb.ClassificationRequest{
        BusinessName: "Acme Corporation",
        Address:      "123 Business St, New York, NY 10001",
    })
    if err != nil {
        log.Fatalf("Classification error: %v", err)
    }
    
    fmt.Printf("NAICS Code: %s\n", classification.PrimaryClassification.NAICSCode)
    
    // Assess risk
    riskAssessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessID:      classification.BusinessID,
        AssessmentType:  "comprehensive",
    })
    if err != nil {
        log.Fatalf("Risk assessment error: %v", err)
    }
    
    fmt.Printf("Risk Level: %s\n", riskAssessment.RiskLevel)
    
    // Check compliance
    compliance, err := client.CheckCompliance(ctx, &kyb.ComplianceRequest{
        BusinessID: classification.BusinessID,
        Frameworks: []string{"soc2", "pci_dss"},
    })
    if err != nil {
        log.Fatalf("Compliance check error: %v", err)
    }
    
    fmt.Printf("Compliance Score: %.2f\n", compliance.OverallComplianceScore)
}
```

**Advanced Usage**:
```go
// Batch processing
businesses := []kyb.Business{
    {Name: "Business A", Address: "123 Main St"},
    {Name: "Business B", Address: "456 Oak Ave"},
    {Name: "Business C", Address: "789 Pine Rd"},
}

results, err := client.ClassifyBatch(ctx, &kyb.BatchClassificationRequest{
    Businesses: businesses,
})
if err != nil {
    log.Fatalf("Batch classification error: %v", err)
}

// Custom retry configuration
client := kyb.NewClient("your_api_key", kyb.WithRetryConfig(&kyb.RetryConfig{
    MaxRetries:          3,
    BackoffFactor:       2,
    RetryOnStatusCodes:  []int{429, 500, 502, 503, 504},
}))

// Concurrent processing
func processBusinesses(ctx context.Context, client *kyb.Client, businesses []string) {
    semaphore := make(chan struct{}, 10) // Limit concurrent requests
    results := make(chan *kyb.Classification, len(businesses))
    
    for _, businessName := range businesses {
        go func(name string) {
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            classification, err := client.Classify(ctx, &kyb.ClassificationRequest{
                BusinessName: name,
            })
            if err != nil {
                log.Printf("Error classifying %s: %v", name, err)
                return
            }
            
            results <- classification
        }(businessName)
    }
    
    // Collect results
    for i := 0; i < len(businesses); i++ {
        result := <-results
        fmt.Printf("Classified: %s -> %s\n", result.BusinessName, result.PrimaryClassification.NAICSCode)
    }
}
```

## REST API Reference

### Core Endpoints

**Business Classification**:
```bash
# Single classification
POST /v1/classify
{
  "business_name": "Acme Corporation",
  "address": "123 Business St, New York, NY 10001",
  "website": "https://acme.com"
}

# Batch classification
POST /v1/classify/batch
{
  "businesses": [
    {"name": "Business A", "address": "123 Main St"},
    {"name": "Business B", "address": "456 Oak Ave"}
  ]
}

# Get classification history
GET /v1/classify/history?business_id=business-123&limit=10&offset=0
```

**Risk Assessment**:
```bash
# Assess risk
POST /v1/risk/assess
{
  "business_id": "business-123",
  "assessment_type": "comprehensive",
  "include_factors": true
}

# Get risk history
GET /v1/risk/history?business_id=business-123&limit=10

# Get risk alerts
GET /v1/risk/alerts?status=active
```

**Compliance Checking**:
```bash
# Check compliance
POST /v1/compliance/check
{
  "business_id": "business-123",
  "frameworks": ["soc2", "pci_dss", "gdpr"]
}

# Generate compliance report
POST /v1/compliance/reports
{
  "business_id": "business-123",
  "report_type": "soc2_audit",
  "date_range": {
    "start": "2024-01-01",
    "end": "2024-12-31"
  }
}
```

### Response Format

All API responses follow a consistent format:

```json
{
  "success": true,
  "data": {
    // Response data here
  },
  "meta": {
    "request_id": "req-456",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**Error Response**:
```json
{
  "success": false,
  "error": {
    "code": "validation_error",
    "message": "Invalid business name",
    "details": {
      "field": "business_name",
      "issue": "Required field is missing"
    }
  },
  "meta": {
    "request_id": "req-456",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Best Practices

### Performance Optimization

**Connection Pooling**:
```python
# Python - Use connection pooling
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

session = requests.Session()
adapter = HTTPAdapter(
    pool_connections=10,
    pool_maxsize=20,
    max_retries=Retry(total=3, backoff_factor=0.1)
)
session.mount('https://', adapter)

client = kyb_platform.Client(api_key="your_api_key", session=session)
```

**Caching**:
```python
# Implement caching for repeated requests
import redis
import json

redis_client = redis.Redis(host='localhost', port=6379, db=0)

def get_cached_classification(business_name):
    cache_key = f"classification:{business_name}"
    cached = redis_client.get(cache_key)
    if cached:
        return json.loads(cached)
    return None

def cache_classification(business_name, result, ttl=3600):
    cache_key = f"classification:{business_name}"
    redis_client.setex(cache_key, ttl, json.dumps(result))
```

**Batch Processing**:
```python
# Process businesses in batches
def process_large_dataset(businesses, batch_size=100):
    results = []
    for i in range(0, len(businesses), batch_size):
        batch = businesses[i:i + batch_size]
        batch_results = client.classify_batch(batch)
        results.extend(batch_results)
        
        # Rate limiting
        time.sleep(1)  # 1 second between batches
    
    return results
```

### Security Best Practices

**Secure API Key Storage**:
```python
# Use environment variables
import os
from dotenv import load_dotenv

load_dotenv()
api_key = os.getenv('KYB_API_KEY')

# Use secret management in production
import boto3

def get_api_key():
    client = boto3.client('secretsmanager')
    response = client.get_secret_value(SecretId='kyb/api-key')
    return response['SecretString']
```

**Input Validation**:
```python
# Validate input before sending to API
def validate_business_data(business_name, address):
    if not business_name or len(business_name.strip()) < 2:
        raise ValueError("Business name must be at least 2 characters")
    
    if address and len(address) > 500:
        raise ValueError("Address too long")
    
    return {
        "business_name": business_name.strip(),
        "address": address.strip() if address else None
    }
```

**HTTPS Only**:
```python
# Ensure HTTPS connections
client = kyb_platform.Client(
    api_key="your_api_key",
    base_url="https://api.kybplatform.com"  # Always use HTTPS
)
```

## Error Handling

### Common Error Codes

**4xx Client Errors**:
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Invalid or missing authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict
- `422 Unprocessable Entity`: Validation errors
- `429 Too Many Requests`: Rate limit exceeded

**5xx Server Errors**:
- `500 Internal Server Error`: Unexpected server error
- `502 Bad Gateway`: External service error
- `503 Service Unavailable`: Service temporarily unavailable
- `504 Gateway Timeout`: External service timeout

### Error Handling Patterns

**Python**:
```python
import kyb_platform
from kyb_platform import KYBPlatformError

def handle_api_call(func, *args, **kwargs):
    try:
        return func(*args, **kwargs)
    except KYBPlatformError as e:
        if e.code == 429:
            # Rate limited - implement backoff
            time.sleep(2 ** e.retry_after)
            return handle_api_call(func, *args, **kwargs)
        elif e.code == 401:
            # Authentication error - refresh token
            client.refresh_token()
            return handle_api_call(func, *args, **kwargs)
        elif e.code >= 500:
            # Server error - retry with exponential backoff
            time.sleep(1)
            return handle_api_call(func, *args, **kwargs)
        else:
            # Client error - log and handle appropriately
            logger.error(f"API Error: {e.message}")
            raise
```

**JavaScript**:
```javascript
async function handleApiCall(apiCall, ...args) {
  try {
    return await apiCall(...args);
  } catch (error) {
    if (error.code === 429) {
      // Rate limited - implement backoff
      await new Promise(resolve => setTimeout(resolve, Math.pow(2, error.retryAfter) * 1000));
      return handleApiCall(apiCall, ...args);
    } else if (error.code === 401) {
      // Authentication error - refresh token
      await client.refreshToken();
      return handleApiCall(apiCall, ...args);
    } else if (error.code >= 500) {
      // Server error - retry with exponential backoff
      await new Promise(resolve => setTimeout(resolve, 1000));
      return handleApiCall(apiCall, ...args);
    } else {
      // Client error - log and handle appropriately
      console.error(`API Error: ${error.message}`);
      throw error;
    }
  }
}
```

## Rate Limiting

### Understanding Rate Limits

**Rate Limit Headers**:
```bash
curl -I https://api.kybplatform.com/v1/classify
# Response headers:
# X-RateLimit-Limit: 1000
# X-RateLimit-Remaining: 999
# X-RateLimit-Reset: 1642233600
# Retry-After: 60
```

**Rate Limit Tiers**:
- **Free**: 1,000 requests/month
- **Professional**: 100,000 requests/month
- **Enterprise**: Custom limits

### Rate Limit Handling

**Exponential Backoff**:
```python
import time
import random

def exponential_backoff(attempt, max_attempts=5):
    if attempt >= max_attempts:
        raise Exception("Max retry attempts exceeded")
    
    # Exponential backoff with jitter
    delay = min(2 ** attempt + random.uniform(0, 1), 60)
    time.sleep(delay)

def api_call_with_backoff(func, *args, **kwargs):
    for attempt in range(5):
        try:
            return func(*args, **kwargs)
        except KYBPlatformError as e:
            if e.code == 429:
                exponential_backoff(attempt)
                continue
            else:
                raise
```

**Token Bucket Algorithm**:
```python
import time
import threading

class RateLimiter:
    def __init__(self, tokens_per_second):
        self.tokens_per_second = tokens_per_second
        self.tokens = tokens_per_second
        self.last_update = time.time()
        self.lock = threading.Lock()
    
    def acquire(self):
        with self.lock:
            now = time.time()
            time_passed = now - self.last_update
            self.tokens = min(
                self.tokens_per_second,
                self.tokens + time_passed * self.tokens_per_second
            )
            
            if self.tokens < 1:
                sleep_time = (1 - self.tokens) / self.tokens_per_second
                time.sleep(sleep_time)
                self.tokens = 0
            else:
                self.tokens -= 1
            
            self.last_update = now

# Usage
rate_limiter = RateLimiter(tokens_per_second=10)

def api_call():
    rate_limiter.acquire()
    return client.classify(business_name="Acme Corp")
```

## Webhooks

### Setting Up Webhooks

**Create Webhook**:
```bash
curl -X POST https://api.kybplatform.com/v1/webhooks \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/kyb",
    "events": ["classification.completed", "risk.alert", "compliance.updated"],
    "secret": "your-webhook-secret"
  }'
```

**Webhook Events**:
- `classification.completed`: Business classification finished
- `risk.alert`: Risk threshold exceeded
- `compliance.updated`: Compliance status changed
- `business.created`: New business added
- `business.updated`: Business information updated

### Webhook Handling

**Python Flask Example**:
```python
from flask import Flask, request, jsonify
import hmac
import hashlib

app = Flask(__name__)

@app.route('/webhooks/kyb', methods=['POST'])
def handle_webhook():
    # Verify webhook signature
    signature = request.headers.get('X-KYB-Signature')
    payload = request.get_data()
    
    expected_signature = hmac.new(
        'your-webhook-secret'.encode('utf-8'),
        payload,
        hashlib.sha256
    ).hexdigest()
    
    if not hmac.compare_digest(signature, expected_signature):
        return jsonify({'error': 'Invalid signature'}), 400
    
    # Process webhook event
    event = request.json
    
    if event['type'] == 'classification.completed':
        handle_classification_completed(event['data'])
    elif event['type'] == 'risk.alert':
        handle_risk_alert(event['data'])
    elif event['type'] == 'compliance.updated':
        handle_compliance_updated(event['data'])
    
    return jsonify({'status': 'success'}), 200

def handle_classification_completed(data):
    business_id = data['business_id']
    classification = data['classification']
    print(f"Classification completed for {business_id}: {classification['naics_code']}")

def handle_risk_alert(data):
    business_id = data['business_id']
    risk_score = data['risk_score']
    print(f"Risk alert for {business_id}: {risk_score}")

def handle_compliance_updated(data):
    business_id = data['business_id']
    compliance_status = data['compliance_status']
    print(f"Compliance updated for {business_id}: {compliance_status}")
```

**Node.js Express Example**:
```javascript
const express = require('express');
const crypto = require('crypto');

const app = express();

app.post('/webhooks/kyb', express.raw({ type: 'application/json' }), (req, res) => {
  // Verify webhook signature
  const signature = req.headers['x-kyb-signature'];
  const payload = req.body;
  
  const expectedSignature = crypto
    .createHmac('sha256', 'your-webhook-secret')
    .update(payload)
    .digest('hex');
  
  if (signature !== expectedSignature) {
    return res.status(400).json({ error: 'Invalid signature' });
  }
  
  // Process webhook event
  const event = JSON.parse(payload);
  
  switch (event.type) {
    case 'classification.completed':
      handleClassificationCompleted(event.data);
      break;
    case 'risk.alert':
      handleRiskAlert(event.data);
      break;
    case 'compliance.updated':
      handleComplianceUpdated(event.data);
      break;
  }
  
  res.status(200).json({ status: 'success' });
});

function handleClassificationCompleted(data) {
  console.log(`Classification completed for ${data.business_id}: ${data.classification.naics_code}`);
}

function handleRiskAlert(data) {
  console.log(`Risk alert for ${data.business_id}: ${data.risk_score}`);
}

function handleComplianceUpdated(data) {
  console.log(`Compliance updated for ${data.business_id}: ${data.compliance_status}`);
}
```

## Advanced Patterns

### Circuit Breaker Pattern

```python
import time
from enum import Enum

class CircuitState(Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"

class CircuitBreaker:
    def __init__(self, failure_threshold=5, recovery_timeout=60):
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.failure_count = 0
        self.last_failure_time = None
        self.state = CircuitState.CLOSED
    
    def call(self, func, *args, **kwargs):
        if self.state == CircuitState.OPEN:
            if time.time() - self.last_failure_time > self.recovery_timeout:
                self.state = CircuitState.HALF_OPEN
            else:
                raise Exception("Circuit breaker is open")
        
        try:
            result = func(*args, **kwargs)
            self.on_success()
            return result
        except Exception as e:
            self.on_failure()
            raise
    
    def on_success(self):
        self.failure_count = 0
        self.state = CircuitState.CLOSED
    
    def on_failure(self):
        self.failure_count += 1
        self.last_failure_time = time.time()
        
        if self.failure_count >= self.failure_threshold:
            self.state = CircuitState.OPEN

# Usage
circuit_breaker = CircuitBreaker()

def api_call():
    return circuit_breaker.call(client.classify, business_name="Acme Corp")
```

### Retry with Exponential Backoff

```python
import time
import random
from functools import wraps

def retry_with_backoff(max_retries=3, base_delay=1, max_delay=60):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            for attempt in range(max_retries + 1):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    if attempt == max_retries:
                        raise
                    
                    # Calculate delay with exponential backoff and jitter
                    delay = min(base_delay * (2 ** attempt) + random.uniform(0, 1), max_delay)
                    time.sleep(delay)
            
            return func(*args, **kwargs)
        return wrapper
    return decorator

@retry_with_backoff(max_retries=3)
def classify_business(business_name):
    return client.classify(business_name=business_name)
```

### Bulk Operations with Progress Tracking

```python
import asyncio
from tqdm import tqdm

async def bulk_classify(businesses, batch_size=100):
    results = []
    total_batches = (len(businesses) + batch_size - 1) // batch_size
    
    with tqdm(total=len(businesses), desc="Classifying businesses") as pbar:
        for i in range(0, len(businesses), batch_size):
            batch = businesses[i:i + batch_size]
            
            try:
                batch_results = await client.classify_batch(businesses=batch)
                results.extend(batch_results)
                pbar.update(len(batch))
            except Exception as e:
                print(f"Error processing batch {i//batch_size + 1}: {e}")
                # Continue with next batch
                continue
    
    return results

# Usage
businesses = [{"name": f"Business {i}"} for i in range(1000)]
results = asyncio.run(bulk_classify(businesses))
```

## Testing

### Unit Testing

```python
import unittest
from unittest.mock import Mock, patch
import kyb_platform

class TestKYBIntegration(unittest.TestCase):
    def setUp(self):
        self.client = kyb_platform.Client(api_key="test_key")
    
    @patch('kyb_platform.Client._make_request')
    def test_classify_business(self, mock_request):
        # Mock successful response
        mock_request.return_value = {
            "success": True,
            "data": {
                "classification_id": "test-123",
                "business_name": "Test Corp",
                "primary_classification": {
                    "naics_code": "541511",
                    "naics_title": "Custom Computer Programming Services"
                }
            }
        }
        
        result = self.client.classify(business_name="Test Corp")
        
        self.assertEqual(result.primary_classification.naics_code, "541511")
        mock_request.assert_called_once()
    
    @patch('kyb_platform.Client._make_request')
    def test_handle_api_error(self, mock_request):
        # Mock error response
        mock_request.side_effect = kyb_platform.KYBPlatformError(
            code=422,
            message="Invalid business name"
        )
        
        with self.assertRaises(kyb_platform.KYBPlatformError):
            self.client.classify(business_name="")
```

### Integration Testing

```python
import pytest
import os

@pytest.fixture
def client():
    api_key = os.getenv('KYB_TEST_API_KEY')
    if not api_key:
        pytest.skip("KYB_TEST_API_KEY not set")
    return kyb_platform.Client(api_key=api_key)

def test_real_classification(client):
    result = client.classify(business_name="Microsoft Corporation")
    
    assert result.success
    assert result.primary_classification.naics_code
    assert result.confidence_score > 0.5

def test_batch_classification(client):
    businesses = [
        {"name": "Apple Inc."},
        {"name": "Google LLC"},
        {"name": "Amazon.com Inc."}
    ]
    
    results = client.classify_batch(businesses)
    
    assert len(results) == 3
    for result in results:
        assert result.primary_classification.naics_code
```

## Production Deployment

### Environment Configuration

**Environment Variables**:
```bash
# Production
export KYB_API_KEY="kyb_live_1234567890abcdef"
export KYB_BASE_URL="https://api.kybplatform.com"
export KYB_TIMEOUT="30"
export KYB_MAX_RETRIES="3"

# Development
export KYB_API_KEY="kyb_test_1234567890abcdef"
export KYB_BASE_URL="https://api-staging.kybplatform.com"
```

**Configuration Management**:
```python
import os
from dataclasses import dataclass

@dataclass
class KYBConfig:
    api_key: str
    base_url: str = "https://api.kybplatform.com"
    timeout: int = 30
    max_retries: int = 3
    
    @classmethod
    def from_env(cls):
        return cls(
            api_key=os.getenv('KYB_API_KEY'),
            base_url=os.getenv('KYB_BASE_URL', 'https://api.kybplatform.com'),
            timeout=int(os.getenv('KYB_TIMEOUT', '30')),
            max_retries=int(os.getenv('KYB_MAX_RETRIES', '3'))
        )

config = KYBConfig.from_env()
client = kyb_platform.Client(
    api_key=config.api_key,
    base_url=config.base_url,
    timeout=config.timeout
)
```

### Monitoring and Logging

```python
import logging
import time
from functools import wraps

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def log_api_calls(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.time()
        
        try:
            result = func(*args, **kwargs)
            duration = time.time() - start_time
            
            logger.info(f"API call {func.__name__} completed in {duration:.2f}s")
            return result
            
        except Exception as e:
            duration = time.time() - start_time
            logger.error(f"API call {func.__name__} failed after {duration:.2f}s: {e}")
            raise
    
    return wrapper

# Usage
@log_api_calls
def classify_business(business_name):
    return client.classify(business_name=business_name)
```

### Health Checks

```python
def health_check():
    try:
        # Test API connectivity
        response = client.health()
        return response.status == "healthy"
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        return False

# Periodic health checks
import schedule
import time

schedule.every(5).minutes.do(health_check)

while True:
    schedule.run_pending()
    time.sleep(1)
```

---

## Conclusion

This API integration guide provides comprehensive information for integrating the KYB Platform into your applications. Key takeaways:

- **Use SDKs** for easier integration and automatic error handling
- **Implement proper error handling** with retry logic and exponential backoff
- **Follow rate limiting** guidelines to avoid API throttling
- **Set up webhooks** for real-time event notifications
- **Monitor and log** API calls for debugging and performance tracking
- **Test thoroughly** before deploying to production

For additional support, refer to our [API Documentation](https://api.kybplatform.com/docs) or contact our support team at support@kybplatform.com.
