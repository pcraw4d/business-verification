'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { getRiskAssessment, startRiskAssessment, getAssessmentStatus } from '@/lib/api';
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
      const data = await getRiskAssessment(merchantId);
      setAssessment(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load risk assessment');
    } finally {
      setLoading(false);
    }
  }

  async function handleStartAssessment() {
    try {
      setProcessing(true);
      setError(null);
      const request: RiskAssessmentRequest = {
        merchantId,
        options: {
          includeHistory: true,
          includePredictions: true,
        },
      };
      const response = await startRiskAssessment(request);
      
      // Poll for status
      const pollInterval = setInterval(async () => {
        try {
          const status = await getAssessmentStatus(response.assessmentId);
          if (status.status === 'completed' || status.status === 'failed') {
            clearInterval(pollInterval);
            await loadAssessment();
            setProcessing(false);
          }
        } catch (err) {
          clearInterval(pollInterval);
          setProcessing(false);
        }
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start assessment');
      setProcessing(false);
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {!assessment && (
        <Card>
          <CardHeader>
            <CardTitle>Risk Assessment</CardTitle>
            <CardDescription>No assessment available</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={handleStartAssessment} disabled={processing}>
              {processing ? 'Processing...' : 'Start Risk Assessment'}
            </Button>
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

