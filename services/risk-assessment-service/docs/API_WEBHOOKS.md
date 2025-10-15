# API Webhooks Documentation

## Overview

Webhooks allow you to receive real-time notifications when events occur in the Risk Assessment Service. This enables you to build reactive applications that respond immediately to risk assessment completions, prediction updates, and other important events.

## How Webhooks Work

1. **Register**: Create a webhook endpoint in your application
2. **Subscribe**: Configure the Risk Assessment Service to send events to your endpoint
3. **Receive**: Your application receives HTTP POST requests with event data
4. **Process**: Handle the event data and update your application state

## Supported Events

### Risk Assessment Events

#### `assessment.completed`
Triggered when a risk assessment is completed successfully.

```json
{
  "event": "assessment.completed",
  "data": {
    "id": "risk_1234567890",
    "business_id": "biz_1234567890",
    "risk_score": 0.75,
    "risk_level": "medium",
    "confidence_score": 0.85,
    "status": "completed",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "webhook_id": "wh_1234567890"
}
```

#### `assessment.failed`
Triggered when a risk assessment fails.

```json
{
  "event": "assessment.failed",
  "data": {
    "id": "risk_1234567890",
    "business_id": "biz_1234567890",
    "status": "failed",
    "error": {
      "code": "EXTERNAL_API_ERROR",
      "message": "Thomson Reuters API temporarily unavailable"
    },
    "created_at": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "webhook_id": "wh_1234567890"
}
```

### Risk Prediction Events

#### `prediction.updated`
Triggered when a risk prediction is generated or updated.

```json
{
  "event": "prediction.updated",
  "data": {
    "business_id": "biz_1234567890",
    "assessment_id": "risk_1234567890",
    "horizon_months": 6,
    "predicted_score": 0.72,
    "predicted_level": "medium",
    "model_type": "lstm",
    "confidence": 0.85,
    "created_at": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "webhook_id": "wh_1234567890"
}
```

### Compliance Events

#### `compliance.alert`
Triggered when a compliance issue is detected.

```json
{
  "event": "compliance.alert",
  "data": {
    "business_id": "biz_1234567890",
    "alert_type": "sanctions_match",
    "severity": "high",
    "details": {
      "sanctions_list": "OFAC SDN",
      "match_score": 0.95,
      "entity_name": "John Doe"
    },
    "created_at": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "webhook_id": "wh_1234567890"
}
```

#### `sanctions.match`
Triggered when a sanctions list match is found.

```json
{
  "event": "sanctions.match",
  "data": {
    "business_id": "biz_1234567890",
    "entity_name": "John Doe",
    "entity_type": "individual",
    "sanctions_list": "OFAC SDN",
    "match_score": 0.95,
    "match_details": {
      "aliases": ["Johnny Doe", "J. Doe"],
      "nationality": "US",
      "date_of_birth": "1980-01-01"
    },
    "created_at": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "webhook_id": "wh_1234567890"
}
```

### Media Monitoring Events

#### `media.alert`
Triggered when adverse media is detected.

```json
{
  "event": "media.alert",
  "data": {
    "business_id": "biz_1234567890",
    "alert_type": "adverse_media",
    "severity": "medium",
    "media_item": {
      "title": "Company faces regulatory investigation",
      "source": "Financial Times",
      "url": "https://ft.com/content/123456",
      "published_at": "2024-01-15T09:00:00Z",
      "sentiment": "negative",
      "risk_level": "high"
    },
    "created_at": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "webhook_id": "wh_1234567890"
}
```

## Webhook Management

### Creating a Webhook

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/webhooks" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/risk-assessment",
    "events": [
      "assessment.completed",
      "assessment.failed",
      "prediction.updated",
      "compliance.alert",
      "sanctions.match",
      "media.alert"
    ],
    "secret": "your_webhook_secret_here",
    "active": true
  }'
```

**Response:**
```json
{
  "webhook_id": "wh_1234567890",
  "url": "https://your-app.com/webhooks/risk-assessment",
  "events": [
    "assessment.completed",
    "assessment.failed",
    "prediction.updated",
    "compliance.alert",
    "sanctions.match",
    "media.alert"
  ],
  "active": true,
  "delivery_stats": {
    "total_deliveries": 0,
    "successful_deliveries": 0,
    "failed_deliveries": 0,
    "success_rate": 0.0
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Retrieving Webhook Details

```bash
curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/webhooks/wh_1234567890" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Updating a Webhook

```bash
curl -X PUT "https://risk-assessment-service-production.up.railway.app/api/v1/webhooks/wh_1234567890" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "events": [
      "assessment.completed",
      "prediction.updated"
    ],
    "active": true
  }'
```

### Deleting a Webhook

```bash
curl -X DELETE "https://risk-assessment-service-production.up.railway.app/api/v1/webhooks/wh_1234567890" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## Webhook Security

### Signature Verification

All webhook payloads include a signature for verification. The signature is generated using HMAC-SHA256:

```javascript
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature, 'hex'),
    Buffer.from(expectedSignature, 'hex')
  );
}

// Express.js webhook handler
app.post('/webhooks/risk-assessment', (req, res) => {
  const signature = req.headers['x-webhook-signature'];
  const payload = JSON.stringify(req.body);
  
  if (!verifyWebhookSignature(payload, signature, WEBHOOK_SECRET)) {
    return res.status(401).send('Invalid signature');
  }
  
  // Process webhook
  const event = req.body;
  console.log(`Received event: ${event.event}`);
  
  res.status(200).send('OK');
});
```

### Security Headers

Webhook requests include the following security headers:

```http
X-Webhook-Signature: sha256=abc123def456...
X-Webhook-Timestamp: 1640995200
X-Webhook-Id: wh_1234567890
X-Webhook-Event: assessment.completed
```

### Timestamp Validation

Validate webhook timestamps to prevent replay attacks:

```javascript
function validateTimestamp(timestamp, toleranceSeconds = 300) {
  const now = Math.floor(Date.now() / 1000);
  const webhookTime = parseInt(timestamp);
  
  return Math.abs(now - webhookTime) <= toleranceSeconds;
}

app.post('/webhooks/risk-assessment', (req, res) => {
  const timestamp = req.headers['x-webhook-timestamp'];
  
  if (!validateTimestamp(timestamp)) {
    return res.status(401).send('Invalid timestamp');
  }
  
  // Process webhook
  res.status(200).send('OK');
});
```

## Webhook Implementation Examples

### Node.js with Express

```javascript
const express = require('express');
const crypto = require('crypto');
const app = express();

app.use(express.json());

const WEBHOOK_SECRET = process.env.WEBHOOK_SECRET;

function verifyWebhookSignature(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature, 'hex'),
    Buffer.from(expectedSignature, 'hex')
  );
}

app.post('/webhooks/risk-assessment', (req, res) => {
  const signature = req.headers['x-webhook-signature'];
  const payload = JSON.stringify(req.body);
  
  // Verify signature
  if (!verifyWebhookSignature(payload, signature, WEBHOOK_SECRET)) {
    console.error('Invalid webhook signature');
    return res.status(401).send('Invalid signature');
  }
  
  // Process event
  const event = req.body;
  
  switch (event.event) {
    case 'assessment.completed':
      handleAssessmentCompleted(event.data);
      break;
    case 'assessment.failed':
      handleAssessmentFailed(event.data);
      break;
    case 'prediction.updated':
      handlePredictionUpdated(event.data);
      break;
    case 'compliance.alert':
      handleComplianceAlert(event.data);
      break;
    case 'sanctions.match':
      handleSanctionsMatch(event.data);
      break;
    case 'media.alert':
      handleMediaAlert(event.data);
      break;
    default:
      console.log(`Unknown event type: ${event.event}`);
  }
  
  res.status(200).send('OK');
});

function handleAssessmentCompleted(data) {
  console.log(`Assessment completed for business ${data.business_id}`);
  console.log(`Risk Score: ${data.risk_score}, Level: ${data.risk_level}`);
  
  // Update your database
  // Send notifications
  // Trigger downstream processes
}

function handleAssessmentFailed(data) {
  console.log(`Assessment failed for business ${data.business_id}`);
  console.log(`Error: ${data.error.message}`);
  
  // Log error
  // Retry assessment
  // Notify administrators
}

function handlePredictionUpdated(data) {
  console.log(`Prediction updated for business ${data.business_id}`);
  console.log(`Predicted Score: ${data.predicted_score} (${data.horizon_months} months)`);
  
  // Update risk dashboard
  // Send alerts if risk increased
  // Update business records
}

function handleComplianceAlert(data) {
  console.log(`Compliance alert for business ${data.business_id}`);
  console.log(`Alert Type: ${data.alert_type}, Severity: ${data.severity}`);
  
  // Block transactions
  // Notify compliance team
  // Update risk status
}

function handleSanctionsMatch(data) {
  console.log(`Sanctions match found for ${data.entity_name}`);
  console.log(`Match Score: ${data.match_score}, List: ${data.sanctions_list}`);
  
  // Block entity
  // Notify compliance team
  // Generate compliance report
}

function handleMediaAlert(data) {
  console.log(`Media alert for business ${data.business_id}`);
  console.log(`Article: ${data.media_item.title}`);
  console.log(`Source: ${data.media_item.source}`);
  
  // Update risk score
  // Notify stakeholders
  // Generate media report
}

app.listen(3000, () => {
  console.log('Webhook server listening on port 3000');
});
```

### Python with Flask

```python
from flask import Flask, request, jsonify
import hmac
import hashlib
import json
import os

app = Flask(__name__)
WEBHOOK_SECRET = os.environ.get('WEBHOOK_SECRET')

def verify_webhook_signature(payload, signature, secret):
    expected_signature = hmac.new(
        secret.encode('utf-8'),
        payload.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    
    return hmac.compare_digest(signature, expected_signature)

@app.route('/webhooks/risk-assessment', methods=['POST'])
def handle_webhook():
    signature = request.headers.get('X-Webhook-Signature')
    payload = json.dumps(request.json)
    
    # Verify signature
    if not verify_webhook_signature(payload, signature, WEBHOOK_SECRET):
        return jsonify({'error': 'Invalid signature'}), 401
    
    # Process event
    event = request.json
    
    if event['event'] == 'assessment.completed':
        handle_assessment_completed(event['data'])
    elif event['event'] == 'assessment.failed':
        handle_assessment_failed(event['data'])
    elif event['event'] == 'prediction.updated':
        handle_prediction_updated(event['data'])
    elif event['event'] == 'compliance.alert':
        handle_compliance_alert(event['data'])
    elif event['event'] == 'sanctions.match':
        handle_sanctions_match(event['data'])
    elif event['event'] == 'media.alert':
        handle_media_alert(event['data'])
    else:
        print(f"Unknown event type: {event['event']}")
    
    return jsonify({'status': 'success'}), 200

def handle_assessment_completed(data):
    print(f"Assessment completed for business {data['business_id']}")
    print(f"Risk Score: {data['risk_score']}, Level: {data['risk_level']}")
    
    # Update your database
    # Send notifications
    # Trigger downstream processes

def handle_assessment_failed(data):
    print(f"Assessment failed for business {data['business_id']}")
    print(f"Error: {data['error']['message']}")
    
    # Log error
    # Retry assessment
    # Notify administrators

def handle_prediction_updated(data):
    print(f"Prediction updated for business {data['business_id']}")
    print(f"Predicted Score: {data['predicted_score']} ({data['horizon_months']} months)")
    
    # Update risk dashboard
    # Send alerts if risk increased
    # Update business records

def handle_compliance_alert(data):
    print(f"Compliance alert for business {data['business_id']}")
    print(f"Alert Type: {data['alert_type']}, Severity: {data['severity']}")
    
    # Block transactions
    # Notify compliance team
    # Update risk status

def handle_sanctions_match(data):
    print(f"Sanctions match found for {data['entity_name']}")
    print(f"Match Score: {data['match_score']}, List: {data['sanctions_list']}")
    
    # Block entity
    # Notify compliance team
    # Generate compliance report

def handle_media_alert(data):
    print(f"Media alert for business {data['business_id']}")
    print(f"Article: {data['media_item']['title']}")
    print(f"Source: {data['media_item']['source']}")
    
    # Update risk score
    # Notify stakeholders
    # Generate media report

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=3000)
```

### Go with Gin

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    
    "github.com/gin-gonic/gin"
)

type WebhookEvent struct {
    Event     string      `json:"event"`
    Data      interface{} `json:"data"`
    Timestamp string      `json:"timestamp"`
    WebhookID string      `json:"webhook_id"`
}

type AssessmentCompleted struct {
    ID              string  `json:"id"`
    BusinessID      string  `json:"business_id"`
    RiskScore       float64 `json:"risk_score"`
    RiskLevel       string  `json:"risk_level"`
    ConfidenceScore float64 `json:"confidence_score"`
    Status          string  `json:"status"`
    CreatedAt       string  `json:"created_at"`
}

func verifyWebhookSignature(payload, signature, secret string) bool {
    expectedSignature := hmac.New(sha256.New, []byte(secret))
    expectedSignature.Write([]byte(payload))
    expectedHex := hex.EncodeToString(expectedSignature.Sum(nil))
    
    return hmac.Equal([]byte(signature), []byte(expectedHex))
}

func handleWebhook(c *gin.Context) {
    signature := c.GetHeader("X-Webhook-Signature")
    payload, _ := c.GetRawData()
    
    // Verify signature
    if !verifyWebhookSignature(string(payload), signature, os.Getenv("WEBHOOK_SECRET")) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
        return
    }
    
    // Parse event
    var event WebhookEvent
    if err := json.Unmarshal(payload, &event); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
        return
    }
    
    // Process event
    switch event.Event {
    case "assessment.completed":
        handleAssessmentCompleted(event.Data)
    case "assessment.failed":
        handleAssessmentFailed(event.Data)
    case "prediction.updated":
        handlePredictionUpdated(event.Data)
    case "compliance.alert":
        handleComplianceAlert(event.Data)
    case "sanctions.match":
        handleSanctionsMatch(event.Data)
    case "media.alert":
        handleMediaAlert(event.Data)
    default:
        log.Printf("Unknown event type: %s", event.Event)
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func handleAssessmentCompleted(data interface{}) {
    // Convert to struct
    jsonData, _ := json.Marshal(data)
    var assessment AssessmentCompleted
    json.Unmarshal(jsonData, &assessment)
    
    log.Printf("Assessment completed for business %s", assessment.BusinessID)
    log.Printf("Risk Score: %.2f, Level: %s", assessment.RiskScore, assessment.RiskLevel)
    
    // Update your database
    // Send notifications
    // Trigger downstream processes
}

func handleAssessmentFailed(data interface{}) {
    log.Printf("Assessment failed: %+v", data)
    
    // Log error
    // Retry assessment
    // Notify administrators
}

func handlePredictionUpdated(data interface{}) {
    log.Printf("Prediction updated: %+v", data)
    
    // Update risk dashboard
    // Send alerts if risk increased
    // Update business records
}

func handleComplianceAlert(data interface{}) {
    log.Printf("Compliance alert: %+v", data)
    
    // Block transactions
    // Notify compliance team
    // Update risk status
}

func handleSanctionsMatch(data interface{}) {
    log.Printf("Sanctions match: %+v", data)
    
    // Block entity
    // Notify compliance team
    // Generate compliance report
}

func handleMediaAlert(data interface{}) {
    log.Printf("Media alert: %+v", data)
    
    // Update risk score
    // Notify stakeholders
    // Generate media report
}

func main() {
    r := gin.Default()
    r.POST("/webhooks/risk-assessment", handleWebhook)
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }
    
    log.Printf("Webhook server listening on port %s", port)
    r.Run(":" + port)
}
```

## Webhook Delivery

### Retry Policy

Webhooks are retried with exponential backoff:

- **Initial retry**: 1 second
- **Maximum retries**: 5 attempts
- **Backoff multiplier**: 2x
- **Maximum delay**: 5 minutes

### Delivery Status

Monitor webhook delivery status:

```bash
curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/webhooks/wh_1234567890" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response:**
```json
{
  "webhook_id": "wh_1234567890",
  "url": "https://your-app.com/webhooks/risk-assessment",
  "active": true,
  "delivery_stats": {
    "total_deliveries": 150,
    "successful_deliveries": 145,
    "failed_deliveries": 5,
    "success_rate": 0.967
  },
  "last_delivery": {
    "timestamp": "2024-01-15T10:30:00Z",
    "status": "success",
    "response_time": 250
  }
}
```

### Testing Webhooks

Test your webhook endpoint:

```bash
curl -X POST "https://your-app.com/webhooks/risk-assessment" \
  -H "X-Webhook-Signature: sha256=test_signature" \
  -H "X-Webhook-Timestamp: 1640995200" \
  -H "X-Webhook-Id: wh_test_123" \
  -H "X-Webhook-Event: assessment.completed" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "assessment.completed",
    "data": {
      "id": "risk_test_123",
      "business_id": "biz_test_123",
      "risk_score": 0.75,
      "risk_level": "medium",
      "confidence_score": 0.85,
      "status": "completed",
      "created_at": "2024-01-15T10:30:00Z"
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "webhook_id": "wh_test_123"
  }'
```

## Best Practices

### 1. Idempotency

Make your webhook handlers idempotent to handle duplicate deliveries:

```javascript
const processedEvents = new Set();

function handleWebhook(event) {
  const eventId = `${event.webhook_id}-${event.timestamp}`;
  
  if (processedEvents.has(eventId)) {
    console.log('Event already processed, skipping');
    return;
  }
  
  processedEvents.add(eventId);
  
  // Process event
  processEvent(event);
}
```

### 2. Error Handling

Handle errors gracefully and return appropriate HTTP status codes:

```javascript
app.post('/webhooks/risk-assessment', (req, res) => {
  try {
    // Process webhook
    processWebhook(req.body);
    res.status(200).send('OK');
  } catch (error) {
    console.error('Webhook processing error:', error);
    
    // Return 5xx for retryable errors
    if (isRetryableError(error)) {
      res.status(500).send('Internal Server Error');
    } else {
      // Return 4xx for non-retryable errors
      res.status(400).send('Bad Request');
    }
  }
});
```

### 3. Timeout Handling

Set appropriate timeouts for webhook processing:

```javascript
app.post('/webhooks/risk-assessment', (req, res) => {
  // Set timeout for webhook processing
  const timeout = setTimeout(() => {
    res.status(408).send('Request Timeout');
  }, 5000); // 5 second timeout
  
  try {
    // Process webhook
    processWebhook(req.body);
    clearTimeout(timeout);
    res.status(200).send('OK');
  } catch (error) {
    clearTimeout(timeout);
    res.status(500).send('Internal Server Error');
  }
});
```

### 4. Logging

Log webhook events for debugging and monitoring:

```javascript
app.post('/webhooks/risk-assessment', (req, res) => {
  const event = req.body;
  
  console.log(`Received webhook: ${event.event}`, {
    webhook_id: event.webhook_id,
    timestamp: event.timestamp,
    business_id: event.data.business_id
  });
  
  // Process webhook
  processWebhook(event);
  
  res.status(200).send('OK');
});
```

## Troubleshooting

### Common Issues

#### 1. Webhook Not Receiving Events

**Check:**
- Webhook URL is accessible from the internet
- Webhook is active and properly configured
- Events are subscribed to correctly
- Firewall/security groups allow incoming connections

#### 2. Signature Verification Failing

**Check:**
- Webhook secret is correct
- Payload is not modified before verification
- Signature header format is correct
- HMAC-SHA256 algorithm is used

#### 3. Webhook Timeouts

**Solutions:**
- Process webhooks asynchronously
- Return 200 OK immediately
- Use message queues for processing
- Optimize webhook handler performance

#### 4. Duplicate Events

**Solutions:**
- Implement idempotency checks
- Use event IDs for deduplication
- Store processed event IDs
- Handle duplicate deliveries gracefully

## Support

For webhook-related issues:

- **Email**: [webhooks@kyb-platform.com](mailto:webhooks@kyb-platform.com)
- **Documentation**: [https://docs.kyb-platform.com/webhooks](https://docs.kyb-platform.com/webhooks)
- **Status Page**: [https://status.kyb-platform.com](https://status.kyb-platform.com)

## Changelog

### v2.0.0 (2024-01-15)
- **NEW**: Enhanced webhook security with signature verification
- **NEW**: Timestamp validation for replay attack prevention
- **NEW**: Delivery statistics and monitoring
- **NEW**: Retry policy with exponential backoff
- **NEW**: Webhook testing tools
- **ENHANCED**: Event payload structure
- **ENHANCED**: Error handling and status codes
- **ENHANCED**: Documentation and examples

### v1.0.0 (2024-01-15)
- Initial webhook support
- Basic event types
- Simple delivery mechanism
