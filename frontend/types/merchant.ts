// TypeScript types for Merchant Details API

export interface Merchant {
  id: string;
  businessName: string;
  industry?: string;
  status: string;
  website?: string;
  email?: string;
  phone?: string;
  address?: Address;
  createdAt: string;
  updatedAt: string;
}

export interface Address {
  street?: string;
  city?: string;
  state?: string;
  postalCode?: string;
  country?: string;
}

export interface AnalyticsData {
  merchantId: string;
  classification: ClassificationData;
  security: SecurityData;
  quality: QualityData;
  intelligence?: IntelligenceData;
  timestamp: string;
}

export interface ClassificationData {
  primaryIndustry: string;
  confidenceScore: number;
  riskLevel: string;
  mccCodes?: IndustryCode[];
  sicCodes?: IndustryCode[];
  naicsCodes?: IndustryCode[];
}

export interface IndustryCode {
  code: string;
  description: string;
  confidence: number;
}

export interface SecurityData {
  trustScore: number;
  sslValid: boolean;
  sslExpiryDate?: string;
  securityHeaders?: SecurityHeader[];
}

export interface SecurityHeader {
  header: string;
  present: boolean;
  value?: string;
}

export interface QualityData {
  completenessScore: number;
  dataPoints: number;
  missingFields?: string[];
}

export interface IntelligenceData {
  businessAge?: number;
  employeeCount?: number;
  annualRevenue?: number;
}

export interface WebsiteAnalysisData {
  merchantId: string;
  websiteUrl: string;
  ssl: SSLData;
  securityHeaders: SecurityHeader[];
  performance: PerformanceData;
  accessibility: AccessibilityData;
  lastAnalyzed: string;
}

export interface SSLData {
  valid: boolean;
  expiryDate?: string;
  issuer?: string;
  grade?: string;
}

export interface PerformanceData {
  loadTime: number;
  pageSize: number;
  requests: number;
  score: number;
}

export interface AccessibilityData {
  score: number;
  issues?: string[];
}

export interface RiskAssessment {
  id: string;
  merchantId: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  options: AssessmentOptions;
  result?: RiskAssessmentResult;
  progress: number;
  estimatedCompletion?: string;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
}

export interface AssessmentOptions {
  includeHistory: boolean;
  includePredictions: boolean;
}

export interface RiskAssessmentResult {
  overallScore: number;
  riskLevel: string;
  factors: RiskFactor[];
}

export interface RiskFactor {
  name: string;
  score: number;
  weight: number;
}

export interface RiskAssessmentRequest {
  merchantId: string;
  options: AssessmentOptions;
}

export interface RiskAssessmentResponse {
  assessmentId: string;
  status: string;
  estimatedCompletion?: string;
}

export interface AssessmentStatusResponse {
  assessmentId: string;
  merchantId: string;
  status: string;
  progress: number;
  estimatedCompletion?: string;
  result?: RiskAssessmentResult;
  completedAt?: string;
}

export interface RiskIndicatorsData {
  merchantId: string;
  indicators: RiskIndicator[];
  timestamp?: string;
}

export interface RiskIndicator {
  id: string;
  title: string;
  description: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  status?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface EnrichmentSource {
  id: string;
  name: string;
  description?: string;
  enabled?: boolean;
}

// Merchant List Types
export interface MerchantListParams {
  page?: number;
  pageSize?: number;
  portfolioType?: string;
  riskLevel?: string;
  status?: string;
  search?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface MerchantListItem {
  id: string;
  name: string;
  legal_name?: string;
  registration_number?: string;
  industry?: string;
  portfolio_type?: string;
  risk_level?: string;
  compliance_status?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface MerchantListResponse {
  merchants: MerchantListItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;
}

