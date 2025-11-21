'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { ClipboardCheck, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import { getComplianceStatus } from '@/lib/api';
import { toast } from 'sonner';

export default function CompliancePage() {
  const [status, setStatus] = useState({
    overallScore: 0,
    pendingReviews: 0,
    complianceTrend: 'Stable',
    regulatoryFrameworks: 0,
  });
  const [loading, setLoading] = useState(true);
  const [mounted, setMounted] = useState(false);

  // Client-side formatted values to prevent hydration errors
  const [formattedPendingReviews, setFormattedPendingReviews] = useState<string>('0');

  useEffect(() => {
    async function fetchStatus() {
      try {
        const data = await getComplianceStatus();
        setStatus(data);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to load compliance status';
        toast.error('Failed to load compliance data', {
          description: errorMessage,
        });
      } finally {
        setLoading(false);
      }
    }

    fetchStatus();
  }, []);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Format numbers on client side only to prevent hydration errors
  useEffect(() => {
    if (!mounted) return;

    if (status.pendingReviews !== undefined && status.pendingReviews !== null) {
      setFormattedPendingReviews(status.pendingReviews.toLocaleString());
    } else {
      setFormattedPendingReviews('0');
    }
  }, [mounted, status.pendingReviews]);

  // Calculate derived metrics
  const compliantCount = Math.round((status.overallScore / 100) * 100); // Simplified calculation
  const nonCompliantCount = 100 - compliantCount;

  return (
    <AppLayout
      title="Compliance Status"
      description="Track compliance across all frameworks"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Compliance Status' },
      ]}
    >
      <div className="space-y-6">
        {/* Compliance Metrics */}
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
                label="Compliance Score"
                value={`${status.overallScore}%`}
                icon={ClipboardCheck}
                variant={status.overallScore >= 90 ? 'success' : status.overallScore >= 70 ? 'warning' : 'danger'}
              />
              <MetricCard
                label="Pending Review"
                value={mounted ? formattedPendingReviews : '0'}
                icon={AlertCircle}
                variant={status.pendingReviews === 0 ? 'success' : status.pendingReviews < 10 ? 'warning' : 'danger'}
              />
              <MetricCard
                label="Compliance Trend"
                value={status.complianceTrend}
                icon={CheckCircle}
                variant={status.complianceTrend === 'Improving' ? 'success' : status.complianceTrend === 'Stable' ? 'info' : 'warning'}
              />
              <MetricCard
                label="Regulatory Frameworks"
                value={status.regulatoryFrameworks.toString()}
                icon={XCircle}
                variant="info"
              />
            </>
          )}
        </div>

        {/* Compliance Details */}
        <Card>
          <CardHeader>
            <CardTitle>Compliance Overview</CardTitle>
            <CardDescription>FATF and regulatory compliance tracking</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Compliance details will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

