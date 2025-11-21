// API Response Validation using Zod
// Provides runtime validation for API responses to catch type mismatches early

import { z } from 'zod';

// Address schema
export const AddressSchema = z.object({
  street: z.string().optional(),
  street1: z.string().optional(),
  street2: z.string().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  postalCode: z.string().optional(),
  country: z.string().optional(),
  countryCode: z.string().optional(),
});

// Merchant schema - matches Merchant interface from types/merchant.ts
export const MerchantSchema = z.object({
  id: z.string(),
  businessName: z.string(),
  name: z.string().optional(),
  legalName: z.string().optional(),
  registrationNumber: z.string().optional(),
  taxId: z.string().optional(),
  industry: z.string().optional(),
  industryCode: z.string().optional(),
  businessType: z.string().optional(),
  description: z.string().optional(),
  status: z.string(),
  website: z.string().optional(),
  email: z.string().email().optional().or(z.literal('')),
  phone: z.string().optional(),
  address: AddressSchema.optional(),
  portfolioType: z.string().optional(),
  riskLevel: z.string().optional(),
  complianceStatus: z.string().optional(),
  // Financial information fields
  foundedDate: z.string().optional(),
  employeeCount: z.number().optional(),
  annualRevenue: z.number().optional(),
  // System information fields
  createdBy: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

// Analytics data schemas
export const IndustryCodeSchema = z.object({
  code: z.string(),
  description: z.string(),
  confidence: z.number(),
});

export const ClassificationDataSchema = z.object({
  primaryIndustry: z.string(),
  confidenceScore: z.number(),
  riskLevel: z.string(),
  mccCodes: z.array(IndustryCodeSchema).optional(),
  sicCodes: z.array(IndustryCodeSchema).optional(),
  naicsCodes: z.array(IndustryCodeSchema).optional(),
});

export const SecurityHeaderSchema = z.object({
  header: z.string(),
  present: z.boolean(),
  value: z.string().optional(),
});

export const SecurityDataSchema = z.object({
  trustScore: z.number(),
  sslValid: z.boolean(),
  sslExpiryDate: z.string().optional(),
  securityHeaders: z.array(SecurityHeaderSchema).optional(),
});

export const QualityDataSchema = z.object({
  completenessScore: z.number(),
  dataPoints: z.number(),
  missingFields: z.array(z.string()).optional(),
});

export const IntelligenceDataSchema = z.object({
  businessAge: z.number().optional(),
  employeeCount: z.number().optional(),
  annualRevenue: z.number().optional(),
});

export const AnalyticsDataSchema = z.object({
  merchantId: z.string(),
  classification: ClassificationDataSchema,
  security: SecurityDataSchema,
  quality: QualityDataSchema,
  intelligence: IntelligenceDataSchema.optional(),
  timestamp: z.string(),
});

// Risk assessment schemas
export const RiskFactorSchema = z.object({
  name: z.string(),
  score: z.number(),
  weight: z.number(),
});

export const RiskAssessmentResultSchema = z.object({
  overallScore: z.number().optional(),
  riskLevel: z.string(),
  factors: z.array(RiskFactorSchema),
});

export const AssessmentOptionsSchema = z.object({
  includeHistory: z.boolean(),
  includePredictions: z.boolean(),
});

export const RiskAssessmentSchema = z.object({
  id: z.string(),
  merchantId: z.string(),
  status: z.enum(['pending', 'processing', 'completed', 'failed']),
  options: AssessmentOptionsSchema,
  result: RiskAssessmentResultSchema.optional(),
  progress: z.number(),
  estimatedCompletion: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
  completedAt: z.string().optional(),
});

// Risk score schema
export const MerchantRiskScoreSchema = z.object({
  merchant_id: z.string(),
  risk_score: z.number().optional(),
  risk_level: z.enum(['low', 'medium', 'high']),
  confidence_score: z.number().optional(),
  assessment_date: z.string(),
  factors: z.array(
    z.object({
      category: z.string(),
      score: z.number().optional(),
      weight: z.number(),
    })
  ),
});

// Portfolio statistics schema
export const PortfolioStatisticsSchema = z.object({
  totalMerchants: z.number(),
  totalAssessments: z.number(),
  averageRiskScore: z.number(),
  riskDistribution: z.object({
    low: z.number(),
    medium: z.number(),
    high: z.number(),
  }),
  industryBreakdown: z.array(
    z.object({
      industry: z.string(),
      count: z.number(),
      averageRiskScore: z.number(),
    })
  ),
  countryBreakdown: z.array(
    z.object({
      country: z.string(),
      count: z.number(),
      averageRiskScore: z.number(),
    })
  ),
  timestamp: z.string(),
});

// Risk benchmarks schema
export const RiskBenchmarksSchema = z.object({
  industry_code: z.string(),
  industry_type: z.enum(['mcc', 'naics', 'sic']),
  average_risk_score: z.number().optional(),
  median_risk_score: z.number().optional(),
  percentile_25: z.number().optional(),
  percentile_75: z.number().optional(),
  percentile_90: z.number().optional(),
  sample_size: z.number(),
  benchmarks: z.object({
    average: z.number().optional(),
    median: z.number().optional(),
    p25: z.number().optional(),
    p75: z.number().optional(),
    p90: z.number().optional(),
  }),
});

// Assessment status response schema
export const AssessmentStatusResponseSchema = z.object({
  assessmentId: z.string(),
  merchantId: z.string(),
  status: z.string(),
  progress: z.number(),
  estimatedCompletion: z.string().optional(),
  result: RiskAssessmentResultSchema.optional(),
  completedAt: z.string().optional(),
});

// Risk history response schema
export const RiskHistoryResponseSchema = z.object({
  merchantId: z.string(),
  history: z.array(RiskAssessmentSchema),
  limit: z.number(),
  offset: z.number(),
  total: z.number(),
});

// Risk recommendations response schema
export const RiskRecommendationsResponseSchema = z.object({
  merchantId: z.string(),
  recommendations: z.array(
    z.object({
      id: z.string(),
      type: z.string(),
      priority: z.string(),
      title: z.string(),
      description: z.string(),
      actionItems: z.array(z.string()),
    })
  ),
  timestamp: z.string(),
});

// Risk indicator schema
export const RiskIndicatorSchema = z.object({
  id: z.string(),
  title: z.string(),
  description: z.string(),
  severity: z.enum(['critical', 'high', 'medium', 'low']),
  status: z.string().optional(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

// Risk indicators data schema
export const RiskIndicatorsDataSchema = z.object({
  merchantId: z.string(),
  indicators: z.array(RiskIndicatorSchema),
  timestamp: z.string().optional(),
});

// Merchant list item schema (snake_case from backend)
export const MerchantListItemSchema = z.object({
  id: z.string(),
  name: z.string(),
  legal_name: z.string().optional(),
  registration_number: z.string().optional(),
  industry: z.string().optional(),
  portfolio_type: z.string().optional(),
  risk_level: z.string().optional(),
  compliance_status: z.string().optional(),
  status: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

// Merchant list response schema
export const MerchantListResponseSchema = z.object({
  merchants: z.array(MerchantListItemSchema),
  total: z.number(),
  page: z.number(),
  page_size: z.number(),
  total_pages: z.number(),
  has_next: z.boolean(),
  has_previous: z.boolean(),
});

// Dashboard metrics schema
export const DashboardMetricsSchema = z.object({
  totalMerchants: z.number(),
  revenue: z.number(),
  growthRate: z.number(),
  analyticsScore: z.number(),
  timestamp: z.string().optional(),
});

// Risk metrics schema
export const RiskMetricsSchema = z.object({
  overallRiskScore: z.number(),
  highRiskMerchants: z.number(),
  riskAssessments: z.number(),
  riskTrend: z.number(),
  riskDistribution: z
    .object({
      low: z.number(),
      medium: z.number(),
      high: z.number(),
      critical: z.number().optional(),
    })
    .optional(),
  timestamp: z.string().optional(),
});

// System metrics schema
export const SystemMetricsSchema = z.object({
  systemHealth: z.number(),
  serverStatus: z.string(),
  databaseStatus: z.string(),
  responseTime: z.number(),
  cpuUsage: z.number().optional(),
  memoryUsage: z.number().optional(),
  diskUsage: z.number().optional(),
  timestamp: z.string().optional(),
});

// Compliance status schema
export const ComplianceStatusSchema = z.object({
  overallScore: z.number(),
  pendingReviews: z.number(),
  complianceTrend: z.string(),
  regulatoryFrameworks: z.number(),
  violations: z.number().optional(),
  timestamp: z.string().optional(),
});

/**
 * Validates API response data against a Zod schema
 * Logs validation errors in development mode
 * @param schema - Zod schema to validate against
 * @param data - Data to validate
 * @param endpoint - API endpoint name for error logging
 * @returns Validated data
 * @throws Error if validation fails
 */
export function validateAPIResponse<T>(
  schema: z.ZodSchema<T>,
  data: unknown,
  endpoint: string
): T {
  try {
    return schema.parse(data);
  } catch (error) {
    if (error instanceof z.ZodError) {
      // ZodError has an 'issues' property, not 'errors'
      const zodError = error;
      const errorDetails = {
        endpoint,
        issues: zodError.issues,
        receivedData: data,
        timestamp: new Date().toISOString(),
      };

      if (process.env.NODE_ENV === 'development') {
        console.error('[API Validation] Validation failed:', errorDetails);
        console.error('[API Validation] Zod issues:', zodError.issues);
        console.error('[API Validation] Received data:', data);
      } else {
        console.error('[API Validation] Validation failed for', endpoint);
      }

      throw new Error(
        `API response validation failed for ${endpoint}: ${zodError.issues.map((e: z.ZodIssue) => e.message).join('; ')}`
      );
    }
    throw error;
  }
}

/**
 * Type guard to check if merchant has financial data
 */
export function hasFinancialData(
  merchant: z.infer<typeof MerchantSchema>
): merchant is z.infer<typeof MerchantSchema> & {
  foundedDate: string;
  employeeCount: number;
  annualRevenue: number;
} {
  return !!(
    merchant.foundedDate &&
    merchant.employeeCount !== undefined &&
    merchant.annualRevenue !== undefined
  );
}

/**
 * Type guard to check if merchant has complete address
 */
export function hasCompleteAddress(
  merchant: z.infer<typeof MerchantSchema>
): merchant is z.infer<typeof MerchantSchema> & {
  address: NonNullable<z.infer<typeof AddressSchema>>;
} {
  return !!(
    merchant.address &&
    merchant.address.street1 &&
    merchant.address.city &&
    merchant.address.country
  );
}

/**
 * Type guard to check if risk assessment has completed result
 */
export function hasRiskAssessmentResult(
  assessment: z.infer<typeof RiskAssessmentSchema>
): assessment is z.infer<typeof RiskAssessmentSchema> & {
  result: NonNullable<z.infer<typeof RiskAssessmentResultSchema>>;
  status: 'completed';
} {
  return assessment.status === 'completed' && !!assessment.result;
}

