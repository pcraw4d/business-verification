# Railway Redis URL Configuration Guide

**Date**: November 13, 2025  
**Status**: ✅ **CONFIRMED CORRECT**

---

## ✅ Redis URL Configuration

### Correct Configuration

**For all services in Railway:**
```bash
REDIS_URL=redis://redis-cache:6379
```

This is **correct** and should be set as a **shared environment variable** at the project level in Railway.

---

## 🔍 Why This Works

### Railway Service Discovery

Railway provides automatic service discovery within a project. Services can reference each other by their **service name**:

1. **Service Name**: `redis-cache` (as defined in `railway.json`)
2. **Internal DNS**: Railway automatically resolves `redis-cache` to the Redis service's internal IP
3. **Port**: `6379` (standard Redis port)
4. **Protocol**: `redis://` (Redis protocol)

### Network Architecture

```
┌─────────────────────────────────────────┐
│         Railway Project Network          │
│                                         │
│  ┌─────────────┐                       │
│  │ API Gateway │                       │
│  └──────┬──────┘                       │
│         │                               │
│  ┌──────▼──────┐                       │
│  │ Classification│                      │
│  └──────┬──────┘                       │
│         │                               │
│  ┌──────▼──────┐                       │
│  │   Merchant   │                      │
│  └──────┬──────┘                       │
│         │                               │
│  ┌──────▼──────┐                       │
│  │Risk Assessment│                      │
│  └──────┬──────┘                       │
│         │                               │
│         │ All connect to:              │
│         │                               │
│  ┌──────▼──────┐                       │
│  │ redis-cache │                       │
│  │   :6379     │                       │
│  └─────────────┘                       │
│                                         │
└─────────────────────────────────────────┘
```

---

## 📋 Service Configuration

### Shared Environment Variable

Set at **Railway Project Level** (applies to all services):

```bash
REDIS_URL=redis://redis-cache:6379
```

### Service-Specific (if needed)

If a service needs different Redis configuration, you can override at the service level:

```bash
# Example: Different Redis DB
REDIS_DB=1

# Example: Custom pool size
REDIS_POOL_SIZE=100
```

---

## 🔧 How Services Use REDIS_URL

### Risk Assessment Service

The service parses `REDIS_URL` and extracts the address:

```go
// Input: redis://redis-cache:6379
// After parsing: redis-cache:6379
// Used as: Addrs: []string{"redis-cache:6379"}
```

**Code Flow**:
1. Reads `REDIS_URL` from environment
2. Removes `redis://` prefix
3. Uses `redis-cache:6379` as the connection address
4. Railway DNS resolves `redis-cache` to the Redis service IP

### Merchant Service

Uses `REDIS_URL` directly for Redis client connection.

### Other Services

All services that use Redis should use the same format.

---

## ✅ Verification

### 1. Check Environment Variables

In Railway Dashboard:
- Go to **Project Settings** → **Variables**
- Verify `REDIS_URL=redis://redis-cache:6379` is set
- Ensure it's marked as **shared** (applies to all services)

### 2. Check Service Logs

After services start, you should see:

**Risk Assessment Service**:
```
🔧 Initializing Redis cache addrs: ["redis-cache:6379"] redis_url: "redis://redis-cache:6379" db: 0
✅ Risk Assessment Service Redis cache initialized successfully
```

**Merchant Service**:
```
Redis cache initialized addr: "redis-cache:6379" db: 0 pool_size: 10
```

### 3. Test Connection

If you have access to a service container:
```bash
redis-cli -h redis-cache ping
# Should return: PONG
```

---

## ⚠️ Important Notes

### 1. Service Name Must Match

The service name in `railway.json` must be exactly `redis-cache`:

```json
{
  "name": "redis-cache",
  ...
}
```

### 2. Internal Network Only

- `redis://redis-cache:6379` works **only within Railway's internal network**
- External access requires exposing the service (not recommended for Redis)
- Services outside the Railway project cannot use this URL

### 3. No Public URL Needed

- Redis service is **unexposed** (internal only) ✅
- This is the **correct** configuration for security
- Services connect via internal service discovery

### 4. URL Format

The format `redis://redis-cache:6379` is correct:
- `redis://` - Protocol prefix
- `redis-cache` - Service name (Railway DNS resolves this)
- `:6379` - Port number

---

## 🔄 Alternative Formats (Not Recommended)

### External URL (if Redis was exposed)

If Redis was exposed publicly (not recommended):
```bash
REDIS_URL=redis://redis-cache-production.up.railway.app:6379
```

**Why not recommended**:
- Security risk (Redis should be internal)
- Slower (goes through external network)
- Unnecessary (internal service discovery is faster)

### With Password (if Redis had password)

If Redis required authentication:
```bash
REDIS_URL=redis://:password@redis-cache:6379
```

**Current setup**: Redis doesn't require password (internal network is secure)

---

## 📊 Current Configuration Status

✅ **Correct Configuration**:
- `REDIS_URL=redis://redis-cache:6379` set at project level
- Redis service name: `redis-cache`
- Redis port: `6379`
- Redis service: Unexposed (internal only)
- All services: Using shared `REDIS_URL`

---

## 🎯 Summary

**Question**: Is `REDIS_URL=redis://redis-cache:6379` correct?

**Answer**: ✅ **YES, this is correct!**

- Railway uses service discovery
- Services reference each other by service name
- `redis-cache` is the service name
- `6379` is the standard Redis port
- Internal network provides secure communication
- No public URL needed (and shouldn't be exposed)

**Action**: No changes needed. The current configuration is correct.

---

**Last Updated**: November 13, 2025

