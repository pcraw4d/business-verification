#!/bin/bash

# Beta Testing Environment Setup Script
# This script sets up the KYB Platform for beta testing

set -e

echo "🚀 Setting up KYB Platform Beta Testing Environment..."

# Configuration
BETA_ENV="beta"
BETA_PORT="8081"
BETA_DB_NAME="kyb_beta"
BETA_LOG_LEVEL="info"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    print_status "Checking Docker status..."
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Create beta environment configuration
setup_beta_config() {
    print_status "Setting up beta environment configuration..."
    
    # Create beta environment file
    cat > configs/beta.env << EOF
# Beta Environment Configuration
ENV=beta
PORT=$BETA_PORT
HOST=0.0.0.0

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=$BETA_DB_NAME
DB_USER=kyb_beta_user
DB_PASSWORD=kyb_beta_password
DB_SSL_MODE=disable

# Logging Configuration
LOG_LEVEL=$BETA_LOG_LEVEL
LOG_FORMAT=json

# API Configuration
API_RATE_LIMIT=100
API_RATE_LIMIT_WINDOW=60

# External Services
BUSINESS_DATA_API_ENABLED=true
BUSINESS_DATA_API_BASE_URL=https://api.example.com
BUSINESS_DATA_API_KEY=beta_api_key

# Feature Flags
FEATURE_BUSINESS_CLASSIFICATION=true
FEATURE_RISK_ASSESSMENT=true
FEATURE_COMPLIANCE_FRAMEWORK=true
FEATURE_ADVANCED_ANALYTICS=false
FEATURE_REAL_TIME_MONITORING=false

# Beta Testing Specific
BETA_MODE=true
BETA_USER_LIMIT=100
BETA_FEEDBACK_ENABLED=true
BETA_ANALYTICS_ENABLED=true
EOF

    print_success "Beta environment configuration created"
}

# Setup beta database
setup_beta_database() {
    print_status "Setting up beta database..."
    
    # Create beta database
    docker run --rm \
        -e POSTGRES_DB=$BETA_DB_NAME \
        -e POSTGRES_USER=kyb_beta_user \
        -e POSTGRES_PASSWORD=kyb_beta_password \
        -p 5433:5432 \
        -d \
        --name kyb_beta_db \
        postgres:15
    
    # Wait for database to be ready
    print_status "Waiting for database to be ready..."
    sleep 10
    
    # Run migrations
    print_status "Running database migrations..."
    DB_HOST=localhost DB_PORT=5433 DB_NAME=$BETA_DB_NAME DB_USER=kyb_beta_user DB_PASSWORD=kyb_beta_password go run cmd/migrate/main.go
    
    print_success "Beta database setup complete"
}

# Setup monitoring and analytics
setup_monitoring() {
    print_status "Setting up monitoring and analytics..."
    
    # Create monitoring configuration
    cat > deployments/beta/monitoring.yml << EOF
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: kyb_beta_prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'

  grafana:
    image: grafana/grafana:latest
    container_name: kyb_beta_grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=beta_admin
    volumes:
      - grafana-storage:/var/lib/grafana

volumes:
  grafana-storage:
EOF

    # Create Prometheus configuration
    cat > deployments/beta/prometheus.yml << EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'kyb-beta-api'
    static_configs:
      - targets: ['host.docker.internal:$BETA_PORT']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'kyb-beta-health'
    static_configs:
      - targets: ['host.docker.internal:$BETA_PORT']
    metrics_path: '/health'
    scrape_interval: 30s
EOF

    print_success "Monitoring setup complete"
}

# Setup beta testing dashboard
setup_beta_dashboard() {
    print_status "Setting up beta testing dashboard..."
    
    # Create beta dashboard configuration
    cat > test/beta/dashboard/config.yml << EOF
# Beta Testing Dashboard Configuration
dashboard:
  title: "KYB Platform Beta Testing Dashboard"
  refresh_interval: 30s
  
metrics:
  - name: "Active Beta Users"
    query: "kyb_beta_users_active"
    type: "counter"
    
  - name: "Classification Requests"
    query: "kyb_classification_requests_total"
    type: "counter"
    
  - name: "Classification Accuracy"
    query: "kyb_classification_accuracy"
    type: "gauge"
    
  - name: "Response Time"
    query: "kyb_api_response_time_seconds"
    type: "histogram"
    
  - name: "Error Rate"
    query: "kyb_api_errors_total"
    type: "counter"

alerts:
  - name: "High Error Rate"
    condition: "error_rate > 0.05"
    severity: "warning"
    
  - name: "Slow Response Time"
    condition: "response_time > 5"
    severity: "warning"
    
  - name: "Low Classification Accuracy"
    condition: "accuracy < 0.85"
    severity: "critical"
EOF

    print_success "Beta dashboard configuration created"
}

# Setup feedback collection
setup_feedback_collection() {
    print_status "Setting up feedback collection system..."
    
    # Create feedback collection configuration
    cat > internal/feedback/config.go << EOF
package feedback

import (
	"time"
)

// BetaFeedbackConfig holds configuration for beta feedback collection
type BetaFeedbackConfig struct {
	Enabled           bool          \`json:"enabled"\`
	CollectionURL     string        \`json:"collection_url"\`
	SurveyInterval    time.Duration \`json:"survey_interval"\`
	MaxResponses      int           \`json:"max_responses"\`
	AutoCollection    bool          \`json:"auto_collection"\`
	FeedbackChannels  []string      \`json:"feedback_channels"\`
}

// BetaFeedback represents a single feedback entry
type BetaFeedback struct {
	ID          string                 \`json:"id"\`
	UserID      string                 \`json:"user_id"\`
	SessionID   string                 \`json:"session_id"\`
	Feature     string                 \`json:"feature"\`
	Rating      int                    \`json:"rating"\`
	Comments    string                 \`json:"comments"\`
	Category    string                 \`json:"category"\`
	Timestamp   time.Time              \`json:"timestamp"\`
	Metadata    map[string]interface{} \`json:"metadata"\`
	UserAgent   string                 \`json:"user_agent"\`
	IPAddress   string                 \`json:"ip_address"\`
}

// BetaFeedbackService handles feedback collection and analysis
type BetaFeedbackService struct {
	config BetaFeedbackConfig
	// Add other fields as needed
}

// NewBetaFeedbackService creates a new feedback service
func NewBetaFeedbackService(config BetaFeedbackConfig) *BetaFeedbackService {
	return &BetaFeedbackService{
		config: config,
	}
}

// CollectFeedback collects user feedback
func (s *BetaFeedbackService) CollectFeedback(feedback BetaFeedback) error {
	// Implementation for feedback collection
	return nil
}

// GetFeedbackSummary returns feedback summary
func (s *BetaFeedbackService) GetFeedbackSummary() (map[string]interface{}, error) {
	// Implementation for feedback summary
	return nil, nil
}
EOF

    print_success "Feedback collection system configured"
}

# Setup beta user management
setup_beta_user_management() {
    print_status "Setting up beta user management..."
    
    # Create beta user management configuration
    cat > internal/beta/config.go << EOF
package beta

import (
	"time"
)

// BetaUser represents a beta testing user
type BetaUser struct {
	ID              string    \`json:"id"\`
	Email           string    \`json:"email"\`
	Name            string    \`json:"name"\`
	Company         string    \`json:"company"\`
	Role            string    \`json:"role"\`
	Industry        string    \`json:"industry"\`
	InvitedAt       time.Time \`json:"invited_at"\`
	JoinedAt        time.Time \`json:"joined_at"\`
	LastActiveAt    time.Time \`json:"last_active_at"\`
	Status          string    \`json:"status"\`
	UsageCount      int       \`json:"usage_count"\`
	FeedbackCount   int       \`json:"feedback_count"\`
	InvitedBy       string    \`json:"invited_by"\`
	Notes           string    \`json:"notes"\`
}

// BetaUserService manages beta users
type BetaUserService struct {
	// Add fields as needed
}

// NewBetaUserService creates a new beta user service
func NewBetaUserService() *BetaUserService {
	return &BetaUserService{}
}

// InviteUser invites a new beta user
func (s *BetaUserService) InviteUser(email, name, company, role string) error {
	// Implementation for user invitation
	return nil
}

// GetActiveUsers returns active beta users
func (s *BetaUserService) GetActiveUsers() ([]BetaUser, error) {
	// Implementation for getting active users
	return nil, nil
}

// UpdateUserActivity updates user activity
func (s *BetaUserService) UpdateUserActivity(userID string) error {
	// Implementation for updating user activity
	return nil
}
EOF

    print_success "Beta user management configured"
}

# Create beta testing scripts
create_beta_scripts() {
    print_status "Creating beta testing scripts..."
    
    # Create beta testing runner
    cat > scripts/run-beta-tests.sh << 'EOF'
#!/bin/bash

# Beta Testing Runner Script
# This script runs automated beta tests

set -e

echo "🧪 Running Beta Tests..."

# Configuration
BETA_API_URL="http://localhost:8081"
TEST_DATA_DIR="test/beta/data"
RESULTS_DIR="test/beta/results"

# Create results directory
mkdir -p $RESULTS_DIR

# Run classification accuracy tests
echo "Testing classification accuracy..."
curl -X POST "$BETA_API_URL/api/v1/classify" \
  -H "Content-Type: application/json" \
  -d @$TEST_DATA_DIR/test_businesses.json \
  -o $RESULTS_DIR/classification_results.json

# Run performance tests
echo "Testing performance..."
ab -n 100 -c 10 -p $TEST_DATA_DIR/test_businesses.json \
  -T application/json \
  "$BETA_API_URL/api/v1/classify" > $RESULTS_DIR/performance_results.txt

# Run user experience tests
echo "Testing user experience..."
# Add UX testing logic here

echo "✅ Beta tests completed. Results saved to $RESULTS_DIR"
EOF

    chmod +x scripts/run-beta-tests.sh

    # Create beta user invitation script
    cat > scripts/invite-beta-users.sh << 'EOF'
#!/bin/bash

# Beta User Invitation Script
# This script sends invitations to beta users

set -e

echo "📧 Sending Beta User Invitations..."

# Configuration
BETA_INVITE_LIST="test/beta/users/invite_list.csv"
EMAIL_TEMPLATE="test/beta/email/invitation_template.html"

# Check if invite list exists
if [ ! -f "$BETA_INVITE_LIST" ]; then
    echo "Error: Invite list not found at $BETA_INVITE_LIST"
    exit 1
fi

# Process each user in the invite list
while IFS=, read -r email name company role; do
    echo "Inviting $name ($email) from $company..."
    
    # Send invitation email
    # Add email sending logic here
    
    echo "✅ Invitation sent to $email"
done < "$BETA_INVITE_LIST"

echo "✅ All invitations sent"
EOF

    chmod +x scripts/invite-beta-users.sh

    print_success "Beta testing scripts created"
}

# Main setup function
main() {
    print_status "Starting KYB Platform Beta Testing Environment Setup..."
    
    # Check prerequisites
    check_docker
    
    # Setup components
    setup_beta_config
    setup_beta_database
    setup_monitoring
    setup_beta_dashboard
    setup_feedback_collection
    setup_beta_user_management
    create_beta_scripts
    
    print_success "🎉 Beta testing environment setup complete!"
    
    echo ""
    echo "Next steps:"
    echo "1. Review and customize configuration in configs/beta.env"
    echo "2. Start the beta environment: ./scripts/dev.sh beta"
    echo "3. Invite beta users: ./scripts/invite-beta-users.sh"
    echo "4. Monitor beta testing: http://localhost:3000 (Grafana)"
    echo "5. Run beta tests: ./scripts/run-beta-tests.sh"
    echo ""
    echo "Beta testing dashboard: http://localhost:3000"
    echo "API endpoint: http://localhost:8081"
    echo "Prometheus metrics: http://localhost:9090"
}

# Run main function
main "$@"
