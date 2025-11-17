'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { BarChart3, TrendingUp, Target, PieChart as PieChartIcon } from 'lucide-react';
import { getBusinessIntelligenceMetrics } from '@/lib/api';
import { toast } from 'sonner';
import { AreaChart } from '@/components/charts/lazy';
import { BarChart } from '@/components/charts/lazy';

export default function BusinessIntelligencePage() {
  const [metrics, setMetrics] = useState({
    revenueGrowth: 0,
    marketShare: 0,
    performanceScore: 0,
    analyticsScore: 0,
  });
  const [loading, setLoading] = useState(true);
  const [chartLoading, setChartLoading] = useState(true);

  // Mock chart data
  const [performanceData, setPerformanceData] = useState<Array<{ name: string; value: number }>>([]);
  const [marketData, setMarketData] = useState<Array<{ name: string; value: number }>>([]);

  useEffect(() => {
    async function fetchMetrics() {
      try {
        const data = await getBusinessIntelligenceMetrics();
        setMetrics(data);
        
        // Generate mock performance data
        const quarters = ['Q1', 'Q2', 'Q3', 'Q4'];
        const performance = quarters.map((quarter) => ({
          name: quarter,
          value: Math.random() * 100 + 50, // Mock performance scores
        }));
        setPerformanceData(performance);
        
        // Generate mock market data
        const market = [
          { name: 'North America', value: 40 },
          { name: 'Europe', value: 30 },
          { name: 'Asia', value: 20 },
          { name: 'Other', value: 10 },
        ];
        setMarketData(market);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to load business intelligence metrics';
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
    return `${value.toFixed(1)}%`;
  };

  return (
    <AppLayout
      title="Business Intelligence"
      description="Advanced analytics and business insights"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Business Intelligence' },
      ]}
    >
      <div className="space-y-6">
        {/* Intelligence Metrics */}
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
                label="Revenue Growth"
                value={formatPercentage(metrics.revenueGrowth)}
                icon={TrendingUp}
                variant="success"
              />
              <MetricCard
                label="Market Share"
                value={formatPercentage(metrics.marketShare)}
                icon={PieChartIcon}
                variant="info"
              />
              <MetricCard
                label="Performance Score"
                value={metrics.performanceScore.toFixed(1)}
                icon={Target}
                variant="info"
              />
              <MetricCard
                label="Analytics Score"
                value={metrics.analyticsScore.toFixed(1)}
                icon={BarChart3}
                variant="info"
              />
            </>
          )}
        </div>

        {/* Intelligence Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <ChartContainer
            title="Business Performance"
            description="Key performance indicators"
            isLoading={chartLoading}
          >
            <AreaChart
              data={performanceData}
              dataKey="name"
              areas={[{ key: 'value', name: 'Performance Score', color: '#8884d8', fillOpacity: 0.6 }]}
              xAxisLabel="Quarter"
              yAxisLabel="Score"
              height={300}
              isLoading={chartLoading}
            />
          </ChartContainer>

          <ChartContainer
            title="Market Analysis"
            description="Market trends and insights"
            isLoading={chartLoading}
          >
            <BarChart
              data={marketData}
              dataKey="name"
              bars={[{ key: 'value', name: 'Market Share (%)', color: '#82ca9d' }]}
              xAxisLabel="Region"
              yAxisLabel="Share (%)"
              height={300}
              isLoading={chartLoading}
            />
          </ChartContainer>
        </div>
      </div>
    </AppLayout>
  );
}

