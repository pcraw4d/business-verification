# KYB Platform API SDK Documentation

This document provides comprehensive SDK documentation for integrating with the KYB Platform API using various programming languages.

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [JavaScript/Node.js SDK](#javascriptnodejs-sdk)
4. [Python SDK](#python-sdk)
5. [Go SDK](#go-sdk)
6. [Java SDK](#java-sdk)
7. [PHP SDK](#php-sdk)
8. [Ruby SDK](#ruby-sdk)
9. [C# SDK](#c-sdk)
10. [Best Practices](#best-practices)
11. [Error Handling](#error-handling)
12. [Rate Limiting](#rate-limiting)
13. [Testing](#testing)

## Overview

The KYB Platform API provides comprehensive business classification, risk assessment, and compliance checking capabilities. This documentation covers official and community SDKs for popular programming languages.

### Base URL
- **Development**: `http://localhost:8080`
- **Production**: `https://api.kybplatform.com`

### API Version
All endpoints are prefixed with `/v1/`

## Authentication

All API requests require authentication using JWT Bearer tokens.

### Getting an Access Token

```bash
curl -X POST https://api.kybplatform.com/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-email@example.com",
    "password": "your-password"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### Using the Token

Include the token in the Authorization header:
```
Authorization: Bearer <your_access_token>
```

## JavaScript/Node.js SDK

### Installation

```bash
npm install kyb-platform-sdk
```

### Basic Usage

```javascript
const KYBPlatform = require('kyb-platform-sdk');

// Initialize the client
const client = new KYBPlatform({
  baseURL: 'https://api.kybplatform.com',
  accessToken: 'your-access-token'
});

// Authenticate
async function authenticate() {
  try {
    const response = await client.auth.login({
      email: 'your-email@example.com',
      password: 'your-password'
    });
    
    console.log('Access token:', response.access_token);
    return response.access_token;
  } catch (error) {
    console.error('Authentication failed:', error.message);
  }
}

// Business Classification
async function classifyBusiness() {
  try {
    const classification = await client.classification.classify({
      business_name: 'Acme Corporation',
      business_address: '123 Main St, New York, NY 10001',
      business_phone: '+1-555-123-4567',
      business_website: 'https://acme.com'
    });
    
    console.log('Classification:', classification);
    return classification;
  } catch (error) {
    console.error('Classification failed:', error.message);
  }
}

// Risk Assessment
async function assessRisk() {
  try {
    const assessment = await client.risk.assess({
      business_id: 'business-123',
      business_name: 'Acme Corporation',
      categories: ['financial', 'operational']
    });
    
    console.log('Risk Assessment:', assessment);
    return assessment;
  } catch (error) {
    console.error('Risk assessment failed:', error.message);
  }
}

// Compliance Check
async function checkCompliance() {
  try {
    const compliance = await client.compliance.check({
      business_id: 'business-123',
      frameworks: ['SOC2', 'PCI_DSS', 'GDPR']
    });
    
    console.log('Compliance Status:', compliance);
    return compliance;
  } catch (error) {
    console.error('Compliance check failed:', error.message);
  }
}
```

### Advanced Usage

```javascript
// Batch Classification
async function batchClassification() {
  const businesses = [
    {
      business_name: 'Acme Corp',
      business_address: '123 Main St, New York, NY'
    },
    {
      business_name: 'Tech Solutions LLC',
      business_address: '456 Oak Ave, San Francisco, CA'
    }
  ];
  
  try {
    const results = await client.classification.batchClassify(businesses);
    console.log('Batch results:', results);
  } catch (error) {
    console.error('Batch classification failed:', error.message);
  }
}

// Get Classification History
async function getHistory() {
  try {
    const history = await client.classification.getHistory({
      business_id: 'business-123',
      limit: 10,
      offset: 0
    });
    
    console.log('Classification history:', history);
  } catch (error) {
    console.error('Failed to get history:', error.message);
  }
}

// Generate Confidence Report
async function generateConfidenceReport() {
  try {
    const report = await client.classification.generateConfidenceReport({
      business_id: 'business-123',
      classification_id: 'class-456'
    });
    
    console.log('Confidence report:', report);
  } catch (error) {
    console.error('Failed to generate report:', error.message);
  }
}
```

## Python SDK

### Installation

```bash
pip install kyb-platform-sdk
```

### Basic Usage

```python
from kyb_platform import KYBPlatform

# Initialize the client
client = KYBPlatform(
    base_url='https://api.kybplatform.com',
    access_token='your-access-token'
)

# Authenticate
def authenticate():
    try:
        response = client.auth.login(
            email='your-email@example.com',
            password='your-password'
        )
        print(f"Access token: {response['access_token']}")
        return response['access_token']
    except Exception as e:
        print(f"Authentication failed: {e}")
        return None

# Business Classification
def classify_business():
    try:
        classification = client.classification.classify({
            'business_name': 'Acme Corporation',
            'business_address': '123 Main St, New York, NY 10001',
            'business_phone': '+1-555-123-4567',
            'business_website': 'https://acme.com'
        })
        
        print(f"Classification: {classification}")
        return classification
    except Exception as e:
        print(f"Classification failed: {e}")
        return None

# Risk Assessment
def assess_risk():
    try:
        assessment = client.risk.assess({
            'business_id': 'business-123',
            'business_name': 'Acme Corporation',
            'categories': ['financial', 'operational']
        })
        
        print(f"Risk Assessment: {assessment}")
        return assessment
    except Exception as e:
        print(f"Risk assessment failed: {e}")
        return None

# Compliance Check
def check_compliance():
    try:
        compliance = client.compliance.check({
            'business_id': 'business-123',
            'frameworks': ['SOC2', 'PCI_DSS', 'GDPR']
        })
        
        print(f"Compliance Status: {compliance}")
        return compliance
    except Exception as e:
        print(f"Compliance check failed: {e}")
        return None
```

### Advanced Usage

```python
# Batch Classification
def batch_classification():
    businesses = [
        {
            'business_name': 'Acme Corp',
            'business_address': '123 Main St, New York, NY'
        },
        {
            'business_name': 'Tech Solutions LLC',
            'business_address': '456 Oak Ave, San Francisco, CA'
        }
    ]
    
    try:
        results = client.classification.batch_classify(businesses)
        print(f"Batch results: {results}")
    except Exception as e:
        print(f"Batch classification failed: {e}")

# Get Risk Categories
def get_risk_categories():
    try:
        categories = client.risk.get_categories()
        print(f"Risk categories: {categories}")
    except Exception as e:
        print(f"Failed to get categories: {e}")

# Get Risk Factors
def get_risk_factors():
    try:
        factors = client.risk.get_factors(category='financial')
        print(f"Risk factors: {factors}")
    except Exception as e:
        print(f"Failed to get factors: {e}")
```

## Go SDK

### Installation

```bash
go get github.com/kybplatform/go-sdk
```

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kybplatform/go-sdk"
)

func main() {
    // Initialize the client
    client := kyb.NewClient(&kyb.Config{
        BaseURL:     "https://api.kybplatform.com",
        AccessToken: "your-access-token",
    })
    
    // Authenticate
    authResponse, err := client.Auth.Login(&kyb.LoginRequest{
        Email:    "your-email@example.com",
        Password: "your-password",
    })
    if err != nil {
        log.Fatalf("Authentication failed: %v", err)
    }
    
    fmt.Printf("Access token: %s\n", authResponse.AccessToken)
    
    // Business Classification
    classification, err := client.Classification.Classify(&kyb.ClassificationRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, New York, NY 10001",
        BusinessPhone:   "+1-555-123-4567",
        BusinessWebsite: "https://acme.com",
    })
    if err != nil {
        log.Fatalf("Classification failed: %v", err)
    }
    
    fmt.Printf("Classification: %+v\n", classification)
    
    // Risk Assessment
    assessment, err := client.Risk.Assess(&kyb.RiskAssessmentRequest{
        BusinessID:   "business-123",
        BusinessName: "Acme Corporation",
        Categories:   []string{"financial", "operational"},
    })
    if err != nil {
        log.Fatalf("Risk assessment failed: %v", err)
    }
    
    fmt.Printf("Risk Assessment: %+v\n", assessment)
    
    // Compliance Check
    compliance, err := client.Compliance.Check(&kyb.ComplianceCheckRequest{
        BusinessID: "business-123",
        Frameworks: []string{"SOC2", "PCI_DSS", "GDPR"},
    })
    if err != nil {
        log.Fatalf("Compliance check failed: %v", err)
    }
    
    fmt.Printf("Compliance Status: %+v\n", compliance)
}
```

### Advanced Usage

```go
// Batch Classification
func batchClassification(client *kyb.Client) {
    businesses := []kyb.ClassificationRequest{
        {
            BusinessName:    "Acme Corp",
            BusinessAddress: "123 Main St, New York, NY",
        },
        {
            BusinessName:    "Tech Solutions LLC",
            BusinessAddress: "456 Oak Ave, San Francisco, CA",
        },
    }
    
    results, err := client.Classification.BatchClassify(&kyb.BatchClassificationRequest{
        Businesses: businesses,
    })
    if err != nil {
        log.Printf("Batch classification failed: %v", err)
        return
    }
    
    fmt.Printf("Batch results: %+v\n", results)
}

// Get Risk Categories
func getRiskCategories(client *kyb.Client) {
    categories, err := client.Risk.GetCategories()
    if err != nil {
        log.Printf("Failed to get categories: %v", err)
        return
    }
    
    fmt.Printf("Risk categories: %+v\n", categories)
}

// Get Risk Factors
func getRiskFactors(client *kyb.Client) {
    factors, err := client.Risk.GetFactors(&kyb.RiskFactorsRequest{
        Category: "financial",
    })
    if err != nil {
        log.Printf("Failed to get factors: %v", err)
        return
    }
    
    fmt.Printf("Risk factors: %+v\n", factors)
}
```

## Java SDK

### Installation

Add to your `pom.xml`:

```xml
<dependency>
    <groupId>com.kybplatform</groupId>
    <artifactId>kyb-platform-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Basic Usage

```java
import com.kybplatform.KYBPlatform;
import com.kybplatform.models.*;

public class KYBExample {
    public static void main(String[] args) {
        // Initialize the client
        KYBPlatform client = new KYBPlatform.Builder()
            .baseUrl("https://api.kybplatform.com")
            .accessToken("your-access-token")
            .build();
        
        // Authenticate
        try {
            LoginRequest loginRequest = new LoginRequest();
            loginRequest.setEmail("your-email@example.com");
            loginRequest.setPassword("your-password");
            
            LoginResponse response = client.auth().login(loginRequest);
            System.out.println("Access token: " + response.getAccessToken());
        } catch (Exception e) {
            System.err.println("Authentication failed: " + e.getMessage());
        }
        
        // Business Classification
        try {
            ClassificationRequest request = new ClassificationRequest();
            request.setBusinessName("Acme Corporation");
            request.setBusinessAddress("123 Main St, New York, NY 10001");
            request.setBusinessPhone("+1-555-123-4567");
            request.setBusinessWebsite("https://acme.com");
            
            ClassificationResponse classification = client.classification().classify(request);
            System.out.println("Classification: " + classification);
        } catch (Exception e) {
            System.err.println("Classification failed: " + e.getMessage());
        }
        
        // Risk Assessment
        try {
            RiskAssessmentRequest request = new RiskAssessmentRequest();
            request.setBusinessId("business-123");
            request.setBusinessName("Acme Corporation");
            request.setCategories(Arrays.asList("financial", "operational"));
            
            RiskAssessmentResponse assessment = client.risk().assess(request);
            System.out.println("Risk Assessment: " + assessment);
        } catch (Exception e) {
            System.err.println("Risk assessment failed: " + e.getMessage());
        }
        
        // Compliance Check
        try {
            ComplianceCheckRequest request = new ComplianceCheckRequest();
            request.setBusinessId("business-123");
            request.setFrameworks(Arrays.asList("SOC2", "PCI_DSS", "GDPR"));
            
            ComplianceCheckResponse compliance = client.compliance().check(request);
            System.out.println("Compliance Status: " + compliance);
        } catch (Exception e) {
            System.err.println("Compliance check failed: " + e.getMessage());
        }
    }
}
```

## PHP SDK

### Installation

```bash
composer require kybplatform/php-sdk
```

### Basic Usage

```php
<?php

require_once 'vendor/autoload.php';

use KYBPlatform\KYBPlatform;
use KYBPlatform\Models\LoginRequest;
use KYBPlatform\Models\ClassificationRequest;
use KYBPlatform\Models\RiskAssessmentRequest;
use KYBPlatform\Models\ComplianceCheckRequest;

// Initialize the client
$client = new KYBPlatform([
    'base_url' => 'https://api.kybplatform.com',
    'access_token' => 'your-access-token'
]);

// Authenticate
try {
    $loginRequest = new LoginRequest();
    $loginRequest->setEmail('your-email@example.com');
    $loginRequest->setPassword('your-password');
    
    $response = $client->auth()->login($loginRequest);
    echo "Access token: " . $response->getAccessToken() . "\n";
} catch (Exception $e) {
    echo "Authentication failed: " . $e->getMessage() . "\n";
}

// Business Classification
try {
    $request = new ClassificationRequest();
    $request->setBusinessName('Acme Corporation');
    $request->setBusinessAddress('123 Main St, New York, NY 10001');
    $request->setBusinessPhone('+1-555-123-4567');
    $request->setBusinessWebsite('https://acme.com');
    
    $classification = $client->classification()->classify($request);
    echo "Classification: " . json_encode($classification) . "\n";
} catch (Exception $e) {
    echo "Classification failed: " . $e->getMessage() . "\n";
}

// Risk Assessment
try {
    $request = new RiskAssessmentRequest();
    $request->setBusinessId('business-123');
    $request->setBusinessName('Acme Corporation');
    $request->setCategories(['financial', 'operational']);
    
    $assessment = $client->risk()->assess($request);
    echo "Risk Assessment: " . json_encode($assessment) . "\n";
} catch (Exception $e) {
    echo "Risk assessment failed: " . $e->getMessage() . "\n";
}

// Compliance Check
try {
    $request = new ComplianceCheckRequest();
    $request->setBusinessId('business-123');
    $request->setFrameworks(['SOC2', 'PCI_DSS', 'GDPR']);
    
    $compliance = $client->compliance()->check($request);
    echo "Compliance Status: " . json_encode($compliance) . "\n";
} catch (Exception $e) {
    echo "Compliance check failed: " . $e->getMessage() . "\n";
}
```

## Ruby SDK

### Installation

Add to your `Gemfile`:

```ruby
gem 'kyb-platform-sdk'
```

### Basic Usage

```ruby
require 'kyb-platform-sdk'

# Initialize the client
client = KYBPlatform::Client.new(
  base_url: 'https://api.kybplatform.com',
  access_token: 'your-access-token'
)

# Authenticate
begin
  response = client.auth.login(
    email: 'your-email@example.com',
    password: 'your-password'
  )
  puts "Access token: #{response.access_token}"
rescue => e
  puts "Authentication failed: #{e.message}"
end

# Business Classification
begin
  classification = client.classification.classify(
    business_name: 'Acme Corporation',
    business_address: '123 Main St, New York, NY 10001',
    business_phone: '+1-555-123-4567',
    business_website: 'https://acme.com'
  )
  puts "Classification: #{classification}"
rescue => e
  puts "Classification failed: #{e.message}"
end

# Risk Assessment
begin
  assessment = client.risk.assess(
    business_id: 'business-123',
    business_name: 'Acme Corporation',
    categories: ['financial', 'operational']
  )
  puts "Risk Assessment: #{assessment}"
rescue => e
  puts "Risk assessment failed: #{e.message}"
end

# Compliance Check
begin
  compliance = client.compliance.check(
    business_id: 'business-123',
    frameworks: ['SOC2', 'PCI_DSS', 'GDPR']
  )
  puts "Compliance Status: #{compliance}"
rescue => e
  puts "Compliance check failed: #{e.message}"
end
```

## C# SDK

### Installation

```bash
dotnet add package KYBPlatform.SDK
```

### Basic Usage

```csharp
using KYBPlatform;
using KYBPlatform.Models;

class Program
{
    static async Task Main(string[] args)
    {
        // Initialize the client
        var client = new KYBPlatformClient(new KYBPlatformConfig
        {
            BaseUrl = "https://api.kybplatform.com",
            AccessToken = "your-access-token"
        });
        
        // Authenticate
        try
        {
            var loginRequest = new LoginRequest
            {
                Email = "your-email@example.com",
                Password = "your-password"
            };
            
            var response = await client.Auth.LoginAsync(loginRequest);
            Console.WriteLine($"Access token: {response.AccessToken}");
        }
        catch (Exception e)
        {
            Console.WriteLine($"Authentication failed: {e.Message}");
        }
        
        // Business Classification
        try
        {
            var request = new ClassificationRequest
            {
                BusinessName = "Acme Corporation",
                BusinessAddress = "123 Main St, New York, NY 10001",
                BusinessPhone = "+1-555-123-4567",
                BusinessWebsite = "https://acme.com"
            };
            
            var classification = await client.Classification.ClassifyAsync(request);
            Console.WriteLine($"Classification: {classification}");
        }
        catch (Exception e)
        {
            Console.WriteLine($"Classification failed: {e.Message}");
        }
        
        // Risk Assessment
        try
        {
            var request = new RiskAssessmentRequest
            {
                BusinessId = "business-123",
                BusinessName = "Acme Corporation",
                Categories = new[] { "financial", "operational" }
            };
            
            var assessment = await client.Risk.AssessAsync(request);
            Console.WriteLine($"Risk Assessment: {assessment}");
        }
        catch (Exception e)
        {
            Console.WriteLine($"Risk assessment failed: {e.Message}");
        }
        
        // Compliance Check
        try
        {
            var request = new ComplianceCheckRequest
            {
                BusinessId = "business-123",
                Frameworks = new[] { "SOC2", "PCI_DSS", "GDPR" }
            };
            
            var compliance = await client.Compliance.CheckAsync(request);
            Console.WriteLine($"Compliance Status: {compliance}");
        }
        catch (Exception e)
        {
            Console.WriteLine($"Compliance check failed: {e.Message}");
        }
    }
}
```

## Best Practices

### 1. Error Handling

Always implement proper error handling:

```javascript
try {
    const result = await client.classification.classify(request);
    // Handle success
} catch (error) {
    if (error.status === 401) {
        // Handle authentication error
        await refreshToken();
    } else if (error.status === 429) {
        // Handle rate limiting
        await delay(error.retryAfter * 1000);
    } else {
        // Handle other errors
        console.error('API Error:', error.message);
    }
}
```

### 2. Retry Logic

Implement exponential backoff for retries:

```javascript
async function withRetry(fn, maxRetries = 3) {
    for (let i = 0; i < maxRetries; i++) {
        try {
            return await fn();
        } catch (error) {
            if (error.status === 429 && i < maxRetries - 1) {
                const delay = Math.pow(2, i) * 1000;
                await new Promise(resolve => setTimeout(resolve, delay));
                continue;
            }
            throw error;
        }
    }
}
```

### 3. Token Management

Implement automatic token refresh:

```javascript
class KYBClient {
    constructor(config) {
        this.config = config;
        this.accessToken = config.accessToken;
        this.refreshToken = config.refreshToken;
    }
    
    async refreshAccessToken() {
        const response = await this.auth.refresh({
            refresh_token: this.refreshToken
        });
        
        this.accessToken = response.access_token;
        this.refreshToken = response.refresh_token;
        
        return this.accessToken;
    }
    
    async makeRequest(endpoint, options = {}) {
        try {
            return await this.request(endpoint, {
                ...options,
                headers: {
                    ...options.headers,
                    'Authorization': `Bearer ${this.accessToken}`
                }
            });
        } catch (error) {
            if (error.status === 401) {
                await this.refreshAccessToken();
                return await this.request(endpoint, {
                    ...options,
                    headers: {
                        ...options.headers,
                        'Authorization': `Bearer ${this.accessToken}`
                    }
                });
            }
            throw error;
        }
    }
}
```

### 4. Logging

Implement proper logging:

```javascript
class KYBClient {
    constructor(config) {
        this.logger = config.logger || console;
    }
    
    async classify(request) {
        this.logger.info('Starting business classification', {
            business_name: request.business_name,
            timestamp: new Date().toISOString()
        });
        
        try {
            const result = await this.makeRequest('/v1/classify', {
                method: 'POST',
                body: JSON.stringify(request)
            });
            
            this.logger.info('Classification completed successfully', {
                business_name: request.business_name,
                classification_id: result.classification_id,
                confidence_score: result.confidence_score
            });
            
            return result;
        } catch (error) {
            this.logger.error('Classification failed', {
                business_name: request.business_name,
                error: error.message,
                status: error.status
            });
            throw error;
        }
    }
}
```

## Error Handling

### Common Error Codes

| Status Code | Error Type | Description | Resolution |
|-------------|------------|-------------|------------|
| 400 | `validation_error` | Invalid request data | Check request format and required fields |
| 401 | `unauthorized` | Invalid or expired token | Refresh authentication token |
| 403 | `insufficient_permissions` | Insufficient permissions | Contact support for access |
| 404 | `resource_not_found` | Resource not found | Verify resource ID |
| 429 | `rate_limit_exceeded` | Rate limit exceeded | Implement exponential backoff |
| 500 | `internal_server_error` | Server error | Retry with exponential backoff |
| 503 | `service_unavailable` | Service temporarily unavailable | Retry after delay |

### Error Response Format

```json
{
  "error": "validation_error",
  "message": "Invalid request data",
  "status_code": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "field": "business_name",
    "issue": "Required field is missing"
  },
  "retry_after": null
}
```

## Rate Limiting

### Limits

- **Authenticated users**: 100 requests per minute
- **Unauthenticated users**: 10 requests per minute

### Headers

- `X-RateLimit-Limit`: Request limit per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the rate limit resets (Unix timestamp)
- `Retry-After`: Seconds to wait before retrying (when rate limited)

### Implementation

```javascript
class RateLimiter {
    constructor() {
        this.requests = [];
    }
    
    async checkLimit() {
        const now = Date.now();
        const window = 60 * 1000; // 1 minute
        
        // Remove old requests
        this.requests = this.requests.filter(time => now - time < window);
        
        if (this.requests.length >= 100) {
            const oldestRequest = this.requests[0];
            const waitTime = window - (now - oldestRequest);
            throw new Error(`Rate limit exceeded. Wait ${waitTime}ms`);
        }
        
        this.requests.push(now);
    }
}
```

## Testing

### Unit Tests

```javascript
// Using Jest
describe('KYB Platform SDK', () => {
    let client;
    
    beforeEach(() => {
        client = new KYBPlatform({
            baseURL: 'https://api.kybplatform.com',
            accessToken: 'test-token'
        });
    });
    
    test('should classify business successfully', async () => {
        const mockResponse = {
            classification_id: 'class-123',
            business_name: 'Acme Corp',
            naics_code: '541511',
            confidence_score: 0.95
        };
        
        // Mock the API call
        client.classification.classify = jest.fn().mockResolvedValue(mockResponse);
        
        const result = await client.classification.classify({
            business_name: 'Acme Corp'
        });
        
        expect(result).toEqual(mockResponse);
        expect(client.classification.classify).toHaveBeenCalledWith({
            business_name: 'Acme Corp'
        });
    });
    
    test('should handle authentication errors', async () => {
        const error = new Error('Unauthorized');
        error.status = 401;
        
        client.auth.login = jest.fn().mockRejectedValue(error);
        
        await expect(client.auth.login({
            email: 'test@example.com',
            password: 'password'
        })).rejects.toThrow('Unauthorized');
    });
});
```

### Integration Tests

```javascript
describe('KYB Platform Integration Tests', () => {
    let client;
    
    beforeAll(async () => {
        client = new KYBPlatform({
            baseURL: process.env.KYB_API_URL || 'https://api.kybplatform.com'
        });
        
        // Authenticate for tests
        const response = await client.auth.login({
            email: process.env.KYB_TEST_EMAIL,
            password: process.env.KYB_TEST_PASSWORD
        });
        
        client.setAccessToken(response.access_token);
    });
    
    test('should perform end-to-end classification', async () => {
        const classification = await client.classification.classify({
            business_name: 'Test Business',
            business_address: '123 Test St, Test City, TS 12345'
        });
        
        expect(classification).toHaveProperty('classification_id');
        expect(classification).toHaveProperty('naics_code');
        expect(classification).toHaveProperty('confidence_score');
        expect(classification.confidence_score).toBeGreaterThan(0);
        expect(classification.confidence_score).toBeLessThanOrEqual(1);
    });
});
```

## Support

For SDK support and questions:

- **Documentation**: https://docs.kybplatform.com
- **GitHub**: https://github.com/kybplatform
- **Email**: support@kybplatform.com
- **Discord**: https://discord.gg/kybplatform

## Changelog

### Version 1.0.0
- Initial SDK release
- Support for business classification
- Support for risk assessment
- Support for compliance checking
- Authentication and token management
- Error handling and retry logic
- Rate limiting support
