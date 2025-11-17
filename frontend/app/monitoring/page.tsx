'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Activity, Server, Database, Zap } from 'lucide-react';
import { getSystemMetrics } from '@/lib/api';
import type { SystemMetrics } from '@/types/dashboard';
import { toast } from 'sonner';

export default function MonitoringPage() {
  const [metrics, setMetrics] = useState<SystemMetrics>({
    systemHealth: 100,
    serverStatus: 'Online',
    databaseStatus: 'Connected',
    responseTime: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchMetrics() {
      try {
        const data = await getSystemMetrics();
        setMetrics({
          ...data,
          cpuUsage: data.cpuUsage ?? 0,
          memoryUsage: data.memoryUsage ?? 0,
        });
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to load system metrics';
        toast.error('Failed to load metrics', {
          description: errorMessage,
        });
      } finally {
        setLoading(false);
      }
    }

    fetchMetrics();
    
    // Refresh metrics every 30 seconds
    const interval = setInterval(fetchMetrics, 30000);
    return () => clearInterval(interval);
  }, []);

  return (
    <AppLayout
      title="Monitoring Dashboard"
      description="System monitoring and performance metrics"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Monitoring' },
      ]}
    >
      <div className="space-y-6">
        {/* Monitoring Metrics */}
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
                label="System Health"
                value={`${metrics.systemHealth}%`}
                icon={Activity}
                variant={metrics.systemHealth >= 90 ? 'success' : metrics.systemHealth >= 70 ? 'warning' : 'danger'}
              />
              <MetricCard
                label="Server Status"
                value={metrics.serverStatus}
                icon={Server}
                variant={metrics.serverStatus === 'Online' ? 'success' : 'danger'}
              />
              <MetricCard
                label="Database Status"
                value={metrics.databaseStatus}
                icon={Database}
                variant={metrics.databaseStatus === 'Connected' ? 'success' : 'danger'}
              />
              <MetricCard
                label="Response Time"
                value={`${metrics.responseTime}ms`}
                icon={Zap}
                variant={metrics.responseTime < 200 ? 'success' : metrics.responseTime < 500 ? 'warning' : 'danger'}
              />
            </>
          )}
        </div>

        {/* Monitoring Details */}
        <Card>
          <CardHeader>
            <CardTitle>System Monitoring</CardTitle>
            <CardDescription>Real-time system performance and health metrics</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Monitoring details will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

