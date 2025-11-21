// Dashboard and Metrics Types

export interface DashboardMetrics {
  totalMerchants: number;
  revenue: number;
  growthRate: number;
  analyticsScore: number;
  timestamp?: string;
}

export interface RiskMetrics {
  overallRiskScore: number;
  highRiskMerchants: number;
  riskAssessments: number;
  riskTrend: number;
  riskDistribution?: {
    low: number;
    medium: number;
    high: number;
    critical?: number;
  };
  timestamp?: string;
}

export interface SystemMetrics {
  systemHealth: number;
  serverStatus: string;
  databaseStatus: string;
  responseTime: number;
  cpuUsage?: number;
  memoryUsage?: number;
  diskUsage?: number;
  timestamp?: string;
}

export interface ComplianceStatus {
  overallScore: number;
  pendingReviews: number;
  complianceTrend: string;
  regulatoryFrameworks: number;
  violations?: number;
  timestamp?: string;
}

export interface BusinessIntelligenceMetrics {
  revenueGrowth: number;
  marketShare: number;
  performanceScore: number;
  analyticsScore: number;
  timestamp?: string;
}

