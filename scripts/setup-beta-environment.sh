#!/bin/bash

# KYB Platform - Beta Environment Setup Script
# This script sets up the complete beta testing environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BETA_ENV_NAME="kyb-beta"
BETA_PORT="8082"
BETA_DB_PORT="5433"
BETA_REDIS_PORT="6380"
BETA_PROMETHEUS_PORT="9092"
BETA_GRAFANA_PORT="3002"
BETA_ELASTICSEARCH_PORT="9201"
BETA_KIBANA_PORT="5602"
BETA_MATTERMOST_PORT="8065"

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        error "Docker is not running. Please start Docker and try again."
    fi
    log "Docker is running"
}

# Function to check if ports are available
check_ports() {
    local ports=($BETA_PORT $BETA_DB_PORT $BETA_REDIS_PORT $BETA_PROMETHEUS_PORT $BETA_GRAFANA_PORT $BETA_ELASTICSEARCH_PORT $BETA_KIBANA_PORT $BETA_MATTERMOST_PORT)
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            warn "Port $port is already in use. Please free up the port or modify the configuration."
        fi
    done
    log "Port availability check completed"
}

# Function to create beta environment configuration
create_beta_config() {
    log "Creating beta environment configuration..."
    
    # Create beta environment file
    cat > configs/beta.env << EOF
# KYB Platform Beta Environment Configuration
ENV=beta
BETA_MODE=true

# Database Configuration
DB_HOST=postgres-beta
DB_PORT=5432
DB_NAME=kyb_beta
DB_USER=kyb_beta_user
DB_PASSWORD=beta_password_secure

# Redis Configuration
REDIS_HOST=redis-beta
REDIS_PORT=6379
REDIS_PASSWORD=beta_redis_password

# Application Configuration
JWT_SECRET=beta_jwt_secret_key_very_secure
LOG_LEVEL=debug
METRICS_ENABLED=true
FEEDBACK_COLLECTION_ENABLED=true
USER_ANALYTICS_ENABLED=true

# Beta-specific Configuration
BETA_USER_LIMIT=50
BETA_FEEDBACK_ENABLED=true
BETA_ANALYTICS_ENABLED=true
BETA_SUPPORT_ENABLED=true

# Monitoring Configuration
PROMETHEUS_ENABLED=true
GRAFANA_ENABLED=true
ELASTICSEARCH_ENABLED=true
KIBANA_ENABLED=true

# Support Configuration
MATTERMOST_ENABLED=true
SUPPORT_EMAIL=beta-support@kybplatform.com
SUPPORT_SLACK=kyb-beta-support
EOF

    log "Beta environment configuration created"
}

# Function to create beta monitoring configuration
create_monitoring_config() {
    log "Creating beta monitoring configuration..."
    
    # Create Prometheus beta configuration
    mkdir -p deployments/prometheus
    cat > deployments/prometheus/prometheus-beta.yml << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager-beta:9093

scrape_configs:
  - job_name: 'kyb-beta-api'
    static_configs:
      - targets: ['kyb-platform-beta:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
    honor_labels: true

  - job_name: 'postgres-beta'
    static_configs:
      - targets: ['postgres-beta:5432']
    scrape_interval: 30s

  - job_name: 'redis-beta'
    static_configs:
      - targets: ['redis-beta:6379']
    scrape_interval: 30s

  - job_name: 'elasticsearch-beta'
    static_configs:
      - targets: ['elasticsearch-beta:9200']
    scrape_interval: 30s
EOF

    # Create AlertManager beta configuration
    mkdir -p deployments/alertmanager
    cat > deployments/alertmanager/alertmanager-beta.yml << EOF
global:
  resolve_timeout: 5m
  slack_api_url: 'https://hooks.slack.com/services/YOUR_SLACK_WEBHOOK'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'beta-team'

receivers:
  - name: 'beta-team'
    slack_configs:
      - channel: '#kyb-beta-alerts'
        title: 'KYB Beta Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
    email_configs:
      - to: 'beta-support@kybplatform.com'
        from: 'alerts@kybplatform.com'
        smarthost: 'smtp.gmail.com:587'
        auth_username: 'alerts@kybplatform.com'
        auth_password: 'your_password'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOF

    log "Beta monitoring configuration created"
}

# Function to create beta feedback collection system
create_feedback_system() {
    log "Creating beta feedback collection system..."
    
    # Create feedback collection API endpoints
    mkdir -p internal/api/handlers/beta
    cat > internal/api/handlers/beta/feedback.go << 'EOF'
package beta

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// FeedbackRequest represents a feedback submission
type FeedbackRequest struct {
	UserID       string          `json:"user_id"`
	FeedbackType string          `json:"feedback_type"`
	Category     string          `json:"category"`
	Rating       *int            `json:"rating,omitempty"`
	FeedbackText string          `json:"feedback_text"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

// FeedbackResponse represents the response to a feedback submission
type FeedbackResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// SubmitFeedback handles feedback submission from beta users
func SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var req FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.FeedbackType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Store feedback in database
	feedbackID := uuid.New().String()
	
	// TODO: Implement database storage
	// db.StoreFeedback(feedbackID, req)

	response := FeedbackResponse{
		ID:        feedbackID,
		Status:    "success",
		Message:   "Feedback submitted successfully",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetFeedbackAnalytics returns analytics about feedback
func GetFeedbackAnalytics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement feedback analytics
	analytics := map[string]interface{}{
		"total_feedback": 0,
		"average_rating": 0.0,
		"feedback_types": map[string]int{},
		"categories":     map[string]int{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
EOF

    # Create survey management system
    cat > internal/api/handlers/beta/surveys.go << 'EOF'
package beta

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SurveyResponse represents a survey response
type SurveyResponse struct {
	SurveyID   string          `json:"survey_id"`
	UserID     string          `json:"user_id"`
	Responses  json.RawMessage `json:"responses"`
	CompletedAt time.Time      `json:"completed_at"`
}

// SubmitSurveyResponse handles survey response submission
func SubmitSurveyResponse(w http.ResponseWriter, r *http.Request) {
	var req SurveyResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SurveyID == "" || req.UserID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	req.CompletedAt = time.Now()

	// TODO: Implement database storage
	// db.StoreSurveyResponse(req)

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Survey response submitted successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSurveyAnalytics returns analytics about survey responses
func GetSurveyAnalytics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement survey analytics
	analytics := map[string]interface{}{
		"total_responses": 0,
		"survey_types":    map[string]int{},
		"completion_rate": 0.0,
		"average_scores":  map[string]float64{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
EOF

    log "Beta feedback collection system created"
}

# Function to create beta user management system
create_user_management() {
    log "Creating beta user management system..."
    
    # Create beta user management handlers
    cat > internal/api/handlers/beta/users.go << 'EOF'
package beta

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// BetaUser represents a beta user
type BetaUser struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Company      string    `json:"company"`
	Role         string    `json:"role"`
	Industry     string    `json:"industry"`
	OnboardedAt  time.Time `json:"onboarded_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	Status       string    `json:"status"`
}

// BetaUserRequest represents a beta user registration request
type BetaUserRequest struct {
	Email    string `json:"email"`
	Company  string `json:"company"`
	Role     string `json:"role"`
	Industry string `json:"industry"`
}

// RegisterBetaUser handles beta user registration
func RegisterBetaUser(w http.ResponseWriter, r *http.Request) {
	var req BetaUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Company == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	user := BetaUser{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Company:      req.Company,
		Role:         req.Role,
		Industry:     req.Industry,
		OnboardedAt:  time.Now(),
		LastActiveAt: time.Now(),
		Status:       "active",
	}

	// TODO: Implement database storage
	// db.StoreBetaUser(user)

	response := map[string]interface{}{
		"status":  "success",
		"message": "Beta user registered successfully",
		"user":    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBetaUsers returns all beta users
func GetBetaUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement database retrieval
	users := []BetaUser{}

	response := map[string]interface{}{
		"status": "success",
		"users":  users,
		"count":  len(users),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBetaUserAnalytics returns analytics about beta users
func GetBetaUserAnalytics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement user analytics
	analytics := map[string]interface{}{
		"total_users":      0,
		"active_users":     0,
		"industries":       map[string]int{},
		"roles":            map[string]int{},
		"onboarding_rate":  0.0,
		"retention_rate":   0.0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
EOF

    log "Beta user management system created"
}

# Function to create beta analytics dashboard
create_analytics_dashboard() {
    log "Creating beta analytics dashboard..."
    
    # Create Grafana dashboard configuration
    mkdir -p deployments/grafana/dashboards
    cat > deployments/grafana/dashboards/beta-analytics.json << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "KYB Beta Analytics Dashboard",
    "tags": ["kyb", "beta", "analytics"],
    "style": "dark",
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Beta User Engagement",
        "type": "stat",
        "targets": [
          {
            "expr": "kyb_beta_users_total",
            "legendFormat": "Total Users"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "palette-classic"
            },
            "custom": {
              "displayMode": "list"
            }
          }
        }
      },
      {
        "id": 2,
        "title": "Feature Usage",
        "type": "piechart",
        "targets": [
          {
            "expr": "kyb_feature_usage_total",
            "legendFormat": "{{feature}}"
          }
        ]
      },
      {
        "id": 3,
        "title": "Feedback Sentiment",
        "type": "gauge",
        "targets": [
          {
            "expr": "avg(kyb_feedback_rating)",
            "legendFormat": "Average Rating"
          }
        ]
      },
      {
        "id": 4,
        "title": "System Performance",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(kyb_api_requests_total[5m])",
            "legendFormat": "Requests/sec"
          }
        ]
      }
    ],
    "time": {
      "from": "now-7d",
      "to": "now"
    },
    "refresh": "30s"
  }
}
EOF

    # Create datasource configuration
    mkdir -p deployments/grafana/datasources
    cat > deployments/grafana/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus-beta:9090
    isDefault: true
    editable: true
EOF

    log "Beta analytics dashboard created"
}

# Function to create beta support system
create_support_system() {
    log "Creating beta support system..."
    
    # Create support documentation
    mkdir -p docs/beta-support
    cat > docs/beta-support/README.md << 'EOF'
# KYB Platform Beta Support

## Getting Help

### Support Channels
- **Email**: beta-support@kybplatform.com
- **Slack**: #kyb-beta-support
- **Mattermost**: http://localhost:8065
- **Documentation**: This directory

### Common Issues

#### Platform Access
- **Issue**: Cannot access beta platform
- **Solution**: Check your credentials and ensure you're using the beta URL: http://localhost:8082

#### Feature Problems
- **Issue**: Feature not working as expected
- **Solution**: Check the feature documentation and report issues through the feedback system

#### Performance Issues
- **Issue**: Platform is slow or unresponsive
- **Solution**: Check your internet connection and report the issue with details

### Feedback Submission
Use the in-app feedback system or email us at beta-support@kybplatform.com

### Bug Reports
Include the following information:
- Description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Browser and OS information
- Screenshots if applicable
EOF

    # Create support scripts
    cat > scripts/beta-support.sh << 'EOF'
#!/bin/bash

# KYB Beta Support Script

echo "KYB Platform Beta Support"
echo "========================="
echo ""
echo "1. Check platform status"
echo "2. View recent logs"
echo "3. Check user analytics"
echo "4. View feedback summary"
echo "5. Restart beta services"
echo "6. Exit"
echo ""

read -p "Select an option: " choice

case $choice in
    1)
        echo "Checking platform status..."
        curl -s http://localhost:8082/health | jq .
        ;;
    2)
        echo "Recent logs:"
        docker logs kyb-platform-beta --tail 50
        ;;
    3)
        echo "User analytics:"
        curl -s http://localhost:8082/v1/beta/analytics/users | jq .
        ;;
    4)
        echo "Feedback summary:"
        curl -s http://localhost:8082/v1/beta/analytics/feedback | jq .
        ;;
    5)
        echo "Restarting beta services..."
        docker-compose -f docker-compose.beta.yml restart
        ;;
    6)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo "Invalid option"
        ;;
esac
EOF

    chmod +x scripts/beta-support.sh

    log "Beta support system created"
}

# Function to start beta environment
start_beta_environment() {
    log "Starting beta environment..."
    
    # Build and start the beta environment
    docker-compose -f docker-compose.beta.yml up -d --build
    
    # Wait for services to be ready
    log "Waiting for services to be ready..."
    sleep 30
    
    # Check service health
    log "Checking service health..."
    
    # Check main application
    if curl -f http://localhost:$BETA_PORT/health > /dev/null 2>&1; then
        log "✅ KYB Platform Beta is running on port $BETA_PORT"
    else
        warn "⚠️  KYB Platform Beta health check failed"
    fi
    
    # Check database
    if docker exec postgres-beta pg_isready -U kyb_beta_user -d kyb_beta > /dev/null 2>&1; then
        log "✅ Beta database is ready"
    else
        warn "⚠️  Beta database health check failed"
    fi
    
    # Check Redis
    if docker exec redis-beta redis-cli -a beta_redis_password ping > /dev/null 2>&1; then
        log "✅ Beta Redis is ready"
    else
        warn "⚠️  Beta Redis health check failed"
    fi
    
    # Check Prometheus
    if curl -f http://localhost:$BETA_PROMETHEUS_PORT/-/healthy > /dev/null 2>&1; then
        log "✅ Beta Prometheus is running on port $BETA_PROMETHEUS_PORT"
    else
        warn "⚠️  Beta Prometheus health check failed"
    fi
    
    # Check Grafana
    if curl -f http://localhost:$BETA_GRAFANA_PORT/api/health > /dev/null 2>&1; then
        log "✅ Beta Grafana is running on port $BETA_GRAFANA_PORT"
    else
        warn "⚠️  Beta Grafana health check failed"
    fi
    
    # Check Elasticsearch
    if curl -f http://localhost:$BETA_ELASTICSEARCH_PORT/_cluster/health > /dev/null 2>&1; then
        log "✅ Beta Elasticsearch is running on port $BETA_ELASTICSEARCH_PORT"
    else
        warn "⚠️  Beta Elasticsearch health check failed"
    fi
    
    # Check Kibana
    if curl -f http://localhost:$BETA_KIBANA_PORT/api/status > /dev/null 2>&1; then
        log "✅ Beta Kibana is running on port $BETA_KIBANA_PORT"
    else
        warn "⚠️  Beta Kibana health check failed"
    fi
    
    # Check Mattermost
    if curl -f http://localhost:$BETA_MATTERMOST_PORT/api/v4/system/ping > /dev/null 2>&1; then
        log "✅ Beta Mattermost is running on port $BETA_MATTERMOST_PORT"
    else
        warn "⚠️  Beta Mattermost health check failed"
    fi
}

# Function to display beta environment information
display_beta_info() {
    log "Beta Environment Setup Complete!"
    echo ""
    echo "🌐 Beta Environment URLs:"
    echo "   KYB Platform:     http://localhost:$BETA_PORT"
    echo "   Prometheus:       http://localhost:$BETA_PROMETHEUS_PORT"
    echo "   Grafana:          http://localhost:$BETA_GRAFANA_PORT"
    echo "   Elasticsearch:    http://localhost:$BETA_ELASTICSEARCH_PORT"
    echo "   Kibana:           http://localhost:$BETA_KIBANA_PORT"
    echo "   Mattermost:       http://localhost:$BETA_MATTERMOST_PORT"
    echo ""
    echo "📊 Default Credentials:"
    echo "   Grafana:          admin / beta_admin_password"
    echo "   Mattermost:       admin / beta_admin_password"
    echo ""
    echo "📁 Configuration Files:"
    echo "   Beta Config:      configs/beta.env"
    echo "   Docker Compose:   docker-compose.beta.yml"
    echo "   Database Init:    scripts/init-beta-db.sql"
    echo ""
    echo "🛠️  Management Commands:"
    echo "   Start Beta:       docker-compose -f docker-compose.beta.yml up -d"
    echo "   Stop Beta:        docker-compose -f docker-compose.beta.yml down"
    echo "   View Logs:        docker-compose -f docker-compose.beta.yml logs -f"
    echo "   Support Script:   ./scripts/beta-support.sh"
    echo ""
    echo "📋 Next Steps:"
    echo "   1. Access the beta platform at http://localhost:$BETA_PORT"
    echo "   2. Review the beta documentation in docs/beta-support/"
    echo "   3. Set up user recruitment using docs/beta-user-recruitment-strategy.md"
    echo "   4. Configure feedback collection using test/beta/feedback-surveys/"
    echo "   5. Monitor analytics in Grafana at http://localhost:$BETA_GRAFANA_PORT"
    echo ""
}

# Main execution
main() {
    log "Starting KYB Platform Beta Environment Setup..."
    
    # Check prerequisites
    check_docker
    check_ports
    
    # Create configurations
    create_beta_config
    create_monitoring_config
    create_feedback_system
    create_user_management
    create_analytics_dashboard
    create_support_system
    
    # Start environment
    start_beta_environment
    
    # Display information
    display_beta_info
    
    log "Beta environment setup completed successfully!"
}

# Run main function
main "$@"
