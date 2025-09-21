# Stakeholder Feedback Implementation Guide

## Overview

This guide provides step-by-step instructions for implementing the stakeholder feedback collection system for the KYB Platform. The system includes user feedback collection, developer feedback collection, and business impact analysis capabilities.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Feedback Collection System                │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   User Feedback │  │ Developer       │  │ Business     │ │
│  │   Collector     │  │ Feedback        │  │ Impact       │ │
│  │                 │  │ Collector       │  │ Analyzer     │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
│           │                     │                     │      │
│           └─────────────────────┼─────────────────────┘      │
│                                 │                            │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Supabase Feedback Storage                  │ │
│  │  - User feedback table                                 │ │
│  │  - Developer feedback table                            │ │
│  │  - Business impact analysis table                      │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                 │                            │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                HTTP API Handlers                        │ │
│  │  - /api/feedback/user (POST, GET)                      │ │
│  │  - /api/feedback/developer (POST, GET)                 │ │
│  │  - /api/feedback/analysis (GET)                        │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Steps

### Step 1: Database Setup

#### 1.1 Create Supabase Tables

```sql
-- User feedback table
CREATE TABLE user_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    comments TEXT,
    specific_features TEXT[],
    improvement_areas TEXT[],
    classification_accuracy DECIMAL(3,2),
    performance_rating INTEGER CHECK (performance_rating >= 1 AND performance_rating <= 5),
    usability_rating INTEGER CHECK (usability_rating >= 1 AND usability_rating <= 5),
    business_impact JSONB,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB
);

-- Developer feedback table
CREATE TABLE developer_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    developer_id VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    comments TEXT,
    code_quality_rating INTEGER CHECK (code_quality_rating >= 1 AND code_quality_rating <= 5),
    architecture_rating INTEGER CHECK (architecture_rating >= 1 AND architecture_rating <= 5),
    performance_rating INTEGER CHECK (performance_rating >= 1 AND performance_rating <= 5),
    maintainability_rating INTEGER CHECK (maintainability_rating >= 1 AND maintainability_rating <= 5),
    technical_debt_level VARCHAR(20),
    improvement_suggestions TEXT[],
    priority_level VARCHAR(20),
    estimated_effort VARCHAR(50),
    impact_assessment TEXT,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB
);

-- Business impact analysis table
CREATE TABLE business_impact_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_id VARCHAR(255) UNIQUE NOT NULL,
    time_period VARCHAR(50) NOT NULL,
    analysis_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    roi_analysis JSONB,
    cost_savings_analysis JSONB,
    productivity_gains_analysis JSONB,
    quality_improvements_analysis JSONB,
    risk_reduction_analysis JSONB,
    user_satisfaction_analysis JSONB,
    technical_metrics_analysis JSONB,
    business_metrics_analysis JSONB,
    key_recommendations TEXT[],
    next_steps TEXT[]
);

-- Create indexes for better performance
CREATE INDEX idx_user_feedback_category ON user_feedback(category);
CREATE INDEX idx_user_feedback_submitted_at ON user_feedback(submitted_at);
CREATE INDEX idx_developer_feedback_category ON developer_feedback(category);
CREATE INDEX idx_developer_feedback_submitted_at ON developer_feedback(submitted_at);
CREATE INDEX idx_business_impact_analysis_date ON business_impact_analysis(analysis_date);
```

#### 1.2 Set up Row Level Security (RLS)

```sql
-- Enable RLS on all tables
ALTER TABLE user_feedback ENABLE ROW LEVEL SECURITY;
ALTER TABLE developer_feedback ENABLE ROW LEVEL SECURITY;
ALTER TABLE business_impact_analysis ENABLE ROW LEVEL SECURITY;

-- Create policies (adjust based on your authentication system)
CREATE POLICY "Users can view their own feedback" ON user_feedback
    FOR SELECT USING (auth.uid()::text = user_id);

CREATE POLICY "Users can insert their own feedback" ON user_feedback
    FOR INSERT WITH CHECK (auth.uid()::text = user_id);

CREATE POLICY "Developers can view their own feedback" ON developer_feedback
    FOR SELECT USING (auth.uid()::text = developer_id);

CREATE POLICY "Developers can insert their own feedback" ON developer_feedback
    FOR INSERT WITH CHECK (auth.uid()::text = developer_id);

CREATE POLICY "Authenticated users can view business impact analysis" ON business_impact_analysis
    FOR SELECT USING (auth.role() = 'authenticated');
```

### Step 2: Go Backend Implementation

#### 2.1 Project Structure

```
internal/
├── feedback/
│   ├── feedback_types.go              # Type definitions
│   ├── user_feedback_collector.go     # User feedback collection
│   ├── developer_feedback_collector.go # Developer feedback collection
│   ├── business_impact_analyzer.go    # Business impact analysis
│   ├── supabase_feedback_storage.go   # Supabase storage implementation
│   └── feedback_test.go               # Unit tests
├── api/
│   └── handlers/
│       └── feedback_handler.go        # HTTP handlers
└── config/
    └── config.go                      # Configuration management
```

#### 2.2 Configuration Setup

```go
// internal/config/config.go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    SupabaseURL    string
    SupabaseKey    string
    ServerPort     string
    LogLevel       string
    CacheEnabled   bool
    CacheTTL       int
}

func Load() *Config {
    return &Config{
        SupabaseURL:    getEnv("SUPABASE_URL", "http://localhost:54321"),
        SupabaseKey:    getEnv("SUPABASE_ANON_KEY", ""),
        ServerPort:     getEnv("SERVER_PORT", "8080"),
        LogLevel:       getEnv("LOG_LEVEL", "info"),
        CacheEnabled:   getEnvBool("CACHE_ENABLED", true),
        CacheTTL:       getEnvInt("CACHE_TTL", 3600),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if parsed, err := strconv.ParseBool(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if parsed, err := strconv.Atoi(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}
```

#### 2.3 Main Application Setup

```go
// cmd/server/main.go
package main

import (
    "database/sql"
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "github.com/lib/pq"
    "github.com/petercrawford/kyb-platform/internal/api/handlers"
    "github.com/petercrawford/kyb-platform/internal/config"
    "github.com/petercrawford/kyb-platform/internal/feedback"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Setup logger
    logger := log.New(os.Stdout, "KYB-FEEDBACK: ", log.LstdFlags)
    
    // Connect to Supabase
    db, err := sql.Open("postgres", cfg.SupabaseURL)
    if err != nil {
        logger.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Initialize feedback storage
    storage := feedback.NewSupabaseFeedbackStorage(db, logger)
    
    // Initialize collectors
    userCollector := feedback.NewUserFeedbackCollector(storage, logger)
    devCollector := feedback.NewDeveloperFeedbackCollector(storage, logger)
    impactAnalyzer := feedback.NewBusinessImpactAnalyzer(storage, logger)
    
    // Initialize handlers
    feedbackHandler := handlers.NewFeedbackHandler(userCollector, devCollector, impactAnalyzer, logger)
    
    // Setup routes
    router := mux.NewRouter()
    router.HandleFunc("/api/feedback/user", feedbackHandler.HandleUserFeedback).Methods("POST")
    router.HandleFunc("/api/feedback/user", feedbackHandler.GetUserFeedback).Methods("GET")
    router.HandleFunc("/api/feedback/developer", feedbackHandler.HandleDeveloperFeedback).Methods("POST")
    router.HandleFunc("/api/feedback/developer", feedbackHandler.GetDeveloperFeedback).Methods("GET")
    router.HandleFunc("/api/feedback/analysis", feedbackHandler.GetFeedbackAnalysis).Methods("GET")
    router.HandleFunc("/api/feedback/export", feedbackHandler.ExportFeedback).Methods("GET")
    
    // Add middleware
    router.Use(loggingMiddleware(logger))
    router.Use(corsMiddleware())
    
    // Start server
    logger.Printf("Starting server on port %s", cfg.ServerPort)
    if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
        logger.Fatalf("Server failed to start: %v", err)
    }
}

func loggingMiddleware(logger *log.Logger) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
            next.ServeHTTP(w, r)
        })
    }
}

func corsMiddleware() mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Step 3: Frontend Integration

#### 3.1 User Feedback Form

```html
<!-- templates/user_feedback_form.html -->
<!DOCTYPE html>
<html>
<head>
    <title>User Feedback - KYB Platform</title>
    <style>
        .feedback-form {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }
        input, select, textarea {
            width: 100%;
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        .rating {
            display: flex;
            gap: 10px;
        }
        .rating input[type="radio"] {
            width: auto;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        button:hover {
            background-color: #0056b3;
        }
    </style>
</head>
<body>
    <div class="feedback-form">
        <h2>User Feedback Form</h2>
        <form id="feedbackForm">
            <div class="form-group">
                <label for="category">Category:</label>
                <select id="category" name="category" required>
                    <option value="">Select a category</option>
                    <option value="database_performance">Database Performance</option>
                    <option value="classification_accuracy">Classification Accuracy</option>
                    <option value="user_experience">User Experience</option>
                    <option value="risk_detection">Risk Detection</option>
                    <option value="overall_satisfaction">Overall Satisfaction</option>
                </select>
            </div>
            
            <div class="form-group">
                <label>Overall Rating:</label>
                <div class="rating">
                    <label><input type="radio" name="rating" value="1"> 1</label>
                    <label><input type="radio" name="rating" value="2"> 2</label>
                    <label><input type="radio" name="rating" value="3"> 3</label>
                    <label><input type="radio" name="rating" value="4"> 4</label>
                    <label><input type="radio" name="rating" value="5"> 5</label>
                </div>
            </div>
            
            <div class="form-group">
                <label for="comments">Comments:</label>
                <textarea id="comments" name="comments" rows="4" placeholder="Please provide detailed feedback..."></textarea>
            </div>
            
            <div class="form-group">
                <label for="performance_rating">Performance Rating:</label>
                <div class="rating">
                    <label><input type="radio" name="performance_rating" value="1"> 1</label>
                    <label><input type="radio" name="performance_rating" value="2"> 2</label>
                    <label><input type="radio" name="performance_rating" value="3"> 3</label>
                    <label><input type="radio" name="performance_rating" value="4"> 4</label>
                    <label><input type="radio" name="performance_rating" value="5"> 5</label>
                </div>
            </div>
            
            <div class="form-group">
                <label for="usability_rating">Usability Rating:</label>
                <div class="rating">
                    <label><input type="radio" name="usability_rating" value="1"> 1</label>
                    <label><input type="radio" name="usability_rating" value="2"> 2</label>
                    <label><input type="radio" name="usability_rating" value="3"> 3</label>
                    <label><input type="radio" name="usability_rating" value="4"> 4</label>
                    <label><input type="radio" name="usability_rating" value="5"> 5</label>
                </div>
            </div>
            
            <button type="submit">Submit Feedback</button>
        </form>
    </div>

    <script>
        document.getElementById('feedbackForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const feedback = {
                user_id: 'current-user-id', // Replace with actual user ID
                category: formData.get('category'),
                rating: parseInt(formData.get('rating')),
                comments: formData.get('comments'),
                performance_rating: parseInt(formData.get('performance_rating')),
                usability_rating: parseInt(formData.get('usability_rating')),
                business_impact: {
                    time_saved: 30, // Example value
                    cost_reduction: 25.0,
                    error_reduction: 40.0,
                    productivity_gain: 35.0,
                    overall_roi: "high"
                }
            };
            
            try {
                const response = await fetch('/api/feedback/user', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(feedback)
                });
                
                if (response.ok) {
                    alert('Feedback submitted successfully!');
                    this.reset();
                } else {
                    alert('Failed to submit feedback. Please try again.');
                }
            } catch (error) {
                console.error('Error:', error);
                alert('An error occurred. Please try again.');
            }
        });
    </script>
</body>
</html>
```

### Step 4: Testing & Validation

#### 4.1 Unit Tests

```bash
# Run unit tests
go test ./internal/feedback/... -v

# Run with coverage
go test ./internal/feedback/... -cover

# Run integration tests
go test ./test/integration/... -v
```

#### 4.2 API Testing

```bash
# Test user feedback submission
curl -X POST http://localhost:8080/api/feedback/user \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user-123",
    "category": "database_performance",
    "rating": 5,
    "comments": "Great performance improvements!",
    "performance_rating": 5,
    "usability_rating": 4,
    "business_impact": {
      "time_saved": 30,
      "cost_reduction": 25.0,
      "error_reduction": 40.0,
      "productivity_gain": 35.0,
      "overall_roi": "high"
    }
  }'

# Test feedback analysis retrieval
curl http://localhost:8080/api/feedback/analysis

# Test feedback export
curl http://localhost:8080/api/feedback/export?format=json&category=database_performance
```

### Step 5: Deployment

#### 5.1 Environment Variables

```bash
# .env file
SUPABASE_URL=your_supabase_url
SUPABASE_ANON_KEY=your_supabase_anon_key
SERVER_PORT=8080
LOG_LEVEL=info
CACHE_ENABLED=true
CACHE_TTL=3600
```

#### 5.2 Docker Deployment

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o feedback-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/feedback-server .
COPY --from=builder /app/templates ./templates

EXPOSE 8080
CMD ["./feedback-server"]
```

#### 5.3 Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  feedback-server:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SUPABASE_URL=${SUPABASE_URL}
      - SUPABASE_ANON_KEY=${SUPABASE_ANON_KEY}
      - SERVER_PORT=8080
      - LOG_LEVEL=info
    depends_on:
      - redis
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    restart: unless-stopped
```

## Monitoring & Maintenance

### 6.1 Health Checks

```go
// Add health check endpoint
func (h *FeedbackHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now(),
        "version":   "1.0.0",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

### 6.2 Metrics Collection

```go
// Add metrics collection
import "github.com/prometheus/client_golang/prometheus"

var (
    feedbackSubmissions = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "feedback_submissions_total",
            Help: "Total number of feedback submissions",
        },
        []string{"category", "rating"},
    )
    
    feedbackProcessingTime = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "feedback_processing_duration_seconds",
            Help: "Time spent processing feedback",
        },
        []string{"category"},
    )
)

func init() {
    prometheus.MustRegister(feedbackSubmissions)
    prometheus.MustRegister(feedbackProcessingTime)
}
```

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   - Check Supabase URL and credentials
   - Verify network connectivity
   - Check database permissions

2. **Performance Issues**
   - Monitor query performance
   - Check index usage
   - Review connection pool settings

3. **Authentication Issues**
   - Verify RLS policies
   - Check user permissions
   - Review authentication flow

### Logging

```go
// Enhanced logging
func (h *FeedbackHandler) HandleUserFeedback(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    h.logger.Printf("Processing user feedback request from %s", r.RemoteAddr)
    
    // ... processing logic ...
    
    duration := time.Since(start)
    h.logger.Printf("User feedback processed in %v", duration)
    
    // Record metrics
    feedbackProcessingTime.WithLabelValues(feedback.Category).Observe(duration.Seconds())
    feedbackSubmissions.WithLabelValues(string(feedback.Category), strconv.Itoa(feedback.Rating)).Inc()
}
```

## Conclusion

This implementation guide provides a comprehensive framework for deploying the stakeholder feedback collection system. The system is designed to be scalable, maintainable, and provide valuable insights for continuous platform improvement.

**Key Benefits:**
- Structured feedback collection from all stakeholders
- Comprehensive business impact analysis
- Real-time feedback processing and analysis
- Export capabilities for further analysis
- Integration with existing KYB Platform infrastructure

**Next Steps:**
1. Deploy the system in a staging environment
2. Conduct user acceptance testing
3. Train users on the feedback system
4. Monitor system performance and usage
5. Iterate based on initial feedback

---

**Document Version:** 1.0  
**Last Updated:** December 19, 2024  
**Next Review:** January 19, 2025
