import { describe, it, expect } from 'vitest';
import {
  formatNumber,
  formatPercent,
  formatPercentWithSign,
  formatPercentile,
  formatNumberWithSeparators,
  formatMultiplied,
} from '@/lib/number-format';

describe('number-format utilities', () => {
  describe('formatNumber', () => {
    it('formats valid numbers correctly', () => {
      expect(formatNumber(123.456, 2)).toBe('123.46');
      expect(formatNumber(0, 1)).toBe('0.0');
      expect(formatNumber(-10.5, 1)).toBe('-10.5');
    });

    it('handles undefined', () => {
      expect(formatNumber(undefined)).toBe('N/A');
      expect(formatNumber(undefined, 2, '--')).toBe('--');
    });

    it('handles null', () => {
      expect(formatNumber(null)).toBe('N/A');
      expect(formatNumber(null, 2, '--')).toBe('--');
    });

    it('handles NaN', () => {
      expect(formatNumber(NaN)).toBe('N/A');
      expect(formatNumber(NaN, 2, '--')).toBe('--');
    });

    it('respects decimal places', () => {
      expect(formatNumber(123.456, 0)).toBe('123');
      expect(formatNumber(123.456, 1)).toBe('123.5');
      expect(formatNumber(123.456, 3)).toBe('123.456');
    });

    it('uses custom fallback', () => {
      expect(formatNumber(undefined, 1, '--')).toBe('--');
      expect(formatNumber(null, 1, '0.0')).toBe('0.0');
    });
  });

  describe('formatPercent', () => {
    it('formats valid percentages correctly', () => {
      expect(formatPercent(0.5)).toBe('50.0%');
      expect(formatPercent(0.123)).toBe('12.3%');
      expect(formatPercent(1)).toBe('100.0%');
      expect(formatPercent(0)).toBe('0.0%');
    });

    it('handles undefined', () => {
      expect(formatPercent(undefined)).toBe('N/A');
    });

    it('handles null', () => {
      expect(formatPercent(null)).toBe('N/A');
    });

    it('handles NaN', () => {
      expect(formatPercent(NaN)).toBe('N/A');
    });

    it('respects decimal places', () => {
      expect(formatPercent(0.5, 0)).toBe('50%');
      expect(formatPercent(0.123, 2)).toBe('12.30%');
    });

    it('handles negative values', () => {
      expect(formatPercent(-0.1)).toBe('-10.0%');
    });
  });

  describe('formatPercentWithSign', () => {
    it('formats positive percentages with + sign', () => {
      expect(formatPercentWithSign(0.5)).toBe('+50.0%');
      expect(formatPercentWithSign(0.1)).toBe('+10.0%');
    });

    it('formats negative percentages with - sign', () => {
      expect(formatPercentWithSign(-0.5)).toBe('-50.0%');
      expect(formatPercentWithSign(-0.1)).toBe('-10.0%');
    });

    it('formats zero without sign', () => {
      expect(formatPercentWithSign(0)).toBe('0.0%');
    });

    it('handles undefined', () => {
      expect(formatPercentWithSign(undefined)).toBe('N/A');
    });

    it('handles null', () => {
      expect(formatPercentWithSign(null)).toBe('N/A');
    });

    it('handles NaN', () => {
      expect(formatPercentWithSign(NaN)).toBe('N/A');
    });
  });

  describe('formatPercentile', () => {
    it('formats valid percentiles correctly', () => {
      expect(formatPercentile(50)).toBe('50th');
      expect(formatPercentile(75.5)).toBe('76th');
      expect(formatPercentile(0)).toBe('0th');
      expect(formatPercentile(100)).toBe('100th');
    });

    it('handles undefined', () => {
      expect(formatPercentile(undefined)).toBe('N/A');
    });

    it('handles null', () => {
      expect(formatPercentile(null)).toBe('N/A');
    });

    it('handles NaN', () => {
      expect(formatPercentile(NaN)).toBe('N/A');
    });

    it('respects decimal places', () => {
      expect(formatPercentile(50.5, 1)).toBe('50.5th');
      expect(formatPercentile(75.25, 2)).toBe('75.25th');
    });
  });

  describe('formatNumberWithSeparators', () => {
    it('formats numbers with thousand separators', () => {
      expect(formatNumberWithSeparators(1000)).toBe('1,000');
      expect(formatNumberWithSeparators(1234567)).toBe('1,234,567');
      expect(formatNumberWithSeparators(123.45, 2)).toBe('123.45');
    });

    it('handles undefined', () => {
      expect(formatNumberWithSeparators(undefined)).toBe('N/A');
    });

    it('handles null', () => {
      expect(formatNumberWithSeparators(null)).toBe('N/A');
    });

    it('handles NaN', () => {
      expect(formatNumberWithSeparators(NaN)).toBe('N/A');
    });

    it('respects decimal places', () => {
      expect(formatNumberWithSeparators(1234.567, 2)).toBe('1,234.57');
      expect(formatNumberWithSeparators(1234.567, 0)).toBe('1,235');
    });
  });

  describe('formatMultiplied', () => {
    it('multiplies and formats correctly', () => {
      expect(formatMultiplied(0.5, 100, 1)).toBe('50.0');
      expect(formatMultiplied(0.123, 100, 2)).toBe('12.30');
      expect(formatMultiplied(2, 10, 0)).toBe('20');
    });

    it('uses default multiplier of 100', () => {
      expect(formatMultiplied(0.5)).toBe('50.0');
      expect(formatMultiplied(0.1)).toBe('10.0');
    });

    it('handles undefined', () => {
      expect(formatMultiplied(undefined)).toBe('N/A');
    });

    it('handles null', () => {
      expect(formatMultiplied(null)).toBe('N/A');
    });

    it('handles NaN', () => {
      expect(formatMultiplied(NaN)).toBe('N/A');
    });

    it('handles custom multiplier', () => {
      expect(formatMultiplied(5, 2, 1)).toBe('10.0');
      expect(formatMultiplied(10, 0.1, 2)).toBe('1.00');
    });
  });

  describe('edge cases', () => {
    it('handles very large numbers', () => {
      expect(formatNumber(Number.MAX_SAFE_INTEGER, 0)).toBe('9007199254740991');
      expect(formatPercent(1)).toBe('100.0%');
    });

    it('handles very small numbers', () => {
      expect(formatNumber(0.0001, 4)).toBe('0.0001');
      expect(formatPercent(0.0001, 4)).toBe('0.0100%');
    });

    it('handles Infinity', () => {
      expect(formatNumber(Infinity)).toBe('N/A');
      expect(formatPercent(Infinity)).toBe('N/A');
    });

    it('handles -Infinity', () => {
      expect(formatNumber(-Infinity)).toBe('N/A');
      expect(formatPercent(-Infinity)).toBe('N/A');
    });
  });
});

