# Enhanced Classification API Documentation

## Overview

The Enhanced Classification API provides advanced business classification capabilities with Machine Learning integration, crosswalk mappings, geographic awareness, and industry-specific algorithms. This API maintains backward compatibility with existing endpoints while offering new enhanced features.

## API Versions

### Supported Versions
- **v1**: Legacy API with basic classification features (deprecated on 2024-12-31)
- **v2**: Enhanced API with ML integration, crosswalk mappings, geographic awareness, and industry-specific algorithms

### Version Selection
API versions can be specified using:
- **Header**: `X-API-Version: v2`
- **Query Parameter**: `?version=v2`
- **Accept Header**: `Accept: application/vnd.api+json; version="v2"`

## Base URL
```
https://api.business-verification.com/v2
```

## Authentication
All API requests require authentication using API keys:
```
Authorization: Bearer YOUR_API_KEY
```

## Enhanced Classification Endpoints

### 1. Enhanced Classification

**Endpoint**: `POST /api/v2/classification/enhanced`

**Description**: Perform enhanced classification with ML integration, crosswalk mappings, geographic awareness, and industry-specific algorithms.

**Request Headers**:
```
Content-Type: application/json
Authorization: Bearer YOUR_API_KEY
X-API-Version: v2
```

**Request Body**:
```json
{
  "business_name": "Acme Corporation",
  "business_type": "Corporation",
  "industry": "Technology",
  "description": "Software development and consulting services",
  "keywords": "software, development, consulting, technology",
  "registration_number": "123456789",
  "ml_model_version": "bert_classifier_v1.0.0",
  "geographic_region": "California, USA",
  "industry_type": "technology",
  "crosswalk_mappings": {
    "enable_naics_to_sic": true,
    "enable_naics_to_mcc": true
  },
  "enhanced_metadata": {
    "source": "api_request",
    "priority": "high"
  }
}
```

**Response**:
```json
{
  "success": true,
  "business_id": "biz_123456789",
  "classifications": [
    {
      "industry_code": "541511",
      "industry_name": "Custom Computer Programming Services",
      "confidence_score": 0.95,
      "classification_method": "ml_enhanced"
    }
  ],
  "primary_industry": {
    "industry_code": "541511",
    "industry_name": "Custom Computer Programming Services",
    "confidence_score": 0.95,
    "classification_method": "ml_enhanced"
  },
  "overall_confidence": 0.95,
  "validation_score": 0.92,
  "classification_method": "ml_enhanced",
  "processing_time": "150ms",
  "timestamp": "2024-01-01T12:00:00Z",
  "ml_model_version": "bert_classifier_v1.0.0",
  "ml_confidence_score": 0.95,
  "crosswalk_mappings": {
    "naics_to_sic": {
      "7371": 0.95,
      "7372": 0.85
    },
    "naics_to_mcc": {
      "5734": 0.90
    },
    "confidence": 0.92
  },
  "geographic_region": "California, USA",
  "region_confidence_score": 0.98,
  "industry_specific_data": {
    "industry_type": "technology",
    "classification_algorithm": "hybrid",
    "confidence_score": 0.95,
    "validation_rules_passed": ["keyword_match", "business_type_validation"]
  },
  "classification_algorithm": "ml_enhanced",
  "validation_rules_applied": ["ml_validation", "crosswalk_validation", "geographic_validation"],
  "enhanced_metadata": {
    "source": "api_request",
    "priority": "high"
  },
  "api_version": "v2"
}
```

### 2. Batch Enhanced Classification

**Endpoint**: `POST /api/v2/classification/enhanced/batch`

**Description**: Perform enhanced classification for multiple businesses in a single request.

**Request Body**:
```json
{
  "businesses": [
    {
      "business_name": "Acme Corporation",
      "description": "Software development services",
      "ml_model_version": "bert_classifier_v1.0.0",
      "geographic_region": "California, USA"
    },
    {
      "business_name": "Tech Solutions Inc",
      "description": "IT consulting services",
      "ml_model_version": "bert_classifier_v1.0.0",
      "geographic_region": "New York, USA"
    }
  ],
  "api_version": "v2"
}
```

**Response**:
```json
{
  "success": true,
  "results": [
    {
      "success": true,
      "business_id": "biz_123456789",
      "primary_industry": {
        "industry_code": "541511",
        "industry_name": "Custom Computer Programming Services",
        "confidence_score": 0.95
      },
      "overall_confidence": 0.95,
      "ml_confidence_score": 0.95,
      "geographic_region": "California, USA",
      "api_version": "v2"
    },
    {
      "success": true,
      "business_id": "biz_987654321",
      "primary_industry": {
        "industry_code": "541618",
        "industry_name": "Other Management Consulting Services",
        "confidence_score": 0.88
      },
      "overall_confidence": 0.88,
      "ml_confidence_score": 0.88,
      "geographic_region": "New York, USA",
      "api_version": "v2"
    }
  ],
  "total_processed": 2,
  "success_count": 2,
  "error_count": 0,
  "processing_time": "300ms",
  "timestamp": "2024-01-01T12:00:00Z",
  "api_version": "v2"
}
```

### 3. Backward Compatible Classification

**Endpoint**: `POST /api/v1/classification`

**Description**: Legacy endpoint for backward compatibility. Uses v1 API format but internally uses enhanced classification.

**Request Body**:
```json
{
  "business_name": "Acme Corporation",
  "business_type": "Corporation",
  "industry": "Technology",
  "description": "Software development services",
  "keywords": "software, development",
  "registration_number": "123456789"
}
```

**Response**:
```json
{
  "success": true,
  "business_id": "biz_123456789",
  "classifications": [
    {
      "industry_code": "541511",
      "industry_name": "Custom Computer Programming Services",
      "confidence_score": 0.95,
      "classification_method": "enhanced"
    }
  ],
  "primary_industry": {
    "industry_code": "541511",
    "industry_name": "Custom Computer Programming Services",
    "confidence_score": 0.95,
    "classification_method": "enhanced"
  },
  "overall_confidence": 0.95,
  "validation_score": 0.92,
  "classification_method": "enhanced",
  "processing_time": "150ms",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 4. API Version Information

**Endpoint**: `GET /api/versions`

**Description**: Get information about available API versions and deprecation schedules.

**Response**:
```json
{
  "current_version": "v2",
  "supported_versions": ["v1", "v2"],
  "version_details": {
    "v1": "Legacy API with basic classification features",
    "v2": "Enhanced API with ML integration, crosswalk mappings, geographic awareness, and industry-specific algorithms"
  },
  "deprecation_info": {
    "v1": "v1 will be deprecated on 2024-12-31. Please migrate to v2 for enhanced features."
  }
}
```

## Request Fields

### Basic Fields (v1 compatible)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `business_name` | string | Yes | Name of the business |
| `business_type` | string | No | Type of business entity |
| `industry` | string | No | Industry category |
| `description` | string | No | Business description |
| `keywords` | string | No | Keywords related to the business |
| `registration_number` | string | No | Business registration number |

### Enhanced Fields (v2 only)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ml_model_version` | string | No | ML model version to use for classification |
| `geographic_region` | string | No | Geographic region for enhanced accuracy |
| `industry_type` | string | No | Industry type for specialized algorithms |
| `crosswalk_mappings` | object | No | Crosswalk mapping configuration |
| `enhanced_metadata` | object | No | Additional metadata for processing |
| `api_version` | string | No | API version (auto-detected if not provided) |

## Response Fields

### Basic Fields (v1 compatible)
| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Whether the classification was successful |
| `business_id` | string | Unique business identifier |
| `classifications` | array | Array of industry classifications |
| `primary_industry` | object | Primary industry classification |
| `overall_confidence` | float | Overall confidence score (0.0-1.0) |
| `validation_score` | float | Validation score (0.0-1.0) |
| `classification_method` | string | Method used for classification |
| `processing_time` | string | Processing time duration |
| `timestamp` | string | Request timestamp |

### Enhanced Fields (v2 only)
| Field | Type | Description |
|-------|------|-------------|
| `ml_model_version` | string | ML model version used |
| `ml_confidence_score` | float | ML-specific confidence score |
| `crosswalk_mappings` | object | Crosswalk mappings between code systems |
| `geographic_region` | string | Detected geographic region |
| `region_confidence_score` | float | Geographic region confidence |
| `industry_specific_data` | object | Industry-specific classification data |
| `classification_algorithm` | string | Algorithm used for classification |
| `validation_rules_applied` | array | Validation rules that were applied |
| `enhanced_metadata` | object | Enhanced metadata from request |
| `api_version` | string | API version used |

## Error Responses

### Standard Error Format
```json
{
  "success": false,
  "error": "Error message description",
  "timestamp": "2024-01-01T12:00:00Z",
  "api_version": "v2"
}
```

### Common Error Codes
| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | Validation failed | Request validation failed |
| 401 | Unauthorized | Invalid or missing API key |
| 429 | Rate limit exceeded | Too many requests |
| 500 | Internal server error | Server processing error |

## Rate Limits

- **Standard**: 1000 requests per hour
- **Enhanced**: 500 requests per hour (due to ML processing)
- **Batch**: 100 requests per hour (max 100 businesses per request)

## Migration Guide

### From v1 to v2

#### 1. Update API Version
```bash
# Old (v1)
curl -H "X-API-Version: v1" https://api.business-verification.com/v1/classification

# New (v2)
curl -H "X-API-Version: v2" https://api.business-verification.com/v2/classification/enhanced
```

#### 2. Update Request Format
```json
// Old v1 format
{
  "business_name": "Acme Corporation",
  "description": "Software development"
}

// New v2 format with enhanced features
{
  "business_name": "Acme Corporation",
  "description": "Software development",
  "ml_model_version": "bert_classifier_v1.0.0",
  "geographic_region": "California, USA",
  "industry_type": "technology"
}
```

#### 3. Handle New Response Fields
```javascript
// Old v1 response handling
const result = response.primary_industry;

// New v2 response handling
const result = {
  primary: response.primary_industry,
  mlConfidence: response.ml_confidence_score,
  crosswalkMappings: response.crosswalk_mappings,
  geographicRegion: response.geographic_region
};
```

#### 4. Update Error Handling
```javascript
// Old v1 error handling
if (!response.success) {
  console.error(response.error);
}

// New v2 error handling
if (!response.success) {
  console.error(`Error (${response.api_version}): ${response.error}`);
}
```

## Code Examples

### JavaScript/Node.js
```javascript
const axios = require('axios');

async function classifyBusiness(businessName, description) {
  try {
    const response = await axios.post('https://api.business-verification.com/v2/classification/enhanced', {
      business_name: businessName,
      description: description,
      ml_model_version: 'bert_classifier_v1.0.0',
      geographic_region: 'California, USA',
      industry_type: 'technology'
    }, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'X-API-Version': 'v2',
        'Content-Type': 'application/json'
      }
    });

    return response.data;
  } catch (error) {
    console.error('Classification error:', error.response.data);
    throw error;
  }
}
```

### Python
```python
import requests

def classify_business(business_name, description):
    url = 'https://api.business-verification.com/v2/classification/enhanced'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'X-API-Version': 'v2',
        'Content-Type': 'application/json'
    }
    data = {
        'business_name': business_name,
        'description': description,
        'ml_model_version': 'bert_classifier_v1.0.0',
        'geographic_region': 'California, USA',
        'industry_type': 'technology'
    }
    
    response = requests.post(url, json=data, headers=headers)
    response.raise_for_status()
    return response.json()
```

### cURL
```bash
curl -X POST https://api.business-verification.com/v2/classification/enhanced \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-API-Version: v2" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "description": "Software development services",
    "ml_model_version": "bert_classifier_v1.0.0",
    "geographic_region": "California, USA",
    "industry_type": "technology"
  }'
```

## Best Practices

### 1. API Version Management
- Always specify the API version explicitly
- Monitor deprecation notices
- Plan migration to newer versions

### 2. Error Handling
- Implement proper error handling for all response codes
- Log errors with context for debugging
- Implement retry logic for transient errors

### 3. Performance Optimization
- Use batch endpoints for multiple classifications
- Cache results when appropriate
- Monitor rate limits

### 4. Data Quality
- Provide detailed business descriptions
- Include relevant keywords
- Specify geographic regions when known

### 5. Security
- Keep API keys secure
- Use HTTPS for all requests
- Rotate API keys regularly

## Support

For API support and questions:
- **Email**: api-support@business-verification.com
- **Documentation**: https://docs.business-verification.com
- **Status Page**: https://status.business-verification.com
