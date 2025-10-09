# KYB Platform Risk Assessment Service Go SDK

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyb-platform/go-sdk)](https://goreportcard.com/report/github.com/kyb-platform/go-sdk)

The official Go SDK for the KYB Platform Risk Assessment Service API. This SDK provides a simple and intuitive interface for performing business risk assessments, compliance checks, and analytics.

## Features

- **Risk Assessment**: Comprehensive business risk evaluation with ML-powered predictions
- **Compliance Checking**: Automated compliance verification and monitoring
- **Sanctions Screening**: Real-time sanctions and watchlist screening
- **Media Monitoring**: Adverse media monitoring and alerting
- **Analytics**: Risk trends and insights analysis
- **Type Safety**: Full type safety with Go's type system
- **Error Handling**: Comprehensive error handling with detailed error information
- **Context Support**: Full context.Context support for cancellation and timeouts
- **HTTP Client**: Configurable HTTP client with retry logic

## Installation

```bash
go get github.com/kyb-platform/go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kyb-platform/go-sdk"
)

func main() {
    // Initialize the client
    client, err := kyb.NewClient(&kyb.Config{
        BaseURL: "https://api.kyb-platform.com/v1",
        APIKey:  "your_api_key_here",
        Timeout: 30 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Perform a risk assessment
    ctx := context.Background()
    assessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
        Phone:           "+1-555-123-4567",
        Email:           "contact@acme.com",
        Website:         "https://www.acme.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Risk Score: %.2f\n", assessment.RiskScore)
    fmt.Printf("Risk Level: %s\n", assessment.RiskLevel)
    fmt.Printf("Confidence: %.2f\n", assessment.ConfidenceScore)
}
```

## Usage Examples

### Risk Assessment

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client, err := kyb.NewClient(&kyb.Config{
        BaseURL: "https://api.kyb-platform.com/v1",
        APIKey:  "your_api_key_here",
        Timeout: 30 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Basic risk assessment
    assessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Assessment ID: %s\n", assessment.ID)
    fmt.Printf("Risk Score: %.2f\n", assessment.RiskScore)
    fmt.Printf("Risk Level: %s\n", assessment.RiskLevel)
    
    // Risk assessment with metadata
    assessmentWithMetadata, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:      "Acme Corporation",
        BusinessAddress:   "123 Main St, Anytown, ST 12345",
        Industry:          "Technology",
        Country:           "US",
        PredictionHorizon: 6,
        Metadata: map[string]interface{}{
            "annual_revenue": 1000000,
            "employee_count": 50,
            "founded_year":   2020,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Assessment with metadata: %+v\n", assessmentWithMetadata)
    
    // Get risk assessment by ID
    retrievedAssessment, err := client.GetRiskAssessment(ctx, assessment.ID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Retrieved assessment: %+v\n", retrievedAssessment)
}
```

### Risk Prediction

```go
// Predict future risk
func predictRisk(client *kyb.Client, assessmentID string) {
    ctx := context.Background()
    
    prediction, err := client.PredictRisk(ctx, assessmentID, &kyb.RiskPredictionRequest{
        HorizonMonths: 6,
        Scenarios:     []string{"optimistic", "realistic", "pessimistic"},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Predicted Score: %.2f\n", prediction.PredictedScore)
    fmt.Printf("Predicted Level: %s\n", prediction.PredictedLevel)
    
    for _, scenario := range prediction.Scenarios {
        fmt.Printf("Scenario %s: Score %.2f, Level %s\n", 
            scenario.Name, scenario.Score, scenario.Level)
    }
}
```

### Compliance Checking

```go
// Check compliance
func checkCompliance(client *kyb.Client) {
    ctx := context.Background()
    
    compliance, err := client.CheckCompliance(ctx, &kyb.ComplianceCheckRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
        ComplianceTypes: []string{"kyc", "aml", "sanctions"},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Compliance Status: %s\n", compliance.ComplianceStatus)
    
    for _, check := range compliance.Checks {
        fmt.Printf("Check %s: %s (Score: %.2f)\n", 
            check.Type, check.Status, check.Score)
    }
}
```

### Sanctions Screening

```go
// Screen for sanctions
func screenSanctions(client *kyb.Client) {
    ctx := context.Background()
    
    sanctions, err := client.ScreenSanctions(ctx, &kyb.SanctionsScreeningRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Country:         "US",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Sanctions Status: %s\n", sanctions.SanctionsStatus)
    
    if len(sanctions.Matches) > 0 {
        fmt.Println("Sanctions matches found:")
        for _, match := range sanctions.Matches {
            fmt.Printf("- Source: %s, Score: %.2f\n", match.Source, match.Score)
        }
    }
}
```

### Media Monitoring

```go
// Set up media monitoring
func monitorMedia(client *kyb.Client) {
    ctx := context.Background()
    
    monitoring, err := client.MonitorMedia(ctx, &kyb.MediaMonitoringRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        MonitoringTypes: []string{"news", "social_media", "regulatory"},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Monitoring ID: %s\n", monitoring.MonitoringID)
    fmt.Printf("Status: %s\n", monitoring.Status)
    
    if len(monitoring.Alerts) > 0 {
        fmt.Println("Active alerts:")
        for _, alert := range monitoring.Alerts {
            fmt.Printf("- %s: %s (Severity: %s)\n", 
                alert.Type, alert.Title, alert.Severity)
        }
    }
}
```

### Analytics

```go
// Get risk trends
func getRiskTrends(client *kyb.Client) {
    ctx := context.Background()
    
    trends, err := client.GetRiskTrends(ctx, &kyb.RiskTrendsOptions{
        Industry:  "Technology",
        Country:   "US",
        Timeframe: "30d",
        Limit:     100,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Average Risk Score: %.2f\n", trends.Summary.AverageRiskScore)
    fmt.Printf("Total Assessments: %d\n", trends.Summary.TotalAssessments)
    
    for _, trend := range trends.Trends {
        fmt.Printf("Trend: %s %s - Score: %.2f, Direction: %s\n",
            trend.Industry, trend.Country, trend.AverageRiskScore, trend.TrendDirection)
    }
}

// Get risk insights
func getRiskInsights(client *kyb.Client) {
    ctx := context.Background()
    
    insights, err := client.GetRiskInsights(ctx, &kyb.RiskInsightsOptions{
        Industry:  "Technology",
        Country:   "US",
        RiskLevel: "high",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Risk Insights:")
    for _, insight := range insights.Insights {
        fmt.Printf("- %s: %s (Impact: %s)\n", 
            insight.Title, insight.Description, insight.Impact)
    }
    
    fmt.Println("Recommendations:")
    for _, rec := range insights.Recommendations {
        fmt.Printf("- %s: %s (Priority: %s)\n", 
            rec.Category, rec.Action, rec.Priority)
    }
}
```

## Error Handling

The SDK provides comprehensive error handling with specific error types:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client, err := kyb.NewClient(&kyb.Config{
        BaseURL: "https://api.kyb-platform.com/v1",
        APIKey:  "your_api_key_here",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Example with error handling
    assessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "", // This will cause a validation error
        BusinessAddress: "123 Main St",
        Industry:        "Technology",
        Country:         "US",
    })
    
    if err != nil {
        switch e := err.(type) {
        case *kyb.APIError:
            if e.IsValidationError() {
                fmt.Printf("Validation Error: %s\n", e.Error())
                fmt.Printf("Request ID: %s\n", e.GetRequestID())
                for _, validationErr := range e.GetValidationErrors() {
                    fmt.Printf("Field: %s, Message: %s\n", 
                        validationErr.Field, validationErr.Message)
                }
            } else if e.IsAuthenticationError() {
                fmt.Printf("Authentication Error: %s\n", e.Error())
            } else if e.IsRateLimitError() {
                fmt.Printf("Rate Limit Error: %s\n", e.Error())
            } else {
                fmt.Printf("API Error: %s (Status: %d)\n", e.Error(), e.StatusCode)
            }
        default:
            fmt.Printf("Unexpected Error: %s\n", err.Error())
        }
        return
    }
    
    fmt.Printf("Assessment successful: %+v\n", assessment)
}
```

## Configuration

```go
package main

import (
    "context"
    "net/http"
    "time"
    
    "github.com/kyb-platform/go-sdk"
)

func main() {
    // Custom HTTP client
    httpClient := &http.Client{
        Timeout: 60 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }
    
    // Custom configuration
    client, err := kyb.NewClient(&kyb.Config{
        BaseURL:    "https://api.kyb-platform.com/v1",
        APIKey:     "your_api_key_here",
        Timeout:    60 * time.Second,
        HTTPClient: httpClient,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Use with custom request options
    ctx := context.Background()
    assessment, err := client.AssessRiskWithOptions(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
    }, &kyb.RequestOptions{
        Headers: map[string]string{
            "X-Custom-Header": "custom-value",
        },
        Timeout: 30 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Assessment: %+v\n", assessment)
}
```

## Context Support

The SDK fully supports Go's context package for cancellation and timeouts:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client, err := kyb.NewClient(&kyb.Config{
        BaseURL: "https://api.kyb-platform.com/v1",
        APIKey:  "your_api_key_here",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Context with cancellation
    ctx2, cancel2 := context.WithCancel(context.Background())
    defer cancel2()
    
    // Use context for requests
    assessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
    })
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            fmt.Println("Request timed out")
        } else if ctx.Err() == context.Canceled {
            fmt.Println("Request was canceled")
        } else {
            fmt.Printf("Request failed: %s\n", err.Error())
        }
        return
    }
    
    fmt.Printf("Assessment: %+v\n", assessment)
}
```

## Testing

```go
package main

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/kyb-platform/go-sdk"
    "github.com/stretchr/testify/assert"
)

func TestKYBClient(t *testing.T) {
    // Create a test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/api/v1/assess" && r.Method == "POST" {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{
                "id": "risk_123",
                "risk_score": 0.75,
                "risk_level": "medium",
                "confidence_score": 0.85
            }`))
        } else {
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer server.Close()
    
    // Create client with test server URL
    client, err := kyb.NewClient(&kyb.Config{
        BaseURL: server.URL + "/api/v1",
        APIKey:  "test_key",
    })
    assert.NoError(t, err)
    
    // Test successful assessment
    ctx := context.Background()
    assessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "risk_123", assessment.ID)
    assert.Equal(t, 0.75, assessment.RiskScore)
    assert.Equal(t, "medium", assessment.RiskLevel)
    assert.Equal(t, 0.85, assessment.ConfidenceScore)
}

func TestValidationError(t *testing.T) {
    client, err := kyb.NewClient(&kyb.Config{
        BaseURL: "https://api.kyb-platform.com/v1",
        APIKey:  "test_key",
    })
    assert.NoError(t, err)
    
    ctx := context.Background()
    _, err = client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "", // Invalid: empty name
        BusinessAddress: "123 Main St",
        Industry:        "Technology",
        Country:         "US",
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "business_name is required")
}
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **Issues**: [GitHub Issues](https://github.com/kyb-platform/go-sdk/issues)
- **Email**: [support@kyb-platform.com](mailto:support@kyb-platform.com)

## Changelog

### v1.0.0 (2024-01-15)
- Initial release
- Risk assessment functionality
- Compliance checking
- Sanctions screening
- Media monitoring
- Analytics and insights
- Comprehensive error handling
- Context support
- Type safety
