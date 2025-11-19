'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { TrendingUp, TrendingDown, BarChart3, AlertCircle } from 'lucide-react';
import { useEffect, useState, useCallback } from 'react';
import { getMerchantAnalytics, getPortfolioAnalytics } from '@/lib/api';
import type { AnalyticsComparison, AnalyticsData, PortfolioAnalytics } from '@/types/merchant';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { BarChart } from '@/components/charts/lazy';

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
        } else {
          console.error('Failed to fetch merchant analytics:', analyticsResult[0].reason);
          setError('Failed to load merchant analytics.');
        }
      }

      // Step 2: Fetch portfolio analytics
      const portfolioResult = await Promise.allSettled([
        getPortfolioAnalytics(),
      ]);

      let portfolioAnalytics: PortfolioAnalytics | null = null;
      if (portfolioResult[0].status === 'fulfilled') {
        portfolioAnalytics = portfolioResult[0].value;
      } else {
        console.error('Failed to fetch portfolio analytics:', portfolioResult[0].reason);
        setError((prev) => (prev ? prev + ' ' : '') + 'Failed to load portfolio analytics.');
      }

      if (analytics && portfolioAnalytics) {
        const merchantClassificationConfidence = analytics.classification?.confidenceScore || 0;
        const merchantSecurityTrustScore = analytics.security?.trustScore || 0;
        const merchantDataQuality = analytics.quality?.completenessScore || 0;

        const portfolioClassificationConfidence = portfolioAnalytics.averageClassificationConfidence || 0;
        const portfolioSecurityTrustScore = portfolioAnalytics.averageSecurityTrustScore || 0;
        const portfolioDataQuality = portfolioAnalytics.averageDataQuality || 0;

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
        setError('Not enough data to perform analytics comparison.');
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
    const sign = difference > 0 ? '+' : '';
    return `${sign}${(percentage).toFixed(1)}%`;
  };

  if (loading) {
    return <Skeleton className="h-96 w-full" />;
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive">Error Loading Analytics Comparison</CardTitle>
          <CardDescription>Could not load analytics comparison data.</CardDescription>
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

  if (!comparison) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Analytics Comparison</CardTitle>
          <CardDescription>No comparison data available.</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            We could not retrieve sufficient data to compare this merchant against portfolio analytics.
          </p>
          <Button onClick={fetchComparisonData} className="mt-4" variant="outline">
            Reload
          </Button>
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
              <p className="text-2xl font-bold">{(comparison.merchant.classificationConfidence * 100).toFixed(1)}%</p>
            </div>
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Portfolio Average</p>
              <p className="text-2xl font-bold">{(comparison.portfolio.averageClassificationConfidence * 100).toFixed(1)}%</p>
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
              <p className="text-2xl font-bold">{(comparison.merchant.securityTrustScore * 100).toFixed(1)}%</p>
            </div>
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Portfolio Average</p>
              <p className="text-2xl font-bold">{(comparison.portfolio.averageSecurityTrustScore * 100).toFixed(1)}%</p>
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
              <p className="text-2xl font-bold">{(comparison.merchant.dataQuality * 100).toFixed(1)}%</p>
            </div>
            <div className="p-4 border rounded-lg">
              <p className="text-sm text-muted-foreground">Portfolio Average</p>
              <p className="text-2xl font-bold">{(comparison.portfolio.averageDataQuality * 100).toFixed(1)}%</p>
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

