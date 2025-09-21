# ML Model Integration Guide

## Overview

This guide provides comprehensive documentation for integrating with the machine learning models in the KYB Platform, including BERT, DistilBERT, custom neural networks, and risk detection models.

## Table of Contents

1. [ML Model Architecture](#ml-model-architecture)
2. [Model Selection Strategy](#model-selection-strategy)
3. [Feature Engineering](#feature-engineering)
4. [Model Performance Optimization](#model-performance-optimization)
5. [Model Explainability](#model-explainability)
6. [A/B Testing and Model Comparison](#ab-testing-and-model-comparison)
7. [Model Monitoring and Drift Detection](#model-monitoring-and-drift-detection)
8. [Best Practices](#best-practices)

## ML Model Architecture

### Available Models

The platform supports multiple ML models for different use cases:

#### Classification Models
- **BERT (bert-base-uncased)**: High-accuracy classification with semantic understanding
- **DistilBERT**: Faster inference with 60% of BERT's parameters
- **Custom Neural Networks**: Specialized models for specific industry sectors

#### Risk Detection Models
- **BERT-based Risk Classification**: Semantic risk assessment
- **Anomaly Detection**: Unusual pattern identification
- **Pattern Recognition**: Complex risk scenario detection

### Model Performance Characteristics

| Model | Accuracy | Inference Time | Use Case |
|-------|----------|----------------|----------|
| BERT | 95%+ | 100-200ms | High-accuracy classification |
| DistilBERT | 92%+ | 50-100ms | Fast classification |
| Custom Neural Net | 90%+ | 30-80ms | Industry-specific classification |
| Risk Detection BERT | 90%+ | 80-150ms | Risk assessment |
| Anomaly Detection | 85%+ | 20-50ms | Unusual pattern detection |

## Model Selection Strategy

### Intelligent Model Routing

The platform automatically selects the best model based on request characteristics:

```javascript
class ModelSelector {
  constructor() {
    this.modelPreferences = {
      highAccuracy: ['bert-base-uncased', 'custom-neural-net'],
      fastInference: ['distilbert-base-uncased', 'keyword-matching'],
      riskAssessment: ['bert-risk-detection', 'anomaly-detection'],
      industrySpecific: ['custom-neural-net', 'bert-base-uncased']
    };
  }
  
  selectModel(request) {
    const { businessData, requirements } = request;
    
    // High accuracy requirement
    if (requirements.accuracy > 0.95) {
      return this.modelPreferences.highAccuracy;
    }
    
    // Fast inference requirement
    if (requirements.maxProcessingTime < 100) {
      return this.modelPreferences.fastInference;
    }
    
    // Risk assessment requirement
    if (requirements.includeRiskAssessment) {
      return this.modelPreferences.riskAssessment;
    }
    
    // Industry-specific classification
    if (this.isIndustrySpecific(businessData.industry)) {
      return this.modelPreferences.industrySpecific;
    }
    
    // Default selection
    return ['bert-base-uncased', 'distilbert-base-uncased'];
  }
  
  isIndustrySpecific(industry) {
    const specializedIndustries = [
      'Healthcare', 'Finance', 'Legal', 'Real Estate', 
      'Manufacturing', 'Technology', 'Retail'
    ];
    return specializedIndustries.includes(industry);
  }
}
```

### Manual Model Selection

You can also manually specify which models to use:

```javascript
const classificationRequest = {
  business_name: "Acme Corporation",
  description: "AI-powered software development company",
  website: "https://acme.com",
  options: {
    strategies: [
      "ml_bert",           // BERT model
      "ml_distilbert",     // DistilBERT model
      "custom_neural_net", // Custom neural network
      "keyword",           // Keyword matching (fallback)
      "similarity"         // Similarity analysis (fallback)
    ],
    model_preferences: {
      primary: "ml_bert",
      fallback: ["ml_distilbert", "keyword"],
      confidence_threshold: 0.8
    }
  }
};
```

## Feature Engineering

### Input Feature Preparation

The platform automatically extracts and prepares features from your input data:

```javascript
class FeatureExtractor {
  extractFeatures(businessData) {
    const features = {
      // Text features
      business_name: this.preprocessText(businessData.business_name),
      description: this.preprocessText(businessData.description),
      website_content: this.extractWebsiteContent(businessData.website),
      
      // Structured features
      industry: businessData.industry,
      keywords: businessData.keywords || [],
      
      // Derived features
      text_length: businessData.description?.length || 0,
      keyword_count: businessData.keywords?.length || 0,
      has_website: !!businessData.website,
      
      // Temporal features
      timestamp: new Date().toISOString()
    };
    
    return features;
  }
  
  preprocessText(text) {
    if (!text) return '';
    
    return text
      .toLowerCase()
      .replace(/[^\w\s]/g, ' ')  // Remove special characters
      .replace(/\s+/g, ' ')      // Normalize whitespace
      .trim();
  }
  
  extractWebsiteContent(website) {
    // This would be handled by the platform's website scraping system
    return null; // Placeholder
  }
}
```

### Feature Importance Analysis

Understand which features contribute most to model decisions:

```javascript
function analyzeFeatureImportance(mlExplainability) {
  if (!mlExplainability?.feature_importance) {
    return null;
  }
  
  const importance = mlExplainability.feature_importance;
  
  // Sort features by importance
  const sortedFeatures = Object.entries(importance)
    .sort(([,a], [,b]) => b - a)
    .map(([feature, score]) => ({
      feature,
      importance: score,
      percentage: (score * 100).toFixed(1) + '%'
    }));
  
  console.log('Feature Importance Analysis:');
  sortedFeatures.forEach(({ feature, importance, percentage }) => {
    console.log(`${feature}: ${percentage} (${importance.toFixed(3)})`);
  });
  
  return sortedFeatures;
}

// Usage
const result = await enhancedClassification(businessData);
if (result.ml_explainability) {
  const featureAnalysis = analyzeFeatureImportance(result.ml_explainability);
}
```

## Model Performance Optimization

### Caching Strategy for ML Models

```javascript
class MLModelCache {
  constructor() {
    this.cache = new Map();
    this.modelCache = new Map();
  }
  
  // Cache model predictions
  cachePrediction(inputHash, modelName, prediction) {
    const key = `${inputHash}_${modelName}`;
    this.cache.set(key, {
      prediction,
      timestamp: Date.now(),
      ttl: 3600000 // 1 hour
    });
  }
  
  // Get cached prediction
  getCachedPrediction(inputHash, modelName) {
    const key = `${inputHash}_${modelName}`;
    const cached = this.cache.get(key);
    
    if (cached && Date.now() - cached.timestamp < cached.ttl) {
      return cached.prediction;
    }
    
    return null;
  }
  
  // Cache model metadata
  cacheModelMetadata(modelName, metadata) {
    this.modelCache.set(modelName, {
      ...metadata,
      cached_at: Date.now()
    });
  }
  
  // Get model metadata
  getModelMetadata(modelName) {
    return this.modelCache.get(modelName);
  }
  
  // Generate input hash for caching
  generateInputHash(businessData) {
    const inputString = JSON.stringify({
      name: businessData.business_name,
      description: businessData.description,
      website: businessData.website
    });
    
    // Simple hash function (in production, use a proper hash library)
    let hash = 0;
    for (let i = 0; i < inputString.length; i++) {
      const char = inputString.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    
    return hash.toString();
  }
}
```

### Batch Processing for ML Models

```javascript
class MLBatchProcessor {
  constructor(batchSize = 10, maxConcurrency = 3) {
    this.batchSize = batchSize;
    this.maxConcurrency = maxConcurrency;
    this.queue = [];
    this.processing = false;
  }
  
  async processBatch(businesses) {
    const batches = this.createBatches(businesses, this.batchSize);
    const results = [];
    
    // Process batches with controlled concurrency
    for (let i = 0; i < batches.length; i += this.maxConcurrency) {
      const concurrentBatches = batches.slice(i, i + this.maxConcurrency);
      
      const batchPromises = concurrentBatches.map(batch => 
        this.processSingleBatch(batch)
      );
      
      const batchResults = await Promise.all(batchPromises);
      results.push(...batchResults.flat());
      
      // Add delay between concurrent batches
      if (i + this.maxConcurrency < batches.length) {
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    }
    
    return results;
  }
  
  createBatches(items, batchSize) {
    const batches = [];
    for (let i = 0; i < items.length; i += batchSize) {
      batches.push(items.slice(i, i + batchSize));
    }
    return batches;
  }
  
  async processSingleBatch(batch) {
    const batchRequest = {
      businesses: batch,
      options: {
        strategies: ["ml_bert", "ml_distilbert"],
        include_ml_explainability: true,
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
    return result.data?.classifications || [];
  }
}
```

## Model Explainability

### Understanding Model Decisions

```javascript
class ModelExplainer {
  explainPrediction(classificationResult) {
    const explanation = {
      primaryClassification: this.explainPrimaryClassification(classificationResult),
      alternatives: this.explainAlternatives(classificationResult),
      confidence: this.explainConfidence(classificationResult),
      riskFactors: this.explainRiskFactors(classificationResult)
    };
    
    return explanation;
  }
  
  explainPrimaryClassification(result) {
    const primary = result.classification.primary_code;
    
    return {
      code: primary.code,
      description: primary.description,
      confidence: primary.confidence,
      reasoning: primary.reasoning,
      modelUsed: primary.ml_model_used,
      confidenceBreakdown: primary.confidence_breakdown,
      explanation: this.generateExplanation(primary)
    };
  }
  
  generateExplanation(classification) {
    const explanations = [];
    
    if (classification.confidence_breakdown) {
      const breakdown = classification.confidence_breakdown;
      
      if (breakdown.ml_bert > 0.8) {
        explanations.push("BERT model shows high confidence in this classification");
      }
      
      if (breakdown.keyword_matching > 0.7) {
        explanations.push("Strong keyword matches support this classification");
      }
      
      if (breakdown.similarity_analysis > 0.6) {
        explanations.push("Similarity analysis confirms this classification");
      }
    }
    
    if (classification.reasoning) {
      explanations.push(classification.reasoning);
    }
    
    return explanations.join('. ');
  }
  
  explainConfidence(result) {
    const confidence = result.ml_explainability?.model_confidence || 0;
    
    let level = 'Low';
    if (confidence > 0.9) level = 'Very High';
    else if (confidence > 0.8) level = 'High';
    else if (confidence > 0.7) level = 'Medium';
    
    return {
      level,
      score: confidence,
      uncertainty: result.ml_explainability?.prediction_uncertainty || 0,
      recommendation: this.getConfidenceRecommendation(confidence)
    };
  }
  
  getConfidenceRecommendation(confidence) {
    if (confidence > 0.9) {
      return "High confidence - suitable for automated processing";
    } else if (confidence > 0.7) {
      return "Medium confidence - consider human review for critical decisions";
    } else {
      return "Low confidence - manual review recommended";
    }
  }
}
```

### Feature Attribution Analysis

```javascript
function analyzeFeatureAttribution(mlExplainability) {
  if (!mlExplainability?.feature_importance) {
    return null;
  }
  
  const attribution = {
    topFeatures: [],
    insights: [],
    recommendations: []
  };
  
  // Get top contributing features
  const features = Object.entries(mlExplainability.feature_importance)
    .sort(([,a], [,b]) => b - a)
    .slice(0, 5);
  
  attribution.topFeatures = features.map(([feature, importance]) => ({
    feature,
    importance,
    contribution: `${(importance * 100).toFixed(1)}%`
  }));
  
  // Generate insights
  const topFeature = features[0];
  if (topFeature && topFeature[1] > 0.4) {
    attribution.insights.push(
      `${topFeature[0]} is the most important factor (${(topFeature[1] * 100).toFixed(1)}%)`
    );
  }
  
  // Generate recommendations
  if (mlExplainability.model_confidence < 0.8) {
    attribution.recommendations.push(
      "Consider providing more detailed business description to improve classification accuracy"
    );
  }
  
  return attribution;
}
```

## A/B Testing and Model Comparison

### Model Performance Comparison

```javascript
class ModelComparator {
  constructor() {
    this.comparisonResults = [];
  }
  
  async compareModels(businessData, models = ['ml_bert', 'ml_distilbert', 'custom_neural_net']) {
    const results = [];
    
    for (const model of models) {
      try {
        const startTime = Date.now();
        
        const request = {
          ...businessData,
          options: {
            strategies: [model],
            include_ml_explainability: true
          }
        };
        
        const response = await fetch('https://api.kyb-platform.com/v2/classify', {
          method: 'POST',
          headers: {
            'Authorization': 'Bearer YOUR_API_KEY',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(request)
        });
        
        const result = await response.json();
        const processingTime = Date.now() - startTime;
        
        if (result.success) {
          results.push({
            model,
            classification: result.data.classification.primary_code,
            confidence: result.data.classification.primary_code.confidence,
            processingTime,
            explainability: result.data.ml_explainability,
            success: true
          });
        } else {
          results.push({
            model,
            error: result.errors,
            success: false
          });
        }
      } catch (error) {
        results.push({
          model,
          error: error.message,
          success: false
        });
      }
    }
    
    return this.analyzeComparison(results);
  }
  
  analyzeComparison(results) {
    const successfulResults = results.filter(r => r.success);
    
    if (successfulResults.length === 0) {
      return {
        success: false,
        error: "All models failed",
        results
      };
    }
    
    // Find best performing model
    const bestModel = successfulResults.reduce((best, current) => {
      return current.confidence > best.confidence ? current : best;
    });
    
    // Calculate performance metrics
    const avgConfidence = successfulResults.reduce((sum, r) => sum + r.confidence, 0) / successfulResults.length;
    const avgProcessingTime = successfulResults.reduce((sum, r) => sum + r.processingTime, 0) / successfulResults.length;
    
    return {
      success: true,
      bestModel: bestModel.model,
      bestConfidence: bestModel.confidence,
      averageConfidence: avgConfidence,
      averageProcessingTime: avgProcessingTime,
      results: successfulResults,
      recommendations: this.generateRecommendations(successfulResults)
    };
  }
  
  generateRecommendations(results) {
    const recommendations = [];
    
    // Accuracy recommendation
    const highAccuracyModels = results.filter(r => r.confidence > 0.9);
    if (highAccuracyModels.length > 0) {
      recommendations.push(
        `Use ${highAccuracyModels[0].model} for high-accuracy requirements (${(highAccuracyModels[0].confidence * 100).toFixed(1)}% confidence)`
      );
    }
    
    // Speed recommendation
    const fastModels = results.filter(r => r.processingTime < 100);
    if (fastModels.length > 0) {
      recommendations.push(
        `Use ${fastModels[0].model} for fast processing (${fastModels[0].processingTime}ms)`
      );
    }
    
    return recommendations;
  }
}
```

### A/B Testing Framework

```javascript
class ABTestFramework {
  constructor() {
    this.tests = new Map();
  }
  
  createTest(testName, variants, trafficSplit = 0.5) {
    const test = {
      name: testName,
      variants: variants,
      trafficSplit: trafficSplit,
      results: {
        variantA: { requests: 0, successes: 0, avgConfidence: 0, avgProcessingTime: 0 },
        variantB: { requests: 0, successes: 0, avgConfidence: 0, avgProcessingTime: 0 }
      },
      startTime: Date.now(),
      active: true
    };
    
    this.tests.set(testName, test);
    return test;
  }
  
  selectVariant(testName, userId) {
    const test = this.tests.get(testName);
    if (!test || !test.active) {
      return null;
    }
    
    // Simple hash-based variant selection
    const hash = this.hashString(userId + testName);
    const variant = hash < test.trafficSplit ? 'variantA' : 'variantB';
    
    return {
      variant,
      config: test.variants[variant]
    };
  }
  
  recordResult(testName, variant, result) {
    const test = this.tests.get(testName);
    if (!test) return;
    
    const variantResults = test.results[variant];
    variantResults.requests++;
    
    if (result.success) {
      variantResults.successes++;
      variantResults.avgConfidence = 
        (variantResults.avgConfidence * (variantResults.successes - 1) + result.confidence) / 
        variantResults.successes;
      variantResults.avgProcessingTime = 
        (variantResults.avgProcessingTime * (variantResults.successes - 1) + result.processingTime) / 
        variantResults.successes;
    }
  }
  
  getTestResults(testName) {
    const test = this.tests.get(testName);
    if (!test) return null;
    
    const { variantA, variantB } = test.results;
    
    return {
      testName,
      duration: Date.now() - test.startTime,
      variantA: {
        ...variantA,
        successRate: variantA.requests > 0 ? variantA.successes / variantA.requests : 0
      },
      variantB: {
        ...variantB,
        successRate: variantB.requests > 0 ? variantB.successes / variantB.requests : 0
      },
      winner: this.determineWinner(variantA, variantB)
    };
  }
  
  determineWinner(variantA, variantB) {
    if (variantA.requests === 0 || variantB.requests === 0) {
      return 'insufficient_data';
    }
    
    const aSuccessRate = variantA.successes / variantA.requests;
    const bSuccessRate = variantB.successes / variantB.requests;
    
    if (Math.abs(aSuccessRate - bSuccessRate) < 0.05) {
      return 'tie';
    }
    
    return aSuccessRate > bSuccessRate ? 'variantA' : 'variantB';
  }
  
  hashString(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    return Math.abs(hash) / 2147483647; // Normalize to 0-1
  }
}
```

## Model Monitoring and Drift Detection

### Performance Monitoring

```javascript
class ModelMonitor {
  constructor() {
    this.metrics = new Map();
    this.alerts = [];
  }
  
  recordMetric(modelName, metricType, value, metadata = {}) {
    const key = `${modelName}_${metricType}`;
    
    if (!this.metrics.has(key)) {
      this.metrics.set(key, {
        values: [],
        timestamps: [],
        metadata: []
      });
    }
    
    const metric = this.metrics.get(key);
    metric.values.push(value);
    metric.timestamps.push(Date.now());
    metric.metadata.push(metadata);
    
    // Keep only last 1000 values
    if (metric.values.length > 1000) {
      metric.values.shift();
      metric.timestamps.shift();
      metric.metadata.shift();
    }
    
    // Check for anomalies
    this.checkForAnomalies(modelName, metricType, value, metric);
  }
  
  checkForAnomalies(modelName, metricType, value, metric) {
    if (metric.values.length < 10) return; // Need minimum data points
    
    const recent = metric.values.slice(-10);
    const historical = metric.values.slice(-100, -10);
    
    if (historical.length === 0) return;
    
    const recentAvg = recent.reduce((sum, val) => sum + val, 0) / recent.length;
    const historicalAvg = historical.reduce((sum, val) => sum + val, 0) / historical.length;
    const historicalStd = this.calculateStandardDeviation(historical, historicalAvg);
    
    // Check for significant deviation
    const deviation = Math.abs(recentAvg - historicalAvg);
    const threshold = historicalStd * 2; // 2 standard deviations
    
    if (deviation > threshold) {
      this.createAlert(modelName, metricType, {
        current: recentAvg,
        historical: historicalAvg,
        deviation,
        threshold,
        severity: deviation > historicalStd * 3 ? 'high' : 'medium'
      });
    }
  }
  
  calculateStandardDeviation(values, mean) {
    const variance = values.reduce((sum, val) => sum + Math.pow(val - mean, 2), 0) / values.length;
    return Math.sqrt(variance);
  }
  
  createAlert(modelName, metricType, details) {
    const alert = {
      id: Date.now().toString(),
      modelName,
      metricType,
      details,
      timestamp: Date.now(),
      acknowledged: false
    };
    
    this.alerts.push(alert);
    
    // Log alert
    console.warn(`Model Alert: ${modelName} - ${metricType}`, details);
    
    return alert;
  }
  
  getModelHealth(modelName) {
    const modelMetrics = Array.from(this.metrics.keys())
      .filter(key => key.startsWith(modelName))
      .map(key => {
        const metric = this.metrics.get(key);
        const metricType = key.split('_').slice(1).join('_');
        
        return {
          type: metricType,
          current: metric.values[metric.values.length - 1],
          average: metric.values.reduce((sum, val) => sum + val, 0) / metric.values.length,
          trend: this.calculateTrend(metric.values),
          dataPoints: metric.values.length
        };
      });
    
    return {
      modelName,
      metrics: modelMetrics,
      health: this.assessHealth(modelMetrics)
    };
  }
  
  calculateTrend(values) {
    if (values.length < 2) return 'insufficient_data';
    
    const first = values.slice(0, Math.floor(values.length / 2));
    const second = values.slice(Math.floor(values.length / 2));
    
    const firstAvg = first.reduce((sum, val) => sum + val, 0) / first.length;
    const secondAvg = second.reduce((sum, val) => sum + val, 0) / second.length;
    
    const change = (secondAvg - firstAvg) / firstAvg;
    
    if (change > 0.1) return 'increasing';
    if (change < -0.1) return 'decreasing';
    return 'stable';
  }
  
  assessHealth(metrics) {
    const issues = metrics.filter(metric => {
      // Define health criteria
      if (metric.type === 'accuracy' && metric.current < 0.8) return true;
      if (metric.type === 'processing_time' && metric.current > 2000) return true;
      if (metric.type === 'confidence' && metric.current < 0.7) return true;
      return false;
    });
    
    if (issues.length === 0) return 'healthy';
    if (issues.length <= 2) return 'warning';
    return 'critical';
  }
}
```

## Best Practices

### 1. Model Selection Guidelines

```javascript
class ModelSelectionGuidelines {
  static selectModelForUseCase(useCase, requirements) {
    const guidelines = {
      highAccuracy: {
        models: ['ml_bert', 'custom_neural_net'],
        description: 'Use for critical business decisions requiring maximum accuracy',
        tradeoffs: 'Higher processing time and cost'
      },
      
      fastProcessing: {
        models: ['ml_distilbert', 'keyword_matching'],
        description: 'Use for real-time applications requiring fast response',
        tradeoffs: 'Slightly lower accuracy'
      },
      
      costOptimized: {
        models: ['keyword_matching', 'similarity_analysis'],
        description: 'Use for high-volume, cost-sensitive applications',
        tradeoffs: 'Lower accuracy, limited semantic understanding'
      },
      
      balanced: {
        models: ['ml_distilbert', 'ml_bert'],
        description: 'Use for general-purpose applications balancing accuracy and speed',
        tradeoffs: 'Moderate processing time and cost'
      }
    };
    
    return guidelines[useCase] || guidelines.balanced;
  }
  
  static getModelRecommendation(businessData, requirements) {
    const { accuracy, speed, cost, volume } = requirements;
    
    if (accuracy > 0.95 && speed < 100) {
      return 'custom_neural_net'; // Specialized for speed + accuracy
    } else if (accuracy > 0.95) {
      return 'ml_bert'; // Maximum accuracy
    } else if (speed < 100) {
      return 'ml_distilbert'; // Fast processing
    } else if (cost < 0.01 && volume > 10000) {
      return 'keyword_matching'; // Cost-optimized
    } else {
      return 'ml_distilbert'; // Balanced default
    }
  }
}
```

### 2. Error Handling and Fallbacks

```javascript
class MLModelErrorHandler {
  constructor() {
    this.fallbackChain = [
      'ml_bert',
      'ml_distilbert', 
      'custom_neural_net',
      'keyword_matching',
      'similarity_analysis'
    ];
  }
  
  async processWithFallback(businessData, primaryModel) {
    const models = [primaryModel, ...this.fallbackChain.filter(m => m !== primaryModel)];
    
    for (const model of models) {
      try {
        const result = await this.tryModel(businessData, model);
        if (result.success) {
          return {
            ...result,
            model_used: model,
            fallback_used: model !== primaryModel
          };
        }
      } catch (error) {
        console.warn(`Model ${model} failed:`, error.message);
        continue;
      }
    }
    
    throw new Error('All models failed');
  }
  
  async tryModel(businessData, model) {
    const request = {
      ...businessData,
      options: {
        strategies: [model],
        include_ml_explainability: true
      }
    };
    
    const response = await fetch('https://api.kyb-platform.com/v2/classify', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(request)
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    return await response.json();
  }
}
```

### 3. Performance Optimization

```javascript
class MLPerformanceOptimizer {
  constructor() {
    this.cache = new Map();
    this.batchQueue = [];
    this.batchProcessor = null;
  }
  
  // Implement request deduplication
  async deduplicateRequest(businessData) {
    const key = this.generateCacheKey(businessData);
    
    if (this.cache.has(key)) {
      const cached = this.cache.get(key);
      if (Date.now() - cached.timestamp < 300000) { // 5 minutes
        return {
          ...cached.result,
          from_cache: true
        };
      }
    }
    
    return null;
  }
  
  // Implement batch processing
  async queueForBatch(businessData) {
    return new Promise((resolve, reject) => {
      this.batchQueue.push({
        data: businessData,
        resolve,
        reject,
        timestamp: Date.now()
      });
      
      // Process batch when it reaches optimal size
      if (this.batchQueue.length >= 10) {
        this.processBatch();
      }
    });
  }
  
  async processBatch() {
    if (this.batchQueue.length === 0) return;
    
    const batch = this.batchQueue.splice(0, 10);
    
    try {
      const batchRequest = {
        businesses: batch.map(item => item.data),
        options: {
          strategies: ['ml_bert', 'ml_distilbert'],
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
        result.data.classifications.forEach((classification, index) => {
          batch[index].resolve(classification);
        });
      } else {
        batch.forEach(item => {
          item.reject(new Error('Batch processing failed'));
        });
      }
    } catch (error) {
      batch.forEach(item => {
        item.reject(error);
      });
    }
  }
  
  generateCacheKey(businessData) {
    return btoa(JSON.stringify({
      name: businessData.business_name,
      description: businessData.description,
      website: businessData.website
    }));
  }
}
```

## Conclusion

This ML model integration guide provides comprehensive documentation for working with the machine learning capabilities of the KYB Platform. By following these patterns and best practices, you can:

- Select the most appropriate models for your use case
- Optimize performance through caching and batch processing
- Understand model decisions through explainability features
- Monitor model performance and detect drift
- Implement robust error handling and fallback strategies
- Conduct A/B testing to compare model performance

For additional support or questions about specific ML model scenarios, please refer to the API reference documentation or contact our support team.
