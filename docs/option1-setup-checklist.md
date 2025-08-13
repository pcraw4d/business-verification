# Option 1: External Supabase Database - Setup Checklist

## ✅ **What We've Completed:**

### **Code Implementation:**
- ✅ **Supabase Client** - `internal/database/supabase.go`
- ✅ **Factory Integration** - Updated `internal/factory.go`
- ✅ **Configuration** - Supabase config in `internal/config/config.go`
- ✅ **Railway Setup** - Dockerfile, startup script, environment variables
- ✅ **Database Schema** - `scripts/setup-supabase-schema.sql`

### **Deployment Files:**
- ✅ **Dockerfile.beta** - Railway deployment with PostgreSQL client
- ✅ **Startup Script** - `scripts/railway-startup.sh`
- ✅ **Environment Template** - `.env.railway.full`
- ✅ **Configuration** - `configs/beta/railway-config.yaml`

## 📋 **What You Need to Do:**

### **Step 1: Set Up Supabase Database Schema**
1. **Go to your Supabase project dashboard**
2. **Navigate to SQL Editor**
3. **Copy and paste the contents of `scripts/setup-supabase-schema.sql`**
4. **Click "Run" to execute the schema**
5. **Verify tables are created in the Table Editor**

### **Step 2: Configure Railway Environment Variables**
In your Railway project Variables tab, add:

#### **Required Supabase Variables:**
```
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
SUPABASE_ENABLED=true
```

#### **Required Security Keys:**
```
JWT_SECRET=GvEHhjPwx6xttws0qScCGzDBMhQ0ORGh
ENCRYPTION_KEY=NUzTbkubsGQPpYysPitxZK4jTPwLCWR
API_SECRET=4gqjV6OM2R2T6DIjdjaspp7G
```

#### **Application Settings:**
```
ENVIRONMENT=beta
BETA_MODE=true
PORT=8080
ANALYTICS_ENABLED=true
FEEDBACK_COLLECTION=true
LOG_LEVEL=info
CORS_ORIGIN=https://your-app.railway.app
```

### **Step 3: Deploy to Railway**
1. **Railway will auto-detect the new commit**
2. **Build will use the updated factory with Supabase client**
3. **Startup script will initialize the database**
4. **Health checks should pass**

### **Step 4: Verify Integration**
Test these endpoints after deployment:

```bash
# Health check
curl https://your-app.railway.app/health

# Test classification (should use Supabase)
curl -X POST https://your-app.railway.app/v1/classification \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company", "website_url": "https://example.com"}'

# Test authentication
curl -X POST https://your-app.railway.app/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

## 🔍 **Verification Steps:**

### **1. Check Railway Logs**
- Look for "Connecting to Supabase" messages
- Verify "Successfully connected to Supabase"
- Check for any database initialization errors

### **2. Check Supabase Dashboard**
- **Table Editor**: Verify all tables are created
- **Authentication**: Check if users can register/login
- **Logs**: Monitor API requests and database queries

### **3. Test Data Persistence**
- Create a classification via API
- Check if it appears in Supabase Table Editor
- Verify Row Level Security is working

## 🚨 **Troubleshooting:**

### **If Health Check Fails:**
1. Check Railway logs for specific errors
2. Verify all environment variables are set
3. Ensure Supabase schema is created
4. Check Supabase project is active

### **If Database Connection Fails:**
1. Verify `SUPABASE_URL` format
2. Check `SUPABASE_ANON_KEY` and `SUPABASE_SERVICE_ROLE_KEY`
3. Ensure Supabase project is not paused
4. Check network connectivity

### **If Authentication Fails:**
1. Verify Supabase auth is enabled
2. Check email templates in Supabase dashboard
3. Ensure RLS policies are correct
4. Test with Supabase dashboard directly

## 🎯 **Success Criteria:**

Your Option 1 setup is complete when:
- ✅ Railway deployment succeeds
- ✅ Health check passes consistently
- ✅ Supabase database schema is created
- ✅ API endpoints work with Supabase backend
- ✅ Data persists in Supabase tables
- ✅ Authentication works via Supabase
- ✅ Row Level Security is active
- ✅ All 8 major features are functional

## 📞 **Next Steps:**

After successful setup:
1. **Test all features** thoroughly
2. **Invite beta testers** to the platform
3. **Monitor Supabase usage** and costs
4. **Collect feedback** from users
5. **Scale as needed** based on usage

## 🔗 **Useful Links:**

- **Supabase Dashboard**: https://supabase.com/dashboard
- **Railway Dashboard**: https://railway.app/dashboard
- **Supabase Documentation**: https://supabase.com/docs
- **Railway Documentation**: https://docs.railway.app

**Your Option 1 setup is ready! Just follow the checklist above to complete the integration.** 🚀
