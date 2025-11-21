'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { TrendingUp, TrendingDown, BarChart3, AlertCircle, RefreshCw } from 'lucide-react';
import { useEffect, useState, useCallback } from 'react';
import { getMerchantAnalytics, getPortfolioAnalytics } from '@/lib/api';
import type { AnalyticsComparison, AnalyticsData, PortfolioAnalytics } from '@/types/merchant';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { BarChart } from '@/components/charts/lazy';
import { formatPercent, formatNumber } from '@/lib/number-format';
import { ErrorCodes, formatErrorWithCode } from '@/lib/error-codes';

interface AnalyticsComparisonProps {
  merchantId: string;
  merchantAnalytics?: AnalyticsData | null; // Optional - can be passed from parent to avoid duplicate fetch
}

export function AnalyticsComparison({ merchantId, merchantAnalytics: providedAnalytics }: AnalyticsComparisonProps) {
  const [comparison, setComparison] = useState<AnalyticsComparison | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [merchantAnalytics, setMerchantAnalytics] = useState<AnalyticsData | null>(providedAnalytics || null);

  const fetchComparisonData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // Step 1: Fetch merchant analytics if not provided
      let analytics: AnalyticsData | null = merchantAnalytics;
      if (!analytics) {
        const analyticsResult = await Promise.allSettled([
          getMerchantAnalytics(merchantId),
        ]);

        if (analyticsResult[0].status === 'fulfilled') {
          analytics = analyticsResult[0].value;
          setMerchantAnalytics(analytics);
          
          if (process.env.NODE_ENV === 'development') {
            console.log('[AnalyticsComparison] Merchant analytics loaded:', {
              hasClassification: !!analytics.classification,
              hasSecurity: !!analytics.security,
              hasQuality: !!analytics.quality,
              classificationConfidence: analytics.classification?.confidenceScore,
              securityTrustScore: analytics.security?.trustScore,
              dataQuality: analytics.quality?.completenessScore,
            });
          }
        } else {
          const reason = analyticsResult[0].reason;
          if (process.env.NODE_ENV === 'development') {
            console.error('[AnalyticsComparison] Failed to fetch merchant analytics:', reason);
          }
          setError(formatErrorWithCode(
            'Unable to fetch merchant analytics. The analytics service may be temporarily unavailable.',
            ErrorCodes.ANALYTICS_COMPARISON.MISSING_MERCHANT_ANALYTICS
          ));
        }
      }

      // Step 2: Fetch portfolio analytics
      const portfolioResult = await Promise.allSettled([
        getPortfolioAnalytics(),
      ]);

      let portfolioAnalytics: PortfolioAnalytics | null = null;
      if (portfolioResult[0].status === 'fulfilled') {
        portfolioAnalytics = portfolioResult[0].value;
        
        if (process.env.NODE_ENV === 'development') {
          console.log('[AnalyticsComparison] Portfolio analytics loaded:', {
            averageClassificationConfidence: portfolioAnalytics.averageClassificationConfidence,
            averageSecurityTrustScore: portfolioAnalytics.averageSecurityTrustScore,
            averageDataQuality: portfolioAnalytics.averageDataQuality,
            totalMerchants: portfolioAnalytics.totalMerchants,
          });
        }
      } else {
        const reason = portfolioResult[0].reason;
        if (process.env.NODE_ENV === 'development') {
          console.error('[AnalyticsComparison] Failed to fetch portfolio analytics:', reason);
        }
        setError((prev) => {
          const baseMsg = prev || 'Unable to fetch portfolio analytics.';
          return formatErrorWithCode(
            `${baseMsg} The statistics service may be temporarily unavailable.`,
            ErrorCodes.ANALYTICS_COMPARISON.MISSING_PORTFOLIO_ANALYTICS
          );
        });
      }

      // Handle partial data scenarios
      if (analytics && portfolioAnalytics) {
        // Validate and extract values with type checking
        const merchantClassificationConfidence = typeof analytics.classification?.confidenceScore === 'number' 
          ? analytics.classification.confidenceScore 
          : 0;
        const merchantSecurityTrustScore = typeof analytics.security?.trustScore === 'number'
          ? analytics.security.trustScore
          : 0;
        const merchantDataQuality = typeof analytics.quality?.completenessScore === 'number'
          ? analytics.quality.completenessScore
          : 0;

        const portfolioClassificationConfidence = typeof portfolioAnalytics.averageClassificationConfidence === 'number'
          ? portfolioAnalytics.averageClassificationConfidence
          : 0;
        const portfolioSecurityTrustScore = typeof portfolioAnalytics.averageSecurityTrustScore === 'number'
          ? portfolioAnalytics.averageSecurityTrustScore
          : 0;
        const portfolioDataQuality = typeof portfolioAnalytics.averageDataQuality === 'number'
          ? portfolioAnalytics.averageDataQuality
          : 0;

        if (process.env.NODE_ENV === 'development') {
          console.log('[AnalyticsComparison] Comparison values:', {
            merchant: {
              classificationConfidence: merchantClassificationConfidence,
              securityTrustScore: merchantSecurityTrustScore,
              dataQuality: merchantDataQuality,
            },
            portfolio: {
              classificationConfidence: portfolioClassificationConfidence,
              securityTrustScore: portfolioSecurityTrustScore,
              dataQuality: portfolioDataQuality,
            },
          });
        }

        const classificationDiff = merchantClassificationConfidence - portfolioClassificationConfidence;
        const securityDiff = merchantSecurityTrustScore - portfolioSecurityTrustScore;
        const dataQualityDiff = merchantDataQuality - portfolioDataQuality;

        const classificationPercentage = portfolioClassificationConfidence !== 0
          ? (classificationDiff / portfolioClassificationConfidence) * 100
          : 0;
        const securityPercentage = portfolioSecurityTrustScore !== 0
          ? (securityDiff / portfolioSecurityTrustScore) * 100
          : 0;
        const dataQualityPercentage = portfolioDataQuality !== 0
          ? (dataQualityDiff / portfolioDataQuality) * 100
          : 0;

        setComparison({
          merchant: {
            classificationConfidence: merchantClassificationConfidence,
            securityTrustScore: merchantSecurityTrustScore,
            dataQuality: merchantDataQuality,
          },
          portfolio: {
            averageClassificationConfidence: portfolioClassificationConfidence,
            averageSecurityTrustScore: portfolioSecurityTrustScore,
            averageDataQuality: portfolioDataQuality,
          },
          differences: {
            classificationConfidence: classificationDiff,
            securityTrustScore: securityDiff,
            dataQuality: dataQualityDiff,
          },
          percentages: {
            classificationConfidence: classificationPercentage,
            securityTrustScore: securityPercentage,
            dataQuality: dataQualityPercentage,
          },
        });
      } else if (!error) {
        // Provide specific error message based on what's missing
        if (!analytics && !portfolioAnalytics) {
          setError(formatErrorWithCode(
            'Unable to fetch merchant analytics and portfolio analytics. Please try again later.',
            ErrorCodes.ANALYTICS_COMPARISON.MISSING_BOTH
          ));
        } else if (!analytics) {
          setError(formatErrorWithCode(
            'Merchant analytics data is not available. Analytics may still be processing for this merchant.',
            ErrorCodes.ANALYTICS_COMPARISON.MISSING_MERCHANT_ANALYTICS
          ));
        } else if (!portfolioAnalytics) {
          setError(formatErrorWithCode(
            'Portfolio analytics are being calculated. Please try again in a few moments.',
            ErrorCodes.ANALYTICS_COMPARISON.MISSING_PORTFOLIO_ANALYTICS
          ));
        } else {
          setError(formatErrorWithCode(
            'Not enough data to perform analytics comparison.',
            ErrorCodes.ANALYTICS_COMPARISON.INVALID_DATA
          ));
        }
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An unexpected error occurred.';
      setError(errorMessage);
      toast.error('Error loading analytics comparison', { description: errorMessage });
    } finally {
      setLoading(false);
    }
  }, [merchantId, merchantAnalytics]);

  useEffect(() => {
    fetchComparisonData();
  }, [fetchComparisonData]);

  const getDifferenceIcon = (difference: number) => {
    if (difference > 0.01) {
      return <TrendingUp className="h-4 w-4 text-green-500" />;
    } else if (difference < -0.01) {
      return <TrendingDown className="h-4 w-4 text-red-500" />;
    } else {
      return <BarChart3 className="h-4 w-4 text-gray-500" />;
    }
  };

  const getDifferenceColor = (difference: number) => {
    if (difference > 0.01) {
      return 'text-green-500';
    } else if (difference < -0.01) {
      return 'text-red-500';
    } else {
      return 'text-gray-500';
    }
  };

  const getDifferenceLabel = (difference: number, percentage: number) => {
    if (Math.abs(difference) < 0.01) {
      return 'Similar';
    }
    // percentage is already a percentage value (0-100), not a decimal (0-1)
    // Use formatNumber for safe formatting, then add sign and % symbol
    const sign = difference > 0 ? '+' : '';
    const formatted = formatNumber(percentage, 1, '0.0');
    return `${sign}${formatted}%`;
  };

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Analytics Comparison</CardTitle>
          <CardDescription>Loading portfolio comparison...</CardDescription>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-96 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive">Error Loading Analytics Comparison</CardTitle>
          <CardDescription>Could not load analytics comparison data.</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription className="mt-2">
              {error.includes('Error ') ? error : formatErrorWithCode(error, ErrorCodes.ANALYTICS_COMPARISON.FETCH_ERROR)}
              <Button 
                onClick={fetchComparisonData} 
                className="mt-3 w-full sm:w-auto"
                variant="outline"
              >
                <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  if (!comparison) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Analytics Comparison</CardTitle>
          <CardDescription>No comparison data available.</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Insufficient Data</AlertTitle>
            <AlertDescription className="mt-2">
            We could not retrieve sufficient data to compare this merchant against portfolio analytics.
              <Button 
                onClick={fetchComparisonData} 
                className="mt-3 w-full sm:w-auto"
                variant="outline"
              >
                <RefreshCw className="h-4 w-4 mr-2" />
            Reload
          </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  // Prepare chart data for comparison
  const classificationChartData = [
    {
      name: 'Merchant',
      value: comparison.merchant.classificationConfidence * 100,
    },
    {
      name: 'Portfolio Average',
      value: comparison.portfolio.averageClassificationConfidence * 100,
    },
  ];

  const securityChartData = [
    {
      name: 'Merchant',
      value: comparison.merchant.securityTrustScore * 100,
    },
    {
      name: 'Portfolio Average',
      value: comparison.portfolio.averageSecurityTrustScore * 100,
    },
  ];

  const dataQualityChartData = [
    {
      name: 'Merchant',
      value: comparison.merchant.dataQuality * 100,
    },
    {
      name: 'Portfolio Average',
      value: comparison.portfolio.averageDataQuality * 100,
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Portfolio Analytics Comparison</CardTitle>
        <CardDescription>How this merchant compares to portfolio averages</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Classification Confidence Comparison */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Classification Confidence</h3>
            <div className="flex items-center gap-2">
              {getDifferenceIcon(comparison.differences.classificationConfidence)}
              <span className={`font-semibold ${getDifferenceColor(comparison.differences.classificationConfidence)}`}>
                {getDifferenceLabel(
                  comparison.differences.classificationConfidence,
                  comparison.percentages.classificationConfidence
                )}
              </span>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Merchant</p>
              <p className="text-2xl font-bold">{formatPercent(comparison.merchant.classificationConfidence)}</p>
            </div>
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Portfolio Average</p>
              <p className="text-2xl font-bold">{formatPercent(comparison.portfolio.averageClassificationConfidence)}</p>
            </div>
          </div>
          <ChartContainer
            title="Classification Confidence Comparison"
            description="Merchant vs portfolio average"
            isLoading={false}
          >
            <BarChart
              data={classificationChartData}
              dataKey="name"
              bars={[
                {
                  key: 'value',
                  name: 'Confidence',
                  color: '#8884d8',
                },
              ]}
              height={200}
              isLoading={false}
            />
          </ChartContainer>
        </div>

        {/* Security Trust Score Comparison */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Security Trust Score</h3>
            <div className="flex items-center gap-2">
              {getDifferenceIcon(comparison.differences.securityTrustScore)}
              <span className={`font-semibold ${getDifferenceColor(comparison.differences.securityTrustScore)}`}>
                {getDifferenceLabel(
                  comparison.differences.securityTrustScore,
                  comparison.percentages.securityTrustScore
                )}
              </span>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Merchant</p>
              <p className="text-2xl font-bold">{formatPercent(comparison.merchant.securityTrustScore)}</p>
            </div>
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Portfolio Average</p>
              <p className="text-2xl font-bold">{formatPercent(comparison.portfolio.averageSecurityTrustScore)}</p>
            </div>
          </div>
          <ChartContainer
            title="Security Trust Score Comparison"
            description="Merchant vs portfolio average"
            isLoading={false}
          >
            <BarChart
              data={securityChartData}
              dataKey="name"
              bars={[
                {
                  key: 'value',
                  name: 'Trust Score',
                  color: '#8884d8',
                },
              ]}
              height={200}
              isLoading={false}
            />
          </ChartContainer>
        </div>

        {/* Data Quality Comparison */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Data Quality</h3>
            <div className="flex items-center gap-2">
              {getDifferenceIcon(comparison.differences.dataQuality)}
              <span className={`font-semibold ${getDifferenceColor(comparison.differences.dataQuality)}`}>
                {getDifferenceLabel(
                  comparison.differences.dataQuality,
                  comparison.percentages.dataQuality
                )}
              </span>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Merchant</p>
              <p className="text-2xl font-bold">{formatPercent(comparison.merchant.dataQuality)}</p>
            </div>
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Portfolio Average</p>
              <p className="text-2xl font-bold">{formatPercent(comparison.portfolio.averageDataQuality)}</p>
            </div>
          </div>
          <ChartContainer
            title="Data Quality Comparison"
            description="Merchant vs portfolio average"
            isLoading={false}
          >
            <BarChart
              data={dataQualityChartData}
              dataKey="name"
              bars={[
                {
                  key: 'value',
                  name: 'Data Quality',
                  color: '#8884d8',
                },
              ]}
              height={200}
              isLoading={false}
            />
          </ChartContainer>
        </div>
      </CardContent>
    </Card>
  );
}

