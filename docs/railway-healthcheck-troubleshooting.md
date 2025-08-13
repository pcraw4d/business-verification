# Railway Healthcheck Failure Troubleshooting Guide

## 🚨 **Healthcheck Failed - Quick Diagnosis**

When Railway shows "Healthcheck failed!" with level "info", it means the application cannot respond to health checks at `/health` within 300 seconds.

## 🔍 **Step-by-Step Troubleshooting**

### **Step 1: Check Railway Logs**

First, check the Railway deployment logs to identify the specific error:

```bash
# View recent logs
railway logs

# View logs with timestamps
railway logs --follow

# View logs for specific service
railway logs --service your-service-name
```

### **Step 2: Verify Environment Variables**

The application requires these **CRITICAL** environment variables:

#### **Required Variables (Check in Railway Dashboard → Variables):**

```bash
# Authentication (CRITICAL)
JWT_SECRET=your-32-character-secret
ENCRYPTION_KEY=your-32-character-key
API_SECRET=your-24-character-secret

# Database (CRITICAL)
DATABASE_URL=postgresql://user:password@host:port/database

# Supabase (if using)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Server Configuration
PORT=8080
HOST=0.0.0.0
```

#### **Quick Environment Check Script:**

```bash
# Add this to your Railway startup script to debug
echo "🔍 Checking environment variables..."
echo "PORT: $PORT"
echo "HOST: $HOST"
echo "JWT_SECRET: ${JWT_SECRET:0:8}..."
echo "DATABASE_URL: ${DATABASE_URL:0:20}..."
echo "SUPABASE_URL: $SUPABASE_URL"
```

### **Step 3: Fix Common Issues**

#### **Issue 1: Missing JWT_SECRET**
```bash
# Generate a new JWT secret
openssl rand -base64 32

# Add to Railway variables
railway variables set JWT_SECRET=your-generated-secret
```

#### **Issue 2: Database Connection Problems**
```bash
# Check if PostgreSQL service is added
railway service add postgresql

# Verify DATABASE_URL is set automatically
railway variables list | grep DATABASE_URL
```

#### **Issue 3: Port Configuration**
```bash
# Ensure PORT is set to 8080
railway variables set PORT=8080
railway variables set HOST=0.0.0.0
```

### **Step 4: Update Railway Configuration**

#### **Update railway.json for Better Healthcheck:**

```json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile.beta"
  },
  "deploy": {
    "startCommand": "./railway-startup.sh",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 300,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

#### **Update Dockerfile.beta for Better Healthcheck:**

```dockerfile
# Add this to your Dockerfile.beta
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=5 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

### **Step 5: Enhanced Startup Script**

#### **Update scripts/railway-startup.sh:**

```bash
#!/bin/bash

# KYB Platform - Railway Startup Script
set -e

echo "🚀 KYB Platform Railway Startup"
echo "================================"

# Function to check environment variables
check_environment() {
    echo "🔍 Checking environment variables..."
    
    required_vars=(
        "JWT_SECRET"
        "DATABASE_URL"
    )
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            echo "❌ Required environment variable $var is not set"
            echo "💡 Add this variable in Railway Dashboard → Variables"
            exit 1
        fi
        echo "✅ $var is set"
    done
    
    echo "✅ All required environment variables are set"
}

# Function to wait for database with better error handling
wait_for_database() {
    echo "⏳ Waiting for database connection..."
    
    if [ -z "$DATABASE_URL" ]; then
        echo "⚠️ No DATABASE_URL provided, skipping database check"
        return 0
    fi
    
    # Extract database connection details
    export DATABASE_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\).*/\1/p')
    export DATABASE_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
    export DATABASE_USER=$(echo $DATABASE_URL | sed -n 's/.*:\/\/\([^:]*\):.*/\1/p')
    
    echo "🔗 Database host: $DATABASE_HOST"
    echo "🔗 Database port: $DATABASE_PORT"
    echo "🔗 Database user: $DATABASE_USER"
    
    # Try to connect to database
    for i in {1..30}; do
        if pg_isready -h $DATABASE_HOST -p $DATABASE_PORT -U $DATABASE_USER > /dev/null 2>&1; then
            echo "✅ Database connection established"
            return 0
        fi
        echo "⏳ Attempt $i/30: Database not ready yet..."
        sleep 2
    done
    
    echo "❌ Database connection failed after 30 attempts"
    echo "💡 Check if PostgreSQL service is added to Railway"
    return 1
}

# Function to start the application
start_application() {
    echo "🚀 Starting KYB Platform application..."
    
    # Set proper permissions
    chmod +x ./kyb-platform
    
    # Start the application with proper signal handling
    exec ./kyb-platform
}

# Main execution
main() {
    echo "📋 Starting KYB Platform Railway deployment..."
    
    # Check environment variables
    check_environment
    
    # Wait for database
    wait_for_database
    
    # Start the application
    start_application
}

# Run main function
main "$@"
```

### **Step 6: Test Healthcheck Locally**

#### **Create a local test script:**

```bash
#!/bin/bash
# test-railway-health.sh

echo "🏥 Testing Railway-style healthcheck..."

# Start application in background
./kyb-platform &
APP_PID=$!

# Wait for startup
echo "⏳ Waiting for application to start..."
sleep 10

# Test health endpoint
echo "🔍 Testing /health endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)

if [ "$response" = "200" ]; then
    echo "✅ Health check passed!"
    echo "🌐 Application is running at: http://localhost:8080"
else
    echo "❌ Health check failed (HTTP $response)"
    echo "🔍 Check application logs"
fi

# Cleanup
kill $APP_PID
```

### **Step 7: Railway-Specific Fixes**

#### **Add PostgreSQL Service:**
```bash
# Add PostgreSQL to your Railway project
railway service add postgresql

# This automatically sets DATABASE_URL
```

#### **Set Required Variables:**
```bash
# Set critical environment variables
railway variables set JWT_SECRET=$(openssl rand -base64 32)
railway variables set ENCRYPTION_KEY=$(openssl rand -base64 32)
railway variables set API_SECRET=$(openssl rand -base64 18)

# Set server configuration
railway variables set PORT=8080
railway variables set HOST=0.0.0.0
```

#### **Redeploy:**
```bash
# Redeploy with new configuration
railway up
```

## 🐛 **Common Error Messages & Solutions**

### **"Failed to connect to database"**
- **Solution**: Add PostgreSQL service to Railway
- **Check**: Verify DATABASE_URL is set automatically

### **"JWT secret is required"**
- **Solution**: Set JWT_SECRET environment variable
- **Generate**: `openssl rand -base64 32`

### **"Port already in use"**
- **Solution**: Ensure PORT=8080 and HOST=0.0.0.0
- **Check**: No other services using port 8080

### **"Permission denied"**
- **Solution**: Ensure startup script is executable
- **Fix**: `chmod +x scripts/railway-startup.sh`

## 📞 **Getting Help**

### **1. Check Railway Status:**
```bash
railway status
railway logs --follow
```

### **2. Test Locally:**
```bash
# Test with Railway environment
railway run ./kyb-platform
```

### **3. Verify Configuration:**
```bash
# List all variables
railway variables list
```

### **4. Contact Support:**
- Check Railway documentation
- Review application logs
- Verify all environment variables are set

## ✅ **Success Checklist**

- [ ] PostgreSQL service added to Railway
- [ ] JWT_SECRET environment variable set
- [ ] DATABASE_URL automatically configured
- [ ] PORT=8080 and HOST=0.0.0.0 set
- [ ] Application starts without errors
- [ ] Health endpoint responds with 200 OK
- [ ] Railway healthcheck passes

## 🔄 **Quick Recovery Commands**

```bash
# Regenerate and set secrets
railway variables set JWT_SECRET=$(openssl rand -base64 32)
railway variables set ENCRYPTION_KEY=$(openssl rand -base64 32)

# Add PostgreSQL if missing
railway service add postgresql

# Redeploy
railway up

# Check status
railway status
railway logs --follow
```

---

**Remember**: The most common cause of healthcheck failures is missing environment variables, especially `JWT_SECRET` and database configuration. Always check Railway logs first to identify the specific error.
