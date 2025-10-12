# Experiment Framework Documentation

## Overview

The Experiment Framework provides comprehensive A/B testing capabilities for ML models in the Risk Assessment Service. It enables systematic comparison of different models, hyperparameters, feature sets, and industry-specific configurations to optimize model performance and ensure reliable deployments.

## Architecture

### Core Components

1. **ABTestFramework**: Core framework for managing experiments
2. **ExperimentManager**: High-level experiment creation and management
3. **ExperimentHandlers**: API handlers for experiment operations
4. **ExperimentRoutes**: HTTP routes for experiment endpoints

### Key Features

- **Model Comparison**: Compare different ML models (XGBoost, LSTM, Random Forest)
- **Hyperparameter Tuning**: Test different hyperparameter configurations
- **Feature Testing**: Evaluate different feature sets and combinations
- **Industry-Specific Testing**: Test models optimized for specific industries
- **Statistical Analysis**: Built-in statistical significance testing
- **Traffic Splitting**: Configurable traffic distribution across variants
- **Real-time Monitoring**: Track experiment performance in real-time

## API Endpoints

### Experiment Management

#### Create Model Comparison Experiment
```http
POST /experiments/model-comparison
Content-Type: application/json

{
  "name": "XGBoost vs LSTM Comparison",
  "models": [
    {
      "name": "XGBoost v1.0",
      "type": "xgboost",
      "version": "1.0",
      "parameters": {
        "n_estimators": 100,
        "max_depth": 6,
        "learning_rate": 0.1
      }
    },
    {
      "name": "LSTM v1.0",
      "type": "lstm",
      "version": "1.0",
      "parameters": {
        "sequence_length": 12,
        "hidden_units": 64,
        "dropout": 0.2
      }
    }
  ]
}
```

#### Create Hyperparameter Experiment
```http
POST /experiments/hyperparameter
Content-Type: application/json

{
  "name": "XGBoost Hyperparameter Tuning",
  "base_model": {
    "name": "XGBoost Base",
    "type": "xgboost",
    "version": "1.0",
    "parameters": {
      "learning_rate": 0.1,
      "subsample": 0.8
    }
  },
  "hyperparams": [
    {
      "name": "Conservative",
      "parameters": {
        "n_estimators": 50,
        "max_depth": 4
      }
    },
    {
      "name": "Aggressive",
      "parameters": {
        "n_estimators": 200,
        "max_depth": 10
      }
    },
    {
      "name": "Balanced",
      "parameters": {
        "n_estimators": 100,
        "max_depth": 6
      }
    }
  ]
}
```

#### Create Feature Experiment
```http
POST /experiments/feature
Content-Type: application/json

{
  "name": "Feature Set Optimization",
  "base_model": {
    "name": "LSTM Base",
    "type": "lstm",
    "version": "1.0",
    "parameters": {
      "sequence_length": 12,
      "hidden_units": 64
    }
  },
  "feature_sets": [
    {
      "name": "Basic Features",
      "features": ["business_name", "industry", "address"]
    },
    {
      "name": "Extended Features",
      "features": ["business_name", "industry", "address", "phone", "email", "website"]
    },
    {
      "name": "Full Features",
      "features": ["business_name", "industry", "address", "phone", "email", "website", "financial_data", "compliance_data"]
    }
  ]
}
```

#### Create Industry Experiment
```http
POST /experiments/industry
Content-Type: application/json

{
  "name": "Finance Industry Model Comparison",
  "industry": "finance",
  "models": [
    {
      "name": "General Model",
      "type": "xgboost",
      "version": "1.0",
      "parameters": {
        "n_estimators": 100
      }
    },
    {
      "name": "Finance Model",
      "type": "finance_specific",
      "version": "1.0",
      "parameters": {
        "finance_features": true,
        "regulatory_focus": true
      }
    }
  ]
}
```

### Experiment Monitoring

#### Get Experiment Status
```http
GET /experiments/{id}/status
```

Response:
```json
{
  "experiment_id": "exp_001",
  "status": "active",
  "start_date": "2024-01-15T10:00:00Z",
  "total_requests": 1500,
  "variant_statuses": {
    "variant_1": {
      "variant_id": "variant_1",
      "name": "XGBoost Model",
      "total_requests": 750,
      "successful_requests": 720,
      "failed_requests": 30,
      "average_latency": 150.5,
      "accuracy": 0.92,
      "error_rate": 0.04,
      "last_updated": "2024-01-15T12:00:00Z"
    },
    "variant_2": {
      "variant_id": "variant_2",
      "name": "LSTM Model",
      "total_requests": 750,
      "successful_requests": 700,
      "failed_requests": 50,
      "average_latency": 200.3,
      "accuracy": 0.88,
      "error_rate": 0.067,
      "last_updated": "2024-01-15T12:00:00Z"
    }
  }
}
```

#### List Experiments
```http
GET /experiments?status=active&limit=10&offset=0
```

Response:
```json
{
  "experiments": [
    {
      "id": "exp_001",
      "name": "XGBoost vs LSTM Comparison",
      "description": "Compare 2 different ML models for risk assessment",
      "status": "active",
      "start_date": "2024-01-15T10:00:00Z",
      "variants": {
        "variant_1": {
          "id": "variant_1",
          "name": "XGBoost Model",
          "model_type": "xgboost",
          "is_control": true,
          "traffic": 0.5
        },
        "variant_2": {
          "id": "variant_2",
          "name": "LSTM Model",
          "model_type": "lstm",
          "is_control": false,
          "traffic": 0.5
        }
      }
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Experiment Control

#### Stop Experiment
```http
POST /experiments/{id}/stop
```

Response:
```json
{
  "experiment_id": "exp_001",
  "winner": "variant_1",
  "is_significant": true,
  "confidence_level": 0.95,
  "start_date": "2024-01-15T10:00:00Z",
  "end_date": "2024-01-15T14:00:00Z",
  "variant_results": {
    "variant_1": {
      "variant_id": "variant_1",
      "name": "XGBoost Model",
      "total_requests": 1500,
      "successful_requests": 1440,
      "failed_requests": 60,
      "average_latency": 150.5,
      "accuracy": 0.92,
      "error_rate": 0.04,
      "conversion_rate": 0.85,
      "last_updated": "2024-01-15T14:00:00Z"
    },
    "variant_2": {
      "variant_id": "variant_2",
      "name": "LSTM Model",
      "model_type": "lstm",
      "total_requests": 1500,
      "successful_requests": 1380,
      "failed_requests": 120,
      "average_latency": 200.3,
      "accuracy": 0.88,
      "error_rate": 0.08,
      "conversion_rate": 0.82,
      "last_updated": "2024-01-15T14:00:00Z"
    }
  },
  "statistical_significance": {
    "p_value": 0.03,
    "confidence_level": 0.95,
    "is_significant": true,
    "effect_size": 0.15,
    "power": 0.85
  },
  "recommendations": [
    "XGBoost model shows better performance",
    "Consider deploying XGBoost model to production",
    "LSTM model has higher latency, optimize if needed"
  ]
}
```

#### Get Experiment Results
```http
GET /experiments/{id}/results
```

## Usage Examples

### 1. Model Comparison Experiment

```go
// Create a model comparison experiment
models := []testing.ModelConfig{
    {
        Name:    "XGBoost v1.0",
        Type:    "xgboost",
        Version: "1.0",
        Parameters: map[string]interface{}{
            "n_estimators": 100,
            "max_depth":    6,
        },
    },
    {
        Name:    "LSTM v1.0",
        Type:    "lstm",
        Version: "1.0",
        Parameters: map[string]interface{}{
            "sequence_length": 12,
            "hidden_units":    64,
        },
    },
}

experiment, err := experimentManager.CreateModelComparisonExperiment(
    context.Background(),
    "Model Comparison Test",
    models,
)
```

### 2. Hyperparameter Tuning Experiment

```go
// Create a hyperparameter tuning experiment
baseModel := testing.ModelConfig{
    Name:    "XGBoost Base",
    Type:    "xgboost",
    Version: "1.0",
    Parameters: map[string]interface{}{
        "learning_rate": 0.1,
        "subsample":     0.8,
    },
}

hyperparams := []testing.HyperparameterConfig{
    {
        Name: "Conservative",
        Parameters: map[string]interface{}{
            "n_estimators": 50,
            "max_depth":    4,
        },
    },
    {
        Name: "Aggressive",
        Parameters: map[string]interface{}{
            "n_estimators": 200,
            "max_depth":    10,
        },
    },
}

experiment, err := experimentManager.CreateHyperparameterExperiment(
    context.Background(),
    "Hyperparameter Tuning",
    baseModel,
    hyperparams,
)
```

### 3. Feature Set Experiment

```go
// Create a feature set experiment
baseModel := testing.ModelConfig{
    Name:    "LSTM Base",
    Type:    "lstm",
    Version: "1.0",
    Parameters: map[string]interface{}{
        "sequence_length": 12,
        "hidden_units":    64,
    },
}

featureSets := []testing.FeatureSetConfig{
    {
        Name:     "Basic Features",
        Features: []string{"business_name", "industry", "address"},
    },
    {
        Name:     "Extended Features",
        Features: []string{"business_name", "industry", "address", "phone", "email", "website"},
    },
}

experiment, err := experimentManager.CreateFeatureExperiment(
    context.Background(),
    "Feature Set Testing",
    baseModel,
    featureSets,
)
```

### 4. Industry-Specific Experiment

```go
// Create an industry-specific experiment
models := []testing.ModelConfig{
    {
        Name:    "General Model",
        Type:    "xgboost",
        Version: "1.0",
        Parameters: map[string]interface{}{
            "n_estimators": 100,
        },
    },
    {
        Name:    "Finance Model",
        Type:    "finance_specific",
        Version: "1.0",
        Parameters: map[string]interface{}{
            "finance_features": true,
        },
    },
}

experiment, err := experimentManager.CreateIndustryExperiment(
    context.Background(),
    "Finance Industry Test",
    "finance",
    models,
)
```

## Configuration

### Experiment Settings

- **MinSampleSize**: Minimum number of samples required for statistical significance
- **ConfidenceLevel**: Statistical confidence level (default: 0.95)
- **TrafficSplit**: Distribution of traffic across variants (default: equal split)
- **Metrics**: Performance metrics to track (accuracy, latency, confidence, error_rate)

### Traffic Splitting

The framework supports configurable traffic splitting:

```go
trafficSplit := map[string]float64{
    "control":    0.5,  // 50% traffic to control
    "variant_a":  0.3,  // 30% traffic to variant A
    "variant_b":  0.2,  // 20% traffic to variant B
}
```

### Statistical Analysis

The framework automatically performs statistical analysis:

- **P-value calculation**: Determines statistical significance
- **Effect size**: Measures the magnitude of differences
- **Power analysis**: Ensures adequate sample size
- **Confidence intervals**: Provides uncertainty bounds

## Best Practices

### 1. Experiment Design

- **Clear Hypothesis**: Define what you're testing and why
- **Control Group**: Always include a control variant
- **Sample Size**: Ensure adequate sample size for statistical power
- **Duration**: Run experiments long enough to capture seasonal effects

### 2. Model Selection

- **Baseline Model**: Use a proven model as the control
- **Incremental Changes**: Test one change at a time
- **Industry Context**: Consider industry-specific requirements
- **Performance Metrics**: Define success criteria upfront

### 3. Monitoring

- **Real-time Tracking**: Monitor experiment performance continuously
- **Early Stopping**: Stop experiments if clear winner emerges
- **Error Handling**: Monitor and alert on errors
- **Data Quality**: Ensure data quality throughout the experiment

### 4. Analysis

- **Statistical Significance**: Wait for statistical significance before concluding
- **Multiple Metrics**: Consider multiple performance metrics
- **Business Impact**: Evaluate business impact, not just technical metrics
- **Documentation**: Document findings and decisions

## Error Handling

The framework includes comprehensive error handling:

- **Validation Errors**: Input validation with clear error messages
- **Model Errors**: Graceful handling of model failures
- **Network Errors**: Retry logic for external dependencies
- **Data Errors**: Validation of experiment data

## Performance Considerations

- **Traffic Splitting**: Efficient traffic routing with minimal overhead
- **Model Loading**: Lazy loading of models to reduce memory usage
- **Caching**: Result caching for improved performance
- **Monitoring**: Lightweight monitoring with minimal impact

## Security

- **Input Validation**: Comprehensive input validation
- **Access Control**: Role-based access to experiment management
- **Data Privacy**: Protection of sensitive experiment data
- **Audit Logging**: Complete audit trail of experiment changes

## Future Enhancements

- **Multi-armed Bandits**: Dynamic traffic allocation based on performance
- **Bayesian Analysis**: Advanced statistical analysis methods
- **Automated Deployment**: Automatic promotion of winning models
- **Integration**: Integration with CI/CD pipelines
- **Dashboard**: Web-based experiment management dashboard
