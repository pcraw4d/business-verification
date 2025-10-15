# Advanced Predictions Tutorial

## Overview

This tutorial covers advanced prediction capabilities of the Risk Assessment Service, including multi-horizon predictions, scenario analysis, and explainable AI features. You'll learn how to leverage machine learning models for predictive risk analytics.

## Table of Contents

1. [Understanding Advanced Predictions](#understanding-advanced-predictions)
2. [Multi-Horizon Predictions](#multi-horizon-predictions)
3. [Scenario Analysis](#scenario-analysis)
4. [Explainable AI (SHAP)](#explainable-ai-shap)
5. [Model Comparison and A/B Testing](#model-comparison-and-ab-testing)
6. [Real-World Use Cases](#real-world-use-cases)

## Understanding Advanced Predictions

### Prediction Horizons

Our service supports multiple prediction horizons:

- **Short-term (1-3 months)**: XGBoost model for immediate risk trends
- **Medium-term (3-6 months)**: Ensemble model combining multiple approaches
- **Long-term (6-12 months)**: LSTM model for time-series forecasting

### Model Types

1. **XGBoost Model**: Gradient boosting for short-term predictions
2. **LSTM Model**: Long Short-Term Memory for time-series forecasting
3. **Ensemble Model**: Combines multiple models for optimal performance

## Multi-Horizon Predictions

### Basic Multi-Horizon Prediction

<details>
<summary><strong>Go</strong></summary>

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client := kyb.NewClient(&kyb.Config{
        APIKey: "your_api_key_here",
    })
    
    request := &kyb.AdvancedPredictionRequest{
        BusinessName:    "FutureTech Corp",
        BusinessAddress: "123 Innovation Blvd, Future City, FC 12345",
        Industry:        "Technology",
        Country:         "US",
        PredictionHorizons: []int{1, 3, 6, 12}, // 1, 3, 6, and 12 months
        IncludeConfidenceIntervals: true,
        IncludeRiskFactors: true,
    }
    
    ctx := context.Background()
    prediction, err := client.PredictAdvanced(ctx, request)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Business ID: %s\n", prediction.BusinessID)
    fmt.Printf("Prediction ID: %s\n", prediction.ID)
    
    // Display predictions for each horizon
    for _, horizon := range prediction.Predictions {
        fmt.Printf("\n--- %d Month Prediction ---\n", horizon.Horizon)
        fmt.Printf("Risk Score: %.3f\n", horizon.RiskScore)
        fmt.Printf("Risk Level: %s\n", horizon.RiskLevel)
        fmt.Printf("Confidence: %.3f\n", horizon.Confidence)
        
        if horizon.ConfidenceInterval != nil {
            fmt.Printf("Confidence Interval: [%.3f, %.3f]\n", 
                horizon.ConfidenceInterval.Lower, horizon.ConfidenceInterval.Upper)
        }
        
        // Display trend
        if horizon.Trend != nil {
            fmt.Printf("Trend: %s (%.2f%% change)\n", 
                horizon.Trend.Direction, horizon.Trend.PercentageChange)
        }
    }
}
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
from kyb_risk_assessment import RiskAssessmentClient

client = RiskAssessmentClient(api_key="your_api_key_here")

request = {
    "business_name": "FutureTech Corp",
    "business_address": "123 Innovation Blvd, Future City, FC 12345",
    "industry": "Technology",
    "country": "US",
    "prediction_horizons": [1, 3, 6, 12],  # 1, 3, 6, and 12 months
    "include_confidence_intervals": True,
    "include_risk_factors": True
}

prediction = client.predict_advanced(request)

print(f"Business ID: {prediction.business_id}")
print(f"Prediction ID: {prediction.id}")

# Display predictions for each horizon
for horizon in prediction.predictions:
    print(f"\n--- {horizon.horizon} Month Prediction ---")
    print(f"Risk Score: {horizon.risk_score:.3f}")
    print(f"Risk Level: {horizon.risk_level}")
    print(f"Confidence: {horizon.confidence:.3f}")
    
    if horizon.confidence_interval:
        print(f"Confidence Interval: [{horizon.confidence_interval.lower:.3f}, {horizon.confidence_interval.upper:.3f}]")
    
    # Display trend
    if horizon.trend:
        print(f"Trend: {horizon.trend.direction} ({horizon.trend.percentage_change:.2f}% change)")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
const { RiskAssessmentClient } = require('@kyb-platform/risk-assessment');

const client = new RiskAssessmentClient({
    apiKey: 'your_api_key_here'
});

const request = {
    businessName: 'FutureTech Corp',
    businessAddress: '123 Innovation Blvd, Future City, FC 12345',
    industry: 'Technology',
    country: 'US',
    predictionHorizons: [1, 3, 6, 12], // 1, 3, 6, and 12 months
    includeConfidenceIntervals: true,
    includeRiskFactors: true
};

async function predictAdvanced() {
    try {
        const prediction = await client.predictAdvanced(request);
        
        console.log(`Business ID: ${prediction.businessId}`);
        console.log(`Prediction ID: ${prediction.id}`);
        
        // Display predictions for each horizon
        prediction.predictions.forEach(horizon => {
            console.log(`\n--- ${horizon.horizon} Month Prediction ---`);
            console.log(`Risk Score: ${horizon.riskScore.toFixed(3)}`);
            console.log(`Risk Level: ${horizon.riskLevel}`);
            console.log(`Confidence: ${horizon.confidence.toFixed(3)}`);
            
            if (horizon.confidenceInterval) {
                console.log(`Confidence Interval: [${horizon.confidenceInterval.lower.toFixed(3)}, ${horizon.confidenceInterval.upper.toFixed(3)}]`);
            }
            
            // Display trend
            if (horizon.trend) {
                console.log(`Trend: ${horizon.trend.direction} (${horizon.trend.percentageChange.toFixed(2)}% change)`);
            }
        });
    } catch (error) {
        console.error('Error:', error.message);
    }
}

predictAdvanced();
```

</details>

### Custom Model Selection

<details>
<summary><strong>Go</strong></summary>

```go
// Use specific models for different horizons
request := &kyb.AdvancedPredictionRequest{
    BusinessName:    "CustomTech Inc",
    BusinessAddress: "456 Custom St, Model City, MC 54321",
    Industry:        "Technology",
    Country:         "US",
    PredictionHorizons: []int{1, 6, 12},
    ModelPreferences: map[int]string{
        1:  "xgboost",  // Use XGBoost for 1-month prediction
        6:  "ensemble", // Use Ensemble for 6-month prediction
        12: "lstm",     // Use LSTM for 12-month prediction
    },
    IncludeModelMetadata: true,
}

prediction, err := client.PredictAdvanced(ctx, request)
if err != nil {
    log.Fatal(err)
}

// Display model information
for _, horizon := range prediction.Predictions {
    fmt.Printf("\n--- %d Month Prediction ---\n", horizon.Horizon)
    fmt.Printf("Model Used: %s\n", horizon.ModelUsed)
    fmt.Printf("Model Version: %s\n", horizon.ModelVersion)
    fmt.Printf("Training Date: %s\n", horizon.TrainingDate)
    fmt.Printf("Risk Score: %.3f\n", horizon.RiskScore)
}
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
# Use specific models for different horizons
request = {
    "business_name": "CustomTech Inc",
    "business_address": "456 Custom St, Model City, MC 54321",
    "industry": "Technology",
    "country": "US",
    "prediction_horizons": [1, 6, 12],
    "model_preferences": {
        1: "xgboost",   # Use XGBoost for 1-month prediction
        6: "ensemble",  # Use Ensemble for 6-month prediction
        12: "lstm"      # Use LSTM for 12-month prediction
    },
    "include_model_metadata": True
}

prediction = client.predict_advanced(request)

# Display model information
for horizon in prediction.predictions:
    print(f"\n--- {horizon.horizon} Month Prediction ---")
    print(f"Model Used: {horizon.model_used}")
    print(f"Model Version: {horizon.model_version}")
    print(f"Training Date: {horizon.training_date}")
    print(f"Risk Score: {horizon.risk_score:.3f}")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
// Use specific models for different horizons
const request = {
    businessName: 'CustomTech Inc',
    businessAddress: '456 Custom St, Model City, MC 54321',
    industry: 'Technology',
    country: 'US',
    predictionHorizons: [1, 6, 12],
    modelPreferences: {
        1: 'xgboost',   // Use XGBoost for 1-month prediction
        6: 'ensemble',  // Use Ensemble for 6-month prediction
        12: 'lstm'      // Use LSTM for 12-month prediction
    },
    includeModelMetadata: true
};

async function predictWithCustomModels() {
    try {
        const prediction = await client.predictAdvanced(request);
        
        // Display model information
        prediction.predictions.forEach(horizon => {
            console.log(`\n--- ${horizon.horizon} Month Prediction ---`);
            console.log(`Model Used: ${horizon.modelUsed}`);
            console.log(`Model Version: ${horizon.modelVersion}`);
            console.log(`Training Date: ${horizon.trainingDate}`);
            console.log(`Risk Score: ${horizon.riskScore.toFixed(3)}`);
        });
    } catch (error) {
        console.error('Error:', error.message);
    }
}

predictWithCustomModels();
```

</details>

## Scenario Analysis

### Monte Carlo Simulation

<details>
<summary><strong>Go</strong></summary>

```go
// Perform scenario analysis with Monte Carlo simulation
scenarioRequest := &kyb.ScenarioAnalysisRequest{
    BusinessName:    "ScenarioTech Ltd",
    BusinessAddress: "789 Scenario Ave, Analysis City, AC 67890",
    Industry:        "Technology",
    Country:         "US",
    Scenarios: []kyb.Scenario{
        {
            Name: "Optimistic",
            Description: "Best-case scenario with favorable market conditions",
            Parameters: map[string]interface{}{
                "market_growth": 1.2,
                "competition_level": 0.8,
                "regulatory_environment": 0.9,
            },
        },
        {
            Name: "Pessimistic",
            Description: "Worst-case scenario with challenging conditions",
            Parameters: map[string]interface{}{
                "market_growth": 0.8,
                "competition_level": 1.3,
                "regulatory_environment": 1.1,
            },
        },
        {
            Name: "Realistic",
            Description: "Most likely scenario based on current trends",
            Parameters: map[string]interface{}{
                "market_growth": 1.0,
                "competition_level": 1.0,
                "regulatory_environment": 1.0,
            },
        },
    },
    MonteCarloSimulations: 1000,
    PredictionHorizon: 12,
}

scenarioResult, err := client.AnalyzeScenarios(ctx, scenarioRequest)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Scenario Analysis ID: %s\n", scenarioResult.ID)
fmt.Printf("Total Simulations: %d\n", scenarioResult.TotalSimulations)

for _, scenario := range scenarioResult.Scenarios {
    fmt.Printf("\n--- %s Scenario ---\n", scenario.Name)
    fmt.Printf("Average Risk Score: %.3f\n", scenario.AverageRiskScore)
    fmt.Printf("Risk Score Range: [%.3f, %.3f]\n", 
        scenario.MinRiskScore, scenario.MaxRiskScore)
    fmt.Printf("Probability of High Risk: %.2f%%\n", 
        scenario.HighRiskProbability * 100)
    
    // Display risk distribution
    fmt.Println("Risk Distribution:")
    for riskLevel, probability := range scenario.RiskDistribution {
        fmt.Printf("  %s: %.2f%%\n", riskLevel, probability * 100)
    }
}
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
# Perform scenario analysis with Monte Carlo simulation
scenario_request = {
    "business_name": "ScenarioTech Ltd",
    "business_address": "789 Scenario Ave, Analysis City, AC 67890",
    "industry": "Technology",
    "country": "US",
    "scenarios": [
        {
            "name": "Optimistic",
            "description": "Best-case scenario with favorable market conditions",
            "parameters": {
                "market_growth": 1.2,
                "competition_level": 0.8,
                "regulatory_environment": 0.9
            }
        },
        {
            "name": "Pessimistic",
            "description": "Worst-case scenario with challenging conditions",
            "parameters": {
                "market_growth": 0.8,
                "competition_level": 1.3,
                "regulatory_environment": 1.1
            }
        },
        {
            "name": "Realistic",
            "description": "Most likely scenario based on current trends",
            "parameters": {
                "market_growth": 1.0,
                "competition_level": 1.0,
                "regulatory_environment": 1.0
            }
        }
    ],
    "monte_carlo_simulations": 1000,
    "prediction_horizon": 12
}

scenario_result = client.analyze_scenarios(scenario_request)

print(f"Scenario Analysis ID: {scenario_result.id}")
print(f"Total Simulations: {scenario_result.total_simulations}")

for scenario in scenario_result.scenarios:
    print(f"\n--- {scenario.name} Scenario ---")
    print(f"Average Risk Score: {scenario.average_risk_score:.3f}")
    print(f"Risk Score Range: [{scenario.min_risk_score:.3f}, {scenario.max_risk_score:.3f}]")
    print(f"Probability of High Risk: {scenario.high_risk_probability * 100:.2f}%")
    
    # Display risk distribution
    print("Risk Distribution:")
    for risk_level, probability in scenario.risk_distribution.items():
        print(f"  {risk_level}: {probability * 100:.2f}%")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
// Perform scenario analysis with Monte Carlo simulation
const scenarioRequest = {
    businessName: 'ScenarioTech Ltd',
    businessAddress: '789 Scenario Ave, Analysis City, AC 67890',
    industry: 'Technology',
    country: 'US',
    scenarios: [
        {
            name: 'Optimistic',
            description: 'Best-case scenario with favorable market conditions',
            parameters: {
                market_growth: 1.2,
                competition_level: 0.8,
                regulatory_environment: 0.9
            }
        },
        {
            name: 'Pessimistic',
            description: 'Worst-case scenario with challenging conditions',
            parameters: {
                market_growth: 0.8,
                competition_level: 1.3,
                regulatory_environment: 1.1
            }
        },
        {
            name: 'Realistic',
            description: 'Most likely scenario based on current trends',
            parameters: {
                market_growth: 1.0,
                competition_level: 1.0,
                regulatory_environment: 1.0
            }
        }
    ],
    monteCarloSimulations: 1000,
    predictionHorizon: 12
};

async function analyzeScenarios() {
    try {
        const scenarioResult = await client.analyzeScenarios(scenarioRequest);
        
        console.log(`Scenario Analysis ID: ${scenarioResult.id}`);
        console.log(`Total Simulations: ${scenarioResult.totalSimulations}`);
        
        scenarioResult.scenarios.forEach(scenario => {
            console.log(`\n--- ${scenario.name} Scenario ---`);
            console.log(`Average Risk Score: ${scenario.averageRiskScore.toFixed(3)}`);
            console.log(`Risk Score Range: [${scenario.minRiskScore.toFixed(3)}, ${scenario.maxRiskScore.toFixed(3)}]`);
            console.log(`Probability of High Risk: ${(scenario.highRiskProbability * 100).toFixed(2)}%`);
            
            // Display risk distribution
            console.log('Risk Distribution:');
            Object.entries(scenario.riskDistribution).forEach(([riskLevel, probability]) => {
                console.log(`  ${riskLevel}: ${(probability * 100).toFixed(2)}%`);
            });
        });
    } catch (error) {
        console.error('Error:', error.message);
    }
}

analyzeScenarios();
```

</details>

### Stress Testing

<details>
<summary><strong>Go</strong></summary>

```go
// Perform stress testing
stressRequest := &kyb.ScenarioAnalysisRequest{
    BusinessName:    "StressTest Corp",
    BusinessAddress: "321 Stress St, Test City, TC 13579",
    Industry:        "Financial Services",
    Country:         "US",
    Scenarios: []kyb.Scenario{
        {
            Name: "Economic Recession",
            Description: "Simulate economic downturn conditions",
            Parameters: map[string]interface{}{
                "gdp_growth": -0.05,
                "unemployment_rate": 0.12,
                "interest_rate": 0.08,
                "market_volatility": 1.5,
            },
        },
        {
            Name: "Regulatory Changes",
            Description: "Simulate new regulatory requirements",
            Parameters: map[string]interface{}{
                "compliance_cost": 1.3,
                "regulatory_complexity": 1.4,
                "audit_frequency": 1.2,
            },
        },
        {
            Name: "Competitive Pressure",
            Description: "Simulate increased competition",
            Parameters: map[string]interface{}{
                "competition_intensity": 1.6,
                "price_pressure": 1.3,
                "market_share_loss": 0.8,
            },
        },
    },
    MonteCarloSimulations: 2000,
    PredictionHorizon: 24, // 2-year stress test
    IncludeSensitivityAnalysis: true,
}

stressResult, err := client.AnalyzeScenarios(ctx, stressRequest)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Stress Test ID: %s\n", stressResult.ID)

for _, scenario := range stressResult.Scenarios {
    fmt.Printf("\n--- %s Stress Test ---\n", scenario.Name)
    fmt.Printf("Worst Case Risk Score: %.3f\n", scenario.MaxRiskScore)
    fmt.Printf("Probability of Failure: %.2f%%\n", 
        scenario.FailureProbability * 100)
    
    // Display sensitivity analysis
    if scenario.SensitivityAnalysis != nil {
        fmt.Println("Sensitivity Analysis:")
        for _, factor := range scenario.SensitivityAnalysis.Factors {
            fmt.Printf("  %s: %.3f (Impact: %s)\n", 
                factor.Name, factor.Sensitivity, factor.Impact)
        }
    }
}
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
# Perform stress testing
stress_request = {
    "business_name": "StressTest Corp",
    "business_address": "321 Stress St, Test City, TC 13579",
    "industry": "Financial Services",
    "country": "US",
    "scenarios": [
        {
            "name": "Economic Recession",
            "description": "Simulate economic downturn conditions",
            "parameters": {
                "gdp_growth": -0.05,
                "unemployment_rate": 0.12,
                "interest_rate": 0.08,
                "market_volatility": 1.5
            }
        },
        {
            "name": "Regulatory Changes",
            "description": "Simulate new regulatory requirements",
            "parameters": {
                "compliance_cost": 1.3,
                "regulatory_complexity": 1.4,
                "audit_frequency": 1.2
            }
        },
        {
            "name": "Competitive Pressure",
            "description": "Simulate increased competition",
            "parameters": {
                "competition_intensity": 1.6,
                "price_pressure": 1.3,
                "market_share_loss": 0.8
            }
        }
    ],
    "monte_carlo_simulations": 2000,
    "prediction_horizon": 24,  # 2-year stress test
    "include_sensitivity_analysis": True
}

stress_result = client.analyze_scenarios(stress_request)

print(f"Stress Test ID: {stress_result.id}")

for scenario in stress_result.scenarios:
    print(f"\n--- {scenario.name} Stress Test ---")
    print(f"Worst Case Risk Score: {scenario.max_risk_score:.3f}")
    print(f"Probability of Failure: {scenario.failure_probability * 100:.2f}%")
    
    # Display sensitivity analysis
    if scenario.sensitivity_analysis:
        print("Sensitivity Analysis:")
        for factor in scenario.sensitivity_analysis.factors:
            print(f"  {factor.name}: {factor.sensitivity:.3f} (Impact: {factor.impact})")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
// Perform stress testing
const stressRequest = {
    businessName: 'StressTest Corp',
    businessAddress: '321 Stress St, Test City, TC 13579',
    industry: 'Financial Services',
    country: 'US',
    scenarios: [
        {
            name: 'Economic Recession',
            description: 'Simulate economic downturn conditions',
            parameters: {
                gdp_growth: -0.05,
                unemployment_rate: 0.12,
                interest_rate: 0.08,
                market_volatility: 1.5
            }
        },
        {
            name: 'Regulatory Changes',
            description: 'Simulate new regulatory requirements',
            parameters: {
                compliance_cost: 1.3,
                regulatory_complexity: 1.4,
                audit_frequency: 1.2
            }
        },
        {
            name: 'Competitive Pressure',
            description: 'Simulate increased competition',
            parameters: {
                competition_intensity: 1.6,
                price_pressure: 1.3,
                market_share_loss: 0.8
            }
        }
    ],
    monteCarloSimulations: 2000,
    predictionHorizon: 24, // 2-year stress test
    includeSensitivityAnalysis: true
};

async function performStressTest() {
    try {
        const stressResult = await client.analyzeScenarios(stressRequest);
        
        console.log(`Stress Test ID: ${stressResult.id}`);
        
        stressResult.scenarios.forEach(scenario => {
            console.log(`\n--- ${scenario.name} Stress Test ---`);
            console.log(`Worst Case Risk Score: ${scenario.maxRiskScore.toFixed(3)}`);
            console.log(`Probability of Failure: ${(scenario.failureProbability * 100).toFixed(2)}%`);
            
            // Display sensitivity analysis
            if (scenario.sensitivityAnalysis) {
                console.log('Sensitivity Analysis:');
                scenario.sensitivityAnalysis.factors.forEach(factor => {
                    console.log(`  ${factor.name}: ${factor.sensitivity.toFixed(3)} (Impact: ${factor.impact})`);
                });
            }
        });
    } catch (error) {
        console.error('Error:', error.message);
    }
}

performStressTest();
```

</details>

## Explainable AI (SHAP)

### Get Model Explanations

<details>
<summary><strong>Go</strong></summary>

```go
// Get explainability for an assessment
explainRequest := &kyb.ExplainabilityRequest{
    AssessmentID: "risk_abc123def456",
    IncludeFeatureImportance: true,
    IncludeSHAPValues: true,
    IncludePartialDependence: true,
}

explanation, err := client.ExplainAssessment(ctx, explainRequest)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Assessment ID: %s\n", explanation.AssessmentID)
fmt.Printf("Model Used: %s\n", explanation.ModelUsed)
fmt.Printf("Overall Risk Score: %.3f\n", explanation.OverallRiskScore)

// Display feature importance
fmt.Println("\nFeature Importance:")
for _, feature := range explanation.FeatureImportance {
    fmt.Printf("  %s: %.3f (%.2f%%)\n", 
        feature.Name, feature.Importance, feature.Percentage)
}

// Display SHAP values
fmt.Println("\nSHAP Values:")
for _, shap := range explanation.SHAPValues {
    fmt.Printf("  %s: %.3f\n", shap.FeatureName, shap.Value)
    if shap.Description != "" {
        fmt.Printf("    Description: %s\n", shap.Description)
    }
}

// Display partial dependence
if explanation.PartialDependence != nil {
    fmt.Println("\nPartial Dependence Analysis:")
    for _, pd := range explanation.PartialDependence {
        fmt.Printf("  %s: %s\n", pd.FeatureName, pd.Description)
    }
}
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
# Get explainability for an assessment
explain_request = {
    "assessment_id": "risk_abc123def456",
    "include_feature_importance": True,
    "include_shap_values": True,
    "include_partial_dependence": True
}

explanation = client.explain_assessment(explain_request)

print(f"Assessment ID: {explanation.assessment_id}")
print(f"Model Used: {explanation.model_used}")
print(f"Overall Risk Score: {explanation.overall_risk_score:.3f}")

# Display feature importance
print("\nFeature Importance:")
for feature in explanation.feature_importance:
    print(f"  {feature.name}: {feature.importance:.3f} ({feature.percentage:.2f}%)")

# Display SHAP values
print("\nSHAP Values:")
for shap in explanation.shap_values:
    print(f"  {shap.feature_name}: {shap.value:.3f}")
    if shap.description:
        print(f"    Description: {shap.description}")

# Display partial dependence
if explanation.partial_dependence:
    print("\nPartial Dependence Analysis:")
    for pd in explanation.partial_dependence:
        print(f"  {pd.feature_name}: {pd.description}")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
// Get explainability for an assessment
const explainRequest = {
    assessmentId: 'risk_abc123def456',
    includeFeatureImportance: true,
    includeSHAPValues: true,
    includePartialDependence: true
};

async function explainAssessment() {
    try {
        const explanation = await client.explainAssessment(explainRequest);
        
        console.log(`Assessment ID: ${explanation.assessmentId}`);
        console.log(`Model Used: ${explanation.modelUsed}`);
        console.log(`Overall Risk Score: ${explanation.overallRiskScore.toFixed(3)}`);
        
        // Display feature importance
        console.log('\nFeature Importance:');
        explanation.featureImportance.forEach(feature => {
            console.log(`  ${feature.name}: ${feature.importance.toFixed(3)} (${feature.percentage.toFixed(2)}%)`);
        });
        
        // Display SHAP values
        console.log('\nSHAP Values:');
        explanation.shapValues.forEach(shap => {
            console.log(`  ${shap.featureName}: ${shap.value.toFixed(3)}`);
            if (shap.description) {
                console.log(`    Description: ${shap.description}`);
            }
        });
        
        // Display partial dependence
        if (explanation.partialDependence) {
            console.log('\nPartial Dependence Analysis:');
            explanation.partialDependence.forEach(pd => {
                console.log(`  ${pd.featureName}: ${pd.description}`);
            });
        }
    } catch (error) {
        console.error('Error:', error.message);
    }
}

explainAssessment();
```

</details>

## Model Comparison and A/B Testing

### Create A/B Test Experiment

<details>
<summary><strong>Go</strong></summary>

```go
// Create an A/B test experiment
experimentRequest := &kyb.ExperimentRequest{
    Name: "Model Comparison Q1 2024",
    Description: "Compare XGBoost vs Ensemble models for Q1 2024",
    Models: []kyb.ExperimentModel{
        {
            Name: "XGBoost Model",
            ModelType: "xgboost",
            Version: "v2.1.0",
            Weight: 0.5, // 50% of traffic
        },
        {
            Name: "Ensemble Model",
            ModelType: "ensemble",
            Version: "v1.8.0",
            Weight: 0.5, // 50% of traffic
        },
    },
    SuccessMetrics: []string{"accuracy", "precision", "recall", "f1_score"},
    Duration: 30, // 30 days
    SampleSize: 1000, // Minimum 1000 assessments per model
}

experiment, err := client.CreateExperiment(ctx, experimentRequest)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Experiment ID: %s\n", experiment.ID)
fmt.Printf("Status: %s\n", experiment.Status)
fmt.Printf("Start Date: %s\n", experiment.StartDate)
fmt.Printf("End Date: %s\n", experiment.EndDate)

// Make assessments with experiment
assessmentRequest := &kyb.RiskAssessmentRequest{
    BusinessName:    "TestBusiness Inc",
    BusinessAddress: "123 Test St, Experiment City, EC 12345",
    Industry:        "Technology",
    Country:         "US",
    ExperimentID:    experiment.ID, // Include experiment ID
}

assessment, err := client.AssessRisk(ctx, assessmentRequest)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Assessment used model: %s\n", assessment.ModelUsed)
fmt.Printf("Experiment group: %s\n", assessment.ExperimentGroup)
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
# Create an A/B test experiment
experiment_request = {
    "name": "Model Comparison Q1 2024",
    "description": "Compare XGBoost vs Ensemble models for Q1 2024",
    "models": [
        {
            "name": "XGBoost Model",
            "model_type": "xgboost",
            "version": "v2.1.0",
            "weight": 0.5  # 50% of traffic
        },
        {
            "name": "Ensemble Model",
            "model_type": "ensemble",
            "version": "v1.8.0",
            "weight": 0.5  # 50% of traffic
        }
    ],
    "success_metrics": ["accuracy", "precision", "recall", "f1_score"],
    "duration": 30,  # 30 days
    "sample_size": 1000  # Minimum 1000 assessments per model
}

experiment = client.create_experiment(experiment_request)

print(f"Experiment ID: {experiment.id}")
print(f"Status: {experiment.status}")
print(f"Start Date: {experiment.start_date}")
print(f"End Date: {experiment.end_date}")

# Make assessments with experiment
assessment_request = {
    "business_name": "TestBusiness Inc",
    "business_address": "123 Test St, Experiment City, EC 12345",
    "industry": "Technology",
    "country": "US",
    "experiment_id": experiment.id  # Include experiment ID
}

assessment = client.assess_risk(assessment_request)

print(f"Assessment used model: {assessment.model_used}")
print(f"Experiment group: {assessment.experiment_group}")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
// Create an A/B test experiment
const experimentRequest = {
    name: 'Model Comparison Q1 2024',
    description: 'Compare XGBoost vs Ensemble models for Q1 2024',
    models: [
        {
            name: 'XGBoost Model',
            modelType: 'xgboost',
            version: 'v2.1.0',
            weight: 0.5 // 50% of traffic
        },
        {
            name: 'Ensemble Model',
            modelType: 'ensemble',
            version: 'v1.8.0',
            weight: 0.5 // 50% of traffic
        }
    ],
    successMetrics: ['accuracy', 'precision', 'recall', 'f1_score'],
    duration: 30, // 30 days
    sampleSize: 1000 // Minimum 1000 assessments per model
};

async function createExperiment() {
    try {
        const experiment = await client.createExperiment(experimentRequest);
        
        console.log(`Experiment ID: ${experiment.id}`);
        console.log(`Status: ${experiment.status}`);
        console.log(`Start Date: ${experiment.startDate}`);
        console.log(`End Date: ${experiment.endDate}`);
        
        // Make assessments with experiment
        const assessmentRequest = {
            businessName: 'TestBusiness Inc',
            businessAddress: '123 Test St, Experiment City, EC 12345',
            industry: 'Technology',
            country: 'US',
            experimentId: experiment.id // Include experiment ID
        };
        
        const assessment = await client.assessRisk(assessmentRequest);
        
        console.log(`Assessment used model: ${assessment.modelUsed}`);
        console.log(`Experiment group: ${assessment.experimentGroup}`);
    } catch (error) {
        console.error('Error:', error.message);
    }
}

createExperiment();
```

</details>

### Get Experiment Results

<details>
<summary><strong>Go</strong></summary>

```go
// Get experiment results
results, err := client.GetExperimentResults(ctx, experiment.ID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Experiment: %s\n", results.ExperimentName)
fmt.Printf("Status: %s\n", results.Status)
fmt.Printf("Total Assessments: %d\n", results.TotalAssessments)

// Display model performance
for _, model := range results.ModelResults {
    fmt.Printf("\n--- %s ---\n", model.ModelName)
    fmt.Printf("Assessments: %d\n", model.AssessmentCount)
    fmt.Printf("Accuracy: %.3f\n", model.Metrics.Accuracy)
    fmt.Printf("Precision: %.3f\n", model.Metrics.Precision)
    fmt.Printf("Recall: %.3f\n", model.Metrics.Recall)
    fmt.Printf("F1 Score: %.3f\n", model.Metrics.F1Score)
    
    // Display confidence intervals
    if model.ConfidenceIntervals != nil {
        fmt.Printf("Accuracy CI: [%.3f, %.3f]\n", 
            model.ConfidenceIntervals.Accuracy.Lower,
            model.ConfidenceIntervals.Accuracy.Upper)
    }
}

// Display statistical significance
if results.StatisticalSignificance != nil {
    fmt.Printf("\nStatistical Significance: %.3f\n", 
        results.StatisticalSignificance.PValue)
    fmt.Printf("Significant: %t\n", 
        results.StatisticalSignificance.IsSignificant)
}
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
# Get experiment results
results = client.get_experiment_results(experiment.id)

print(f"Experiment: {results.experiment_name}")
print(f"Status: {results.status}")
print(f"Total Assessments: {results.total_assessments}")

# Display model performance
for model in results.model_results:
    print(f"\n--- {model.model_name} ---")
    print(f"Assessments: {model.assessment_count}")
    print(f"Accuracy: {model.metrics.accuracy:.3f}")
    print(f"Precision: {model.metrics.precision:.3f}")
    print(f"Recall: {model.metrics.recall:.3f}")
    print(f"F1 Score: {model.metrics.f1_score:.3f}")
    
    # Display confidence intervals
    if model.confidence_intervals:
        print(f"Accuracy CI: [{model.confidence_intervals.accuracy.lower:.3f}, {model.confidence_intervals.accuracy.upper:.3f}]")

# Display statistical significance
if results.statistical_significance:
    print(f"\nStatistical Significance: {results.statistical_significance.p_value:.3f}")
    print(f"Significant: {results.statistical_significance.is_significant}")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
// Get experiment results
async function getExperimentResults() {
    try {
        const results = await client.getExperimentResults(experiment.id);
        
        console.log(`Experiment: ${results.experimentName}`);
        console.log(`Status: ${results.status}`);
        console.log(`Total Assessments: ${results.totalAssessments}`);
        
        // Display model performance
        results.modelResults.forEach(model => {
            console.log(`\n--- ${model.modelName} ---`);
            console.log(`Assessments: ${model.assessmentCount}`);
            console.log(`Accuracy: ${model.metrics.accuracy.toFixed(3)}`);
            console.log(`Precision: ${model.metrics.precision.toFixed(3)}`);
            console.log(`Recall: ${model.metrics.recall.toFixed(3)}`);
            console.log(`F1 Score: ${model.metrics.f1Score.toFixed(3)}`);
            
            // Display confidence intervals
            if (model.confidenceIntervals) {
                console.log(`Accuracy CI: [${model.confidenceIntervals.accuracy.lower.toFixed(3)}, ${model.confidenceIntervals.accuracy.upper.toFixed(3)}]`);
            }
        });
        
        // Display statistical significance
        if (results.statisticalSignificance) {
            console.log(`\nStatistical Significance: ${results.statisticalSignificance.pValue.toFixed(3)}`);
            console.log(`Significant: ${results.statisticalSignificance.isSignificant}`);
        }
    } catch (error) {
        console.error('Error:', error.message);
    }
}

getExperimentResults();
```

</details>

## Real-World Use Cases

### 1. Portfolio Risk Management

```go
func analyzePortfolioRisk(client *kyb.Client, businessIDs []string) (*PortfolioRiskAnalysis, error) {
    // Get multi-horizon predictions for all businesses
    var predictions []*kyb.AdvancedPredictionResponse
    for _, businessID := range businessIDs {
        request := &kyb.AdvancedPredictionRequest{
            BusinessID: businessID,
            PredictionHorizons: []int{3, 6, 12},
            IncludeConfidenceIntervals: true,
        }
        
        prediction, err := client.PredictAdvanced(context.Background(), request)
        if err != nil {
            return nil, err
        }
        
        predictions = append(predictions, prediction)
    }
    
    // Calculate portfolio risk metrics
    analysis := &PortfolioRiskAnalysis{
        TotalBusinesses: len(businessIDs),
        AverageRiskScore: calculateAverageRisk(predictions),
        RiskDistribution: calculateRiskDistribution(predictions),
        CorrelationMatrix: calculateCorrelations(predictions),
        StressTestResults: performPortfolioStressTest(predictions),
    }
    
    return analysis, nil
}
```

### 2. Dynamic Risk Monitoring

```python
class DynamicRiskMonitor:
    def __init__(self, client):
        self.client = client
        self.monitored_businesses = {}
    
    def add_business(self, business_id, thresholds):
        """Add a business to monitoring with risk thresholds"""
        self.monitored_businesses[business_id] = {
            'thresholds': thresholds,
            'last_assessment': None,
            'risk_trend': []
        }
    
    def check_risk_changes(self, business_id):
        """Check for significant risk changes"""
        # Get current assessment
        current = self.client.get_assessment(business_id)
        
        # Get historical trend
        historical = self.client.get_risk_history(business_id, days=30)
        
        # Calculate trend
        trend = self.calculate_trend(historical)
        
        # Check thresholds
        alerts = []
        if current.risk_score > self.monitored_businesses[business_id]['thresholds']['high_risk']:
            alerts.append({
                'type': 'high_risk',
                'message': f'Risk score {current.risk_score:.3f} exceeds high risk threshold',
                'business_id': business_id
            })
        
        if trend.slope > 0.1:  # Increasing risk trend
            alerts.append({
                'type': 'increasing_risk',
                'message': f'Risk trend is increasing: {trend.slope:.3f}',
                'business_id': business_id
            })
        
        return alerts
    
    def run_monitoring_cycle(self):
        """Run monitoring cycle for all businesses"""
        all_alerts = []
        for business_id in self.monitored_businesses:
            alerts = self.check_risk_changes(business_id)
            all_alerts.extend(alerts)
        
        return all_alerts
```

### 3. Predictive Compliance Monitoring

```javascript
class PredictiveComplianceMonitor {
    constructor(client) {
        this.client = client;
    }
    
    async monitorComplianceRisk(businessId) {
        // Get current compliance assessment
        const current = await this.client.getComprehensiveExternalData(businessId);
        
        // Get predictive risk assessment
        const prediction = await this.client.predictAdvanced({
            businessId: businessId,
            predictionHorizons: [3, 6, 12],
            includeComplianceFactors: true
        });
        
        // Analyze compliance trends
        const complianceTrend = this.analyzeComplianceTrend(current, prediction);
        
        // Generate alerts
        const alerts = [];
        
        if (complianceTrend.ofacRisk > 0.7) {
            alerts.push({
                type: 'ofac_risk',
                severity: 'high',
                message: 'High OFAC risk predicted in next 3 months',
                businessId: businessId
            });
        }
        
        if (complianceTrend.adverseMediaTrend > 0.5) {
            alerts.push({
                type: 'adverse_media',
                severity: 'medium',
                message: 'Increasing adverse media risk detected',
                businessId: businessId
            });
        }
        
        return {
            businessId: businessId,
            currentCompliance: current,
            predictedCompliance: prediction,
            complianceTrend: complianceTrend,
            alerts: alerts
        };
    }
    
    analyzeComplianceTrend(current, prediction) {
        return {
            ofacRisk: prediction.predictions[0].complianceFactors.ofacRisk,
            adverseMediaTrend: this.calculateTrend(prediction.predictions.map(p => p.complianceFactors.adverseMediaRisk)),
            sanctionsRisk: prediction.predictions[0].complianceFactors.sanctionsRisk
        };
    }
}
```

## Best Practices

### 1. Model Selection

- **Short-term predictions (1-3 months)**: Use XGBoost for accuracy
- **Medium-term predictions (3-6 months)**: Use Ensemble for balanced performance
- **Long-term predictions (6-12 months)**: Use LSTM for time-series patterns

### 2. Confidence Intervals

- Always include confidence intervals for uncertainty quantification
- Use wider confidence intervals for longer prediction horizons
- Consider confidence scores when making business decisions

### 3. Scenario Analysis

- Test multiple scenarios including optimistic, pessimistic, and realistic cases
- Use Monte Carlo simulation for robust risk assessment
- Include sensitivity analysis for key risk factors

### 4. Model Validation

- Use A/B testing for model comparison
- Monitor model performance over time
- Implement model drift detection

### 5. Explainability

- Use SHAP values for model interpretability
- Provide feature importance rankings
- Include partial dependence analysis for complex models

## Troubleshooting

### Common Issues

1. **Low Prediction Confidence**
   - Provide more complete business information
   - Use shorter prediction horizons
   - Consider manual review for low-confidence predictions

2. **Inconsistent Predictions**
   - Check for data quality issues
   - Verify model versions and training dates
   - Use ensemble models for stability

3. **Scenario Analysis Errors**
   - Validate scenario parameters
   - Ensure sufficient simulation runs
   - Check parameter ranges and constraints

### Getting Help

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **API Reference**: [https://docs.kyb-platform.com/api](https://docs.kyb-platform.com/api)
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)
- **Email Support**: [dev-support@kyb-platform.com](mailto:dev-support@kyb-platform.com)

---

**Last Updated**: January 15, 2024  
**Version**: 2.0.0  
**Next Review**: April 15, 2024
