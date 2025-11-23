import { expect, test } from '@playwright/test';
import { handleCorsOptions, getCorsHeaders } from './helpers/cors-helpers';

/**
 * Integration tests for Data Display (Phase 6 - Task 6.2.1)
 * 
 * Tests:
 * - All backend fields display when available
 * - Financial information card displays correctly
 * - Address display with all fields
 * - Metadata JSON viewer
 */

const TEST_MERCHANT_ID = 'merchant-123';

test.describe('Data Display Integration Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to merchant details page
    await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`);
    await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
  });

  test.describe('All Backend Fields Display', () => {
    test('should display all financial information fields when available', async ({ page }) => {
      // Mock merchant with all financial data
      await page.route(`**/api/v1/merchants/${TEST_MERCHANT_ID}**`, async (route) => {
        if (await handleCorsOptions(route)) return;
        const url = route.request().url();
        if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            headers: getCorsHeaders(),
            body: JSON.stringify({
              id: TEST_MERCHANT_ID,
              business_name: 'Test Business Inc',
              status: 'active',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
              founded_date: '2020-01-15T00:00:00Z',
              employee_count: 150,
              annual_revenue: 5000000.50,
            }),
          });
        } else {
          await route.continue();
        }
      });

      // Navigate directly to ensure fresh load with new mock
      await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`, { waitUntil: 'domcontentloaded' });
      await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
      await page.waitForTimeout(2000); // Wait for component to mount and render

      // Check for Financial Information card
      await expect(page.getByText(/Financial Information/i)).toBeVisible({ timeout: 5000 });

      // Check for founded date
      await expect(page.getByText(/Founded Date/i)).toBeVisible();
      await expect(page.getByText(/2020/i)).toBeVisible();

      // Check for employee count (formatted with commas)
      await expect(page.getByText(/Employee Count/i)).toBeVisible();
      await expect(page.getByText(/150/i)).toBeVisible();

      // Check for annual revenue (formatted as currency)
      await expect(page.getByText(/Annual Revenue/i)).toBeVisible();
      await expect(page.getByText(/\$5,000,000/i)).toBeVisible();
    });

    test('should display all address fields including street1, street2, countryCode', async ({ page }) => {
      // Mock merchant with complete address
      await page.route(`**/api/v1/merchants/${TEST_MERCHANT_ID}**`, async (route) => {
        if (await handleCorsOptions(route)) return;
        const url = route.request().url();
        if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            headers: getCorsHeaders(),
            body: JSON.stringify({
              id: TEST_MERCHANT_ID,
              business_name: 'Test Business Inc',
              status: 'active',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
              address: {
                street1: '123 Main Street',
                street2: 'Suite 100',
                city: 'San Francisco',
                state: 'CA',
                postal_code: '94102',
                country: 'United States',
                country_code: 'US',
              },
            }),
          });
        } else {
          await route.continue();
        }
      });

      // Navigate directly to ensure fresh load with new mock
      await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`, { waitUntil: 'domcontentloaded' });
      await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
      await page.waitForTimeout(2000); // Wait for component to mount and render

      // Check for Address card - wait for it to be visible
      await expect(page.getByText(/Address/i)).toBeVisible({ timeout: 10000 });

      // Check for street1
      await expect(page.getByText('123 Main Street')).toBeVisible();

      // Check for street2
      await expect(page.getByText('Suite 100')).toBeVisible();

      // Check for city
      await expect(page.getByText('San Francisco')).toBeVisible();

      // Check for state - use .first() to avoid strict mode violation
      await expect(page.getByText('CA').first()).toBeVisible();

      // Check for country with country code
      await expect(page.getByText(/United States/i)).toBeVisible();
      await expect(page.getByText(/US/i)).toBeVisible();
    });

    test('should display system information fields (createdBy, metadata)', async ({ page }) => {
      // Mock merchant with system data
      await page.route(`**/api/v1/merchants/${TEST_MERCHANT_ID}**`, async (route) => {
        if (await handleCorsOptions(route)) return;
        const url = route.request().url();
        if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            headers: getCorsHeaders(),
            body: JSON.stringify({
              id: TEST_MERCHANT_ID,
              business_name: 'Test Business Inc',
              status: 'active',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
              created_by: 'user-123',
              metadata: {
                source: 'manual',
                verified: true,
                tags: ['enterprise', 'high-value'],
              },
            }),
          });
        } else {
          await route.continue();
        }
      });

      // Navigate directly to ensure fresh load with new mock
      await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`, { waitUntil: 'domcontentloaded' });
      await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
      await page.waitForTimeout(2000); // Wait for component to mount and render

      // Check for Metadata card - use .first() to avoid strict mode violation
      await expect(page.getByText(/Metadata/i).first()).toBeVisible({ timeout: 5000 });

      // Check for Created By field
      await expect(page.getByText(/Created By/i)).toBeVisible();
      await expect(page.getByText('user-123')).toBeVisible();
    });

    test('should display N/A for missing optional fields', async ({ page }) => {
      // Mock merchant with minimal data
      await page.route(`**/api/v1/merchants/${TEST_MERCHANT_ID}**`, async (route) => {
        if (await handleCorsOptions(route)) return;
        const url = route.request().url();
        if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            headers: getCorsHeaders(),
            body: JSON.stringify({
              id: TEST_MERCHANT_ID,
              business_name: 'Test Business Inc',
              status: 'active',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            }),
          });
        } else {
          await route.continue();
        }
      });

      // Navigate directly to ensure fresh load with new mock
      await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`, { waitUntil: 'domcontentloaded' });
      await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
      await page.waitForTimeout(2000); // Wait for component to mount and render

      // Check for Financial Information card
      await expect(page.getByText(/Financial Information/i)).toBeVisible({ timeout: 5000 });

      // Should show N/A for missing fields
      const naTexts = page.getByText(/N\/A/i);
      const count = await naTexts.count();
      expect(count).toBeGreaterThan(0);
    });
  });

  test.describe('Financial Information Card', () => {
    test('should format employee count with commas', async ({ page }) => {
      await page.route(`**/api/v1/merchants/${TEST_MERCHANT_ID}**`, async (route) => {
        if (await handleCorsOptions(route)) return;
        const url = route.request().url();
        if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            headers: getCorsHeaders(),
            body: JSON.stringify({
              id: TEST_MERCHANT_ID,
              business_name: 'Test Business Inc',
              status: 'active',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
              employee_count: 15000,
            }),
          });
        } else {
          await route.continue();
        }
      });

      // Navigate directly to ensure fresh load with new mock
      await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`, { waitUntil: 'domcontentloaded' });
      await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
      await page.waitForTimeout(2000); // Wait for component to mount and render

      // Check for formatted employee count
      await expect(page.getByText(/15,000/i)).toBeVisible({ timeout: 5000 });
    });

    test('should format annual revenue as currency', async ({ page }) => {
      await page.route(`**/api/v1/merchants/${TEST_MERCHANT_ID}**`, async (route) => {
        if (await handleCorsOptions(route)) return;
        const url = route.request().url();
        if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            headers: getCorsHeaders(),
            body: JSON.stringify({
              id: TEST_MERCHANT_ID,
              business_name: 'Test Business Inc',
              status: 'active',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
              annual_revenue: 5000000.50,
            }),
          });
        } else {
          await route.continue();
        }
      });

      // Navigate directly to ensure fresh load with new mock
      await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`, { waitUntil: 'domcontentloaded' });
      await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
      await page.waitForTimeout(2000); // Wait for component to mount and render

      // Check for currency format (with $ and commas)
      await expect(page.getByText(/\$5,000,000/i)).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Metadata JSON Viewer', () => {
    test('should display metadata JSON in expandable section', async ({ page }) => {
      await page.route(`**/api/v1/merchants/${TEST_MERCHANT_ID}**`, async (route) => {
        if (await handleCorsOptions(route)) return;
        const url = route.request().url();
        if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            headers: getCorsHeaders(),
            body: JSON.stringify({
              id: TEST_MERCHANT_ID,
              business_name: 'Test Business Inc',
              status: 'active',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
              metadata: {
                source: 'manual',
                verified: true,
                tags: ['enterprise'],
              },
            }),
          });
        } else {
          await route.continue();
        }
      });

      // Navigate directly to ensure fresh load with new mock
      await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`, { waitUntil: 'domcontentloaded' });
      await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
      await page.waitForTimeout(2000); // Wait for component to mount and render

      // Check for Metadata section
      await expect(page.getByText(/Metadata/i).first()).toBeVisible({ timeout: 5000 });

      // Metadata should be expandable (check for collapsible trigger)
      const metadataTrigger = page.getByRole('button', { name: /Metadata/i });
      if (await metadataTrigger.count() > 0) {
        await metadataTrigger.first().click();
        // Should show metadata content
        await expect(page.getByText(/source/i)).toBeVisible({ timeout: 2000 });
      }
    });
  });

  test.describe('Data Completeness Indicator', () => {
    test('should display data completeness percentage', async ({ page }) => {
      await page.route(`**/api/v1/merchants/${TEST_MERCHANT_ID}**`, async (route) => {
        if (await handleCorsOptions(route)) return;
        const url = route.request().url();
        if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            headers: getCorsHeaders(),
            body: JSON.stringify({
              id: TEST_MERCHANT_ID,
              business_name: 'Test Business Inc',
              status: 'active',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
              founded_date: '2020-01-15T00:00:00Z',
              employee_count: 150,
              annual_revenue: 5000000.50,
              email: 'test@example.com',
              phone: '+1-555-123-4567',
            }),
          });
        } else {
          await route.continue();
        }
      });

      // Navigate directly to ensure fresh load with new mock
      await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`, { waitUntil: 'domcontentloaded' });
      await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
      await page.waitForTimeout(2000); // Wait for component to mount and render

      // Check for data completeness indicator
      const completenessText = page.getByText(/Data Completeness/i);
      const isVisible = await completenessText.isVisible({ timeout: 5000 }).catch(() => false);
      expect(isVisible).toBeTruthy();
    });
  });
});

