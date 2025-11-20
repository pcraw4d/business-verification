/**
 * Safe number formatting utilities
 * Handles undefined, null, and NaN values gracefully
 */

/**
 * Safely formats a number with fixed decimal places
 * @param value - The number to format (can be undefined, null, or NaN)
 * @param decimals - Number of decimal places (default: 1)
 * @param fallback - Fallback value if number is invalid (default: 'N/A')
 * @returns Formatted string
 */
export function formatNumber(
  value: number | undefined | null,
  decimals: number = 1,
  fallback: string = 'N/A'
): string {
  if (value == null || isNaN(value) || !isFinite(value)) {
    return fallback;
  }
  return value.toFixed(decimals);
}

/**
 * Formats a number as a percentage (multiplies by 100)
 * @param value - The number to format (0-1 scale, can be undefined, null, or NaN)
 * @param decimals - Number of decimal places (default: 1)
 * @param fallback - Fallback value if number is invalid (default: 'N/A')
 * @returns Formatted percentage string with % symbol
 */
export function formatPercent(
  value: number | undefined | null,
  decimals: number = 1,
  fallback: string = 'N/A'
): string {
  if (value == null || isNaN(value) || !isFinite(value)) {
    return fallback;
  }
  return `${(value * 100).toFixed(decimals)}%`;
}

/**
 * Formats a number as a percentage with optional sign prefix
 * @param value - The number to format (can be undefined, null, or NaN)
 * @param decimals - Number of decimal places (default: 1)
 * @param fallback - Fallback value if number is invalid (default: 'N/A')
 * @returns Formatted percentage string with + or - prefix
 */
export function formatPercentWithSign(
  value: number | undefined | null,
  decimals: number = 1,
  fallback: string = 'N/A'
): string {
  if (value == null || isNaN(value) || !isFinite(value)) {
    return fallback;
  }
  const sign = value > 0 ? '+' : (value < 0 ? '' : ''); // No sign for zero
  return `${sign}${(value * 100).toFixed(decimals)}%`;
}

/**
 * Formats a percentile value (0-100 scale)
 * @param value - The percentile value (can be undefined, null, or NaN)
 * @param decimals - Number of decimal places (default: 0)
 * @param fallback - Fallback value if number is invalid (default: 'N/A')
 * @returns Formatted percentile string with "th" suffix
 */
export function formatPercentile(
  value: number | undefined | null,
  decimals: number = 0,
  fallback: string = 'N/A'
): string {
  if (value == null || isNaN(value) || !isFinite(value)) {
    return fallback;
  }
  return `${value.toFixed(decimals)}th`;
}

/**
 * Formats a number with thousand separators
 * @param value - The number to format (can be undefined, null, or NaN)
 * @param decimals - Number of decimal places (default: 0)
 * @param fallback - Fallback value if number is invalid (default: 'N/A')
 * @returns Formatted string with thousand separators
 */
export function formatNumberWithSeparators(
  value: number | undefined | null,
  decimals: number = 0,
  fallback: string = 'N/A'
): string {
  if (value == null || isNaN(value) || !isFinite(value)) {
    return fallback;
  }
  return value.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

/**
 * Safely multiplies a number by a multiplier and formats it
 * @param value - The number to multiply (can be undefined, null, or NaN)
 * @param multiplier - The multiplier (default: 100 for percentage conversion)
 * @param decimals - Number of decimal places (default: 1)
 * @param fallback - Fallback value if number is invalid (default: 'N/A')
 * @returns Formatted string
 */
export function formatMultiplied(
  value: number | undefined | null,
  multiplier: number = 100,
  decimals: number = 1,
  fallback: string = 'N/A'
): string {
  if (value == null || isNaN(value) || !isFinite(value)) {
    return fallback;
  }
  return (value * multiplier).toFixed(decimals);
}

