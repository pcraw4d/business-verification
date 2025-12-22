# Cache Hit Rate Investigation - Track 8.1

## Executive Summary

Investigation of cache hit rate reveals **cache is enabled and configured**, but **cache hit rate patterns need analysis from logs**. Cache uses both Redis (if enabled) and in-memory fallback, with normalized cache keys to improve hit rates.

**Status**: ⚠️ **MEDIUM** - Cache configured correctly, but hit rate needs verification

## Cache Configuration

### Cache Settings

**Location**: `services/classification-service/internal/config/config.go:108-111`

**Configuration**:
- `CacheEnabled`: Default `true` (configurable via `CACHE_ENABLED`)
- `CacheTTL`: Default `10*time.Minute` (configurable via `CACHE_TTL`)
- `RedisEnabled`: Default `false` (configurable via `REDIS_ENABLED`)
- `RedisURL`: Empty by default (configurable via `REDIS_URL`)

**Status**: ✅ Cache enabled by default, TTL set to 10 minutes

### Cache Implementation

**Location**: `services/classification-service/internal/handlers/classification.go:740-818`

**Cache Strategy**:
1. **Redis Cache** (if enabled):
   - Distributed cache across service instances
   - Primary cache layer
   - Falls back to in-memory if Redis unavailable

2. **In-Memory Cache** (always):
   - Fallback if Redis disabled or unavailable
   - Local to each service instance
   - Cleaned up periodically

**Status**: ✅ Dual-layer caching implemented

## Cache Key Generation

### Key Generation Logic

**Location**: `services/classification-service/internal/handlers/classification.go:575-600`

**Key Generation**:
```go
// Normalize inputs: trim whitespace, lowercase for case-insensitive matching
businessName := strings.TrimSpace(strings.ToLower(req.BusinessName))
description := strings.TrimSpace(strings.ToLower(req.Description))
websiteURL := strings.TrimSpace(strings.ToLower(req.WebsiteURL))

// Create hash of normalized inputs
data := fmt.Sprintf("%s|%s|%s", businessName, description, websiteURL)
hash := sha256.Sum256([]byte(data))
cacheKey := fmt.Sprintf("classification:%x", hash)
```

**Key Characteristics**:
- **Normalization**: Lowercase, trimmed whitespace
- **Uniqueness**: SHA256 hash of normalized inputs
- **Format**: `classification:{hash}`
- **Collision Risk**: Very low (SHA256)

**Status**: ✅ Key generation looks correct

### Potential Issues

1. **Too Specific Keys** ⚠️ **MEDIUM**
   - Keys include business name, description, and website URL
   - Small variations in input create different keys
   - **Impact**: Lower cache hit rate for similar requests
   - **Example**: "Acme Corp" vs "Acme Corporation" = different keys

2. **Website URL Variations** ⚠️ **MEDIUM**
   - `https://example.com` vs `https://www.example.com` = different keys
   - Trailing slashes, query parameters = different keys
   - **Impact**: Lower cache hit rate for same website
   - **Recommendation**: Normalize URLs (remove www, trailing slashes, query params)

3. **Description Variations** ⚠️ **LOW**
   - Slight wording differences = different keys
   - **Impact**: Lower cache hit rate
   - **Recommendation**: Consider fuzzy matching or description normalization

## Cache Hit/Miss Patterns

### Log Patterns

**Cache Hit Logs**:
```
✅ [CACHE-HIT] Cache hit from Redis
✅ [CACHE-HIT] Cache hit from in-memory cache
✅ [CACHE-HIT] Classification served from cache
```

**Cache Miss Logs**:
```
❌ [CACHE-MISS] Cache miss, processing new request
Cache miss from Redis
Cache miss from in-memory cache
```

**Status**: ⏳ **NEEDS ANALYSIS** - Need to analyze Railway logs for hit/miss patterns

### Expected Cache Hit Rate

**Target**: >60% (from config comment: "improve cache hit rate from 49.6% to 60-70%")

**Factors Affecting Hit Rate**:
1. **Request Similarity**: More similar requests = higher hit rate
2. **Cache TTL**: Longer TTL = higher hit rate (but stale data)
3. **Traffic Patterns**: Repeat requests = higher hit rate
4. **Key Normalization**: Better normalization = higher hit rate

**Status**: ⏳ **NEEDS VERIFICATION** - Current hit rate unknown

## Cache Types

### 1. Classification Results Cache

**Location**: `services/classification-service/internal/handlers/classification.go:390-391`

**Storage**:
- Redis (if enabled)
- In-memory fallback

**Key Format**: `classification:{sha256_hash}`

**TTL**: 10 minutes (default)

**Status**: ✅ Implemented

### 2. Website Content Cache

**Location**: `services/classification-service/internal/cache/website_content_cache.go`

**Storage**:
- Redis (if enabled)
- In-memory fallback

**Key Format**: Website URL-based

**TTL**: 24 hours (default)

**Status**: ✅ Implemented

### 3. Code Metadata Cache

**Location**: `internal/classification/repository/supabase_repository.go`

**Storage**:
- In-memory cache
- Cache keys: `classification_codes:industry:{id}`, `classification_codes:type:{type}`

**TTL**: Not explicitly set (may be indefinite)

**Status**: ✅ Implemented

## Redis Cache Configuration

### Redis Setup

**Location**: `services/classification-service/internal/cache/redis_cache.go`

**Initialization**:
- Checks `REDIS_URL` environment variable
- Parses Redis URL
- Creates Redis client
- Falls back to in-memory if Redis unavailable

**Status**: ✅ Redis cache implemented with fallback

### Redis Health Check

**Location**: `services/classification-service/internal/cache/redis_cache.go:Health()`

**Health Check**:
- Pings Redis server
- Returns error if unavailable
- Service continues with in-memory cache if Redis fails

**Status**: ✅ Health check implemented

## Investigation Steps

### Step 1: Check Cache Configuration

**Verify Settings**:
- `CACHE_ENABLED` should be `true` (default)
- `CACHE_TTL` should be `10m` (default)
- `REDIS_ENABLED` should be `true` if Redis is available
- `REDIS_URL` should be set if Redis is enabled

**Status**: ⏳ **PENDING** - Need to verify in Railway

### Step 2: Analyze Cache Hit/Miss Patterns

**From Railway Logs**:
- Count `[CACHE-HIT]` log entries
- Count `[CACHE-MISS]` log entries
- Calculate hit rate: `hits / (hits + misses) * 100%`
- Identify patterns in cache misses

**Status**: ⏳ **PENDING** - Need to analyze logs

### Step 3: Review Cache Key Generation

**Test Scenarios**:
1. Same business name, different case → Should hit cache
2. Same business name, different whitespace → Should hit cache
3. Same website, different URL format → May miss cache
4. Similar descriptions → May miss cache

**Status**: ⏳ **PENDING** - Need to test

### Step 4: Check Redis Availability

**Verify**:
- Redis is deployed in Railway
- `REDIS_URL` is set correctly
- Redis health check passes
- Redis is being used (not just in-memory)

**Status**: ⏳ **PENDING** - Need to verify

## Root Cause Analysis

### Potential Issues

1. **Redis Not Enabled** ⚠️ **MEDIUM**
   - `REDIS_ENABLED` may be `false`
   - `REDIS_URL` may not be set
   - **Impact**: Only in-memory cache (not shared across instances)
   - **Evidence**: Default is `false`

2. **Cache Key Too Specific** ⚠️ **MEDIUM**
   - Keys include all inputs (name, description, URL)
   - Small variations create different keys
   - **Impact**: Lower cache hit rate
   - **Evidence**: Key generation includes all fields

3. **URL Normalization Missing** ⚠️ **MEDIUM**
   - URLs not normalized (www, trailing slashes, query params)
   - **Impact**: Same website = different cache keys
   - **Evidence**: URL used as-is in key generation

4. **Cache TTL Too Short** ⚠️ **LOW**
   - 10 minutes may be too short for some use cases
   - **Impact**: Lower hit rate
   - **Evidence**: TTL is 10 minutes (was 5 minutes, increased)

5. **Low Request Similarity** ⚠️ **LOW**
   - If requests are mostly unique, hit rate will be low
   - **Impact**: Low cache hit rate (expected)
   - **Evidence**: Need to analyze request patterns

## Recommendations

### Immediate Actions (High Priority)

1. **Verify Cache Configuration**:
   - Check `CACHE_ENABLED` is `true` in Railway
   - Check `CACHE_TTL` is set appropriately
   - Verify Redis is enabled if available

2. **Analyze Cache Hit Rate**:
   - Parse Railway logs for cache hit/miss patterns
   - Calculate actual hit rate
   - Identify patterns in cache misses

3. **Improve URL Normalization**:
   - Normalize URLs in cache key generation
   - Remove `www.`, trailing slashes, query parameters
   - **Expected Impact**: Higher cache hit rate for same websites

### Medium Priority Actions

4. **Optimize Cache Key Generation**:
   - Consider fuzzy matching for descriptions
   - Normalize business names (remove common suffixes)
   - **Expected Impact**: Higher cache hit rate for similar requests

5. **Enable Redis Cache**:
   - Set `REDIS_ENABLED=true` if Redis is available
   - Set `REDIS_URL` correctly
   - **Expected Impact**: Shared cache across instances, higher hit rate

6. **Add Cache Metrics**:
   - Track cache hit/miss rates
   - Monitor cache size
   - Alert on low hit rates

### Low Priority Actions

7. **Review Cache TTL**:
   - Consider increasing if hit rate is low
   - Balance freshness vs. hit rate
   - **Expected Impact**: Higher hit rate (but potentially stale data)

8. **Implement Cache Warming**:
   - Pre-populate cache with common requests
   - **Expected Impact**: Higher initial hit rate

## Code Locations

- **Cache Configuration**: `services/classification-service/internal/config/config.go:108-111`
- **Cache Implementation**: `services/classification-service/internal/handlers/classification.go:740-818`
- **Cache Key Generation**: `services/classification-service/internal/handlers/classification.go:575-600`
- **Redis Cache**: `services/classification-service/internal/cache/redis_cache.go`
- **Website Content Cache**: `services/classification-service/internal/cache/website_content_cache.go`

## Next Steps

1. ✅ **Complete Track 8.1 Investigation** - This document
2. **Verify Cache Configuration** - Check Railway settings
3. **Analyze Cache Hit Rate** - Parse logs for patterns
4. **Improve URL Normalization** - Normalize URLs in cache keys
5. **Enable Redis Cache** - If available
6. **Add Cache Metrics** - Track hit/miss rates

## Expected Impact

After fixing issues:

1. **Cache Hit Rate**: Current → >60% (target)
2. **Response Time**: Improved for cached requests
3. **Database Load**: Reduced with higher cache hit rate
4. **Service Performance**: Improved with distributed Redis cache

## References

- Cache Implementation: `services/classification-service/internal/handlers/classification.go`
- Redis Cache: `services/classification-service/internal/cache/redis_cache.go`
- Config: `services/classification-service/internal/config/config.go`

