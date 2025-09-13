# Merchant Portfolio API Integration Guide

**Version**: 1.0.0  
**Last Updated**: January 2025  

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication Setup](#authentication-setup)
3. [Basic Integration Examples](#basic-integration-examples)
4. [Advanced Integration Patterns](#advanced-integration-patterns)
5. [Error Handling Best Practices](#error-handling-best-practices)
6. [Performance Optimization](#performance-optimization)
7. [Testing Your Integration](#testing-your-integration)
8. [Troubleshooting](#troubleshooting)

## Getting Started

### Prerequisites

- Valid API credentials (API key or JWT token)
- HTTP client library (axios, requests, curl, etc.)
- Understanding of REST API principles
- Basic knowledge of JSON data format

### Base URL

```
Production: https://api.kyb-platform.com/v1
Staging: https://staging-api.kyb-platform.com/v1
```

### API Versioning

The API uses URL-based versioning. Current version is `v1`. Future versions will be available at `/v2`, `/v3`, etc.

## Authentication Setup

### JWT Token Authentication

All API requests require a valid JWT token in the Authorization header:

```http
Authorization: Bearer <your-jwt-token>
```

### Getting Your API Token

1. Log into the KYB Platform dashboard
2. Navigate to Settings > API Keys
3. Generate a new API key
4. Copy the token for use in your application

### Token Refresh

JWT tokens expire after 24 hours. Implement token refresh logic:

```javascript
// JavaScript example
class APIClient {
  constructor(baseURL, apiKey) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
    this.token = null;
    this.tokenExpiry = null;
  }

  async getValidToken() {
    if (!this.token || this.isTokenExpired()) {
      await this.refreshToken();
    }
    return this.token;
  }

  isTokenExpired() {
    return this.tokenExpiry && Date.now() >= this.tokenExpiry;
  }

  async refreshToken() {
    // Implement token refresh logic
    const response = await fetch('/auth/refresh', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`
      }
    });
    
    const data = await response.json();
    this.token = data.access_token;
    this.tokenExpiry = Date.now() + (data.expires_in * 1000);
  }
}
```

## Basic Integration Examples

### 1. Creating a Merchant

```javascript
// JavaScript with axios
const axios = require('axios');

async function createMerchant(merchantData) {
  try {
    const response = await axios.post(
      'https://api.kyb-platform.com/v1/merchants',
      merchantData,
      {
        headers: {
          'Authorization': `Bearer ${process.env.API_TOKEN}`,
          'Content-Type': 'application/json'
        }
      }
    );
    
    console.log('Merchant created:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error creating merchant:', error.response?.data || error.message);
    throw error;
  }
}

// Example usage
const newMerchant = await createMerchant({
  name: "TechStart Inc",
  legal_name: "TechStart Incorporated",
  industry: "Technology",
  portfolio_type: "prospective",
  risk_level: "medium",
  address: {
    street1: "456 Innovation Drive",
    city: "Austin",
    state: "TX",
    postal_code: "78701",
    country: "United States",
    country_code: "US"
  },
  contact_info: {
    email: "info@techstart.com",
    phone: "+1-512-555-0123"
  }
});
```

```python
# Python with requests
import requests
import os

def create_merchant(merchant_data):
    url = "https://api.kyb-platform.com/v1/merchants"
    headers = {
        "Authorization": f"Bearer {os.getenv('API_TOKEN')}",
        "Content-Type": "application/json"
    }
    
    try:
        response = requests.post(url, json=merchant_data, headers=headers)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error creating merchant: {e.response.json() if e.response else str(e)}")
        raise

# Example usage
new_merchant = create_merchant({
    "name": "TechStart Inc",
    "legal_name": "TechStart Incorporated",
    "industry": "Technology",
    "portfolio_type": "prospective",
    "risk_level": "medium",
    "address": {
        "street1": "456 Innovation Drive",
        "city": "Austin",
        "state": "TX",
        "postal_code": "78701",
        "country": "United States",
        "country_code": "US"
    },
    "contact_info": {
        "email": "info@techstart.com",
        "phone": "+1-512-555-0123"
    }
})
```

### 2. Searching Merchants

```javascript
// Advanced search with filters
async function searchMerchants(filters = {}) {
  try {
    const response = await axios.get(
      'https://api.kyb-platform.com/v1/merchants',
      {
        params: {
          page: filters.page || 1,
          page_size: filters.page_size || 20,
          portfolio_type: filters.portfolio_type,
          risk_level: filters.risk_level,
          industry: filters.industry,
          search: filters.search
        },
        headers: {
          'Authorization': `Bearer ${process.env.API_TOKEN}`
        }
      }
    );
    
    return response.data;
  } catch (error) {
    console.error('Error searching merchants:', error.response?.data || error.message);
    throw error;
  }
}

// Example usage
const results = await searchMerchants({
  portfolio_type: "onboarded",
  risk_level: "medium",
  industry: "Technology",
  search: "software",
  page: 1,
  page_size: 50
});

console.log(`Found ${results.total} merchants`);
console.log(`Page ${results.page} of ${Math.ceil(results.total / results.page_size)}`);
```

### 3. Bulk Operations

```javascript
// Bulk update portfolio type
async function bulkUpdatePortfolioType(merchantIds, portfolioType) {
  try {
    const response = await axios.post(
      'https://api.kyb-platform.com/v1/merchants/bulk/portfolio-type',
      {
        merchant_ids: merchantIds,
        portfolio_type: portfolioType
      },
      {
        headers: {
          'Authorization': `Bearer ${process.env.API_TOKEN}`,
          'Content-Type': 'application/json'
        }
      }
    );
    
    return response.data;
  } catch (error) {
    console.error('Error in bulk update:', error.response?.data || error.message);
    throw error;
  }
}

// Example usage
const result = await bulkUpdatePortfolioType(
  ["merchant_123", "merchant_456", "merchant_789"],
  "onboarded"
);

console.log(`Updated ${result.successful_updates} out of ${result.total_merchants} merchants`);
```

## Advanced Integration Patterns

### 1. Pagination Handling

```javascript
// Fetch all merchants with automatic pagination
async function getAllMerchants(filters = {}) {
  const allMerchants = [];
  let page = 1;
  let hasMore = true;
  
  while (hasMore) {
    try {
      const response = await axios.get(
        'https://api.kyb-platform.com/v1/merchants',
        {
          params: {
            ...filters,
            page,
            page_size: 100 // Maximum page size
          },
          headers: {
            'Authorization': `Bearer ${process.env.API_TOKEN}`
          }
        }
      );
      
      const data = response.data;
      allMerchants.push(...data.merchants);
      hasMore = data.has_more;
      page++;
      
      // Add delay to respect rate limits
      if (hasMore) {
        await new Promise(resolve => setTimeout(resolve, 100));
      }
    } catch (error) {
      console.error(`Error fetching page ${page}:`, error.response?.data || error.message);
      throw error;
    }
  }
  
  return allMerchants;
}
```

### 2. Session Management

```javascript
// Merchant session management
class MerchantSessionManager {
  constructor(apiClient) {
    this.apiClient = apiClient;
    this.activeSession = null;
  }
  
  async startSession(merchantId) {
    try {
      // End current session if exists
      if (this.activeSession) {
        await this.endCurrentSession();
      }
      
      const response = await this.apiClient.post(
        `/merchants/${merchantId}/session`
      );
      
      this.activeSession = response.data;
      console.log(`Started session for merchant: ${this.activeSession.merchant_name}`);
      return this.activeSession;
    } catch (error) {
      console.error('Error starting session:', error.response?.data || error.message);
      throw error;
    }
  }
  
  async endCurrentSession() {
    if (!this.activeSession) return;
    
    try {
      await this.apiClient.delete(
        `/merchants/${this.activeSession.merchant_id}/session`
      );
      console.log(`Ended session for merchant: ${this.activeSession.merchant_name}`);
      this.activeSession = null;
    } catch (error) {
      console.error('Error ending session:', error.response?.data || error.message);
      throw error;
    }
  }
  
  async getActiveSession() {
    try {
      const response = await this.apiClient.get('/merchants/session/active');
      this.activeSession = response.data;
      return this.activeSession;
    } catch (error) {
      if (error.response?.status === 404) {
        this.activeSession = null;
        return null;
      }
      throw error;
    }
  }
}
```

### 3. Real-time Updates with Webhooks

```javascript
// Webhook handler for real-time updates
const express = require('express');
const app = express();

app.use(express.json());

// Webhook endpoint
app.post('/webhooks/merchant-updates', (req, res) => {
  const { event, data, timestamp } = req.body;
  
  console.log(`Received webhook: ${event} at ${timestamp}`);
  
  switch (event) {
    case 'merchant.created':
      handleMerchantCreated(data);
      break;
    case 'merchant.updated':
      handleMerchantUpdated(data);
      break;
    case 'merchant.portfolio_type_changed':
      handlePortfolioTypeChanged(data);
      break;
    case 'bulk_operation.completed':
      handleBulkOperationCompleted(data);
      break;
    default:
      console.log(`Unhandled event: ${event}`);
  }
  
  res.status(200).json({ received: true });
});

function handleMerchantCreated(data) {
  console.log(`New merchant created: ${data.merchant_id}`);
  // Update your local cache or database
}

function handleMerchantUpdated(data) {
  console.log(`Merchant updated: ${data.merchant_id}`);
  // Refresh merchant data in your application
}

function handlePortfolioTypeChanged(data) {
  console.log(`Portfolio type changed for ${data.merchant_id}: ${data.changes.portfolio_type.old} -> ${data.changes.portfolio_type.new}`);
  // Update UI or trigger notifications
}

function handleBulkOperationCompleted(data) {
  console.log(`Bulk operation completed: ${data.operation_id}`);
  // Refresh affected data or show completion notification
}

app.listen(3000, () => {
  console.log('Webhook server running on port 3000');
});
```

### 4. Caching Strategy

```javascript
// Simple in-memory cache with TTL
class APICache {
  constructor(ttl = 300000) { // 5 minutes default TTL
    this.cache = new Map();
    this.ttl = ttl;
  }
  
  get(key) {
    const item = this.cache.get(key);
    if (!item) return null;
    
    if (Date.now() > item.expiry) {
      this.cache.delete(key);
      return null;
    }
    
    return item.data;
  }
  
  set(key, data) {
    this.cache.set(key, {
      data,
      expiry: Date.now() + this.ttl
    });
  }
  
  clear() {
    this.cache.clear();
  }
}

// Enhanced API client with caching
class CachedAPIClient {
  constructor(baseURL, apiKey) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
    this.cache = new APICache();
  }
  
  async getMerchant(merchantId, useCache = true) {
    const cacheKey = `merchant_${merchantId}`;
    
    if (useCache) {
      const cached = this.cache.get(cacheKey);
      if (cached) {
        console.log(`Cache hit for merchant ${merchantId}`);
        return cached;
      }
    }
    
    const response = await axios.get(
      `${this.baseURL}/merchants/${merchantId}`,
      {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`
        }
      }
    );
    
    if (useCache) {
      this.cache.set(cacheKey, response.data);
    }
    
    return response.data;
  }
  
  async searchMerchants(filters, useCache = true) {
    const cacheKey = `search_${JSON.stringify(filters)}`;
    
    if (useCache) {
      const cached = this.cache.get(cacheKey);
      if (cached) {
        console.log('Cache hit for search');
        return cached;
      }
    }
    
    const response = await axios.get(
      `${this.baseURL}/merchants`,
      {
        params: filters,
        headers: {
          'Authorization': `Bearer ${this.apiKey}`
        }
      }
    );
    
    if (useCache) {
      this.cache.set(cacheKey, response.data);
    }
    
    return response.data;
  }
}
```

## Error Handling Best Practices

### 1. Comprehensive Error Handling

```javascript
class APIError extends Error {
  constructor(message, status, code, details) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

async function handleAPIRequest(requestFn) {
  try {
    return await requestFn();
  } catch (error) {
    if (error.response) {
      // API returned an error response
      const { status, data } = error.response;
      const errorData = data.error || {};
      
      throw new APIError(
        errorData.message || 'API request failed',
        status,
        errorData.code || 'UNKNOWN_ERROR',
        errorData.details
      );
    } else if (error.request) {
      // Network error
      throw new APIError(
        'Network error - unable to reach API',
        0,
        'NETWORK_ERROR',
        error.message
      );
    } else {
      // Other error
      throw new APIError(
        'Unexpected error',
        0,
        'UNEXPECTED_ERROR',
        error.message
      );
    }
  }
}

// Usage with retry logic
async function retryableRequest(requestFn, maxRetries = 3) {
  let lastError;
  
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await handleAPIRequest(requestFn);
    } catch (error) {
      lastError = error;
      
      // Don't retry on client errors (4xx)
      if (error.status >= 400 && error.status < 500) {
        throw error;
      }
      
      // Don't retry on last attempt
      if (attempt === maxRetries) {
        throw error;
      }
      
      // Exponential backoff
      const delay = Math.pow(2, attempt) * 1000;
      console.log(`Attempt ${attempt} failed, retrying in ${delay}ms...`);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
  
  throw lastError;
}
```

### 2. Rate Limit Handling

```javascript
class RateLimitHandler {
  constructor() {
    this.requests = [];
    this.limit = 100; // requests per minute
    this.window = 60000; // 1 minute in milliseconds
  }
  
  async waitIfNeeded() {
    const now = Date.now();
    
    // Remove old requests outside the window
    this.requests = this.requests.filter(time => now - time < this.window);
    
    if (this.requests.length >= this.limit) {
      const oldestRequest = Math.min(...this.requests);
      const waitTime = this.window - (now - oldestRequest);
      
      if (waitTime > 0) {
        console.log(`Rate limit reached, waiting ${waitTime}ms...`);
        await new Promise(resolve => setTimeout(resolve, waitTime));
      }
    }
    
    this.requests.push(now);
  }
}

// Enhanced API client with rate limiting
class RateLimitedAPIClient {
  constructor(baseURL, apiKey) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
    this.rateLimiter = new RateLimitHandler();
  }
  
  async request(method, endpoint, data = null) {
    await this.rateLimiter.waitIfNeeded();
    
    const config = {
      method,
      url: `${this.baseURL}${endpoint}`,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json'
      }
    };
    
    if (data) {
      config.data = data;
    }
    
    return axios(config);
  }
}
```

## Performance Optimization

### 1. Batch Operations

```javascript
// Batch merchant creation
async function batchCreateMerchants(merchants, batchSize = 10) {
  const results = [];
  
  for (let i = 0; i < merchants.length; i += batchSize) {
    const batch = merchants.slice(i, i + batchSize);
    
    console.log(`Processing batch ${Math.floor(i / batchSize) + 1} of ${Math.ceil(merchants.length / batchSize)}`);
    
    const batchPromises = batch.map(merchant => 
      createMerchant(merchant).catch(error => ({
        error: error.message,
        merchant: merchant.name
      }))
    );
    
    const batchResults = await Promise.all(batchPromises);
    results.push(...batchResults);
    
    // Add delay between batches to respect rate limits
    if (i + batchSize < merchants.length) {
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }
  
  return results;
}
```

### 2. Parallel Processing

```javascript
// Parallel merchant updates
async function parallelUpdateMerchants(updates) {
  const updatePromises = updates.map(update => 
    updateMerchant(update.id, update.data)
      .then(result => ({ success: true, id: update.id, result }))
      .catch(error => ({ success: false, id: update.id, error: error.message }))
  );
  
  const results = await Promise.all(updatePromises);
  
  const successful = results.filter(r => r.success);
  const failed = results.filter(r => !r.success);
  
  console.log(`Updated ${successful.length} merchants successfully`);
  if (failed.length > 0) {
    console.log(`${failed.length} updates failed:`, failed);
  }
  
  return { successful, failed };
}
```

### 3. Data Streaming

```javascript
// Stream large merchant lists
async function* streamMerchants(filters = {}) {
  let page = 1;
  let hasMore = true;
  
  while (hasMore) {
    const response = await axios.get(
      'https://api.kyb-platform.com/v1/merchants',
      {
        params: {
          ...filters,
          page,
          page_size: 100
        },
        headers: {
          'Authorization': `Bearer ${process.env.API_TOKEN}`
        }
      }
    );
    
    const data = response.data;
    
    for (const merchant of data.merchants) {
      yield merchant;
    }
    
    hasMore = data.has_more;
    page++;
  }
}

// Usage
async function processAllMerchants() {
  for await (const merchant of streamMerchants({ portfolio_type: 'onboarded' })) {
    console.log(`Processing merchant: ${merchant.name}`);
    // Process each merchant as it arrives
  }
}
```

## Testing Your Integration

### 1. Unit Tests

```javascript
// Jest test example
const axios = require('axios');
const { createMerchant, searchMerchants } = require('./merchant-api');

// Mock axios
jest.mock('axios');
const mockedAxios = axios;

describe('Merchant API', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });
  
  test('should create merchant successfully', async () => {
    const mockResponse = {
      data: {
        id: 'merchant_123',
        name: 'Test Corp',
        portfolio_type: 'prospective'
      }
    };
    
    mockedAxios.post.mockResolvedValue(mockResponse);
    
    const result = await createMerchant({
      name: 'Test Corp',
      portfolio_type: 'prospective'
    });
    
    expect(result).toEqual(mockResponse.data);
    expect(mockedAxios.post).toHaveBeenCalledWith(
      'https://api.kyb-platform.com/v1/merchants',
      expect.any(Object),
      expect.any(Object)
    );
  });
  
  test('should handle API errors', async () => {
    const mockError = {
      response: {
        status: 400,
        data: {
          error: {
            code: 'VALIDATION_ERROR',
            message: 'Invalid request data'
          }
        }
      }
    };
    
    mockedAxios.post.mockRejectedValue(mockError);
    
    await expect(createMerchant({})).rejects.toThrow('Invalid request data');
  });
});
```

### 2. Integration Tests

```javascript
// Integration test with real API
describe('Merchant API Integration', () => {
  const testMerchant = {
    name: 'Integration Test Corp',
    legal_name: 'Integration Test Corporation',
    industry: 'Technology',
    portfolio_type: 'prospective',
    risk_level: 'medium',
    address: {
      street1: '123 Test St',
      city: 'Test City',
      state: 'TS',
      postal_code: '12345',
      country: 'United States',
      country_code: 'US'
    },
    contact_info: {
      email: 'test@integration.com',
      phone: '+1-555-123-4567'
    }
  };
  
  let createdMerchantId;
  
  test('should create, read, update, and delete merchant', async () => {
    // Create
    const created = await createMerchant(testMerchant);
    expect(created.id).toBeDefined();
    createdMerchantId = created.id;
    
    // Read
    const retrieved = await getMerchant(createdMerchantId);
    expect(retrieved.name).toBe(testMerchant.name);
    
    // Update
    const updated = await updateMerchant(createdMerchantId, {
      portfolio_type: 'onboarded'
    });
    expect(updated.portfolio_type).toBe('onboarded');
    
    // Delete
    await deleteMerchant(createdMerchantId);
    
    // Verify deletion
    await expect(getMerchant(createdMerchantId)).rejects.toThrow();
  });
});
```

### 3. Load Testing

```javascript
// Load test with multiple concurrent requests
const loadTest = async (concurrency = 10, totalRequests = 100) => {
  const results = [];
  const startTime = Date.now();
  
  const makeRequest = async (index) => {
    try {
      const response = await searchMerchants({
        page: Math.floor(Math.random() * 10) + 1,
        page_size: 20
      });
      return { success: true, index, responseTime: Date.now() - startTime };
    } catch (error) {
      return { success: false, index, error: error.message };
    }
  };
  
  // Create batches of concurrent requests
  for (let i = 0; i < totalRequests; i += concurrency) {
    const batch = [];
    for (let j = 0; j < concurrency && i + j < totalRequests; j++) {
      batch.push(makeRequest(i + j));
    }
    
    const batchResults = await Promise.all(batch);
    results.push(...batchResults);
    
    // Small delay between batches
    await new Promise(resolve => setTimeout(resolve, 100));
  }
  
  const successful = results.filter(r => r.success);
  const failed = results.filter(r => !r.success);
  
  console.log(`Load test completed:`);
  console.log(`- Total requests: ${totalRequests}`);
  console.log(`- Successful: ${successful.length}`);
  console.log(`- Failed: ${failed.length}`);
  console.log(`- Success rate: ${(successful.length / totalRequests * 100).toFixed(2)}%`);
  
  return { successful, failed, totalTime: Date.now() - startTime };
};

// Run load test
loadTest(20, 200).then(results => {
  console.log('Load test results:', results);
});
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Authentication Errors

**Problem**: `401 Unauthorized` errors

**Solutions**:
- Verify your API token is correct and not expired
- Check that the Authorization header format is correct: `Bearer <token>`
- Ensure you're using the correct base URL
- Check if your token has the required permissions

```javascript
// Debug authentication
async function debugAuth() {
  try {
    const response = await axios.get(
      'https://api.kyb-platform.com/v1/merchants/portfolio-types',
      {
        headers: {
          'Authorization': `Bearer ${process.env.API_TOKEN}`
        }
      }
    );
    console.log('Authentication successful');
  } catch (error) {
    console.error('Authentication failed:', error.response?.data);
  }
}
```

#### 2. Rate Limit Issues

**Problem**: `429 Too Many Requests` errors

**Solutions**:
- Implement exponential backoff
- Reduce request frequency
- Use bulk operations instead of individual requests
- Cache responses when appropriate

```javascript
// Rate limit handler with exponential backoff
async function withRetry(fn, maxRetries = 3) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (error.response?.status === 429 && attempt < maxRetries) {
        const retryAfter = error.response.headers['retry-after'] || Math.pow(2, attempt);
        console.log(`Rate limited, waiting ${retryAfter}s before retry ${attempt + 1}`);
        await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
        continue;
      }
      throw error;
    }
  }
}
```

#### 3. Validation Errors

**Problem**: `400 Bad Request` with validation errors

**Solutions**:
- Check required fields are provided
- Validate data types and formats
- Review field length limits
- Ensure enum values are correct

```javascript
// Validation helper
function validateMerchantData(data) {
  const errors = [];
  
  if (!data.name || data.name.trim().length === 0) {
    errors.push('Name is required');
  }
  
  if (!['onboarded', 'deactivated', 'prospective', 'pending'].includes(data.portfolio_type)) {
    errors.push('Invalid portfolio type');
  }
  
  if (!['high', 'medium', 'low'].includes(data.risk_level)) {
    errors.push('Invalid risk level');
  }
  
  if (data.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(data.email)) {
    errors.push('Invalid email format');
  }
  
  return errors;
}
```

#### 4. Network Issues

**Problem**: Connection timeouts or network errors

**Solutions**:
- Implement retry logic with exponential backoff
- Use connection pooling
- Set appropriate timeouts
- Handle network interruptions gracefully

```javascript
// Network error handler
const axios = require('axios');

const apiClient = axios.create({
  baseURL: 'https://api.kyb-platform.com/v1',
  timeout: 30000, // 30 seconds
  headers: {
    'Authorization': `Bearer ${process.env.API_TOKEN}`
  }
});

// Add retry interceptor
apiClient.interceptors.response.use(
  response => response,
  async error => {
    const config = error.config;
    
    if (!config || !config.retry) {
      config.retry = 0;
    }
    
    if (config.retry >= 3) {
      return Promise.reject(error);
    }
    
    config.retry++;
    
    // Wait before retrying
    await new Promise(resolve => setTimeout(resolve, Math.pow(2, config.retry) * 1000));
    
    return apiClient(config);
  }
);
```

### Debugging Tips

1. **Enable request/response logging**:
```javascript
// Log all API requests and responses
axios.interceptors.request.use(request => {
  console.log('Request:', request.method.toUpperCase(), request.url);
  return request;
});

axios.interceptors.response.use(
  response => {
    console.log('Response:', response.status, response.config.url);
    return response;
  },
  error => {
    console.error('Error:', error.response?.status, error.config?.url, error.response?.data);
    return Promise.reject(error);
  }
);
```

2. **Use request IDs for tracing**:
```javascript
// Add request ID to all requests
const requestId = require('crypto').randomUUID();

axios.interceptors.request.use(config => {
  config.headers['X-Request-ID'] = requestId;
  return config;
});
```

3. **Monitor API usage**:
```javascript
// Track API usage metrics
class APIMetrics {
  constructor() {
    this.requests = 0;
    this.errors = 0;
    this.totalTime = 0;
  }
  
  recordRequest(duration, success) {
    this.requests++;
    this.totalTime += duration;
    if (!success) this.errors++;
  }
  
  getStats() {
    return {
      totalRequests: this.requests,
      errorRate: this.errors / this.requests,
      averageResponseTime: this.totalTime / this.requests
    };
  }
}
```

## Support and Resources

- **API Documentation**: [https://docs.kyb-platform.com/api](https://docs.kyb-platform.com/api)
- **Status Page**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: api-support@kyb-platform.com
- **GitHub Repository**: [https://github.com/kyb-platform/api-examples](https://github.com/kyb-platform/api-examples)

For additional help or questions about integration, please contact our API support team.
