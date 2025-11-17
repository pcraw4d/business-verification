'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { Skeleton } from '@/components/ui/skeleton';
import { TrendingUp, BarChart3, PieChart as PieChartIcon } from 'lucide-react';
import { LineChart } from '@/components/charts/lazy';

export default function MarketAnalysisPage() {
  const [loading, setLoading] = useState(true);
  const [chartLoading, setChartLoading] = useState(true);
  const [trendData, setTrendData] = useState<Array<{ name: string; value: number }>>([]);

  useEffect(() => {
    // Simulate data loading
    setTimeout(() => {
      // Generate mock trend data
      const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      const trend = months.map((month) => ({
        name: month,
        value: Math.floor(Math.random() * 1000000) + 500000, // Mock market size
      }));
      setTrendData(trend);
      setLoading(false);
      setChartLoading(false);
    }, 1000);
  }, []);

  return (
    <AppLayout
      title="Market Analysis"
      description="Comprehensive market analysis and insights"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Market Analysis' },
      ]}
    >
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {loading ? (
            <>
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
            </>
          ) : (
            <>
              <MetricCard label="Market Size" value="$2.5M" icon={BarChart3} variant="info" />
              <MetricCard label="Growth Rate" value="+12.5%" icon={TrendingUp} variant="success" />
              <MetricCard label="Market Share" value="15.2%" icon={PieChartIcon} variant="info" />
            </>
          )}
        </div>

        <ChartContainer
          title="Market Trends"
          description="Market analysis and trends"
          isLoading={chartLoading}
        >
          <LineChart
            data={trendData}
            dataKey="name"
            lines={[{ key: 'value', name: 'Market Size', color: '#8884d8' }]}
            xAxisLabel="Month"
            yAxisLabel="Market Size ($)"
            height={300}
            isLoading={chartLoading}
          />
        </ChartContainer>
      </div>
    </AppLayout>
  );
}

