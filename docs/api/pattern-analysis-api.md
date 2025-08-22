# Pattern Analysis API Documentation

## Overview

The Pattern Analysis API provides comprehensive analysis of misclassification patterns to help identify and address classification accuracy issues. This API enables deep analysis of misclassification data to detect patterns, generate recommendations, and provide insights for improving classification accuracy.

## Base URL

```
/api/v1/pattern-analysis
```

## Authentication

All endpoints require authentication. Include your API key in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

## Endpoints

### 1. Analyze Misclassifications

Analyzes a set of misclassification records to identify patterns and generate insights.

**Endpoint:** `POST /api/v1/pattern-analysis/analyze`

**Request Body:**
```json
{
  "misclassifications": [
    {
      "id": "mis_1234567890",
      "timestamp": "2024-01-15T10:30:00Z",
      "business_name": "Acme Corporation",
      "expected_classification": "Technology",
      "actual_classification": "Finance",
      "confidence_score": 0.85,
      "classification_method": "ml",
      "input_data": {
        "text": "software development company",
        "industry": "technology"
      },
      "error_type": "misclassification",
      "severity": "high",
      "root_cause": "ambiguous business description"
    }
  ],
  "config": {
    "min_confidence_threshold": 0.5,
    "max_patterns_to_track": 100,
    "analysis_window_hours": 24,
    "min_occurrences_for_pattern": 3
  }
}
```

**Response:**
```json
{
  "success": true,
  "result": {
    "id": "analysis_1234567890",
    "analysis_time": "2024-01-15T10:35:00Z",
    "patterns_found": 5,
    "new_patterns": 2,
    "updated_patterns": 3,
    "recommendations": [
      {
        "id": "rec_1234567890",
        "type": "algorithm_tuning",
        "priority": "high",
        "title": "Improve Technology-Finance Classification",
        "description": "High confidence misclassifications between Technology and Finance categories",
        "actions": [
          "Review training data for Technology category",
          "Add more examples of technology companies",
          "Adjust confidence thresholds"
        ],
        "impact": "high",
        "effort": "medium",
        "metadata": {
          "affected_categories": ["Technology", "Finance"],
          "confidence_range": [0.8, 0.95]
        }
      }
    ],
    "summary": {
      "total_misclassifications": 50,
      "patterns_by_type": {
        "temporal": 2,
        "semantic": 3,
        "confidence": 1
      },
      "patterns_by_severity": {
        "critical": 1,
        "high": 3,
        "medium": 1
      },
      "risk_level": "high"
    }
  },
  "metadata": {
    "analyzed_at": "2024-01-15T10:35:00Z",
    "count": 50
  }
}
```

### 2. Get All Patterns

Retrieves all detected misclassification patterns.

**Endpoint:** `GET /api/v1/pattern-analysis/patterns`

**Response:**
```json
{
  "success": true,
  "patterns": {
    "pattern_1234567890": {
      "id": "pattern_1234567890",
      "pattern_type": "semantic",
      "category": "classification_error",
      "severity": "high",
      "confidence": 0.92,
      "impact_score": 0.85,
      "occurrences": 15,
      "first_seen": "2024-01-10T08:00:00Z",
      "last_seen": "2024-01-15T10:30:00Z",
      "characteristics": {
        "keywords": ["software", "tech", "development"],
        "phrases": ["software company", "tech startup"],
        "confidence_range": [0.8, 0.95],
        "time_of_day_distribution": {
          "morning": 0.4,
          "afternoon": 0.35,
          "evening": 0.25
        }
      },
      "affected_categories": ["Technology", "Finance"],
      "root_causes": [
        {
          "type": "ambiguous_description",
          "description": "Business descriptions contain ambiguous terms",
          "confidence": 0.88
        }
      ]
    }
  },
  "count": 5,
  "metadata": {
    "retrieved_at": "2024-01-15T10:35:00Z"
  }
}
```

### 3. Get Patterns by Type

Retrieves patterns filtered by pattern type.

**Endpoint:** `GET /api/v1/pattern-analysis/patterns/type/{type}`

**Path Parameters:**
- `type`: Pattern type (temporal, semantic, confidence, input, cross_dimensional)

**Response:**
```json
{
  "success": true,
  "patterns": {
    "pattern_1234567890": {
      "id": "pattern_1234567890",
      "pattern_type": "temporal",
      "category": "classification_error",
      "severity": "medium",
      "confidence": 0.78,
      "impact_score": 0.65,
      "occurrences": 8,
      "first_seen": "2024-01-10T08:00:00Z",
      "last_seen": "2024-01-15T10:30:00Z",
      "characteristics": {
        "time_periods": ["morning", "afternoon"],
        "frequency": "daily",
        "trend": "increasing"
      }
    }
  },
  "pattern_type": "temporal",
  "count": 2,
  "metadata": {
    "retrieved_at": "2024-01-15T10:35:00Z"
  }
}
```

### 4. Get Patterns by Severity

Retrieves patterns filtered by severity level.

**Endpoint:** `GET /api/v1/pattern-analysis/patterns/severity/{severity}`

**Path Parameters:**
- `severity`: Pattern severity (critical, high, medium, low)

**Response:**
```json
{
  "success": true,
  "patterns": {
    "pattern_1234567890": {
      "id": "pattern_1234567890",
      "pattern_type": "semantic",
      "category": "classification_error",
      "severity": "high",
      "confidence": 0.92,
      "impact_score": 0.85,
      "occurrences": 15,
      "first_seen": "2024-01-10T08:00:00Z",
      "last_seen": "2024-01-15T10:30:00Z",
      "characteristics": {
        "keywords": ["software", "tech", "development"],
        "phrases": ["software company", "tech startup"],
        "confidence_range": [0.8, 0.95]
      }
    }
  },
  "severity": "high",
  "count": 3,
  "metadata": {
    "retrieved_at": "2024-01-15T10:35:00Z"
  }
}
```

### 5. Get Pattern Details

Retrieves detailed information about a specific pattern.

**Endpoint:** `GET /api/v1/pattern-analysis/patterns/{id}`

**Path Parameters:**
- `id`: Pattern ID

**Response:**
```json
{
  "success": true,
  "pattern": {
    "id": "pattern_1234567890",
    "pattern_type": "semantic",
    "category": "classification_error",
    "severity": "high",
    "confidence": 0.92,
    "impact_score": 0.85,
    "occurrences": 15,
    "first_seen": "2024-01-10T08:00:00Z",
    "last_seen": "2024-01-15T10:30:00Z",
    "characteristics": {
      "keywords": ["software", "tech", "development"],
      "phrases": ["software company", "tech startup"],
      "confidence_range": [0.8, 0.95],
      "time_of_day_distribution": {
        "morning": 0.4,
        "afternoon": 0.35,
        "evening": 0.25
      }
    },
    "affected_categories": ["Technology", "Finance"],
    "root_causes": [
      {
        "type": "ambiguous_description",
        "description": "Business descriptions contain ambiguous terms",
        "confidence": 0.88
      }
    ]
  },
  "metadata": {
    "retrieved_at": "2024-01-15T10:35:00Z"
  }
}
```

### 6. Get Pattern History

Retrieves the history of pattern analysis results.

**Endpoint:** `GET /api/v1/pattern-analysis/history`

**Query Parameters:**
- `limit`: Maximum number of history entries to return (default: 50, max: 100)

**Response:**
```json
{
  "success": true,
  "history": [
    {
      "id": "analysis_1234567890",
      "analysis_time": "2024-01-15T10:35:00Z",
      "patterns_found": 5,
      "new_patterns": 2,
      "updated_patterns": 3,
      "misclassifications_analyzed": 50,
      "summary": {
        "total_misclassifications": 50,
        "patterns_by_type": {
          "temporal": 2,
          "semantic": 3
        },
        "risk_level": "high"
      }
    }
  ],
  "count": 10,
  "limit": 50,
  "metadata": {
    "retrieved_at": "2024-01-15T10:35:00Z"
  }
}
```

### 7. Get Pattern Summary

Retrieves a summary of pattern analysis statistics.

**Endpoint:** `GET /api/v1/pattern-analysis/summary`

**Response:**
```json
{
  "success": true,
  "summary": {
    "total_patterns": 5,
    "patterns_by_type": {
      "temporal": 2,
      "semantic": 3
    },
    "patterns_by_category": {
      "classification_error": 4,
      "data_quality": 1
    },
    "patterns_by_severity": {
      "critical": 1,
      "high": 3,
      "medium": 1
    },
    "critical_patterns": 1,
    "high_impact_patterns": 3,
    "average_impact": 0.75,
    "risk_level": "high",
    "analysis_history": {
      "total_analyses": 10,
      "last_analysis": {
        "id": "analysis_1234567890",
        "analysis_time": "2024-01-15T10:35:00Z",
        "patterns_found": 5,
        "new_patterns": 2
      }
    }
  },
  "metadata": {
    "generated_at": "2024-01-15T10:35:00Z"
  }
}
```

### 8. Get Recommendations

Retrieves recommendations based on current patterns.

**Endpoint:** `GET /api/v1/pattern-analysis/recommendations`

**Response:**
```json
{
  "success": true,
  "recommendations": [
    {
      "id": "rec_1234567890",
      "type": "algorithm_tuning",
      "priority": "high",
      "title": "Improve Technology-Finance Classification",
      "description": "High confidence misclassifications between Technology and Finance categories",
      "actions": [
        "Review training data for Technology category",
        "Add more examples of technology companies",
        "Adjust confidence thresholds"
      ],
      "impact": "high",
      "effort": "medium",
      "metadata": {
        "affected_categories": ["Technology", "Finance"],
        "confidence_range": [0.8, 0.95]
      }
    }
  ],
  "count": 3,
  "metadata": {
    "generated_at": "2024-01-15T10:35:00Z"
  }
}
```

### 9. Health Check

Provides health status for the pattern analysis engine.

**Endpoint:** `GET /api/v1/pattern-analysis/health`

**Response:**
```json
{
  "status": "healthy",
  "stats": {
    "active_patterns": 5,
    "analysis_history": 10,
    "uptime": "5h30m15s"
  },
  "metadata": {
    "checked_at": "2024-01-15T10:35:00Z"
  }
}
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request body",
  "message": "No misclassifications provided"
}
```

### 404 Not Found
```json
{
  "error": "Pattern not found",
  "message": "Pattern with ID 'nonexistent' was not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Analysis failed",
  "message": "Failed to analyze misclassifications"
}
```

## Data Types

### MisclassificationRecord
```json
{
  "id": "string",
  "timestamp": "datetime",
  "business_name": "string",
  "expected_classification": "string",
  "actual_classification": "string",
  "confidence_score": "float",
  "classification_method": "string",
  "input_data": "object",
  "error_type": "string",
  "severity": "string",
  "root_cause": "string",
  "action_required": "boolean"
}
```

### PatternAnalysisConfig
```json
{
  "min_confidence_threshold": "float",
  "max_patterns_to_track": "integer",
  "analysis_window_hours": "integer",
  "min_occurrences_for_pattern": "integer"
}
```

### MisclassificationPattern
```json
{
  "id": "string",
  "pattern_type": "string",
  "category": "string",
  "severity": "string",
  "confidence": "float",
  "impact_score": "float",
  "occurrences": "integer",
  "first_seen": "datetime",
  "last_seen": "datetime",
  "characteristics": "object",
  "affected_categories": "array",
  "root_causes": "array"
}
```

## Rate Limiting

- **Rate Limit:** 100 requests per minute per API key
- **Burst Limit:** 10 requests per second

## Pagination

For endpoints that return large datasets, pagination is supported using query parameters:

- `limit`: Number of items per page (default: 50, max: 100)
- `offset`: Number of items to skip

## Webhooks

The Pattern Analysis API supports webhooks for real-time notifications when new patterns are detected or when pattern severity changes.

**Webhook Event Types:**
- `pattern.detected`: New pattern detected
- `pattern.updated`: Existing pattern updated
- `pattern.severity_changed`: Pattern severity level changed
- `recommendation.generated`: New recommendation generated

## SDK Support

Official SDKs are available for:
- Go
- Python
- JavaScript/TypeScript
- Java

## Support

For API support and questions:
- Email: api-support@company.com
- Documentation: https://docs.company.com/api/pattern-analysis
- Status Page: https://status.company.com
