# Merchant Details Page - API Endpoints Documentation

## Overview

This document provides a comprehensive inventory of all API endpoints used by the merchant-details page, including their purpose, request/response structures, and data flow.

**Last Updated:** December 19, 2024  
**Page:** `cmd/frontend-service/static/merchant-details.html`

---

## Data Flow Overview

```
add-merchant form → API → Database → Session Storage → merchant-details page
```

### Data Sources

1. **Session Storage**: Primary source for merchant data from form submission
   - `merchantData`: Form submission data
   - `merchantApiResults`: API response data

2. **API Endpoints**: Secondary source for real-time data and additional information

---

## API Endpoints Inventory

### 1. Base Merchant Data

#### `GET /api/v1/merchants/{merchantId}`
- **Purpose**: Retrieve base merchant information
- **Used By**: MerchantContext component, RealDataIntegration
- **Request**: 
  ```javascript
  GET /api/v1/merchants/{merchantId}
  Headers: {
    'Authorization': 'Bearer {token}',
    'Content-Type': 'application/json'
  }
  ```
- **Response Structure**:
  ```json
  {
    "id": "merchant_id",
    "businessName": "Company Name",
    "industry": "Industry",
    "status": "active",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
  ```
- **Status**: Currently loaded from session storage, API endpoint available for future use

---

### 2. Risk Assessment Endpoints

#### `GET /api/v1/merchants/{merchantId}/risk-score`
- **Purpose**: Retrieve risk score data for merchant
- **Used By**: Risk Score Panel component
- **Location**: `merchant-details.html` line 2551
- **Request**:
  ```javascript
  GET /api/v1/merchants/${merchantId}/risk-score
  Headers: {
    'Authorization': `Bearer ${authToken}`
  }
  ```
- **Response Structure**:
  ```json
  {
    "merchantId": "merchant_id",
    "overallScore": 0.75,
    "riskLevel": "medium",
    "factors": [
      {
        "name": "Financial Stability",
        "score": 0.8,
        "weight": 0.3
      }
    ],
    "lastUpdated": "2024-01-01T00:00:00Z"
  }
  ```
- **Error Handling**: Shows placeholder message if data unavailable
- **Status**: ✅ Implemented with error handling

#### `GET /api/v1/merchants/{merchantId}/website-risk`
- **Purpose**: Retrieve website risk assessment data
- **Used By**: Website Risk Display component
- **Location**: `merchant-details.html` line 2584
- **Request**:
  ```javascript
  GET /api/v1/merchants/${merchantId}/website-risk
  Headers: {
    'Authorization': `Bearer ${authToken}`
  }
  ```
- **Response Structure**:
  ```json
  {
    "merchantId": "merchant_id",
    "websiteUrl": "https://example.com",
    "riskScore": 0.65,
    "indicators": [
      {
        "type": "ssl",
        "status": "valid",
        "score": 0.9
      }
    ],
    "lastAnalyzed": "2024-01-01T00:00:00Z"
  }
  ```
- **Error Handling**: Shows placeholder message if data unavailable
- **Status**: ✅ Implemented with error handling

---

### 3. Data Enrichment Endpoints

#### `GET /api/v1/merchants/{merchantId}/enrichment/sources`
- **Purpose**: Get available enrichment sources
- **Used By**: Data Enrichment component
- **Location**: `js/components/data-enrichment.js`
- **Request**:
  ```javascript
  GET /api/v1/merchants/${merchantId}/enrichment/sources
  ```
- **Response Structure**:
  ```json
  {
    "sources": [
      {
        "id": "thomson-reuters",
        "name": "Thomson Reuters",
        "status": "available",
        "description": "Business data enrichment"
      }
    ]
  }
  ```
- **Status**: ✅ Component initialized, UI integrated

#### `POST /api/v1/merchants/{merchantId}/enrichment/trigger`
- **Purpose**: Trigger data enrichment from external sources
- **Used By**: Data Enrichment component
- **Request**:
  ```javascript
  POST /api/v1/merchants/${merchantId}/enrichment/trigger
  Body: {
    "source": "thomson-reuters"
  }
  ```
- **Response Structure**:
  ```json
  {
    "status": "processing",
    "jobId": "enrichment_job_id",
    "estimatedCompletion": "2024-01-01T00:05:00Z"
  }
  ```
- **Status**: ✅ Component initialized, UI integrated

---

### 4. External Data Sources Endpoints

#### `GET /api/v1/merchants/{merchantId}/external-sources`
- **Purpose**: Get list of external data sources and their status
- **Used By**: External Data Sources component
- **Request**:
  ```javascript
  GET /api/v1/merchants/${merchantId}/external-sources
  ```
- **Response Structure**:
  ```json
  {
    "sources": [
      {
        "id": "source_id",
        "name": "Source Name",
        "status": "active",
        "lastSync": "2024-01-01T00:00:00Z",
        "dataPoints": 150
      }
    ]
  }
  ```
- **Status**: ✅ Component initialized, UI integrated

---

### 5. Risk Indicators Endpoints (From API Config)

#### `GET /api/v1/merchants/{merchantId}/risk-indicators`
- **Purpose**: Get risk indicators for merchant
- **Used By**: Risk Indicators tab
- **Location**: `js/api-config.js` line 51
- **Request**:
  ```javascript
  GET /api/v1/merchants/${merchantId}/risk-indicators
  ```
- **Response Structure**: (To be documented)
- **Status**: ⚠️ Endpoint defined in config, implementation pending

#### `GET /api/v1/merchants/{merchantId}/website-analysis`
- **Purpose**: Get website analysis data
- **Used By**: Business Analytics tab
- **Location**: `js/api-config.js` line 52
- **Request**:
  ```javascript
  GET /api/v1/merchants/${merchantId}/website-analysis
  ```
- **Response Structure**: (To be documented)
- **Status**: ⚠️ Endpoint defined in config, implementation pending

#### `GET /api/v1/merchants/{merchantId}/risk-recommendations`
- **Purpose**: Get risk mitigation recommendations
- **Used By**: Risk Assessment tab
- **Location**: `js/api-config.js` line 53
- **Request**:
  ```javascript
  GET /api/v1/merchants/${merchantId}/risk-recommendations
  ```
- **Response Structure**: (To be documented)
- **Status**: ⚠️ Endpoint defined in config, implementation pending

#### `GET /api/v1/merchants/{merchantId}/risk-alerts`
- **Purpose**: Get active risk alerts
- **Used By**: Risk Indicators tab
- **Location**: `js/api-config.js` line 54
- **Request**:
  ```javascript
  GET /api/v1/merchants/${merchantId}/risk-alerts
  ```
- **Response Structure**: (To be documented)
- **Status**: ⚠️ Endpoint defined in config, implementation pending

---

### 6. Analytics Endpoints (From API Config)

#### `GET /api/v1/merchants/analytics`
- **Purpose**: Get merchant analytics data
- **Used By**: Business Analytics tab
- **Location**: `js/api-config.js` line 36
- **Request**:
  ```javascript
  GET /api/v1/merchants/analytics?merchantId={merchantId}
  ```
- **Response Structure**: (To be documented)
- **Status**: ⚠️ Endpoint defined in config, implementation pending

---

### 7. Risk Assessment Endpoints (From API Config)

#### `POST /api/v1/risk/assess`
- **Purpose**: Assess merchant risk
- **Used By**: Risk Assessment tab
- **Location**: `js/api-config.js` line 43
- **Request**:
  ```javascript
  POST /api/v1/risk/assess
  Body: {
    "merchantId": "merchant_id"
  }
  ```
- **Response Structure**: (To be documented)
- **Status**: ⚠️ Endpoint defined in config, implementation pending

#### `GET /api/v1/risk/history/{merchantId}`
- **Purpose**: Get risk assessment history
- **Used By**: Risk Assessment tab
- **Location**: `js/api-config.js` line 44
- **Request**:
  ```javascript
  GET /api/v1/risk/history/${merchantId}
  ```
- **Response Structure**: (To be documented)
- **Status**: ⚠️ Endpoint defined in config, implementation pending

#### `GET /api/v1/risk/predictions/{merchantId}`
- **Purpose**: Get risk predictions
- **Used By**: Risk Assessment tab
- **Location**: `js/api-config.js` line 45
- **Request**:
  ```javascript
  GET /api/v1/risk/predictions/${merchantId}
  ```
- **Response Structure**: (To be documented)
- **Status**: ⚠️ Endpoint defined in config, implementation pending

---

## Tab-to-Endpoint Mapping

### Overview Tab
- **Data Source**: Session Storage (`merchantData`, `merchantApiResults`)
- **API Endpoints**: None (uses session storage)
- **Status**: ✅ Implemented

### Business Analytics Tab
- **Data Source**: Session Storage + API
- **API Endpoints**:
  - `/api/v1/merchants/analytics` (pending)
  - `/api/v1/merchants/{merchantId}/website-analysis` (pending)
- **Status**: ⚠️ Partially implemented (uses session storage, API endpoints pending)

### Risk Assessment Tab
- **Data Source**: API
- **API Endpoints**:
  - `/api/v1/merchants/{merchantId}/risk-score` ✅
  - `/api/v1/merchants/{merchantId}/website-risk` ✅
  - `/api/v1/risk/assess` (pending)
  - `/api/v1/risk/history/{merchantId}` (pending)
  - `/api/v1/risk/predictions/{merchantId}` (pending)
  - `/api/v1/merchants/{merchantId}/risk-recommendations` (pending)
- **Status**: ⚠️ Partially implemented (risk-score and website-risk working, others pending)

### Risk Indicators Tab
- **Data Source**: API
- **API Endpoints**:
  - `/api/v1/merchants/{merchantId}/risk-indicators` (pending)
  - `/api/v1/merchants/{merchantId}/risk-alerts` (pending)
- **Status**: ⚠️ Endpoints defined, implementation pending

---

## Authentication

All API endpoints require authentication via Bearer token:
```javascript
Headers: {
  'Authorization': `Bearer ${authToken}`
}
```

Token is retrieved from:
- `sessionStorage.getItem('authToken')` (primary)
- Cookie fallback (if implemented)

---

## Error Handling

### Current Implementation
- Risk Score Panel: Shows placeholder message on error
- Website Risk Display: Shows placeholder message on error
- Data Enrichment: Shows error state in UI
- External Data Sources: Shows error state in UI

### Recommended Improvements
1. Implement consistent error handling across all endpoints
2. Add retry logic for failed requests
3. Implement exponential backoff for rate-limited endpoints
4. Add user-friendly error messages
5. Log errors for debugging

---

## Data Mapping: Form Fields → Database → Display

### Form Fields (add-merchant.html)
- Business Name → `merchantData.businessName`
- Business Address → `merchantData.address`
- Phone Number → `merchantData.phone`
- Email → `merchantData.email`
- Website → `merchantData.website`
- Industry → `merchantData.industry`
- Revenue → `merchantData.revenue`
- Employee Count → `merchantData.employeeCount`

### Database Columns (Inferred)
- `merchants.id` → Merchant ID
- `merchants.business_name` → Business Name
- `merchants.address` → Address
- `merchants.phone` → Phone
- `merchants.email` → Email
- `merchants.website` → Website
- `merchants.industry` → Industry
- `merchants.revenue` → Revenue
- `merchants.employee_count` → Employee Count
- `merchants.created_at` → Created Date
- `merchants.updated_at` → Updated Date

### Display Components
- **Overview Card**: Business Name, Industry, Status, Key Metrics
- **Contact Card**: Address, Phone, Email, Website
- **Financial Card**: Revenue, Employee Count, Founded Year, Transactions
- **Compliance Card**: KYB Status, Verification Date, Compliance Score, Certifications

---

## Next Steps

1. **Backend Verification**: Verify all endpoints are implemented and accessible
2. **Response Structure Documentation**: Document actual response structures from backend
3. **Error Response Handling**: Implement consistent error handling
4. **Data Flow Testing**: Test complete flow from form to display
5. **Mock Data Fallback**: Implement mock data for unavailable endpoints
6. **API Integration**: Connect all tabs to their respective endpoints
7. **Performance Optimization**: Implement caching and request batching

---

## Notes

- Most data currently comes from session storage (form submission)
- API endpoints are defined but many are not yet fully integrated
- Risk Score and Website Risk endpoints are working with error handling
- Data Enrichment and External Data Sources components are initialized but need backend integration
- Tab-specific endpoints need to be connected to their respective tabs

