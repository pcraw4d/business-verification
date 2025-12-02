# Redis Setup Guide for Testing

## Overview

This guide explains how to set up Redis for running Redis-dependent tests in the classification service.

---

## Quick Start

### Option 1: Homebrew (macOS)

```bash
# Install Redis
brew install redis

# Start Redis
brew services start redis

# Verify Redis is running
redis-cli ping
# Should return: PONG
```

### Option 2: Docker

```bash
# Run Redis in Docker
docker run -d --name redis-test -p 6379:6379 redis:latest

# Verify Redis is running
docker ps | grep redis
redis-cli -h localhost -p 6379 ping
```

### Option 3: System Package Manager (Linux)

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install redis-server
sudo systemctl start redis-server

# Verify Redis is running
redis-cli ping
```

---

## Running Tests with Redis

### Set Environment Variable

```bash
export REDIS_URL=redis://localhost:6379
```

### Run Cache Tests

```bash
# Run all cache tests
go test -v ./services/classification-service/internal/cache -run TestWebsiteContentCache

# Run specific Redis-dependent tests
go test -v ./services/classification-service/internal/cache -run TestWebsiteContentCache_GetKey_Indirect
go test -v ./services/classification-service/internal/cache -run TestWebsiteContentCache_Get_UnmarshalError
go test -v ./services/classification-service/internal/cache -run TestWebsiteContentCache_Set_MarshalError
go test -v ./services/classification-service/internal/cache -run TestWebsiteContentCache_GetKey_Format
```

### Run Integration Tests

```bash
export INTEGRATION_TESTS=true
export REDIS_URL=redis://localhost:6379
go test -v ./services/classification-service/internal/integration -run TestRedisCache_WebsiteContent
```

---

## Test Behavior

### Without Redis

Tests gracefully skip when Redis is not available:
- ✅ No test failures
- ✅ Tests skip with informative messages
- ✅ All non-Redis tests continue to pass

### With Redis

All Redis-dependent tests should:
- ✅ Pass successfully
- ✅ Validate cache operations
- ✅ Test error handling
- ✅ Improve code coverage

---

## Expected Coverage Improvements

With Redis available, coverage should improve:

| Function | Without Redis | With Redis (Expected) |
|----------|---------------|----------------------|
| `Get` | 13.3% | ~80%+ |
| `Set` | 18.2% | ~80%+ |
| `Delete` | 28.6% | ~80%+ |
| `getKey` | 0.0% | 100% |

---

## Troubleshooting

### Redis Connection Refused

**Error**: `dial tcp [::1]:6379: connect: connection refused`

**Solution**:
1. Verify Redis is running: `redis-cli ping`
2. Check Redis is listening on port 6379: `lsof -i :6379`
3. Start Redis if not running: `brew services start redis` or `sudo systemctl start redis-server`

### Redis Not Found

**Error**: `command not found: redis-cli`

**Solution**:
1. Install Redis using one of the methods above
2. Add Redis to PATH if installed manually
3. Verify installation: `which redis-cli`

### Docker Redis Issues

**Error**: `Cannot connect to the Docker daemon`

**Solution**:
1. Start Docker Desktop (macOS/Windows)
2. Verify Docker is running: `docker ps`
3. Start Redis container: `docker run -d --name redis-test -p 6379:6379 redis:latest`

---

## CI/CD Integration

### GitHub Actions Example

```yaml
services:
  redis:
    image: redis:latest
    ports:
      - 6379:6379
    options: >-
      --health-cmd "redis-cli ping"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5

env:
  REDIS_URL: redis://localhost:6379
```

### GitLab CI Example

```yaml
services:
  - redis:latest

variables:
  REDIS_URL: redis://redis:6379
```

---

## Production Considerations

### Redis Configuration

For production use:
- Use Redis Sentinel for high availability
- Configure persistence (RDB or AOF)
- Set appropriate memory limits
- Use password authentication
- Enable TLS for secure connections

### Connection Pooling

The cache implementation uses connection pooling:
- Max connections: Configurable
- Connection timeout: 5 seconds
- Read/Write timeouts: Configurable

---

## Next Steps

1. **Set Up Redis**: Choose one of the installation methods above
2. **Run Tests**: Execute Redis-dependent tests
3. **Verify Coverage**: Check coverage improvements
4. **Document Results**: Update test results documentation

---

## Files

- **Cache Tests**: `services/classification-service/internal/cache/website_content_cache_critical_paths_test.go`
- **Integration Tests**: `services/classification-service/internal/integration/classification_optimizations_integration_test.go`
- **Documentation**: `docs/redis-setup-guide.md` (this document)

