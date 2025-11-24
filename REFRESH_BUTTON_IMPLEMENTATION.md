# Refresh Button Implementation for Business Analytics Tab

## Overview

Implemented functionality to trigger classification and website analysis jobs for existing merchants when the refresh button is clicked on the Business Analytics tab.

## Changes Made

### Backend Changes

#### 1. New Endpoint: `POST /api/v1/merchants/{id}/analytics/refresh`

**File**: `services/merchant-service/internal/handlers/merchant.go`

**Function**: `HandleTriggerAnalyticsRefresh()`

**Functionality**:
- Fetches merchant data from Supabase
- Triggers classification job for the merchant
- Triggers website analysis job if website URL is provided
- Marks website analysis as "skipped" if no website URL
- Returns status response with job information

**Response Format**:
```json
{
  "merchant_id": "merchant_123",
  "status": "triggered",
  "message": "Analytics refresh jobs have been triggered",
  "jobs": {
    "classification": "triggered",
    "website_analysis": {
      "status": "triggered" | "skipped",
      "reason": "website URL provided" | "no website URL"
    }
  },
  "timestamp": "2025-11-24T...",
  "processing_time": "50ms"
}
```

#### 2. Route Registration

**File**: `services/merchant-service/cmd/main.go`

Added route registration:
```go
router.HandleFunc("/api/v1/merchants/{id}/analytics/refresh", merchantHandler.HandleTriggerAnalyticsRefresh).Methods("POST", "OPTIONS")
```

**Note**: Route is registered before `/api/v1/merchants/{id}/analytics` to ensure proper matching.

### Frontend Changes

#### 1. API Function

**File**: `frontend/lib/api.ts`

**New Function**: `triggerAnalyticsRefresh(merchantId: string)`

- Makes POST request to `/api/v1/merchants/{id}/analytics/refresh`
- Returns `TriggerAnalyticsRefreshResponse` with job status
- Includes retry logic and error handling

**Type Definition**:
```typescript
export interface TriggerAnalyticsRefreshResponse {
  merchant_id: string;
  status: string;
  message: string;
  jobs: {
    classification: string;
    website_analysis: {
      status: string;
      reason: string;
    };
  };
  timestamp: string;
  processing_time: string;
}
```

#### 2. API Endpoint Configuration

**File**: `frontend/lib/api-config.ts`

Added endpoint:
```typescript
triggerAnalyticsRefresh: (id: string) => buildApiUrl(`/api/v1/merchants/${id}/analytics/refresh`),
```

#### 3. Refresh Handler Update

**File**: `frontend/components/merchant/BusinessAnalyticsTab.tsx`

**Updated Function**: `handleRefresh()`

**Previous Behavior**:
- Only reloaded existing analytics data
- Did not trigger new analysis jobs

**New Behavior**:
1. Calls `triggerAnalyticsRefresh()` to start background jobs
2. Waits 1 second for jobs to initialize
3. Reloads analytics data to show updated status
4. Status indicators automatically update via polling

**Implementation**:
```typescript
const handleRefresh = async () => {
  try {
    setIsRefreshing(true);
    setError(null);
    
    // Trigger analytics refresh jobs
    await triggerAnalyticsRefresh(merchantId);
    
    // Wait a moment for jobs to start, then reload data
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Reload analytics data
    await loadAnalytics(true);
  } catch (err) {
    setError(err instanceof Error ? err.message : 'Failed to trigger analytics refresh');
    setIsRefreshing(false);
  }
};
```

## User Experience

### Before
- Clicking refresh button only reloaded existing data
- No way to trigger analysis for existing merchants
- Status indicators showed "pending" indefinitely

### After
- Clicking refresh button:
  1. Triggers classification job (always)
  2. Triggers website analysis job (if website URL exists)
  3. Shows loading spinner while triggering
  4. Reloads data after 1 second
  5. Status indicators update automatically via polling
  6. User sees "Processing..." status immediately
  7. Status updates to "Completed" when jobs finish

## Status Indicator Integration

The existing `AnalyticsStatusIndicator` component automatically:
- Polls status endpoint every 3 seconds
- Shows "Processing..." badge while jobs run
- Updates to "Completed" when jobs finish
- Shows "Failed" if jobs error
- Shows "Skipped" for website analysis if no URL

## Testing

### Manual Testing Steps

1. Navigate to a merchant details page
2. Go to Business Analytics tab
3. Click the "Refresh" button
4. Verify:
   - Button shows loading spinner
   - Status indicators show "Processing..."
   - After completion, status updates to "Completed"
   - Analytics data refreshes with new results

### API Testing

```bash
# Trigger analytics refresh
curl -X POST "http://localhost:8082/api/v1/merchants/{merchant_id}/analytics/refresh" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}"

# Check status
curl "http://localhost:8082/api/v1/merchants/{merchant_id}/analytics/status" \
  -H "Authorization: Bearer {token}"
```

## Error Handling

- **Merchant Not Found**: Returns 404 with error message
- **Database Error**: Returns 500 with error message
- **Job Processor Not Initialized**: Logs warning, returns success (jobs will fail gracefully)
- **Network Error**: Frontend shows error message, allows retry

## Notes

- Jobs run asynchronously in background
- Response is immediate (HTTP 202 Accepted)
- Status can be checked via `/analytics/status` endpoint
- Frontend automatically polls status after triggering
- Website analysis is skipped if no website URL is provided
- Classification always runs (required for all merchants)

