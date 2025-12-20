# hrequests-service Build Failure Analysis
## December 20, 2025

---

## Problem Summary

**Service**: `hrequests_service` (Python microservice)  
**Status**: ❌ **BUILD FAILURE** - Health check failing  
**Impact**: Service cannot start, blocking website scraping functionality

---

## Root Cause

The `hrequests` Python library (v0.9.2) attempts to download a native library (`hrequests-cgo-3.1-linux-amd64.so`) from GitHub **at runtime** during import:

1. **Download Timeout**: `httpcore.ReadTimeout: The read operation timed out`
2. **Corrupted Download**: `OSError: file too short` (incomplete download)

**Error Location**: `services/hrequests-scraper/app.py:15` → `import hrequests`

---

## Error Details

### Primary Error
```
OSError: /usr/local/lib/python3.11/site-packages/hrequests/bin/hrequests-cgo-3.1-linux-amd64.so: file too short
```

### Download Attempt
```
Downloading hrequests-cgo library from daijro/hrequests...
httpcore.ReadTimeout: The read operation timed out
```

### Stack Trace
- `hrequests/__init__.py:31` → `from .reqs import *`
- `hrequests/reqs.py:14` → `from hrequests.response import Response`
- `hrequests/response.py:14` → `from hrequests.cffi import library`
- `hrequests/cffi.py:155` → `self.library: ctypes.CDLL = LibraryManager.load_library()`
- `hrequests/cffi.py:134` → `return ctypes.cdll.LoadLibrary(libman.full_path)`
- **FAILURE**: File is corrupted/incomplete

---

## Current Configuration

**File**: `services/hrequests-scraper/Dockerfile`
```dockerfile
FROM python:3.11-slim
WORKDIR /app
RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY app.py .
CMD ["python", "app.py"]
```

**File**: `services/hrequests-scraper/requirements.txt`
```
flask==3.0.0
beautifulsoup4==4.12.2
hrequests==0.9.2
lxml==4.9.3
```

**Issue**: The library download happens **at runtime** (when `import hrequests` is executed), not during Docker build.

---

## Solutions

### Option 1: Pre-download Library During Build (Recommended)

Modify `Dockerfile` to pre-download the library during build:

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Pre-download hrequests native library during build
# This prevents runtime download failures
RUN python -c "import hrequests; print('Library pre-loaded')" || \
    (echo "Pre-downloading library..." && \
     mkdir -p /usr/local/lib/python3.11/site-packages/hrequests/bin && \
     curl -L -o /usr/local/lib/python3.11/site-packages/hrequests/bin/hrequests-cgo-3.1-linux-amd64.so \
     https://github.com/daijro/hrequests/releases/download/v3.1/hrequests-cgo-3.1-linux-amd64.so && \
     python -c "import hrequests; print('Library verified')")

# Copy application code
COPY app.py .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/health').read()"

# Run the application
CMD ["python", "app.py"]
```

### Option 2: Use Alternative Library

Replace `hrequests` with a more stable alternative:
- `httpx` (similar API, no runtime downloads)
- `requests` (standard library, well-tested)
- `aiohttp` (async support)

### Option 3: Add Retry Logic in App

Modify `app.py` to retry library initialization:

```python
import os
import time
import logging

# Retry hrequests import with exponential backoff
max_retries = 3
for attempt in range(max_retries):
    try:
        import hrequests
        break
    except Exception as e:
        if attempt < max_retries - 1:
            wait_time = 2 ** attempt
            logging.warning(f"hrequests import failed (attempt {attempt + 1}/{max_retries}), retrying in {wait_time}s...")
            time.sleep(wait_time)
        else:
            raise
```

### Option 4: Pin to Stable Version

Check if a newer version of `hrequests` has fixed this issue, or use a version that doesn't require runtime downloads.

---

## Recommended Fix

**Option 1** is recommended because:
1. ✅ Pre-downloads library during build (faster startup)
2. ✅ Fails fast during build (not at runtime)
3. ✅ No code changes required
4. ✅ More reliable in containerized environments

---

## Impact Assessment

### Classification Service
- ✅ **NOT AFFECTED**: Classification service can still function
- ⚠️ **DEGRADED**: Website scraping will fall back to other strategies (SimpleHTTP, Playwright)
- ⚠️ **PERFORMANCE**: May be slower without hrequests (fastest scraping strategy)

### Other Services
- ✅ **NOT AFFECTED**: All other services are independent

---

## Next Steps

1. ⏳ **Implement Fix**: Apply Option 1 (pre-download during build)
2. ⏳ **Test**: Verify service starts successfully
3. ⏳ **Deploy**: Push fix to Railway
4. ⏳ **Monitor**: Verify health checks pass

---

## Related Issues

- This is a **separate issue** from the classification accuracy fixes
- Classification service fixes (Coca-Cola, Entertainment, Technology) are **unaffected**
- This only impacts the `hrequests-scraper` microservice

---

**Status**: ⚠️ **BLOCKING DEPLOYMENT** - Service cannot start  
**Priority**: **MEDIUM** - Service has fallback strategies  
**Date**: December 20, 2025

