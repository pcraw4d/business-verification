# Railway Environment Setup Guide
## Risk Assessment Service - Environment Variables Configuration

### Overview

This guide provides step-by-step instructions for configuring environment variables in Railway for the Risk Assessment Service across different environments (development, staging, production).

---

## Prerequisites

- Railway account with appropriate permissions
- Railway CLI installed and authenticated
- Access to Supabase project
- Redis instance (shared or dedicated)
- External API keys (NewsAPI, OpenCorporates, Thomson Reuters, OFAC)

---

## Environment Setup

### 1. Development Environment

#### 1.1 Create Development Project
```bash
# Create new Railway project for development
railway login
railway init risk-assessment-service-dev
```

#### 1.2 Set Development Environment Variables
```bash
# Set basic service configuration
railway variables set ENVIRONMENT=development
railway variables set SERVICE_NAME=risk-assessment-service
railway variables set SERVICE_VERSION=1.0.0
railway variables set PORT=8080

# Set database configuration
railway variables set DATABASE_URL=postgresql://postgres:[DEV-PASSWORD]@db.[DEV-PROJECT-REF].supabase.co:5432/postgres
railway variables set SUPABASE_URL=https://[DEV-PROJECT-REF].supabase.co
railway variables set SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.[DEV-ANON-KEY]
railway variables set SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.[DEV-SERVICE-ROLE-KEY]

# Set Redis configuration
railway variables set REDIS_URL=redis://default:[DEV-REDIS-PASSWORD]@[DEV-REDIS-HOST]:6379
railway variables set REDIS_PASSWORD=[DEV-REDIS-PASSWORD]

# Set logging configuration
railway variables set LOG_LEVEL=debug
railway variables set LOG_FORMAT=console

# Set development-specific flags
railway variables set AUTH_REQUIRE_AUTH=false
railway variables set DEV_ENABLE_DEBUG_ENDPOINTS=true
railway variables set DEV_ENABLE_SWAGGER=true
```

#### 1.3 Deploy Development Service
```bash
# Deploy to development
railway up --detach
```

### 2. Staging Environment

#### 2.1 Create Staging Project
```bash
# Create new Railway project for staging
railway init risk-assessment-service-staging
```

#### 2.2 Set Staging Environment Variables
```bash
# Set basic service configuration
railway variables set ENVIRONMENT=staging
railway variables set SERVICE_NAME=risk-assessment-service
railway variables set SERVICE_VERSION=1.0.0
railway variables set PORT=8080

# Set database configuration
railway variables set DATABASE_URL=postgresql://postgres:[STAGING-PASSWORD]@db.[STAGING-PROJECT-REF].supabase.co:5432/postgres
railway variables set SUPABASE_URL=https://[STAGING-PROJECT-REF].supabase.co
railway variables set SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.[STAGING-ANON-KEY]
railway variables set SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.[STAGING-SERVICE-ROLE-KEY]

# Set Redis configuration
railway variables set REDIS_URL=redis://default:[STAGING-REDIS-PASSWORD]@[STAGING-REDIS-HOST]:6379
railway variables set REDIS_PASSWORD=[STAGING-REDIS-PASSWORD]

# Set logging configuration
railway variables set LOG_LEVEL=info
railway variables set LOG_FORMAT=json

# Set authentication
railway variables set AUTH_REQUIRE_AUTH=true
railway variables set JWT_SECRET=[STAGING-JWT-SECRET]
railway variables set SERVICE_TOKEN=[STAGING-SERVICE-TOKEN]

# Set external API keys
railway variables set NEWSAPI_API_KEY=[STAGING-NEWSAPI-KEY]
railway variables set OPENCORPORATES_API_KEY=[STAGING-OPENCORPORATES-KEY]
railway variables set THOMSON_REUTERS_API_KEY=[STAGING-THOMSON-REUTERS-KEY]
railway variables set OFAC_API_KEY=[STAGING-OFAC-KEY]

# Set monitoring
railway variables set MONITORING_ENABLED=true
railway variables set METRICS_ENABLED=true
railway variables set SLACK_WEBHOOK_URL=[STAGING-SLACK-WEBHOOK]
```

#### 2.3 Deploy Staging Service
```bash
# Deploy to staging
railway up --detach
```

### 3. Production Environment

#### 3.1 Create Production Project
```bash
# Create new Railway project for production
railway init risk-assessment-service-production
```

#### 3.2 Set Production Environment Variables
```bash
# Set basic service configuration
railway variables set ENVIRONMENT=production
railway variables set SERVICE_NAME=risk-assessment-service
railway variables set SERVICE_VERSION=1.0.0
railway variables set PORT=8080

# Set database configuration
railway variables set DATABASE_URL=postgresql://postgres:[PROD-PASSWORD]@db.[PROD-PROJECT-REF].supabase.co:5432/postgres
railway variables set SUPABASE_URL=https://[PROD-PROJECT-REF].supabase.co
railway variables set SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.[PROD-ANON-KEY]
railway variables set SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.[PROD-SERVICE-ROLE-KEY]

# Set Redis configuration
railway variables set REDIS_URL=redis://default:[PROD-REDIS-PASSWORD]@[PROD-REDIS-HOST]:6379
railway variables set REDIS_PASSWORD=[PROD-REDIS-PASSWORD]

# Set logging configuration
railway variables set LOG_LEVEL=info
railway variables set LOG_FORMAT=json

# Set authentication
railway variables set AUTH_REQUIRE_AUTH=true
railway variables set JWT_SECRET=[PROD-JWT-SECRET]
railway variables set SERVICE_TOKEN=[PROD-SERVICE-TOKEN]

# Set external API keys
railway variables set NEWSAPI_API_KEY=[PROD-NEWSAPI-KEY]
railway variables set OPENCORPORATES_API_KEY=[PROD-OPENCORPORATES-KEY]
railway variables set THOMSON_REUTERS_API_KEY=[PROD-THOMSON-REUTERS-KEY]
railway variables set OFAC_API_KEY=[PROD-OFAC-KEY]

# Set monitoring and alerting
railway variables set MONITORING_ENABLED=true
railway variables set METRICS_ENABLED=true
railway variables set SLACK_WEBHOOK_URL=[PROD-SLACK-WEBHOOK]
railway variables set PAGERDUTY_INTEGRATION_KEY=[PAGERDUTY-KEY]
railway variables set ALERT_EMAIL_RECIPIENTS=alerts@kyb-platform.com,devops@kyb-platform.com

# Set security
railway variables set SECURITY_ENABLE_HSTS=true
railway variables set SECURITY_ENABLE_CSP=true
railway variables set CORS_ALLOWED_ORIGINS=https://kyb-platform.com,https://app.kyb-platform.com

# Set performance
railway variables set PERFORMANCE_ENABLE_TRACING=true
railway variables set PERFORMANCE_TRACE_SAMPLE_RATE=0.01
```

#### 3.3 Deploy Production Service
```bash
# Deploy to production
railway up --detach
```

---

## Environment-Specific Configurations

### Development Environment
- **Authentication**: Disabled for easier development
- **Logging**: Debug level with console format
- **External APIs**: Mock or test keys
- **Monitoring**: Basic monitoring only
- **Security**: Relaxed CORS and security headers

### Staging Environment
- **Authentication**: Enabled with test credentials
- **Logging**: Info level with JSON format
- **External APIs**: Staging/test API keys
- **Monitoring**: Full monitoring with staging alerts
- **Security**: Production-like security settings

### Production Environment
- **Authentication**: Full authentication enabled
- **Logging**: Info level with JSON format
- **External APIs**: Production API keys
- **Monitoring**: Full monitoring with production alerts
- **Security**: Strict security settings

---

## Railway Dashboard Configuration

### 1. Access Railway Dashboard
1. Go to [Railway.app](https://railway.app)
2. Log in to your account
3. Navigate to your project

### 2. Configure Environment Variables
1. Click on your service
2. Go to the "Variables" tab
3. Add environment variables using the "New Variable" button
4. Set the variable name and value
5. Click "Add" to save

### 3. Configure Service Settings
1. Go to the "Settings" tab
2. Configure the following:
   - **Port**: 8080
   - **Health Check Path**: /health
   - **Start Command**: ./risk-assessment-service
   - **Build Command**: go build -o risk-assessment-service ./cmd/main.go

### 4. Configure Domains
1. Go to the "Domains" tab
2. Add custom domains for each environment:
   - Development: `dev-risk-assessment.railway.app`
   - Staging: `staging-risk-assessment.railway.app`
   - Production: `risk-assessment.railway.app`

---

## Automated Environment Setup

### 1. Using Railway CLI Script
```bash
#!/bin/bash
# setup-railway-environments.sh

# Function to setup environment variables
setup_environment() {
    local env=$1
    local project_name=$2
    
    echo "Setting up $env environment..."
    
    # Switch to project
    railway link $project_name
    
    # Set common variables
    railway variables set ENVIRONMENT=$env
    railway variables set SERVICE_NAME=risk-assessment-service
    railway variables set SERVICE_VERSION=1.0.0
    railway variables set PORT=8080
    
    # Set environment-specific variables
    case $env in
        "development")
            railway variables set LOG_LEVEL=debug
            railway variables set LOG_FORMAT=console
            railway variables set AUTH_REQUIRE_AUTH=false
            ;;
        "staging")
            railway variables set LOG_LEVEL=info
            railway variables set LOG_FORMAT=json
            railway variables set AUTH_REQUIRE_AUTH=true
            ;;
        "production")
            railway variables set LOG_LEVEL=info
            railway variables set LOG_FORMAT=json
            railway variables set AUTH_REQUIRE_AUTH=true
            railway variables set SECURITY_ENABLE_HSTS=true
            ;;
    esac
    
    echo "$env environment setup complete!"
}

# Setup all environments
setup_environment "development" "risk-assessment-service-dev"
setup_environment "staging" "risk-assessment-service-staging"
setup_environment "production" "risk-assessment-service-production"
```

### 2. Using Railway API
```bash
#!/bin/bash
# setup-railway-api.sh

RAILWAY_TOKEN="your-railway-token"
PROJECT_ID="your-project-id"

# Function to set variable via API
set_variable() {
    local key=$1
    local value=$2
    
    curl -X POST \
        -H "Authorization: Bearer $RAILWAY_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"key\":\"$key\",\"value\":\"$value\"}" \
        "https://backboard.railway.app/graphql/v1"
}

# Set environment variables
set_variable "ENVIRONMENT" "production"
set_variable "SERVICE_NAME" "risk-assessment-service"
set_variable "PORT" "8080"
```

---

## Environment Variable Validation

### 1. Validation Script
```bash
#!/bin/bash
# validate-environment.sh

# Required variables for each environment
REQUIRED_VARS=(
    "ENVIRONMENT"
    "SERVICE_NAME"
    "PORT"
    "DATABASE_URL"
    "SUPABASE_URL"
    "SUPABASE_ANON_KEY"
    "SUPABASE_SERVICE_ROLE_KEY"
    "REDIS_URL"
    "REDIS_PASSWORD"
)

# Check if all required variables are set
validate_variables() {
    local env=$1
    echo "Validating $env environment variables..."
    
    for var in "${REQUIRED_VARS[@]}"; do
        if [ -z "${!var}" ]; then
            echo "ERROR: $var is not set for $env environment"
            exit 1
        fi
    done
    
    echo "All required variables are set for $env environment"
}

# Validate each environment
validate_variables "development"
validate_variables "staging"
validate_variables "production"
```

### 2. Health Check Validation
```bash
#!/bin/bash
# validate-health-checks.sh

# Function to check service health
check_health() {
    local url=$1
    local env=$2
    
    echo "Checking health for $env environment at $url..."
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url/health")
    
    if [ "$response" = "200" ]; then
        echo "✅ $env environment is healthy"
    else
        echo "❌ $env environment health check failed (HTTP $response)"
        exit 1
    fi
}

# Check all environments
check_health "https://dev-risk-assessment.railway.app" "development"
check_health "https://staging-risk-assessment.railway.app" "staging"
check_health "https://risk-assessment.railway.app" "production"
```

---

## Troubleshooting

### Common Issues

#### 1. Environment Variables Not Loading
```bash
# Check if variables are set
railway variables

# Restart service to reload variables
railway service restart
```

#### 2. Database Connection Issues
```bash
# Test database connection
railway run psql $DATABASE_URL -c "SELECT 1;"

# Check database URL format
echo $DATABASE_URL
```

#### 3. Redis Connection Issues
```bash
# Test Redis connection
railway run redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping

# Check Redis URL format
echo $REDIS_URL
```

#### 4. Service Not Starting
```bash
# Check service logs
railway logs --tail 100

# Check service status
railway status
```

### Debug Commands

```bash
# View all environment variables
railway variables

# View service logs
railway logs --tail 100

# Check service status
railway status

# Restart service
railway service restart

# View service metrics
railway metrics
```

---

## Security Best Practices

### 1. Environment Variable Security
- Never commit `.env` files to version control
- Use Railway's secure environment variable storage
- Rotate secrets regularly
- Use different credentials for each environment
- Enable 2FA on Railway account

### 2. Access Control
- Limit Railway project access to necessary team members
- Use service-specific API keys
- Implement proper authentication and authorization
- Monitor access logs

### 3. Monitoring and Alerting
- Set up alerts for failed deployments
- Monitor environment variable changes
- Track service health and performance
- Implement proper logging and monitoring

---

## Maintenance

### 1. Regular Tasks
- Review and rotate API keys monthly
- Update environment variables as needed
- Monitor service performance and costs
- Review and update security settings

### 2. Backup and Recovery
- Backup environment variable configurations
- Document all environment-specific settings
- Test disaster recovery procedures
- Maintain runbooks and documentation

---

## Support

### Railway Support
- [Railway Documentation](https://docs.railway.app)
- [Railway Discord](https://discord.gg/railway)
- [Railway Support](https://railway.app/support)

### Internal Support
- DevOps Team: devops@kyb-platform.com
- Platform Team: platform@kyb-platform.com
- Emergency: +1-XXX-XXX-XXXX

---

**Document Version**: 1.0.0  
**Last Updated**: December 2024  
**Next Review**: March 2025
