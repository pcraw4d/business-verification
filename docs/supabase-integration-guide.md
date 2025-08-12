# Supabase Integration Guide for Beta Testing

## 🔗 Integrating Supabase with Railway Deployment

This guide explains how to integrate Supabase with your Railway deployment for the KYB Platform beta testing.

## 🏗️ Architecture Options

### Option 1: External Supabase Database (Recommended)

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Railway App   │────│   Supabase      │────│   Supabase      │
│   (Web + API)   │    │   Database      │    │   Auth          │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**Benefits:**
- ✅ Persistent data across deployments
- ✅ Supabase's built-in auth and real-time features
- ✅ No data loss during app updates
- ✅ Advanced database features (RLS, functions, etc.)

### Option 2: Railway PostgreSQL + Supabase Features

```
┌─────────────────┐    ┌─────────────────┐
│   Railway App   │────│   Railway       │
│   (Web + API)   │    │   PostgreSQL    │
└─────────────────┘    └─────────────────┘
         │
         └─────────────┐
                       │
              ┌─────────────────┐
              │   Supabase      │
              │   (Auth only)   │
              └─────────────────┘
```

**Benefits:**
- ✅ Lower latency (database in same network)
- ✅ Railway's managed PostgreSQL
- ✅ Still get Supabase auth features

## 🚀 Setup Instructions

### Step 1: Create Supabase Project

1. **Go to [supabase.com](https://supabase.com)**
2. **Create a new project**
3. **Note down your project URL and API keys**

### Step 2: Configure Supabase Database

#### **Create Database Schema**

```sql
-- Users table (extends Supabase auth.users)
CREATE TABLE public.profiles (
    id UUID REFERENCES auth.users(id) PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT,
    role TEXT CHECK (role IN ('compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'other')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Business classifications table
CREATE TABLE public.business_classifications (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    business_name TEXT NOT NULL,
    website_url TEXT,
    description TEXT,
    primary_industry JSONB,
    secondary_industries JSONB,
    confidence_score DECIMAL(3,2),
    classification_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Risk assessments table
CREATE TABLE public.risk_assessments (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    business_id UUID REFERENCES public.business_classifications(id),
    risk_factors JSONB,
    risk_score DECIMAL(3,2),
    risk_level TEXT CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    assessment_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Compliance checks table
CREATE TABLE public.compliance_checks (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    business_id UUID REFERENCES public.business_classifications(id),
    compliance_frameworks JSONB,
    compliance_status JSONB,
    gap_analysis JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Feedback table
CREATE TABLE public.feedback (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    feedback_type TEXT CHECK (feedback_type IN ('bug', 'feature', 'improvement', 'general')),
    message TEXT NOT NULL,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable Row Level Security (RLS)
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.business_classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.risk_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.compliance_checks ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.feedback ENABLE ROW LEVEL SECURITY;

-- Create RLS policies
CREATE POLICY "Users can view own profile" ON public.profiles
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON public.profiles
    FOR UPDATE USING (auth.uid() = id);

CREATE POLICY "Users can view own classifications" ON public.business_classifications
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own classifications" ON public.business_classifications
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can view own risk assessments" ON public.risk_assessments
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own risk assessments" ON public.risk_assessments
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can view own compliance checks" ON public.compliance_checks
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own compliance checks" ON public.compliance_checks
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can view own feedback" ON public.feedback
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own feedback" ON public.feedback
    FOR INSERT WITH CHECK (auth.uid() = user_id);

-- Create indexes for better performance
CREATE INDEX idx_business_classifications_user_id ON public.business_classifications(user_id);
CREATE INDEX idx_business_classifications_created_at ON public.business_classifications(created_at);
CREATE INDEX idx_risk_assessments_user_id ON public.risk_assessments(user_id);
CREATE INDEX idx_compliance_checks_user_id ON public.compliance_checks(user_id);
CREATE INDEX idx_feedback_user_id ON public.feedback(user_id);
```

#### **Set Up Authentication**

1. **Configure Auth Settings in Supabase Dashboard**
2. **Set up email templates for verification**
3. **Configure OAuth providers (if needed)**

### Step 3: Update Railway Deployment

#### **Set Environment Variables**

```bash
# Set Supabase credentials
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_ANON_KEY="your-anon-key"
export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"

# Deploy with Supabase integration
./scripts/deploy-beta-railway.sh
```

#### **Update Application Configuration**

Create `configs/beta/supabase-config.yaml`:

```yaml
supabase:
  url: ${SUPABASE_URL}
  anon_key: ${SUPABASE_ANON_KEY}
  service_role_key: ${SUPABASE_SERVICE_ROLE_KEY}
  
database:
  type: "supabase"
  connection_string: "postgresql://postgres:[password]@[host]:5432/postgres"
  
auth:
  provider: "supabase"
  jwt_secret: ${SUPABASE_JWT_SECRET}
  session_duration: "24h"
  
features:
  real_time: true
  edge_functions: true
  storage: false  # Enable if needed for file uploads
```

### Step 4: Update Application Code

#### **Add Supabase Client**

```go
// internal/database/supabase.go
package database

import (
    "github.com/supabase-community/supabase-go"
    "os"
)

type SupabaseClient struct {
    client *supabase.Client
}

func NewSupabaseClient() (*SupabaseClient, error) {
    supabaseURL := os.Getenv("SUPABASE_URL")
    supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
    
    client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
    if err != nil {
        return nil, err
    }
    
    return &SupabaseClient{client: client}, nil
}

func (s *SupabaseClient) SaveClassification(userID string, classification *BusinessClassification) error {
    _, err := s.client.DB.From("business_classifications").Insert(classification).Execute()
    return err
}

func (s *SupabaseClient) GetUserClassifications(userID string) ([]BusinessClassification, error) {
    var classifications []BusinessClassification
    _, err := s.client.DB.From("business_classifications").
        Select("*").
        Eq("user_id", userID).
        Execute(&classifications)
    return classifications, err
}
```

#### **Update Authentication Handler**

```go
// internal/auth/supabase_auth.go
package auth

import (
    "github.com/supabase-community/supabase-go"
)

type SupabaseAuth struct {
    client *supabase.Client
}

func (s *SupabaseAuth) AuthenticateUser(email, password string) (*User, error) {
    // Use Supabase auth
    auth, err := s.client.Auth.SignIn(context.Background(), supabase.UserCredentials{
        Email:    email,
        Password: password,
    })
    if err != nil {
        return nil, err
    }
    
    return &User{
        ID:    auth.User.ID,
        Email: auth.User.Email,
        Role:  auth.User.UserMetadata["role"].(string),
    }, nil
}
```

### Step 5: Update Web Interface

#### **Add Supabase JavaScript Client**

```html
<!-- Add to web/index.html -->
<script src="https://cdn.jsdelivr.net/npm/@supabase/supabase-js@2"></script>
<script>
    // Initialize Supabase client
    const supabaseUrl = 'https://your-project.supabase.co'
    const supabaseKey = 'your-anon-key'
    const supabase = supabase.createClient(supabaseUrl, supabaseKey)
    
    // Update authentication functions
    async function loginWithSupabase(email, password) {
        const { data, error } = await supabase.auth.signInWithPassword({
            email: email,
            password: password
        })
        
        if (error) {
            throw error
        }
        
        return data
    }
    
    async function registerWithSupabase(email, password, userData) {
        const { data, error } = await supabase.auth.signUp({
            email: email,
            password: password,
            options: {
                data: userData
            }
        })
        
        if (error) {
            throw error
        }
        
        return data
    }
    
    // Save classification to Supabase
    async function saveClassification(classificationData) {
        const { data, error } = await supabase
            .from('business_classifications')
            .insert([classificationData])
        
        if (error) {
            throw error
        }
        
        return data
    }
</script>
```

## 🔧 Configuration Options

### **Environment Variables**

```bash
# Required for Supabase integration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Optional
SUPABASE_JWT_SECRET=your-jwt-secret
SUPABASE_DB_PASSWORD=your-db-password
```

### **Database Connection String**

```
postgresql://postgres:[password]@[host]:5432/postgres?sslmode=require
```

## 📊 Monitoring and Analytics

### **Supabase Dashboard**

1. **Database Analytics**
   - Query performance
   - Connection usage
   - Storage usage

2. **Auth Analytics**
   - User sign-ups
   - Login attempts
   - Email verification rates

3. **Real-time Analytics**
   - Active connections
   - Message throughput

### **Custom Analytics**

```sql
-- User activity analytics
SELECT 
    DATE(created_at) as date,
    COUNT(*) as classifications,
    COUNT(DISTINCT user_id) as active_users
FROM business_classifications
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Feature usage analytics
SELECT 
    feedback_type,
    COUNT(*) as count
FROM feedback
GROUP BY feedback_type
ORDER BY count DESC;
```

## 🔒 Security Considerations

### **Row Level Security (RLS)**

- All tables have RLS enabled
- Users can only access their own data
- Admin users can access all data

### **API Security**

- Use service role key only on server-side
- Use anon key for client-side operations
- Implement proper JWT validation

### **Data Protection**

- Enable SSL for all connections
- Use environment variables for secrets
- Regular security audits

## 🚀 Deployment Checklist

### **Pre-Deployment**
- [ ] Supabase project created
- [ ] Database schema created
- [ ] RLS policies configured
- [ ] Environment variables set
- [ ] Application code updated

### **Deployment**
- [ ] Railway deployment with Supabase integration
- [ ] Health checks passing
- [ ] Authentication working
- [ ] Data persistence verified

### **Post-Deployment**
- [ ] Monitor database performance
- [ ] Check auth flows
- [ ] Verify data isolation
- [ ] Test backup/restore

## 💰 Cost Optimization

### **Supabase Free Tier Limits**
- 500MB database
- 2GB bandwidth
- 50,000 monthly active users

### **Upgrade Triggers**
- Database size > 400MB
- Bandwidth > 1.5GB
- Users > 40,000

### **Cost Monitoring**
```bash
# Check Supabase usage
curl -X GET "https://api.supabase.com/v1/projects/{project_id}/usage" \
  -H "Authorization: Bearer {service_role_key}"
```

## 🐛 Troubleshooting

### **Common Issues**

#### **Connection Errors**
```bash
# Check Supabase status
curl https://status.supabase.com/api/v2/status.json

# Verify environment variables
echo $SUPABASE_URL
echo $SUPABASE_ANON_KEY
```

#### **Authentication Issues**
- Verify JWT secret configuration
- Check email templates in Supabase dashboard
- Ensure RLS policies are correct

#### **Performance Issues**
- Monitor query performance in Supabase dashboard
- Add database indexes for slow queries
- Optimize application queries

This integration guide ensures your Railway deployment works seamlessly with Supabase for a robust beta testing environment.
