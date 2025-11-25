// TypeScript types for Merchant Details API

export interface Merchant {
  id: string;
  businessName: string;
  name?: string; // Alternative field name from backend
  legalName?: string;
  registrationNumber?: string;
  taxId?: string;
  industry?: string;
  industryCode?: string;
  businessType?: string;
  description?: string;
  status: string;
  website?: string;
  email?: string;
  phone?: string;
  address?: Address;
  portfolioType?: string;
  riskLevel?: string;
  complianceStatus?: string;
  // Financial information fields
  foundedDate?: string; // ISO date string from backend founded_date
  employeeCount?: number; // From backend employee_count
  annualRevenue?: number; // From backend annual_revenue
  // System information fields
  createdBy?: string; // From backend created_by
  metadata?: Record<string, unknown>; // JSONB metadata if available
  createdAt: string;
  updatedAt: string;
}

export interface Address {
  // Primary address fields
  street?: string; // Legacy field, may be street1
  street1?: string; // From backend address.street1 or address_street1
  street2?: string; // From backend address.street2 or address_street2
  city?: string;
  state?: string;
  postalCode?: string; // From backend address.postal_code or address_postal_code
  country?: string;
  countryCode?: string; // From backend address.country_code or address_country_code
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
  metadata?: {
    pageAnalysis?: {
      pagesAnalyzed?: number;
      analysisMethod?: 'multi_page' | 'single_page' | 'url_only';
      structuredDataFound?: boolean;
    };
    brandMatch?: {
      isBrandMatch?: boolean;
      brandName?: string;
      confidence?: number;
    };
    dataSourcePriority?: {
      websiteContent?: 'primary' | 'secondary' | 'none';
      businessName?: 'primary' | 'secondary' | 'none';
    };
    codeGeneration?: {
      method: 'industry_only' | 'keyword_only' | 'hybrid';
      sources: string[];
      industriesAnalyzed: string[];
      keywordMatches: number;
      industryMatches: number;
      totalCodesGenerated: number;
    };
  };
}

export interface IndustryCode {
  code: string;
  description: string;
  confidence: number;
  // New optional fields for hybrid code generation:
  source?: ('industry' | 'keyword' | 'both')[];
  matchType?: 'exact' | 'partial' | 'synonym';
  relevanceScore?: number;
  industries?: string[];
  isPrimary?: boolean;
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
  overallScore?: number;
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

// Portfolio-level analytics types
export interface PortfolioAnalytics {
  totalMerchants: number;
  averageRiskScore: number;
  averageClassificationConfidence: number;
  averageSecurityTrustScore: number;
  averageDataQuality: number;
  riskDistribution: {
    low: number;
    medium: number;
    high: number;
  };
  industryDistribution: Record<string, number>;
  countryDistribution: Record<string, number>;
  timestamp: string;
}

export interface PortfolioStatistics {
  totalMerchants: number;
  totalAssessments: number;
  averageRiskScore: number;
  riskDistribution: {
    low: number;
    medium: number;
    high: number;
  };
  industryBreakdown: Array<{
    industry: string;
    count: number;
    averageRiskScore: number;
  }>;
  countryBreakdown: Array<{
    country: string;
    count: number;
    averageRiskScore: number;
  }>;
  timestamp: string;
}

// Risk trends types
export interface RiskTrends {
  trends: RiskTrend[];
  summary: TrendSummary;
}

export interface RiskTrend {
  industry: string;
  country: string;
  average_risk_score: number;
  trend_direction: 'improving' | 'worsening' | 'stable';
  change_percentage: number;
  sample_size: number;
}

export interface TrendSummary {
  total_assessments: number;
  average_risk_score: number;
  high_risk_percentage: number;
}

// Risk insights types
export interface RiskInsights {
  insights: RiskInsight[];
  recommendations: Recommendation[];
}

export interface RiskInsight {
  type: string;
  title: string;
  description: string;
  impact: 'low' | 'medium' | 'high';
  recommendation: string;
}

export interface Recommendation {
  category: string;
  action: string;
  priority: 'low' | 'medium' | 'high';
}

// Risk benchmarks types
export interface RiskBenchmarks {
  industry_code: string;
  industry_type: 'mcc' | 'naics' | 'sic';
  average_risk_score?: number;
  median_risk_score?: number;
  percentile_25?: number;
  percentile_75?: number;
  percentile_90?: number;
  sample_size: number;
  benchmarks: {
    average?: number;
    median?: number;
    p25?: number;
    p75?: number;
    p90?: number;
  };
}

// Merchant risk score type
export interface MerchantRiskScore {
  merchant_id: string;
  risk_score?: number;
  risk_level: 'low' | 'medium' | 'high';
  confidence_score?: number;
  assessment_date: string;
  factors: Array<{
    category: string;
    score?: number;
    weight: number;
  }>;
}

// Comparison result types
export interface PortfolioComparison {
  merchantScore: number;
  portfolioAverage: number;
  portfolioMedian: number;
  percentile: number;
  position: 'above_average' | 'below_average' | 'average';
  difference: number;
  differencePercentage: number;
}

export interface BenchmarkComparison {
  merchantScore: number;
  industryAverage: number;
  industryMedian: number;
  industryPercentile75: number;
  industryPercentile90: number;
  percentile: number;
  position: 'top_10' | 'top_25' | 'average' | 'bottom_25' | 'bottom_10';
  difference: number;
  differencePercentage: number;
}

export interface AnalyticsComparison {
  merchant: {
    classificationConfidence: number;
    securityTrustScore: number;
    dataQuality: number;
  };
  portfolio: {
    averageClassificationConfidence: number;
    averageSecurityTrustScore: number;
    averageDataQuality: number;
  };
  differences: {
    classificationConfidence: number;
    securityTrustScore: number;
    dataQuality: number;
  };
  percentages: {
    classificationConfidence: number;
    securityTrustScore: number;
    dataQuality: number;
  };
}

