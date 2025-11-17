'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { ChartLine, TrendingUp, Users, DollarSign } from 'lucide-react';
import { getDashboardMetrics } from '@/lib/api';
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

  // Mock chart data - will be replaced with real API data
  const [trendData, setTrendData] = useState<Array<{ name: string; value: number }>>([]);
  const [distributionData, setDistributionData] = useState<Array<{ name: string; value: number }>>([]);

  useEffect(() => {
    async function fetchMetrics() {
      try {
        const data = await getDashboardMetrics();
        setMetrics(data);
        
        // Generate mock trend data (last 12 months)
        const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
        const trend = months.map((month, index) => ({
          name: month,
          value: Math.floor(Math.random() * 10000) + 5000, // Mock revenue data
        }));
        setTrendData(trend);
        
        // Generate mock distribution data
        const distribution = [
          { name: 'Technology', value: 35 },
          { name: 'Retail', value: 25 },
          { name: 'Finance', value: 20 },
          { name: 'Healthcare', value: 15 },
          { name: 'Other', value: 5 },
        ];
        setDistributionData(distribution);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to load dashboard metrics';
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

  const formatPercentage = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(1)}%`;
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
                value={formatPercentage(metrics.growthRate)}
                icon={TrendingUp}
                variant="success"
                trend={{ value: metrics.growthRate, isPositive: metrics.growthRate >= 0 }}
              />
              <MetricCard
                label="Analytics Score"
                value={metrics.analyticsScore.toFixed(1)}
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

