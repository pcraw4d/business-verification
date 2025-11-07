/**
 * Data Type Definitions
 * TypeScript definitions for shared component library
 */

export interface RiskData {
    merchantId: string;
    current: RiskAssessment | null;
    history: RiskHistory | null;
    predictions: RiskPredictions | null;
    benchmarks: IndustryBenchmarks | null;
    lastUpdated: string;
    dataSources: {
        assessment: boolean;
        history: boolean;
        predictions: boolean;
        benchmarks: boolean;
    };
}

export interface RiskAssessment {
    id: string;
    merchantId: string;
    overallScore: number;
    overallLevel: RiskLevel;
    categoryScores: Record<string, CategoryScore>;
    riskFactors: RiskFactor[];
    recommendations: Recommendation[];
    explanations?: SHAPExplanation;
    timestamp: string;
}

export interface RiskHistory {
    merchantId: string;
    timeRange: string;
    dataPoints: RiskHistoryDataPoint[];
    trends: RiskTrend[];
}

export interface RiskHistoryDataPoint {
    timestamp: string;
    overallScore: number;
    categoryScores: Record<string, number>;
    industryAverage?: number;
}

export interface RiskTrend {
    category: string;
    direction: 'improving' | 'stable' | 'rising';
    change: number;
    period: string;
}

export interface RiskPredictions {
    merchantId: string;
    horizons: number[]; // months
    predictions: PredictionDataPoint[];
    scenarios: ScenarioData | null;
    confidence: ConfidenceInterval | null;
    drivers: RiskDriver[];
}

export interface PredictionDataPoint {
    horizon: number; // months
    predictedScore: number;
    confidenceInterval: {
        lower: number;
        upper: number;
    };
    categoryPredictions: Record<string, number>;
}

export interface ScenarioData {
    optimistic: PredictionDataPoint[];
    realistic: PredictionDataPoint[];
    pessimistic: PredictionDataPoint[];
}

export interface ConfidenceInterval {
    lower: number;
    upper: number;
    confidence: number; // 0-1
}

export interface RiskDriver {
    category: string;
    factor: string;
    impact: number; // 0-1
    description: string;
}

export interface IndustryBenchmarks {
    industryCodes: {
        mcc?: string;
        naics?: string;
        sic?: string;
    };
    industryName: string;
    averages: Record<string, number>;
    percentiles: {
        p10: Record<string, number>;
        p25: Record<string, number>;
        p50: Record<string, number>;
        p75: Record<string, number>;
        p90: Record<string, number>;
    };
    sampleSize: number;
    isFallback?: boolean;
}

export type RiskLevel = 'low' | 'medium' | 'high' | 'critical';

export interface CategoryScore {
    category: string;
    score: number;
    level: RiskLevel;
    subCategories: Record<string, number>;
    trend?: TrendDirection;
    lastUpdated?: string;
}

export interface TrendDirection {
    direction: 'improving' | 'stable' | 'rising';
    change: number;
    icon: string;
    label: string;
}

export interface RiskFactor {
    id: string;
    category: string;
    name: string;
    description: string;
    score: number;
    level: RiskLevel;
    detectedAt?: string;
}

export interface Recommendation {
    id: string;
    type: 'ml_based' | 'manual_verification' | 'document_verification' | 'compliance_check' | 'security_audit';
    priority: 'critical' | 'high' | 'medium' | 'low';
    title: string;
    description: string;
    impactScore: number; // 0-1
    difficulty: 'low' | 'medium' | 'high';
    actionRequired: string;
    status: 'pending' | 'in_progress' | 'completed';
}

export interface SHAPExplanation {
    featureContributions: Array<{
        feature: string;
        contribution: number;
        value: number;
    }>;
    baseValue: number;
    prediction: number;
}

export interface MerchantData {
    id: string;
    name: string;
    email?: string;
    phone?: string;
    website?: string;
    address?: Address;
    classification?: MerchantClassification;
    analytics?: MerchantAnalytics;
    riskSummary?: RiskSummary;
    createdAt?: string;
    updatedAt?: string;
}

export interface Address {
    street?: string;
    city?: string;
    state?: string;
    zip?: string;
    country?: string;
}

export interface MerchantClassification {
    mcc_codes?: Array<{ code: string; description: string; confidence: number }>;
    naics_codes?: Array<{ code: string; description: string; confidence: number }>;
    sic_codes?: Array<{ code: string; description: string; confidence: number }>;
    primary_industry?: string;
    confidence?: number;
}

export interface MerchantAnalytics {
    id: string;
    merchantId: string;
    dataQuality: DataQualityMetrics;
    risk_keywords?: Array<{ keyword: string; severity: string; context: string }>;
    sentiment_analysis?: SentimentAnalysis;
    backlinks?: BacklinkData[];
    createdAt?: string;
}

export interface DataQualityMetrics {
    completeness: number; // 0-1
    consistency: number; // 0-1
    agreement: number; // 0-1
    overall: number; // 0-1
}

export interface SentimentAnalysis {
    overall_sentiment: 'positive' | 'neutral' | 'negative';
    score: number; // -1 to 1
    confidence: number; // 0-1
}

export interface BacklinkData {
    domain: string;
    quality: 'high' | 'medium' | 'low';
    type: string;
}

export interface RiskSummary {
    overallScore: number;
    overallLevel: RiskLevel;
    categoryScores: Record<string, number>;
}

export interface ComplianceData {
    status: ComplianceStatus;
    gaps?: ComplianceGap[];
    progress?: ComplianceProgress;
    alerts?: ComplianceAlert[];
    lastUpdated: string;
}

export interface ComplianceStatus {
    merchantId?: string;
    frameworks: ComplianceFramework[];
    overallScore: number;
    compliantCount: number;
    nonCompliantCount: number;
    inProgressCount: number;
}

export interface ComplianceFramework {
    name: string; // SOC2, PCI_DSS, GDPR, etc.
    status: 'compliant' | 'non_compliant' | 'in_progress';
    score: number; // 0-100
    lastAssessed: string;
}

export interface ComplianceGap {
    id: string;
    framework: string;
    requirement: string;
    severity: 'critical' | 'high' | 'medium' | 'low';
    description: string;
    remediation: string;
    priority: number;
}

export interface ComplianceProgress {
    merchantId?: string;
    milestones: ProgressMilestone[];
    completionRate: number; // 0-1
    velocity: number; // milestones per week
}

export interface ProgressMilestone {
    id: string;
    name: string;
    targetDate: string;
    status: 'pending' | 'in_progress' | 'completed';
    completionDate?: string;
}

export interface ComplianceAlert {
    id: string;
    type: 'gap' | 'deadline' | 'violation' | 'review';
    severity: 'critical' | 'high' | 'medium' | 'low';
    title: string;
    message: string;
    framework?: string;
    merchantId?: string;
    createdAt: string;
    acknowledgedAt?: string;
    resolvedAt?: string;
}

