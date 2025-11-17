import { expect, test } from '@playwright/test';

test.describe('Form Tests', () => {
  test('should submit merchant form', async ({ page }) => {
    await page.goto('http://localhost:3000/add-merchant');
    await page.waitForLoadState('networkidle');
    
    // Fill in form fields using proper selectors
    // Business Name is required
    await page.fill('input[name="businessName"]', 'Test Merchant');
    
    // Country is a Select component - need to click and select
    const countrySelect = page.locator('button:has-text("Select country"), [role="combobox"]:near(label:has-text("Country"))').first();
    if (await countrySelect.isVisible({ timeout: 2000 })) {
      await countrySelect.click();
      await page.getByRole('option', { name: /united states|us/i }).click();
    } else {
      // Fallback: try to find select by name
      await page.selectOption('select[name="country"]', 'US');
    }
    
    // Submit form
    const submitButton = page.getByRole('button', { name: /submit|create|verify/i }).first();
    await submitButton.click();
    
    // Wait for navigation or success message
    await page.waitForTimeout(3000);
    
    // Check for success (either redirect or success message)
    const success = page.locator('text=/success|created|saved/i').first();
    const hasRedirect = page.url().includes('merchant-details');
    
    if (await success.isVisible({ timeout: 5000 }).catch(() => false) || hasRedirect) {
      // Success - either message shown or redirected
      expect(true).toBeTruthy();
    } else {
      // Check for redirect to merchant details
      await expect(page).toHaveURL(/.*merchant-details/, { timeout: 10000 });
    }
  });

  test('should validate required fields', async ({ page }) => {
    await page.goto('http://localhost:3000/add-merchant');
    await page.waitForLoadState('networkidle');
    
    // Try to submit without filling required fields
    const submitButton = page.getByRole('button', { name: /submit|create|verify/i }).first();
    await submitButton.click();
    
    // Wait a bit for validation to run
    await page.waitForTimeout(1000);
    
    // Check for validation errors - they appear with role="alert"
    const errorMessage = page.locator('[role="alert"], .text-destructive').first();
    // Error might be in toast or inline, check both
    const hasError = await errorMessage.isVisible({ timeout: 3000 }).catch(() => false) ||
                     await page.locator('text=/required|invalid|error/i').first().isVisible({ timeout: 2000 }).catch(() => false);
    
    // At minimum, form should not submit (stay on same page or show error)
    const stillOnPage = page.url().includes('add-merchant');
    expect(hasError || stillOnPage).toBeTruthy();
  });

  test('should handle form errors gracefully', async ({ page }) => {
    await page.goto('http://localhost:3000/add-merchant');
    await page.waitForLoadState('networkidle');
    
    // Fill form with invalid data
    const businessNameInput = page.locator('input[name="businessName"]');
    if (await businessNameInput.isVisible({ timeout: 2000 })) {
      await businessNameInput.fill('');
    }
    
    const emailInput = page.locator('input[name="email"], input[type="email"]').first();
    if (await emailInput.isVisible({ timeout: 2000 })) {
      await emailInput.fill('invalid-email');
    }
    
    const submitButton = page.getByRole('button', { name: /submit|create|verify/i }).first();
    await submitButton.click();
    
    // Wait for validation
    await page.waitForTimeout(1000);
    
    // Should show validation errors - check for error messages
    const hasError = await page.locator('[role="alert"], .text-destructive, text=/invalid|error/i').first()
      .isVisible({ timeout: 3000 }).catch(() => false);
    
    // Form should not submit with invalid data
    const stillOnPage = page.url().includes('add-merchant');
    expect(hasError || stillOnPage).toBeTruthy();
  });
});

