# Python ML Service Models Endpoint Fix

## Issue Summary

The Python ML service was returning `503 Service Unavailable` when other services tried to access the `/models` endpoint during initialization. This caused other services to think the ML service wasn't ready, even though:

1. The service was running and responding to health checks
2. Models were loading lazily in the background
3. Classification requests worked after models loaded

## Root Cause

The `/models` endpoint was checking if `model_manager` was `None` and returning 503:

```python
@app.get("/models", response_model=List[ModelInfo])
async def get_models():
    """Get available models"""
    if model_manager is None:
        raise HTTPException(status_code=503, detail="Models are still loading. Please try again in a moment.")
    return model_manager.get_available_models()
```

The `model_manager` was initialized as `None` and only created lazily when the first classification request arrived. This meant:

- Other services calling `/models` during initialization got 503 errors
- The service appeared unavailable even though it was running
- Models were loading in the background but the endpoint didn't reflect this

## Solution

### 1. Initialize `model_manager` at Startup

Changed from:
```python
model_manager: Optional[ModelManager] = None
```

To:
```python
# Initialize model_manager at startup so /models endpoint doesn't return 503
# Models themselves are still loaded lazily on first request
model_manager: Optional[ModelManager] = ModelManager(load_models=False)
```

### 2. Make `/models` Return Empty List Instead of 503

Updated the endpoint to return an empty list when models aren't loaded yet:

```python
@app.get("/models", response_model=List[ModelInfo])
async def get_models():
    """Get available models"""
    # Initialize model_manager if somehow it's None (shouldn't happen, but defensive)
    if model_manager is None:
        global model_manager
        model_manager = ModelManager(load_models=False)
    
    # Return available models (empty list if models haven't been loaded yet)
    # This allows other services to check the endpoint without getting 503 errors
    return model_manager.get_available_models()
```

### 3. Fix `/models/{model_id}/metrics` Endpoint

Updated to handle the case where `model_manager` might be None (defensive):

```python
@app.get("/models/{model_id}/metrics", response_model=ModelMetrics)
async def get_model_metrics(model_id: str):
    """Get model metrics"""
    # Initialize model_manager if somehow it's None (shouldn't happen, but defensive)
    if model_manager is None:
        global model_manager
        model_manager = ModelManager(load_models=False)
    
    if model_id not in model_manager.models:
        raise HTTPException(status_code=404, detail="Model not found")
    # ... rest of function
```

### 4. Update Go Service to Handle 503 Gracefully

Added handling for 503 errors in the Go service (backward compatibility):

```go
// Handle 503 Service Unavailable (models still loading) - return empty list
if resp.StatusCode == http.StatusServiceUnavailable {
    pms.logger.Printf("⚠️ Models are still loading, returning empty list")
    return []*MLModel{}, nil
}
```

## Benefits

1. **No More 503 Errors**: The `/models` endpoint always returns 200 OK, with an empty list initially
2. **Better Service Discovery**: Other services can check the endpoint without errors
3. **Backward Compatible**: Go service handles 503 gracefully if Python service hasn't been updated
4. **Models Still Load Lazily**: Models are still loaded in the background, not blocking startup
5. **Defensive Programming**: Added checks to ensure `model_manager` is never None

## Testing

After deployment, verify:

1. `/models` endpoint returns 200 OK with empty list initially
2. `/models` endpoint returns populated list after models load
3. Other services can call `/models` without getting 503 errors
4. Classification requests still work and trigger model loading
5. Health checks continue to work

## Deployment Notes

- This is a backward-compatible change
- No breaking changes to the API
- Models still load lazily in the background
- Service starts immediately and accepts requests

