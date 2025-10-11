# ML Model Validation System

## Overview

The ML Model Validation System provides comprehensive validation capabilities for machine learning models used in risk assessment. It includes cross-validation, historical data generation, performance metrics, and actionable recommendations.

## Features

### 🔬 Cross-Validation Framework
- **K-Fold Cross-Validation**: Configurable number of folds (default: 5)
- **Stratified Sampling**: Ensures balanced representation across folds
- **Confidence Intervals**: Statistical confidence intervals for all metrics
- **Parallel Processing**: Optional parallel fold processing for faster validation

### 📊 Performance Metrics
- **Accuracy**: Overall prediction accuracy
- **Precision**: True positive rate
- **Recall**: Sensitivity/true positive rate
- **F1-Score**: Harmonic mean of precision and recall
- **AUC**: Area Under the Curve (ROC)
- **Log Loss**: Logarithmic loss for probability predictions
- **MCC**: Matthews Correlation Coefficient
- **Specificity**: True negative rate

### 📈 Historical Data Generation
- **Realistic Business Data**: Industry, geographic, and size distributions
- **Temporal Patterns**: Seasonal effects and trends
- **Risk Factors**: 20+ realistic risk factors including:
  - Financial health (revenue growth, profit margin, debt-to-equity)
  - Operational factors (employee count, market share, customer concentration)
  - Market factors (volatility, competition, barriers to entry)
  - Technology factors (digital maturity, cybersecurity, innovation)
  - Environmental factors (climate risk, sustainability, ESG rating)
- **External Events**: Economic crises, regulatory changes, market disruptions
- **Data Quality Metrics**: Completeness, consistency, accuracy, timeliness

### 🎯 Validation Reports
- **Comprehensive Analysis**: Detailed performance breakdown
- **Model Comparison**: Side-by-side comparison of multiple models
- **Actionable Recommendations**: Prioritized improvement suggestions
- **Risk Assessment**: Model deployment risk evaluation
- **JSON Export**: Machine-readable validation reports

## Usage

### Command Line Interface

#### Basic Validation
```bash
# Run standard validation (5-fold, 1000 samples, 1 year data)
make validate-ml

# Quick validation (3-fold, 100 samples, 30 days data)
make validate-ml-quick

# Comprehensive validation (10-fold, 5000 samples, 2 years data)
make validate-ml-comprehensive
```

#### Custom Validation
```bash
# Custom parameters
go run cmd/validate_model.go \
  -k 7 \
  -samples 2000 \
  -time-range 180d \
  -confidence 0.99 \
  -output my_validation_report.json \
  -verbose
```

#### Command Line Options
- `-k`: Number of folds for cross-validation (default: 5)
- `-samples`: Total number of samples to generate (default: 1000)
- `-time-range`: Time range for historical data (default: 365d)
- `-confidence`: Confidence level for intervals (default: 0.95)
- `-output`: Output file for validation report (JSON format)
- `-verbose`: Enable verbose logging
- `-parallel`: Enable parallel fold processing
- `-seed`: Random seed for reproducibility

### Programmatic Usage

```go
package main

import (
    "context"
    "log"
    
    "go.uber.org/zap"
    "kyb-platform/services/risk-assessment-service/internal/ml/validation"
    "kyb-platform/services/risk-assessment-service/internal/ml/models"
)

func main() {
    logger, _ := zap.NewProduction()
    
    // Create validation service
    validationService := validation.NewValidationService(logger)
    
    // Create model
    model := models.NewXGBoostModel(logger)
    
    // Configure validation
    config := validation.ValidationConfig{
        CrossValidation: validation.CrossValidationConfig{
            KFolds:          5,
            ConfidenceLevel: 0.95,
            RandomSeed:      12345,
            ParallelFolds:   true,
            MaxConcurrency:  4,
        },
        DataGeneration: validation.DataGenerationConfig{
            TotalSamples:     1000,
            TimeRange:        365 * 24 * time.Hour,
            RiskCategories:   getDefaultRiskCategories(),
            IndustryWeights:  getDefaultIndustryWeights(),
            GeographicBias:   getDefaultGeographicBias(),
            SeasonalPatterns: true,
            TrendStrength:    0.02,
            NoiseLevel:       0.05,
        },
        OutputFormat: "json",
        SaveResults:  true,
        ResultsPath:  "validation_report.json",
    }
    
    // Run validation
    ctx := context.Background()
    report, err := validationService.ValidateModel(ctx, model, config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use results
    log.Printf("Overall Score: %.3f", report.Summary.OverallScore)
    log.Printf("Recommendation: %s", report.Summary.Recommendation)
}
```

## Configuration

### Cross-Validation Configuration

```go
type CrossValidationConfig struct {
    KFolds           int     `json:"k_folds"`           // Number of folds (default: 5)
    ConfidenceLevel  float64 `json:"confidence_level"`  // Confidence level (default: 0.95)
    RandomSeed       int64   `json:"random_seed"`       // Random seed for reproducibility
    ParallelFolds    bool    `json:"parallel_folds"`    // Enable parallel processing
    MaxConcurrency   int     `json:"max_concurrency"`   // Max concurrent folds
}
```

### Data Generation Configuration

```go
type DataGenerationConfig struct {
    TotalSamples     int                    `json:"total_samples"`     // Number of samples
    TimeRange        time.Duration          `json:"time_range"`        // Time range
    RiskCategories   []RiskCategory         `json:"risk_categories"`   // Risk categories
    IndustryWeights  map[string]float64     `json:"industry_weights"`  // Industry distribution
    GeographicBias   map[string]float64     `json:"geographic_bias"`   // Geographic distribution
    SeasonalPatterns bool                   `json:"seasonal_patterns"` // Enable seasonal effects
    TrendStrength    float64                `json:"trend_strength"`    // Trend strength
    NoiseLevel       float64                `json:"noise_level"`       // Noise level
}
```

### Risk Categories

```go
type RiskCategory struct {
    Name            string                 `json:"name"`
    BaseRisk        float64                `json:"base_risk"`
    Volatility      float64                `json:"volatility"`
    IndustryBias    map[string]float64     `json:"industry_bias"`
    GeographicBias  map[string]float64     `json:"geographic_bias"`
    SizeBias        map[string]float64     `json:"size_bias"`
    AgeBias         map[string]float64     `json:"age_bias"`
}
```

## Output Format

### Validation Report Structure

```json
{
  "summary": {
    "overall_score": 0.847,
    "accuracy_score": 0.823,
    "reliability_score": 0.891,
    "performance_score": 0.827,
    "recommendation": "Good model performance. Consider minor improvements before production.",
    "risk_level": "Low-Medium",
    "confidence_level": 0.85
  },
  "cross_validation": {
    "model_name": "XGBoost",
    "k_folds": 5,
    "total_samples": 1000,
    "fold_results": [...],
    "overall_metrics": {
      "mean_accuracy": 0.823,
      "std_accuracy": 0.021,
      "mean_precision": 0.815,
      "std_precision": 0.025,
      "mean_recall": 0.831,
      "std_recall": 0.019,
      "mean_f1_score": 0.823,
      "std_f1_score": 0.020,
      "mean_auc": 0.891,
      "std_auc": 0.015,
      "mean_log_loss": 0.342,
      "std_log_loss": 0.028
    },
    "confidence_interval": {
      "accuracy": {"lower": 0.801, "upper": 0.845},
      "precision": {"lower": 0.790, "upper": 0.840},
      "recall": {"lower": 0.812, "upper": 0.850},
      "f1_score": {"lower": 0.803, "upper": 0.843},
      "auc": {"lower": 0.876, "upper": 0.906},
      "confidence_level": 0.95
    },
    "validation_time": "2m34s",
    "timestamp": "2025-01-09T13:44:30Z"
  },
  "historical_data": {
    "total_samples": 1000,
    "time_range": "8760h0m0s",
    "industries": {
      "Technology": 200,
      "Finance": 150,
      "Healthcare": 120,
      ...
    },
    "countries": {
      "United States": 350,
      "Canada": 80,
      "United Kingdom": 70,
      ...
    },
    "business_sizes": {
      "Small": 400,
      "Medium": 300,
      "Large": 200,
      ...
    },
    "risk_distribution": {
      "Low": 200,
      "Medium-Low": 300,
      "Medium": 250,
      "Medium-High": 150,
      "High": 100
    },
    "data_quality": {
      "completeness": 0.95,
      "consistency": 0.87,
      "accuracy": 0.85,
      "timeliness": 0.92,
      "relevance": 0.80,
      "overall_quality": 0.88
    }
  },
  "model_comparison": [
    {
      "model_name": "XGBoost",
      "metrics": {
        "accuracy": 0.823,
        "precision": 0.815,
        "recall": 0.831,
        "f1_score": 0.823,
        "auc": 0.891,
        "log_loss": 0.342
      },
      "overall_score": 0.847,
      "rank": 1,
      "strengths": ["High accuracy", "Consistent performance", "Good F1 score"],
      "weaknesses": [],
      "use_case": "General risk assessment"
    }
  ],
  "recommendations": [
    {
      "type": "Model Improvement",
      "priority": "Medium",
      "title": "Improve Model Consistency",
      "description": "Model shows moderate variance across folds. Consider regularization or ensemble methods.",
      "impact": "Medium",
      "effort": "Medium",
      "timeline": "1-2 weeks",
      "actions": [
        "Add regularization techniques",
        "Implement ensemble methods",
        "Cross-validation with more folds"
      ]
    }
  ],
  "generated_at": "2025-01-09T13:44:30Z",
  "validation_time": "2m34s",
  "configuration": {...}
}
```

## Testing

### Run Tests
```bash
# Run all validation tests
make test-ml-validation

# Run specific test
go test -v ./internal/ml/validation/... -run TestCrossValidator

# Run with coverage
go test -v -cover ./internal/ml/validation/...
```

### Test Coverage
The validation system includes comprehensive tests covering:
- Cross-validation with various configurations
- Historical data generation with different parameters
- Edge cases and error handling
- Performance metrics calculations
- Confidence interval calculations
- Multi-model comparison

## Performance Considerations

### Optimization Tips
1. **Parallel Processing**: Enable `ParallelFolds` for faster validation
2. **Sample Size**: Balance between accuracy and speed
3. **K-Folds**: More folds = better estimates but slower execution
4. **Memory Usage**: Large sample sizes require more memory

### Performance Benchmarks
- **100 samples, 3-fold**: ~5 seconds
- **1000 samples, 5-fold**: ~30 seconds
- **5000 samples, 10-fold**: ~3 minutes
- **Parallel processing**: 2-3x speedup

## Best Practices

### Model Validation
1. **Use Multiple Metrics**: Don't rely on accuracy alone
2. **Check Confidence Intervals**: Ensure statistical significance
3. **Validate on Historical Data**: Use realistic data distributions
4. **Consider Business Context**: Align metrics with business objectives

### Data Quality
1. **Monitor Data Quality**: Track completeness, consistency, accuracy
2. **Handle Missing Data**: Implement appropriate imputation strategies
3. **Validate Data Sources**: Ensure external data is reliable
4. **Regular Updates**: Keep historical data current

### Deployment Readiness
1. **Performance Thresholds**: Set minimum acceptable performance levels
2. **Risk Assessment**: Evaluate deployment risk based on validation results
3. **Monitoring Plan**: Implement ongoing model performance monitoring
4. **Rollback Strategy**: Prepare for model rollback if performance degrades

## Troubleshooting

### Common Issues

#### Low Accuracy
- **Cause**: Insufficient training data or poor feature engineering
- **Solution**: Increase sample size, improve feature selection, try different algorithms

#### High Variance
- **Cause**: Model instability or insufficient regularization
- **Solution**: Add regularization, use ensemble methods, increase training data

#### Poor Data Quality
- **Cause**: Missing or inconsistent data
- **Solution**: Improve data collection, implement data validation, use imputation

#### Slow Performance
- **Cause**: Large datasets or complex models
- **Solution**: Enable parallel processing, reduce sample size, optimize algorithms

### Debug Mode
```bash
# Enable verbose logging for debugging
go run cmd/validate_model.go -verbose -k 3 -samples 100
```

## Integration

### With Existing Models
The validation system integrates with any model implementing the `ModelValidator` interface:

```go
type ModelValidator interface {
    Train(ctx context.Context, features [][]float64, labels []float64) error
    Predict(ctx context.Context, features [][]float64) ([]float64, error)
    PredictProba(ctx context.Context, features [][]float64) ([][]float64, error)
    GetName() string
}
```

### With CI/CD Pipeline
```yaml
# Example GitHub Actions workflow
- name: Validate ML Model
  run: |
    make validate-ml
    if [ $? -ne 0 ]; then
      echo "Model validation failed"
      exit 1
    fi
```

### With Monitoring Systems
The validation reports can be integrated with monitoring systems for automated model performance tracking and alerting.

## Future Enhancements

### Planned Features
1. **Automated Hyperparameter Tuning**: Integration with optimization libraries
2. **Real-time Validation**: Continuous model validation during production
3. **A/B Testing Framework**: Compare model versions in production
4. **Advanced Metrics**: Additional performance metrics and visualizations
5. **Model Interpretability**: SHAP values and feature importance analysis
6. **Automated Retraining**: Trigger retraining based on performance degradation

### Contributing
To contribute to the ML validation system:
1. Follow the existing code structure and patterns
2. Add comprehensive tests for new features
3. Update documentation for any changes
4. Ensure backward compatibility
5. Follow Go best practices and conventions
