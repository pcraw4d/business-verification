'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Settings, Server, Users, Activity } from 'lucide-react';
import { getSystemMetrics } from '@/lib/api';
import type { SystemMetrics } from '@/types/dashboard';
import { toast } from 'sonner';

export default function AdminPage() {
  const [metrics, setMetrics] = useState<SystemMetrics>({
    systemHealth: 100,
    serverStatus: 'Online',
    databaseStatus: 'Connected',
    responseTime: 0,
    cpuUsage: 0,
    memoryUsage: 0,
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
      title="Admin Dashboard"
      description="System administration and monitoring"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Admin Dashboard' },
      ]}
    >
      <div className="space-y-6">
        {/* System Metrics */}
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
                label="System Status"
                value={metrics.serverStatus}
                icon={Server}
                variant={metrics.serverStatus === 'Online' ? 'success' : 'danger'}
              />
              <MetricCard
                label="System Health"
                value={`${metrics.systemHealth}%`}
                icon={Activity}
                variant={metrics.systemHealth >= 90 ? 'success' : metrics.systemHealth >= 70 ? 'warning' : 'danger'}
              />
              <MetricCard
                label="System Load"
                value={metrics.cpuUsage ? `${metrics.cpuUsage.toFixed(1)}%` : 'N/A'}
                icon={Activity}
                variant={metrics.cpuUsage && metrics.cpuUsage < 70 ? 'success' : metrics.cpuUsage && metrics.cpuUsage < 90 ? 'warning' : 'danger'}
              />
              <MetricCard
                label="Response Time"
                value={`${metrics.responseTime}ms`}
                icon={Settings}
                variant={metrics.responseTime < 200 ? 'success' : metrics.responseTime < 500 ? 'warning' : 'danger'}
              />
            </>
          )}
        </div>

        {/* Admin Panels */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>System Metrics</CardTitle>
              <CardDescription>Memory, CPU, and performance monitoring</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-muted-foreground">
                System metrics will be displayed here
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>User Management</CardTitle>
              <CardDescription>Manage users and permissions</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-muted-foreground">
                User management interface will be displayed here
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </AppLayout>
  );
}

