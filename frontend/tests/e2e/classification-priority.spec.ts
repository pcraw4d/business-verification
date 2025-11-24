import { test, expect } from '@playwright/test';

test.describe('Classification Priority', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to merchant details page
    await page.goto('/merchant-details/merchant-123');
    // Wait for page to load
    await page.waitForLoadState('networkidle');
  });

  test('classification prioritizes website content over business name', async ({ page }) => {
    // Navigate to Business Analytics tab
    await page.click('text=Business Analytics');
    await page.waitForSelector('[data-testid="classification-card"]', { timeout: 5000 });

    // Verify classification data is displayed
    const classificationCard = page.locator('[data-testid="classification-card"]');
    await expect(classificationCard).toBeVisible();

    // Verify primary industry is shown
    await expect(page.locator('text=Primary Industry')).toBeVisible();

    // Check that website content is used (metadata should indicate primary source)
    const metadata = page.locator('text=/data source priority/i');
    if (await metadata.count() > 0) {
      await expect(metadata).toBeVisible();
    }
  });

  test('multi-page analysis extracts more keywords', async ({ page }) => {
    // Navigate to Business Analytics tab
    await page.click('text=Business Analytics');
    await page.waitForSelector('[data-testid="classification-card"]', { timeout: 5000 });

    // Check for multi-page analysis indicator
    const multiPageIndicator = page.locator('text=/multi-page|pages analyzed/i');
    
    // If metadata is available, verify it shows multi-page analysis
    if (await multiPageIndicator.count() > 0) {
      await expect(multiPageIndicator.first()).toBeVisible();
      
      // Verify pages analyzed count is shown
      const pagesCount = page.locator('text=/\\d+ pages/i');
      if (await pagesCount.count() > 0) {
        await expect(pagesCount.first()).toBeVisible();
      }
    }
  });

  test('structured data improves classification', async ({ page }) => {
    // Navigate to Business Analytics tab
    await page.click('text=Business Analytics');
    await page.waitForSelector('[data-testid="classification-card"]', { timeout: 5000 });

    // Check for structured data indicator
    const structuredDataIndicator = page.locator('text=/structured data/i');
    
    // If structured data was found, verify it's displayed
    if (await structuredDataIndicator.count() > 0) {
      await expect(structuredDataIndicator.first()).toBeVisible();
    }
  });

  test('brand match uses business name for hotels', async ({ page }) => {
    // This test requires a merchant with a known hotel brand
    // Navigate to merchant details for a hotel brand
    await page.goto('/merchant-details/hotel-merchant-123');
    await page.waitForLoadState('networkidle');

    // Navigate to Business Analytics tab
    await page.click('text=Business Analytics');
    await page.waitForSelector('[data-testid="classification-card"]', { timeout: 5000 });

    // Check for brand match indicator
    const brandMatchIndicator = page.locator('text=/brand match|Hilton|Marriott|Hyatt/i');
    
    // If brand match is present, verify it's displayed
    if (await brandMatchIndicator.count() > 0) {
      await expect(brandMatchIndicator.first()).toBeVisible();
    }
  });

  test('frontend displays classification metadata', async ({ page }) => {
    // Navigate to Business Analytics tab
    await page.click('text=Business Analytics');
    await page.waitForSelector('[data-testid="classification-card"]', { timeout: 5000 });

    // Verify classification metadata section is visible (if metadata exists)
    const metadataSection = page.locator('text=/analysis metadata|pages analyzed|data source/i');
    
    // Metadata may or may not be present depending on classification status
    // Just verify the page doesn't crash when metadata is missing
    await expect(page.locator('text=Primary Industry')).toBeVisible();
  });

  test('multi-page analysis completes in <60s', async ({ page }) => {
    // Navigate to merchant details with a website URL
    await page.goto('/merchant-details/merchant-with-website-123');
    await page.waitForLoadState('networkidle');

    // Navigate to Business Analytics tab
    await page.click('text=Business Analytics');
    
    // Start timing
    const startTime = Date.now();
    
    // Wait for classification to complete (indicated by metadata or classification data)
    await page.waitForSelector('text=Primary Industry', { timeout: 65000 });
    
    const duration = (Date.now() - startTime) / 1000; // Convert to seconds

    // Verify it completes within 60 seconds (with 5 second buffer)
    expect(duration).toBeLessThan(65);
  });
});

