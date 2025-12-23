# Resource Constraints Investigation - Track 8.2

## Executive Summary

Investigation of resource constraints reveals **concurrent request limits are set to prevent OOM kills**, but **Railway resource limits and memory usage patterns need verification**. The service has been optimized to reduce memory pressure (50% reduction), but actual resource usage needs monitoring.

**Status**: ⚠️ **MEDIUM** - Limits configured, but actual usage needs verification

## Railway Resource Limits

### Service Configuration

**Location**: `railway.json`

**Classification Service**:
- No explicit memory limit set
- No explicit CPU limit set
- Health check timeout: 30s
- Restart policy: ON_FAILURE (max 10 retries)

**Status**: ⏳ **NEEDS VERIFICATION** - Default Railway limits apply

### Memory Limit Configuration

**Location**: `services/classification-service/cmd/main.go:73`

**Function**: `applyMemoryLimit()`

**Purpose**: Apply Go memory limit if provided (helps avoid OOM kills on Railway)

**Status**: ✅ Memory limit can be set via `GOMEMLIMIT` environment variable

**Expected Usage**:
- Set `GOMEMLIMIT` to limit Go runtime memory
- Helps prevent OOM kills
- **Status**: ⏳ **NEEDS VERIFICATION** - May not be set in Railway

## Concurrent Request Limits

### Configuration

**Location**: `services/classification-service/internal/config/config.go:105`

**Setting**: `MaxConcurrentRequests: 20` (default)

**Comment**: "Reduced from 40 to 20 to prevent OOM kills (50% reduction in memory pressure)"

**Environment Variable**: `MAX_CONCURRENT_REQUESTS`

**Status**: ✅ Limit configured to prevent OOM kills

### Request Queue

**Location**: `services/classification-service/internal/handlers/classification.go:395, 421-424`

**Queue Configuration**:
- Max queue size: `MaxConcurrentRequests` (default: 20)
- Worker pool: 30% of `MaxConcurrentRequests` (default: 6 workers)
- **Status**: ✅ Queue implemented to manage concurrent requests

### Queue Management

**Location**: `services/classification-service/internal/handlers/classification.go:1142-1146`

**Queue Check**:
```go
if h.requestQueue != nil && h.requestQueue.Size() >= h.config.Classification.MaxConcurrentRequests {
    // Reject request if queue is full
    http.Error(w, "Service busy, please try again later", http.StatusServiceUnavailable)
}
```

**Status**: ✅ Queue size check implemented to prevent overload

## Memory Usage Analysis

### Memory Optimization

**Previous Optimization**:
- Reduced `MaxConcurrentRequests` from 40 to 20 (50% reduction)
- Purpose: Prevent OOM kills
- **Status**: ✅ Optimization applied

### Potential Memory Leaks

**Areas to Check**:
1. **Goroutine Leaks** ⚠️ **MEDIUM**
   - Background goroutines not being cleaned up
   - **Impact**: Memory growth over time
   - **Evidence**: Need to check goroutine usage

2. **Cache Growth** ⚠️ **LOW**
   - In-memory cache growing unbounded
   - **Impact**: Memory growth
   - **Evidence**: Cache cleanup implemented (periodic cleanup)

3. **Unclosed Connections** ⚠️ **MEDIUM**
   - HTTP connections not being closed
   - Database connections not being closed
   - **Impact**: Memory and connection leaks
   - **Evidence**: Need to verify connection cleanup

4. **Large Response Buffers** ⚠️ **LOW**
   - Large response bodies in memory
   - **Impact**: Memory spikes
   - **Evidence**: Need to check response handling

## CPU Usage Analysis

### CPU Limits

**Railway Configuration**:
- No explicit CPU limit set in `railway.json`
- Default Railway limits apply
- **Status**: ⏳ **NEEDS VERIFICATION** - Default limits unknown

### CPU-Intensive Operations

**Potential Bottlenecks**:
1. **Website Scraping** ⚠️ **MEDIUM**
   - Multiple concurrent scrapes
   - **Impact**: High CPU usage
   - **Evidence**: Concurrent pages: 5 (default)

2. **ML Classification** ⚠️ **MEDIUM**
   - Python ML service calls
   - **Impact**: Network I/O, but may be CPU-intensive
   - **Evidence**: ML service calls are async

3. **Database Queries** ⚠️ **LOW**
   - Multiple concurrent queries
   - **Impact**: CPU usage for query processing
   - **Evidence**: Parallel code queries implemented

## OOM (Out of Memory) Kills

### OOM Detection

**Check Railway Logs For**:
- "Out of memory" errors
- "OOM kill" messages
- Service restarts
- Memory limit exceeded errors

**Status**: ⏳ **PENDING** - Need to check Railway logs

### OOM Prevention

**Current Measures**:
1. **Concurrent Request Limit**: 20 (reduced from 40)
2. **Memory Limit**: Can be set via `GOMEMLIMIT`
3. **Request Queue**: Prevents overload
4. **Cache Cleanup**: Periodic cleanup of expired entries

**Status**: ✅ Multiple prevention measures in place

## Investigation Steps

### Step 1: Check Railway Resource Limits

**Verify**:
- Memory limit (if set)
- CPU limit (if set)
- Request timeout limits
- Health check configuration

**Status**: ⏳ **PENDING** - Need to check Railway dashboard

### Step 2: Analyze Memory Usage

**From Railway Logs**:
- Check for OOM kills
- Review memory usage patterns
- Identify memory spikes
- Check for memory leaks

**Status**: ⏳ **PENDING** - Need to analyze logs

### Step 3: Check Concurrent Request Limits

**Verify**:
- `MAX_CONCURRENT_REQUESTS` is set appropriately
- Queue is not constantly full
- Requests are not being rejected unnecessarily

**Status**: ⏳ **PENDING** - Need to verify

### Step 4: Review Goroutine Usage

**Check**:
- Number of goroutines
- Goroutine leaks
- Background goroutine cleanup

**Status**: ⏳ **PENDING** - Need to add monitoring

## Root Cause Analysis

### Potential Issues

1. **Memory Limit Not Set** ⚠️ **MEDIUM**
   - `GOMEMLIMIT` may not be set
   - **Impact**: Go runtime may use too much memory
   - **Evidence**: Memory limit is optional

2. **Concurrent Request Limit Too Low** ⚠️ **LOW**
   - Limit of 20 may be too restrictive
   - **Impact**: Requests queued or rejected
   - **Evidence**: Reduced from 40 to prevent OOM

3. **Memory Leaks** ⚠️ **MEDIUM**
   - Goroutine leaks possible
   - Unclosed connections possible
   - **Impact**: Memory growth over time
   - **Evidence**: Need to verify

4. **Cache Growth** ⚠️ **LOW**
   - In-memory cache may grow unbounded
   - **Impact**: Memory growth
   - **Evidence**: Cache cleanup implemented

5. **High Memory Usage Per Request** ⚠️ **MEDIUM**
   - Each request may use significant memory
   - **Impact**: OOM kills with many concurrent requests
   - **Evidence**: Reduced concurrent limit to prevent OOM

## Recommendations

### Immediate Actions (High Priority)

1. **Set Memory Limit**:
   - Set `GOMEMLIMIT` in Railway
   - Use appropriate value (e.g., 512MB, 1GB)
   - **Expected Impact**: Prevents OOM kills

2. **Monitor Memory Usage**:
   - Add memory usage metrics
   - Track memory over time
   - Alert on high memory usage

3. **Check Railway Logs for OOM**:
   - Review logs for OOM kills
   - Identify patterns
   - Document findings

### Medium Priority Actions

4. **Review Concurrent Request Limit**:
   - Test with different limits
   - Monitor queue depth
   - Adjust if needed

5. **Fix Memory Leaks**:
   - Review goroutine usage
   - Ensure connections are closed
   - Fix any leaks found

6. **Optimize Memory Usage**:
   - Reduce memory per request
   - Optimize data structures
   - Use streaming where possible

### Low Priority Actions

7. **Add Resource Monitoring**:
   - Track CPU usage
   - Track memory usage
   - Track goroutine count
   - Alert on resource constraints

8. **Review Cache Memory Usage**:
   - Monitor cache size
   - Set cache size limits
   - Implement cache eviction

## Code Locations

- **Memory Limit**: `services/classification-service/cmd/main.go:73`
- **Concurrent Requests**: `services/classification-service/internal/config/config.go:105`
- **Request Queue**: `services/classification-service/internal/handlers/classification.go:395, 421-424`
- **Queue Check**: `services/classification-service/internal/handlers/classification.go:1142-1146`

## Next Steps

1. ✅ **Complete Track 8.2 Investigation** - This document
2. **Check Railway Resource Limits** - Review dashboard
3. **Analyze Memory Usage** - Review logs for OOM
4. **Set Memory Limit** - Configure `GOMEMLIMIT`
5. **Monitor Resource Usage** - Add metrics
6. **Review Concurrent Limits** - Test and adjust

## Expected Impact

After fixing issues:

1. **OOM Kills**: Reduced with memory limit
2. **Memory Usage**: Optimized with limits and monitoring
3. **Request Throughput**: Improved with appropriate limits
4. **Service Stability**: Improved with resource management

## References

- Config: `services/classification-service/internal/config/config.go`
- Memory Limit: `services/classification-service/cmd/main.go:73`
- Request Queue: `services/classification-service/internal/handlers/classification.go`


