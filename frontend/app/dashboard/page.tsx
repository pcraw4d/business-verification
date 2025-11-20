'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { ChartLine, TrendingUp, Users, DollarSign } from 'lucide-react';
import { getDashboardMetrics, getPortfolioAnalytics, getPortfolioStatistics } from '@/lib/api';
import { formatPercentWithSign, formatNumber } from '@/lib/number-format';
import { toast } from 'sonner';
import { LineChart } from '@/components/charts/lazy';
import { PieChart } from '@/components/charts/lazy';

export default function DashboardPage() {
  const [metrics, setMetrics] = useState({
    totalMerchants: 0,
    revenue: 0,
    growthRate: 0,
    analyticsScore: 0,
  });
  const [loading, setLoading] = useState(true);
  const [chartLoading, setChartLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Portfolio data
  const [trendData, setTrendData] = useState<Array<{ name: string; value: number }>>([]);
  const [distributionData, setDistributionData] = useState<Array<{ name: string; value: number }>>([]);

  useEffect(() => {
    async function fetchMetrics() {
      try {
        setError(null);
        setLoading(true);
        setChartLoading(true);

        // Fetch portfolio data in parallel
        const [portfolioAnalytics, portfolioStatistics, dashboardMetrics] = await Promise.allSettled([
          getPortfolioAnalytics(),
          getPortfolioStatistics(),
          getDashboardMetrics(), // Keep v3 endpoint as fallback/supplement
        ]);

        // Use portfolio statistics for primary metrics
        if (portfolioStatistics.status === 'fulfilled') {
          const stats = portfolioStatistics.value;
          setMetrics({
            totalMerchants: stats.totalMerchants || 0,
            revenue: 0, // Revenue not in portfolio statistics
            growthRate: 0, // Growth rate not in portfolio statistics
            analyticsScore: stats.averageRiskScore ? (1 - stats.averageRiskScore) * 100 : 0, // Convert risk score to analytics score
          });

          // Use industry breakdown for distribution chart
          if (stats.industryBreakdown && stats.industryBreakdown.length > 0) {
            const distribution = stats.industryBreakdown.map((item) => ({
              name: item.industry || 'Unknown',
              value: item.count || 0,
            }));
            setDistributionData(distribution);
          }
        }

        // Use portfolio analytics for additional metrics
        let hasDistributionData = false;
        if (portfolioAnalytics.status === 'fulfilled') {
          const analytics = portfolioAnalytics.value;
          // Update analytics score if available
          if (analytics.averageClassificationConfidence) {
            setMetrics((prev) => ({
              ...prev,
              analyticsScore: analytics.averageClassificationConfidence * 100,
            }));
          }

          // Use industry distribution for chart if statistics didn't provide it
          if (analytics.industryDistribution) {
            const distribution = Object.entries(analytics.industryDistribution).map(([name, value]) => ({
              name,
              value: typeof value === 'number' ? value : 0,
            }));
            setDistributionData(distribution);
            hasDistributionData = true;
          }
        }

        // Check if we got distribution data from statistics
        if (portfolioStatistics.status === 'fulfilled') {
          const stats = portfolioStatistics.value;
          if (stats.industryBreakdown && stats.industryBreakdown.length > 0) {
            hasDistributionData = true;
          }
        }

        // Fallback to v3 dashboard metrics if portfolio endpoints fail
        if (dashboardMetrics.status === 'fulfilled') {
          const v3Data = dashboardMetrics.value;
          setMetrics((prev) => ({
            totalMerchants: prev.totalMerchants || v3Data.totalMerchants || 0,
            revenue: v3Data.revenue || 0,
            growthRate: v3Data.growthRate || 0,
            analyticsScore: prev.analyticsScore || v3Data.analyticsScore || 0,
          }));
        }

        // Generate trend data (placeholder - can be enhanced with time series data)
        const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
        const trend = months.map((month) => ({
          name: month,
          value: Math.floor(Math.random() * 10000) + 5000, // Placeholder - should use real time series data
        }));
        setTrendData(trend);

        // If no distribution data from portfolio endpoints, use placeholder
        if (!hasDistributionData) {
          const distribution = [
            { name: 'Technology', value: 35 },
            { name: 'Retail', value: 25 },
            { name: 'Finance', value: 20 },
            { name: 'Healthcare', value: 15 },
            { name: 'Other', value: 5 },
          ];
          setDistributionData(distribution);
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to load dashboard metrics';
        setError(errorMessage);
        toast.error('Failed to load metrics', {
          description: errorMessage,
        });
      } finally {
        setLoading(false);
        setChartLoading(false);
      }
    }

    fetchMetrics();
  }, []);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };


  return (
    <AppLayout
      title="Business Intelligence"
      description="Comprehensive business analytics and insights"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Business Intelligence' },
      ]}
    >
      <div className="space-y-6">
        {/* Metrics */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {loading ? (
            <>
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
            </>
          ) : (
            <>
              <MetricCard
                label="Total Merchants"
                value={metrics.totalMerchants.toLocaleString()}
                icon={Users}
                variant="info"
              />
              <MetricCard
                label="Revenue"
                value={formatCurrency(metrics.revenue)}
                icon={DollarSign}
                variant="success"
              />
              <MetricCard
                label="Growth Rate"
                value={formatPercentWithSign(metrics.growthRate)}
                icon={TrendingUp}
                variant="success"
                trend={{ value: metrics.growthRate, isPositive: metrics.growthRate >= 0 }}
              />
              <MetricCard
                label="Analytics Score"
                value={formatNumber(metrics.analyticsScore, 1)}
                icon={ChartLine}
                variant="info"
              />
            </>
          )}
        </div>

        {/* Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <ChartContainer
            title="Business Trends"
            description="Monthly business growth trends"
            isLoading={chartLoading}
          >
            <LineChart
              data={trendData}
              dataKey="name"
              lines={[{ key: 'value', name: 'Revenue', color: '#8884d8' }]}
              xAxisLabel="Month"
              yAxisLabel="Revenue ($)"
              height={300}
              isLoading={chartLoading}
            />
          </ChartContainer>

          <ChartContainer
            title="Merchant Distribution"
            description="Distribution by industry"
            isLoading={chartLoading}
          >
            <PieChart
              data={distributionData}
              height={300}
              isLoading={chartLoading}
            />
          </ChartContainer>
        </div>

        {/* Additional Analytics */}
        <Card>
          <CardHeader>
            <CardTitle>Analytics Overview</CardTitle>
            <CardDescription>Detailed business intelligence metrics</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Analytics data will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

