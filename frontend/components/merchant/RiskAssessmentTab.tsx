'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { EmptyState } from '@/components/ui/empty-state';
import { ProgressIndicator } from '@/components/ui/progress-indicator';
import { getRiskAssessment, startRiskAssessment, getAssessmentStatus } from '@/lib/api';
import { ErrorHandler } from '@/lib/error-handler';
import { toast } from 'sonner';
import type { RiskAssessment, RiskAssessmentRequest } from '@/types/merchant';

interface RiskAssessmentTabProps {
  merchantId: string;
}

export function RiskAssessmentTab({ merchantId }: RiskAssessmentTabProps) {
  const [assessment, setAssessment] = useState<RiskAssessment | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [processing, setProcessing] = useState(false);

  useEffect(() => {
    loadAssessment();
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
        <Card>
          <CardHeader>
            <CardTitle>Risk Assessment</CardTitle>
            <CardDescription>Current risk assessment results</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Status</p>
              <Badge variant="outline">{assessment.status}</Badge>
            </div>
            {assessment.result && (
              <>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Overall Score</p>
                  <p className="text-2xl font-bold">{assessment.result.overallScore.toFixed(1)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Risk Level</p>
                  <Badge variant="outline">{assessment.result.riskLevel}</Badge>
                </div>
                {assessment.result.factors && assessment.result.factors.length > 0 && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground mb-2">Risk Factors</p>
                    <div className="space-y-2">
                      {assessment.result.factors.map((factor, index) => (
                        <div key={index} className="flex justify-between">
                          <span>{factor.name}</span>
                          <span>{factor.score.toFixed(1)}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </>
            )}
            {assessment.status === 'pending' || assessment.status === 'processing' ? (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Progress</p>
                <p>{assessment.progress}%</p>
              </div>
            ) : null}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

