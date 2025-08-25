# KYB Platform Caching API Overview

## Overview

The KYB Platform Caching API provides intelligent caching with optimization, monitoring, and invalidation capabilities.

## Base URL
- Production: `https://api.kyb-platform.com/v1/cache`
- Staging: `https://staging-api.kyb-platform.com/v1/cache`

## Authentication
```http
Authorization: Bearer your-api-key-here
```

## Key Endpoints

### Cache Operations
- `GET /cache/{key}` - Get cached value
- `PUT /cache/{key}` - Set cache value
- `DELETE /cache/{key}` - Delete cache value
- `DELETE /cache` - Clear all cache

### Statistics & Analytics
- `GET /cache/stats` - Get cache statistics
- `GET /cache/analytics?time_range=24h` - Get detailed analytics

### Optimization
- `GET /cache/optimization/plans` - List optimization plans
- `POST /cache/optimization/plans` - Generate optimization plan
- `POST /cache/optimization/plans/{plan_id}` - Execute optimization plan

### Invalidation
- `GET /cache/invalidation/rules` - List invalidation rules
- `POST /cache/invalidation/rules` - Create invalidation rule
- `POST /cache/invalidation/execute` - Execute invalidation

### Health
- `GET /cache/health` - Get cache health status

## Example Usage

### Store Value
```bash
curl -X PUT "https://api.kyb-platform.com/v1/cache/user:12345:profile" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "value": {"name": "John Doe", "email": "john@example.com"},
    "ttl": 3600,
    "tags": ["user", "profile"]
  }'
```

### Get Value
```bash
curl -X GET "https://api.kyb-platform.com/v1/cache/user:12345:profile" \
  -H "Authorization: Bearer your-api-key"
```

### Generate Optimization Plan
```bash
curl -X POST "https://api.kyb-platform.com/v1/cache/optimization/plans" \
  -H "Authorization: Bearer your-api-key"
```

## Error Handling
```json
{
  "error": "Invalid request data",
  "code": "INVALID_REQUEST",
  "timestamp": "2024-12-19T11:00:00Z",
  "request_id": "req_1234567890"
}
```

## Rate Limiting
- Standard: 1000 requests/minute
- Premium: 5000 requests/minute
- Enterprise: 10000 requests/minute

## Support
- Email: api-support@kyb-platform.com
- Documentation: https://docs.kyb-platform.com/api
