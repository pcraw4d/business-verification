# Migration Guides

## Overview

This document provides comprehensive migration guides for upgrading between different versions of the Risk Assessment Service. Each guide includes step-by-step instructions, code examples, and best practices for a smooth migration experience.

## Table of Contents

1. [Migration from v2.x to v3.0.0](#migration-from-v2x-to-v300)
2. [Migration from v1.x to v2.0.0](#migration-from-v1x-to-v200)
3. [SDK Migration Examples](#sdk-migration-examples)
4. [Database Migration](#database-migration)
5. [Configuration Migration](#configuration-migration)
6. [Authentication Migration](#authentication-migration)
7. [API Endpoint Migration](#api-endpoint-migration)
8. [Webhook Migration](#webhook-migration)
9. [Troubleshooting Migration Issues](#troubleshooting-migration-issues)

## Migration from v2.x to v3.0.0

### Overview

Version 3.0.0 introduces significant improvements including advanced ML models, enhanced API structure, and improved authentication. This migration requires code changes due to breaking changes in the API.

### Breaking Changes

#### 1. API Version Change

**Before (v2.x):**
```bash
curl -X POST https://api.kyb-platform.com/v2/assess \
  -H "X-API-Key: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Example Corp"}'
```

**After (v3.0.0):**
```bash
curl -X POST https://api.kyb-platform.com/v3/assess \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Example Corp"}'
```

#### 2. Authentication Method Change

**Before (v2.x) - API Key:**
```javascript
const client = new RiskAssessmentClient('your_api_key');
```

**After (v3.0.0) - JWT Token:**
```javascript
const client = new RiskAssessmentClient({
  apiKey: 'your_jwt_token',
  baseURL: 'https://api.kyb-platform.com/v3'
});
```

#### 3. Response Format Changes

**Before (v2.x):**
```json
{
  "risk_score": 75,
  "risk_level": "medium",
  "business_name": "Example Corp",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**After (v3.0.0):**
```json
{
  "id": "risk_abc123def456",
  "business_name": "Example Corp",
  "risk_score": 0.75,
  "risk_level": "medium",
  "confidence": 0.89,
  "model_used": "xgboost",
  "created_at": "2024-01-15T10:30:00Z",
  "metadata": {
    "version": "3.0.0",
    "processing_time_ms": 245
  }
}
```

#### 4. Risk Score Scale Change

**Before (v2.x) - 0-100 scale:**
```javascript
if (response.risk_score > 70) {
  // High risk
}
```

**After (v3.0.0) - 0.0-1.0 scale:**
```javascript
if (response.risk_score > 0.7) {
  // High risk
}
```

### Migration Steps

#### Step 1: Update API Endpoints

1. **Change Base URL:**
   ```javascript
   // Before
   const baseURL = 'https://api.kyb-platform.com/v2';
   
   // After
   const baseURL = 'https://api.kyb-platform.com/v3';
   ```

2. **Update All API Calls:**
   ```javascript
   // Before
   const response = await fetch(`${baseURL}/assess`, {
     method: 'POST',
     headers: {
       'X-API-Key': apiKey,
       'Content-Type': 'application/json'
     },
     body: JSON.stringify(data)
   });
   
   // After
   const response = await fetch(`${baseURL}/assess`, {
     method: 'POST',
     headers: {
       'Authorization': `Bearer ${jwtToken}`,
       'Content-Type': 'application/json'
     },
     body: JSON.stringify(data)
   });
   ```

#### Step 2: Update Authentication

1. **Generate JWT Token:**
   ```javascript
   // Get JWT token from authentication endpoint
   const authResponse = await fetch('https://api.kyb-platform.com/v3/auth/token', {
     method: 'POST',
     headers: {
       'Content-Type': 'application/json'
     },
     body: JSON.stringify({
       api_key: 'your_api_key'
     })
   });
   
   const { access_token, refresh_token, expires_in } = await authResponse.json();
   ```

2. **Implement Token Refresh:**
   ```javascript
   class TokenManager {
     constructor(apiKey) {
       this.apiKey = apiKey;
       this.accessToken = null;
       this.refreshToken = null;
       this.expiresAt = null;
     }
   
     async getValidToken() {
       if (!this.accessToken || this.isTokenExpired()) {
         await this.refreshAccessToken();
       }
       return this.accessToken;
     }
   
     async refreshAccessToken() {
       const response = await fetch('https://api.kyb-platform.com/v3/auth/refresh', {
         method: 'POST',
         headers: {
           'Content-Type': 'application/json'
         },
         body: JSON.stringify({
           refresh_token: this.refreshToken
         })
       });
   
       const { access_token, refresh_token, expires_in } = await response.json();
       this.accessToken = access_token;
       this.refreshToken = refresh_token;
       this.expiresAt = Date.now() + (expires_in * 1000);
     }
   
     isTokenExpired() {
       return Date.now() >= this.expiresAt - 60000; // 1 minute buffer
     }
   }
   ```

#### Step 3: Update Response Handling

1. **Handle New Response Format:**
   ```javascript
   // Before
   function handleResponse(response) {
     const { risk_score, risk_level, business_name } = response;
     console.log(`Risk for ${business_name}: ${risk_score} (${risk_level})`);
   }
   
   // After
   function handleResponse(response) {
     const { 
       id, 
       business_name, 
       risk_score, 
       risk_level, 
       confidence, 
       model_used,
       metadata 
     } = response;
     
     console.log(`Risk for ${business_name}: ${risk_score} (${risk_level})`);
     console.log(`Confidence: ${confidence}, Model: ${model_used}`);
     console.log(`Processing time: ${metadata.processing_time_ms}ms`);
   }
   ```

2. **Update Risk Score Comparisons:**
   ```javascript
   // Before
   function getRiskCategory(score) {
     if (score >= 80) return 'high';
     if (score >= 50) return 'medium';
     return 'low';
   }
   
   // After
   function getRiskCategory(score) {
     if (score >= 0.8) return 'high';
     if (score >= 0.5) return 'medium';
     return 'low';
   }
   ```

#### Step 4: Update Error Handling

1. **Handle New Error Format:**
   ```javascript
   // Before
   try {
     const response = await client.assessRisk(data);
   } catch (error) {
     console.error('Assessment failed:', error.message);
   }
   
   // After
   try {
     const response = await client.assessRisk(data);
   } catch (error) {
     if (error.response) {
       const { status, data } = error.response;
       console.error(`API Error ${status}:`, data.message);
       console.error('Error code:', data.code);
       console.error('Request ID:', data.request_id);
     } else {
       console.error('Network error:', error.message);
     }
   }
   ```

### SDK-Specific Migration

#### Go SDK Migration

**Before (v2.x):**
```go
package main

import (
    "fmt"
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client := kyb.NewClient("your_api_key")
    
    assessment, err := client.AssessRisk("Example Corp", "123 Main St")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Risk Score: %d\n", assessment.RiskScore)
    fmt.Printf("Risk Level: %s\n", assessment.RiskLevel)
}
```

**After (v3.0.0):**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client := kyb.NewClient(&kyb.Config{
        APIKey: "your_jwt_token",
        BaseURL: "https://api.kyb-platform.com/v3",
    })
    
    request := &kyb.RiskAssessmentRequest{
        BusinessName:    "Example Corp",
        BusinessAddress: "123 Main St",
        Industry:        "Technology",
        Country:         "US",
    }
    
    ctx := context.Background()
    assessment, err := client.AssessRisk(ctx, request)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Assessment ID: %s\n", assessment.ID)
    fmt.Printf("Risk Score: %.3f\n", assessment.RiskScore)
    fmt.Printf("Risk Level: %s\n", assessment.RiskLevel)
    fmt.Printf("Confidence: %.3f\n", assessment.Confidence)
    fmt.Printf("Model Used: %s\n", assessment.ModelUsed)
}
```

#### Python SDK Migration

**Before (v2.x):**
```python
from kyb_risk_assessment import RiskAssessmentClient

client = RiskAssessmentClient(api_key="your_api_key")

assessment = client.assess_risk("Example Corp", "123 Main St")

print(f"Risk Score: {assessment.risk_score}")
print(f"Risk Level: {assessment.risk_level}")
```

**After (v3.0.0):**
```python
from kyb_risk_assessment import RiskAssessmentClient

client = RiskAssessmentClient(api_key="your_jwt_token")

request = {
    "business_name": "Example Corp",
    "business_address": "123 Main St",
    "industry": "Technology",
    "country": "US"
}

assessment = client.assess_risk(request)

print(f"Assessment ID: {assessment.id}")
print(f"Risk Score: {assessment.risk_score:.3f}")
print(f"Risk Level: {assessment.risk_level}")
print(f"Confidence: {assessment.confidence:.3f}")
print(f"Model Used: {assessment.model_used}")
```

#### Node.js SDK Migration

**Before (v2.x):**
```javascript
const { RiskAssessmentClient } = require('@kyb-platform/risk-assessment');

const client = new RiskAssessmentClient('your_api_key');

async function assessRisk() {
    const assessment = await client.assessRisk('Example Corp', '123 Main St');
    
    console.log(`Risk Score: ${assessment.risk_score}`);
    console.log(`Risk Level: ${assessment.risk_level}`);
}

assessRisk();
```

**After (v3.0.0):**
```javascript
const { RiskAssessmentClient } = require('@kyb-platform/risk-assessment');

const client = new RiskAssessmentClient({
    apiKey: 'your_jwt_token',
    baseURL: 'https://api.kyb-platform.com/v3'
});

async function assessRisk() {
    const request = {
        businessName: 'Example Corp',
        businessAddress: '123 Main St',
        industry: 'Technology',
        country: 'US'
    };
    
    const assessment = await client.assessRisk(request);
    
    console.log(`Assessment ID: ${assessment.id}`);
    console.log(`Risk Score: ${assessment.riskScore.toFixed(3)}`);
    console.log(`Risk Level: ${assessment.riskLevel}`);
    console.log(`Confidence: ${assessment.confidence.toFixed(3)}`);
    console.log(`Model Used: ${assessment.modelUsed}`);
}

assessRisk();
```

## Migration from v1.x to v2.0.0

### Overview

Version 2.0.0 introduced the first major API redesign with improved structure, better error handling, and enhanced features. This migration requires significant code changes.

### Breaking Changes

#### 1. Authentication Method Change

**Before (v1.x) - Basic Auth:**
```bash
curl -u username:password https://api.kyb-platform.com/v1/assess
```

**After (v2.0.0) - API Key:**
```bash
curl -H "X-API-Key: your_api_key" https://api.kyb-platform.com/v2/assess
```

#### 2. Request Format Changes

**Before (v1.x):**
```json
{
  "company_name": "Example Corp",
  "address": "123 Main St"
}
```

**After (v2.0.0):**
```json
{
  "business_name": "Example Corp",
  "business_address": "123 Main St",
  "industry": "Technology",
  "country": "US"
}
```

#### 3. Response Format Changes

**Before (v1.x):**
```json
{
  "score": 75,
  "level": "medium",
  "company": "Example Corp"
}
```

**After (v2.0.0):**
```json
{
  "risk_score": 75,
  "risk_level": "medium",
  "business_name": "Example Corp",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Migration Steps

#### Step 1: Update Authentication

1. **Generate API Key:**
   - Log into admin dashboard
   - Navigate to API Keys section
   - Generate new API key
   - Store securely

2. **Update Authentication Code:**
   ```javascript
   // Before
   const auth = btoa(`${username}:${password}`);
   const headers = {
     'Authorization': `Basic ${auth}`
   };
   
   // After
   const headers = {
     'X-API-Key': 'your_api_key'
   };
   ```

#### Step 2: Update Request Format

1. **Map Field Names:**
   ```javascript
   // Before
   const requestData = {
     company_name: businessName,
     address: businessAddress
   };
   
   // After
   const requestData = {
     business_name: businessName,
     business_address: businessAddress,
     industry: industry,
     country: country
   };
   ```

2. **Add Required Fields:**
   ```javascript
   // Add industry and country fields
   const requestData = {
     business_name: "Example Corp",
     business_address: "123 Main St",
     industry: "Technology", // New required field
     country: "US"           // New required field
   };
   ```

#### Step 3: Update Response Handling

1. **Map Response Fields:**
   ```javascript
   // Before
   const { score, level, company } = response;
   
   // After
   const { risk_score, risk_level, business_name } = response;
   ```

2. **Handle New Fields:**
   ```javascript
   // Handle new timestamp field
   const { 
     risk_score, 
     risk_level, 
     business_name, 
     created_at 
   } = response;
   
   console.log(`Assessment completed at: ${created_at}`);
   ```

## SDK Migration Examples

### Ruby SDK Migration

**Before (v2.x):**
```ruby
require 'kyb-risk-assessment'

client = KYB::RiskAssessmentClient.new('your_api_key')

assessment = client.assess_risk('Example Corp', '123 Main St')

puts "Risk Score: #{assessment.risk_score}"
puts "Risk Level: #{assessment.risk_level}"
```

**After (v3.0.0):**
```ruby
require 'kyb-risk-assessment'

client = KYB::RiskAssessmentClient.new(
  api_key: 'your_jwt_token',
  base_url: 'https://api.kyb-platform.com/v3'
)

request = {
  business_name: 'Example Corp',
  business_address: '123 Main St',
  industry: 'Technology',
  country: 'US'
}

assessment = client.assess_risk(request)

puts "Assessment ID: #{assessment.id}"
puts "Risk Score: #{assessment.risk_score}"
puts "Risk Level: #{assessment.risk_level}"
puts "Confidence: #{assessment.confidence}"
puts "Model Used: #{assessment.model_used}"
```

### Java SDK Migration

**Before (v2.x):**
```java
import com.kyb.riskassessment.RiskAssessmentClient;

RiskAssessmentClient client = new RiskAssessmentClient("your_api_key");

Assessment assessment = client.assessRisk("Example Corp", "123 Main St");

System.out.println("Risk Score: " + assessment.getRiskScore());
System.out.println("Risk Level: " + assessment.getRiskLevel());
```

**After (v3.0.0):**
```java
import com.kyb.riskassessment.RiskAssessmentClient;
import com.kyb.riskassessment.models.RiskAssessmentRequest;

RiskAssessmentClient client = new RiskAssessmentClient.Builder()
    .apiKey("your_jwt_token")
    .baseUrl("https://api.kyb-platform.com/v3")
    .build();

RiskAssessmentRequest request = RiskAssessmentRequest.builder()
    .businessName("Example Corp")
    .businessAddress("123 Main St")
    .industry("Technology")
    .country("US")
    .build();

Assessment assessment = client.assessRisk(request);

System.out.println("Assessment ID: " + assessment.getId());
System.out.println("Risk Score: " + assessment.getRiskScore());
System.out.println("Risk Level: " + assessment.getRiskLevel());
System.out.println("Confidence: " + assessment.getConfidence());
System.out.println("Model Used: " + assessment.getModelUsed());
```

### PHP SDK Migration

**Before (v2.x):**
```php
<?php
require_once 'vendor/autoload.php';

use KYB\RiskAssessment\RiskAssessmentClient;

$client = new RiskAssessmentClient('your_api_key');

$assessment = $client->assessRisk('Example Corp', '123 Main St');

echo "Risk Score: " . $assessment->risk_score . "\n";
echo "Risk Level: " . $assessment->risk_level . "\n";
?>
```

**After (v3.0.0):**
```php
<?php
require_once 'vendor/autoload.php';

use KYB\RiskAssessment\RiskAssessmentClient;
use KYB\RiskAssessment\Models\RiskAssessmentRequest;

$client = new RiskAssessmentClient([
    'api_key' => 'your_jwt_token',
    'base_url' => 'https://api.kyb-platform.com/v3'
]);

$request = new RiskAssessmentRequest([
    'business_name' => 'Example Corp',
    'business_address' => '123 Main St',
    'industry' => 'Technology',
    'country' => 'US'
]);

$assessment = $client->assessRisk($request);

echo "Assessment ID: " . $assessment->id . "\n";
echo "Risk Score: " . $assessment->risk_score . "\n";
echo "Risk Level: " . $assessment->risk_level . "\n";
echo "Confidence: " . $assessment->confidence . "\n";
echo "Model Used: " . $assessment->model_used . "\n";
?>
```

## Database Migration

### Schema Changes

#### v2.0.0 Schema Changes

1. **New Tables:**
   ```sql
   -- Risk factors table
   CREATE TABLE risk_factors (
       id UUID PRIMARY KEY,
       assessment_id UUID REFERENCES risk_assessments(id),
       category VARCHAR(50) NOT NULL,
       name VARCHAR(100) NOT NULL,
       score DECIMAL(3,2) NOT NULL,
       weight DECIMAL(3,2) NOT NULL,
       created_at TIMESTAMP DEFAULT NOW()
   );
   
   -- Predictions table
   CREATE TABLE predictions (
       id UUID PRIMARY KEY,
       assessment_id UUID REFERENCES risk_assessments(id),
       horizon_months INTEGER NOT NULL,
       predicted_score DECIMAL(3,2) NOT NULL,
       confidence DECIMAL(3,2) NOT NULL,
       model_used VARCHAR(50) NOT NULL,
       created_at TIMESTAMP DEFAULT NOW()
   );
   ```

2. **Modified Tables:**
   ```sql
   -- Add new columns to risk_assessments
   ALTER TABLE risk_assessments 
   ADD COLUMN confidence DECIMAL(3,2),
   ADD COLUMN model_used VARCHAR(50),
   ADD COLUMN metadata JSONB;
   ```

#### v3.0.0 Schema Changes

1. **New Tables:**
   ```sql
   -- Compliance checks table
   CREATE TABLE compliance_checks (
       id UUID PRIMARY KEY,
       assessment_id UUID REFERENCES risk_assessments(id),
       check_type VARCHAR(50) NOT NULL,
       status VARCHAR(20) NOT NULL,
       result JSONB,
       checked_at TIMESTAMP DEFAULT NOW()
   );
   
   -- Webhook events table
   CREATE TABLE webhook_events (
       id UUID PRIMARY KEY,
       webhook_url VARCHAR(500) NOT NULL,
       event_type VARCHAR(50) NOT NULL,
       payload JSONB NOT NULL,
       status VARCHAR(20) NOT NULL,
       attempts INTEGER DEFAULT 0,
       created_at TIMESTAMP DEFAULT NOW()
   );
   ```

2. **Indexes:**
   ```sql
   -- Performance indexes
   CREATE INDEX idx_risk_assessments_user_created ON risk_assessments(user_id, created_at DESC);
   CREATE INDEX idx_risk_factors_assessment_category ON risk_factors(assessment_id, category);
   CREATE INDEX idx_predictions_assessment_horizon ON predictions(assessment_id, horizon_months);
   ```

### Migration Scripts

#### v2.0.0 Migration Script

```sql
-- Migration script for v2.0.0
BEGIN;

-- Create new tables
CREATE TABLE IF NOT EXISTS risk_factors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assessment_id UUID REFERENCES risk_assessments(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    score DECIMAL(3,2) NOT NULL,
    weight DECIMAL(3,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assessment_id UUID REFERENCES risk_assessments(id) ON DELETE CASCADE,
    horizon_months INTEGER NOT NULL,
    predicted_score DECIMAL(3,2) NOT NULL,
    confidence DECIMAL(3,2) NOT NULL,
    model_used VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Add new columns
ALTER TABLE risk_assessments 
ADD COLUMN IF NOT EXISTS confidence DECIMAL(3,2),
ADD COLUMN IF NOT EXISTS model_used VARCHAR(50),
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_risk_factors_assessment_id ON risk_factors(assessment_id);
CREATE INDEX IF NOT EXISTS idx_predictions_assessment_id ON predictions(assessment_id);

COMMIT;
```

#### v3.0.0 Migration Script

```sql
-- Migration script for v3.0.0
BEGIN;

-- Create compliance checks table
CREATE TABLE IF NOT EXISTS compliance_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assessment_id UUID REFERENCES risk_assessments(id) ON DELETE CASCADE,
    check_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    result JSONB,
    checked_at TIMESTAMP DEFAULT NOW()
);

-- Create webhook events table
CREATE TABLE IF NOT EXISTS webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_url VARCHAR(500) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL,
    attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create performance indexes
CREATE INDEX IF NOT EXISTS idx_risk_assessments_user_created ON risk_assessments(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_factors_assessment_category ON risk_factors(assessment_id, category);
CREATE INDEX IF NOT EXISTS idx_predictions_assessment_horizon ON predictions(assessment_id, horizon_months);
CREATE INDEX IF NOT EXISTS idx_compliance_checks_assessment_type ON compliance_checks(assessment_id, check_type);
CREATE INDEX IF NOT EXISTS idx_webhook_events_status_created ON webhook_events(status, created_at);

COMMIT;
```

## Configuration Migration

### Environment Variables

#### v2.0.0 Configuration Changes

**Before (v1.x):**
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=risk_assessment
DB_USER=postgres
DB_PASSWORD=password

# API
API_PORT=8080
API_KEY_SECRET=your_secret
```

**After (v2.0.0):**
```bash
# Database
DATABASE_URL=postgresql://postgres:password@localhost:5432/risk_assessment

# API
PORT=8080
API_KEY_SECRET=your_secret
JWT_SECRET=your_jwt_secret
JWT_EXPIRY=24h

# External APIs
NEWS_API_KEY=your_news_api_key
MARKET_DATA_API_KEY=your_market_api_key
```

#### v3.0.0 Configuration Changes

**Before (v2.x):**
```bash
# Basic configuration
DATABASE_URL=postgresql://postgres:password@localhost:5432/risk_assessment
PORT=8080
JWT_SECRET=your_jwt_secret
```

**After (v3.0.0):**
```bash
# Enhanced configuration
DATABASE_URL=postgresql://postgres:password@localhost:5432/risk_assessment
PORT=8080

# Authentication
JWT_SECRET=your_jwt_secret
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=7d
JWT_ISSUER=kyb-platform
JWT_AUDIENCE=risk-assessment-api

# ML Models
ML_MODEL_PATH=/app/models
XGBOOST_MODEL_VERSION=v2.1.0
LSTM_MODEL_VERSION=v1.8.0
ENSEMBLE_MODEL_VERSION=v1.5.0

# External APIs
THOMSON_REUTERS_API_KEY=your_tr_api_key
OFAC_API_KEY=your_ofac_api_key
CREDIT_BUREAU_API_KEY=your_credit_api_key

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
GRAFANA_ENABLED=true
GRAFANA_PORT=3000

# Caching
REDIS_URL=redis://localhost:6379
CACHE_TTL=15m
CACHE_MAX_SIZE=1000

# Webhooks
WEBHOOK_RETRY_ATTEMPTS=3
WEBHOOK_RETRY_DELAY=5s
WEBHOOK_TIMEOUT=30s
```

### Configuration Migration Script

```bash
#!/bin/bash
# Configuration migration script

echo "Migrating configuration from v2.x to v3.0.0..."

# Backup current configuration
cp .env .env.backup.$(date +%Y%m%d_%H%M%S)

# Add new configuration variables
cat >> .env << EOF

# Enhanced JWT configuration
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=7d
JWT_ISSUER=kyb-platform
JWT_AUDIENCE=risk-assessment-api

# ML Models
ML_MODEL_PATH=/app/models
XGBOOST_MODEL_VERSION=v2.1.0
LSTM_MODEL_VERSION=v1.8.0
ENSEMBLE_MODEL_VERSION=v1.5.0

# External APIs
THOMSON_REUTERS_API_KEY=your_tr_api_key
OFAC_API_KEY=your_ofac_api_key
CREDIT_BUREAU_API_KEY=your_credit_api_key

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
GRAFANA_ENABLED=true
GRAFANA_PORT=3000

# Caching
REDIS_URL=redis://localhost:6379
CACHE_TTL=15m
CACHE_MAX_SIZE=1000

# Webhooks
WEBHOOK_RETRY_ATTEMPTS=3
WEBHOOK_RETRY_DELAY=5s
WEBHOOK_TIMEOUT=30s
EOF

echo "Configuration migration completed!"
echo "Please update the API keys and secrets in the .env file."
```

## Authentication Migration

### JWT Token Migration

#### Step 1: Update Authentication Flow

**Before (v2.x) - API Key:**
```javascript
class APIClient {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async makeRequest(endpoint, data) {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'X-API-Key': this.apiKey,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });
    
    return response.json();
  }
}
```

**After (v3.0.0) - JWT Token:**
```javascript
class APIClient {
  constructor(config) {
    this.baseURL = config.baseURL;
    this.apiKey = config.apiKey;
    this.accessToken = null;
    this.refreshToken = null;
    this.tokenExpiry = null;
  }
  
  async getValidToken() {
    if (!this.accessToken || this.isTokenExpired()) {
      await this.refreshAccessToken();
    }
    return this.accessToken;
  }
  
  async refreshAccessToken() {
    const response = await fetch(`${this.baseURL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        refresh_token: this.refreshToken
      })
    });
    
    const { access_token, refresh_token, expires_in } = await response.json();
    this.accessToken = access_token;
    this.refreshToken = refresh_token;
    this.tokenExpiry = Date.now() + (expires_in * 1000);
  }
  
  isTokenExpired() {
    return Date.now() >= this.tokenExpiry - 60000; // 1 minute buffer
  }
  
  async makeRequest(endpoint, data) {
    const token = await this.getValidToken();
    
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });
    
    return response.json();
  }
}
```

#### Step 2: Implement Token Storage

```javascript
class TokenStorage {
  constructor() {
    this.storageKey = 'kyb_tokens';
  }
  
  saveTokens(tokens) {
    const tokenData = {
      access_token: tokens.access_token,
      refresh_token: tokens.refresh_token,
      expires_at: Date.now() + (tokens.expires_in * 1000)
    };
    
    localStorage.setItem(this.storageKey, JSON.stringify(tokenData));
  }
  
  getTokens() {
    const tokenData = localStorage.getItem(this.storageKey);
    return tokenData ? JSON.parse(tokenData) : null;
  }
  
  clearTokens() {
    localStorage.removeItem(this.storageKey);
  }
  
  isTokenValid() {
    const tokens = this.getTokens();
    if (!tokens) return false;
    
    return Date.now() < tokens.expires_at;
  }
}
```

## API Endpoint Migration

### Endpoint Changes

#### v2.0.0 Endpoint Changes

**Before (v1.x):**
```
POST /v1/assess
GET /v1/status
GET /v1/health
```

**After (v2.0.0):**
```
POST /v2/assess
GET /v2/assessments/{id}
GET /v2/assessments
GET /v2/health
GET /v2/status
```

#### v3.0.0 Endpoint Changes

**Before (v2.x):**
```
POST /v2/assess
GET /v2/assessments/{id}
GET /v2/assessments
```

**After (v3.0.0):**
```
POST /v3/assess
GET /v3/assessments/{id}
GET /v3/assessments
POST /v3/predict
GET /v3/predictions/{id}
POST /v3/batch/assess
GET /v3/webhooks
POST /v3/webhooks
```

### Endpoint Migration Script

```javascript
// Endpoint migration utility
class EndpointMigrator {
  constructor(fromVersion, toVersion) {
    this.fromVersion = fromVersion;
    this.toVersion = toVersion;
    this.endpointMap = this.createEndpointMap();
  }
  
  createEndpointMap() {
    return {
      'v1': {
        'v2': {
          '/v1/assess': '/v2/assess',
          '/v1/status': '/v2/status',
          '/v1/health': '/v2/health'
        }
      },
      'v2': {
        'v3': {
          '/v2/assess': '/v3/assess',
          '/v2/assessments/{id}': '/v3/assessments/{id}',
          '/v2/assessments': '/v3/assessments'
        }
      }
    };
  }
  
  migrateEndpoint(endpoint) {
    const versionMap = this.endpointMap[this.fromVersion]?.[this.toVersion];
    if (!versionMap) {
      throw new Error(`No migration path from ${this.fromVersion} to ${this.toVersion}`);
    }
    
    return versionMap[endpoint] || endpoint;
  }
  
  migrateRequest(endpoint, requestData) {
    const newEndpoint = this.migrateEndpoint(endpoint);
    
    // Apply request data transformations
    if (this.fromVersion === 'v1' && this.toVersion === 'v2') {
      return this.migrateV1ToV2Request(requestData);
    } else if (this.fromVersion === 'v2' && this.toVersion === 'v3') {
      return this.migrateV2ToV3Request(requestData);
    }
    
    return { endpoint: newEndpoint, data: requestData };
  }
  
  migrateV1ToV2Request(data) {
    return {
      business_name: data.company_name,
      business_address: data.address,
      industry: data.industry || 'Unknown',
      country: data.country || 'US'
    };
  }
  
  migrateV2ToV3Request(data) {
    // v2 to v3 requests are mostly compatible
    return data;
  }
}
```

## Webhook Migration

### Webhook Format Changes

#### v2.0.0 Webhook Changes

**Before (v1.x):**
```json
{
  "event": "assessment_completed",
  "data": {
    "id": "123",
    "score": 75,
    "level": "medium"
  }
}
```

**After (v2.0.0):**
```json
{
  "event_type": "assessment.completed",
  "event_id": "evt_abc123",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "id": "risk_abc123def456",
    "risk_score": 75,
    "risk_level": "medium",
    "business_name": "Example Corp"
  }
}
```

#### v3.0.0 Webhook Changes

**Before (v2.x):**
```json
{
  "event_type": "assessment.completed",
  "event_id": "evt_abc123",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "id": "risk_abc123def456",
    "risk_score": 75,
    "risk_level": "medium"
  }
}
```

**After (v3.0.0):**
```json
{
  "event_type": "assessment.completed",
  "event_id": "evt_abc123",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "3.0.0",
  "data": {
    "id": "risk_abc123def456",
    "risk_score": 0.75,
    "risk_level": "medium",
    "confidence": 0.89,
    "model_used": "xgboost",
    "business_name": "Example Corp",
    "metadata": {
      "processing_time_ms": 245,
      "features_used": 150
    }
  }
}
```

### Webhook Migration Script

```javascript
// Webhook migration utility
class WebhookMigrator {
  constructor(fromVersion, toVersion) {
    this.fromVersion = fromVersion;
    this.toVersion = toVersion;
  }
  
  migrateWebhook(webhookData) {
    if (this.fromVersion === 'v1' && this.toVersion === 'v2') {
      return this.migrateV1ToV2Webhook(webhookData);
    } else if (this.fromVersion === 'v2' && this.toVersion === 'v3') {
      return this.migrateV2ToV3Webhook(webhookData);
    }
    
    return webhookData;
  }
  
  migrateV1ToV2Webhook(data) {
    return {
      event_type: this.mapEventType(data.event),
      event_id: this.generateEventId(),
      timestamp: new Date().toISOString(),
      data: {
        id: data.data.id,
        risk_score: data.data.score,
        risk_level: data.data.level,
        business_name: data.data.business_name || 'Unknown'
      }
    };
  }
  
  migrateV2ToV3Webhook(data) {
    return {
      ...data,
      version: '3.0.0',
      data: {
        ...data.data,
        risk_score: data.data.risk_score / 100, // Convert to 0.0-1.0 scale
        confidence: 0.85, // Default confidence
        model_used: 'xgboost', // Default model
        metadata: {
          processing_time_ms: 200,
          features_used: 150
        }
      }
    };
  }
  
  mapEventType(event) {
    const eventMap = {
      'assessment_completed': 'assessment.completed',
      'assessment_failed': 'assessment.failed',
      'assessment_started': 'assessment.started'
    };
    
    return eventMap[event] || event;
  }
  
  generateEventId() {
    return 'evt_' + Math.random().toString(36).substr(2, 9);
  }
}
```

## Troubleshooting Migration Issues

### Common Migration Issues

#### 1. Authentication Errors

**Issue**: 401 Unauthorized errors after migration

**Solution**:
```javascript
// Check token validity
async function validateToken(token) {
  try {
    const response = await fetch('https://api.kyb-platform.com/v3/auth/validate', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Token validation failed');
    }
    
    return true;
  } catch (error) {
    console.error('Token validation error:', error);
    return false;
  }
}
```

#### 2. Response Format Errors

**Issue**: Parsing errors due to response format changes

**Solution**:
```javascript
// Robust response parsing
function parseResponse(response) {
  try {
    const data = response.json();
    
    // Handle different response formats
    if (data.risk_score !== undefined) {
      // v3.0.0 format
      return {
        id: data.id,
        riskScore: data.risk_score,
        riskLevel: data.risk_level,
        confidence: data.confidence,
        modelUsed: data.model_used
      };
    } else if (data.score !== undefined) {
      // v1.x format
      return {
        id: data.id,
        riskScore: data.score / 100, // Convert to 0.0-1.0
        riskLevel: data.level,
        confidence: 0.85, // Default
        modelUsed: 'legacy'
      };
    }
    
    throw new Error('Unknown response format');
  } catch (error) {
    console.error('Response parsing error:', error);
    throw error;
  }
}
```

#### 3. Database Migration Errors

**Issue**: Database schema migration failures

**Solution**:
```sql
-- Check for existing columns before adding
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'risk_assessments' 
                   AND column_name = 'confidence') THEN
        ALTER TABLE risk_assessments ADD COLUMN confidence DECIMAL(3,2);
    END IF;
END $$;
```

#### 4. Configuration Issues

**Issue**: Missing environment variables

**Solution**:
```bash
#!/bin/bash
# Configuration validation script

required_vars=(
  "DATABASE_URL"
  "JWT_SECRET"
  "API_PORT"
)

missing_vars=()

for var in "${required_vars[@]}"; do
  if [ -z "${!var}" ]; then
    missing_vars+=("$var")
  fi
done

if [ ${#missing_vars[@]} -ne 0 ]; then
  echo "Missing required environment variables:"
  printf '%s\n' "${missing_vars[@]}"
  exit 1
fi

echo "All required environment variables are set."
```

### Migration Testing

#### Automated Migration Tests

```javascript
// Migration test suite
describe('API Migration Tests', () => {
  test('v2 to v3 authentication migration', async () => {
    const client = new RiskAssessmentClient({
      apiKey: 'test_jwt_token',
      baseURL: 'https://api.kyb-platform.com/v3'
    });
    
    const response = await client.assessRisk({
      business_name: 'Test Corp',
      business_address: '123 Test St',
      industry: 'Technology',
      country: 'US'
    });
    
    expect(response.id).toBeDefined();
    expect(response.risk_score).toBeGreaterThanOrEqual(0);
    expect(response.risk_score).toBeLessThanOrEqual(1);
    expect(response.confidence).toBeDefined();
    expect(response.model_used).toBeDefined();
  });
  
  test('response format compatibility', () => {
    const v2Response = {
      risk_score: 75,
      risk_level: 'medium',
      business_name: 'Test Corp'
    };
    
    const v3Response = migrateV2ToV3Response(v2Response);
    
    expect(v3Response.risk_score).toBe(0.75);
    expect(v3Response.risk_level).toBe('medium');
    expect(v3Response.confidence).toBeDefined();
    expect(v3Response.model_used).toBeDefined();
  });
});
```

### Migration Support

#### Getting Help

1. **Documentation**: Check the latest documentation for your version
2. **Migration Guides**: Follow the step-by-step migration guides
3. **Support Team**: Contact support for migration assistance
4. **Community**: Join the community forum for peer support

#### Support Channels

- **Email**: [migration@kyb-platform.com](mailto:migration@kyb-platform.com)
- **Documentation**: [https://docs.kyb-platform.com/migration](https://docs.kyb-platform.com/migration)
- **GitHub Issues**: [https://github.com/kyb-platform/risk-assessment-service/issues](https://github.com/kyb-platform/risk-assessment-service/issues)
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

---

**Last Updated**: January 15, 2024  
**Next Review**: April 15, 2024
