# 🚀 **Supabase Project Setup Guide for Keyword Classification System**

## 📋 **Overview**

This guide will walk you through setting up a Supabase project and implementing the 7-table keyword classification schema. The system will provide a robust foundation for dynamic keyword management and industry classification.

## 🎯 **Prerequisites**

- **Supabase Account**: Free tier account at [supabase.com](https://supabase.com)
- **Go Environment**: Go 1.22+ with all dependencies installed
- **Environment Variables**: Ready to configure with real Supabase credentials

## 🏗️ **Step 1: Create Supabase Project**

### **1.1 Create New Project**
1. **Login to Supabase Dashboard** at [app.supabase.com](https://app.supabase.com)
2. **Click "New Project"**
3. **Choose Organization** (or create one if needed)
4. **Project Details**:
   - **Name**: `keyword-classification-system` (or your preferred name)
   - **Database Password**: Generate a strong password (save this!)
   - **Region**: Choose closest to your users
   - **Pricing Plan**: Select **Free tier** (500MB storage, 2 projects)

### **1.2 Wait for Project Setup**
- **Database Creation**: ~2-3 minutes
- **API Generation**: Automatic
- **Dashboard Access**: Available immediately

## 🔑 **Step 2: Get Project Credentials**

### **2.1 Project Settings**
1. **Go to Project Dashboard**
2. **Navigate to Settings → API**
3. **Copy the following values**:

```bash
# Project URL
SUPABASE_URL=https://your-project-id.supabase.co

# Anon/Public Key (for client-side access)
SUPABASE_API_KEY=your-anon-key-here

# Service Role Key (for admin operations)
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key-here

# JWT Secret (for authentication)
SUPABASE_JWT_SECRET=your-jwt-secret-here
```

### **2.2 Update Environment File**
Update `configs/development.env` with real values:

```env
# Supabase Configuration
SUPABASE_URL=https://your-project-id.supabase.co
SUPABASE_API_KEY=your-anon-key-here
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key-here
SUPABASE_JWT_SECRET=your-jwt-secret-here
```

## 🗄️ **Step 3: Apply Database Schema**

### **3.1 Open SQL Editor**
1. **In Supabase Dashboard**, go to **SQL Editor**
2. **Click "New Query"**

### **3.2 Apply Initial Schema**
1. **Copy the contents** of `supabase-migrations/001_initial_keyword_classification_schema.sql`
2. **Paste into SQL Editor**
3. **Click "Run"** to execute the migration

### **3.3 Verify Schema Creation**
After running the migration, you should see:
- **7 tables created** in the **Table Editor**
- **Indexes created** for performance
- **RLS policies** configured for security
- **Triggers** set up for automatic timestamps

## 🧪 **Step 4: Test the Setup**

### **4.1 Test Go Code Connection**
```bash
# Build and run the test script
go build -o test-supabase ./test-supabase-connection.go
./test-supabase
```

**Expected Output**:
```
🔍 Testing Supabase connection with new Go code...
📋 Configuration loaded:
   URL: https://your-project-id.supabase.co
   API Key: your-anon-key...
   Service Role Key: your-service-role-key...
✅ Supabase client created successfully
🔌 Connecting to Supabase at https://your-project-id.supabase.co
✅ Successfully connected to Supabase!
🔍 Testing repository operations...
✅ Found 0 industries in database
✅ Found 0 keywords matching 'tech'
✅ Business classified as: General Business (confidence: 50.0%)
🎉 Supabase connection test completed successfully!
```

### **4.2 Verify Database Tables**
In Supabase Dashboard → **Table Editor**, verify:
- ✅ `industries` table exists
- ✅ `industry_keywords` table exists
- ✅ `classification_codes` table exists
- ✅ `industry_patterns` table exists
- ✅ `keyword_weights` table exists
- ✅ `audit_logs` table exists
- ✅ `migrations` table exists

## 🌱 **Step 5: Seed Initial Data (Optional)**

### **5.1 Create Basic Industries**
```sql
-- Insert some basic industries for testing
INSERT INTO industries (name, description, category, confidence_threshold, is_active) VALUES
('Technology', 'Software, hardware, and IT services', 'traditional', 0.80, true),
('Healthcare', 'Medical services and health products', 'traditional', 0.80, true),
('Finance', 'Banking, insurance, and financial services', 'traditional', 0.80, true),
('Retail', 'Consumer goods and retail services', 'traditional', 0.80, true),
('Manufacturing', 'Industrial production and manufacturing', 'traditional', 0.80, true);
```

### **5.2 Add Sample Keywords**
```sql
-- Insert sample keywords for Technology industry
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active) VALUES
(1, 'software', 0.9, true),
(1, 'technology', 0.9, true),
(1, 'digital', 0.8, true),
(1, 'computer', 0.8, true),
(1, 'app', 0.7, true);

-- Insert sample keywords for Healthcare industry
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active) VALUES
(2, 'health', 0.9, true),
(2, 'medical', 0.9, true),
(2, 'doctor', 0.8, true),
(2, 'hospital', 0.8, true),
(2, 'pharmacy', 0.7, true);
```

## 🔒 **Step 6: Security Configuration**

### **6.1 Row Level Security (RLS)**
The migration script already configures RLS policies:
- **Public Read Access**: Anyone can read industries and keywords
- **Authenticated Write Access**: Only authenticated users can modify data
- **Admin Access**: Service role key has full access

### **6.2 API Access Control**
- **Client Applications**: Use `SUPABASE_API_KEY` (anon key)
- **Admin Operations**: Use `SUPABASE_SERVICE_ROLE_KEY`
- **Authentication**: JWT-based user authentication

## 📊 **Step 7: Monitor and Optimize**

### **7.1 Database Monitoring**
In Supabase Dashboard:
- **Database**: Monitor query performance
- **API**: Track API usage and limits
- **Storage**: Monitor database size (free tier: 500MB)

### **7.2 Performance Optimization**
- **Indexes**: Already created for common queries
- **Connection Pooling**: Configured for optimal performance
- **Query Optimization**: Monitor slow queries

## 🚨 **Troubleshooting**

### **Common Issues**

#### **1. Connection Failed**
```
❌ Failed to connect to Supabase: ping failed - database schema may not be initialized
```
**Solution**: Ensure the schema migration has been run successfully

#### **2. Environment Variables Not Set**
```
❌ SUPABASE_URL is required
```
**Solution**: Check `configs/development.env` and ensure all variables are set

#### **3. Permission Denied**
```
❌ permission denied for table industries
```
**Solution**: Verify RLS policies are correctly configured

### **Debug Steps**
1. **Check Environment Variables**: Verify all Supabase credentials
2. **Verify Schema**: Ensure all tables exist in Table Editor
3. **Check RLS Policies**: Verify security policies are active
4. **Test API Keys**: Verify keys have correct permissions

## 🎉 **Success Criteria**

✅ **Supabase project created** and accessible  
✅ **Database schema applied** with all 7 tables  
✅ **Go code connects successfully** to Supabase  
✅ **Repository operations work** without errors  
✅ **Security policies active** and protecting data  
✅ **Performance indexes created** for optimal queries  

## 🔄 **Next Steps**

After completing this setup:
1. **Test the complete system** with real data
2. **Implement keyword classification logic** in Go
3. **Add more industries and keywords** as needed
4. **Monitor performance** and optimize queries
5. **Scale up** when approaching free tier limits

---

**🎯 Ready to proceed?** Let's set up your Supabase project and get the keyword classification system running!
