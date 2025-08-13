# Railway Full-Featured Deployment Guide

## 🚀 Complete KYB Platform Deployment

### Prerequisites
- Railway account
- Supabase account  
- GitHub repository access

### Step 1: Create Supabase Project
1. Go to supabase.com
2. Create new project
3. Get API keys from Settings → API
4. Run database schema from supabase-integration-guide.md

### Step 2: Create Railway Project
1. Go to railway.app
2. New Project → Deploy from GitHub
3. Select your repository
4. Add PostgreSQL database

### Step 3: Environment Variables
Add to Railway Variables:
```
JWT_SECRET=GvEHhjPwx6xttws0qScCGzDBMhQ0ORGh
ENCRYPTION_KEY=NUzTbkubsGQPpYysPitxZK4jTPwLCWR
API_SECRET=4gqjV6OM2R2T6DIjdjaspp7G
ENVIRONMENT=beta
BETA_MODE=true
PORT=8080
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
SUPABASE_ENABLED=true
ANALYTICS_ENABLED=true
FEEDBACK_COLLECTION=true
LOG_LEVEL=info
CORS_ORIGIN=https://your-app.railway.app
```

### Step 4: Deploy
1. Railway auto-detects Dockerfile.beta
2. Builds and deploys application
3. Runs health checks
4. Application goes live

### Step 5: Verify
Test endpoints:
- GET /health
- GET /v1/health  
- POST /v1/classification
- POST /v1/auth/register
- GET /beta

### Features Available
✅ Business Classification
✅ Risk Assessment  
✅ Compliance Framework
✅ Authentication & Authorization
✅ API Gateway
✅ Database Management
✅ Observability & Monitoring
✅ Security & Compliance

### Monitoring
- Railway dashboard for logs and metrics
- Supabase dashboard for database and auth
- Application health checks
- Performance monitoring

### Beta Testing
- Share Railway URL with testers
- Collect feedback via web interface
- Monitor usage and performance
- Iterate based on feedback
