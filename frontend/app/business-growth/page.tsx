'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { Skeleton } from '@/components/ui/skeleton';
import { TrendingUp, Target, ArrowUpRight } from 'lucide-react';
import { AreaChart } from '@/components/charts/lazy';

export default function BusinessGrowthPage() {
  const [loading, setLoading] = useState(true);
  const [chartLoading, setChartLoading] = useState(true);
  const [growthData, setGrowthData] = useState<Array<{ name: string; revenue: number; customers: number }>>([]);

  useEffect(() => {
    // Simulate data loading
    setTimeout(() => {
      // Generate mock growth data
      const quarters = ['Q1 2023', 'Q2 2023', 'Q3 2023', 'Q4 2023', 'Q1 2024', 'Q2 2024'];
      const growth = quarters.map((quarter) => ({
        name: quarter,
        revenue: Math.floor(Math.random() * 50) + 20, // Mock revenue growth %
        customers: Math.floor(Math.random() * 30) + 10, // Mock customer growth %
      }));
      setGrowthData(growth);
      setLoading(false);
      setChartLoading(false);
    }, 1000);
  }, []);

  return (
    <AppLayout
      title="Business Growth Analytics"
      description="Track business growth and performance metrics"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Business Growth' },
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
              <MetricCard
                label="Revenue Growth"
                value="+15.2%"
                icon={TrendingUp}
                variant="success"
                trend={{ value: 15.2, isPositive: true }}
              />
              <MetricCard
                label="Customer Growth"
                value="+8.5%"
                icon={Target}
                variant="success"
                trend={{ value: 8.5, isPositive: true }}
              />
              <MetricCard
                label="Market Expansion"
                value="12"
                icon={ArrowUpRight}
                variant="info"
              />
            </>
          )}
        </div>

        <ChartContainer
          title="Growth Trends"
          description="Business growth trends over time"
          isLoading={chartLoading}
        >
          <AreaChart
            data={growthData.map(item => ({ name: item.name, value: item.revenue }))}
            dataKey="name"
            areas={[
              { key: 'value', name: 'Revenue Growth (%)', color: '#8884d8', fillOpacity: 0.6 },
            ]}
            xAxisLabel="Quarter"
            yAxisLabel="Growth (%)"
            height={300}
            isLoading={chartLoading}
          />
        </ChartContainer>
      </div>
    </AppLayout>
  );
}

