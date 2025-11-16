// Manual mock for @/lib/api
// Use global jest which is available in Jest test environment
// @ts-ignore - jest is available globally in Jest
export const getMerchant = jest.fn();
export const getMerchantAnalytics = jest.fn();
export const getWebsiteAnalysis = jest.fn();
export const getRiskAssessment = jest.fn();
export const startRiskAssessment = jest.fn();
export const getAssessmentStatus = jest.fn();
export const getRiskHistory = jest.fn();
export const getRiskPredictions = jest.fn();
export const explainRiskAssessment = jest.fn();
export const getRiskRecommendations = jest.fn();
export const getRiskIndicators = jest.fn();
export const getEnrichmentSources = jest.fn();
export const triggerEnrichment = jest.fn();

