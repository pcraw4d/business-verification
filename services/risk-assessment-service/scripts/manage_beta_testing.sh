#!/bin/bash

# Beta Testing Management Script
# This script manages the beta testing program for the Risk Assessment Service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BETA_SERVICE_URL="https://risk-assessment-service-production.up.railway.app"
BETA_MANAGER_URL="http://localhost:8081"
FEEDBACK_COLLECTOR_URL="http://localhost:8080"
INVITATION_SYSTEM_URL="http://localhost:8082"

echo -e "${BLUE}🧪 Beta Testing Management Script${NC}"
echo "================================================================"

# Function to check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"
    
    # Check if required tools are installed
    local required_tools=("curl" "jq" "go")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            echo -e "${RED}❌ $tool is not installed${NC}"
            exit 1
        fi
    done
    
    # Check if beta services are running
    if ! curl -s "$BETA_SERVICE_URL/health" &> /dev/null; then
        echo -e "${YELLOW}⚠️  Beta service is not running at $BETA_SERVICE_URL${NC}"
    fi
    
    echo -e "${GREEN}✅ Prerequisites check passed${NC}"
}

# Function to create beta tester
create_beta_tester() {
    echo -e "${YELLOW}👥 Creating beta tester...${NC}"
    
    read -p "Enter tester name: " name
    read -p "Enter tester email: " email
    read -p "Enter company: " company
    read -p "Enter role: " role
    read -p "Enter experience level (beginner/intermediate/advanced): " experience
    read -p "Enter preferred SDK (go/python/nodejs): " preferred_sdk
    read -p "Enter integration type (web/mobile/desktop/api): " integration_type
    
    # Create beta tester
    local response=$(curl -s -X POST "$BETA_MANAGER_URL/api/v1/beta/testers" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$name\",
            \"email\": \"$email\",
            \"company\": \"$company\",
            \"role\": \"$role\",
            \"experience\": \"$experience\",
            \"preferred_sdk\": \"$preferred_sdk\",
            \"integration_type\": \"$integration_type\"
        }")
    
    if echo "$response" | jq -e '.id' > /dev/null; then
        local tester_id=$(echo "$response" | jq -r '.id')
        local api_key=$(echo "$response" | jq -r '.api_key')
        
        echo -e "${GREEN}✅ Beta tester created successfully${NC}"
        echo -e "${BLUE}Tester ID: $tester_id${NC}"
        echo -e "${BLUE}API Key: $api_key${NC}"
        
        # Send welcome email
        echo -e "${YELLOW}📧 Sending welcome email...${NC}"
        # This would trigger the welcome email in a real implementation
        
    else
        echo -e "${RED}❌ Failed to create beta tester${NC}"
        echo "$response"
    fi
}

# Function to send invitation
send_invitation() {
    echo -e "${YELLOW}📧 Sending beta testing invitation...${NC}"
    
    read -p "Enter invitee name: " name
    read -p "Enter invitee email: " email
    read -p "Enter company: " company
    read -p "Enter role: " role
    read -p "Enter experience level (beginner/intermediate/advanced): " experience
    read -p "Enter preferred SDK (go/python/nodejs): " preferred_sdk
    read -p "Enter integration type (web/mobile/desktop/api): " integration_type
    read -p "Enter personal message (optional): " message
    
    # Send invitation
    local response=$(curl -s -X POST "$INVITATION_SYSTEM_URL/api/v1/beta/invitations" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$name\",
            \"email\": \"$email\",
            \"company\": \"$company\",
            \"role\": \"$role\",
            \"experience\": \"$experience\",
            \"preferred_sdk\": \"$preferred_sdk\",
            \"integration_type\": \"$integration_type\",
            \"message\": \"$message\"
        }")
    
    if echo "$response" | jq -e '.id' > /dev/null; then
        local invitation_id=$(echo "$response" | jq -r '.id')
        
        echo -e "${GREEN}✅ Invitation sent successfully${NC}"
        echo -e "${BLUE}Invitation ID: $invitation_id${NC}"
        echo -e "${BLUE}Expires: $(echo "$response" | jq -r '.expires_at')${NC}"
        
    else
        echo -e "${RED}❌ Failed to send invitation${NC}"
        echo "$response"
    fi
}

# Function to view beta testers
view_beta_testers() {
    echo -e "${YELLOW}👥 Viewing beta testers...${NC}"
    
    local response=$(curl -s "$BETA_MANAGER_URL/api/v1/beta/testers")
    
    if echo "$response" | jq -e '.testers' > /dev/null; then
        echo "$response" | jq -r '.testers[] | "\(.name) (\(.company)) - \(.status) - \(.email)"'
    else
        echo -e "${RED}❌ Failed to retrieve beta testers${NC}"
        echo "$response"
    fi
}

# Function to view feedback
view_feedback() {
    echo -e "${YELLOW}💬 Viewing feedback...${NC}"
    
    local response=$(curl -s "$FEEDBACK_COLLECTOR_URL/api/v1/beta/feedback")
    
    if echo "$response" | jq -e '.feedback' > /dev/null; then
        echo "$response" | jq -r '.feedback[] | "\(.beta_tester_name) (\(.company)) - Rating: \(.overall_rating)/5 - \(.submitted_at)"'
    else
        echo -e "${RED}❌ Failed to retrieve feedback${NC}"
        echo "$response"
    fi
}

# Function to view feedback stats
view_feedback_stats() {
    echo -e "${YELLOW}📊 Viewing feedback statistics...${NC}"
    
    local response=$(curl -s "$FEEDBACK_COLLECTOR_URL/api/v1/beta/feedback/stats")
    
    if echo "$response" | jq -e '.total_feedback' > /dev/null; then
        echo -e "${BLUE}Total Feedback: $(echo "$response" | jq -r '.total_feedback')${NC}"
        echo -e "${BLUE}Average Rating: $(echo "$response" | jq -r '.average_rating')${NC}"
        echo -e "${BLUE}Bug Reports: $(echo "$response" | jq -r '.bug_count')${NC}"
        echo -e "${BLUE}Feature Requests: $(echo "$response" | jq -r '.feature_request_count')${NC}"
        
        echo -e "\n${YELLOW}Category Ratings:${NC}"
        echo "$response" | jq -r '.category_ratings | to_entries[] | "\(.key): \(.value)"'
        
        echo -e "\n${YELLOW}SDK Usage:${NC}"
        echo "$response" | jq -r '.sdk_usage | to_entries[] | "\(.key): \(.value)"'
        
    else
        echo -e "${RED}❌ Failed to retrieve feedback stats${NC}"
        echo "$response"
    fi
}

# Function to view beta program stats
view_beta_program_stats() {
    echo -e "${YELLOW}📊 Viewing beta program statistics...${NC}"
    
    local response=$(curl -s "$BETA_MANAGER_URL/api/v1/beta/stats")
    
    if echo "$response" | jq -e '.total_invites' > /dev/null; then
        echo -e "${BLUE}Total Invites: $(echo "$response" | jq -r '.total_invites')${NC}"
        echo -e "${BLUE}Accepted Invites: $(echo "$response" | jq -r '.accepted_invites')${NC}"
        echo -e "${BLUE}Active Testers: $(echo "$response" | jq -r '.active_testers')${NC}"
        echo -e "${BLUE}Completed Testers: $(echo "$response" | jq -r '.completed_testers')${NC}"
        echo -e "${BLUE}Total Feedback: $(echo "$response" | jq -r '.total_feedback')${NC}"
        echo -e "${BLUE}Average Rating: $(echo "$response" | jq -r '.average_rating')${NC}"
        
        echo -e "\n${YELLOW}SDK Usage:${NC}"
        echo "$response" | jq -r '.sdk_usage | to_entries[] | "\(.key): \(.value)"'
        
        echo -e "\n${YELLOW}Integration Types:${NC}"
        echo "$response" | jq -r '.integration_types | to_entries[] | "\(.key): \(.value)"'
        
        echo -e "\n${YELLOW}Experience Levels:${NC}"
        echo "$response" | jq -r '.experience_levels | to_entries[] | "\(.key): \(.value)"'
        
    else
        echo -e "${RED}❌ Failed to retrieve beta program stats${NC}"
        echo "$response"
    fi
}

# Function to run performance test
run_performance_test() {
    echo -e "${YELLOW}🚀 Running performance test...${NC}"
    
    read -p "Enter test duration (minutes): " duration
    read -p "Enter number of users: " users
    read -p "Enter requests per second: " rps
    
    echo -e "${BLUE}Running performance test for $duration minutes with $users users at $rps RPS...${NC}"
    
    go run ./cmd/load_test.go \
        -url="$BETA_SERVICE_URL" \
        -duration="${duration}m" \
        -users="$users" \
        -rps="$rps" \
        -type=load \
        -verbose
}

# Function to check service health
check_service_health() {
    echo -e "${YELLOW}🏥 Checking service health...${NC}"
    
    # Check main service
    echo -e "${BLUE}Checking main service...${NC}"
    local main_health=$(curl -s "$BETA_SERVICE_URL/health")
    if echo "$main_health" | jq -e '.status' > /dev/null; then
        echo -e "${GREEN}✅ Main service is healthy${NC}"
    else
        echo -e "${RED}❌ Main service is unhealthy${NC}"
    fi
    
    # Check performance health
    echo -e "${BLUE}Checking performance health...${NC}"
    local perf_health=$(curl -s "$BETA_SERVICE_URL/api/v1/performance/health")
    if echo "$perf_health" | jq -e '.status' > /dev/null; then
        echo -e "${GREEN}✅ Performance monitoring is healthy${NC}"
    else
        echo -e "${RED}❌ Performance monitoring is unhealthy${NC}"
    fi
    
    # Check beta manager
    echo -e "${BLUE}Checking beta manager...${NC}"
    if curl -s "$BETA_MANAGER_URL/health" &> /dev/null; then
        echo -e "${GREEN}✅ Beta manager is healthy${NC}"
    else
        echo -e "${RED}❌ Beta manager is unhealthy${NC}"
    fi
    
    # Check feedback collector
    echo -e "${BLUE}Checking feedback collector...${NC}"
    if curl -s "$FEEDBACK_COLLECTOR_URL/health" &> /dev/null; then
        echo -e "${GREEN}✅ Feedback collector is healthy${NC}"
    else
        echo -e "${RED}❌ Feedback collector is unhealthy${NC}"
    fi
}

# Function to generate beta testing report
generate_report() {
    echo -e "${YELLOW}📋 Generating beta testing report...${NC}"
    
    local report_file="beta_testing_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# Beta Testing Report
Generated on: $(date)

## Service Health
EOF
    
    # Add service health to report
    echo "### Main Service" >> "$report_file"
    curl -s "$BETA_SERVICE_URL/health" | jq '.' >> "$report_file"
    
    echo -e "\n### Performance Health" >> "$report_file"
    curl -s "$BETA_SERVICE_URL/api/v1/performance/health" | jq '.' >> "$report_file"
    
    # Add beta program stats to report
    echo -e "\n## Beta Program Statistics" >> "$report_file"
    curl -s "$BETA_MANAGER_URL/api/v1/beta/stats" | jq '.' >> "$report_file"
    
    # Add feedback stats to report
    echo -e "\n## Feedback Statistics" >> "$report_file"
    curl -s "$FEEDBACK_COLLECTOR_URL/api/v1/beta/feedback/stats" | jq '.' >> "$report_file"
    
    echo -e "${GREEN}✅ Report generated: $report_file${NC}"
}

# Function to start beta testing services
start_services() {
    echo -e "${YELLOW}🚀 Starting beta testing services...${NC}"
    
    # Start beta manager
    echo -e "${BLUE}Starting beta manager...${NC}"
    go run ./beta-testing/beta-manager.go &
    local beta_manager_pid=$!
    echo "Beta manager PID: $beta_manager_pid"
    
    # Start feedback collector
    echo -e "${BLUE}Starting feedback collector...${NC}"
    go run ./beta-testing/feedback-collector.go &
    local feedback_collector_pid=$!
    echo "Feedback collector PID: $feedback_collector_pid"
    
    # Start invitation system
    echo -e "${BLUE}Starting invitation system...${NC}"
    go run ./beta-testing/invitation-system.go &
    local invitation_system_pid=$!
    echo "Invitation system PID: $invitation_system_pid"
    
    # Wait for services to start
    echo -e "${YELLOW}⏳ Waiting for services to start...${NC}"
    sleep 5
    
    # Check if services are running
    check_service_health
    
    echo -e "${GREEN}✅ Beta testing services started${NC}"
    echo -e "${BLUE}Beta Manager: http://localhost:8081${NC}"
    echo -e "${BLUE}Feedback Collector: http://localhost:8080${NC}"
    echo -e "${BLUE}Invitation System: http://localhost:8082${NC}"
}

# Function to stop beta testing services
stop_services() {
    echo -e "${YELLOW}🛑 Stopping beta testing services...${NC}"
    
    # Kill all Go processes related to beta testing
    pkill -f "beta-manager.go" || true
    pkill -f "feedback-collector.go" || true
    pkill -f "invitation-system.go" || true
    
    echo -e "${GREEN}✅ Beta testing services stopped${NC}"
}

# Function to show help
show_help() {
    echo -e "${BLUE}Beta Testing Management Script${NC}"
    echo "================================================================"
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  create-tester     - Create a new beta tester"
    echo "  send-invitation   - Send beta testing invitation"
    echo "  view-testers      - View all beta testers"
    echo "  view-feedback     - View feedback from beta testers"
    echo "  view-feedback-stats - View feedback statistics"
    echo "  view-program-stats - View beta program statistics"
    echo "  performance-test  - Run performance test"
    echo "  check-health      - Check service health"
    echo "  generate-report   - Generate beta testing report"
    echo "  start-services    - Start beta testing services"
    echo "  stop-services     - Stop beta testing services"
    echo "  help              - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 create-tester"
    echo "  $0 send-invitation"
    echo "  $0 view-feedback-stats"
    echo "  $0 performance-test"
    echo "  $0 generate-report"
}

# Main function
main() {
    case "${1:-}" in
        "create-tester")
            check_prerequisites
            create_beta_tester
            ;;
        "send-invitation")
            check_prerequisites
            send_invitation
            ;;
        "view-testers")
            check_prerequisites
            view_beta_testers
            ;;
        "view-feedback")
            check_prerequisites
            view_feedback
            ;;
        "view-feedback-stats")
            check_prerequisites
            view_feedback_stats
            ;;
        "view-program-stats")
            check_prerequisites
            view_beta_program_stats
            ;;
        "performance-test")
            check_prerequisites
            run_performance_test
            ;;
        "check-health")
            check_prerequisites
            check_service_health
            ;;
        "generate-report")
            check_prerequisites
            generate_report
            ;;
        "start-services")
            start_services
            ;;
        "stop-services")
            stop_services
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            echo -e "${RED}❌ Unknown command: ${1:-}${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
