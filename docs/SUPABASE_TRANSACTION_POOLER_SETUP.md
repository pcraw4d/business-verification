# Supabase Transaction Pooler Setup Guide

**Date**: November 13, 2025  
**Status**: ✅ **RECOMMENDED FOR RAILWAY**

---

## 🎯 Recommendation: Use Transaction Pooler

**For Risk Assessment Service**: ✅ **Transaction Pooler** is the correct choice.

**Why Transaction Pooler?**
- ✅ Ideal for stateless applications (microservices)
- ✅ Perfect for brief, isolated database queries
- ✅ Efficient connection management for serverless/containerized environments
- ✅ Better suited for Railway deployments
- ✅ Handles high concurrency better

**Why NOT Session Pooler?**
- ❌ Designed for persistent connections (not our use case)
- ❌ Only recommended as alternative to Direct Connection
- ❌ Better for traditional backend servers with long-lived connections

---

## 📋 Setup Steps

### Step 1: Get Transaction Pooler Connection String from Supabase

1. **Open Supabase Dashboard**
   - Go to your Supabase project
   - Navigate to **Settings** → **Database**

2. **Find Connection Pooling Section**
   - Look for "Connection Pooling" or "Pooler" settings
   - Select **Transaction pooler** (should be selected by default)

3. **Copy Connection String**
   - Find the connection string for Transaction pooler
   - Format: `postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres?pgbouncer=true`
   - Or use the "Connection string" provided by Supabase

4. **Get Database Password**
   - The password is **NOT** the service role key
   - It's the database password (set when creating the project)
   - Find it in Supabase Dashboard → Settings → Database → Database Password
   - Or reset it if needed

---

### Step 2: Set DATABASE_URL in Railway

1. **Open Railway Dashboard**
   - Go to your Railway project
   - Navigate to **Project Settings** → **Variables**

2. **Add DATABASE_URL Variable**
   - Click **"New Variable"** or **"Add Variable"**
   - Variable Name: `DATABASE_URL`
   - Value: Paste the Transaction Pooler connection string from Supabase
   - Scope: **Shared** (or specific to risk-assessment-service)

3. **Example Format**:
   ```
   DATABASE_URL=postgresql://postgres.qpqhuqqmkjxsltzshfam:[YOUR-DB-PASSWORD]@aws-0-us-east-1.pooler.supabase.com:6543/postgres?pgbouncer=true
   ```

   **Important**: Replace:
   - `[YOUR-DB-PASSWORD]` with your actual database password
   - `qpqhuqqmkjxsltzshfam` with your project reference
   - `us-east-1` with your actual region (if different)

---

### Step 3: Verify Connection String Format

**Transaction Pooler Format**:
```
postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres?pgbouncer=true
```

**Key Components**:
- **Protocol**: `postgresql://`
- **User**: `postgres.[project-ref]` (note the dot before project-ref)
- **Password**: Your database password (NOT service role key)
- **Host**: `aws-0-[region].pooler.supabase.com`
- **Port**: `6543` (Transaction pooler port)
- **Database**: `postgres`
- **Parameters**: `?pgbouncer=true` (enables connection pooling)

---

### Step 4: Optional - Set SUPABASE_DB_PASSWORD

If you want the code to auto-construct the connection string (not recommended), you can set:

```
SUPABASE_DB_PASSWORD=[your-database-password]
```

**But it's better to set `DATABASE_URL` directly** with the full connection string from Supabase.

---

## ✅ Verification

### 1. Check Environment Variable

**In Railway Dashboard**:
- Go to **Risk Assessment Service** → **Variables** tab
- Verify `DATABASE_URL` is set with Transaction Pooler connection string
- Check that it includes `:6543` (Transaction pooler port)
- Verify it includes `?pgbouncer=true`

### 2. Check Service Logs

**After Redeploy**, you should see:
```
✅ Database connection established with performance optimizations
```

**If connection fails**, you'll see:
```
Failed to initialize database with performance optimizations - continuing without database
error: "failed to ping database: ..."
```

### 3. Test Database Features

After successful connection:
- Performance components should initialize
- Connection pool should be created
- Query optimizer should be available
- Database-dependent features should work

---

## 🔍 Troubleshooting

### Issue: Connection Still Fails

**Check**:
1. **Password is Correct**: Database password, not service role key
2. **Port is 6543**: Transaction pooler uses port 6543
3. **Format is Correct**: Includes `?pgbouncer=true`
4. **Project Reference**: Matches your Supabase project
5. **Region**: Matches your Supabase region

### Issue: "Invalid Password"

**Solution**:
- Verify you're using the **database password**, not service role key
- Reset database password in Supabase if needed
- Update `DATABASE_URL` in Railway

### Issue: "Connection Refused"

**Solution**:
- Verify port is `6543` (Transaction pooler)
- Check that Transaction pooler is enabled in Supabase
- Verify network connectivity from Railway

---

## 📊 Transaction Pooler vs Session Pooler

| Feature | Transaction Pooler | Session Pooler |
|---------|-------------------|----------------|
| **Port** | 6543 | 5432 |
| **Use Case** | Stateless, brief queries | Persistent connections |
| **Prepared Statements** | ❌ Not supported | ✅ Supported |
| **Connection Reuse** | ✅ High (per transaction) | ✅ Per session |
| **Ideal For** | Serverless, microservices | Traditional backends |
| **Our Service** | ✅ **RECOMMENDED** | ❌ Not recommended |

---

## 🎯 Summary

**Action Required**:
1. ✅ Use **Transaction Pooler** (port 6543)
2. ✅ Get connection string from Supabase dashboard
3. ✅ Set `DATABASE_URL` in Railway with full connection string
4. ✅ Use database password (not service role key)
5. ✅ Include `?pgbouncer=true` parameter

**Expected Result**:
- Database connection succeeds
- Performance components initialize
- Database-dependent features enabled
- No more IPv6 connection errors

---

**Last Updated**: November 13, 2025

