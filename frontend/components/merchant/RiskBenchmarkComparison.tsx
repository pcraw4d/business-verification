'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { AlertCircle, TrendingUp, TrendingDown, BarChart3 } from 'lucide-react';
import { useEffect, useState, useCallback } from 'react';
import { getMerchantAnalytics, getRiskBenchmarks, getMerchantRiskScore } from '@/lib/api';
import type { BenchmarkComparison, RiskBenchmarks, MerchantRiskScore, AnalyticsData } from '@/types/merchant';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { BarChart } from '@/components/charts/lazy';

interface RiskBenchmarkComparisonProps {
  merchantId: string;
}

export function RiskBenchmarkComparison({ merchantId }: RiskBenchmarkComparisonProps) {
  const [comparison, setComparison] = useState<BenchmarkComparison | null>(null);
  const [industryCode, setIndustryCode] = useState<{ code: string; type: 'mcc' | 'naics' | 'sic'; description: string } | null>(null);
  const [benchmarks, setBenchmarks] = useState<RiskBenchmarks | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchComparisonData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // Step 1: Fetch merchant analytics to get industry codes
      const analyticsResult = await Promise.allSettled([
        getMerchantAnalytics(merchantId),
      ]);

      let analytics: AnalyticsData | null = null;
      if (analyticsResult[0].status === 'fulfilled') {
        analytics = analyticsResult[0].value;
      } else {
        console.error('Failed to fetch merchant analytics:', analyticsResult[0].reason);
      }

      // Step 2: Extract industry code (prefer MCC, then NAICS, then SIC)
      let extractedCode: { code: string; type: 'mcc' | 'naics' | 'sic'; description: string } | null = null;
      
      if (analytics?.classification) {
        // Try MCC first
        if (analytics.classification.mccCodes && analytics.classification.mccCodes.length > 0) {
          const topMcc = analytics.classification.mccCodes[0]; // Already sorted by confidence
          extractedCode = {
            code: topMcc.code,
            type: 'mcc',
            description: topMcc.description,
          };
        }
        // Try NAICS second
        else if (analytics.classification.naicsCodes && analytics.classification.naicsCodes.length > 0) {
          const topNaics = analytics.classification.naicsCodes[0];
          extractedCode = {
            code: topNaics.code,
            type: 'naics',
            description: topNaics.description,
          };
        }
        // Try SIC third
        else if (analytics.classification.sicCodes && analytics.classification.sicCodes.length > 0) {
          const topSic = analytics.classification.sicCodes[0];
          extractedCode = {
            code: topSic.code,
            type: 'sic',
            description: topSic.description,
          };
        }
      }

      if (!extractedCode) {
        setError('No industry code available for this merchant. Cannot perform benchmark comparison.');
        setLoading(false);
        return;
      }

      setIndustryCode(extractedCode);

      // Step 3: Fetch benchmarks and merchant risk score in parallel
      const [benchmarksResult, merchantRiskScoreResult] = await Promise.allSettled([
        getRiskBenchmarks({ [extractedCode.type]: extractedCode.code }),
        getMerchantRiskScore(merchantId),
      ]);

      let benchmarks: RiskBenchmarks | null = null;
      if (benchmarksResult.status === 'fulfilled') {
        benchmarks = benchmarksResult.value;
      } else {
        console.error('Failed to fetch risk benchmarks:', benchmarksResult.reason);
        setError('Failed to load industry benchmarks.');
      }

      let merchantRiskScore: MerchantRiskScore | null = null;
      if (merchantRiskScoreResult.status === 'fulfilled') {
        merchantRiskScore = merchantRiskScoreResult.value;
      } else {
        console.error('Failed to fetch merchant risk score:', merchantRiskScoreResult.reason);
        setError((prev) => (prev ? prev + ' ' : '') + 'Failed to load merchant risk score.');
      }

      if (benchmarks && merchantRiskScore) {
        const merchantScore = merchantRiskScore.risk_score; // Already 0-1 scale
        const industryAverage = benchmarks.average_risk_score;
        const industryMedian = benchmarks.median_risk_score;
        const industryP25 = benchmarks.percentile_25;
        const industryP75 = benchmarks.percentile_75;
        const industryP90 = benchmarks.percentile_90;

        const difference = merchantScore - industryAverage;
        const differencePercentage = industryAverage !== 0 ? (difference / industryAverage) * 100 : 0;

        // Calculate percentile position
        let percentile = 50; // Default to median
        if (merchantScore <= industryP25) {
          // Bottom 25%
          percentile = (merchantScore / industryP25) * 25;
        } else if (merchantScore <= industryMedian) {
          // 25th to 50th percentile
          percentile = 25 + ((merchantScore - industryP25) / (industryMedian - industryP25)) * 25;
        } else if (merchantScore <= industryP75) {
          // 50th to 75th percentile
          percentile = 50 + ((merchantScore - industryMedian) / (industryP75 - industryMedian)) * 25;
        } else if (merchantScore <= industryP90) {
          // 75th to 90th percentile
          percentile = 75 + ((merchantScore - industryP75) / (industryP90 - industryP75)) * 15;
        } else {
          // Top 10%
          percentile = 90 + ((merchantScore - industryP90) / (1 - industryP90)) * 10;
        }
        percentile = Math.min(100, Math.max(0, percentile)); // Ensure between 0 and 100

        // Determine position
        let position: BenchmarkComparison['position'] = 'average';
        if (percentile >= 90) {
          position = 'top_10';
        } else if (percentile >= 75) {
          position = 'top_25';
        } else if (percentile >= 25) {
          position = 'average';
        } else if (percentile >= 10) {
          position = 'bottom_25';
        } else {
          position = 'bottom_10';
        }

        setComparison({
          merchantScore,
          industryAverage,
          industryMedian,
          industryPercentile75: industryP75,
          industryPercentile90: industryP90,
          percentile,
          position,
          difference,
          differencePercentage,
        });
        
        // Store benchmarks for later use in chart
        setBenchmarks(benchmarks);
      } else if (!error) {
        setError('Not enough data to perform benchmark comparison.');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An unexpected error occurred.';
      setError(errorMessage);
      toast.error('Error loading benchmark comparison', { description: errorMessage });
    } finally {
      setLoading(false);
    }
  }, [merchantId, error]);

  useEffect(() => {
    fetchComparisonData();
  }, [fetchComparisonData]);

  const getPositionBadgeVariant = (position: BenchmarkComparison['position']) => {
    switch (position) {
      case 'top_10':
      case 'top_25':
        return 'default'; // Lower risk is better
      case 'average':
        return 'secondary';
      case 'bottom_25':
      case 'bottom_10':
        return 'destructive'; // Higher risk is worse
      default:
        return 'outline';
    }
  };

  const getPositionIcon = (position: BenchmarkComparison['position']) => {
    switch (position) {
      case 'top_10':
      case 'top_25':
        return <TrendingDown className="h-4 w-4 text-green-500" />; // Lower risk is better
      case 'bottom_25':
      case 'bottom_10':
        return <TrendingUp className="h-4 w-4 text-red-500" />; // Higher risk is worse
      case 'average':
      default:
        return <BarChart3 className="h-4 w-4 text-gray-500" />;
    }
  };

  const getPositionLabel = (position: BenchmarkComparison['position']) => {
    switch (position) {
      case 'top_10':
        return 'Top 10% (Lowest Risk)';
      case 'top_25':
        return 'Top 25% (Low Risk)';
      case 'average':
        return 'Average';
      case 'bottom_25':
        return 'Bottom 25% (High Risk)';
      case 'bottom_10':
        return 'Bottom 10% (Highest Risk)';
      default:
        return 'Unknown';
    }
  };

  if (loading) {
    return <Skeleton className="h-96 w-full" />;
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive">Error Loading Benchmark Comparison</CardTitle>
          <CardDescription>Could not load industry benchmark comparison data.</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-destructive-foreground">{error}</p>
          <Button onClick={fetchComparisonData} className="mt-4" variant="outline">
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (!comparison || !industryCode) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Industry Benchmark Comparison</CardTitle>
          <CardDescription>No benchmark comparison data available.</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            We could not retrieve sufficient data to compare this merchant against industry benchmarks.
          </p>
          <Button onClick={fetchComparisonData} className="mt-4" variant="outline">
            Reload
          </Button>
        </CardContent>
      </Card>
    );
  }

  // Prepare chart data for comparison
  const chartData = [
    {
      name: 'Merchant',
      value: comparison.merchantScore,
      isMerchant: 1, // Use number instead of boolean for chart compatibility
    },
    {
      name: 'Industry Average',
      value: comparison.industryAverage,
      isMerchant: 0, // Use number instead of boolean for chart compatibility
    },
    {
      name: 'Industry Median',
      value: comparison.industryMedian,
      isMerchant: 0, // Use number instead of boolean for chart compatibility
    },
    {
      name: '75th Percentile',
      value: comparison.industryPercentile75,
      isMerchant: 0, // Use number instead of boolean for chart compatibility
    },
    {
      name: '90th Percentile',
      value: comparison.industryPercentile90,
      isMerchant: 0, // Use number instead of boolean for chart compatibility
    },
  ];
  
  // Determine merchant bar color based on position
  const merchantBarColor = 
    comparison.position === 'top_10' || comparison.position === 'top_25'
      ? '#22c55e' // Green for low risk
      : comparison.position === 'bottom_25' || comparison.position === 'bottom_10'
      ? '#ef4444' // Red for high risk
      : '#3b82f6'; // Blue for average

  return (
    <Card>
      <CardHeader>
        <CardTitle>Industry Benchmark Comparison</CardTitle>
        <CardDescription>
          How this merchant compares to industry benchmarks ({industryCode.type.toUpperCase()}: {industryCode.code} - {industryCode.description})
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Key Metrics */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Merchant Risk Score</p>
            <p className="text-2xl font-bold">{(comparison.merchantScore * 100).toFixed(1)}%</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Industry Average</p>
            <p className="text-2xl font-bold">{(comparison.industryAverage * 100).toFixed(1)}%</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Difference</p>
            <div className="flex items-center space-x-1">
              {getPositionIcon(comparison.position)}
              <span
                className={`text-2xl font-bold ${
                  comparison.difference < 0
                    ? 'text-green-500'
                    : comparison.difference > 0
                    ? 'text-red-500'
                    : 'text-gray-500'
                }`}
              >
                {comparison.difference > 0 ? '+' : ''}
                {(comparison.difference * 100).toFixed(1)}%
              </span>
            </div>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Percentile</p>
            <p className="text-2xl font-bold">{comparison.percentile.toFixed(0)}th</p>
          </div>
        </div>

        {/* Position Badge */}
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium text-muted-foreground">Industry Position</p>
          <Badge variant={getPositionBadgeVariant(comparison.position)}>
            {getPositionIcon(comparison.position)}
            <span className="ml-1">{getPositionLabel(comparison.position)}</span>
          </Badge>
        </div>

        {/* Comparison Chart */}
        <ChartContainer
          title="Risk Score Comparison"
          description="Merchant risk score compared to industry benchmarks"
          isLoading={false}
        >
          <BarChart
            data={chartData}
            dataKey="name"
            bars={[
              {
                key: 'value',
                name: 'Risk Score',
                color: '#8884d8', // Default color - merchant will be highlighted by label
              },
            ]}
            height={300}
            isLoading={false}
          />
          <div className="mt-4 flex items-center justify-center gap-4 text-sm">
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded" style={{ backgroundColor: merchantBarColor }} />
              <span className="font-medium">Merchant ({comparison.position.replace(/_/g, ' ')})</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded bg-gray-400" />
              <span>Industry Benchmarks</span>
            </div>
          </div>
        </ChartContainer>

        {/* Detailed Benchmarks */}
        {benchmarks && (
          <div className="space-y-2">
            <p className="text-sm font-medium text-muted-foreground">Industry Benchmarks</p>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="p-3 border rounded-lg">
                <p className="text-xs text-muted-foreground">25th Percentile</p>
                <p className="text-lg font-semibold">{(benchmarks.percentile_25 * 100).toFixed(1)}%</p>
              </div>
              <div className="p-3 border rounded-lg">
                <p className="text-xs text-muted-foreground">Median</p>
                <p className="text-lg font-semibold">{(benchmarks.median_risk_score * 100).toFixed(1)}%</p>
              </div>
              <div className="p-3 border rounded-lg">
                <p className="text-xs text-muted-foreground">75th Percentile</p>
                <p className="text-lg font-semibold">{(benchmarks.percentile_75 * 100).toFixed(1)}%</p>
              </div>
              <div className="p-3 border rounded-lg">
                <p className="text-xs text-muted-foreground">90th Percentile</p>
                <p className="text-lg font-semibold">{(benchmarks.percentile_90 * 100).toFixed(1)}%</p>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

