'use client';

import { ExportButton } from '@/components/export/ExportButton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { getRiskIndicators } from '@/lib/api';
import { deferNonCriticalDataLoad } from '@/lib/lazy-loader';
import type { RiskIndicatorsData } from '@/types/merchant';
import { useEffect, useState } from 'react';

interface RiskIndicatorsTabProps {
  merchantId: string;
}

export function RiskIndicatorsTab({ merchantId }: RiskIndicatorsTabProps) {
  const [indicators, setIndicators] = useState<RiskIndicatorsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Defer loading risk indicators as non-critical data
    deferNonCriticalDataLoad(async () => {
      try {
        setLoading(true);
        const data = await getRiskIndicators(merchantId);
        setIndicators(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load risk indicators');
      } finally {
        setLoading(false);
      }
    });
  }, [merchantId]);

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error && !indicators) {
    return (
      <EmptyState
        type="error"
        title="Error Loading Indicators"
        message={error}
        actionLabel="Retry"
        onAction={() => {
          setError(null);
          setLoading(true);
          getRiskIndicators(merchantId)
            .then(setIndicators)
            .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load'))
            .finally(() => setLoading(false));
        }}
      />
    );
  }

  if (!loading && !indicators) {
    return (
      <EmptyState
        type="noData"
        title="No Risk Indicators"
        message="No active risk indicators found for this merchant."
      />
    );
  }

  const getExportData = async () => {
    return {
      indicators,
      merchantId,
      exportedAt: new Date().toISOString(),
    };
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Risk Indicators</CardTitle>
              <CardDescription>Active risk indicators for this merchant</CardDescription>
            </div>
            {indicators && (
              <ExportButton
                data={getExportData}
                exportType="risk"
                merchantId={merchantId}
                formats={['csv', 'json', 'excel', 'pdf']}
              />
            )}
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <Skeleton className="h-32 w-full" />
          ) : indicators && indicators.indicators.length > 0 ? (
            <div className="space-y-4">
              {indicators.indicators.map((indicator) => (
                <Alert key={indicator.id} variant={indicator.severity === 'critical' ? 'destructive' : 'default'}>
                  <AlertDescription>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium">{indicator.title}</p>
                        <p className="text-sm text-muted-foreground">{indicator.description}</p>
                      </div>
                      <Badge variant={indicator.severity === 'critical' ? 'destructive' : 'outline'}>
                        {indicator.severity}
                      </Badge>
                    </div>
                  </AlertDescription>
                </Alert>
              ))}
            </div>
          ) : (
            <EmptyState
              type="noData"
              title="No Active Indicators"
              message="No active risk indicators at this time."
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}

