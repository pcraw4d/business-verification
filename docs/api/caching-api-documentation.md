# KYB Platform Caching API Documentation

## Overview

The KYB Platform Caching API provides a comprehensive intelligent caching system with advanced optimization capabilities, performance monitoring, and sophisticated invalidation strategies. This API enables efficient data caching with multiple eviction policies, real-time analytics, and automated optimization.

## Table of Contents

1. [Authentication](#authentication)
2. [Base URL](#base-url)
3. [API Endpoints](#api-endpoints)
4. [Request/Response Formats](#requestresponse-formats)
5. [Error Handling](#error-handling)
6. [Rate Limiting](#rate-limiting)
7. [Usage Examples](#usage-examples)
8. [Integration Guidelines](#integration-guidelines)
9. [Best Practices](#best-practices)

## Authentication

The API supports two authentication methods:

### API Key Authentication
```http
Authorization: Bearer your-api-key-here
```

### JWT Token Authentication
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Base URL

- **Production**: `https://api.kyb-platform.com/v1/cache`
- **Staging**: `https://staging-api.kyb-platform.com/v1/cache`
- **Development**: `http://localhost:8080/v1/cache`

## API Endpoints

### Cache Operations

#### Get Cache Value
```http
GET /cache/{key}
```

Retrieves a value from the intelligent cache.

**Parameters:**
- `key` (path, required): Cache key

**Response:**
```json
{
  "key": "user:12345:profile",
  "value": "cached data",
  "ttl": 1800,
  "created_at": "2024-12-19T10:30:00Z",
  "accessed_at": "2024-12-19T11:00:00Z",
  "access_count": 5,
  "tags": ["user", "profile"],
  "metadata": {
    "source": "database",
    "version": "1.0"
  }
}
```

#### Set Cache Value
```http
PUT /cache/{key}
```

Stores a value in the intelligent cache with optional TTL and metadata.

**Parameters:**
- `key` (path, required): Cache key

**Request Body:**
```json
{
  "value": "cached data",
  "ttl": 3600,
  "priority": 1,
  "tags": ["user", "profile"],
  "metadata": {
    "source": "database",
    "version": "1.0"
  }
}
```

**Response:**
```json
{
  "key": "user:12345:profile",
  "success": true,
  "message": "Value cached successfully",
  "ttl": 3600,
  "size": 1024
}
```

#### Delete Cache Value
```http
DELETE /cache/{key}
```

Removes a value from the cache.

**Parameters:**
- `key` (path, required): Cache key

**Response:**
```json
{
  "key": "user:12345:profile",
  "success": true,
  "message": "Value deleted successfully"
}
```

#### Clear All Cache
```http
DELETE /cache
```

Removes all values from the cache.

**Response:**
```json
{
  "success": true,
  "message": "Cache cleared successfully",
  "deleted_count": 1000
}
```

### Cache Statistics

#### Get Cache Statistics
```http
GET /cache/stats
```

Retrieves comprehensive cache statistics and performance metrics.

**Response:**
```json
{
  "hits": 15000,
  "misses": 3000,
  "evictions": 500,
  "expirations": 200,
  "total_size": 104857600,
  "entry_count": 5000,
  "hit_rate": 0.833,
  "miss_rate": 0.167,
  "average_access_time": 2.5,
  "memory_usage": 52428800,
  "throughput": 1000,
  "shard_count": 8
}
```

### Cache Analytics

#### Get Cache Analytics
```http
GET /cache/analytics?time_range=24h
```

Retrieves detailed cache analytics including access patterns and performance insights.

**Parameters:**
- `time_range` (query, optional): Time range for analytics (e.g., "1h", "24h", "7d")

**Response:**
```json
{
  "time_range": "24h",
  "hit_rate": 0.85,
  "miss_rate": 0.15,
  "eviction_rate": 0.05,
  "expiration_rate": 0.02,
  "average_entry_size": 2048,
  "average_access_time": 3.2,
  "popular_keys": [
    {
      "key": "user:12345:profile",
      "access_count": 150,
      "last_access": "2024-12-19T11:00:00Z"
    }
  ],
  "hot_keys": [
    {
      "key": "session:67890",
      "access_count": 25,
      "last_access": "2024-12-19T11:05:00Z"
    }
  ],
  "cold_keys": [
    {
      "key": "config:old",
      "access_count": 1,
      "last_access": "2024-12-18T10:00:00Z"
    }
  ],
  "access_patterns": {
    "read_write_ratio": 0.8,
    "temporal_patterns": {
      "00:00": 100,
      "06:00": 500,
      "12:00": 1000,
      "18:00": 800
    }
  },
  "size_distribution": {
    "small": 2000,
    "medium": 2500,
    "large": 500
  }
}
```

### Optimization

#### List Optimization Plans
```http
GET /cache/optimization/plans
```

Retrieves all optimization plans.

**Response:**
```json
{
  "plans": [
    {
      "id": "plan_1234567890",
      "name": "Cache Performance Optimization",
      "description": "Optimize cache performance based on current metrics",
      "actions": [
        {
          "id": "action_1234567890",
          "strategy": "size_adjustment",
          "description": "Increase cache size to reduce evictions",
          "parameters": {
            "new_size": 209715200
          },
          "priority": 1,
          "impact": "high",
          "risk": "low",
          "estimated_gain": 0.15,
          "estimated_cost": 0.05,
          "roi": 3.0
        }
      ],
      "estimated_total_gain": 0.25,
      "estimated_total_cost": 0.05,
      "estimated_roi": 5.0,
      "risk_level": "low",
      "execution_time": 30,
      "created_at": "2024-12-19T11:00:00Z",
      "status": "pending"
    }
  ],
  "total_count": 10
}
```

#### Generate Optimization Plan
```http
POST /cache/optimization/plans
```

Generates a new optimization plan based on current cache performance.

**Request Body:**
```json
{
  "force_generation": false
}
```

**Response:**
```json
{
  "id": "plan_1234567890",
  "name": "Cache Performance Optimization",
  "description": "Optimize cache performance based on current metrics",
  "actions": [
    {
      "id": "action_1234567890",
      "strategy": "size_adjustment",
      "description": "Increase cache size to reduce evictions",
      "parameters": {
        "new_size": 209715200
      },
      "priority": 1,
      "impact": "high",
      "risk": "low",
      "estimated_gain": 0.15,
      "estimated_cost": 0.05,
      "roi": 3.0
    }
  ],
  "estimated_total_gain": 0.25,
  "estimated_total_cost": 0.05,
  "estimated_roi": 5.0,
  "risk_level": "low",
  "execution_time": 30,
  "created_at": "2024-12-19T11:00:00Z",
  "status": "pending"
}
```

#### Get Optimization Plan
```http
GET /cache/optimization/plans/{plan_id}
```

Retrieves a specific optimization plan.

**Parameters:**
- `plan_id` (path, required): Optimization plan ID

#### Execute Optimization Plan
```http
POST /cache/optimization/plans/{plan_id}
```

Executes a specific optimization plan.

**Parameters:**
- `plan_id` (path, required): Optimization plan ID

**Response:**
```json
{
  "id": "result_1234567890",
  "plan_id": "plan_1234567890",
  "success": true,
  "error_message": "",
  "actual_gain": 0.18,
  "actual_cost": 0.03,
  "execution_duration": 25,
  "timestamp": "2024-12-19T11:05:00Z"
}
```

#### List Optimization Results
```http
GET /cache/optimization/results
```

Retrieves all optimization execution results.

**Response:**
```json
{
  "results": [
    {
      "id": "result_1234567890",
      "plan_id": "plan_1234567890",
      "success": true,
      "error_message": "",
      "actual_gain": 0.18,
      "actual_cost": 0.03,
      "execution_duration": 25,
      "timestamp": "2024-12-19T11:05:00Z"
    }
  ],
  "total_count": 15
}
```

### Invalidation

#### List Invalidation Rules
```http
GET /cache/invalidation/rules
```

Retrieves all cache invalidation rules.

**Response:**
```json
{
  "rules": [
    {
      "id": "rule_1234567890",
      "name": "User profile invalidation",
      "strategy": "pattern",
      "pattern": "user:*:profile",
      "tags": ["user", "profile"],
      "dependencies": ["user:12345:data"],
      "conditions": {
        "time_range": {
          "start": "00:00",
          "end": "06:00"
        }
      },
      "priority": 1,
      "enabled": true,
      "created_at": "2024-12-19T10:00:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total_count": 5
}
```

#### Create Invalidation Rule
```http
POST /cache/invalidation/rules
```

Creates a new cache invalidation rule.

**Request Body:**
```json
{
  "name": "User profile invalidation",
  "strategy": "pattern",
  "pattern": "user:*:profile",
  "tags": ["user", "profile"],
  "dependencies": ["user:12345:data"],
  "conditions": {
    "time_range": {
      "start": "00:00",
      "end": "06:00"
    }
  },
  "priority": 1,
  "enabled": true
}
```

**Response:**
```json
{
  "id": "rule_1234567890",
  "name": "User profile invalidation",
  "strategy": "pattern",
  "pattern": "user:*:profile",
  "tags": ["user", "profile"],
  "dependencies": ["user:12345:data"],
  "conditions": {
    "time_range": {
      "start": "00:00",
      "end": "06:00"
    }
  },
  "priority": 1,
  "enabled": true,
  "created_at": "2024-12-19T10:00:00Z",
  "updated_at": "2024-12-19T10:00:00Z"
}
```

#### Get Invalidation Rule
```http
GET /cache/invalidation/rules/{rule_id}
```

Retrieves a specific invalidation rule.

**Parameters:**
- `rule_id` (path, required): Invalidation rule ID

#### Update Invalidation Rule
```http
PUT /cache/invalidation/rules/{rule_id}
```

Updates an existing invalidation rule.

**Parameters:**
- `rule_id` (path, required): Invalidation rule ID

**Request Body:** Same as Create Invalidation Rule

#### Delete Invalidation Rule
```http
DELETE /cache/invalidation/rules/{rule_id}
```

Deletes an invalidation rule.

**Parameters:**
- `rule_id` (path, required): Invalidation rule ID

**Response:**
```json
{
  "id": "rule_1234567890",
  "success": true,
  "message": "Invalidation rule deleted successfully"
}
```

#### Execute Invalidation
```http
POST /cache/invalidation/execute
```

Executes cache invalidation by key, pattern, or all.

**Request Body:**
```json
{
  "strategy": "pattern",
  "pattern": "user:*:profile"
}
```

**Response:**
```json
{
  "strategy": "pattern",
  "pattern": "user:*:profile",
  "invalidated_count": 150,
  "success": true,
  "message": "Invalidation executed successfully"
}
```

### Health and Status

#### Get Cache Health
```http
GET /cache/health
```

Retrieves cache health status and basic information.

**Response:**
```json
{
  "status": "healthy",
  "uptime": 86400,
  "version": "1.0.0",
  "eviction_policy": "lru",
  "shard_count": 8,
  "total_entries": 5000,
  "total_size": 104857600,
  "memory_usage": 52428800,
  "hit_rate": 0.85,
  "last_optimization": "2024-12-19T10:00:00Z",
  "checks": {
    "cache_accessible": true,
    "memory_ok": true,
    "performance_ok": true,
    "optimization_enabled": true
  }
}
```

## Request/Response Formats

### Request Headers
```http
Content-Type: application/json
Authorization: Bearer your-api-key-here
Accept: application/json
```

### Response Headers
```http
Content-Type: application/json
X-Request-ID: req_1234567890
X-Cache-Hit: true
```

### Pagination
For endpoints that return lists, pagination is supported:

```http
GET /cache/optimization/plans?page=1&per_page=20
```

**Response:**
```json
{
  "plans": [...],
  "total_count": 100,
  "page": 1,
  "per_page": 20,
  "total_pages": 5
}
```

## Error Handling

### Error Response Format
```json
{
  "error": "Invalid request data",
  "code": "INVALID_REQUEST",
  "details": {
    "field": "ttl",
    "reason": "Value must be positive"
  },
  "timestamp": "2024-12-19T11:00:00Z",
  "request_id": "req_1234567890"
}
```

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `413` - Request Entity Too Large
- `429` - Too Many Requests
- `500` - Internal Server Error

### Error Codes
- `INVALID_REQUEST` - Invalid request data
- `KEY_NOT_FOUND` - Cache key not found
- `VALUE_TOO_LARGE` - Value exceeds maximum cache size
- `PLAN_NOT_FOUND` - Optimization plan not found
- `RULE_NOT_FOUND` - Invalidation rule not found
- `NO_OPTIMIZATION_NEEDED` - No optimization needed
- `INVALIDATION_FAILED` - Invalidation execution failed
- `INTERNAL_ERROR` - Internal server error

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Standard Plan**: 1000 requests per minute
- **Premium Plan**: 5000 requests per minute
- **Enterprise Plan**: 10000 requests per minute

Rate limit headers are included in responses:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1640000000
```

## Usage Examples

### Basic Cache Operations

#### Store User Profile
```bash
curl -X PUT "https://api.kyb-platform.com/v1/cache/user:12345:profile" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "id": "12345",
      "name": "John Doe",
      "email": "john@example.com"
    },
    "ttl": 3600,
    "tags": ["user", "profile"]
  }'
```

#### Retrieve User Profile
```bash
curl -X GET "https://api.kyb-platform.com/v1/cache/user:12345:profile" \
  -H "Authorization: Bearer your-api-key"
```

#### Delete User Profile
```bash
curl -X DELETE "https://api.kyb-platform.com/v1/cache/user:12345:profile" \
  -H "Authorization: Bearer your-api-key"
```

### Optimization Management

#### Generate Optimization Plan
```bash
curl -X POST "https://api.kyb-platform.com/v1/cache/optimization/plans" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "force_generation": false
  }'
```

#### Execute Optimization Plan
```bash
curl -X POST "https://api.kyb-platform.com/v1/cache/optimization/plans/plan_1234567890" \
  -H "Authorization: Bearer your-api-key"
```

### Invalidation Management

#### Create Invalidation Rule
```bash
curl -X POST "https://api.kyb-platform.com/v1/cache/invalidation/rules" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Session cleanup",
    "strategy": "pattern",
    "pattern": "session:.*",
    "priority": 2,
    "enabled": true
  }'
```

#### Execute Invalidation
```bash
curl -X POST "https://api.kyb-platform.com/v1/cache/invalidation/execute" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "strategy": "pattern",
    "pattern": "user:*:profile"
  }'
```

### Monitoring and Analytics

#### Get Cache Statistics
```bash
curl -X GET "https://api.kyb-platform.com/v1/cache/stats" \
  -H "Authorization: Bearer your-api-key"
```

#### Get Cache Analytics
```bash
curl -X GET "https://api.kyb-platform.com/v1/cache/analytics?time_range=24h" \
  -H "Authorization: Bearer your-api-key"
```

#### Check Cache Health
```bash
curl -X GET "https://api.kyb-platform.com/v1/cache/health" \
  -H "Authorization: Bearer your-api-key"
```

## Integration Guidelines

### SDK Usage

The KYB Platform provides official SDKs for popular programming languages:

#### Go SDK
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kyb-platform/go-sdk/cache"
)

func main() {
    client := cache.NewClient("your-api-key")
    
    // Store value
    err := client.Set("user:12345:profile", map[string]interface{}{
        "id":    "12345",
        "name":  "John Doe",
        "email": "john@example.com",
    }, cache.WithTTL(time.Hour))
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve value
    value, found, err := client.Get("user:12345:profile")
    if err != nil {
        log.Fatal(err)
    }
    
    if found {
        fmt.Printf("Retrieved: %+v\n", value)
    }
}
```

#### Python SDK
```python
from kyb_cache import CacheClient

client = CacheClient("your-api-key")

# Store value
client.set("user:12345:profile", {
    "id": "12345",
    "name": "John Doe",
    "email": "john@example.com"
}, ttl=3600)

# Retrieve value
value = client.get("user:12345:profile")
if value:
    print(f"Retrieved: {value}")
```

#### JavaScript SDK
```javascript
const { CacheClient } = require('@kyb-platform/cache');

const client = new CacheClient('your-api-key');

// Store value
await client.set('user:12345:profile', {
    id: '12345',
    name: 'John Doe',
    email: 'john@example.com'
}, { ttl: 3600 });

// Retrieve value
const value = await client.get('user:12345:profile');
if (value) {
    console.log('Retrieved:', value);
}
```

### Webhook Integration

The API supports webhooks for real-time notifications:

#### Configure Webhook
```bash
curl -X POST "https://api.kyb-platform.com/v1/webhooks" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/cache",
    "events": ["cache.hit", "cache.miss", "cache.eviction"],
    "secret": "your-webhook-secret"
  }'
```

#### Webhook Payload
```json
{
  "event": "cache.hit",
  "timestamp": "2024-12-19T11:00:00Z",
  "data": {
    "key": "user:12345:profile",
    "hit_rate": 0.85,
    "access_count": 150
  }
}
```

## Best Practices

### Cache Key Design
- Use descriptive, hierarchical keys: `user:{id}:profile`
- Include version information: `user:{id}:profile:v2`
- Use consistent naming conventions
- Avoid special characters in keys

### TTL Management
- Set appropriate TTL based on data freshness requirements
- Use shorter TTL for frequently changing data
- Use longer TTL for static data
- Monitor cache hit rates to optimize TTL

### Optimization
- Regularly monitor cache performance
- Generate and review optimization plans
- Execute optimizations during low-traffic periods
- Monitor optimization results

### Invalidation
- Use pattern-based invalidation for related data
- Set up automatic invalidation rules
- Monitor invalidation effectiveness
- Use dependency-based invalidation for complex relationships

### Error Handling
- Implement retry logic with exponential backoff
- Handle cache misses gracefully
- Monitor error rates and patterns
- Use circuit breakers for external dependencies

### Security
- Use HTTPS for all API calls
- Rotate API keys regularly
- Implement proper authentication
- Monitor for suspicious activity

### Performance
- Use connection pooling
- Implement request batching
- Monitor response times
- Use appropriate cache sizes

## Support

For API support and questions:

- **Email**: api-support@kyb-platform.com
- **Documentation**: https://docs.kyb-platform.com/api
- **Status Page**: https://status.kyb-platform.com
- **Community**: https://community.kyb-platform.com

## Changelog

### Version 1.0.0 (2024-12-19)
- Initial release of caching API
- Support for intelligent caching algorithms
- Performance monitoring and analytics
- Optimization strategies
- Advanced invalidation mechanisms
