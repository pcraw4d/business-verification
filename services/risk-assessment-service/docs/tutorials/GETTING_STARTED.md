# Getting Started with Risk Assessment Service

## Overview

This tutorial will help you get started with the Risk Assessment Service in just 5 minutes. You'll learn how to install the SDK, make your first API call, and understand the basic concepts.

## Prerequisites

- An API key from the KYB Platform dashboard
- Basic knowledge of your chosen programming language
- Internet connection

## Quick Start (5 Minutes)

### Step 1: Get Your API Key

1. Log in to your [KYB Platform dashboard](https://dashboard.kyb-platform.com)
2. Navigate to "API Settings" → "API Keys"
3. Generate a new API key
4. Copy and securely store your API key

### Step 2: Install the SDK

Choose your preferred language:

<details>
<summary><strong>Go</strong></summary>

```bash
go get github.com/kyb-platform/go-sdk
```

</details>

<details>
<summary><strong>Python</strong></summary>

```bash
pip install kyb-risk-assessment
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```bash
npm install @kyb-platform/risk-assessment
```

</details>

<details>
<summary><strong>Ruby</strong></summary>

```bash
gem install kyb-risk-assessment
```

</details>

<details>
<summary><strong>Java</strong></summary>

Add to your `pom.xml`:

```xml
<dependency>
    <groupId>com.kyb-platform</groupId>
    <artifactId>risk-assessment-sdk</artifactId>
    <version>2.0.0</version>
</dependency>
```

</details>

<details>
<summary><strong>PHP</strong></summary>

```bash
composer require kyb-platform/risk-assessment-sdk
```

</details>

### Step 3: Make Your First API Call

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
    // Initialize the client
    client := kyb.NewClient(&kyb.Config{
        APIKey: "your_api_key_here",
    })
    
    // Create a risk assessment request
    request := &kyb.RiskAssessmentRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
    }
    
    // Make the API call
    ctx := context.Background()
    assessment, err := client.AssessRisk(ctx, request)
    if err != nil {
        log.Fatal(err)
    }
    
    // Print the result
    fmt.Printf("Risk Score: %.2f\n", assessment.RiskScore)
    fmt.Printf("Risk Level: %s\n", assessment.RiskLevel)
}
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
from kyb_risk_assessment import RiskAssessmentClient

# Initialize the client
client = RiskAssessmentClient(api_key="your_api_key_here")

# Create a risk assessment request
request = {
    "business_name": "Acme Corporation",
    "business_address": "123 Main St, Anytown, ST 12345",
    "industry": "Technology",
    "country": "US"
}

# Make the API call
assessment = client.assess_risk(request)

# Print the result
print(f"Risk Score: {assessment.risk_score:.2f}")
print(f"Risk Level: {assessment.risk_level}")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
const { RiskAssessmentClient } = require('@kyb-platform/risk-assessment');

// Initialize the client
const client = new RiskAssessmentClient({
    apiKey: 'your_api_key_here'
});

// Create a risk assessment request
const request = {
    businessName: 'Acme Corporation',
    businessAddress: '123 Main St, Anytown, ST 12345',
    industry: 'Technology',
    country: 'US'
};

// Make the API call
async function assessRisk() {
    try {
        const assessment = await client.assessRisk(request);
        
        // Print the result
        console.log(`Risk Score: ${assessment.riskScore.toFixed(2)}`);
        console.log(`Risk Level: ${assessment.riskLevel}`);
    } catch (error) {
        console.error('Error:', error.message);
    }
}

assessRisk();
```

</details>

<details>
<summary><strong>Ruby</strong></summary>

```ruby
require 'kyb-risk-assessment'

# Initialize the client
client = KybRiskAssessment::Client.new(api_key: 'your_api_key_here')

# Create a risk assessment request
request = {
    business_name: 'Acme Corporation',
    business_address: '123 Main St, Anytown, ST 12345',
    industry: 'Technology',
    country: 'US'
}

# Make the API call
assessment = client.assess_risk(request)

# Print the result
puts "Risk Score: #{assessment.risk_score.round(2)}"
puts "Risk Level: #{assessment.risk_level}"
```

</details>

<details>
<summary><strong>Java</strong></summary>

```java
import com.kybplatform.riskassessment.RiskAssessmentClient;
import com.kybplatform.riskassessment.models.RiskAssessmentRequest;
import com.kybplatform.riskassessment.models.RiskAssessmentResponse;

public class QuickStart {
    public static void main(String[] args) {
        // Initialize the client
        RiskAssessmentClient client = new RiskAssessmentClient("your_api_key_here");
        
        // Create a risk assessment request
        RiskAssessmentRequest request = RiskAssessmentRequest.builder()
            .businessName("Acme Corporation")
            .businessAddress("123 Main St, Anytown, ST 12345")
            .industry("Technology")
            .country("US")
            .build();
        
        // Make the API call
        try {
            RiskAssessmentResponse assessment = client.assessRisk(request);
            
            // Print the result
            System.out.printf("Risk Score: %.2f%n", assessment.getRiskScore());
            System.out.printf("Risk Level: %s%n", assessment.getRiskLevel());
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
        }
    }
}
```

</details>

<details>
<summary><strong>PHP</strong></summary>

```php
<?php
require_once 'vendor/autoload.php';

use KybPlatform\RiskAssessment\RiskAssessmentClient;

// Initialize the client
$client = new RiskAssessmentClient('your_api_key_here');

// Create a risk assessment request
$request = [
    'business_name' => 'Acme Corporation',
    'business_address' => '123 Main St, Anytown, ST 12345',
    'industry' => 'Technology',
    'country' => 'US'
];

// Make the API call
try {
    $assessment = $client->assessRisk($request);
    
    // Print the result
    echo "Risk Score: " . number_format($assessment->riskScore, 2) . "\n";
    echo "Risk Level: " . $assessment->riskLevel . "\n";
} catch (Exception $e) {
    echo "Error: " . $e->getMessage() . "\n";
}
?>
```

</details>

### Step 4: Understand the Response

The API response contains several key fields:

```json
{
  "id": "risk_abc123def456",
  "business_id": "biz_xyz789uvw012",
  "risk_score": 0.65,
  "risk_level": "medium",
  "risk_factors": [
    {
      "category": "financial",
      "name": "Credit Score",
      "score": 0.7,
      "weight": 0.3,
      "description": "Business credit score analysis",
      "source": "internal",
      "confidence": 0.9
    }
  ],
  "prediction_horizon": 3,
  "confidence_score": 0.88,
  "status": "completed",
  "created_at": "2024-01-15T12:00:00Z",
  "updated_at": "2024-01-15T12:00:00Z"
}
```

**Key Fields:**
- `risk_score`: Overall risk score (0.0 to 1.0)
- `risk_level`: Human-readable risk level (low, medium, high)
- `risk_factors`: Detailed breakdown of risk factors
- `confidence_score`: Confidence in the assessment
- `prediction_horizon`: Time horizon for predictions (months)

## Next Steps

Now that you've made your first API call, explore these tutorials:

1. **[Basic Risk Assessment](BASIC_RISK_ASSESSMENT.md)** - Learn about different risk assessment types
2. **[Advanced Predictions](ADVANCED_PREDICTIONS.md)** - Multi-horizon predictions and scenario analysis
3. **[Compliance Screening](COMPLIANCE_SCREENING.md)** - OFAC, UN, and EU sanctions screening
4. **[Real-time Monitoring](REAL_TIME_MONITORING.md)** - Webhooks and real-time updates
5. **[Batch Processing](BATCH_PROCESSING.md)** - Process multiple assessments efficiently

## Common Use Cases

### 1. Merchant Onboarding

```go
// Go example
func onboardMerchant(client *kyb.Client, merchantData MerchantData) error {
    request := &kyb.RiskAssessmentRequest{
        BusinessName:    merchantData.Name,
        BusinessAddress: merchantData.Address,
        Industry:        merchantData.Industry,
        Country:         merchantData.Country,
        PredictionHorizon: 6, // 6-month prediction for onboarding
    }
    
    assessment, err := client.AssessRisk(context.Background(), request)
    if err != nil {
        return err
    }
    
    // Approve if risk is low or medium
    if assessment.RiskLevel == "low" || assessment.RiskLevel == "medium" {
        return approveMerchant(merchantData)
    }
    
    return rejectMerchant(merchantData, assessment.RiskFactors)
}
```

### 2. Transaction Monitoring

```python
# Python example
def monitor_transaction(client, transaction_data):
    # Get business risk assessment
    assessment = client.get_assessment(transaction_data.business_id)
    
    # Check if risk level has changed
    if assessment.risk_level == "high":
        # Flag for manual review
        flag_for_review(transaction_data, assessment)
    
    # Get real-time risk factors
    risk_factors = client.get_risk_factors(transaction_data.business_id)
    
    return {
        "approved": assessment.risk_level != "high",
        "risk_score": assessment.risk_score,
        "risk_factors": risk_factors
    }
```

### 3. Compliance Reporting

```javascript
// Node.js example
async function generateComplianceReport(client, businessId) {
    const assessment = await client.getAssessment(businessId);
    const sanctions = await client.checkSanctions(businessId);
    const adverseMedia = await client.checkAdverseMedia(businessId);
    
    return {
        business_id: businessId,
        risk_assessment: assessment,
        sanctions_check: sanctions,
        adverse_media: adverseMedia,
        compliance_status: determineComplianceStatus(assessment, sanctions, adverseMedia),
        generated_at: new Date().toISOString()
    };
}
```

## Error Handling

All SDKs provide consistent error handling:

<details>
<summary><strong>Go</strong></summary>

```go
assessment, err := client.AssessRisk(ctx, request)
if err != nil {
    switch e := err.(type) {
    case *kyb.APIError:
        switch e.StatusCode {
        case 400:
            fmt.Printf("Bad Request: %s\n", e.Message)
        case 401:
            fmt.Printf("Unauthorized: %s\n", e.Message)
        case 429:
            fmt.Printf("Rate Limited: %s\n", e.Message)
        default:
            fmt.Printf("API Error: %s\n", e.Message)
        }
    case *kyb.NetworkError:
        fmt.Printf("Network Error: %s\n", e.Message)
    default:
        fmt.Printf("Unknown Error: %s\n", err.Error())
    }
    return
}
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
try:
    assessment = client.assess_risk(request)
except kyb.APIError as e:
    if e.status_code == 400:
        print(f"Bad Request: {e.message}")
    elif e.status_code == 401:
        print(f"Unauthorized: {e.message}")
    elif e.status_code == 429:
        print(f"Rate Limited: {e.message}")
    else:
        print(f"API Error: {e.message}")
except kyb.NetworkError as e:
    print(f"Network Error: {e.message}")
except Exception as e:
    print(f"Unknown Error: {e}")
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
try {
    const assessment = await client.assessRisk(request);
} catch (error) {
    if (error instanceof APIError) {
        switch (error.statusCode) {
            case 400:
                console.log(`Bad Request: ${error.message}`);
                break;
            case 401:
                console.log(`Unauthorized: ${error.message}`);
                break;
            case 429:
                console.log(`Rate Limited: ${error.message}`);
                break;
            default:
                console.log(`API Error: ${error.message}`);
        }
    } else if (error instanceof NetworkError) {
        console.log(`Network Error: ${error.message}`);
    } else {
        console.log(`Unknown Error: ${error.message}`);
    }
}
```

</details>

## Configuration Options

All SDKs support various configuration options:

<details>
<summary><strong>Go</strong></summary>

```go
client := kyb.NewClient(&kyb.Config{
    APIKey:     "your_api_key_here",
    BaseURL:    "https://api.kyb-platform.com/v1", // Optional: custom base URL
    Timeout:    30 * time.Second,                  // Optional: request timeout
    Retries:    3,                                 // Optional: number of retries
    UserAgent:  "MyApp/1.0",                      // Optional: custom user agent
})
```

</details>

<details>
<summary><strong>Python</strong></summary>

```python
client = RiskAssessmentClient(
    api_key="your_api_key_here",
    base_url="https://api.kyb-platform.com/v1",  # Optional: custom base URL
    timeout=30,                                   # Optional: request timeout
    retries=3,                                    # Optional: number of retries
    user_agent="MyApp/1.0"                       # Optional: custom user agent
)
```

</details>

<details>
<summary><strong>Node.js</strong></summary>

```javascript
const client = new RiskAssessmentClient({
    apiKey: 'your_api_key_here',
    baseURL: 'https://api.kyb-platform.com/v1',  // Optional: custom base URL
    timeout: 30000,                               // Optional: request timeout
    retries: 3,                                   // Optional: number of retries
    userAgent: 'MyApp/1.0'                       // Optional: custom user agent
});
```

</details>

## Best Practices

### 1. API Key Management

- Store API keys in environment variables
- Never commit API keys to version control
- Rotate API keys regularly
- Use different keys for different environments

### 2. Error Handling

- Always handle errors gracefully
- Implement retry logic for transient failures
- Log errors for debugging
- Provide meaningful error messages to users

### 3. Performance

- Use connection pooling for high-volume applications
- Implement caching for frequently accessed data
- Use batch processing for multiple assessments
- Monitor API usage and rate limits

### 4. Security

- Use HTTPS for all API calls
- Validate input data before sending requests
- Implement proper authentication
- Follow security best practices

## Support

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **API Reference**: [https://docs.kyb-platform.com/api](https://docs.kyb-platform.com/api)
- **GitHub**: [https://github.com/kyb-platform/risk-assessment-service](https://github.com/kyb-platform/risk-assessment-service)
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)
- **Email Support**: [dev-support@kyb-platform.com](mailto:dev-support@kyb-platform.com)

---

**Last Updated**: January 15, 2024  
**Version**: 2.0.0  
**Next Review**: April 15, 2024
