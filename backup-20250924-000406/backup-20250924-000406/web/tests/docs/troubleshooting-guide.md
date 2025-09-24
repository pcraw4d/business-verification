# Visual Regression Testing Troubleshooting Guide

## Quick Reference

| Issue | Quick Fix | Detailed Solution |
|-------|-----------|-------------------|
| Test timeout | Increase timeout | [Timeout Issues](#timeout-issues) |
| Element not found | Check selector | [Element Not Found](#element-not-found) |
| Screenshot mismatch | Review changes | [Screenshot Mismatches](#screenshot-mismatches) |
| Railway connection failed | Check deployment | [Connection Issues](#connection-issues) |
| Baseline update failed | Check permissions | [Baseline Management](#baseline-management) |

## Common Issues and Solutions

### Timeout Issues

#### Problem
Tests fail with timeout errors like:
```
Error: Test timeout of 30000ms exceeded.
```

#### Causes
- Slow page loading
- Network connectivity issues
- Railway deployment not responding
- Heavy page content

#### Solutions

**1. Increase Test Timeout**
```javascript
test('my test', async ({ page }) => {
  await page.goto('/my-page', { timeout: 60000 }); // 60 seconds
});
```

**2. Increase Global Timeout**
```javascript
// In playwright.config.js
export default defineConfig({
  timeout: 60000, // 60 seconds
  expect: {
    timeout: 10000, // 10 seconds
  },
});
```

**3. Wait for Specific Conditions**
```javascript
test('my test', async ({ page }) => {
  await page.goto('/my-page');
  await page.waitForLoadState('networkidle'); // Wait for network to be idle
  await page.waitForSelector('.main-content'); // Wait for specific element
});
```

**4. Check Railway Deployment**
```bash
# Test Railway connectivity
curl -I https://shimmering-comfort-production.up.railway.app/risk-dashboard.html
```

### Element Not Found

#### Problem
Tests fail because elements cannot be found:
```
Error: locator.click: Timeout 10000ms exceeded.
```

#### Causes
- Incorrect selectors
- Elements not yet loaded
- Dynamic content not rendered
- Page structure changes

#### Solutions

**1. Use Better Selectors**
```javascript
// Bad - brittle CSS selector
await page.click('.container > div:nth-child(2) > button');

// Good - semantic selector
await page.click('[data-testid="submit-button"]');

// Good - role-based selector
await page.click('button[type="submit"]');
```

**2. Wait for Elements**
```javascript
test('my test', async ({ page }) => {
  await page.goto('/my-page');
  
  // Wait for element to be visible
  await page.waitForSelector('.my-element', { state: 'visible' });
  
  // Wait for element to be stable
  await page.waitForSelector('.my-element', { state: 'attached' });
  
  // Then interact with it
  await page.click('.my-element');
});
```

**3. Check Element State**
```javascript
test('my test', async ({ page }) => {
  const element = page.locator('.my-element');
  
  // Check if element exists
  if (await element.count() > 0) {
    await element.click();
  } else {
    console.log('Element not found');
  }
});
```

**4. Debug Element Selection**
```javascript
test('my test', async ({ page }) => {
  await page.goto('/my-page');
  
  // Debug: log all buttons on the page
  const buttons = await page.locator('button').all();
  console.log(`Found ${buttons.length} buttons`);
  
  // Debug: take screenshot
  await page.screenshot({ path: 'debug-page.png' });
});
```

### Screenshot Mismatches

#### Problem
Visual regression tests fail because screenshots don't match baselines:
```
Error: Screenshot comparison failed:
Expected: baseline.png
Received: actual.png
```

#### Causes
- Intentional UI changes
- Timing issues
- Dynamic content
- Browser differences
- Font rendering differences

#### Solutions

**1. Review the Changes**
```bash
# View the diff
npx playwright show-report
```

**2. Update Baselines (if changes are intentional)**
```bash
# Update all baselines
npx playwright test --update-snapshots

# Update specific test baselines
npx playwright test tests/visual/baseline-screenshots.spec.js --update-snapshots
```

**3. Fix Timing Issues**
```javascript
test('my test', async ({ page }) => {
  await page.goto('/my-page');
  
  // Wait for animations to complete
  await page.waitForTimeout(1000);
  
  // Wait for specific elements to be stable
  await page.waitForSelector('.chart', { state: 'visible' });
  await page.waitForLoadState('networkidle');
  
  // Then take screenshot
  await expect(page).toHaveScreenshot('my-page.png');
});
```

**4. Handle Dynamic Content**
```javascript
test('my test', async ({ page }) => {
  await page.goto('/my-page');
  
  // Mock dynamic content
  await page.addInitScript(() => {
    // Override Date.now() to return fixed timestamp
    Date.now = () => 1640995200000; // 2022-01-01
  });
  
  await expect(page).toHaveScreenshot('my-page.png');
});
```

### Connection Issues

#### Problem
Tests fail to connect to Railway deployment:
```
Error: net::ERR_CONNECTION_REFUSED at https://shimmering-comfort-production.up.railway.app
```

#### Causes
- Railway deployment down
- Network connectivity issues
- Incorrect URL
- Firewall blocking

#### Solutions

**1. Check Railway Deployment Status**
```bash
# Test basic connectivity
curl -I https://shimmering-comfort-production.up.railway.app

# Test specific page
curl -I https://shimmering-comfort-production.up.railway.app/risk-dashboard.html
```

**2. Verify URL Configuration**
```javascript
// In test files, ensure correct baseURL
const baseUrl = process.env.BASE_URL || 'https://shimmering-comfort-production.up.railway.app';
```

**3. Add Retry Logic**
```javascript
test('my test', async ({ page }) => {
  // Retry connection
  let retries = 3;
  while (retries > 0) {
    try {
      await page.goto('/my-page', { timeout: 30000 });
      break;
    } catch (error) {
      retries--;
      if (retries === 0) throw error;
      await page.waitForTimeout(5000); // Wait 5 seconds before retry
    }
  }
});
```

**4. Use Local Development Server**
```javascript
// For local development, use local server
const baseUrl = process.env.BASE_URL || 'http://localhost:8080';
```

### Baseline Management

#### Problem
Baseline updates fail or baselines are not committed properly.

#### Causes
- Git permissions issues
- File path problems
- CI/CD configuration issues

#### Solutions

**1. Check Git Configuration**
```bash
# Ensure git is configured
git config --global user.email "your-email@example.com"
git config --global user.name "Your Name"
```

**2. Verify File Paths**
```bash
# Check if baseline files exist
ls -la web/tests/visual/*.spec.js-snapshots/
```

**3. Manual Baseline Update**
```bash
# Update baselines locally
npx playwright test --update-snapshots

# Commit changes
git add web/tests/visual/*.spec.js-snapshots/
git commit -m "chore: update visual regression test baselines"
git push
```

**4. Use GitHub Actions Workflow**
- Go to GitHub Actions
- Run "Update Visual Test Baselines" workflow
- Select the test type to update
- Review the results

### Performance Issues

#### Problem
Tests run slowly or consume too many resources.

#### Causes
- Too many parallel tests
- Large screenshots
- Slow network
- Resource-intensive operations

#### Solutions

**1. Optimize Test Execution**
```javascript
// In playwright.config.js
export default defineConfig({
  workers: 2, // Reduce parallel workers
  timeout: 30000, // Set appropriate timeout
});
```

**2. Reduce Screenshot Size**
```javascript
test('my test', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 720 }); // Smaller viewport
  await expect(page).toHaveScreenshot('my-page.png', {
    maxDiffPixels: 100, // Allow small differences
  });
});
```

**3. Use Selective Testing**
```bash
# Run only specific tests
npx playwright test --grep="critical tests"
```

### Browser-Specific Issues

#### Problem
Tests pass in one browser but fail in another.

#### Causes
- Browser rendering differences
- Font rendering variations
- CSS compatibility issues
- JavaScript execution differences

#### Solutions

**1. Browser-Specific Baselines**
```javascript
// Create browser-specific baselines
test('my test', async ({ page, browserName }) => {
  await expect(page).toHaveScreenshot(`my-page-${browserName}.png`);
});
```

**2. Adjust Browser Settings**
```javascript
// In playwright.config.js
export default defineConfig({
  use: {
    // Disable animations for consistent screenshots
    reducedMotion: 'reduce',
    // Set consistent font rendering
    fontFamily: 'Arial, sans-serif',
  },
});
```

**3. Handle Browser Differences**
```javascript
test('my test', async ({ page, browserName }) => {
  // Adjust expectations based on browser
  const maxDiffPixels = browserName === 'webkit' ? 200 : 100;
  
  await expect(page).toHaveScreenshot('my-page.png', {
    maxDiffPixels,
  });
});
```

## Debugging Techniques

### 1. Enable Debug Mode
```bash
npx playwright test --debug
```

### 2. Take Screenshots
```javascript
test('my test', async ({ page }) => {
  await page.goto('/my-page');
  await page.screenshot({ path: 'debug-screenshot.png' });
});
```

### 3. Log Page Content
```javascript
test('my test', async ({ page }) => {
  await page.goto('/my-page');
  
  // Log page title
  console.log('Page title:', await page.title());
  
  // Log all buttons
  const buttons = await page.locator('button').all();
  console.log(`Found ${buttons.length} buttons`);
  
  // Log page HTML
  const html = await page.content();
  console.log('Page HTML length:', html.length);
});
```

### 4. Use Playwright Inspector
```bash
npx playwright test --debug
```

### 5. Check Network Requests
```javascript
test('my test', async ({ page }) => {
  // Listen to network requests
  page.on('request', request => {
    console.log('Request:', request.url());
  });
  
  page.on('response', response => {
    console.log('Response:', response.url(), response.status());
  });
  
  await page.goto('/my-page');
});
```

## Getting Help

### 1. Check Test Reports
```bash
npx playwright show-report
```

### 2. Review GitHub Actions Logs
- Go to GitHub Actions
- Click on the failed workflow
- Review the logs for specific errors

### 3. Check Railway Deployment
- Verify the deployment is running
- Check Railway logs for errors
- Test the application manually

### 4. Create an Issue
If you can't resolve the issue:
1. Gather relevant information (error messages, logs, screenshots)
2. Create a detailed issue description
3. Include steps to reproduce
4. Attach relevant files

### 5. Community Resources
- [Playwright Documentation](https://playwright.dev/)
- [Playwright GitHub Issues](https://github.com/microsoft/playwright/issues)
- [Stack Overflow](https://stackoverflow.com/questions/tagged/playwright)

## Prevention

### 1. Regular Maintenance
- Update baselines regularly
- Review test stability
- Monitor test performance
- Keep dependencies updated

### 2. Good Test Practices
- Use stable selectors
- Add proper waits
- Handle dynamic content
- Test on multiple browsers

### 3. CI/CD Monitoring
- Monitor test success rates
- Set up alerts for failures
- Review test reports regularly
- Update workflows as needed

### 4. Documentation
- Keep this guide updated
- Document new issues and solutions
- Share knowledge with team
- Maintain test documentation
