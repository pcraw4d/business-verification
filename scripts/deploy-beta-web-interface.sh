#!/bin/bash

# Beta Web Interface Deployment Script
# This script deploys the KYB Platform web interface for public beta testing

set -e

echo "🚀 Deploying KYB Platform Beta Web Interface..."

# Configuration
BETA_DOMAIN="${BETA_DOMAIN:-beta.kybplatform.com}"
BETA_ENVIRONMENT="${BETA_ENVIRONMENT:-staging}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io/pcraw4d}"

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

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check if docker-compose is installed
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check if required environment variables are set
    if [[ -z "$BETA_API_KEY" ]]; then
        print_warning "BETA_API_KEY not set. Will use default configuration."
    fi
    
    print_success "Prerequisites check completed"
}

# Create Docker Compose configuration for beta environment
create_beta_compose() {
    print_status "Creating beta environment configuration..."
    
    cat > docker-compose.beta.yml << EOF
version: '3.8'

services:
  # Web Interface
  web-interface:
    image: ${DOCKER_REGISTRY}/kyb-web-interface:beta
    container_name: kyb-beta-web
    ports:
      - "80:80"
      - "443:443"
    environment:
      - NODE_ENV=production
      - API_BASE_URL=https://api.${BETA_DOMAIN}
      - BETA_MODE=true
      - ANALYTICS_ENABLED=true
    volumes:
      - ./configs/beta/web-interface.conf:/etc/nginx/conf.d/default.conf
    depends_on:
      - api-gateway
    networks:
      - kyb-beta-network

  # API Gateway
  api-gateway:
    image: ${DOCKER_REGISTRY}/kyb-api-gateway:beta
    container_name: kyb-beta-api
    ports:
      - "8081:8081"
    environment:
      - ENVIRONMENT=beta
      - DATABASE_URL=${BETA_DATABASE_URL}
      - REDIS_URL=${BETA_REDIS_URL}
      - JWT_SECRET=${BETA_JWT_SECRET}
      - API_RATE_LIMIT=1000
      - CORS_ORIGIN=https://${BETA_DOMAIN}
    volumes:
      - ./configs/beta/api-config.yaml:/app/config.yaml
    networks:
      - kyb-beta-network

  # Database
  database:
    image: postgres:15-alpine
    container_name: kyb-beta-db
    environment:
      - POSTGRES_DB=kyb_beta
      - POSTGRES_USER=${BETA_DB_USER:-kyb_beta_user}
      - POSTGRES_PASSWORD=${BETA_DB_PASSWORD:-kyb_beta_password}
    volumes:
      - kyb-beta-db-data:/var/lib/postgresql/data
      - ./configs/beta/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - kyb-beta-network

  # Redis for caching and sessions
  redis:
    image: redis:7-alpine
    container_name: kyb-beta-redis
    volumes:
      - kyb-beta-redis-data:/data
    networks:
      - kyb-beta-network

  # Monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: kyb-beta-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./configs/beta/prometheus.yml:/etc/prometheus/prometheus.yml
      - kyb-beta-prometheus-data:/prometheus
    networks:
      - kyb-beta-network

  grafana:
    image: grafana/grafana:latest
    container_name: kyb-beta-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${BETA_GRAFANA_PASSWORD:-admin123}
    volumes:
      - kyb-beta-grafana-data:/var/lib/grafana
    networks:
      - kyb-beta-network

volumes:
  kyb-beta-db-data:
  kyb-beta-redis-data:
  kyb-beta-prometheus-data:
  kyb-beta-grafana-data:

networks:
  kyb-beta-network:
    driver: bridge
EOF

    print_success "Beta environment configuration created"
}

# Create beta configuration files
create_beta_configs() {
    print_status "Creating beta configuration files..."
    
    # Create configs directory
    mkdir -p configs/beta
    
    # Web interface nginx configuration
    cat > configs/beta/web-interface.conf << EOF
server {
    listen 80;
    server_name ${BETA_DOMAIN};
    
    # Redirect HTTP to HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name ${BETA_DOMAIN};
    
    # SSL configuration (you'll need to add your SSL certificates)
    ssl_certificate /etc/ssl/certs/${BETA_DOMAIN}.crt;
    ssl_certificate_key /etc/ssl/private/${BETA_DOMAIN}.key;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    
    # Root directory
    root /usr/share/nginx/html;
    index index.html;
    
    # Handle client-side routing
    location / {
        try_files \$uri \$uri/ /index.html;
    }
    
    # API proxy
    location /api/ {
        proxy_pass http://api-gateway:8081;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # Health check
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF

    # API configuration
    cat > configs/beta/api-config.yaml << EOF
environment: beta
server:
  port: 8081
  host: 0.0.0.0

database:
  host: database
  port: 5432
  name: kyb_beta
  user: ${BETA_DB_USER:-kyb_beta_user}
  password: ${BETA_DB_PASSWORD:-kyb_beta_password}
  ssl_mode: disable

redis:
  host: redis
  port: 6379
  password: ""
  db: 0

jwt:
  secret: ${BETA_JWT_SECRET:-your-super-secret-jwt-key-change-this}
  expiration: 24h

cors:
  allowed_origins:
    - https://${BETA_DOMAIN}
    - http://localhost:3000
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS

rate_limiting:
  enabled: true
  requests_per_minute: 1000
  burst_size: 100

monitoring:
  prometheus:
    enabled: true
    path: /metrics
  health_check:
    enabled: true
    path: /health

beta_features:
  enabled: true
  feedback_collection: true
  analytics: true
  user_tracking: true
EOF

    # Prometheus configuration
    cat > configs/beta/prometheus.yml << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'kyb-api'
    static_configs:
      - targets: ['api-gateway:8081']
    metrics_path: /metrics

  - job_name: 'kyb-web'
    static_configs:
      - targets: ['web-interface:80']
    metrics_path: /metrics
EOF

    print_success "Beta configuration files created"
}

# Deploy the beta environment
deploy_beta() {
    print_status "Deploying beta environment..."
    
    # Pull latest images
    docker-compose -f docker-compose.beta.yml pull
    
    # Start services
    docker-compose -f docker-compose.beta.yml up -d
    
    print_success "Beta environment deployed successfully"
}

# Setup SSL certificates (basic self-signed for testing)
setup_ssl() {
    print_status "Setting up SSL certificates..."
    
    mkdir -p configs/beta/ssl
    
    # Generate self-signed certificate for testing
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout configs/beta/ssl/${BETA_DOMAIN}.key \
        -out configs/beta/ssl/${BETA_DOMAIN}.crt \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=${BETA_DOMAIN}"
    
    print_warning "Self-signed SSL certificate generated for testing"
    print_warning "For production, replace with proper SSL certificates"
}

# Generate shareable links
generate_shareable_links() {
    print_status "Generating shareable links..."
    
    cat > SHAREABLE_LINKS.md << EOF
# KYB Platform Beta Testing - Shareable Links

## 🌐 Public Access Links

### Web Interface (For Non-Technical Users)
- **Main Dashboard**: https://${BETA_DOMAIN}
- **User Registration**: https://${BETA_DOMAIN}/register
- **User Login**: https://${BETA_DOMAIN}/login

### API Documentation (For Technical Users)
- **API Documentation**: https://${BETA_DOMAIN}/docs
- **API Base URL**: https://${BETA_DOMAIN}/api/v1

### Monitoring & Analytics
- **Grafana Dashboard**: https://${BETA_DOMAIN}:3000
- **Prometheus Metrics**: https://${BETA_DOMAIN}:9090

## 🔑 Beta Testing Credentials

### Test Users
- **Compliance Officer**: compliance@beta.kybplatform.com / password123
- **Risk Manager**: risk@beta.kybplatform.com / password123
- **Business Analyst**: analyst@beta.kybplatform.com / password123

### API Access
- **API Key**: ${BETA_API_KEY:-your-api-key-here}
- **Rate Limit**: 1000 requests per minute

## 📋 Beta Testing Instructions

### For Non-Technical Users:
1. Visit https://${BETA_DOMAIN}
2. Click "Register" to create an account
3. Use the web interface to test business classification
4. Explore risk assessment and compliance features
5. Provide feedback through the built-in feedback system

### For Technical Users:
1. Visit https://${BETA_DOMAIN}/docs for API documentation
2. Use the API endpoints for integration testing
3. Test authentication and rate limiting
4. Validate all platform features programmatically

## 🚨 Important Notes
- This is a beta environment - data may be reset periodically
- SSL certificate is self-signed for testing
- Report any issues through the feedback system
- Beta testing period: [Start Date] to [End Date]

## 📞 Support
- **Email**: beta-support@kybplatform.com
- **Documentation**: https://${BETA_DOMAIN}/docs
- **Feedback**: Use the feedback form in the web interface
EOF

    print_success "Shareable links generated in SHAREABLE_LINKS.md"
}

# Health check
health_check() {
    print_status "Performing health check..."
    
    # Wait for services to start
    sleep 30
    
    # Check web interface
    if curl -f -s https://${BETA_DOMAIN}/health > /dev/null; then
        print_success "Web interface is healthy"
    else
        print_error "Web interface health check failed"
        return 1
    fi
    
    # Check API
    if curl -f -s https://${BETA_DOMAIN}/api/v1/health > /dev/null; then
        print_success "API is healthy"
    else
        print_error "API health check failed"
        return 1
    fi
    
    print_success "All services are healthy"
}

# Main deployment process
main() {
    echo "🚀 KYB Platform Beta Web Interface Deployment"
    echo "=============================================="
    
    check_prerequisites
    create_beta_compose
    create_beta_configs
    setup_ssl
    deploy_beta
    generate_shareable_links
    health_check
    
    echo ""
    echo "🎉 Beta deployment completed successfully!"
    echo ""
    echo "📋 Next Steps:"
    echo "1. Review SHAREABLE_LINKS.md for access information"
    echo "2. Send the shareable links to your beta testers"
    echo "3. Monitor the deployment using Grafana dashboard"
    echo "4. Collect feedback through the web interface"
    echo ""
    echo "🌐 Web Interface: https://${BETA_DOMAIN}"
    echo "📊 Monitoring: https://${BETA_DOMAIN}:3000"
    echo ""
}

# Run main function
main "$@"
