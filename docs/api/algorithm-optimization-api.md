# Algorithm Optimization API Documentation

## Overview

The Algorithm Optimization API provides automated optimization of classification algorithms based on pattern analysis. This API enables intelligent tuning of algorithm parameters, thresholds, and features to improve classification accuracy and reduce misclassifications.

## Base URL

```
/api/v1/algorithm-optimization
```

## Authentication

All endpoints require authentication. Include your API key in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

## Endpoints

### 1. Analyze and Optimize

Triggers analysis of misclassification patterns and performs automated optimizations.

**Endpoint:** `POST /api/v1/algorithm-optimization/analyze`

**Request Body:**
```json
{
  "force_optimization": false
}
```

**Parameters:**
- `force_optimization` (boolean, optional): Force optimization even if conditions aren't ideal

**Response:**
```json
{
  "status": "success",
  "message": "Analysis and optimization completed",
  "time": "2025-01-15T10:30:00Z"
}
```

**Status Codes:**
- `200 OK`: Optimization completed successfully
- `400 Bad Request`: Invalid request body
- `500 Internal Server Error`: Optimization failed

### 2. Get Optimization History

Retrieves the history of all optimizations performed.

**Endpoint:** `GET /api/v1/algorithm-optimization/history`

**Query Parameters:**
- `limit` (integer, optional): Maximum number of optimizations to return (default: 100)

**Response:**
```json
{
  "history": [
    {
      "id": "opt_threshold_1705312200",
      "algorithm_id": "ml_classifier",
      "optimization_type": "threshold",
      "status": "completed",
      "triggered_by_patterns": ["pattern-001", "pattern-002"],
      "before_metrics": {
        "accuracy": 0.85,
        "precision": 0.82,
        "recall": 0.88,
        "f1_score": 0.85,
        "misclassification_rate": 0.15,
        "confidence_score": 0.78,
        "processing_time": 150.5,
        "throughput": 66.4,
        "error_rate": 0.12
      },
      "after_metrics": {
        "accuracy": 0.89,
        "precision": 0.86,
        "recall": 0.92,
        "f1_score": 0.89,
        "misclassification_rate": 0.11,
        "confidence_score": 0.82,
        "processing_time": 145.2,
        "throughput": 68.9,
        "error_rate": 0.08
      },
      "improvement": {
        "accuracy_improvement": 0.04,
        "precision_improvement": 0.04,
        "recall_improvement": 0.04,
        "f1_score_improvement": 0.04,
        "misclassification_reduction": 0.04,
        "confidence_improvement": 0.04,
        "processing_time_improvement": 5.3,
        "overall_improvement": 0.04
      },
      "changes": [
        {
          "parameter": "confidence_threshold",
          "old_value": 0.7,
          "new_value": 0.75,
          "change_type": "threshold_adjustment",
          "impact": "medium",
          "confidence": 0.8
        }
      ],
      "optimization_time": "2025-01-15T10:30:00Z",
      "completion_time": "2025-01-15T10:32:15Z"
    }
  ],
  "total": 1,
  "limit": 100
}
```

### 3. Get Active Optimizations

Retrieves currently active optimizations.

**Endpoint:** `GET /api/v1/algorithm-optimization/active`

**Response:**
```json
{
  "active_optimizations": {
    "opt_weights_1705312300": {
      "id": "opt_weights_1705312300",
      "algorithm_id": "feature_classifier",
      "optimization_type": "weights",
      "status": "running",
      "optimization_time": "2025-01-15T10:35:00Z"
    }
  },
  "count": 1
}
```

### 4. Get Optimization Summary

Retrieves a summary of optimization performance.

**Endpoint:** `GET /api/v1/algorithm-optimization/summary`

**Response:**
```json
{
  "total_optimizations": 15,
  "successful_optimizations": 12,
  "failed_optimizations": 2,
  "average_improvement": 0.08,
  "optimizations_by_type": {
    "threshold": 8,
    "weights": 4,
    "features": 2,
    "model": 1
  },
  "optimizations_by_category": {
    "ml_classifier": 10,
    "feature_classifier": 3,
    "ensemble_classifier": 2
  }
}
```

### 5. Get Optimization by ID

Retrieves a specific optimization by its ID.

**Endpoint:** `GET /api/v1/algorithm-optimization/{id}`

**Path Parameters:**
- `id` (string, required): Optimization ID

**Response:**
```json
{
  "id": "opt_threshold_1705312200",
  "algorithm_id": "ml_classifier",
  "optimization_type": "threshold",
  "status": "completed",
  "triggered_by_patterns": ["pattern-001", "pattern-002"],
  "before_metrics": {
    "accuracy": 0.85,
    "precision": 0.82,
    "recall": 0.88,
    "f1_score": 0.85,
    "misclassification_rate": 0.15,
    "confidence_score": 0.78,
    "processing_time": 150.5,
    "throughput": 66.4,
    "error_rate": 0.12
  },
  "after_metrics": {
    "accuracy": 0.89,
    "precision": 0.86,
    "recall": 0.92,
    "f1_score": 0.89,
    "misclassification_rate": 0.11,
    "confidence_score": 0.82,
    "processing_time": 145.2,
    "throughput": 68.9,
    "error_rate": 0.08
  },
  "improvement": {
    "accuracy_improvement": 0.04,
    "precision_improvement": 0.04,
    "recall_improvement": 0.04,
    "f1_score_improvement": 0.04,
    "misclassification_reduction": 0.04,
    "confidence_improvement": 0.04,
    "processing_time_improvement": 5.3,
    "overall_improvement": 0.04
  },
  "changes": [
    {
      "parameter": "confidence_threshold",
      "old_value": 0.7,
      "new_value": 0.75,
      "change_type": "threshold_adjustment",
      "impact": "medium",
      "confidence": 0.8
    }
  ],
  "optimization_time": "2025-01-15T10:30:00Z",
  "completion_time": "2025-01-15T10:32:15Z"
}
```

**Status Codes:**
- `200 OK`: Optimization found
- `404 Not Found`: Optimization not found

### 6. Get Optimizations by Type

Retrieves optimizations filtered by type.

**Endpoint:** `GET /api/v1/algorithm-optimization/type/{type}`

**Path Parameters:**
- `type` (string, required): Optimization type (threshold, weights, features, model, ensemble, hyperparams)

**Response:**
```json
{
  "optimizations": [
    {
      "id": "opt_threshold_1705312200",
      "algorithm_id": "ml_classifier",
      "optimization_type": "threshold",
      "status": "completed",
      "optimization_time": "2025-01-15T10:30:00Z"
    }
  ],
  "type": "threshold",
  "count": 1
}
```

### 7. Get Optimizations by Algorithm

Retrieves optimizations filtered by algorithm ID.

**Endpoint:** `GET /api/v1/algorithm-optimization/algorithm/{algorithm_id}`

**Path Parameters:**
- `algorithm_id` (string, required): Algorithm ID

**Response:**
```json
{
  "optimizations": [
    {
      "id": "opt_threshold_1705312200",
      "algorithm_id": "ml_classifier",
      "optimization_type": "threshold",
      "status": "completed",
      "optimization_time": "2025-01-15T10:30:00Z"
    }
  ],
  "algorithm_id": "ml_classifier",
  "count": 1
}
```

### 8. Cancel Optimization

Cancels an active optimization.

**Endpoint:** `POST /api/v1/algorithm-optimization/{id}/cancel`

**Path Parameters:**
- `id` (string, required): Optimization ID

**Response:**
```json
{
  "status": "success",
  "message": "Optimization cancellation requested",
  "id": "opt_weights_1705312300",
  "time": "2025-01-15T10:40:00Z"
}
```

**Status Codes:**
- `200 OK`: Cancellation requested successfully
- `404 Not Found`: Optimization not found or not active

### 9. Rollback Optimization

Rolls back a completed optimization.

**Endpoint:** `POST /api/v1/algorithm-optimization/{id}/rollback`

**Path Parameters:**
- `id` (string, required): Optimization ID

**Request Body:**
```json
{
  "reason": "Performance degradation detected"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Optimization rollback requested",
  "id": "opt_threshold_1705312200",
  "reason": "Performance degradation detected",
  "time": "2025-01-15T10:45:00Z"
}
```

### 10. Get Optimization Recommendations

Retrieves optimization recommendations based on current patterns.

**Endpoint:** `GET /api/v1/algorithm-optimization/recommendations`

**Response:**
```json
{
  "recommendations": [
    {
      "id": "rec-001",
      "type": "threshold",
      "priority": "high",
      "title": "Adjust confidence thresholds",
      "description": "High-confidence misclassifications detected",
      "impact": "medium",
      "effort": "low",
      "confidence": 0.8,
      "actions": [
        "Lower confidence threshold for high-confidence errors",
        "Implement adaptive threshold adjustment"
      ]
    },
    {
      "id": "rec-002",
      "type": "features",
      "priority": "medium",
      "title": "Enhance feature extraction",
      "description": "Semantic patterns suggest feature improvements",
      "impact": "high",
      "effort": "medium",
      "confidence": 0.6,
      "actions": [
        "Add semantic feature extraction",
        "Implement context-aware features"
      ]
    }
  ],
  "count": 2,
  "generated_at": "2025-01-15T10:50:00Z"
}
```

## Data Models

### OptimizationResult

```json
{
  "id": "string",
  "algorithm_id": "string",
  "optimization_type": "string",
  "status": "string",
  "triggered_by_patterns": ["string"],
  "before_metrics": "AlgorithmMetrics",
  "after_metrics": "AlgorithmMetrics",
  "improvement": "ImprovementMetrics",
  "changes": ["AlgorithmChange"],
  "optimization_time": "datetime",
  "completion_time": "datetime",
  "error": "string",
  "recommendations": ["OptimizationRecommendation"]
}
```

### AlgorithmMetrics

```json
{
  "accuracy": "float",
  "precision": "float",
  "recall": "float",
  "f1_score": "float",
  "misclassification_rate": "float",
  "confidence_score": "float",
  "processing_time": "float",
  "throughput": "float",
  "error_rate": "float"
}
```

### ImprovementMetrics

```json
{
  "accuracy_improvement": "float",
  "precision_improvement": "float",
  "recall_improvement": "float",
  "f1_score_improvement": "float",
  "misclassification_reduction": "float",
  "confidence_improvement": "float",
  "processing_time_improvement": "float",
  "overall_improvement": "float"
}
```

### AlgorithmChange

```json
{
  "parameter": "string",
  "old_value": "any",
  "new_value": "any",
  "change_type": "string",
  "impact": "string",
  "confidence": "float"
}
```

### OptimizationRecommendation

```json
{
  "id": "string",
  "type": "string",
  "priority": "string",
  "title": "string",
  "description": "string",
  "impact": "string",
  "effort": "string",
  "confidence": "float",
  "actions": ["string"],
  "metadata": "object"
}
```

## Optimization Types

- `threshold`: Adjusts confidence thresholds
- `weights`: Modifies feature weights
- `features`: Enhances feature extraction
- `model`: Retrains or fine-tunes models
- `ensemble`: Optimizes ensemble methods
- `hyperparams`: Tunes hyperparameters

## Optimization Status

- `pending`: Optimization is queued
- `running`: Optimization is in progress
- `completed`: Optimization completed successfully
- `failed`: Optimization failed
- `rolled_back`: Optimization was rolled back

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- `400 Bad Request`: Invalid request parameters or body
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error during optimization

Error responses include a message describing the issue:

```json
{
  "error": "Optimization failed",
  "message": "Insufficient patterns for optimization",
  "code": "INSUFFICIENT_PATTERNS"
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse:

- Maximum 10 optimization requests per minute
- Maximum 100 API calls per hour per API key

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 8
X-RateLimit-Reset: 1705312800
```

## Best Practices

1. **Monitor Optimization Results**: Regularly check optimization history and summary to track improvements
2. **Review Recommendations**: Use the recommendations endpoint to get actionable insights
3. **Test Changes**: Always test optimizations in a staging environment before production
4. **Rollback Strategy**: Be prepared to rollback optimizations if performance degrades
5. **Pattern Analysis**: Ensure sufficient misclassification patterns exist before triggering optimizations

## Examples

### Trigger Optimization

```bash
curl -X POST https://api.example.com/api/v1/algorithm-optimization/analyze \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"force_optimization": false}'
```

### Get Optimization History

```bash
curl -X GET "https://api.example.com/api/v1/algorithm-optimization/history?limit=10" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Rollback Optimization

```bash
curl -X POST https://api.example.com/api/v1/algorithm-optimization/opt_123/rollback \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Performance degradation detected"}'
```
