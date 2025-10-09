# KYB Platform Risk Assessment Service Python SDK

[![PyPI version](https://badge.fury.io/py/kyb-sdk.svg)](https://badge.fury.io/py/kyb-sdk)
[![Python Support](https://img.shields.io/pypi/pyversions/kyb-sdk.svg)](https://pypi.org/project/kyb-sdk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The official Python SDK for the KYB Platform Risk Assessment Service API. This SDK provides a simple and intuitive interface for performing business risk assessments, compliance checks, and analytics.

## Features

- **Risk Assessment**: Comprehensive business risk evaluation with ML-powered predictions
- **Compliance Checking**: Automated compliance verification and monitoring
- **Sanctions Screening**: Real-time sanctions and watchlist screening
- **Media Monitoring**: Adverse media monitoring and alerting
- **Analytics**: Risk trends and insights analysis
- **Type Safety**: Full type hints for better development experience
- **Error Handling**: Comprehensive error handling with detailed error information
- **Retry Logic**: Automatic retry with exponential backoff
- **Context Manager**: Support for context manager usage

## Installation

```bash
pip install kyb-sdk
```

## Quick Start

```python
from kyb_sdk import KYBClient

# Initialize the client
client = KYBClient(api_key="your_api_key_here")

# Perform a risk assessment
assessment = client.assess_risk(
    business_name="Acme Corporation",
    business_address="123 Main St, Anytown, ST 12345",
    industry="Technology",
    country="US",
    phone="+1-555-123-4567",
    email="contact@acme.com",
    website="https://www.acme.com"
)

print(f"Risk Score: {assessment['risk_score']}")
print(f"Risk Level: {assessment['risk_level']}")
print(f"Confidence: {assessment['confidence_score']}")
```

## Usage Examples

### Risk Assessment

```python
from kyb_sdk import KYBClient

client = KYBClient(api_key="your_api_key")

# Basic risk assessment
assessment = client.assess_risk(
    business_name="Acme Corporation",
    business_address="123 Main St, Anytown, ST 12345",
    industry="Technology",
    country="US"
)

# Risk assessment with metadata
assessment = client.assess_risk(
    business_name="Acme Corporation",
    business_address="123 Main St, Anytown, ST 12345",
    industry="Technology",
    country="US",
    prediction_horizon=6,
    metadata={
        "annual_revenue": 1000000,
        "employee_count": 50,
        "founded_year": 2020
    }
)

# Get risk assessment by ID
assessment = client.get_risk_assessment("risk_1234567890")
```

### Risk Prediction

```python
# Predict future risk
prediction = client.predict_risk(
    assessment_id="risk_1234567890",
    horizon_months=6,
    scenarios=["optimistic", "realistic", "pessimistic"]
)

print(f"Predicted Score: {prediction['predicted_score']}")
print(f"Predicted Level: {prediction['predicted_level']}")
```

### Compliance Checking

```python
# Check compliance
compliance = client.check_compliance(
    business_name="Acme Corporation",
    business_address="123 Main St, Anytown, ST 12345",
    industry="Technology",
    country="US",
    compliance_types=["kyc", "aml", "sanctions"]
)

print(f"Compliance Status: {compliance['compliance_status']}")
```

### Sanctions Screening

```python
# Screen for sanctions
sanctions = client.screen_sanctions(
    business_name="Acme Corporation",
    business_address="123 Main St, Anytown, ST 12345",
    country="US"
)

print(f"Sanctions Status: {sanctions['sanctions_status']}")
```

### Media Monitoring

```python
# Set up media monitoring
monitoring = client.monitor_media(
    business_name="Acme Corporation",
    business_address="123 Main St, Anytown, ST 12345",
    monitoring_types=["news", "social_media", "regulatory"]
)

print(f"Monitoring ID: {monitoring['monitoring_id']}")
```

### Analytics

```python
# Get risk trends
trends = client.get_risk_trends(
    industry="Technology",
    country="US",
    timeframe="30d",
    limit=100
)

print(f"Average Risk Score: {trends['summary']['average_risk_score']}")

# Get risk insights
insights = client.get_risk_insights(
    industry="Technology",
    country="US",
    risk_level="high"
)

for insight in insights['insights']:
    print(f"Insight: {insight['title']}")
```

## Error Handling

The SDK provides comprehensive error handling with specific exception types:

```python
from kyb_sdk import KYBClient, ValidationError, AuthenticationError, APIError

client = KYBClient(api_key="your_api_key")

try:
    assessment = client.assess_risk(
        business_name="",  # This will cause a validation error
        business_address="123 Main St",
        industry="Technology",
        country="US"
    )
except ValidationError as e:
    print(f"Validation Error: {e.message}")
    print(f"Request ID: {e.get_request_id()}")
    for error in e.get_validation_errors():
        print(f"Field: {error['field']}, Message: {error['message']}")
except AuthenticationError as e:
    print(f"Authentication Error: {e.message}")
except APIError as e:
    print(f"API Error: {e.message}")
    print(f"Status Code: {e.status_code}")
```

## Configuration

```python
from kyb_sdk import KYBClient

# Custom configuration
client = KYBClient(
    api_key="your_api_key",
    base_url="https://api.kyb-platform.com/v1",
    timeout=60,
    max_retries=5,
    user_agent="my-app/1.0.0"
)
```

## Context Manager

The client supports context manager usage for automatic resource cleanup:

```python
from kyb_sdk import KYBClient

with KYBClient(api_key="your_api_key") as client:
    assessment = client.assess_risk(
        business_name="Acme Corporation",
        business_address="123 Main St, Anytown, ST 12345",
        industry="Technology",
        country="US"
    )
    # Client is automatically closed when exiting the context
```

## Async Support

For async applications, you can use the client with asyncio:

```python
import asyncio
from kyb_sdk import KYBClient

async def main():
    client = KYBClient(api_key="your_api_key")
    
    # Run in thread pool for async compatibility
    loop = asyncio.get_event_loop()
    assessment = await loop.run_in_executor(
        None,
        client.assess_risk,
        "Acme Corporation",
        "123 Main St, Anytown, ST 12345",
        "Technology",
        "US"
    )
    
    print(f"Risk Score: {assessment['risk_score']}")

asyncio.run(main())
```

## Rate Limiting

The SDK automatically handles rate limiting with exponential backoff:

```python
client = KYBClient(
    api_key="your_api_key",
    max_retries=5  # Maximum number of retries
)
```

## Logging

Enable logging to see request details:

```python
import logging

# Enable logging
logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger('kyb_sdk')
logger.setLevel(logging.DEBUG)
```

## Testing

```python
import unittest
from unittest.mock import patch, Mock
from kyb_sdk import KYBClient, ValidationError

class TestKYBClient(unittest.TestCase):
    def setUp(self):
        self.client = KYBClient(api_key="test_key")
    
    def test_assess_risk_validation(self):
        with self.assertRaises(ValidationError):
            self.client.assess_risk(
                business_name="",  # Invalid: empty name
                business_address="123 Main St",
                industry="Technology",
                country="US"
            )
    
    @patch('kyb_sdk.client.requests.Session.post')
    def test_assess_risk_success(self, mock_post):
        # Mock successful response
        mock_response = Mock()
        mock_response.json.return_value = {
            "id": "risk_123",
            "risk_score": 0.75,
            "risk_level": "medium"
        }
        mock_response.status_code = 200
        mock_post.return_value = mock_response
        
        result = self.client.assess_risk(
            business_name="Acme Corporation",
            business_address="123 Main St",
            industry="Technology",
            country="US"
        )
        
        self.assertEqual(result["risk_score"], 0.75)
        self.assertEqual(result["risk_level"], "medium")

if __name__ == '__main__':
    unittest.main()
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **Issues**: [GitHub Issues](https://github.com/kyb-platform/python-sdk/issues)
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
- Type hints support
- Context manager support
