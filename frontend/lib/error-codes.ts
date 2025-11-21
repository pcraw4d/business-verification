/**
 * Error codes for merchant detail components
 * Format: COMPONENT-ERROR_TYPE-NUMBER
 * 
 * Components:
 * - PC: PortfolioComparison
 * - RS: RiskScore
 * - AC: AnalyticsComparison
 * - RB: RiskBenchmark
 * - RA: RiskAssessment
 */

export const ErrorCodes = {
  // PortfolioComparison errors
  PORTFOLIO_COMPARISON: {
    MISSING_RISK_SCORE: 'PC-001',
    MISSING_PORTFOLIO_STATS: 'PC-002',
    MISSING_BOTH: 'PC-003',
    INVALID_DATA: 'PC-004',
    FETCH_ERROR: 'PC-005',
  },
  
  // RiskScore errors
  RISK_SCORE: {
    NOT_FOUND: 'RS-001',
    INVALID_DATA: 'RS-002',
    FETCH_ERROR: 'RS-003',
  },
  
  // AnalyticsComparison errors
  ANALYTICS_COMPARISON: {
    MISSING_MERCHANT_ANALYTICS: 'AC-001',
    MISSING_PORTFOLIO_ANALYTICS: 'AC-002',
    MISSING_BOTH: 'AC-003',
    INVALID_DATA: 'AC-004',
    FETCH_ERROR: 'AC-005',
  },
  
  // RiskBenchmark errors
  RISK_BENCHMARK: {
    MISSING_INDUSTRY_CODE: 'RB-001',
    BENCHMARKS_UNAVAILABLE: 'RB-002',
    MISSING_RISK_SCORE: 'RB-003',
    INVALID_DATA: 'RB-004',
    FETCH_ERROR: 'RB-005',
  },
  
  // RiskAssessment errors
  RISK_ASSESSMENT: {
    NOT_FOUND: 'RA-001',
    FETCH_ERROR: 'RA-002',
    START_FAILED: 'RA-003',
  },
} as const;

/**
 * Formats an error message with error code
 */
export function formatErrorWithCode(message: string, code: string): string {
  return `Error ${code}: ${message}`;
}

/**
 * Gets a support documentation link for an error code
 */
export function getErrorSupportLink(code: string): string | null {
  // In a real implementation, this would link to support documentation
  // For now, return null or a generic support link
  return process.env.NEXT_PUBLIC_SUPPORT_URL 
    ? `${process.env.NEXT_PUBLIC_SUPPORT_URL}/errors/${code}`
    : null;
}

