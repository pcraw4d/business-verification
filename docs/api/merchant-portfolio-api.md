# Merchant Portfolio Management API Documentation

**Version**: 1.0.0  
**Base URL**: `/api/v1/merchants`  
**Last Updated**: January 2025  

## Overview

The Merchant Portfolio Management API provides comprehensive functionality for managing merchant portfolios in the KYB Platform. This API supports CRUD operations, search and filtering, bulk operations, session management, and analytics for merchant data.

## Authentication

All API endpoints require authentication using Bearer tokens in the Authorization header:

```http
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

- **Standard endpoints**: 100 requests per minute per user
- **Bulk operations**: 10 requests per minute per user
- **Search endpoints**: 200 requests per minute per user

Rate limit headers are included in all responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the rate limit resets

## Data Models

### Merchant

```json
{
  "id": "string",
  "name": "string",
  "legal_name": "string",
  "registration_number": "string",
  "tax_id": "string",
  "industry": "string",
  "industry_code": "string",
  "business_type": "string",
  "founded_date": "2023-01-01T00:00:00Z",
  "employee_count": 50,
  "annual_revenue": 1000000.00,
  "address": {
    "street1": "string",
    "street2": "string",
    "city": "string",
    "state": "string",
    "postal_code": "string",
    "country": "string",
    "country_code": "string"
  },
  "contact_info": {
    "phone": "string",
    "email": "string",
    "website": "string",
    "primary_contact": "string"
  },
  "portfolio_type": "onboarded|deactivated|prospective|pending",
  "risk_level": "high|medium|low",
  "compliance_status": "string",
  "status": "string",
  "created_by": "string",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### Portfolio Types

- `onboarded`: Active merchants in the portfolio
- `deactivated`: Merchants that have been deactivated
- `prospective`: Potential merchants under consideration
- `pending`: Merchants awaiting approval

### Risk Levels

- `high`: High-risk merchants requiring additional monitoring
- `medium`: Medium-risk merchants with standard monitoring
- `low`: Low-risk merchants with minimal monitoring

## Endpoints

### Merchant CRUD Operations

#### Create Merchant

**POST** `/api/v1/merchants`

Creates a new merchant in the portfolio.

**Request Body:**
```json
{
  "name": "Acme Corporation",
  "legal_name": "Acme Corporation LLC",
  "registration_number": "REG123456",
  "tax_id": "TAX789012",
  "industry": "Technology",
  "industry_code": "541511",
  "business_type": "Corporation",
  "founded_date": "2020-01-15T00:00:00Z",
  "employee_count": 25,
  "annual_revenue": 500000.00,
  "address": {
    "street1": "123 Main Street",
    "street2": "Suite 100",
    "city": "San Francisco",
    "state": "CA",
    "postal_code": "94105",
    "country": "United States",
    "country_code": "US"
  },
  "contact_info": {
    "phone": "+1-555-123-4567",
    "email": "contact@acme.com",
    "website": "https://www.acme.com",
    "primary_contact": "John Doe"
  },
  "portfolio_type": "prospective",
  "risk_level": "medium"
}
```

**Response:**
```json
{
  "id": "merchant_1234567890",
  "name": "Acme Corporation",
  "legal_name": "Acme Corporation LLC",
  "registration_number": "REG123456",
  "tax_id": "TAX789012",
  "industry": "Technology",
  "industry_code": "541511",
  "business_type": "Corporation",
  "founded_date": "2020-01-15T00:00:00Z",
  "employee_count": 25,
  "annual_revenue": 500000.00,
  "address": {
    "street1": "123 Main Street",
    "street2": "Suite 100",
    "city": "San Francisco",
    "state": "CA",
    "postal_code": "94105",
    "country": "United States",
    "country_code": "US"
  },
  "contact_info": {
    "phone": "+1-555-123-4567",
    "email": "contact@acme.com",
    "website": "https://www.acme.com",
    "primary_contact": "John Doe"
  },
  "portfolio_type": "prospective",
  "risk_level": "medium",
  "compliance_status": "pending",
  "status": "active",
  "created_by": "user_123",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

**Status Codes:**
- `201 Created`: Merchant created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `409 Conflict`: Merchant already exists
- `429 Too Many Requests`: Rate limit exceeded

#### Get Merchant

**GET** `/api/v1/merchants/{id}`

Retrieves a specific merchant by ID.

**Path Parameters:**
- `id` (string, required): Merchant ID

**Response:**
```json
{
  "id": "merchant_1234567890",
  "name": "Acme Corporation",
  "legal_name": "Acme Corporation LLC",
  "registration_number": "REG123456",
  "tax_id": "TAX789012",
  "industry": "Technology",
  "industry_code": "541511",
  "business_type": "Corporation",
  "founded_date": "2020-01-15T00:00:00Z",
  "employee_count": 25,
  "annual_revenue": 500000.00,
  "address": {
    "street1": "123 Main Street",
    "street2": "Suite 100",
    "city": "San Francisco",
    "state": "CA",
    "postal_code": "94105",
    "country": "United States",
    "country_code": "US"
  },
  "contact_info": {
    "phone": "+1-555-123-4567",
    "email": "contact@acme.com",
    "website": "https://www.acme.com",
    "primary_contact": "John Doe"
  },
  "portfolio_type": "onboarded",
  "risk_level": "medium",
  "compliance_status": "compliant",
  "status": "active",
  "created_by": "user_123",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T15:45:00Z"
}
```

**Status Codes:**
- `200 OK`: Merchant retrieved successfully
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Merchant not found
- `429 Too Many Requests`: Rate limit exceeded

#### Update Merchant

**PUT** `/api/v1/merchants/{id}`

Updates an existing merchant.

**Path Parameters:**
- `id` (string, required): Merchant ID

**Request Body:**
```json
{
  "name": "Acme Corporation Updated",
  "portfolio_type": "onboarded",
  "risk_level": "low",
  "compliance_status": "compliant"
}
```

**Response:**
```json
{
  "id": "merchant_1234567890",
  "name": "Acme Corporation Updated",
  "legal_name": "Acme Corporation LLC",
  "registration_number": "REG123456",
  "tax_id": "TAX789012",
  "industry": "Technology",
  "industry_code": "541511",
  "business_type": "Corporation",
  "founded_date": "2020-01-15T00:00:00Z",
  "employee_count": 25,
  "annual_revenue": 500000.00,
  "address": {
    "street1": "123 Main Street",
    "street2": "Suite 100",
    "city": "San Francisco",
    "state": "CA",
    "postal_code": "94105",
    "country": "United States",
    "country_code": "US"
  },
  "contact_info": {
    "phone": "+1-555-123-4567",
    "email": "contact@acme.com",
    "website": "https://www.acme.com",
    "primary_contact": "John Doe"
  },
  "portfolio_type": "onboarded",
  "risk_level": "low",
  "compliance_status": "compliant",
  "status": "active",
  "created_by": "user_123",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T16:20:00Z"
}
```

**Status Codes:**
- `200 OK`: Merchant updated successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Merchant not found
- `429 Too Many Requests`: Rate limit exceeded

#### Delete Merchant

**DELETE** `/api/v1/merchants/{id}`

Deletes a merchant from the portfolio.

**Path Parameters:**
- `id` (string, required): Merchant ID

**Response:**
```
204 No Content
```

**Status Codes:**
- `204 No Content`: Merchant deleted successfully
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Merchant not found
- `429 Too Many Requests`: Rate limit exceeded

### Merchant Search and Listing

#### List Merchants

**GET** `/api/v1/merchants`

Retrieves a paginated list of merchants with optional filtering.

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Number of items per page (default: 20, max: 100)
- `portfolio_type` (string, optional): Filter by portfolio type
- `risk_level` (string, optional): Filter by risk level
- `industry` (string, optional): Filter by industry
- `status` (string, optional): Filter by status
- `search` (string, optional): Search query for name, legal name, or industry

**Example Request:**
```http
GET /api/v1/merchants?page=1&page_size=20&portfolio_type=onboarded&risk_level=medium&search=technology
```

**Response:**
```json
{
  "merchants": [
    {
      "id": "merchant_1234567890",
      "name": "Acme Corporation",
      "legal_name": "Acme Corporation LLC",
      "industry": "Technology",
      "portfolio_type": "onboarded",
      "risk_level": "medium",
      "compliance_status": "compliant",
      "status": "active",
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T15:45:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "page_size": 20,
  "has_more": true
}
```

**Status Codes:**
- `200 OK`: Merchants retrieved successfully
- `400 Bad Request`: Invalid query parameters
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded

#### Advanced Search

**POST** `/api/v1/merchants/search`

Performs advanced search with complex filtering criteria.

**Request Body:**
```json
{
  "filters": {
    "portfolio_type": "onboarded",
    "risk_level": "high",
    "industry": "Technology",
    "status": "active",
    "search_query": "software development"
  },
  "pagination": {
    "page": 1,
    "page_size": 50
  },
  "sorting": {
    "field": "created_at",
    "order": "desc"
  }
}
```

**Response:**
```json
{
  "merchants": [
    {
      "id": "merchant_1234567890",
      "name": "Acme Corporation",
      "legal_name": "Acme Corporation LLC",
      "industry": "Technology",
      "portfolio_type": "onboarded",
      "risk_level": "high",
      "compliance_status": "compliant",
      "status": "active",
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T15:45:00Z"
    }
  ],
  "total": 25,
  "page": 1,
  "page_size": 50,
  "has_more": false,
  "filters_applied": {
    "portfolio_type": "onboarded",
    "risk_level": "high",
    "industry": "Technology",
    "status": "active",
    "search_query": "software development"
  }
}
```

**Status Codes:**
- `200 OK`: Search completed successfully
- `400 Bad Request`: Invalid search criteria
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded

### Bulk Operations

#### Bulk Update Portfolio Type

**POST** `/api/v1/merchants/bulk/portfolio-type`

Updates the portfolio type for multiple merchants.

**Request Body:**
```json
{
  "merchant_ids": [
    "merchant_1234567890",
    "merchant_0987654321",
    "merchant_1122334455"
  ],
  "portfolio_type": "onboarded"
}
```

**Response:**
```json
{
  "operation_id": "bulk_op_7890123456",
  "status": "completed",
  "total_merchants": 3,
  "successful_updates": 3,
  "failed_updates": 0,
  "errors": [],
  "started_at": "2025-01-15T16:30:00Z",
  "completed_at": "2025-01-15T16:30:05Z"
}
```

**Status Codes:**
- `200 OK`: Bulk operation completed successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded

#### Bulk Update Risk Level

**POST** `/api/v1/merchants/bulk/risk-level`

Updates the risk level for multiple merchants.

**Request Body:**
```json
{
  "merchant_ids": [
    "merchant_1234567890",
    "merchant_0987654321"
  ],
  "risk_level": "low"
}
```

**Response:**
```json
{
  "operation_id": "bulk_op_7890123457",
  "status": "completed",
  "total_merchants": 2,
  "successful_updates": 2,
  "failed_updates": 0,
  "errors": [],
  "started_at": "2025-01-15T16:35:00Z",
  "completed_at": "2025-01-15T16:35:03Z"
}
```

**Status Codes:**
- `200 OK`: Bulk operation completed successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded

#### Get Bulk Operation Status

**GET** `/api/v1/merchants/bulk/{operation_id}`

Retrieves the status of a bulk operation.

**Path Parameters:**
- `operation_id` (string, required): Bulk operation ID

**Response:**
```json
{
  "operation_id": "bulk_op_7890123456",
  "status": "in_progress",
  "total_merchants": 100,
  "processed_merchants": 45,
  "successful_updates": 43,
  "failed_updates": 2,
  "errors": [
    {
      "merchant_id": "merchant_9999999999",
      "error": "merchant not found"
    }
  ],
  "started_at": "2025-01-15T16:30:00Z",
  "estimated_completion": "2025-01-15T16:32:00Z"
}
```

**Status Codes:**
- `200 OK`: Operation status retrieved successfully
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Operation not found
- `429 Too Many Requests`: Rate limit exceeded

### Session Management

#### Start Merchant Session

**POST** `/api/v1/merchants/{id}/session`

Starts a session for a specific merchant (only one merchant can be active at a time).

**Path Parameters:**
- `id` (string, required): Merchant ID

**Response:**
```json
{
  "session_id": "session_1234567890",
  "user_id": "user_123",
  "merchant_id": "merchant_1234567890",
  "merchant_name": "Acme Corporation",
  "started_at": "2025-01-15T16:40:00Z",
  "expires_at": "2025-01-15T20:40:00Z"
}
```

**Status Codes:**
- `201 Created`: Session started successfully
- `400 Bad Request`: Invalid merchant ID
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Merchant not found
- `409 Conflict`: Another merchant session is already active
- `429 Too Many Requests`: Rate limit exceeded

#### End Merchant Session

**DELETE** `/api/v1/merchants/{id}/session`

Ends the current merchant session.

**Path Parameters:**
- `id` (string, required): Merchant ID

**Response:**
```
204 No Content
```

**Status Codes:**
- `204 No Content`: Session ended successfully
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Session not found
- `429 Too Many Requests`: Rate limit exceeded

#### Get Active Session

**GET** `/api/v1/merchants/session/active`

Retrieves the currently active merchant session.

**Response:**
```json
{
  "session_id": "session_1234567890",
  "user_id": "user_123",
  "merchant_id": "merchant_1234567890",
  "merchant_name": "Acme Corporation",
  "started_at": "2025-01-15T16:40:00Z",
  "expires_at": "2025-01-15T20:40:00Z"
}
```

**Status Codes:**
- `200 OK`: Active session retrieved successfully
- `401 Unauthorized`: Authentication required
- `404 Not Found`: No active session
- `429 Too Many Requests`: Rate limit exceeded

### Analytics and Reporting

#### Get Merchant Analytics

**GET** `/api/v1/merchants/analytics`

Retrieves comprehensive analytics and insights for the merchant portfolio.

**Query Parameters:**
- `period` (string, optional): Time period for analytics (7d, 30d, 90d, 1y)
- `portfolio_type` (string, optional): Filter by portfolio type
- `risk_level` (string, optional): Filter by risk level

**Response:**
```json
{
  "summary": {
    "total_merchants": 1250,
    "onboarded": 800,
    "deactivated": 150,
    "prospective": 200,
    "pending": 100,
    "high_risk": 300,
    "medium_risk": 600,
    "low_risk": 350
  },
  "trends": {
    "new_merchants_7d": 25,
    "new_merchants_30d": 120,
    "deactivated_7d": 5,
    "deactivated_30d": 20,
    "risk_changes_7d": 15,
    "risk_changes_30d": 80
  },
  "industry_distribution": [
    {
      "industry": "Technology",
      "count": 400,
      "percentage": 32.0
    },
    {
      "industry": "Financial Services",
      "count": 300,
      "percentage": 24.0
    },
    {
      "industry": "Healthcare",
      "count": 200,
      "percentage": 16.0
    }
  ],
  "risk_trends": [
    {
      "date": "2025-01-15",
      "high_risk": 300,
      "medium_risk": 600,
      "low_risk": 350
    }
  ],
  "compliance_status": {
    "compliant": 1000,
    "pending": 200,
    "non_compliant": 50
  }
}
```

**Status Codes:**
- `200 OK`: Analytics retrieved successfully
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded

#### Get Portfolio Types

**GET** `/api/v1/merchants/portfolio-types`

Retrieves available portfolio types with their descriptions.

**Response:**
```json
{
  "portfolio_types": [
    {
      "type": "onboarded",
      "name": "Onboarded",
      "description": "Active merchants in the portfolio",
      "count": 800
    },
    {
      "type": "deactivated",
      "name": "Deactivated",
      "description": "Merchants that have been deactivated",
      "count": 150
    },
    {
      "type": "prospective",
      "name": "Prospective",
      "description": "Potential merchants under consideration",
      "count": 200
    },
    {
      "type": "pending",
      "name": "Pending",
      "description": "Merchants awaiting approval",
      "count": 100
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Portfolio types retrieved successfully
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded

#### Get Risk Levels

**GET** `/api/v1/merchants/risk-levels`

Retrieves available risk levels with their descriptions.

**Response:**
```json
{
  "risk_levels": [
    {
      "level": "high",
      "name": "High Risk",
      "description": "High-risk merchants requiring additional monitoring",
      "count": 300,
      "color": "#dc3545"
    },
    {
      "level": "medium",
      "name": "Medium Risk",
      "description": "Medium-risk merchants with standard monitoring",
      "count": 600,
      "color": "#ffc107"
    },
    {
      "level": "low",
      "name": "Low Risk",
      "description": "Low-risk merchants with minimal monitoring",
      "count": 350,
      "color": "#28a745"
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Risk levels retrieved successfully
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded

#### Get Merchant Statistics

**GET** `/api/v1/merchants/statistics`

Retrieves detailed statistics for the merchant portfolio.

**Response:**
```json
{
  "portfolio_statistics": {
    "total_merchants": 1250,
    "active_merchants": 800,
    "inactive_merchants": 450,
    "average_employee_count": 45.5,
    "total_annual_revenue": 1250000000.00,
    "average_annual_revenue": 1000000.00
  },
  "geographic_distribution": [
    {
      "country": "United States",
      "count": 800,
      "percentage": 64.0
    },
    {
      "country": "Canada",
      "count": 200,
      "percentage": 16.0
    },
    {
      "country": "United Kingdom",
      "count": 150,
      "percentage": 12.0
    }
  ],
  "business_type_distribution": [
    {
      "business_type": "Corporation",
      "count": 600,
      "percentage": 48.0
    },
    {
      "business_type": "LLC",
      "count": 400,
      "percentage": 32.0
    },
    {
      "business_type": "Partnership",
      "count": 150,
      "percentage": 12.0
    }
  ],
  "compliance_metrics": {
    "compliance_rate": 80.0,
    "pending_reviews": 200,
    "overdue_reviews": 15,
    "average_review_time_days": 7.5
  }
}
```

**Status Codes:**
- `200 OK`: Statistics retrieved successfully
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded

## Error Handling

### Error Response Format

All error responses follow a consistent format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error details",
    "request_id": "req_1234567890",
    "timestamp": "2025-01-15T16:45:00Z"
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | Invalid request data or parameters |
| `UNAUTHORIZED` | 401 | Authentication required or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict (e.g., duplicate merchant) |
| `VALIDATION_ERROR` | 422 | Data validation failed |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit exceeded |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

### Validation Errors

When validation fails, the error response includes field-specific details:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": "One or more fields contain invalid data",
    "validation_errors": [
      {
        "field": "name",
        "message": "merchant name is required and cannot be empty"
      },
      {
        "field": "portfolio_type",
        "message": "invalid portfolio type"
      }
    ],
    "request_id": "req_1234567890",
    "timestamp": "2025-01-15T16:45:00Z"
  }
}
```

## SDK Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

class MerchantPortfolioAPI {
  constructor(baseURL, apiKey) {
    this.client = axios.create({
      baseURL: `${baseURL}/api/v1/merchants`,
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async createMerchant(merchantData) {
    try {
      const response = await this.client.post('/', merchantData);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to create merchant: ${error.response?.data?.error?.message || error.message}`);
    }
  }

  async getMerchant(merchantId) {
    try {
      const response = await this.client.get(`/${merchantId}`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get merchant: ${error.response?.data?.error?.message || error.message}`);
    }
  }

  async searchMerchants(filters = {}) {
    try {
      const response = await this.client.get('/', { params: filters });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to search merchants: ${error.response?.data?.error?.message || error.message}`);
    }
  }

  async bulkUpdatePortfolioType(merchantIds, portfolioType) {
    try {
      const response = await this.client.post('/bulk/portfolio-type', {
        merchant_ids: merchantIds,
        portfolio_type: portfolioType
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to bulk update portfolio type: ${error.response?.data?.error?.message || error.message}`);
    }
  }
}

// Usage example
const api = new MerchantPortfolioAPI('https://api.kyb-platform.com', 'your-api-key');

// Create a new merchant
const newMerchant = await api.createMerchant({
  name: 'Example Corp',
  legal_name: 'Example Corporation LLC',
  industry: 'Technology',
  portfolio_type: 'prospective',
  risk_level: 'medium',
  address: {
    street1: '123 Example St',
    city: 'San Francisco',
    state: 'CA',
    postal_code: '94105',
    country: 'United States',
    country_code: 'US'
  },
  contact_info: {
    email: 'contact@example.com',
    phone: '+1-555-123-4567'
  }
});

console.log('Created merchant:', newMerchant);
```

### Python

```python
import requests
from typing import Dict, List, Optional

class MerchantPortfolioAPI:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = f"{base_url}/api/v1/merchants"
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }

    def create_merchant(self, merchant_data: Dict) -> Dict:
        """Create a new merchant"""
        try:
            response = requests.post(
                self.base_url,
                json=merchant_data,
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            error_msg = e.response.json().get('error', {}).get('message', str(e)) if e.response else str(e)
            raise Exception(f"Failed to create merchant: {error_msg}")

    def get_merchant(self, merchant_id: str) -> Dict:
        """Get a merchant by ID"""
        try:
            response = requests.get(
                f"{self.base_url}/{merchant_id}",
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            error_msg = e.response.json().get('error', {}).get('message', str(e)) if e.response else str(e)
            raise Exception(f"Failed to get merchant: {error_msg}")

    def search_merchants(self, filters: Optional[Dict] = None) -> Dict:
        """Search merchants with optional filters"""
        try:
            response = requests.get(
                self.base_url,
                params=filters or {},
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            error_msg = e.response.json().get('error', {}).get('message', str(e)) if e.response else str(e)
            raise Exception(f"Failed to search merchants: {error_msg}")

    def bulk_update_portfolio_type(self, merchant_ids: List[str], portfolio_type: str) -> Dict:
        """Bulk update portfolio type for multiple merchants"""
        try:
            response = requests.post(
                f"{self.base_url}/bulk/portfolio-type",
                json={
                    'merchant_ids': merchant_ids,
                    'portfolio_type': portfolio_type
                },
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            error_msg = e.response.json().get('error', {}).get('message', str(e)) if e.response else str(e)
            raise Exception(f"Failed to bulk update portfolio type: {error_msg}")

# Usage example
api = MerchantPortfolioAPI('https://api.kyb-platform.com', 'your-api-key')

# Create a new merchant
new_merchant = api.create_merchant({
    'name': 'Example Corp',
    'legal_name': 'Example Corporation LLC',
    'industry': 'Technology',
    'portfolio_type': 'prospective',
    'risk_level': 'medium',
    'address': {
        'street1': '123 Example St',
        'city': 'San Francisco',
        'state': 'CA',
        'postal_code': '94105',
        'country': 'United States',
        'country_code': 'US'
    },
    'contact_info': {
        'email': 'contact@example.com',
        'phone': '+1-555-123-4567'
    }
})

print(f"Created merchant: {new_merchant}")
```

## Integration Guides

### Webhook Integration

The API supports webhooks for real-time notifications of merchant events.

#### Webhook Events

- `merchant.created`: New merchant created
- `merchant.updated`: Merchant information updated
- `merchant.deleted`: Merchant deleted
- `merchant.portfolio_type_changed`: Portfolio type changed
- `merchant.risk_level_changed`: Risk level changed
- `bulk_operation.completed`: Bulk operation completed

#### Webhook Payload

```json
{
  "event": "merchant.updated",
  "data": {
    "merchant_id": "merchant_1234567890",
    "changes": {
      "portfolio_type": {
        "old": "prospective",
        "new": "onboarded"
      }
    }
  },
  "timestamp": "2025-01-15T16:45:00Z",
  "webhook_id": "webhook_1234567890"
}
```

### Rate Limiting Best Practices

1. **Implement exponential backoff** for rate limit errors
2. **Cache responses** when appropriate to reduce API calls
3. **Use bulk operations** instead of individual requests when possible
4. **Monitor rate limit headers** to avoid hitting limits

### Pagination Best Practices

1. **Use appropriate page sizes** (20-50 items for most use cases)
2. **Implement infinite scroll** or "load more" functionality
3. **Cache paginated results** to improve user experience
4. **Handle edge cases** like empty results or large datasets

### Error Handling Best Practices

1. **Implement retry logic** for transient errors (5xx status codes)
2. **Provide user-friendly error messages** based on error codes
3. **Log errors** for debugging and monitoring
4. **Handle network timeouts** gracefully

## Changelog

### Version 1.0.0 (January 2025)

**Initial Release:**
- Merchant CRUD operations
- Search and filtering capabilities
- Bulk operations for portfolio type and risk level updates
- Session management (single merchant active at a time)
- Analytics and reporting endpoints
- Comprehensive error handling
- Rate limiting and authentication
- SDK examples for JavaScript and Python

## Support

For API support and questions:
- **Documentation**: [https://docs.kyb-platform.com/api](https://docs.kyb-platform.com/api)
- **Support Email**: api-support@kyb-platform.com
- **Status Page**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
