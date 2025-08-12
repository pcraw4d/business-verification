# Beta Testing Deployment Guide

## 🚀 Quick Setup for Public Beta Testing

This guide will help you deploy the KYB Platform to a public URL so non-technical users can access it for beta testing.

## 📋 Prerequisites

- Docker and Docker Compose installed
- A domain name (optional, but recommended)
- Railway account (for quick deployment)

## 🎯 Deployment Options

### Option 1: Railway Deployment (Recommended - Quick & Easy)

Railway provides a simple way to deploy your application with a public URL.

#### Step 1: Install Railway CLI
```bash
npm install -g @railway/cli
```

#### Step 2: Login to Railway
```bash
railway login
```

#### Step 3: Deploy
```bash
# Run the Railway deployment script
./scripts/deploy-beta-railway.sh
```

#### Step 4: Get Your Public URL
The script will generate a `SHAREABLE_LINKS_RAILWAY.md` file with your public URL.

### Option 2: Custom Cloud Deployment

For more control over your deployment environment.

#### Step 1: Set Environment Variables
```bash
export BETA_DOMAIN="your-domain.com"
export BETA_API_KEY="your-api-key"
export BETA_JWT_SECRET="your-jwt-secret"
```

#### Step 2: Deploy
```bash
# Run the custom deployment script
./scripts/deploy-beta-web-interface.sh
```

#### Step 3: Configure DNS
Point your domain to your server's IP address.

## 🌐 Public Access URLs

After deployment, you'll have access to:

- **Web Interface**: `https://your-deployment-url.com`
- **API Documentation**: `https://your-deployment-url.com/docs`
- **Health Check**: `https://your-deployment-url.com/health`

## 📧 Shareable Links for Beta Testers

### For Non-Technical Users:
```
🌐 KYB Platform Beta Testing

Access the platform here: https://your-deployment-url.com

Instructions:
1. Click "Register" to create an account
2. Choose your role (Compliance Officer, Risk Manager, etc.)
3. Use the web interface to test business classification
4. Explore risk assessment and compliance features
5. Provide feedback through the built-in feedback system

Test Credentials:
- Email: test@example.com
- Password: password123

This is a beta environment - data may be reset periodically.
```

### For Technical Users:
```
🔌 KYB Platform API Beta Testing

API Base URL: https://your-deployment-url.com/api/v1
Documentation: https://your-deployment-url.com/docs

Test the API endpoints:
- POST /api/v1/auth/login
- POST /api/v1/classify
- POST /api/v1/risk/assess
- POST /api/v1/compliance/check

API Key: your-api-key-here
Rate Limit: 1000 requests per minute
```

## 🔧 Management Commands

### Railway Management
```bash
# View deployment status
railway status

# View logs
railway logs

# Update environment variables
railway variables set KEY=value

# Scale resources
railway scale
```

### Docker Management
```bash
# View running containers
docker-compose -f docker-compose.beta.yml ps

# View logs
docker-compose -f docker-compose.beta.yml logs

# Restart services
docker-compose -f docker-compose.beta.yml restart

# Stop all services
docker-compose -f docker-compose.beta.yml down
```

## 📊 Monitoring

### Health Checks
- **Web Interface**: `https://your-deployment-url.com/health`
- **API**: `https://your-deployment-url.com/api/v1/health`

### Metrics Dashboard
- **Grafana**: `https://your-deployment-url.com:3000`
- **Prometheus**: `https://your-deployment-url.com:9090`

## 🔒 Security Considerations

### SSL Certificates
- Railway automatically provides SSL certificates
- For custom deployments, use Let's Encrypt or your own certificates

### Environment Variables
- Store sensitive data in environment variables
- Never commit API keys or secrets to version control
- Use Railway's environment variable management

### Rate Limiting
- API requests are limited to 1000 per minute
- Web interface has built-in protection against abuse

## 🐛 Troubleshooting

### Common Issues

#### Deployment Fails
```bash
# Check logs
railway logs

# Verify environment variables
railway variables

# Restart deployment
railway up
```

#### API Not Responding
```bash
# Check health endpoint
curl https://your-deployment-url.com/health

# Check API health
curl https://your-deployment-url.com/api/v1/health
```

#### Web Interface Issues
- Clear browser cache
- Check browser console for errors
- Verify API base URL in web interface

### Support
- Check Railway dashboard for deployment issues
- Review application logs for errors
- Test API endpoints directly with curl or Postman

## 📈 Scaling

### Railway Scaling
```bash
# Scale up resources
railway scale

# Add more instances
railway scale --instances 3
```

### Custom Deployment Scaling
```bash
# Scale web interface
docker-compose -f docker-compose.beta.yml up -d --scale web-interface=3

# Scale API gateway
docker-compose -f docker-compose.beta.yml up -d --scale api-gateway=2
```

## 🎉 Success!

Once deployed, you'll have:
- ✅ Public web interface accessible to all beta testers
- ✅ API endpoints for technical users
- ✅ Monitoring and health checks
- ✅ Feedback collection system
- ✅ Secure HTTPS access

Share the generated links with your beta testers and start collecting feedback!
