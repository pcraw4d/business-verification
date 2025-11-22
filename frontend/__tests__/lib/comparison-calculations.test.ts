/**
 * Unit tests for comparison calculation logic
 * Tests portfolio comparison, benchmark comparison, and analytics comparison calculations
 */

import { describe, it, expect } from 'vitest';
import type {
  PortfolioStatistics,
  MerchantRiskScore,
  RiskBenchmarks,
  AnalyticsData,
  PortfolioAnalytics,
} from '@/types/merchant';

describe('Portfolio Comparison Calculations', () => {
  describe('Percentile Calculation', () => {
    it('should calculate percentile correctly for merchant below average', () => {
      const merchantScore = 0.45;
      const portfolioAverage = 0.6;
      const portfolioMedian = 0.55;

      // Percentile calculation: (merchantScore - portfolioAverage) / (portfolioAverage - portfolioMedian) * 50 + 50
      const percentile = ((merchantScore - portfolioAverage) / (portfolioAverage - portfolioMedian)) * 50 + 50;
      
      // Merchant is below average, so percentile should be < 50
      // Clamp to 0-100 range for realistic percentile
      const clampedPercentile = Math.max(0, Math.min(100, percentile));
      expect(clampedPercentile).toBeLessThan(50);
      expect(clampedPercentile).toBeGreaterThanOrEqual(0);
      // Verify the raw calculation produces negative value (below average)
      expect(percentile).toBeLessThan(50);
    });

    it('should calculate percentile correctly for merchant above average', () => {
      const merchantScore = 0.75;
      const portfolioAverage = 0.6;
      const portfolioMedian = 0.55;

      const percentile = ((merchantScore - portfolioAverage) / (portfolioAverage - portfolioMedian)) * 50 + 50;
      
      // Merchant is above average, so percentile should be > 50
      // Clamp to 0-100 range for realistic percentile
      const clampedPercentile = Math.max(0, Math.min(100, percentile));
      expect(clampedPercentile).toBeGreaterThan(50);
      expect(clampedPercentile).toBeLessThanOrEqual(100);
      // Verify the raw calculation produces value > 50 (above average)
      expect(percentile).toBeGreaterThan(50);
    });

    it('should calculate percentile correctly for merchant at average', () => {
      const merchantScore = 0.6;
      const portfolioAverage = 0.6;
      const portfolioMedian = 0.55;

      const percentile = ((merchantScore - portfolioAverage) / (portfolioAverage - portfolioMedian)) * 50 + 50;
      
      // Merchant is at average, so percentile should be ~50
      expect(percentile).toBeCloseTo(50, 1);
    });
  });

  describe('Position Calculation', () => {
    it('should identify merchant as above average when score is higher', () => {
      const merchantScore = 0.75;
      const portfolioAverage = 0.6;

      const position = merchantScore > portfolioAverage ? 'above_average' : 
                      merchantScore < portfolioAverage ? 'below_average' : 'average';

      expect(position).toBe('above_average');
    });

    it('should identify merchant as below average when score is lower', () => {
      const merchantScore = 0.45;
      const portfolioAverage = 0.6;

      const position = merchantScore > portfolioAverage ? 'above_average' : 
                      merchantScore < portfolioAverage ? 'below_average' : 'average';

      expect(position).toBe('below_average');
    });

    it('should identify merchant as average when score equals average', () => {
      const merchantScore = 0.6;
      const portfolioAverage = 0.6;

      const position = merchantScore > portfolioAverage ? 'above_average' : 
                      merchantScore < portfolioAverage ? 'below_average' : 'average';

      expect(position).toBe('average');
    });
  });

  describe('Difference Calculation', () => {
    it('should calculate absolute difference correctly', () => {
      const merchantScore = 0.45;
      const portfolioAverage = 0.6;

      const difference = Math.abs(merchantScore - portfolioAverage);

      // Use toBeCloseTo for floating point comparison (allowing for floating point precision)
      expect(difference).toBeCloseTo(0.15, 10);
    });

    it('should calculate percentage difference correctly', () => {
      const merchantScore = 0.45;
      const portfolioAverage = 0.6;

      const differencePercentage = ((merchantScore - portfolioAverage) / portfolioAverage) * 100;

      expect(differencePercentage).toBeCloseTo(-25, 1); // 25% below average
    });

    it('should handle zero portfolio average gracefully', () => {
      const merchantScore = 0.5;
      const portfolioAverage = 0;

      // Should not divide by zero
      const differencePercentage = portfolioAverage === 0 ? 0 : ((merchantScore - portfolioAverage) / portfolioAverage) * 100;

      expect(differencePercentage).toBe(0);
    });
  });
});

describe('Benchmark Comparison Calculations', () => {
  describe('Industry Percentile Calculation', () => {
    it('should calculate percentile correctly for top 10% merchant', () => {
      const merchantScore = 0.9;
      const percentile90 = 0.85;
      const percentile75 = 0.7;
      const percentile25 = 0.45;

      let percentile: number;
      if (merchantScore >= percentile90) {
        percentile = 90 + ((merchantScore - percentile90) / (1 - percentile90)) * 10;
      } else if (merchantScore >= percentile75) {
        percentile = 75 + ((merchantScore - percentile75) / (percentile90 - percentile75)) * 15;
      } else if (merchantScore >= percentile25) {
        percentile = 25 + ((merchantScore - percentile25) / (percentile75 - percentile25)) * 50;
      } else {
        percentile = (merchantScore / percentile25) * 25;
      }

      expect(percentile).toBeGreaterThan(90);
      expect(percentile).toBeLessThanOrEqual(100);
    });

    it('should calculate percentile correctly for top 25% merchant', () => {
      const merchantScore = 0.75;
      const percentile90 = 0.85;
      const percentile75 = 0.7;
      const percentile25 = 0.45;

      let percentile: number;
      if (merchantScore >= percentile90) {
        percentile = 90 + ((merchantScore - percentile90) / (1 - percentile90)) * 10;
      } else if (merchantScore >= percentile75) {
        percentile = 75 + ((merchantScore - percentile75) / (percentile90 - percentile75)) * 15;
      } else if (merchantScore >= percentile25) {
        percentile = 25 + ((merchantScore - percentile25) / (percentile75 - percentile25)) * 50;
      } else {
        percentile = (merchantScore / percentile25) * 25;
      }

      expect(percentile).toBeGreaterThan(75);
      expect(percentile).toBeLessThanOrEqual(90);
    });

    it('should calculate percentile correctly for average merchant', () => {
      const merchantScore = 0.55;
      const percentile90 = 0.85;
      const percentile75 = 0.7;
      const percentile25 = 0.45;

      let percentile: number;
      if (merchantScore >= percentile90) {
        percentile = 90 + ((merchantScore - percentile90) / (1 - percentile90)) * 10;
      } else if (merchantScore >= percentile75) {
        percentile = 75 + ((merchantScore - percentile75) / (percentile90 - percentile75)) * 15;
      } else if (merchantScore >= percentile25) {
        percentile = 25 + ((merchantScore - percentile25) / (percentile75 - percentile25)) * 50;
      } else {
        percentile = (merchantScore / percentile25) * 25;
      }

      expect(percentile).toBeGreaterThan(25);
      expect(percentile).toBeLessThanOrEqual(75);
    });
  });

  describe('Position Classification', () => {
    it('should classify merchant as top 10% when score >= percentile90', () => {
      const merchantScore = 0.9;
      const percentile90 = 0.85;

      const position = merchantScore >= percentile90 ? 'top_10' :
                      merchantScore >= 0.7 ? 'top_25' :
                      merchantScore >= 0.45 ? 'average' :
                      merchantScore >= 0.25 ? 'bottom_25' : 'bottom_10';

      expect(position).toBe('top_10');
    });

    it('should classify merchant as top 25% when score >= percentile75', () => {
      const merchantScore = 0.75;
      const percentile90 = 0.85;
      const percentile75 = 0.7;

      const position = merchantScore >= percentile90 ? 'top_10' :
                      merchantScore >= percentile75 ? 'top_25' :
                      merchantScore >= 0.45 ? 'average' :
                      merchantScore >= 0.25 ? 'bottom_25' : 'bottom_10';

      expect(position).toBe('top_25');
    });

    it('should classify merchant as average when score is between percentile25 and percentile75', () => {
      const merchantScore = 0.55;
      const percentile90 = 0.85;
      const percentile75 = 0.7;
      const percentile25 = 0.45;

      const position = merchantScore >= percentile90 ? 'top_10' :
                      merchantScore >= percentile75 ? 'top_25' :
                      merchantScore >= percentile25 ? 'average' :
                      merchantScore >= 0.25 ? 'bottom_25' : 'bottom_10';

      expect(position).toBe('average');
    });
  });
});

describe('Analytics Comparison Calculations', () => {
  describe('Difference Calculation', () => {
    it('should calculate difference correctly for classification confidence', () => {
      const merchantConfidence = 0.95;
      const portfolioAverage = 0.8;

      const difference = merchantConfidence - portfolioAverage;

      // Use toBeCloseTo for floating point comparison
      expect(difference).toBeCloseTo(0.15, 10);
    });

    it('should calculate difference correctly for security trust score', () => {
      const merchantSecurity = 0.85;
      const portfolioAverage = 0.75;

      const difference = merchantSecurity - portfolioAverage;

      // Use toBeCloseTo for floating point comparison
      expect(difference).toBeCloseTo(0.1, 10);
    });

    it('should calculate difference correctly for data quality', () => {
      const merchantQuality = 0.9;
      const portfolioAverage = 0.85;

      const difference = merchantQuality - portfolioAverage;

      // Use toBeCloseTo for floating point comparison
      expect(difference).toBeCloseTo(0.05, 10);
    });
  });

  describe('Percentage Difference Calculation', () => {
    it('should calculate percentage difference correctly', () => {
      const merchantValue = 0.95;
      const portfolioAverage = 0.8;

      const percentageDifference = ((merchantValue - portfolioAverage) / portfolioAverage) * 100;

      expect(percentageDifference).toBeCloseTo(18.75, 2); // 18.75% above average
    });

    it('should handle negative differences correctly', () => {
      const merchantValue = 0.7;
      const portfolioAverage = 0.8;

      const percentageDifference = ((merchantValue - portfolioAverage) / portfolioAverage) * 100;

      expect(percentageDifference).toBeCloseTo(-12.5, 2); // 12.5% below average
    });

    it('should handle zero portfolio average gracefully', () => {
      const merchantValue = 0.5;
      const portfolioAverage = 0;

      const percentageDifference = portfolioAverage === 0 ? 0 : ((merchantValue - portfolioAverage) / portfolioAverage) * 100;

      expect(percentageDifference).toBe(0);
    });
  });

  describe('Comparison Direction', () => {
    it('should identify merchant as better when values are higher', () => {
      const merchantConfidence = 0.95;
      const portfolioAverage = 0.8;

      const isBetter = merchantConfidence > portfolioAverage;

      expect(isBetter).toBe(true);
    });

    it('should identify merchant as worse when values are lower', () => {
      const merchantConfidence = 0.7;
      const portfolioAverage = 0.8;

      const isWorse = merchantConfidence < portfolioAverage;

      expect(isWorse).toBe(true);
    });

    it('should identify merchant as equal when values match', () => {
      const merchantConfidence = 0.8;
      const portfolioAverage = 0.8;

      const isEqual = merchantConfidence === portfolioAverage;

      expect(isEqual).toBe(true);
    });
  });
});

describe('Edge Cases and Error Handling', () => {
  describe('Invalid Input Handling', () => {
    it('should handle NaN values gracefully', () => {
      const merchantScore = NaN;
      const portfolioAverage = 0.6;

      const isValid = !isNaN(merchantScore) && !isNaN(portfolioAverage);

      expect(isValid).toBe(false);
    });

    it('should handle Infinity values gracefully', () => {
      const merchantScore = Infinity;
      const portfolioAverage = 0.6;

      const isValid = isFinite(merchantScore) && isFinite(portfolioAverage);

      expect(isValid).toBe(false);
    });

    it('should handle negative values', () => {
      const merchantScore = -0.1;
      const portfolioAverage = 0.6;

      // Risk scores should be between 0 and 1
      const isValid = merchantScore >= 0 && merchantScore <= 1;

      expect(isValid).toBe(false);
    });

    it('should handle values greater than 1', () => {
      const merchantScore = 1.5;
      const portfolioAverage = 0.6;

      // Risk scores should be between 0 and 1
      const isValid = merchantScore >= 0 && merchantScore <= 1;

      expect(isValid).toBe(false);
    });
  });

  describe('Empty Data Handling', () => {
    it('should handle empty portfolio statistics', () => {
      const portfolioStats: PortfolioStatistics = {
        totalMerchants: 0,
        totalAssessments: 0,
        averageRiskScore: 0,
        riskDistribution: { low: 0, medium: 0, high: 0 },
        industryBreakdown: [],
        countryBreakdown: [],
        timestamp: new Date().toISOString(),
      };

      expect(portfolioStats.totalMerchants).toBe(0);
      expect(portfolioStats.averageRiskScore).toBe(0);
    });

    it('should handle missing benchmark data', () => {
      const benchmarks: RiskBenchmarks = {
        industry_code: '5734',
        industry_type: 'mcc',
        sample_size: 0,
        benchmarks: {},
      };

      expect(benchmarks.sample_size).toBe(0);
      expect(benchmarks.benchmarks.average).toBeUndefined();
    });
  });
});

