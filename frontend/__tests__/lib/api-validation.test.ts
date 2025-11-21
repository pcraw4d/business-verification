import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  validateAPIResponse,
  MerchantSchema,
  RiskAssessmentSchema,
  DashboardMetricsSchema,
  RiskMetricsSchema,
  hasFinancialData,
  hasCompleteAddress,
  hasRiskAssessmentResult,
} from '@/lib/api-validation';
import { z } from 'zod';

describe('API Validation', () => {
  const originalEnv = process.env.NODE_ENV;

  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    process.env.NODE_ENV = originalEnv;
    vi.restoreAllMocks();
  });

  describe('validateAPIResponse', () => {
    it('should validate correct merchant data', () => {
      const validMerchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };

      const result = validateAPIResponse(MerchantSchema, validMerchant, 'getMerchant');
      expect(result).toEqual(validMerchant);
    });

    it('should throw error for invalid merchant data', () => {
      const invalidMerchant = {
        id: 'merchant-123',
        // Missing required fields: businessName, status, createdAt, updatedAt
      };

      expect(() => {
        validateAPIResponse(MerchantSchema, invalidMerchant, 'getMerchant');
      }).toThrow('API response validation failed');
    });

    it('should validate optional fields correctly', () => {
      const merchantWithOptionalFields = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        foundedDate: '2020-01-01',
        employeeCount: 50,
        annualRevenue: 1000000,
        email: 'test@example.com',
      };

      const result = validateAPIResponse(MerchantSchema, merchantWithOptionalFields, 'getMerchant');
      expect(result.foundedDate).toBe('2020-01-01');
      expect(result.employeeCount).toBe(50);
      expect(result.annualRevenue).toBe(1000000);
    });

    it('should log validation errors in development mode', () => {
      process.env.NODE_ENV = 'development';
      const consoleErrorSpy = vi.spyOn(console, 'error');

      const invalidData = { id: 'test' };

      try {
        validateAPIResponse(MerchantSchema, invalidData, 'test-endpoint');
      } catch (error) {
        // Expected to throw
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        '[API Validation] Validation failed:',
        expect.objectContaining({
          endpoint: 'test-endpoint',
        })
      );
    });

    it('should not log detailed errors in production mode', () => {
      process.env.NODE_ENV = 'production';
      const consoleErrorSpy = vi.spyOn(console, 'error');

      const invalidData = { id: 'test' };

      try {
        validateAPIResponse(MerchantSchema, invalidData, 'test-endpoint');
      } catch (error) {
        // Expected to throw
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        '[API Validation] Validation failed for',
        'test-endpoint'
      );
      expect(consoleErrorSpy).not.toHaveBeenCalledWith(
        expect.stringContaining('Zod issues'),
        expect.anything()
      );
    });

    it('should validate risk assessment data', () => {
      const validAssessment = {
        id: 'assessment-123',
        merchantId: 'merchant-123',
        status: 'completed' as const,
        options: {
          includeHistory: true,
          includePredictions: false,
        },
        progress: 100,
        result: {
          overallScore: 7.5,
          riskLevel: 'medium',
          factors: [],
        },
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };

      const result = validateAPIResponse(RiskAssessmentSchema, validAssessment, 'getRiskAssessment');
      expect(result.id).toBe('assessment-123');
      expect(result.status).toBe('completed');
      expect(result.options.includeHistory).toBe(true);
      expect(result.progress).toBe(100);
    });

    it('should validate dashboard metrics', () => {
      const validMetrics = {
        totalMerchants: 100,
        revenue: 1000000,
        growthRate: 0.15,
        analyticsScore: 85,
      };

      const result = validateAPIResponse(DashboardMetricsSchema, validMetrics, 'getDashboardMetrics');
      expect(result.totalMerchants).toBe(100);
      expect(result.growthRate).toBe(0.15);
    });

    it('should validate risk metrics with optional critical field', () => {
      const validRiskMetrics = {
        overallRiskScore: 6.5,
        highRiskMerchants: 10,
        riskAssessments: 50,
        riskTrend: 0.1,
        riskDistribution: {
          low: 30,
          medium: 15,
          high: 5,
          // critical is optional
        },
      };

      const result = validateAPIResponse(RiskMetricsSchema, validRiskMetrics, 'getRiskMetrics');
      expect(result.overallRiskScore).toBe(6.5);
      expect(result.riskDistribution?.critical).toBeUndefined();
    });

    it('should validate risk metrics with critical field', () => {
      const validRiskMetrics = {
        overallRiskScore: 6.5,
        highRiskMerchants: 10,
        riskAssessments: 50,
        riskTrend: 0.1,
        riskDistribution: {
          low: 30,
          medium: 15,
          high: 5,
          critical: 2,
        },
      };

      const result = validateAPIResponse(RiskMetricsSchema, validRiskMetrics, 'getRiskMetrics');
      expect(result.riskDistribution?.critical).toBe(2);
    });
  });

  describe('Type Guards', () => {
    describe('hasFinancialData', () => {
      it('should return true for merchant with all financial data', () => {
        const merchant = {
          id: 'test',
          businessName: 'Test',
          status: 'active',
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
          foundedDate: '2020-01-01',
          employeeCount: 50,
          annualRevenue: 1000000,
        };

        expect(hasFinancialData(merchant)).toBe(true);
      });

      it('should return false for merchant missing financial data', () => {
        const merchant = {
          id: 'test',
          businessName: 'Test',
          status: 'active',
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
        };

        expect(hasFinancialData(merchant)).toBe(false);
      });

      it('should return false for merchant with partial financial data', () => {
        const merchant = {
          id: 'test',
          businessName: 'Test',
          status: 'active',
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
          foundedDate: '2020-01-01',
          // Missing employeeCount and annualRevenue
        };

        expect(hasFinancialData(merchant)).toBe(false);
      });
    });

    describe('hasCompleteAddress', () => {
      it('should return true for merchant with complete address', () => {
        const merchant = {
          id: 'test',
          businessName: 'Test',
          status: 'active',
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
          address: {
            street1: '123 Main St',
            city: 'New York',
            country: 'USA',
          },
        };

        expect(hasCompleteAddress(merchant)).toBe(true);
      });

      it('should return false for merchant without address', () => {
        const merchant = {
          id: 'test',
          businessName: 'Test',
          status: 'active',
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
        };

        expect(hasCompleteAddress(merchant)).toBe(false);
      });

      it('should return false for merchant with incomplete address', () => {
        const merchant = {
          id: 'test',
          businessName: 'Test',
          status: 'active',
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
          address: {
            street1: '123 Main St',
            // Missing city and country
          },
        };

        expect(hasCompleteAddress(merchant)).toBe(false);
      });
    });

    describe('hasRiskAssessmentResult', () => {
      it('should return true for completed assessment with result', () => {
        const assessment = {
          id: 'test',
          merchantId: 'merchant-123',
          status: 'completed' as const,
          result: {
            overallScore: 7.5,
            riskLevel: 'medium',
            factors: [],
          },
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
        };

        expect(hasRiskAssessmentResult(assessment)).toBe(true);
      });

      it('should return false for pending assessment', () => {
        const assessment = {
          id: 'test',
          merchantId: 'merchant-123',
          status: 'pending' as const,
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
        };

        expect(hasRiskAssessmentResult(assessment)).toBe(false);
      });

      it('should return false for completed assessment without result', () => {
        const assessment = {
          id: 'test',
          merchantId: 'merchant-123',
          status: 'completed' as const,
          createdAt: '2024-01-01',
          updatedAt: '2024-01-01',
        };

        expect(hasRiskAssessmentResult(assessment)).toBe(false);
      });
    });
  });
});

