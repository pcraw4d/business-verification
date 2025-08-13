# Railway Healthcheck Failure - Immediate Fix Guide

## 🚨 **Issue Summary**

**Error**: `Healthcheck failed!` with level `"info"`

**Cause**: Railway cannot reach the `/health` endpoint within 300 seconds, usually due to:
1. Missing environment variables (most common)
2. Database connection issues
3. Application startup failures
4. Port configuration problems

## 🚀 **Immediate Solutions**

### **Step 1: Quick Diagnosis**

Run our diagnostic script to identify the specific issue:

```bash
./scripts/diagnose-railway-issue.sh
```

This will check:
- Railway CLI and login status
- Environment variables
- PostgreSQL service
- Recent logs
- Health endpoint accessibility

### **Step 2: Generate Required Secrets**

If you need to generate new security keys:

```bash
./scripts/fix-railway-deployment.sh generate
```

This creates:
- JWT_SECRET (32 characters)
- ENCRYPTION_KEY (32 characters)  
- API_SECRET (24 characters)

### **Step 3: Fix Common Issues**

Run the automated fix script:

```bash
./scripts/fix-railway-deployment.sh fix
```

This will:
- Set PORT=8080 and HOST=0.0.0.0
- Check PostgreSQL service
- Verify environment variables

### **Step 4: Manual Environment Variable Setup**

If the automated fix doesn't work, manually set these in Railway Dashboard → Variables:

#### **Critical Variables (Required):**
```bash
JWT_SECRET=your-32-character-secret
DATABASE_URL=postgresql://user:password@host:port/database
PORT=8080
HOST=0.0.0.0
```

#### **Optional Variables:**
```bash
ENCRYPTION_KEY=your-32-character-key
API_SECRET=your-24-character-secret
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
```

### **Step 5: Add PostgreSQL Service**

If you don't have a database:

```bash
railway service add postgresql
```

This automatically sets `DATABASE_URL`.

### **Step 6: Redeploy**

After fixing the issues:

```bash
railway up
```

### **Step 7: Verify Fix**

Check if the healthcheck passes:

```bash
./scripts/fix-railway-deployment.sh health
```

## 🔍 **Common Root Causes**

### **1. Missing JWT_SECRET (Most Common)**
- **Error**: "JWT secret is required"
- **Fix**: Set JWT_SECRET environment variable
- **Generate**: `openssl rand -base64 32`

### **2. Missing Database**
- **Error**: "Failed to connect to database"
- **Fix**: Add PostgreSQL service to Railway
- **Check**: Verify DATABASE_URL is set

### **3. Port Issues**
- **Error**: "Port already in use" or connection refused
- **Fix**: Set PORT=8080 and HOST=0.0.0.0
- **Check**: No other services using port 8080

### **4. Application Startup Failures**
- **Error**: Various startup errors in logs
- **Fix**: Check Railway logs for specific errors
- **Command**: `railway logs --follow`

## 📋 **Troubleshooting Checklist**

- [ ] Run diagnostic script: `./scripts/diagnose-railway-issue.sh`
- [ ] Check Railway logs: `railway logs --follow`
- [ ] Verify environment variables: `railway variables list`
- [ ] Add PostgreSQL service if missing: `railway service add postgresql`
- [ ] Set required variables (JWT_SECRET, PORT, HOST)
- [ ] Redeploy: `railway up`
- [ ] Test health endpoint: `./scripts/fix-railway-deployment.sh health`

## 🛠️ **Advanced Troubleshooting**

### **Check Application Logs**
```bash
railway logs --follow
```

### **Test Health Endpoint Manually**
```bash
# Get your app URL
railway status

# Test health endpoint
curl https://your-app.railway.app/health
```

### **Verify Environment Variables**
```bash
railway variables list
```

### **Check Service Status**
```bash
railway status
railway service list
```

## 📞 **Getting Help**

### **1. Check Documentation**
- [Railway Healthcheck Troubleshooting Guide](railway-healthcheck-troubleshooting.md)
- [Railway Deployment Keys Guide](railway-deployment-keys-guide.md)

### **2. Use Diagnostic Tools**
- `./scripts/diagnose-railway-issue.sh` - Comprehensive diagnosis
- `./scripts/fix-railway-deployment.sh` - Automated fixes

### **3. Check Railway Status**
- Railway Dashboard → Your Project → Deployments
- Railway Dashboard → Your Project → Variables
- Railway Dashboard → Your Project → Logs

## ✅ **Success Indicators**

When the fix is successful, you should see:
- ✅ Railway deployment status: "Deployed"
- ✅ Health endpoint returns HTTP 200
- ✅ Application logs show successful startup
- ✅ No error messages in Railway logs

## 🔄 **Quick Recovery Commands**

```bash
# Full diagnostic and fix workflow
./scripts/diagnose-railway-issue.sh
./scripts/fix-railway-deployment.sh generate
./scripts/fix-railway-deployment.sh fix
railway up
./scripts/fix-railway-deployment.sh health
```

---

**Remember**: The most common cause of healthcheck failures is missing environment variables, especially `JWT_SECRET`. Always start with the diagnostic script to identify the specific issue.
