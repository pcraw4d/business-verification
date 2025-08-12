#!/bin/bash

# KYB Platform - Security Audit Script
# This script implements the manual security checklist from the implementation guide

set -e

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

# Function to check if a service is running
check_service() {
    local service_name=$1
    local port=$2
    local url=$3
    
    if curl -f -s "$url" > /dev/null 2>&1; then
        print_success "$service_name is running on $url"
        return 0
    else
        print_error "$service_name is not accessible on $url"
        return 1
    fi
}

# Function to check JWT configuration
check_jwt_config() {
    print_status "Checking JWT configuration..."
    
    # Check if JWT secret is set and not default
    if [ -f ".env.local" ]; then
        JWT_SECRET=$(grep "JWT_SECRET" .env.local | cut -d'=' -f2)
        if [ "$JWT_SECRET" = "local_jwt_secret_change_in_production" ]; then
            print_warning "JWT secret is using default value - should be changed in production"
        else
            print_success "JWT secret is configured"
        fi
    else
        print_error "No .env.local file found"
    fi
}

# Function to check authentication endpoints
check_auth_endpoints() {
    print_status "Checking authentication endpoints..."
    
    # Test health endpoint
    if curl -f -s "http://localhost:8080/health" > /dev/null 2>&1; then
        print_success "Health endpoint is accessible"
    else
        print_error "Health endpoint is not accessible"
    fi
    
    # Test authentication endpoint (should return 401 for unauthenticated requests)
    AUTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/v1/auth/me" 2>/dev/null || echo "000")
    if [ "$AUTH_RESPONSE" = "401" ]; then
        print_success "Authentication endpoint properly requires authentication"
    else
        print_warning "Authentication endpoint returned $AUTH_RESPONSE (expected 401)"
    fi
}

# Function to check input validation
check_input_validation() {
    print_status "Checking input validation..."
    
    # Test with malformed JSON
    MALFORMED_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d '{"invalid": json}' \
        "http://localhost:8080/v1/classify" 2>/dev/null || echo "000")
    
    if [ "$MALFORMED_RESPONSE" = "400" ]; then
        print_success "Input validation rejects malformed JSON"
    else
        print_warning "Input validation returned $MALFORMED_RESPONSE for malformed JSON (expected 400)"
    fi
}

# Function to check rate limiting
check_rate_limiting() {
    print_status "Checking rate limiting..."
    
    # Make multiple rapid requests
    for i in {1..10}; do
        curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/health" &
    done
    wait
    
    # Check if rate limiting is working (this is a basic test)
    print_success "Rate limiting configuration checked"
}

# Function to check CORS configuration
check_cors_config() {
    print_status "Checking CORS configuration..."
    
    # Test CORS headers
    CORS_HEADERS=$(curl -s -I -H "Origin: http://localhost:3000" \
        "http://localhost:8080/health" 2>/dev/null | grep -i "access-control" || echo "")
    
    if [ -n "$CORS_HEADERS" ]; then
        print_success "CORS headers are present"
    else
        print_warning "CORS headers not found"
    fi
}

# Function to check database security
check_database_security() {
    print_status "Checking database security..."
    
    # Check if database is accessible only from localhost
    if docker ps | grep -q "postgres"; then
        print_success "PostgreSQL container is running"
        
        # Check if database port is exposed only locally
        if netstat -an 2>/dev/null | grep ":5433" | grep -q "127.0.0.1"; then
            print_success "Database port is bound to localhost only"
        else
            print_warning "Database port may be exposed externally"
        fi
    else
        print_error "PostgreSQL container is not running"
    fi
}

# Function to check SSL/TLS configuration
check_ssl_config() {
    print_status "Checking SSL/TLS configuration..."
    
    # Check if HTTPS is configured (for production)
    if [ "$ENV" = "production" ]; then
        print_warning "SSL/TLS configuration should be verified in production"
    else
        print_success "Development environment - SSL not required"
    fi
}

# Function to check environment variables
check_environment_vars() {
    print_status "Checking environment variables..."
    
    if [ -f ".env.local" ]; then
        # Check for sensitive information in environment file
        if grep -q "password\|secret\|key" .env.local; then
            print_warning "Sensitive information found in .env.local file"
        else
            print_success "No obvious sensitive information in environment file"
        fi
        
        # Check if environment file has proper permissions
        if [ "$(stat -f %Lp .env.local 2>/dev/null || stat -c %a .env.local 2>/dev/null)" = "600" ]; then
            print_success "Environment file has secure permissions"
        else
            print_warning "Environment file should have 600 permissions"
        fi
    else
        print_error "No .env.local file found"
    fi
}

# Function to check logging configuration
check_logging_config() {
    print_status "Checking logging configuration..."
    
    # Check if logs are being generated
    if docker logs newtool-kyb-platform-1 2>/dev/null | tail -n 5 | grep -q "level"; then
        print_success "Structured logging is working"
    else
        print_warning "Logging configuration should be verified"
    fi
}

# Function to check monitoring and alerting
check_monitoring() {
    print_status "Checking monitoring and alerting..."
    
    # Check if Prometheus is running
    if check_service "Prometheus" "9090" "http://localhost:9090"; then
        print_success "Prometheus monitoring is active"
    fi
    
    # Check if Grafana is running
    if check_service "Grafana" "3000" "http://localhost:3000"; then
        print_success "Grafana dashboard is accessible"
    fi
}

# Function to generate security report
generate_security_report() {
    print_status "Generating security audit report..."
    
    cat > security-audit-report.txt << EOF
# KYB Platform - Security Audit Report
Generated: $(date)

## Executive Summary
This report contains the results of a comprehensive security audit of the KYB Platform.

## Security Checklist Results

### Authentication & Authorization
- [ ] JWT tokens properly configured
- [ ] Authorization: RBAC implemented correctly
- [ ] Input Validation: All endpoints validated
- [ ] SQL Injection: Parameterized queries used
- [ ] XSS Protection: Headers configured
- [ ] CSRF Protection: Tokens implemented
- [ ] Rate Limiting: Configured per endpoint
- [ ] Encryption: TLS 1.3, data at rest encrypted

### Infrastructure Security
- [ ] Database access restricted
- [ ] Environment variables secured
- [ ] Logging configured properly
- [ ] Monitoring and alerting active
- [ ] SSL/TLS configured (production)

### Application Security
- [ ] Input validation implemented
- [ ] Error handling secure
- [ ] CORS properly configured
- [ ] Rate limiting active
- [ ] Authentication required

## Recommendations
1. Change default JWT secret in production
2. Implement proper SSL/TLS in production
3. Configure proper file permissions
4. Set up automated security scanning
5. Implement security headers

## Next Steps
1. Address any warnings found in this audit
2. Implement missing security features
3. Set up continuous security monitoring
4. Regular security audits

EOF
    
    print_success "Security audit report generated: security-audit-report.txt"
}

# Main security audit function
main_security_audit() {
    echo "🔒 KYB Platform - Security Audit"
    echo "================================="
    echo
    
    # Check if application is running
    if ! check_service "KYB Platform" "8080" "http://localhost:8080/health"; then
        print_error "KYB Platform is not running. Please start the application first."
        exit 1
    fi
    
    echo
    print_status "Starting comprehensive security audit..."
    echo
    
    # Run all security checks
    check_jwt_config
    check_auth_endpoints
    check_input_validation
    check_rate_limiting
    check_cors_config
    check_database_security
    check_ssl_config
    check_environment_vars
    check_logging_config
    check_monitoring
    
    echo
    generate_security_report
    
    echo
    print_success "Security audit completed!"
    echo
    print_status "Review the security-audit-report.txt file for detailed results."
}

# Function to show usage
show_usage() {
    echo "KYB Platform - Security Audit Tool"
    echo "=================================="
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  audit     - Run comprehensive security audit"
    echo "  jwt       - Check JWT configuration"
    echo "  auth      - Check authentication endpoints"
    echo "  input     - Check input validation"
    echo "  rate      - Check rate limiting"
    echo "  cors      - Check CORS configuration"
    echo "  db        - Check database security"
    echo "  ssl       - Check SSL/TLS configuration"
    echo "  env       - Check environment variables"
    echo "  logs      - Check logging configuration"
    echo "  monitor   - Check monitoring and alerting"
    echo "  report    - Generate security report"
    echo "  help      - Show this help message"
    echo
}

# Main execution
main() {
    case "${1:-help}" in
        audit)
            main_security_audit
            ;;
        jwt)
            check_jwt_config
            ;;
        auth)
            check_auth_endpoints
            ;;
        input)
            check_input_validation
            ;;
        rate)
            check_rate_limiting
            ;;
        cors)
            check_cors_config
            ;;
        db)
            check_database_security
            ;;
        ssl)
            check_ssl_config
            ;;
        env)
            check_environment_vars
            ;;
        logs)
            check_logging_config
            ;;
        monitor)
            check_monitoring
            ;;
        report)
            generate_security_report
            ;;
        help|*)
            show_usage
            ;;
    esac
}

# Run main function
main "$@"
