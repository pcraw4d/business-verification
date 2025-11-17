'use client';

import { RiskGauge } from '@/components/charts/lazy';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { ExportButton } from '@/components/export/ExportButton';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ProgressIndicator } from '@/components/ui/progress-indicator';
import { Skeleton } from '@/components/ui/skeleton';
import { RiskWebSocketProvider, useRiskWebSocket, WebSocketStatusIndicator } from '@/components/websocket/RiskWebSocketProvider';
import { getAssessmentStatus, getRiskAssessment, startRiskAssessment } from '@/lib/api';
import { ErrorHandler } from '@/lib/error-handler';
import type { RiskAssessment, RiskAssessmentRequest } from '@/types/merchant';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';

interface RiskAssessmentTabProps {
  merchantId: string;
}

function RiskAssessmentTabContent({ merchantId }: RiskAssessmentTabProps) {
  useRiskWebSocket(); // Initialize WebSocket connection
  const [assessment, setAssessment] = useState<RiskAssessment | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [processing, setProcessing] = useState(false);

  useEffect(() => {
    loadAssessment();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [merchantId]);

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

  async function loadAssessment() {
    try {
      setLoading(true);
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
    } catch (err) {
      if (process.env.NODE_ENV === 'test') {
        console.error('[RiskAssessmentTab] Error loading assessment:', err);
      }
      setError(err instanceof Error ? err.message : 'Failed to load risk assessment');
    } finally {
      setLoading(false);
      if (process.env.NODE_ENV === 'test') {
        console.log('[RiskAssessmentTab] Loading complete');
      }
    }
  }

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
    <div className="space-y-6">
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
                  value={assessment.result.overallScore}
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
                    <p className="text-2xl font-bold">{assessment.result.overallScore.toFixed(1)}</p>
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
                                {factor.score.toFixed(1)}
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
        </>
      )}
    </div>
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

