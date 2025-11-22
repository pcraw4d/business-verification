/**
 * Unit tests for Merchant and Address type definitions
 * Tests type guards and type safety for all Phase 1-5 fields
 */

import { describe, it, expect } from 'vitest';
import type { Merchant, Address } from '@/types/merchant';

describe('Merchant Type', () => {
  describe('Required Fields', () => {
    it('should require id, businessName, status, createdAt, updatedAt', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };

      expect(merchant.id).toBe('merchant-123');
      expect(merchant.businessName).toBe('Test Business');
      expect(merchant.status).toBe('active');
      expect(merchant.createdAt).toBe('2024-01-01T00:00:00Z');
      expect(merchant.updatedAt).toBe('2024-01-01T00:00:00Z');
    });
  });

  describe('Financial Information Fields (Phase 1)', () => {
    it('should support foundedDate as ISO string', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        foundedDate: '2020-01-15T00:00:00Z',
      };

      expect(merchant.foundedDate).toBe('2020-01-15T00:00:00Z');
      expect(new Date(merchant.foundedDate!)).toBeInstanceOf(Date);
    });

    it('should support employeeCount as number', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        employeeCount: 150,
      };

      expect(merchant.employeeCount).toBe(150);
      expect(typeof merchant.employeeCount).toBe('number');
    });

    it('should support annualRevenue as number', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        annualRevenue: 5000000.50,
      };

      expect(merchant.annualRevenue).toBe(5000000.50);
      expect(typeof merchant.annualRevenue).toBe('number');
    });

    it('should support all financial fields together', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        foundedDate: '2020-01-15T00:00:00Z',
        employeeCount: 150,
        annualRevenue: 5000000.50,
      };

      expect(merchant.foundedDate).toBeDefined();
      expect(merchant.employeeCount).toBeDefined();
      expect(merchant.annualRevenue).toBeDefined();
    });
  });

  describe('System Information Fields (Phase 1)', () => {
    it('should support createdBy as string', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        createdBy: 'user-123',
      };

      expect(merchant.createdBy).toBe('user-123');
      expect(typeof merchant.createdBy).toBe('string');
    });

    it('should support metadata as Record<string, unknown>', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        metadata: {
          source: 'manual',
          verified: true,
          tags: ['enterprise', 'high-value'],
        },
      };

      expect(merchant.metadata).toBeDefined();
      expect(merchant.metadata?.source).toBe('manual');
      expect(merchant.metadata?.verified).toBe(true);
      expect(Array.isArray(merchant.metadata?.tags)).toBe(true);
    });

    it('should support empty metadata object', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        metadata: {},
      };

      expect(merchant.metadata).toBeDefined();
      expect(Object.keys(merchant.metadata!)).toHaveLength(0);
    });
  });

  describe('Optional Fields', () => {
    it('should allow all optional fields to be undefined', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };

      expect(merchant.foundedDate).toBeUndefined();
      expect(merchant.employeeCount).toBeUndefined();
      expect(merchant.annualRevenue).toBeUndefined();
      expect(merchant.createdBy).toBeUndefined();
      expect(merchant.metadata).toBeUndefined();
      expect(merchant.address).toBeUndefined();
    });
  });
});

describe('Address Type', () => {
  describe('Primary Address Fields', () => {
    it('should support street1 and street2 separately', () => {
      const address: Address = {
        street1: '123 Main Street',
        street2: 'Suite 100',
        city: 'San Francisco',
        state: 'CA',
        postalCode: '94102',
        country: 'United States',
        countryCode: 'US',
      };

      expect(address.street1).toBe('123 Main Street');
      expect(address.street2).toBe('Suite 100');
      expect(address.city).toBe('San Francisco');
      expect(address.state).toBe('CA');
      expect(address.postalCode).toBe('94102');
      expect(address.country).toBe('United States');
      expect(address.countryCode).toBe('US');
    });

    it('should support legacy street field for backward compatibility', () => {
      const address: Address = {
        street: '123 Main Street',
        city: 'San Francisco',
        state: 'CA',
        postalCode: '94102',
        country: 'United States',
      };

      expect(address.street).toBe('123 Main Street');
      expect(address.street1).toBeUndefined();
    });

    it('should support both street and street1', () => {
      const address: Address = {
        street: '123 Main Street',
        street1: '123 Main Street',
        street2: 'Suite 100',
        city: 'San Francisco',
        state: 'CA',
        postalCode: '94102',
        country: 'United States',
        countryCode: 'US',
      };

      expect(address.street).toBe('123 Main Street');
      expect(address.street1).toBe('123 Main Street');
      expect(address.street2).toBe('Suite 100');
    });
  });

  describe('Country Code Field (Phase 1)', () => {
    it('should support countryCode as string', () => {
      const address: Address = {
        street1: '123 Main Street',
        city: 'San Francisco',
        state: 'CA',
        postalCode: '94102',
        country: 'United States',
        countryCode: 'US',
      };

      expect(address.countryCode).toBe('US');
      expect(typeof address.countryCode).toBe('string');
    });

    it('should allow countryCode to be undefined', () => {
      const address: Address = {
        street1: '123 Main Street',
        city: 'San Francisco',
        state: 'CA',
        postalCode: '94102',
        country: 'United States',
      };

      expect(address.countryCode).toBeUndefined();
    });
  });

  describe('All Fields Optional', () => {
    it('should allow all address fields to be undefined', () => {
      const address: Address = {};

      expect(address.street).toBeUndefined();
      expect(address.street1).toBeUndefined();
      expect(address.street2).toBeUndefined();
      expect(address.city).toBeUndefined();
      expect(address.state).toBeUndefined();
      expect(address.postalCode).toBeUndefined();
      expect(address.country).toBeUndefined();
      expect(address.countryCode).toBeUndefined();
    });
  });
});

describe('Type Guards', () => {
  describe('hasFinancialData', () => {
    it('should identify merchant with all financial data', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        foundedDate: '2020-01-15T00:00:00Z',
        employeeCount: 150,
        annualRevenue: 5000000.50,
      };

      const hasFinancialData = !!(
        merchant.foundedDate &&
        merchant.employeeCount &&
        merchant.annualRevenue
      );

      expect(hasFinancialData).toBe(true);
    });

    it('should identify merchant with partial financial data', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
        foundedDate: '2020-01-15T00:00:00Z',
      };

      const hasFinancialData = !!(
        merchant.foundedDate &&
        merchant.employeeCount &&
        merchant.annualRevenue
      );

      expect(hasFinancialData).toBe(false);
    });

    it('should identify merchant with no financial data', () => {
      const merchant: Merchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };

      const hasFinancialData = !!(
        merchant.foundedDate &&
        merchant.employeeCount &&
        merchant.annualRevenue
      );

      expect(hasFinancialData).toBe(false);
    });
  });

  describe('hasCompleteAddress', () => {
    it('should identify address with all required fields', () => {
      const address: Address = {
        street1: '123 Main Street',
        city: 'San Francisco',
        state: 'CA',
        postalCode: '94102',
        country: 'United States',
        countryCode: 'US',
      };

      const hasCompleteAddress = !!(
        (address.street1 || address.street) &&
        address.city &&
        address.country
      );

      expect(hasCompleteAddress).toBe(true);
    });

    it('should identify incomplete address', () => {
      const address: Address = {
        street1: '123 Main Street',
        city: 'San Francisco',
      };

      const hasCompleteAddress = !!(
        (address.street1 || address.street) &&
        address.city &&
        address.country
      );

      expect(hasCompleteAddress).toBe(false);
    });
  });
});

