# Beta Testing Verification Guide

## ✅ Production Deployment Status

All changes are automatically deployed to **production Railway services** when pushed to the `main` branch on GitHub.

### Production URLs

1. **Frontend Service (UI)**
   - URL: `https://frontend-service-production-b225.up.railway.app`
   - This is where users interact with the merchant form
   - **All HTML/JavaScript changes are live here**

2. **API Gateway Service**
   - URL: `https://api-gateway-service-production-21fd.up.railway.app`
   - Handles all API requests from the frontend
   - **All backend fixes are live here**

3. **Classification Service**
   - URL: `https://classification-service-production.up.railway.app`
   - Processes business classification requests

## 🔄 Automatic Deployment

Railway is configured to automatically deploy when:
- Code is pushed to the `main` branch on GitHub
- Changes are detected in the service directories

### Recent Deployments

The following fixes have been deployed to production:

1. ✅ **API Gateway Request Body Bug Fix** (Commit: `c3e8a7b20`)
   - Fixed request body being consumed twice
   - Now properly forwards requests to classification service

2. ✅ **Promise.allSettled() Bug Fix** (Commit: `32de0aedf`)
   - Fixed error handling in merchant form
   - Now properly tracks API call failures

3. ✅ **Enhanced Error Logging** (Commit: `5b6da37bc`)
   - Added comprehensive error logging
   - Better debugging information in browser console

4. ✅ **XSS Vulnerability Fixes** (Commit: `9e2fa57c9`)
   - Fixed XSS vulnerabilities in merchant details page
   - All user input is now properly escaped

## 🧪 Beta Testing Instructions

### Test 1: Add Merchant Form Submission

1. **Navigate to the form:**
   ```
   https://frontend-service-production-b225.up.railway.app/add-merchant
   ```

2. **Fill out the form:**
   - Business Name: (any name)
   - Website URL: (optional)
   - Country: (select from dropdown)
   - Other required fields

3. **Submit the form:**
   - Click "Verify Merchant" button
   - **Expected:** Form should redirect to merchant details page

4. **Check browser console (F12):**
   - Look for logs starting with `🔍`
   - Should see API call logs
   - Check for any error messages

5. **Verify redirect:**
   - Should automatically navigate to:
     ```
     https://frontend-service-production-b225.up.railway.app/merchant-details
     ```

### Test 2: Verify API Calls Are Working

1. **Open browser console (F12 → Console tab)**

2. **Submit the merchant form**

3. **Check for these logs:**
   ```
   🔍 Making Business Intelligence API call
   🔍 API URL: https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify
   🔍 Request data: {...}
   🔍 Response status: 200
   🔍 Business Intelligence API result: {...}
   ```

4. **If API calls fail, you'll see:**
   ```
   🔍 Business Intelligence API call failed
   🔍 Error name: ...
   🔍 Error message: ...
   ```

### Test 3: Verify Error Handling

1. **Test with network issues:**
   - Disable network temporarily
   - Submit form
   - **Expected:** Should still redirect to merchant details page
   - **Expected:** Error should be logged in console

2. **Test with invalid data:**
   - Submit form with missing required fields
   - **Expected:** Form validation should prevent submission

### Test 4: Verify Data Persistence

1. **Submit merchant form**

2. **Check sessionStorage:**
   - Open browser console (F12)
   - Run: `JSON.parse(sessionStorage.getItem('merchantData'))`
   - **Expected:** Should see merchant form data

3. **Check API results:**
   - Run: `JSON.parse(sessionStorage.getItem('merchantApiResults'))`
   - **Expected:** Should see API results or error information

## 🔍 Verification Checklist

- [ ] Frontend service is accessible at production URL
- [ ] Add merchant form loads correctly
- [ ] Form validation works
- [ ] Form submission redirects to merchant details page
- [ ] API calls are being made (check browser console)
- [ ] Error handling works (form redirects even if APIs fail)
- [ ] Data is stored in sessionStorage
- [ ] Merchant details page loads with data

## 🐛 Troubleshooting

### If form doesn't redirect:

1. **Check browser console for errors**
2. **Verify API gateway is accessible:**
   ```
   curl https://api-gateway-service-production-21fd.up.railway.app/health
   ```
3. **Check Railway deployment logs:**
   - Go to Railway Dashboard
   - Select the service
   - Check "Deployments" tab for recent deployments
   - Check logs for any errors

### If API calls are failing:

1. **Check CORS errors in console**
2. **Verify API gateway health:**
   ```
   https://api-gateway-service-production-21fd.up.railway.app/health
   ```
3. **Check classification service health:**
   ```
   https://classification-service-production.up.railway.app/health
   ```

## 📊 Monitoring Deployment Status

### Check Railway Dashboard

1. Go to https://railway.app
2. Select your project
3. Check each service:
   - **Frontend Service** - Should show "Active" status
   - **API Gateway Service** - Should show "Active" status
   - **Classification Service** - Should show "Active" status

### Check Deployment Logs

1. In Railway Dashboard, select a service
2. Go to "Deployments" tab
3. Click on latest deployment
4. Check build logs for any errors
5. Check runtime logs for application errors

## 🚀 Quick Test Script

Run this in browser console on the add-merchant page:

```javascript
// Test API connectivity
fetch('https://api-gateway-service-production-21fd.up.railway.app/health')
  .then(r => r.json())
  .then(data => console.log('✅ API Gateway is accessible:', data))
  .catch(err => console.error('❌ API Gateway error:', err));

// Test classification endpoint
fetch('https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    business_name: 'Test Company',
    description: 'Test description',
    website_url: 'https://example.com'
  })
})
  .then(r => r.json())
  .then(data => console.log('✅ Classification API works:', data))
  .catch(err => console.error('❌ Classification API error:', err));
```

## 📝 Notes

- All changes pushed to `main` branch are automatically deployed
- Deployment typically takes 2-5 minutes after push
- Check Railway dashboard to confirm deployment completion
- Production URLs are hardcoded in the frontend code
- All fixes are live and testable immediately after deployment

