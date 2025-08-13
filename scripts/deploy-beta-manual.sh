#!/bin/bash

# KYB Platform Beta - Manual Deployment Setup
# This script prepares everything for Railway deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
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

# Configuration
PROJECT_NAME="${PROJECT_NAME:-kyb-platform-beta}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io}"

echo "🚀 KYB Platform Beta - Manual Deployment Setup"
echo "=============================================="

print_status "Creating Railway deployment files..."

# Create railway.json configuration
cat > railway.json << EOF
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile.beta"
  },
  "deploy": {
    "startCommand": "./kyb-platform",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 300,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
EOF

print_success "Created railway.json"

# Create Dockerfile.beta for Railway
cat > Dockerfile.beta << EOF
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-platform ./cmd/api

# Create final image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary and configs
COPY --from=builder /app/kyb-platform .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/web ./web

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \\
    adduser -u 1001 -S appuser -G appgroup

# Change ownership
RUN chown -R appuser:appgroup /root/

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \\
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./kyb-platform"]
EOF

print_success "Created Dockerfile.beta"

# Create .dockerignore for Railway
cat > .dockerignore << EOF
.git
.gitignore
README.md
docs/
test/
*.md
.env*
.air.toml
Makefile
Dockerfile*
docker-compose*
railway.json
EOF

print_success "Created .dockerignore"

# Create environment template
cat > .env.railway.template << EOF
# Railway Environment Variables Template
# Copy this to Railway dashboard and fill in your values

# Application
ENVIRONMENT=beta
BETA_MODE=true
PORT=8080

# Database (Supabase or Railway PostgreSQL)
DATABASE_URL=postgresql://postgres:password@localhost:5432/kyb_beta
DATABASE_TYPE=postgresql

# Supabase Integration (Optional)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Redis (for caching and sessions)
REDIS_URL=redis://localhost:6379

# Security
JWT_SECRET=your-jwt-secret-key-here
ENCRYPTION_KEY=your-encryption-key-here

# External APIs
BUSINESS_DATA_API_KEY=your-api-key
BUSINESS_DATA_API_URL=https://api.example.com

# Monitoring
ANALYTICS_ENABLED=true
FEEDBACK_COLLECTION=true
LOG_LEVEL=info

# CORS
CORS_ORIGIN=https://your-app.railway.app
EOF

print_success "Created .env.railway.template"

# Create deployment instructions
cat > DEPLOYMENT_INSTRUCTIONS.md << EOF
# Railway Deployment Instructions

## 🚀 Quick Deploy to Railway

### Step 1: Create Railway Project
1. Go to [railway.app](https://railway.app)
2. Click "New Project"
3. Choose "Deploy from GitHub repo"
4. Select this repository

### Step 2: Configure Environment Variables
1. In your Railway project dashboard
2. Go to "Variables" tab
3. Copy variables from \`.env.railway.template\`
4. Fill in your actual values:
   - \`JWT_SECRET\`: Generate a random 32-character string
   - \`ENCRYPTION_KEY\`: Generate a random 32-character string
   - \`SUPABASE_URL\`: Your Supabase project URL
   - \`SUPABASE_ANON_KEY\`: Your Supabase anon key
   - \`SUPABASE_SERVICE_ROLE_KEY\`: Your Supabase service role key

### Step 3: Add PostgreSQL Database
1. In Railway dashboard, click "New"
2. Select "Database" → "PostgreSQL"
3. Railway will automatically set \`DATABASE_URL\`

### Step 4: Deploy
1. Railway will automatically detect the Dockerfile.beta
2. Click "Deploy" to start the build
3. Wait for deployment to complete

### Step 5: Get Your URLs
1. Go to "Settings" tab
2. Copy your deployment URL
3. Share with beta testers

## 🔧 Manual Deployment Commands

If you prefer using Railway CLI:

\`\`\`bash
# Login to Railway
railway login

# Link to project
railway link

# Deploy
railway up

# Get deployment URL
railway domain
\`\`\`

## 📊 Monitoring

- **Logs**: View in Railway dashboard
- **Metrics**: Built-in monitoring
- **Health Checks**: Automatic health monitoring

## 🔒 Security Notes

- All secrets are encrypted in Railway
- Environment variables are secure
- SSL is automatically enabled
- Health checks ensure uptime

## 🐛 Troubleshooting

### Build Failures
- Check Dockerfile.beta syntax
- Verify all files are present
- Check Railway logs

### Runtime Errors
- Verify environment variables
- Check database connection
- Review application logs

### Performance Issues
- Monitor resource usage
- Scale up if needed
- Optimize Docker image
EOF

print_success "Created DEPLOYMENT_INSTRUCTIONS.md"

# Create a simple deployment script for Railway CLI (when available)
cat > deploy-railway.sh << 'EOF'
#!/bin/bash

# Railway CLI Deployment Script
# Run this after logging in with: railway login

set -e

echo "🚀 Deploying KYB Platform to Railway..."

# Check if logged in
if ! railway whoami > /dev/null 2>&1; then
    echo "❌ Not logged in to Railway. Please run: railway login"
    exit 1
fi

# Deploy
echo "📦 Deploying application..."
railway up

# Get deployment URL
echo "🌐 Getting deployment URL..."
DEPLOYMENT_URL=$(railway domain)

echo "✅ Deployment complete!"
echo "🌐 Your beta testing URL: $DEPLOYMENT_URL"
echo ""
echo "📋 Next steps:"
echo "1. Configure environment variables in Railway dashboard"
echo "2. Add PostgreSQL database if needed"
echo "3. Test the deployment"
echo "4. Share URL with beta testers"
EOF

chmod +x deploy-railway.sh

print_success "Created deploy-railway.sh"

# Create a health check endpoint test
cat > test-health.sh << 'EOF'
#!/bin/bash

# Health Check Test Script
# Usage: ./test-health.sh [URL]

URL="${1:-http://localhost:8080}"

echo "🏥 Testing health endpoint: $URL/health"

response=$(curl -s -o /dev/null -w "%{http_code}" "$URL/health")

if [ "$response" = "200" ]; then
    echo "✅ Health check passed!"
    echo "🌐 Application is running at: $URL"
else
    echo "❌ Health check failed (HTTP $response)"
    echo "🔍 Check the application logs"
fi
EOF

chmod +x test-health.sh

print_success "Created test-health.sh"

print_success "✅ Manual deployment setup complete!"

echo ""
echo "📋 Next Steps:"
echo "1. Go to [railway.app](https://railway.app)"
echo "2. Create a new project"
echo "3. Connect this GitHub repository"
echo "4. Configure environment variables from .env.railway.template"
echo "5. Deploy!"
echo ""
echo "📖 See DEPLOYMENT_INSTRUCTIONS.md for detailed steps"
echo "🔧 Run ./deploy-railway.sh if you get Railway CLI working"
echo ""
echo "🚀 Ready for beta testing deployment!"
