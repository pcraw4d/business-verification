'use client';

import { RiskGauge, LineChart, BarChart, AreaChart } from '@/components/charts/lazy';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { ExportButton } from '@/components/export/ExportButton';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ProgressIndicator } from '@/components/ui/progress-indicator';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { RiskWebSocketProvider, useRiskWebSocket, WebSocketStatusIndicator } from '@/components/websocket/RiskWebSocketProvider';
import { getAssessmentStatus, getRiskAssessment, getRiskHistory, getRiskPredictions, startRiskAssessment } from '@/lib/api';
import { ErrorHandler } from '@/lib/error-handler';
import { formatNumber, formatPercent } from '@/lib/number-format';
import type { RiskAssessment, RiskAssessmentRequest, RiskFactor } from '@/types/merchant';
import { useEffect, useState, useMemo } from 'react';
import { RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import { toast } from 'sonner';
import { RiskBenchmarkComparison } from './RiskBenchmarkComparison';
import { RiskExplainabilitySection } from './RiskExplainabilitySection';
import { RiskRecommendationsSection } from './RiskRecommendationsSection';

interface RiskAssessmentTabProps {
  merchantId: string;
}

function RiskAssessmentTabContent({ merchantId }: RiskAssessmentTabProps) {
  useRiskWebSocket(); // Initialize WebSocket connection
  const [assessment, setAssessment] = useState<RiskAssessment | null>(null);
  const [riskHistory, setRiskHistory] = useState<RiskAssessment[]>([]);
  const [riskPredictions, setRiskPredictions] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [processing, setProcessing] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastRefreshTime, setLastRefreshTime] = useState<Date | null>(null);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    loadAssessment();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [merchantId]);

  // Keyboard shortcut: R to refresh
  useKeyboardShortcuts([
    {
      key: 'r',
      handler: handleRefresh,
      description: 'Refresh risk assessment data',
    },
  ]);

  // Listen for WebSocket risk updates
  useEffect(() => {
    const handleRiskUpdate = (event: CustomEvent) => {
      const data = event.detail;
      if (data.merchantId === merchantId && data.assessment) {
        setAssessment(data.assessment);
        toast.info('Risk assessment updated', {
          description: 'Real-time update received',
        });
      }
    };

    const handleRiskAlert = (event: CustomEvent) => {
      const data = event.detail;
      if (data.merchantId === merchantId) {
        toast.warning('Risk Alert', {
          description: data.message || 'New risk alert received',
        });
      }
    };

    window.addEventListener('riskUpdate', handleRiskUpdate as EventListener);
    window.addEventListener('riskAlert', handleRiskAlert as EventListener);

    return () => {
      window.removeEventListener('riskUpdate', handleRiskUpdate as EventListener);
      window.removeEventListener('riskAlert', handleRiskAlert as EventListener);
    };
  }, [merchantId]);

  async function loadAssessment(bypassCache = false) {
    try {
      if (!bypassCache) {
        setLoading(true);
      } else {
        setIsRefreshing(true);
      }
      setError(null);
      if (process.env.NODE_ENV === 'test') {
        console.log('[RiskAssessmentTab] Starting to load assessment:', merchantId);
      }
      const data = await getRiskAssessment(merchantId);
      if (process.env.NODE_ENV === 'test') {
        console.log('[RiskAssessmentTab] Assessment data received:', data);
      }
      setAssessment(data);
      if (process.env.NODE_ENV === 'test') {
        console.log('[RiskAssessmentTab] State updated with assessment data');
      }

      // Load risk history if assessment exists (optional - endpoint may not be implemented)
      if (data) {
        try {
          setHistoryLoading(true);
          const history = await getRiskHistory(merchantId, 10, 0);
          setRiskHistory(history.history || []);
        } catch (historyErr) {
          // Silently handle 404s for optional endpoints - don't log to console
          const is404 = historyErr instanceof Error && historyErr.message.includes('404');
          if (!is404) {
          console.error('Failed to load risk history:', historyErr);
          }
          // Don't fail the whole component if history fails
        } finally {
          setHistoryLoading(false);
        }

        // Load risk predictions
        try {
          const predictions = await getRiskPredictions(merchantId, [3, 6, 12], false, true);
          setRiskPredictions(predictions);
        } catch (predictionsErr) {
          console.error('Failed to load risk predictions:', predictionsErr);
          // Don't fail the whole component if predictions fail
        }
      }
    } catch (err) {
      if (process.env.NODE_ENV === 'test') {
        console.error('[RiskAssessmentTab] Error loading assessment:', err);
      }
      setError(err instanceof Error ? err.message : 'Failed to load risk assessment');
    } finally {
      setLoading(false);
      setIsRefreshing(false);
      setLastRefreshTime(new Date());
      if (process.env.NODE_ENV === 'test') {
        console.log('[RiskAssessmentTab] Loading complete');
      }
    }
  }

  const handleRefresh = () => {
    loadAssessment(true);
  };

  // Format relative time for last refresh
  const formatRelativeTime = (date: Date): string => {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffSecs < 60) return 'just now';
    if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
    if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
    if (diffDays < 7) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
    return date.toLocaleDateString();
  };

  async function handleStartAssessment() {
    try {
      setProcessing(true);
      setError(null);
      toast.info('Starting risk assessment...');
      const request: RiskAssessmentRequest = {
        merchantId,
        options: {
          includeHistory: true,
          includePredictions: true,
        },
      };
      const response = await startRiskAssessment(request);
      toast.success('Risk assessment started successfully');
      
      // Poll for status
      const pollInterval = setInterval(async () => {
        try {
          const status = await getAssessmentStatus(response.assessmentId);
          if (status.status === 'completed') {
            clearInterval(pollInterval);
            await loadAssessment();
            setProcessing(false);
            toast.success('Risk assessment completed');
          } else if (status.status === 'failed') {
            clearInterval(pollInterval);
            setProcessing(false);
            toast.error('Risk assessment failed');
          }
        } catch (err) {
          clearInterval(pollInterval);
          setProcessing(false);
          await ErrorHandler.handleAPIError(err);
        }
      }, 2000);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to start assessment';
      setError(errorMessage);
      setProcessing(false);
      await ErrorHandler.handleAPIError(err);
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error && !assessment) {
    return (
      <EmptyState
        type="error"
        title="Error Loading Assessment"
        message={error}
        actionLabel="Retry"
        onAction={loadAssessment}
      />
    );
  }

  if (!assessment && !processing) {
    return (
      <EmptyState
        type="noData"
        title="No Risk Assessment"
        message="No risk assessment has been performed for this merchant yet."
        actionLabel="Start Assessment"
        onAction={handleStartAssessment}
      />
    );
  }

  return (
    <section className="space-y-6" aria-labelledby="risk-assessment-heading">
      {/* Header with refresh button */}
      <header className="flex items-center justify-between">
        <div>
          <h2 id="risk-assessment-heading" className="text-2xl font-bold">Risk Assessment</h2>
          {lastRefreshTime && (
            <p className="text-sm text-muted-foreground mt-1" aria-live="polite">
              Updated {formatRelativeTime(lastRefreshTime)}
            </p>
          )}
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleRefresh}
          disabled={isRefreshing || loading || processing}
          aria-label="Refresh risk assessment data"
          title="Refresh data (R)"
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${isRefreshing ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </header>

      {processing && (
        <Card>
          <CardHeader>
            <CardTitle>Processing Assessment</CardTitle>
            <CardDescription>Risk assessment is being processed...</CardDescription>
          </CardHeader>
          <CardContent>
            <ProgressIndicator progress={50} label="Assessment Progress" />
          </CardContent>
        </Card>
      )}

      {assessment && (
        <>
          {assessment.result && (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <ChartContainer
                title="Risk Score Gauge"
                description="Overall risk assessment score"
                isLoading={false}
              >
                <RiskGauge
                  value={assessment.result?.overallScore ?? 0}
                  max={10}
                  height={300}
                  width={300}
                  isLoading={false}
                  showNeedle={true}
                  label="Risk Score"
                />
              </ChartContainer>

              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle>Risk Assessment Details</CardTitle>
                      <CardDescription>Current risk assessment results</CardDescription>
                    </div>
                    <ExportButton
                      data={async () => ({
                        assessment,
                        merchantId,
                        exportedAt: new Date().toISOString(),
                      })}
                      exportType="risk"
                      merchantId={merchantId}
                      formats={['csv', 'json', 'excel', 'pdf']}
                    />
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Status</p>
                    <Badge variant="outline">{assessment.status}</Badge>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Overall Score</p>
                    <p className="text-2xl font-bold">{formatNumber(assessment.result?.overallScore, 1)}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Risk Level</p>
                    <Badge
                      variant={
                        assessment.result.riskLevel === 'low'
                          ? 'default'
                          : assessment.result.riskLevel === 'medium'
                          ? 'secondary'
                          : 'destructive'
                      }
                    >
                      {assessment.result.riskLevel}
                    </Badge>
                  </div>
                  {assessment.result.factors && assessment.result.factors.length > 0 && (
                    <div>
                      <p className="text-sm font-medium text-muted-foreground mb-2">Risk Factors</p>
                      <div className="space-y-2">
                        {assessment.result.factors.map((factor, index) => (
                          <div key={index} className="flex justify-between items-center">
                            <span>{factor.name}</span>
                            <div className="flex items-center gap-2">
                              <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                                <div
                                  className="h-full bg-primary"
                                  style={{ width: `${(factor.score / 10) * 100}%` }}
                                />
                              </div>
                              <span className="text-sm font-medium w-12 text-right">
                                {formatNumber(factor.score, 1)}
                              </span>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          )}

          {(!assessment.result || assessment.status === 'pending' || assessment.status === 'processing') && (
            <Card>
              <CardHeader>
                <CardTitle>Risk Assessment</CardTitle>
                <CardDescription>Current risk assessment status</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Status</p>
                  <Badge variant="outline">{assessment.status}</Badge>
                </div>
                {(assessment.status === 'pending' || assessment.status === 'processing') && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground mb-2">Progress</p>
                    <ProgressIndicator progress={assessment.progress} label="Assessment Progress" />
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Industry Benchmark Comparison */}
          <RiskBenchmarkComparison merchantId={merchantId} />

          {/* Risk Assessment Explainability */}
          <RiskExplainabilitySection merchantId={merchantId} />

          {/* Risk Recommendations */}
          <RiskRecommendationsSection merchantId={merchantId} />

          {/* Risk Factors Table */}
          {assessment.result?.factors && assessment.result.factors.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Risk Factors</CardTitle>
                <CardDescription>Detailed risk factor analysis</CardDescription>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Factor</TableHead>
                      <TableHead>Score</TableHead>
                      <TableHead>Weight</TableHead>
                      <TableHead className="text-right">Impact</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {assessment.result.factors
                      .sort((a, b) => b.score - a.score)
                      .map((factor, index) => (
                        <TableRow key={index}>
                          <TableCell className="font-medium">{factor.name}</TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <div className="w-32 h-2 bg-muted rounded-full overflow-hidden">
                                <div
                                  className="h-full bg-primary"
                                  style={{ width: `${(factor.score / 10) * 100}%` }}
                                />
                              </div>
                              <span className="text-sm font-medium w-12">
                                {formatNumber(factor.score, 1)}/10
                              </span>
                            </div>
                          </TableCell>
                          <TableCell>{formatPercent(factor.weight)}</TableCell>
                          <TableCell className="text-right">
                            {formatNumber(factor.score * factor.weight, 2)}
                          </TableCell>
                        </TableRow>
                      ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* Risk History Table and Charts */}
          {(riskHistory.length > 0 || riskPredictions) && (
            <div className="space-y-6">
              {riskHistory.length > 0 && (
                <>
                  <Card>
                    <CardHeader>
                      <CardTitle>Risk History</CardTitle>
                      <CardDescription>Historical risk assessment scores</CardDescription>
                    </CardHeader>
                    <CardContent>
                      {historyLoading ? (
                        <Skeleton className="h-64 w-full" />
                      ) : (
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead>Date</TableHead>
                              <TableHead>Status</TableHead>
                              <TableHead>Score</TableHead>
                              <TableHead>Risk Level</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {riskHistory.map((historyItem) => (
                              <TableRow key={historyItem.id}>
                                <TableCell suppressHydrationWarning>
                                  {mounted ? new Date(historyItem.createdAt).toLocaleDateString() : 'Loading...'}
                                </TableCell>
                                <TableCell>
                                  <Badge variant="outline">{historyItem.status}</Badge>
                                </TableCell>
                                <TableCell>
                                  {formatNumber(historyItem.result?.overallScore, 1)}
                                </TableCell>
                                <TableCell>
                                  <Badge
                                    variant={
                                      historyItem.result?.riskLevel === 'low'
                                        ? 'default'
                                        : historyItem.result?.riskLevel === 'medium'
                                        ? 'secondary'
                                        : 'destructive'
                                    }
                                  >
                                    {historyItem.result?.riskLevel || 'N/A'}
                                  </Badge>
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      )}
                    </CardContent>
                  </Card>

                  {/* Risk Score History Chart */}
                  <ChartContainer
                    title="Risk Score History"
                    description="Risk assessment scores over time"
                    isLoading={historyLoading}
                  >
                    <LineChart
                      data={mounted && riskHistory.length > 0 ? riskHistory
                        .filter((h) => h.result?.overallScore != null)
                        .map((h) => {
                          // Format date on client side only
                          const dateStr = mounted 
                            ? new Date(h.createdAt).toLocaleDateString()
                            : 'Loading...';
                          return {
                            name: dateStr,
                            value: h.result?.overallScore ?? 0,
                          };
                        }) : []}
                      dataKey="value"
                      lines={[{ key: 'value', name: 'Risk Score', color: '#8884d8' }]}
                      height={300}
                      isLoading={historyLoading}
                    />
                  </ChartContainer>
                </>
              )}

              {/* Risk Factors Comparison Chart */}
              {assessment.result?.factors && assessment.result.factors.length > 0 && (
                <ChartContainer
                  title="Risk Factors Comparison"
                  description="Comparison of risk factor scores"
                  isLoading={false}
                >
                  <BarChart
                    data={assessment.result.factors
                      .sort((a, b) => b.score - a.score)
                      .map((factor) => ({
                        name: factor.name,
                        value: factor.score,
                      }))}
                    dataKey="value"
                    bars={[{ key: 'value', name: 'Score', color: '#8884d8' }]}
                    height={300}
                    isLoading={false}
                  />
                </ChartContainer>
              )}

              {/* Risk Predictions Chart */}
              {riskPredictions && riskPredictions.predictions && riskPredictions.predictions.length > 0 && (
                <ChartContainer
                  title="Risk Predictions"
                  description="Predicted risk scores over time"
                  isLoading={false}
                >
                  <AreaChart
                    data={riskPredictions.predictions.map((pred: any) => ({
                      name: `${pred.months} months`,
                      value: pred.predictedScore || 0,
                    }))}
                    dataKey="name"
                    areas={[{ key: 'value', name: 'Predicted Score', color: '#8884d8', fillOpacity: 0.6 }]}
                    height={300}
                    isLoading={false}
                  />
                </ChartContainer>
              )}
            </div>
          )}
        </>
      )}
    </section>
  );
}

export function RiskAssessmentTab({ merchantId }: RiskAssessmentTabProps) {
  return (
    <RiskWebSocketProvider merchantId={merchantId}>
      <div className="space-y-4">
        <div className="flex items-center justify-end">
          <WebSocketStatusIndicator />
        </div>
        <RiskAssessmentTabContent merchantId={merchantId} />
      </div>
    </RiskWebSocketProvider>
  );
}

