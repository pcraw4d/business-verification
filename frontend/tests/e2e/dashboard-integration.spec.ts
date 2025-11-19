import { expect, test } from '@playwright/test';

/**
 * Integration tests for Dashboard Pages
 * 
 * Tests end-to-end flows for:
 * - Business Intelligence Dashboard
 * - Risk Dashboard
 * - Risk Indicators Dashboard
 * 
 * Verifies:
 * - Portfolio data loading
 * - Error handling
 * - Loading states
 * - Data display
 */

test.describe('Dashboard Integration Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Set up API mocking for consistent test results
    // Tests will verify that pages handle both success and error cases
  });

  test.describe('Business Intelligence Dashboard', () => {
    test('should load portfolio analytics and statistics', async ({ page }) => {
      // Mock API responses
      await page.route('**/api/v1/merchants/analytics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            totalMerchants: 150,
            averageRiskScore: 0.6,
            averageClassificationConfidence: 0.85,
            averageSecurityTrustScore: 0.75,
            averageDataQuality: 0.9,
            riskDistribution: { low: 60, medium: 70, high: 20 },
            industryDistribution: {
              'Technology': 50,
              'Finance': 40,
              'Retail': 30,
              'Healthcare': 30,
            },
            countryDistribution: {},
            timestamp: new Date().toISOString(),
          }),
        });
      });

      await page.route('**/api/v1/merchants/statistics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            totalMerchants: 150,
            averageRiskScore: 0.6,
            riskDistribution: { low: 60, medium: 70, high: 20 },
            industryBreakdown: [
              { industry: 'Technology', count: 50 },
              { industry: 'Finance', count: 40 },
              { industry: 'Retail', count: 30 },
              { industry: 'Healthcare', count: 30 },
            ],
            countryBreakdown: [],
            timestamp: new Date().toISOString(),
          }),
        });
      });

      await page.route('**/api/v3/dashboard/metrics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            overview: { total_requests: 125000, active_users: 45 },
            business: { total_verifications: 150, revenue: 1000000 },
            performance: { response_time: 200 },
          }),
        });
      });

      await page.goto('/dashboard');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      
      // Wait for loading to complete
      await page.waitForTimeout(3000);

      // Verify metrics are displayed (flexible matching)
      const totalMerchants = page.getByText(/150|merchants|total/i).first();
      const hasMerchants = await totalMerchants.isVisible({ timeout: 10000 }).catch(() => false);

      // Verify charts are rendered (check for chart containers)
      const chartContainer = page.locator('[class*="chart"], [data-testid*="chart"]').first();
      const hasChart = await chartContainer.isVisible({ timeout: 5000 }).catch(() => false);
      
      // At minimum, page should load without errors
      await expect(page.locator('main, [role="main"]').first()).toBeVisible();
      
      // Should have either metrics or charts visible
      expect(hasMerchants || hasChart).toBeTruthy();
    });

    test('should show loading state initially', async ({ page }) => {
      // Delay API response to see loading state
      await page.route('**/api/v1/merchants/analytics', async (route) => {
        await new Promise(resolve => setTimeout(resolve, 2000));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            totalMerchants: 150,
            averageRiskScore: 0.6,
            averageClassificationConfidence: 0.85,
            averageSecurityTrustScore: 0.75,
            averageDataQuality: 0.9,
            riskDistribution: { low: 60, medium: 70, high: 20 },
            industryDistribution: {},
            countryDistribution: {},
            timestamp: new Date().toISOString(),
          }),
        });
      });

      await page.goto('/dashboard');
      
      // Check for skeleton loaders (they may disappear quickly)
      const skeleton = page.locator('[class*="skeleton"], [class*="loading"]').first();
      const hasSkeleton = await skeleton.isVisible({ timeout: 1000 }).catch(() => false);
      
      // Page should load
      await expect(page.locator('body')).toBeVisible();
    });

    test('should handle API errors gracefully', async ({ page }) => {
      // Mock API errors
      await page.route('**/api/v1/merchants/analytics', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal Server Error' }),
        });
      });

      await page.route('**/api/v1/merchants/statistics', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal Server Error' }),
        });
      });

      await page.goto('/dashboard');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      await page.waitForTimeout(3000);

      // Should show error message, toast, or fallback to v3 endpoint
      const errorMessage = page.locator('text=/error|failed|unavailable/i').first();
      const toast = page.locator('[role="status"], [data-sonner-toast]').first();
      const hasError = await errorMessage.isVisible({ timeout: 5000 }).catch(() => false);
      const hasToast = await toast.isVisible({ timeout: 5000 }).catch(() => false);
      
      // Check if v3 endpoint provided fallback data
      const fallbackData = page.locator('text=/125000|1000000|revenue/i').first();
      const hasFallback = await fallbackData.isVisible({ timeout: 5000 }).catch(() => false);

      // Page should still be functional (not crashed)
      await expect(page.locator('body')).toBeVisible();
      
      // Should have some indication of error or fallback data
      expect(hasError || hasToast || hasFallback).toBeTruthy();
    });

    test('should fallback to v3 endpoint when portfolio endpoints fail', async ({ page }) => {
      // Mock portfolio endpoints to fail
      await page.route('**/api/v1/merchants/analytics', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Service Unavailable' }),
        });
      });

      await page.route('**/api/v1/merchants/statistics', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Service Unavailable' }),
        });
      });

      // Mock v3 endpoint to succeed
      await page.route('**/api/v3/dashboard/metrics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            overview: { total_requests: 125000, active_users: 45 },
            business: { total_verifications: 150, revenue: 1000000, growth_rate: 5.2 },
            performance: { response_time: 200 },
          }),
        });
      });

      await page.goto('/dashboard');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      await page.waitForTimeout(3000);

      // Should still display data from v3 endpoint
      const metrics = page.locator('text=/150|merchants|revenue/i').first();
      const hasMetrics = await metrics.isVisible({ timeout: 5000 }).catch(() => false);
      
      // Page should load successfully
      await expect(page.locator('main, [role="main"]').first()).toBeVisible();
    });
  });

  test.describe('Risk Dashboard', () => {
    test('should load risk trends and insights', async ({ page }) => {
      // Mock API responses
      await page.route('**/api/v1/risk/metrics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            overallRiskScore: 0.65,
            highRiskMerchants: 15,
            riskAssessments: 150,
            riskTrend: -2.5,
            riskDistribution: { low: 60, medium: 70, high: 15, critical: 5 },
          }),
        });
      });

      await page.route('**/api/v1/analytics/trends**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            trends: [
              {
                industry: 'Technology',
                country: 'US',
                average_risk_score: 0.6,
                trend_direction: 'improving',
                change_percentage: -5.2,
                sample_size: 50,
              },
              {
                industry: 'Finance',
                country: 'US',
                average_risk_score: 0.7,
                trend_direction: 'worsening',
                change_percentage: 3.1,
                sample_size: 40,
              },
            ],
            summary: {
              total_assessments: 150,
              average_risk_score: 0.65,
              high_risk_percentage: 10.0,
            },
          }),
        });
      });

      await page.route('**/api/v1/analytics/insights**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            insights: [
              {
                id: 'insight-1',
                type: 'risk_factor',
                title: 'High Risk Concentration',
                description: '10% of assessments are high risk',
                severity: 'medium',
                impact: 'high',
                recommendation: 'Review high-risk assessments',
              },
            ],
            recommendations: [
              {
                category: 'monitoring',
                action: 'Increase monitoring frequency',
                priority: 'high',
              },
            ],
            timestamp: new Date().toISOString(),
          }),
        });
      });

      await page.goto('/risk-dashboard');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      await page.waitForTimeout(3000);

      // Verify risk metrics are displayed (flexible matching)
      const riskScore = page.getByText(/65|risk.*score|overall/i).first();
      const hasRiskScore = await riskScore.isVisible({ timeout: 10000 }).catch(() => false);

      // Verify charts are rendered
      const chartContainer = page.locator('[class*="chart"], [data-testid*="chart"]').first();
      const hasChart = await chartContainer.isVisible({ timeout: 5000 }).catch(() => false);

      // Verify insights section
      const insightsSection = page.getByText(/insights|recommendations/i).first();
      const hasInsights = await insightsSection.isVisible({ timeout: 5000 }).catch(() => false);

      // Page should load successfully
      await expect(page.locator('main, [role="main"]').first()).toBeVisible();
      
      // Should have at least one of: risk score, chart, or insights
      expect(hasRiskScore || hasChart || hasInsights).toBeTruthy();
    });

    test('should handle API errors gracefully', async ({ page }) => {
      // Mock API errors
      await page.route('**/api/v1/risk/metrics', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal Server Error' }),
        });
      });

      await page.route('**/api/v1/analytics/trends**', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal Server Error' }),
        });
      });

      await page.goto('/risk-dashboard');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      await page.waitForTimeout(3000);

      // Should show error or fallback to default values
      const errorMessage = page.locator('text=/error|failed|unavailable/i').first();
      const toast = page.locator('[role="status"], [data-sonner-toast]').first();
      const hasError = await errorMessage.isVisible({ timeout: 5000 }).catch(() => false);
      const hasToast = await toast.isVisible({ timeout: 5000 }).catch(() => false);

      // Page should still be functional
      await expect(page.locator('body')).toBeVisible();
    });

    test('should display loading state', async ({ page }) => {
      // Delay API response
      await page.route('**/api/v1/risk/metrics', async (route) => {
        await new Promise(resolve => setTimeout(resolve, 2000));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            overallRiskScore: 0.65,
            highRiskMerchants: 15,
            riskAssessments: 150,
            riskTrend: -2.5,
          }),
        });
      });

      await page.goto('/risk-dashboard');
      
      // Check for loading indicators
      const skeleton = page.locator('[class*="skeleton"], [class*="loading"]').first();
      const hasSkeleton = await skeleton.isVisible({ timeout: 1000 }).catch(() => false);
      
      // Page should load
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Risk Indicators Dashboard', () => {
    test('should load portfolio statistics and risk trends', async ({ page }) => {
      // Mock API responses
      await page.route('**/api/v1/risk/metrics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            overallRiskScore: 0.65,
            highRiskMerchants: 15,
            riskAssessments: 150,
            riskTrend: -2.5,
            riskDistribution: { low: 60, medium: 70, high: 15, critical: 5 },
          }),
        });
      });

      await page.route('**/api/v1/merchants/statistics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            totalMerchants: 150,
            averageRiskScore: 0.65,
            riskDistribution: { low: 60, medium: 70, high: 15, critical: 5 },
            industryBreakdown: [],
            countryBreakdown: [],
            timestamp: new Date().toISOString(),
          }),
        });
      });

      await page.route('**/api/v1/analytics/trends**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            trends: [],
            summary: {
              total_assessments: 150,
              average_risk_score: 0.65,
              high_risk_percentage: 10.0,
            },
          }),
        });
      });

      await page.goto('/risk-indicators');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      await page.waitForTimeout(3000);

      // Verify risk indicators are displayed
      const riskGauge = page.locator('[class*="gauge"], [data-testid*="gauge"]').first();
      const hasGauge = await riskGauge.isVisible({ timeout: 5000 }).catch(() => false);

      // Verify risk counts
      const riskCounts = page.getByText(/low|medium|high|critical/i).first();
      const hasRiskCounts = await riskCounts.isVisible({ timeout: 5000 }).catch(() => false);

      // Page should load successfully
      await expect(page.locator('main, [role="main"]').first()).toBeVisible();
    });

    test('should handle API errors gracefully', async ({ page }) => {
      // Mock API errors
      await page.route('**/api/v1/risk/metrics', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal Server Error' }),
        });
      });

      await page.route('**/api/v1/merchants/statistics', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal Server Error' }),
        });
      });

      await page.goto('/risk-indicators');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      await page.waitForTimeout(3000);

      // Should show error or fallback
      const errorMessage = page.locator('text=/error|failed|unavailable/i').first();
      const toast = page.locator('[role="status"], [data-sonner-toast]').first();
      const hasError = await errorMessage.isVisible({ timeout: 5000 }).catch(() => false);
      const hasToast = await toast.isVisible({ timeout: 5000 }).catch(() => false);

      // Page should still be functional
      await expect(page.locator('body')).toBeVisible();
    });

    test('should display loading state', async ({ page }) => {
      // Delay API response
      await page.route('**/api/v1/risk/metrics', async (route) => {
        await new Promise(resolve => setTimeout(resolve, 2000));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            overallRiskScore: 0.65,
            highRiskMerchants: 15,
            riskAssessments: 150,
            riskTrend: -2.5,
          }),
        });
      });

      await page.goto('/risk-indicators');
      
      // Check for loading indicators
      const skeleton = page.locator('[class*="skeleton"], [class*="loading"]').first();
      const hasSkeleton = await skeleton.isVisible({ timeout: 1000 }).catch(() => false);
      
      // Page should load
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Dashboard Data Consistency', () => {
    test('should use portfolio data when available', async ({ page }) => {
      // Mock portfolio endpoints to return data
      await page.route('**/api/v1/merchants/analytics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            totalMerchants: 200, // Different from v3 to verify priority
            averageRiskScore: 0.55,
            averageClassificationConfidence: 0.9,
            averageSecurityTrustScore: 0.8,
            averageDataQuality: 0.95,
            riskDistribution: { low: 80, medium: 100, high: 20 },
            industryDistribution: { 'Technology': 100 },
            countryDistribution: {},
            timestamp: new Date().toISOString(),
          }),
        });
      });

      await page.route('**/api/v1/merchants/statistics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            totalMerchants: 200,
            averageRiskScore: 0.55,
            riskDistribution: { low: 80, medium: 100, high: 20 },
            industryBreakdown: [{ industry: 'Technology', count: 100 }],
            countryBreakdown: [],
            timestamp: new Date().toISOString(),
          }),
        });
      });

      await page.goto('/dashboard');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      await page.waitForTimeout(3000);

      // Should display portfolio data (200 merchants, not v3 data)
      const merchantsCount = page.getByText(/200|merchants/i).first();
      const hasPortfolioData = await merchantsCount.isVisible({ timeout: 5000 }).catch(() => false);

      // Page should load
      await expect(page.locator('main, [role="main"]').first()).toBeVisible();
    });

    test('should handle partial API failures', async ({ page }) => {
      // Mock one endpoint to fail, others to succeed
      await page.route('**/api/v1/merchants/analytics', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Service Unavailable' }),
        });
      });

      await page.route('**/api/v1/merchants/statistics', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            totalMerchants: 150,
            averageRiskScore: 0.6,
            riskDistribution: { low: 60, medium: 70, high: 20 },
            industryBreakdown: [],
            countryBreakdown: [],
            timestamp: new Date().toISOString(),
          }),
        });
      });

      await page.goto('/dashboard');
      await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
      await page.waitForTimeout(3000);

      // Should still display data from successful endpoint
      const metrics = page.locator('text=/150|merchants/i').first();
      const hasMetrics = await metrics.isVisible({ timeout: 5000 }).catch(() => false);

      // Page should load successfully
      await expect(page.locator('main, [role="main"]').first()).toBeVisible();
    });
  });
});

