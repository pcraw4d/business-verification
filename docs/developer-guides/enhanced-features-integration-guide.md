# Enhanced Features Integration Guide

## Overview

This guide provides comprehensive documentation for integrating with the enhanced features of the KYB Platform, including advanced ML-powered classification, comprehensive risk assessment, and intelligent routing capabilities.

## Table of Contents

1. [Enhanced Classification Integration](#enhanced-classification-integration)
2. [Risk Assessment Integration](#risk-assessment-integration)
3. [ML Model Integration](#ml-model-integration)
4. [Feature Flag Management](#feature-flag-management)
5. [Error Handling and Fallbacks](#error-handling-and-fallbacks)
6. [Performance Optimization](#performance-optimization)
7. [Best Practices](#best-practices)

## Enhanced Classification Integration

### Basic Enhanced Classification

The enhanced classification system provides improved accuracy through ML models and comprehensive analysis.

```javascript
// Enhanced classification request
const classificationRequest = {
  business_name: "Acme Corporation",
  description: "AI-powered software development company specializing in machine learning solutions",
  website: "https://acme.com",
  industry: "Technology",
  keywords: ["AI", "machine learning", "software development", "consulting"],
  options: {
    include_alternatives: true,
    max_results: 5,
    confidence_threshold: 0.8,
    strategies: ["ml_bert", "ml_distilbert", "custom_neural_net", "keyword", "similarity"],
    include_risk_assessment: true,
    include_ml_explainability: true,
    include_confidence_breakdown: true
  }
};

// Make API call
const response = await fetch('https://api.kyb-platform.com/v2/classify', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(classificationRequest)
});

const result = await response.json();
```

### Handling Enhanced Classification Response

```javascript
function handleEnhancedClassificationResponse(result) {
  if (result.success) {
    const classification = result.data;
    
    // Access primary classification
    const primaryCode = classification.classification.primary_code;
    console.log(`Primary Classification: ${primaryCode.type} ${primaryCode.code}`);
    console.log(`Description: ${primaryCode.description}`);
    console.log(`Confidence: ${primaryCode.confidence}`);
    
    // Access ML model information
    if (primaryCode.ml_model_used) {
      console.log(`ML Model: ${primaryCode.ml_model_used}`);
    }
    
    // Access confidence breakdown
    if (primaryCode.confidence_breakdown) {
      console.log('Confidence Breakdown:', primaryCode.confidence_breakdown);
    }
    
    // Access risk assessment if included
    if (classification.risk_assessment) {
      console.log(`Risk Level: ${classification.risk_assessment.overall_risk_level}`);
      console.log(`Risk Score: ${classification.risk_assessment.risk_score}`);
    }
    
    // Access ML explainability
    if (classification.ml_explainability) {
      console.log('Feature Importance:', classification.ml_explainability.feature_importance);
      console.log('Model Confidence:', classification.ml_explainability.model_confidence);
    }
    
    // Access alternative classifications
    if (classification.classification.alternatives) {
      classification.classification.alternatives.forEach((alt, index) => {
        console.log(`Alternative ${index + 1}: ${alt.type} ${alt.code} (${alt.confidence})`);
      });
    }
  } else {
    console.error('Classification failed:', result.errors);
  }
}
```

### Batch Enhanced Classification

For processing multiple businesses efficiently:

```javascript
const batchRequest = {
  businesses: [
    {
      business_name: "Acme Corporation",
      description: "AI-powered software development",
      website: "https://acme.com"
    },
    {
      business_name: "Tech Solutions Inc",
      description: "Cloud computing and data analytics",
      website: "https://techsolutions.com"
    }
  ],
  options: {
    include_alternatives: true,
    max_results: 3,
    confidence_threshold: 0.8,
    strategies: ["ml_bert", "keyword", "similarity"],
    include_risk_assessment: true,
    parallel_processing: true
  }
};

const batchResponse = await fetch('https://api.kyb-platform.com/v2/classify/batch', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(batchRequest)
});

const batchResult = await batchResponse.json();
```

## Risk Assessment Integration

### Enhanced Risk Assessment

The enhanced risk assessment system provides comprehensive risk analysis with ML-powered detection.

```javascript
const riskAssessmentRequest = {
  business_name: "Acme Corporation",
  website: "https://acme.com",
  industry: "Technology",
  business_description: "Software development company specializing in AI solutions",
  options: {
    include_security_analysis: true,
    include_financial_analysis: true,
    include_compliance_analysis: true,
    include_reputation_analysis: true,
    include_ml_risk_detection: true,
    include_keyword_analysis: true,
    include_trend_analysis: true
  }
};

const riskResponse = await fetch('https://api.kyb-platform.com/v1/risk/enhanced/assess', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(riskAssessmentRequest)
});

const riskResult = await riskResponse.json();
```

### Handling Risk Assessment Response

```javascript
function handleRiskAssessmentResponse(result) {
  if (result.success) {
    const assessment = result.data;
    
    // Access overall risk information
    console.log(`Overall Risk Level: ${assessment.overall_risk_level}`);
    console.log(`Overall Risk Score: ${assessment.overall_risk_score}`);
    console.log(`Confidence Score: ${assessment.confidence_score}`);
    
    // Access individual risk factors
    if (assessment.risk_factors) {
      assessment.risk_factors.forEach(factor => {
        console.log(`Risk Factor: ${factor.factor_name}`);
        console.log(`Category: ${factor.category}`);
        console.log(`Score: ${factor.score}`);
        console.log(`Level: ${factor.level}`);
        console.log(`Explanation: ${factor.explanation}`);
        console.log('Evidence:', factor.evidence);
      });
    }
    
    // Access recommendations
    if (assessment.recommendations) {
      assessment.recommendations.forEach(rec => {
        console.log(`Recommendation: ${rec.title}`);
        console.log(`Priority: ${rec.priority}`);
        console.log(`Action: ${rec.action}`);
        console.log(`Timeline: ${rec.timeline}`);
      });
    }
    
    // Access trend data
    if (assessment.trend_data) {
      console.log(`Risk Trend: ${assessment.trend_data.risk_trend}`);
      console.log(`Change Percentage: ${assessment.trend_data.change_percentage}`);
    }
    
    // Access alerts
    if (assessment.alerts && assessment.alerts.length > 0) {
      assessment.alerts.forEach(alert => {
        console.log(`Alert: ${alert.message}`);
        console.log(`Level: ${alert.level}`);
        console.log(`Score: ${alert.score}`);
      });
    }
  } else {
    console.error('Risk assessment failed:', result.errors);
  }
}
```

### Risk Factor Calculation

For calculating specific risk factors:

```javascript
const riskFactorRequest = {
  business_id: "business_1234567890",
  factors: ["security", "financial", "compliance", "reputational"],
  options: {
    include_ml_analysis: true,
    include_historical_data: true
  }
};

const factorResponse = await fetch('https://api.kyb-platform.com/v1/risk/factors/calculate', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(riskFactorRequest)
});

const factorResult = await factorResponse.json();
```

### Risk Alerts Management

```javascript
// Get active alerts
const alertsResponse = await fetch('https://api.kyb-platform.com/v1/risk/alerts?business_id=business_1234567890', {
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY'
  }
});

const alertsResult = await alertsResponse.json();

// Acknowledge an alert
if (alertsResult.data.alerts.length > 0) {
  const alertId = alertsResult.data.alerts[0].id;
  
  const acknowledgeResponse = await fetch(`https://api.kyb-platform.com/v1/risk/alerts/${alertId}/acknowledge`, {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });
  
  const acknowledgeResult = await acknowledgeResponse.json();
}
```

## ML Model Integration

### Model Selection and Configuration

The platform supports multiple ML models with intelligent routing:

```javascript
// Configure ML model preferences
const mlConfig = {
  preferred_models: {
    classification: ["bert-base-uncased", "distilbert-base-uncased"],
    risk_detection: ["bert-risk-detection", "anomaly-detection"],
    fallback_models: ["keyword-matching", "similarity-analysis"]
  },
  confidence_thresholds: {
    high_confidence: 0.9,
    medium_confidence: 0.7,
    low_confidence: 0.5
  },
  performance_requirements: {
    max_processing_time_ms: 2000,
    min_accuracy: 0.85
  }
};

// Use in classification request
const classificationWithMLConfig = {
  business_name: "Acme Corporation",
  description: "AI-powered software development",
  website: "https://acme.com",
  options: {
    strategies: ["ml_bert", "ml_distilbert", "keyword"],
    ml_config: mlConfig,
    include_ml_explainability: true
  }
};
```

### Model Explainability

Access detailed explanations of ML model decisions:

```javascript
function analyzeMLExplainability(classificationResult) {
  if (classificationResult.ml_explainability) {
    const explainability = classificationResult.ml_explainability;
    
    // Feature importance analysis
    console.log('Feature Importance:');
    Object.entries(explainability.feature_importance).forEach(([feature, importance]) => {
      console.log(`${feature}: ${(importance * 100).toFixed(1)}%`);
    });
    
    // Model confidence
    console.log(`Model Confidence: ${(explainability.model_confidence * 100).toFixed(1)}%`);
    console.log(`Prediction Uncertainty: ${(explainability.prediction_uncertainty * 100).toFixed(1)}%`);
    
    // Use for decision making
    if (explainability.model_confidence < 0.8) {
      console.warn('Low model confidence - consider manual review');
    }
  }
}
```

## Feature Flag Management

### Dynamic Feature Toggle

The platform supports dynamic feature flags for gradual rollout and A/B testing:

```javascript
// Check feature flags
async function checkFeatureFlags() {
  const flagsResponse = await fetch('https://api.kyb-platform.com/v1/features/flags', {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });
  
  const flags = await flagsResponse.json();
  
  return {
    enhancedClassification: flags.data.enhanced_classification_enabled,
    mlModels: flags.data.ml_models_enabled,
    riskAssessment: flags.data.risk_assessment_enabled,
    specificModels: {
      bert: flags.data.bert_model_enabled,
      distilbert: flags.data.distilbert_model_enabled,
      customNeuralNet: flags.data.custom_neural_net_enabled
    }
  };
}

// Use feature flags in requests
async function makeIntelligentRequest(businessData) {
  const flags = await checkFeatureFlags();
  
  const requestOptions = {
    strategies: []
  };
  
  // Add strategies based on feature flags
  if (flags.specificModels.bert) {
    requestOptions.strategies.push("ml_bert");
  }
  if (flags.specificModels.distilbert) {
    requestOptions.strategies.push("ml_distilbert");
  }
  if (flags.specificModels.customNeuralNet) {
    requestOptions.strategies.push("custom_neural_net");
  }
  
  // Fallback to rule-based methods
  if (requestOptions.strategies.length === 0) {
    requestOptions.strategies = ["keyword", "similarity"];
  }
  
  const request = {
    ...businessData,
    options: requestOptions
  };
  
  return request;
}
```

## Error Handling and Fallbacks

### Comprehensive Error Handling

```javascript
async function robustClassificationRequest(businessData) {
  try {
    // Try enhanced classification first
    const enhancedResponse = await fetch('https://api.kyb-platform.com/v2/classify', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(businessData)
    });
    
    if (enhancedResponse.ok) {
      const result = await enhancedResponse.json();
      if (result.success) {
        return result;
      }
    }
    
    // Fallback to standard classification
    console.warn('Enhanced classification failed, falling back to standard classification');
    
    const standardResponse = await fetch('https://api.kyb-platform.com/v1/classify', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(businessData)
    });
    
    if (standardResponse.ok) {
      const result = await standardResponse.json();
      if (result.success) {
        return result;
      }
    }
    
    throw new Error('Both enhanced and standard classification failed');
    
  } catch (error) {
    console.error('Classification request failed:', error);
    
    // Return fallback result
    return {
      success: false,
      error: 'Classification service unavailable',
      fallback: true,
      data: {
        classification: {
          primary_code: {
            type: "UNKNOWN",
            code: "0000",
            description: "Classification unavailable",
            confidence: 0.0
          }
        }
      }
    };
  }
}
```

### Retry Logic with Exponential Backoff

```javascript
async function retryWithBackoff(fn, maxRetries = 3, baseDelay = 1000) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (attempt === maxRetries) {
        throw error;
      }
      
      const delay = baseDelay * Math.pow(2, attempt - 1);
      console.log(`Attempt ${attempt} failed, retrying in ${delay}ms...`);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
}

// Usage
const result = await retryWithBackoff(async () => {
  const response = await fetch('https://api.kyb-platform.com/v2/classify', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(businessData)
  });
  
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }
  
  return await response.json();
});
```

## Performance Optimization

### Caching Strategy

```javascript
class ClassificationCache {
  constructor(ttl = 3600000) { // 1 hour default TTL
    this.cache = new Map();
    this.ttl = ttl;
  }
  
  generateKey(businessData) {
    return btoa(JSON.stringify({
      name: businessData.business_name,
      description: businessData.description,
      website: businessData.website
    }));
  }
  
  get(businessData) {
    const key = this.generateKey(businessData);
    const cached = this.cache.get(key);
    
    if (cached && Date.now() - cached.timestamp < this.ttl) {
      return cached.data;
    }
    
    return null;
  }
  
  set(businessData, result) {
    const key = this.generateKey(businessData);
    this.cache.set(key, {
      data: result,
      timestamp: Date.now()
    });
  }
}

// Usage
const cache = new ClassificationCache();

async function getCachedClassification(businessData) {
  // Check cache first
  const cached = cache.get(businessData);
  if (cached) {
    console.log('Cache hit');
    return cached;
  }
  
  // Make API request
  const result = await robustClassificationRequest(businessData);
  
  // Cache successful results
  if (result.success) {
    cache.set(businessData, result);
  }
  
  return result;
}
```

### Batch Processing Optimization

```javascript
async function processBatchEfficiently(businesses, batchSize = 10) {
  const results = [];
  
  // Process in batches to avoid overwhelming the API
  for (let i = 0; i < businesses.length; i += batchSize) {
    const batch = businesses.slice(i, i + batchSize);
    
    try {
      const batchRequest = {
        businesses: batch,
        options: {
          include_alternatives: true,
          max_results: 3,
          confidence_threshold: 0.8,
          strategies: ["ml_bert", "keyword", "similarity"],
          parallel_processing: true
        }
      };
      
      const response = await fetch('https://api.kyb-platform.com/v2/classify/batch', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(batchRequest)
      });
      
      const result = await response.json();
      
      if (result.success) {
        results.push(...result.data.classifications);
      } else {
        console.error('Batch processing failed:', result.errors);
        // Add fallback results for failed batch
        batch.forEach(business => {
          results.push({
            business_name: business.business_name,
            classification: {
              primary_code: {
                type: "UNKNOWN",
                code: "0000",
                description: "Classification failed",
                confidence: 0.0
              }
            },
            error: true
          });
        });
      }
    } catch (error) {
      console.error('Batch processing error:', error);
      // Add fallback results
      batch.forEach(business => {
        results.push({
          business_name: business.business_name,
          classification: {
            primary_code: {
              type: "UNKNOWN",
              code: "0000",
              description: "Classification error",
              confidence: 0.0
            }
          },
          error: true
        });
      });
    }
    
    // Add delay between batches to respect rate limits
    if (i + batchSize < businesses.length) {
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }
  
  return results;
}
```

## Best Practices

### 1. Progressive Enhancement

```javascript
async function progressiveClassification(businessData) {
  // Start with basic classification
  let result = await basicClassification(businessData);
  
  // Enhance with ML if available
  if (result.success && result.data.confidence < 0.9) {
    try {
      const enhancedResult = await enhancedClassification(businessData);
      if (enhancedResult.success && enhancedResult.data.confidence > result.data.confidence) {
        result = enhancedResult;
      }
    } catch (error) {
      console.warn('Enhanced classification failed, using basic result:', error);
    }
  }
  
  // Add risk assessment if needed
  if (result.success && shouldAssessRisk(businessData)) {
    try {
      const riskResult = await riskAssessment(businessData);
      result.data.risk_assessment = riskResult.data;
    } catch (error) {
      console.warn('Risk assessment failed:', error);
    }
  }
  
  return result;
}
```

### 2. Monitoring and Observability

```javascript
class ClassificationMonitor {
  constructor() {
    this.metrics = {
      requests: 0,
      successes: 0,
      failures: 0,
      averageResponseTime: 0,
      cacheHits: 0,
      cacheMisses: 0
    };
  }
  
  recordRequest(startTime, success, fromCache = false) {
    const responseTime = Date.now() - startTime;
    
    this.metrics.requests++;
    if (success) {
      this.metrics.successes++;
    } else {
      this.metrics.failures++;
    }
    
    if (fromCache) {
      this.metrics.cacheHits++;
    } else {
      this.metrics.cacheMisses++;
    }
    
    // Update average response time
    this.metrics.averageResponseTime = 
      (this.metrics.averageResponseTime * (this.metrics.requests - 1) + responseTime) / 
      this.metrics.requests;
  }
  
  getMetrics() {
    return {
      ...this.metrics,
      successRate: this.metrics.requests > 0 ? this.metrics.successes / this.metrics.requests : 0,
      cacheHitRate: (this.metrics.cacheHits + this.metrics.cacheMisses) > 0 ? 
        this.metrics.cacheHits / (this.metrics.cacheHits + this.metrics.cacheMisses) : 0
    };
  }
}

// Usage
const monitor = new ClassificationMonitor();

async function monitoredClassification(businessData) {
  const startTime = Date.now();
  
  try {
    const result = await getCachedClassification(businessData);
    monitor.recordRequest(startTime, result.success, result.fromCache);
    return result;
  } catch (error) {
    monitor.recordRequest(startTime, false);
    throw error;
  }
}
```

### 3. Configuration Management

```javascript
class ClassificationConfig {
  constructor() {
    this.config = {
      apiBaseUrl: 'https://api.kyb-platform.com',
      apiKey: process.env.KYB_API_KEY,
      timeout: 30000,
      retryAttempts: 3,
      retryDelay: 1000,
      cacheTTL: 3600000,
      batchSize: 10,
      strategies: {
        default: ['ml_bert', 'keyword', 'similarity'],
        fallback: ['keyword', 'similarity'],
        highConfidence: ['ml_bert', 'ml_distilbert', 'custom_neural_net']
      },
      thresholds: {
        highConfidence: 0.9,
        mediumConfidence: 0.7,
        lowConfidence: 0.5
      }
    };
  }
  
  getStrategy(confidence = 'default') {
    return this.config.strategies[confidence] || this.config.strategies.default;
  }
  
  shouldUseEnhanced(confidence) {
    return confidence >= this.config.thresholds.mediumConfidence;
  }
  
  getApiUrl(endpoint) {
    return `${this.config.apiBaseUrl}${endpoint}`;
  }
}

// Usage
const config = new ClassificationConfig();

async function configuredClassification(businessData) {
  const strategy = config.getStrategy('default');
  
  const request = {
    ...businessData,
    options: {
      strategies: strategy,
      confidence_threshold: config.thresholds.mediumConfidence
    }
  };
  
  const response = await fetch(config.getApiUrl('/v2/classify'), {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${config.config.apiKey}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(request)
  });
  
  return await response.json();
}
```

## Conclusion

This integration guide provides comprehensive documentation for working with the enhanced features of the KYB Platform. By following these patterns and best practices, you can:

- Leverage advanced ML models for improved classification accuracy
- Implement comprehensive risk assessment capabilities
- Handle errors gracefully with appropriate fallbacks
- Optimize performance through caching and batch processing
- Monitor and observe system behavior
- Configure the system for your specific needs

For additional support or questions about specific integration scenarios, please refer to the API reference documentation or contact our support team.
