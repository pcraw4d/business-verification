# Enhanced Business Intelligence System - API Reference

## Overview

This document provides comprehensive API reference documentation for the Enhanced Business Intelligence System. It covers all endpoints, request/response formats, authentication, error handling, and usage examples.

## Table of Contents

1. [Authentication](#authentication)
2. [Base URL and Versioning](#base-url-and-versioning)
3. [Common Response Formats](#common-response-formats)
4. [Error Handling](#error-handling)
5. [Classification Endpoints](#classification-endpoints)
6. [Enhanced Classification Endpoints](#enhanced-classification-endpoints)
7. [Risk Assessment Endpoints](#risk-assessment-endpoints)
8. [Data Discovery Endpoints](#data-discovery-endpoints)
9. [Caching Endpoints](#caching-endpoints)
10. [Monitoring Endpoints](#monitoring-endpoints)
11. [Health and Status Endpoints](#health-and-status-endpoints)
12. [Data Models](#data-models)

## Authentication

### API Key Authentication

The API supports API key authentication for secure access control.

**Header**: `Authorization: Bearer YOUR_API_KEY`

**Example**:
```bash
curl -H "Authorization: Bearer sk_live_1234567890abcdef" \
     -H "Content-Type: application/json" \
     -X POST https://api.kyb-platform.com/v1/classify \
     -d '{"business_name": "Acme Corporation"}'
```

### JWT Token Authentication

For advanced use cases, JWT token authentication is supported.

**Header**: `Authorization: Bearer JWT_TOKEN`

**Token Format**:
```json
{
  "sub": "user_1234567890",
  "iss": "kyb-platform",
  "aud": "kyb-api",
  "iat": 1640995200,
  "exp": 1641081600,
  "permissions": ["classify", "risk_assess", "data_discover"]
}
```

## Base URL and Versioning

### Base URL
- **Production**: `https://api.kyb-platform.com`
- **Staging**: `https://staging-api.kyb-platform.com`
- **Development**: `http://localhost:8080`

### API Versioning
The API uses URL-based versioning:
- **Current Version**: `/v1`
- **Future Versions**: `/v2`, `/v3`, etc.

### Rate Limiting
- **Standard Plan**: 1,000 requests per hour
- **Professional Plan**: 10,000 requests per hour
- **Enterprise Plan**: 100,000 requests per hour

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Common Response Formats

### Success Response
All successful API responses follow this format:
```json
{
  "success": true,
  "data": {
    // Response data here
  },
  "metadata": {
    "request_id": "req_1234567890",
    "timestamp": "2024-12-19T10:30:00Z",
    "processing_time": "1.2s",
    "cache_hit": false
  }
}
```

### Error Response
All error responses follow this format:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid business name provided",
    "details": {
      "field": "business_name",
      "reason": "Business name cannot be empty"
    },
    "request_id": "req_1234567890",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

## Error Handling

### HTTP Status Codes
- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service temporarily unavailable

### Error Codes
- `VALIDATION_ERROR`: Input validation failed
- `CLASSIFICATION_ERROR`: Classification processing failed
- `RISK_ASSESSMENT_ERROR`: Risk assessment processing failed
- `DATA_DISCOVERY_ERROR`: Data discovery processing failed
- `RATE_LIMIT_EXCEEDED`: Rate limit exceeded
- `AUTHENTICATION_ERROR`: Authentication failed
- `AUTHORIZATION_ERROR`: Authorization failed
- `INTERNAL_ERROR`: Internal server error
- `SERVICE_UNAVAILABLE`: Service temporarily unavailable

## Classification Endpoints

### POST /v1/classify

Classifies a business using industry codes (NAICS, SIC, MCC).

**Request**:
```json
{
  "business_name": "Acme Corporation",
  "description": "Technology consulting services",
  "website": "https://acme.com",
  "industry": "Technology",
  "keywords": ["consulting", "technology", "services"],
  "options": {
    "include_alternatives": true,
    "max_results": 3,
    "confidence_threshold": 0.7,
    "strategies": ["keyword", "ml", "similarity"]
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "class_1234567890",
    "business_name": "Acme Corporation",
    "classification": {
      "primary_code": {
        "type": "NAICS",
        "code": "541511",
        "description": "Custom Computer Programming Services",
        "confidence": 0.95,
        "reasoning": "Strong keyword matches: 'consulting', 'technology', 'services'"
      },
      "alternatives": [
        {
          "type": "SIC",
          "code": "7371",
          "description": "Computer Programming Services",
          "confidence": 0.92,
          "reasoning": "High similarity to primary classification"
        },
        {
          "type": "MCC",
          "code": "7392",
          "description": "Management Consulting Services",
          "confidence": 0.88,
          "reasoning": "Keyword match: 'consulting'"
        }
      ]
    },
    "strategies_used": ["keyword", "ml", "similarity"],
    "processing_time": "1.2s"
  },
  "metadata": {
    "request_id": "req_1234567890",
    "timestamp": "2024-12-19T10:30:00Z",
    "processing_time": "1.2s",
    "cache_hit": false
  }
}
```

**cURL Example**:
```bash
curl -X POST https://api.kyb-platform.com/v1/classify \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "description": "Technology consulting services",
    "website": "https://acme.com"
  }'
```

### GET /v1/classify/{classification_id}

Retrieves a specific classification result.

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "class_1234567890",
    "business_name": "Acme Corporation",
    "classification": {
      "primary_code": {
        "type": "NAICS",
        "code": "541511",
        "description": "Custom Computer Programming Services",
        "confidence": 0.95
      },
      "alternatives": [...]
    },
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### GET /v1/classify/history/{business_id}

Retrieves classification history for a business.

**Query Parameters**:
- `limit` (optional): Number of results to return (default: 10, max: 100)
- `offset` (optional): Number of results to skip (default: 0)
- `start_date` (optional): Start date for filtering (ISO 8601 format)
- `end_date` (optional): End date for filtering (ISO 8601 format)

**Response**:
```json
{
  "success": true,
  "data": {
    "business_id": "business_1234567890",
    "classifications": [
      {
        "id": "class_1234567890",
        "primary_code": {
          "type": "NAICS",
          "code": "541511",
          "description": "Custom Computer Programming Services",
          "confidence": 0.95
        },
        "created_at": "2024-12-19T10:30:00Z"
      }
    ],
    "total_count": 1,
    "has_more": false
  }
}
```

## Enhanced Classification Endpoints

### POST /v2/classify

Performs enhanced business classification using advanced ML models and comprehensive analysis.

**Request**:
```json
{
  "business_name": "Acme Corporation",
  "description": "AI-powered software development company specializing in machine learning solutions",
  "website": "https://acme.com",
  "industry": "Technology",
  "keywords": ["AI", "machine learning", "software development", "consulting"],
  "options": {
    "include_alternatives": true,
    "max_results": 5,
    "confidence_threshold": 0.8,
    "strategies": ["ml_bert", "ml_distilbert", "custom_neural_net", "keyword", "similarity"],
    "include_risk_assessment": true,
    "include_ml_explainability": true,
    "include_confidence_breakdown": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "class_enhanced_1234567890",
    "business_name": "Acme Corporation",
    "classification": {
      "primary_code": {
        "type": "NAICS",
        "code": "541511",
        "description": "Custom Computer Programming Services",
        "confidence": 0.97,
        "reasoning": "BERT model analysis: Strong semantic match for software development with AI/ML specialization",
        "ml_model_used": "bert-base-uncased-v2.1.0",
        "confidence_breakdown": {
          "ml_bert": 0.95,
          "keyword_matching": 0.92,
          "similarity_analysis": 0.89
        }
      },
      "alternatives": [
        {
          "type": "NAICS",
          "code": "541512",
          "description": "Computer Systems Design Services",
          "confidence": 0.94,
          "reasoning": "DistilBERT analysis: High confidence for systems design with AI components",
          "ml_model_used": "distilbert-base-uncased-v1.5.0"
        },
        {
          "type": "SIC",
          "code": "7371",
          "description": "Computer Programming Services",
          "confidence": 0.91,
          "reasoning": "Custom neural network: Strong match for programming services"
        },
        {
          "type": "MCC",
          "code": "7372",
          "description": "Computer Programming, Data Processing and Integrated Systems Design Services",
          "confidence": 0.88,
          "reasoning": "Keyword and similarity analysis"
        }
      ]
    },
    "risk_assessment": {
      "overall_risk_level": "LOW",
      "risk_score": 0.15,
      "risk_factors": [
        {
          "factor": "Industry Risk",
          "level": "LOW",
          "score": 0.10,
          "explanation": "Technology industry with established business model"
        }
      ]
    },
    "ml_explainability": {
      "feature_importance": {
        "business_name": 0.35,
        "description": 0.45,
        "keywords": 0.20
      },
      "model_confidence": 0.97,
      "prediction_uncertainty": 0.03
    },
    "strategies_used": ["ml_bert", "ml_distilbert", "custom_neural_net", "keyword", "similarity"],
    "processing_time": "2.1s",
    "ml_processing_time": "1.8s"
  },
  "metadata": {
    "request_id": "req_enhanced_1234567890",
    "timestamp": "2024-12-19T10:30:00Z",
    "version": "v2.0",
    "ml_models_used": ["bert-base-uncased-v2.1.0", "distilbert-base-uncased-v1.5.0", "custom-neural-net-v1.0"],
    "cache_hit": false
  }
}
```

### POST /v2/classify/batch

Performs enhanced batch classification for multiple businesses using ML models.

**Request**:
```json
{
  "businesses": [
    {
      "business_name": "Acme Corporation",
      "description": "AI-powered software development",
      "website": "https://acme.com"
    },
    {
      "business_name": "Tech Solutions Inc",
      "description": "Cloud computing and data analytics",
      "website": "https://techsolutions.com"
    }
  ],
  "options": {
    "include_alternatives": true,
    "max_results": 3,
    "confidence_threshold": 0.8,
    "strategies": ["ml_bert", "keyword", "similarity"],
    "include_risk_assessment": true,
    "parallel_processing": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "batch_id": "batch_enhanced_1234567890",
    "total_businesses": 2,
    "successful_classifications": 2,
    "failed_classifications": 0,
    "classifications": [
      {
        "business_name": "Acme Corporation",
        "classification": {
          "primary_code": {
            "type": "NAICS",
            "code": "541511",
            "description": "Custom Computer Programming Services",
            "confidence": 0.97
          }
        },
        "risk_assessment": {
          "overall_risk_level": "LOW",
          "risk_score": 0.15
        }
      },
      {
        "business_name": "Tech Solutions Inc",
        "classification": {
          "primary_code": {
            "type": "NAICS",
            "code": "541512",
            "description": "Computer Systems Design Services",
            "confidence": 0.94
          }
        },
        "risk_assessment": {
          "overall_risk_level": "LOW",
          "risk_score": 0.18
        }
      }
    ],
    "processing_time": "3.2s",
    "average_processing_time": "1.6s"
  }
}
```

## Risk Assessment Endpoints

### POST /v1/risk/assess

Assesses business risk factors and provides comprehensive risk scoring.

### POST /v1/risk/enhanced/assess

Performs enhanced risk assessment using advanced ML models and comprehensive risk factor analysis.

**Request**:
```json
{
  "business_name": "Acme Corporation",
  "website": "https://acme.com",
  "industry": "Technology",
  "business_description": "Software development company specializing in AI solutions",
  "options": {
    "include_security_analysis": true,
    "include_financial_analysis": true,
    "include_compliance_analysis": true,
    "include_reputation_analysis": true,
    "include_ml_risk_detection": true,
    "include_keyword_analysis": true,
    "include_trend_analysis": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "assessment_id": "risk_enhanced_1234567890",
    "business_id": "business_1234567890",
    "timestamp": "2024-12-19T10:30:00Z",
    "overall_risk_score": 0.25,
    "overall_risk_level": "LOW",
    "risk_factors": [
      {
        "factor_id": "security_001",
        "factor_name": "Website Security",
        "category": "cybersecurity",
        "score": 0.15,
        "level": "LOW",
        "confidence": 0.95,
        "explanation": "Website has valid SSL certificate and security headers",
        "evidence": ["Valid SSL certificate", "Security headers present"],
        "calculated_at": "2024-12-19T10:30:00Z"
      }
    ],
    "recommendations": [
      {
        "id": "rec_001",
        "risk_factor": "security_001",
        "title": "Implement Additional Security Measures",
        "description": "Consider implementing additional security headers and monitoring",
        "priority": "LOW",
        "action": "Add Content Security Policy headers",
        "impact": "Improved security posture",
        "timeline": "1-2 weeks"
      }
    ],
    "trend_data": {
      "risk_trend": "stable",
      "change_percentage": 0.05,
      "change_period": "30_days"
    },
    "correlation_data": {
      "industry_average": 0.30,
      "peer_comparison": "below_average"
    },
    "alerts": [],
    "confidence_score": 0.92,
    "processing_time_ms": 1250,
    "metadata": {
      "ml_model_version": "v2.1.0",
      "assessment_method": "enhanced_ml_analysis"
    }
  }
}
```

### POST /v1/risk/factors/calculate

Calculates specific risk factors for a business.

**Request**:
```json
{
  "business_id": "business_1234567890",
  "factors": ["security", "financial", "compliance", "reputational"],
  "options": {
    "include_ml_analysis": true,
    "include_historical_data": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "business_id": "business_1234567890",
    "calculated_factors": [
      {
        "factor_id": "security_001",
        "factor_name": "Website Security",
        "category": "cybersecurity",
        "score": 0.15,
        "level": "LOW",
        "confidence": 0.95,
        "explanation": "Website security analysis completed",
        "evidence": ["Valid SSL certificate", "Security headers present"],
        "calculated_at": "2024-12-19T10:30:00Z"
      }
    ],
    "processing_time_ms": 850
  }
}
```

### POST /v1/risk/recommendations

Generates risk mitigation recommendations based on assessment results.

**Request**:
```json
{
  "business_id": "business_1234567890",
  "risk_factors": ["security_001", "financial_001"],
  "options": {
    "include_priority_ranking": true,
    "include_implementation_guidance": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "business_id": "business_1234567890",
    "recommendations": [
      {
        "id": "rec_001",
        "risk_factor": "security_001",
        "title": "Implement Additional Security Measures",
        "description": "Consider implementing additional security headers and monitoring",
        "priority": "LOW",
        "action": "Add Content Security Policy headers",
        "impact": "Improved security posture",
        "timeline": "1-2 weeks",
        "created_at": "2024-12-19T10:30:00Z"
      }
    ],
    "total_recommendations": 1
  }
}
```

### POST /v1/risk/trends/analyze

Analyzes risk trends for a business over time.

**Request**:
```json
{
  "business_id": "business_1234567890",
  "time_period": "90_days",
  "categories": ["security", "financial", "compliance"],
  "options": {
    "include_predictions": true,
    "include_peer_comparison": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "business_id": "business_1234567890",
    "trend_analysis": {
      "overall_trend": "stable",
      "change_percentage": 0.05,
      "change_period": "90_days",
      "category_trends": [
        {
          "category": "security",
          "trend": "improving",
          "change_percentage": -0.10,
          "current_score": 0.15,
          "previous_score": 0.25
        }
      ]
    },
    "predictions": [
      {
        "factor_id": "security_001",
        "predicted_score": 0.12,
        "predicted_level": "LOW",
        "confidence": 0.85,
        "horizon": "3_months",
        "predicted_at": "2024-12-19T10:30:00Z"
      }
    ],
    "peer_comparison": {
      "industry_average": 0.30,
      "peer_percentile": 25,
      "comparison_status": "below_average"
    }
  }
}
```

### GET /v1/risk/alerts

Retrieves active risk alerts for a business.

**Query Parameters**:
- `business_id` (optional): Filter by business ID
- `level` (optional): Filter by alert level (low, medium, high, critical)
- `category` (optional): Filter by risk category
- `limit` (optional): Number of results to return (default: 20, max: 100)
- `offset` (optional): Number of results to skip (default: 0)

**Response**:
```json
{
  "success": true,
  "data": {
    "alerts": [
      {
        "id": "alert_001",
        "business_id": "business_1234567890",
        "risk_factor": "security_001",
        "level": "MEDIUM",
        "message": "Security risk score has increased above threshold",
        "score": 0.65,
        "threshold": 0.60,
        "triggered_at": "2024-12-19T10:30:00Z",
        "acknowledged": false,
        "acknowledged_at": null
      }
    ],
    "total_count": 1,
    "has_more": false
  }
}
```

### POST /v1/risk/alerts/{alert_id}/acknowledge

Acknowledges a risk alert.

**Response**:
```json
{
  "success": true,
  "data": {
    "alert_id": "alert_001",
    "acknowledged": true,
    "acknowledged_at": "2024-12-19T10:35:00Z"
  }
}
```

### POST /v1/risk/alerts/{alert_id}/resolve

Resolves a risk alert.

**Response**:
```json
{
  "success": true,
  "data": {
    "alert_id": "alert_001",
    "resolved": true,
    "resolved_at": "2024-12-19T10:35:00Z"
  }
}
```

### GET /v1/risk/factors/{factor_id}/history

Retrieves historical data for a specific risk factor.

**Query Parameters**:
- `start_date` (optional): Start date for filtering (ISO 8601 format)
- `end_date` (optional): End date for filtering (ISO 8601 format)
- `limit` (optional): Number of results to return (default: 50, max: 200)

**Response**:
```json
{
  "success": true,
  "data": {
    "factor_id": "security_001",
    "history": [
      {
        "score": 0.15,
        "level": "LOW",
        "recorded_at": "2024-12-19T10:30:00Z",
        "change_from": 0.05,
        "change_period": "7_days"
      }
    ],
    "total_count": 1,
    "has_more": false
  }
}
```

**Request**:
```json
{
  "business_name": "Acme Corporation",
  "website": "https://acme.com",
  "industry": "Technology",
  "options": {
    "include_security_analysis": true,
    "include_financial_analysis": true,
    "include_compliance_analysis": true,
    "include_reputation_analysis": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "risk_1234567890",
    "business_name": "Acme Corporation",
    "overall_risk": "LOW",
    "risk_score": 0.15,
    "risk_factors": {
      "security_risk": {
        "level": "LOW",
        "score": 0.1,
        "factors": [
          {
            "factor": "SSL Certificate",
            "status": "VALID",
            "details": "Valid SSL certificate until 2025-12-19"
          },
          {
            "factor": "Security Headers",
            "status": "GOOD",
            "details": "HSTS, CSP, and other security headers present"
          }
        ]
      },
      "financial_risk": {
        "level": "MEDIUM",
        "score": 0.3,
        "factors": [
          {
            "factor": "Company Size",
            "status": "UNKNOWN",
            "details": "Company size information not available"
          }
        ]
      },
      "compliance_risk": {
        "level": "LOW",
        "score": 0.05,
        "factors": [
          {
            "factor": "Data Protection",
            "status": "COMPLIANT",
            "details": "Privacy policy and data protection measures in place"
          }
        ]
      },
      "reputation_risk": {
        "level": "LOW",
        "score": 0.1,
        "factors": [
          {
            "factor": "Online Presence",
            "status": "GOOD",
            "details": "Professional website and social media presence"
          }
        ]
      }
    },
    "recommendations": [
      "Consider obtaining company size information for better financial risk assessment",
      "Monitor security headers regularly for compliance"
    ],
    "created_at": "2024-12-19T10:30:00Z"
  }
}
```

### GET /v1/risk/assess/{risk_assessment_id}

Retrieves a specific risk assessment result.

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "risk_1234567890",
    "business_name": "Acme Corporation",
    "overall_risk": "LOW",
    "risk_score": 0.15,
    "risk_factors": {...},
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### GET /v1/risk/history/{business_id}

Retrieves risk assessment history for a business.

**Query Parameters**:
- `limit` (optional): Number of results to return (default: 10, max: 100)
- `offset` (optional): Number of results to skip (default: 0)
- `start_date` (optional): Start date for filtering (ISO 8601 format)
- `end_date` (optional): End date for filtering (ISO 8601 format)

## Data Discovery Endpoints

### POST /v1/discover

Discovers and extracts comprehensive business information from multiple sources.

**Request**:
```json
{
  "business_name": "Acme Corporation",
  "website": "https://acme.com",
  "options": {
    "include_website_analysis": true,
    "include_web_search": true,
    "include_social_media": true,
    "max_results": 10
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "discover_1234567890",
    "business_name": "Acme Corporation",
    "website": "https://acme.com",
    "discovered_data": {
      "company_info": {
        "name": "Acme Corporation",
        "description": "Leading technology consulting firm",
        "founded": "2010",
        "headquarters": "San Francisco, CA",
        "contact": {
          "email": "contact@acme.com",
          "phone": "+1-555-123-4567",
          "address": "123 Tech Street, San Francisco, CA 94105"
        }
      },
      "team_info": {
        "size": "50-100 employees",
        "leadership": [
          {
            "name": "John Doe",
            "title": "CEO",
            "linkedin": "https://linkedin.com/in/johndoe"
          }
        ]
      },
      "products_services": [
        "Technology Consulting",
        "Software Development",
        "Digital Transformation",
        "Cloud Solutions"
      ],
      "business_model": "B2B",
      "technology_stack": [
        "React",
        "Node.js",
        "AWS",
        "Docker"
      ],
      "market_presence": {
        "regions": ["North America", "Europe"],
        "industries": ["Technology", "Finance", "Healthcare"],
        "competitors": ["TechCorp", "InnovateTech"]
      }
    },
    "data_quality": {
      "completeness": 0.85,
      "accuracy": 0.92,
      "consistency": 0.88,
      "freshness": 0.95,
      "overall": 0.90
    },
    "sources": [
      {
        "type": "website",
        "url": "https://acme.com",
        "confidence": 0.95
      },
      {
        "type": "web_search",
        "query": "Acme Corporation technology consulting",
        "confidence": 0.88
      }
    ],
    "created_at": "2024-12-19T10:30:00Z"
  }
}
```

### GET /v1/discover/{discovery_id}

Retrieves a specific data discovery result.

### GET /v1/discover/history/{business_id}

Retrieves data discovery history for a business.

## Caching Endpoints

### GET /v1/cache/{key}

Retrieves a cached value.

**Response**:
```json
{
  "success": true,
  "data": {
    "key": "classification_acme_corp",
    "value": {
      "classification": {...},
      "risk_assessment": {...}
    },
    "metadata": {
      "ttl": 3600,
      "created_at": "2024-12-19T10:30:00Z",
      "expires_at": "2024-12-19T11:30:00Z",
      "access_count": 5,
      "last_accessed": "2024-12-19T10:25:00Z"
    }
  }
}
```

### PUT /v1/cache/{key}

Stores a value in the cache.

**Request**:
```json
{
  "value": {
    "classification": {...},
    "risk_assessment": {...}
  },
  "ttl": 3600,
  "tags": ["classification", "acme_corp"],
  "priority": "high"
}
```

### DELETE /v1/cache/{key}

Removes a cached value.

### GET /v1/cache/stats

Retrieves cache statistics.

**Response**:
```json
{
  "success": true,
  "data": {
    "total_entries": 1000,
    "total_size": "50MB",
    "hit_rate": 0.85,
    "miss_rate": 0.15,
    "eviction_rate": 0.05,
    "average_ttl": 1800,
    "memory_usage": "25MB",
    "disk_usage": "25MB"
  }
}
```

### POST /v1/cache/optimize

Triggers cache optimization.

**Request**:
```json
{
  "strategy": "size_adjustment",
  "parameters": {
    "target_size": "100MB",
    "eviction_policy": "lru"
  }
}
```

## Monitoring Endpoints

### GET /v1/monitoring/metrics

Retrieves system metrics.

**Query Parameters**:
- `type` (optional): Metric type (performance, quality, errors)
- `timeframe` (optional): Timeframe for metrics (1h, 24h, 7d, 30d)
- `granularity` (optional): Metric granularity (1m, 5m, 1h, 1d)

**Response**:
```json
{
  "success": true,
  "data": {
    "performance": {
      "response_time": {
        "p50": 150,
        "p95": 300,
        "p99": 500,
        "average": 180
      },
      "throughput": {
        "requests_per_second": 100,
        "concurrent_users": 50
      }
    },
    "quality": {
      "accuracy_rate": 0.95,
      "confidence_average": 0.88,
      "misclassification_rate": 0.05
    },
    "errors": {
      "error_rate": 0.02,
      "error_types": {
        "validation_error": 0.01,
        "classification_error": 0.005,
        "internal_error": 0.005
      }
    },
    "resources": {
      "cpu_usage": 0.45,
      "memory_usage": 0.60,
      "disk_usage": 0.30
    }
  }
}
```

### GET /v1/monitoring/alerts

Retrieves active alerts.

**Response**:
```json
{
  "success": true,
  "data": {
    "alerts": [
      {
        "id": "alert_1234567890",
        "type": "performance",
        "severity": "warning",
        "message": "Response time exceeded threshold",
        "details": {
          "metric": "response_time_p95",
          "value": 350,
          "threshold": 300
        },
        "created_at": "2024-12-19T10:30:00Z",
        "status": "active"
      }
    ],
    "total_count": 1
  }
}
```

### POST /v1/monitoring/alerts/{alert_id}/acknowledge

Acknowledges an alert.

### GET /v1/monitoring/patterns

Retrieves pattern analysis results.

**Response**:
```json
{
  "success": true,
  "data": {
    "patterns": [
      {
        "id": "pattern_1234567890",
        "type": "misclassification",
        "severity": "medium",
        "description": "High misclassification rate for technology companies",
        "affected_businesses": 25,
        "confidence": 0.85,
        "recommendations": [
          "Update keyword database for technology sector",
          "Retrain ML model with more technology examples"
        ],
        "created_at": "2024-12-19T10:30:00Z"
      }
    ]
  }
}
```

## Health and Status Endpoints

### GET /v1/health

Retrieves system health status.

**Response**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "7d 12h 30m 15s",
    "timestamp": "2024-12-19T10:30:00Z",
    "components": {
      "database": {
        "status": "healthy",
        "response_time": "5ms"
      },
      "cache": {
        "status": "healthy",
        "hit_rate": 0.85
      },
      "external_apis": {
        "status": "healthy",
        "response_time": "150ms"
      }
    }
  }
}
```

### GET /v1/status

Retrieves detailed system status.

**Response**:
```json
{
  "success": true,
  "data": {
    "system": {
      "version": "1.0.0",
      "environment": "production",
      "region": "us-west-2",
      "instance_id": "i-1234567890abcdef0"
    },
    "performance": {
      "cpu_usage": 0.45,
      "memory_usage": 0.60,
      "disk_usage": 0.30,
      "network_io": "10MB/s"
    },
    "requests": {
      "total_requests": 1000000,
      "requests_per_second": 100,
      "error_rate": 0.02,
      "average_response_time": 180
    },
    "cache": {
      "total_entries": 1000,
      "hit_rate": 0.85,
      "memory_usage": "25MB",
      "disk_usage": "25MB"
    },
    "database": {
      "connections": 10,
      "active_queries": 5,
      "slow_queries": 0
    }
  }
}
```

### GET /v1/version

Retrieves API version information.

**Response**:
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "build_date": "2024-12-19T10:30:00Z",
    "git_commit": "abc123def456",
    "features": [
      "classification",
      "risk_assessment",
      "data_discovery",
      "caching",
      "monitoring"
    ],
    "deprecated_features": [],
    "upcoming_features": [
      "advanced_analytics",
      "machine_learning_enhancements"
    ]
  }
}
```

## SDK Examples

### Python SDK

```python
import kyb_client

# Initialize client
client = kyb_client.Client(api_key="YOUR_API_KEY")

# Classify business
result = client.classify(
    business_name="Acme Corporation",
    description="Technology consulting services",
    website="https://acme.com"
)

print(f"Primary classification: {result.primary_code.code}")
print(f"Confidence: {result.primary_code.confidence}")

# Assess risk
risk = client.assess_risk(
    business_name="Acme Corporation",
    website="https://acme.com"
)

print(f"Overall risk: {risk.overall_risk}")
print(f"Risk score: {risk.risk_score}")

# Discover data
discovery = client.discover_data(
    business_name="Acme Corporation",
    website="https://acme.com"
)

print(f"Company size: {discovery.team_info.size}")
print(f"Data quality: {discovery.data_quality.overall}")
```

### JavaScript SDK

```javascript
const { KYBClient } = require('kyb-client');

// Initialize client
const client = new KYBClient('YOUR_API_KEY');

// Classify business
const result = await client.classify({
    business_name: 'Acme Corporation',
    description: 'Technology consulting services',
    website: 'https://acme.com'
});

console.log(`Primary classification: ${result.primary_code.code}`);
console.log(`Confidence: ${result.primary_code.confidence}`);

// Assess risk
const risk = await client.assessRisk({
    business_name: 'Acme Corporation',
    website: 'https://acme.com'
});

console.log(`Overall risk: ${risk.overall_risk}`);
console.log(`Risk score: ${risk.risk_score}`);

// Discover data
const discovery = await client.discoverData({
    business_name: 'Acme Corporation',
    website: 'https://acme.com'
});

console.log(`Company size: ${discovery.team_info.size}`);
console.log(`Data quality: ${discovery.data_quality.overall}`);
```

## Webhook Integration

### Webhook Configuration

Configure webhooks to receive real-time notifications:

```json
{
  "url": "https://your-app.com/webhooks/kyb",
  "events": ["classification.completed", "risk_assessment.completed"],
  "secret": "webhook_secret_1234567890"
}
```

### Webhook Payload

```json
{
  "event": "classification.completed",
  "timestamp": "2024-12-19T10:30:00Z",
  "data": {
    "id": "class_1234567890",
    "business_name": "Acme Corporation",
    "classification": {
      "primary_code": {
        "type": "NAICS",
        "code": "541511",
        "description": "Custom Computer Programming Services",
        "confidence": 0.95
      }
    }
  }
}
```

## Rate Limiting

### Rate Limit Headers

All API responses include rate limit headers:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
X-RateLimit-Reset-Time: 2024-12-19T11:30:00Z
```

### Rate Limit Exceeded Response

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later.",
    "details": {
      "limit": 1000,
      "reset_time": "2024-12-19T11:30:00Z"
    },
    "request_id": "req_1234567890",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

## Best Practices

### Request Optimization

1. **Use appropriate timeouts**: Set reasonable timeouts for your requests
2. **Implement retry logic**: Retry failed requests with exponential backoff
3. **Cache responses**: Cache responses to reduce API calls
4. **Batch requests**: Use batch endpoints when available

### Error Handling

1. **Check status codes**: Always check HTTP status codes
2. **Handle rate limits**: Implement rate limit handling
3. **Log errors**: Log errors for debugging and monitoring
4. **Provide fallbacks**: Implement fallback mechanisms

### Security

1. **Secure API keys**: Keep API keys secure and rotate regularly
2. **Use HTTPS**: Always use HTTPS for API calls
3. **Validate responses**: Validate API responses before processing
4. **Monitor usage**: Monitor API usage for anomalies

## Data Models

### Enhanced Risk Assessment Models

#### RiskFactor
Represents a specific risk factor that contributes to overall risk assessment.

```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "category": "operational|financial|regulatory|reputational|cybersecurity",
  "weight": 0.0-1.0,
  "thresholds": {
    "minimal": 0.0-0.2,
    "low": 0.2-0.4,
    "medium": 0.4-0.6,
    "high": 0.6-0.8,
    "critical": 0.8-1.0
  },
  "metadata": {},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

#### RiskScore
Represents a calculated risk score for a specific factor or overall assessment.

```json
{
  "factor_id": "string",
  "factor_name": "string",
  "category": "operational|financial|regulatory|reputational|cybersecurity",
  "score": 0.0-100.0,
  "level": "minimal|low|medium|high|critical",
  "confidence": 0.0-1.0,
  "explanation": "string",
  "evidence": ["string"],
  "calculated_at": "2024-12-19T10:30:00Z"
}
```

#### EnhancedRiskAssessment
Represents a complete enhanced risk assessment for a business.

```json
{
  "id": "string",
  "business_id": "string",
  "business_name": "string",
  "overall_score": 0.0-100.0,
  "overall_level": "minimal|low|medium|high|critical",
  "category_scores": {
    "operational": { /* RiskScore object */ },
    "financial": { /* RiskScore object */ },
    "regulatory": { /* RiskScore object */ },
    "reputational": { /* RiskScore object */ },
    "cybersecurity": { /* RiskScore object */ }
  },
  "factor_scores": [/* RiskScore objects */],
  "recommendations": [/* RiskRecommendation objects */],
  "alerts": [/* RiskAlert objects */],
  "alert_level": "minimal|low|medium|high|critical",
  "assessed_at": "2024-12-19T10:30:00Z",
  "valid_until": "2024-12-19T10:30:00Z",
  "metadata": {
    "ml_model_version": "string",
    "assessment_method": "string",
    "processing_time_ms": 1250
  }
}
```

#### RiskRecommendation
Represents a recommendation to mitigate or address a risk.

```json
{
  "id": "string",
  "risk_factor": "string",
  "title": "string",
  "description": "string",
  "priority": "minimal|low|medium|high|critical",
  "action": "string",
  "impact": "string",
  "timeline": "string",
  "created_at": "2024-12-19T10:30:00Z"
}
```

#### RiskAlert
Represents an alert triggered by risk assessment.

```json
{
  "id": "string",
  "business_id": "string",
  "risk_factor": "string",
  "level": "minimal|low|medium|high|critical",
  "message": "string",
  "score": 0.0-100.0,
  "threshold": 0.0-100.0,
  "triggered_at": "2024-12-19T10:30:00Z",
  "acknowledged": false,
  "acknowledged_at": "2024-12-19T10:30:00Z"
}
```

### Enhanced Classification Models

#### EnhancedClassificationResult
Represents the result of enhanced business classification with ML models.

```json
{
  "id": "string",
  "business_name": "string",
  "classification": {
    "primary_code": {
      "type": "NAICS|SIC|MCC",
      "code": "string",
      "description": "string",
      "confidence": 0.0-1.0,
      "reasoning": "string",
      "ml_model_used": "string",
      "confidence_breakdown": {
        "ml_bert": 0.0-1.0,
        "ml_distilbert": 0.0-1.0,
        "custom_neural_net": 0.0-1.0,
        "keyword_matching": 0.0-1.0,
        "similarity_analysis": 0.0-1.0
      }
    },
    "alternatives": [/* ClassificationCode objects */]
  },
  "risk_assessment": {
    "overall_risk_level": "minimal|low|medium|high|critical",
    "risk_score": 0.0-1.0,
    "risk_factors": [/* RiskFactor objects */]
  },
  "ml_explainability": {
    "feature_importance": {
      "business_name": 0.0-1.0,
      "description": 0.0-1.0,
      "keywords": 0.0-1.0,
      "website_content": 0.0-1.0
    },
    "model_confidence": 0.0-1.0,
    "prediction_uncertainty": 0.0-1.0
  },
  "strategies_used": ["ml_bert", "ml_distilbert", "custom_neural_net", "keyword", "similarity"],
  "processing_time": "string",
  "ml_processing_time": "string"
}
```

#### ClassificationCode
Represents a classification code (NAICS, SIC, MCC) with confidence and reasoning.

```json
{
  "type": "NAICS|SIC|MCC",
  "code": "string",
  "description": "string",
  "confidence": 0.0-1.0,
  "reasoning": "string",
  "ml_model_used": "string"
}
```

### Risk Keywords Models

#### RiskKeyword
Represents a risk keyword used for risk detection and assessment.

```json
{
  "id": "integer",
  "keyword": "string",
  "risk_category": "illegal|prohibited|high_risk|tbml|sanctions|fraud",
  "risk_severity": "low|medium|high|critical",
  "description": "string",
  "mcc_codes": ["string"],
  "naics_codes": ["string"],
  "sic_codes": ["string"],
  "card_brand_restrictions": ["Visa", "Mastercard", "Amex"],
  "detection_patterns": ["string"],
  "synonyms": ["string"],
  "is_active": true,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

#### BusinessRiskAssessment
Represents a business risk assessment result with detected keywords and patterns.

```json
{
  "id": "uuid",
  "business_id": "uuid",
  "risk_keyword_id": "integer",
  "detected_keywords": ["string"],
  "risk_score": 0.0-1.0,
  "risk_level": "low|medium|high|critical",
  "assessment_method": "string",
  "website_content": "string",
  "detected_patterns": {},
  "assessment_date": "2024-12-19T10:30:00Z",
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### Industry Code Crosswalk Models

#### IndustryCodeCrosswalk
Represents the mapping between industries and various classification codes.

```json
{
  "id": "integer",
  "industry_id": "integer",
  "mcc_code": "string",
  "naics_code": "string",
  "sic_code": "string",
  "code_description": "string",
  "confidence_score": 0.0-1.0,
  "is_primary": false,
  "is_active": true,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### ML Model Metadata

#### MLModelInfo
Represents information about ML models used in classification and risk assessment.

```json
{
  "model_name": "string",
  "model_version": "string",
  "model_type": "bert|distilbert|custom_neural_net|anomaly_detection|pattern_recognition",
  "accuracy": 0.0-1.0,
  "confidence_threshold": 0.0-1.0,
  "training_date": "2024-12-19T10:30:00Z",
  "last_updated": "2024-12-19T10:30:00Z",
  "performance_metrics": {
    "precision": 0.0-1.0,
    "recall": 0.0-1.0,
    "f1_score": 0.0-1.0,
    "inference_time_ms": 100
  },
  "feature_importance": {},
  "is_active": true
}
```

## Conclusion

This API reference provides comprehensive documentation for all endpoints in the Enhanced Business Intelligence System. The API is designed to be:

- **RESTful**: Follows REST principles for consistency
- **Secure**: Multiple authentication methods and security features
- **Scalable**: Rate limiting and performance optimization
- **Reliable**: Comprehensive error handling and monitoring
- **Extensible**: Versioned API with backward compatibility
- **ML-Enhanced**: Advanced machine learning models for improved accuracy
- **Risk-Aware**: Comprehensive risk assessment and monitoring capabilities

For additional support, please refer to the SDK documentation or contact our support team.
