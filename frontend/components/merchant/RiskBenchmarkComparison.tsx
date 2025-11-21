'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { AlertCircle, TrendingUp, TrendingDown, BarChart3, RefreshCw, Sparkles } from 'lucide-react';
import { useEffect, useState, useCallback, useRef } from 'react';
import { getMerchantAnalytics, getRiskBenchmarks, getMerchantRiskScore } from '@/lib/api';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import type { BenchmarkComparison, RiskBenchmarks, MerchantRiskScore, AnalyticsData } from '@/types/merchant';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { BarChart } from '@/components/charts/lazy';
import { formatPercent, formatPercentile, formatPercentWithSign } from '@/lib/number-format';
import { EnrichmentButton } from './EnrichmentButton';
import { ErrorCodes, formatErrorWithCode } from '@/lib/error-codes';

interface RiskBenchmarkComparisonProps {
  merchantId: string;
}

export function RiskBenchmarkComparison({ merchantId }: RiskBenchmarkComparisonProps) {
  const [comparison, setComparison] = useState<BenchmarkComparison | null>(null);
  const [industryCode, setIndustryCode] = useState<{ code: string; type: 'mcc' | 'naics' | 'sic'; description: string } | null>(null);
  const [benchmarks, setBenchmarks] = useState<RiskBenchmarks | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastRefreshTime, setLastRefreshTime] = useState<Date | null>(null);
  const fetchingRef = useRef(false); // Track if fetch is in progress to prevent infinite loops

  const fetchComparisonData = useCallback(async (bypassCache = false) => {
    // Prevent concurrent fetches
    if (fetchingRef.current) {
      return;
    }
    fetchingRef.current = true;
    if (!bypassCache) {
      setLoading(true);
    } else {
      setIsRefreshing(true);
    }
    setError(null);
    try {
      // Step 1: Fetch merchant analytics to get industry codes
      const analyticsResult = await Promise.allSettled([
        getMerchantAnalytics(merchantId),
      ]);

      let analytics: AnalyticsData | null = null;
      if (analyticsResult[0].status === 'fulfilled') {
        analytics = analyticsResult[0].value;
        
        if (process.env.NODE_ENV === 'development') {
          console.log('[RiskBenchmarkComparison] Merchant analytics loaded:', {
            hasClassification: !!analytics.classification,
            mccCodes: analytics.classification?.mccCodes?.length || 0,
            naicsCodes: analytics.classification?.naicsCodes?.length || 0,
            sicCodes: analytics.classification?.sicCodes?.length || 0,
          });
        }
      } else {
        const reason = analyticsResult[0].reason;
        if (process.env.NODE_ENV === 'development') {
          console.error('[RiskBenchmarkComparison] Failed to fetch merchant analytics:', reason);
        }
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
        setError(formatErrorWithCode(
          'Industry code is required for benchmark comparison. Use the Enrich Data button to add industry information.',
          ErrorCodes.RISK_BENCHMARK.MISSING_INDUSTRY_CODE
        ));
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
        
        if (process.env.NODE_ENV === 'development') {
          console.log('[RiskBenchmarkComparison] Benchmarks loaded:', {
            average_risk_score: benchmarks.average_risk_score,
            median_risk_score: benchmarks.median_risk_score,
            percentile_25: benchmarks.percentile_25,
            percentile_75: benchmarks.percentile_75,
            percentile_90: benchmarks.percentile_90,
          });
        }
      } else {
        const reason = benchmarksResult.reason;
        if (process.env.NODE_ENV === 'development') {
          console.error('[RiskBenchmarkComparison] Failed to fetch risk benchmarks:', reason);
        }
        setError(formatErrorWithCode(
          'Benchmark data for this industry is currently unavailable. Please try again later.',
          ErrorCodes.RISK_BENCHMARK.BENCHMARKS_UNAVAILABLE
        ));
      }

      let merchantRiskScore: MerchantRiskScore | null = null;
      if (merchantRiskScoreResult.status === 'fulfilled') {
        merchantRiskScore = merchantRiskScoreResult.value;
        
        if (process.env.NODE_ENV === 'development') {
          console.log('[RiskBenchmarkComparison] Merchant risk score loaded:', {
            risk_score: merchantRiskScore.risk_score,
            risk_level: merchantRiskScore.risk_level,
          });
        }
      } else {
        const reason = merchantRiskScoreResult.reason;
        if (process.env.NODE_ENV === 'development') {
          console.error('[RiskBenchmarkComparison] Failed to fetch merchant risk score:', reason);
        }
        // Use functional update to avoid needing error in dependency array
        setError((prev) => {
          // Only set error if we don't already have a more specific error
          if (prev && prev.includes('Error RB-')) {
            return prev; // Keep existing error
          }
          return formatErrorWithCode(
            'Unable to fetch merchant risk score. Please ensure a risk assessment has been completed.',
            ErrorCodes.RISK_BENCHMARK.MISSING_RISK_SCORE
          );
        });
      }

      if (benchmarks && merchantRiskScore) {
        const merchantScore = merchantRiskScore.risk_score; // Already 0-1 scale
        const industryAverage = benchmarks.average_risk_score;
        const industryMedian = benchmarks.median_risk_score;
        const industryP25 = benchmarks.percentile_25;
        const industryP75 = benchmarks.percentile_75;
        const industryP90 = benchmarks.percentile_90;

        // Validate required values and provide specific error messages with codes
        if (merchantScore == null) {
          setError(formatErrorWithCode(
            'Merchant risk score is required for benchmark comparison. Please ensure a risk assessment has been completed.',
            ErrorCodes.RISK_BENCHMARK.MISSING_RISK_SCORE
          ));
          setLoading(false);
          return;
        }
        
        if (industryAverage == null || industryMedian == null || 
            industryP25 == null || industryP75 == null || industryP90 == null) {
          setError(formatErrorWithCode(
            'Incomplete benchmark data. Cannot perform comparison. Please try again later.',
            ErrorCodes.RISK_BENCHMARK.INVALID_DATA
          ));
          setLoading(false);
          return;
        }

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
        
        // Update last refresh time
        setLastRefreshTime(new Date());
      } else {
        // If we don't have both benchmarks and merchantRiskScore, but no specific error was set,
        // set a generic error. We check if error is already set by checking if we've set it earlier.
        // Since we can't read error state here (would cause dependency issues), we rely on the
        // fact that specific errors are set above, so this is a fallback.
        setError((prev) => {
          // Only set generic error if no specific error was already set
          if (prev && prev.includes('Error RB-')) {
            return prev; // Keep existing specific error
          }
          return formatErrorWithCode(
            'Not enough data to perform benchmark comparison.',
            ErrorCodes.RISK_BENCHMARK.INVALID_DATA
          );
        });
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An unexpected error occurred.';
      setError(errorMessage);
      toast.error('Error loading benchmark comparison', { description: errorMessage });
    } finally {
      setLoading(false);
      setIsRefreshing(false);
      fetchingRef.current = false; // Reset fetch flag
    }
  }, [merchantId]); // Removed 'error' from dependencies to prevent infinite loop

  // Format relative time for last refresh
  const formatRelativeTime = (date: Date): string => {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffSecs < 60) return 'just now';
    if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
    if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
    if (diffDays < 7) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
    return date.toLocaleDateString();
  };

  const handleRefresh = () => {
    fetchComparisonData(true);
  };

  useEffect(() => {
    fetchComparisonData();
  }, [fetchComparisonData]);

  // Keyboard shortcut: R to refresh
  useKeyboardShortcuts([
    {
      key: 'r',
      handler: handleRefresh,
      description: 'Refresh benchmark comparison data',
    },
  ]);

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
    return (
      <Card>
        <CardHeader>
          <CardTitle>Industry Benchmark Comparison</CardTitle>
          <CardDescription>Fetching industry benchmarks...</CardDescription>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-96 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    const isMissingIndustryCode = error.includes('Industry code is required');
    
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive">Error Loading Benchmark Comparison</CardTitle>
          <CardDescription>Could not load industry benchmark comparison data.</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription className="mt-2">
              {error.includes('Error ') ? error : formatErrorWithCode(error, ErrorCodes.RISK_BENCHMARK.FETCH_ERROR)}
              <div className="mt-3 space-y-2">
                {isMissingIndustryCode ? (
                  <EnrichmentButton merchantId={merchantId} variant="default" size="sm" />
                ) : (
                  <Button 
                    onClick={fetchComparisonData} 
                    className="w-full sm:w-auto"
                    variant="outline"
                  >
                    <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
                )}
              </div>
            </AlertDescription>
          </Alert>
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
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Insufficient Data</AlertTitle>
            <AlertDescription className="mt-2">
            We could not retrieve sufficient data to compare this merchant against industry benchmarks.
              <div className="mt-3 space-y-2">
                <Button 
                  onClick={fetchComparisonData} 
                  className="w-full sm:w-auto"
                  variant="outline"
                >
                  <RefreshCw className="h-4 w-4 mr-2" />
            Reload
          </Button>
              </div>
            </AlertDescription>
          </Alert>
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
    <Card role="region" aria-labelledby="benchmark-heading" aria-describedby="benchmark-description">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle id="benchmark-heading">Industry Benchmark Comparison</CardTitle>
            <CardDescription id="benchmark-description">
              How this merchant compares to industry benchmarks ({industryCode.type.toUpperCase()}: {industryCode.code} - {industryCode.description})
            </CardDescription>
            {lastRefreshTime && (
              <p className="text-xs text-muted-foreground mt-1" aria-live="polite">
                Updated {formatRelativeTime(lastRefreshTime)}
              </p>
            )}
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={handleRefresh}
            disabled={isRefreshing || loading}
            aria-label="Refresh benchmark comparison data"
            title="Refresh data (R)"
          >
            <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Key Metrics */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Merchant Risk Score</p>
            <p className="text-2xl font-bold">{formatPercent(comparison.merchantScore)}</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Industry Average</p>
            <p className="text-2xl font-bold">{formatPercent(comparison.industryAverage)}</p>
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
                {formatPercentWithSign(comparison.difference)}
              </span>
            </div>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Percentile</p>
            <p className="text-2xl font-bold">{formatPercentile(comparison.percentile)}</p>
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
          <div role="img" aria-label="Bar chart showing merchant risk score compared to industry benchmarks">
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
          </div>
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
                <p className="text-lg font-semibold">{formatPercent(benchmarks.percentile_25)}</p>
              </div>
              <div className="p-3 border rounded-lg">
                <p className="text-xs text-muted-foreground">Median</p>
                <p className="text-lg font-semibold">{formatPercent(benchmarks.median_risk_score)}</p>
              </div>
              <div className="p-3 border rounded-lg">
                <p className="text-xs text-muted-foreground">75th Percentile</p>
                <p className="text-lg font-semibold">{formatPercent(benchmarks.percentile_75)}</p>
              </div>
              <div className="p-3 border rounded-lg">
                <p className="text-xs text-muted-foreground">90th Percentile</p>
                <p className="text-lg font-semibold">{formatPercent(benchmarks.percentile_90)}</p>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

