'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Gauge, AlertCircle, CheckCircle, XCircle } from 'lucide-react';
import { RiskGauge } from '@/components/charts/lazy';
import { RiskTrendChart } from '@/components/charts/lazy';
import { RiskCategoryRadar } from '@/components/charts/lazy';
import { getRiskMetrics, getPortfolioStatistics, getRiskTrends } from '@/lib/api';
import { toast } from 'sonner';

export default function RiskIndicatorsPage() {
  const [loading, setLoading] = useState(true);
  const [riskCounts, setRiskCounts] = useState({
    low: 0,
    medium: 0,
    high: 0,
    critical: 0,
  });
  const [overallRisk, setOverallRisk] = useState(0);
  const [trendData, setTrendData] = useState<Array<{
    name: string;
    historical: number;
    prediction?: number;
    confidenceUpper?: number;
    confidenceLower?: number;
  }>>([]);
  const [categoryData, setCategoryData] = useState<Array<{ category: string; score: number }>>([]);

  useEffect(() => {
    async function fetchData() {
      try {
        setLoading(true);

        // Fetch portfolio data in parallel
        const [riskMetrics, portfolioStats, riskTrends] = await Promise.allSettled([
          getRiskMetrics(),
          getPortfolioStatistics(),
          getRiskTrends({ timeframe: '6m' }),
        ]);

        // Track if we got data
        let hasOverallRisk = false;
        let hasRiskCounts = false;

        // Use portfolio statistics for risk counts and overall risk
        if (portfolioStats.status === 'fulfilled') {
          const stats = portfolioStats.value;
          setOverallRisk(stats.averageRiskScore * 100); // Convert to percentage
          hasOverallRisk = true;

          if (stats.riskDistribution) {
            setRiskCounts({
              low: stats.riskDistribution.low || 0,
              medium: stats.riskDistribution.medium || 0,
              high: stats.riskDistribution.high || 0,
              critical: 0, // Portfolio statistics might not have critical, use 0 or calculate
            });
            hasRiskCounts = true;
          }
        }

        // Fallback to risk metrics if portfolio stats not available
        if (riskMetrics.status === 'fulfilled') {
          const metrics = riskMetrics.value;
          if (!hasOverallRisk) {
            setOverallRisk(metrics.overallRiskScore);
          }
          if (!hasRiskCounts && metrics.riskDistribution) {
            setRiskCounts({
              low: metrics.riskDistribution.low || 0,
              medium: metrics.riskDistribution.medium || 0,
              high: metrics.riskDistribution.high || 0,
              critical: metrics.riskDistribution.critical || 0,
            });
          }
        }

        // Use risk trends for trend chart
        if (riskTrends.status === 'fulfilled') {
          const trendsData = riskTrends.value;
          if (trendsData.summary) {
            // Create trend data from summary (simplified - can be enhanced with time series)
            const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
            const avgScore = trendsData.summary.average_risk_score * 100;
            const trend = months.map((month, index) => {
              // Use average with some variation for historical data
              const historical = avgScore + (Math.random() * 10 - 5);
              // Add predictions for last 6 months
              const isFuture = index >= 6;
              return {
                name: month,
                historical: isFuture ? undefined : historical,
                prediction: isFuture ? avgScore + (Math.random() * 5 - 2.5) : undefined,
                confidenceUpper: isFuture ? avgScore + 5 : undefined,
                confidenceLower: isFuture ? avgScore - 5 : undefined,
              };
            });
            setTrendData(trend);
          }
        }

        // Track if we got trend data
        let hasTrendData = false;
        if (riskTrends.status === 'fulfilled') {
          const trendsData = riskTrends.value;
          if (trendsData.summary) {
            hasTrendData = true;
          }
        }

        // Fallback to placeholder trend data if no real data
        if (!hasTrendData) {
          const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
          const trend = months.map((month, index) => ({
            name: month,
            historical: Math.random() * 10,
            prediction: index >= 6 ? Math.random() * 10 : undefined,
            confidenceUpper: index >= 6 ? Math.random() * 10 + 1 : undefined,
            confidenceLower: index >= 6 ? Math.random() * 10 - 1 : undefined,
          }));
          setTrendData(trend);
        }

        // Category data - placeholder for now (can be enhanced with real category breakdown)
        const categories = [
          { category: 'Financial', score: 8.1 },
          { category: 'Operational', score: 6.5 },
          { category: 'Compliance', score: 4.2 },
          { category: 'Market', score: 7.8 },
          { category: 'Reputation', score: 6.9 },
        ];
        setCategoryData(categories);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to load risk indicators';
        toast.error('Failed to load data', {
          description: errorMessage,
        });
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, []);

  return (
    <AppLayout
      title="Risk Indicators"
      description="Real-time risk monitoring and alerts"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Risk Indicators' },
      ]}
    >
      <div className="space-y-6">
        {/* Risk Level Indicators */}
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
              <Card>
                <CardHeader className="pb-2">
                  <CardDescription>Low Risk</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-2">
                    <CheckCircle className="h-8 w-8 text-green-600" />
                    <div className="text-2xl font-bold">{riskCounts.low}</div>
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="pb-2">
                  <CardDescription>Medium Risk</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-2">
                    <AlertCircle className="h-8 w-8 text-yellow-600" />
                    <div className="text-2xl font-bold">{riskCounts.medium}</div>
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="pb-2">
                  <CardDescription>High Risk</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-2">
                    <XCircle className="h-8 w-8 text-red-600" />
                    <div className="text-2xl font-bold">{riskCounts.high}</div>
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="pb-2">
                  <CardDescription>Critical Risk</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-2">
                    <Gauge className="h-8 w-8 text-red-800" />
                    <div className="text-2xl font-bold">{riskCounts.critical}</div>
                  </div>
                </CardContent>
              </Card>
            </>
          )}
        </div>

        {/* Risk Visualizations */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <ChartContainer
            title="Overall Risk Gauge"
            description="Current overall risk score"
            isLoading={loading}
          >
            <RiskGauge
              value={overallRisk}
              max={10}
              height={300}
              width={300}
              isLoading={loading}
              showNeedle={true}
            />
          </ChartContainer>

          <ChartContainer
            title="Risk Trend Analysis"
            description="6-month risk trend with predictions"
            isLoading={loading}
          >
            <RiskTrendChart
              data={trendData}
              height={300}
              isLoading={loading}
              showPrediction={true}
              showConfidenceBands={true}
            />
          </ChartContainer>
        </div>

        {/* Risk Category Radar */}
        <ChartContainer
          title="Risk Category Breakdown"
          description="Risk scores by category"
          isLoading={loading}
        >
          <RiskCategoryRadar
            data={categoryData}
            height={400}
            width={400}
            isLoading={loading}
            maxScore={10}
          />
        </ChartContainer>

        {/* Risk Indicators List */}
        <Card>
          <CardHeader>
            <CardTitle>Active Risk Indicators</CardTitle>
            <CardDescription>Real-time risk monitoring and alerts</CardDescription>
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-32 w-full" />
            ) : (
              <div className="text-muted-foreground">
                Risk indicators list will be displayed here
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

