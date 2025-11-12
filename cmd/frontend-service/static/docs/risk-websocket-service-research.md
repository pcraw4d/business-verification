# Risk WebSocket Service - Research and Implementation Plan

## Service Purpose

The Risk WebSocket Service provides real-time updates for risk assessment data, enabling live monitoring of risk scores, predictions, and alerts without requiring page refreshes or polling.

### Benefits

1. **Real-time Updates**: Instant notification of risk score changes
2. **Live Monitoring**: Continuous risk assessment updates as data changes
3. **Reduced Server Load**: Push-based updates instead of polling
4. **Better User Experience**: Immediate feedback on risk changes
5. **Proactive Alerts**: Real-time risk alerts and notifications

## Implementation Details

### Backend Requirements

**Endpoint**: `/ws/risk-assessment/{assessment_id}`

**WebSocket Protocol**: 
- Connection: WebSocket (ws:// or wss://)
- Message Format: JSON
- Events:
  - `riskUpdate`: Risk score updated
  - `riskPrediction`: New risk prediction available
  - `riskAlert`: Risk alert triggered

### Frontend Integration

**Component**: `RiskWebSocketClient` (already exists in `js/components/risk-websocket-client.js`)

**Integration Points**:
1. Enable WebSocket connection in `MerchantRiskTab` (uncomment lines 67-71)
2. Add connection status indicator
3. Show "Real-time updates" badge when connected
4. Handle reconnection logic
5. Display real-time updates in UI

### UI Indicators

1. **Connection Status**: Visual indicator showing connection state (connected/disconnected)
2. **Real-time Badge**: Badge indicating "Real-time updates" when connected
3. **Reconnection Messages**: Status messages during reconnection attempts
4. **Update Notifications**: Toast notifications for important risk updates

## Implementation Steps

1. **Backend Verification** (2-3 hours)
   - Verify WebSocket endpoint exists and is accessible
   - Test WebSocket connection and message format
   - Document event structure and payload

2. **Frontend Integration** (5-8 hours)
   - Enable WebSocket connection in MerchantRiskTab
   - Add connection status indicator
   - Implement event handlers for riskUpdate, riskPrediction, riskAlert
   - Add reconnection logic with exponential backoff
   - Update UI components to reflect real-time data

3. **UI Enhancements** (3-5 hours)
   - Add connection status indicator
   - Create "Real-time updates" badge
   - Add reconnection status messages
   - Implement toast notifications for updates
   - Style and position indicators appropriately

4. **Testing** (3-5 hours)
   - Test WebSocket connection and disconnection
   - Test reconnection logic
   - Test event handling and UI updates
   - Test error scenarios
   - Cross-browser testing

5. **Documentation** (2-3 hours)
   - Document WebSocket API
   - Document event structure
   - Update user documentation

## Effort Estimate

**Total Estimated Time**: 15-23 hours

- Backend Verification: 2-3 hours
- Frontend Integration: 5-8 hours
- UI Enhancements: 3-5 hours
- Testing: 3-5 hours
- Documentation: 2-3 hours

## Decision

**Status**: Pending Approval

**Recommendation**: Implement for MVP if backend WebSocket service is available. If not available, defer to post-MVP phase.

**Priority**: Medium (enhances user experience but not critical for MVP)

## Implementation Code Reference

The WebSocket client is already implemented in:
- `cmd/frontend-service/static/js/components/risk-websocket-client.js`

Integration point in:
- `cmd/frontend-service/static/js/merchant-risk-tab.js` (lines 67-71, currently commented)

## Notes

- WebSocket service requires backend support
- Consider fallback to polling if WebSocket unavailable
- Implement graceful degradation for browsers without WebSocket support
- Monitor WebSocket connection health and performance

