'use client';

import { ExportButton } from '@/components/export/ExportButton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { PieChart, BarChart } from '@/components/charts/lazy';
import { getRiskIndicators } from '@/lib/api';
import { deferNonCriticalDataLoad } from '@/lib/lazy-loader';
import type { RiskIndicatorsData, RiskIndicator } from '@/types/merchant';
import { useEffect, useState, useMemo } from 'react';
import { RiskAlertsSection } from './RiskAlertsSection';

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

  // Group indicators by severity
  const groupedIndicators = useMemo(() => {
    if (!indicators?.indicators) return {};
    const grouped: Record<string, RiskIndicator[]> = {
      critical: [],
      high: [],
      medium: [],
      low: [],
    };
    indicators.indicators.forEach((indicator) => {
      const severity = indicator.severity || 'medium';
      if (grouped[severity]) {
        grouped[severity].push(indicator);
      }
    });
    return grouped;
  }, [indicators]);

  // Prepare chart data
  const severityDistributionData = useMemo(() => {
    if (!indicators?.indicators) return [];
    const counts: Record<string, number> = {};
    indicators.indicators.forEach((indicator) => {
      const severity = indicator.severity || 'medium';
      counts[severity] = (counts[severity] || 0) + 1;
    });
    return Object.entries(counts).map(([name, value]) => ({ name, value }));
  }, [indicators]);

  const getSeverityBadgeVariant = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'destructive';
      case 'high':
        return 'destructive';
      case 'medium':
        return 'secondary';
      case 'low':
        return 'outline';
      default:
        return 'outline';
    }
  };

  return (
    <div className="space-y-6">
      {/* Risk Alerts Section */}
      <RiskAlertsSection merchantId={merchantId} />

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
            <div className="space-y-6">
              {/* Severity Distribution Chart */}
              {severityDistributionData.length > 0 && (
                <ChartContainer
                  title="Severity Distribution"
                  description="Distribution of risk indicators by severity"
                  isLoading={false}
                >
                  <PieChart
                    data={severityDistributionData}
                    height={250}
                    isLoading={false}
                  />
                </ChartContainer>
              )}

              {/* Risk Indicators Table */}
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Severity</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Date</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {indicators.indicators
                    .sort((a, b) => {
                      const severityOrder: Record<string, number> = {
                        critical: 0,
                        high: 1,
                        medium: 2,
                        low: 3,
                      };
                      const aOrder = severityOrder[a.severity] ?? 99;
                      const bOrder = severityOrder[b.severity] ?? 99;
                      return aOrder - bOrder;
                    })
                    .map((indicator) => (
                      <TableRow key={indicator.id}>
                        <TableCell className="font-medium">{indicator.title}</TableCell>
                        <TableCell className="max-w-[300px] truncate" title={indicator.description}>
                          {indicator.description}
                        </TableCell>
                        <TableCell>
                          <Badge variant={getSeverityBadgeVariant(indicator.severity)}>
                            {indicator.severity}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{indicator.status || 'active'}</Badge>
                        </TableCell>
                        <TableCell className="text-right text-sm text-muted-foreground">
                          {indicator.createdAt
                            ? new Date(indicator.createdAt).toLocaleDateString()
                            : 'N/A'}
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>

              {/* Grouped by Severity (Alternative View) */}
              <div className="space-y-4">
                {Object.entries(groupedIndicators)
                  .filter(([_, indicators]) => indicators.length > 0)
                  .map(([severity, severityIndicators]) => (
                    <Card key={severity}>
                      <CardHeader>
                        <div className="flex items-center justify-between">
                          <CardTitle className="text-lg capitalize">{severity} Risk</CardTitle>
                          <Badge variant={getSeverityBadgeVariant(severity)}>
                            {severityIndicators.length} indicator{severityIndicators.length !== 1 ? 's' : ''}
                          </Badge>
                        </div>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-2">
                          {severityIndicators.map((indicator) => (
                            <Alert
                              key={indicator.id}
                              variant={indicator.severity === 'critical' ? 'destructive' : 'default'}
                            >
                              <AlertDescription>
                                <div className="flex items-center justify-between">
                                  <div>
                                    <p className="font-medium">{indicator.title}</p>
                                    <p className="text-sm text-muted-foreground">{indicator.description}</p>
                                  </div>
                                </div>
                              </AlertDescription>
                            </Alert>
                          ))}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
              </div>
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

