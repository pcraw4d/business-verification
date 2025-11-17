'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { AlertTriangle, TrendingUp, Shield, Activity } from 'lucide-react';
import { getRiskMetrics } from '@/lib/api';
import { toast } from 'sonner';
import { LineChart } from '@/components/charts/lazy';
import { BarChart } from '@/components/charts/lazy';

export default function RiskDashboardPage() {
  const [metrics, setMetrics] = useState({
    overallRiskScore: 0,
    highRiskMerchants: 0,
    riskAssessments: 0,
    riskTrend: 0,
  });
  const [loading, setLoading] = useState(true);
  const [chartLoading, setChartLoading] = useState(true);

  // Mock chart data
  const [trendData, setTrendData] = useState<Array<{ name: string; value: number }>>([]);
  const [distributionData, setDistributionData] = useState<Array<{ name: string; value: number }>>([]);

  useEffect(() => {
    async function fetchMetrics() {
      try {
        const data = await getRiskMetrics();
        setMetrics(data);
        
        // Generate mock trend data (last 6 months)
        const months = ['Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
        const trend = months.map((month) => ({
          name: month,
          value: Math.random() * 10 + 5, // Mock risk scores
        }));
        setTrendData(trend);
        
        // Generate mock distribution data
        const distribution = [
          { name: 'Low', value: 45 },
          { name: 'Medium', value: 30 },
          { name: 'High', value: 20 },
          { name: 'Critical', value: 5 },
        ];
        setDistributionData(distribution);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to load risk metrics';
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

  const formatPercentage = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(1)}%`;
  };

  return (
    <AppLayout
      title="Risk Assessment"
      description="Advanced risk scoring and assessment tools"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Risk Assessment' },
      ]}
    >
      <div className="space-y-6">
        {/* Risk Metrics */}
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
                label="Overall Risk Score"
                value={metrics.overallRiskScore.toFixed(1)}
                icon={AlertTriangle}
                variant="warning"
              />
              <MetricCard
                label="High Risk Merchants"
                value={metrics.highRiskMerchants.toLocaleString()}
                icon={Shield}
                variant="danger"
              />
              <MetricCard
                label="Risk Assessments"
                value={metrics.riskAssessments.toLocaleString()}
                icon={Activity}
                variant="info"
              />
              <MetricCard
                label="Risk Trend"
                value={formatPercentage(metrics.riskTrend)}
                icon={TrendingUp}
                variant="warning"
                trend={{ value: metrics.riskTrend, isPositive: metrics.riskTrend < 0 }}
              />
            </>
          )}
        </div>

        {/* Risk Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <ChartContainer
            title="Risk Trend Analysis"
            description="6-month risk trend history"
            isLoading={chartLoading}
          >
            <LineChart
              data={trendData}
              dataKey="name"
              lines={[{ key: 'value', name: 'Risk Score', color: '#ef4444' }]}
              xAxisLabel="Month"
              yAxisLabel="Risk Score"
              height={300}
              isLoading={chartLoading}
            />
          </ChartContainer>

          <ChartContainer
            title="Risk Distribution"
            description="Risk level distribution across portfolio"
            isLoading={chartLoading}
          >
            <BarChart
              data={distributionData}
              dataKey="name"
              bars={[{ key: 'value', name: 'Merchants', color: '#8884d8' }]}
              xAxisLabel="Risk Level"
              yAxisLabel="Count"
              height={300}
              isLoading={chartLoading}
            />
          </ChartContainer>
        </div>

        {/* Risk Details */}
        <Card>
          <CardHeader>
            <CardTitle>Risk Assessment Details</CardTitle>
            <CardDescription>Comprehensive risk analysis and scoring</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Risk assessment details will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

