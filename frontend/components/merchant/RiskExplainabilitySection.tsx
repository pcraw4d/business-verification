'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { AlertCircle, BarChart3, Info, RefreshCw } from 'lucide-react';
import { useEffect, useState, useCallback, useMemo } from 'react';
import { explainRiskAssessment, getRiskAssessment } from '@/lib/api';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { BarChart } from '@/components/charts/lazy';

interface RiskExplainabilitySectionProps {
  merchantId: string;
}

type RiskExplanationResponse = {
  assessmentId: string;
  factors: Array<{ name: string; score: number; weight: number }>;
  shapValues: Record<string, number>;
  baseValue: number;
  prediction?: number;
};

export function RiskExplainabilitySection({ merchantId }: RiskExplainabilitySectionProps) {
  const [explanation, setExplanation] = useState<RiskExplanationResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [assessmentId, setAssessmentId] = useState<string | null>(null);

  const fetchExplanation = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      // First, get the current risk assessment to get the assessment ID
      let currentAssessmentId = assessmentId;
      if (!currentAssessmentId) {
        const assessment = await getRiskAssessment(merchantId);
        if (assessment && assessment.id) {
          currentAssessmentId = assessment.id;
          setAssessmentId(currentAssessmentId);
        } else {
          setError('No risk assessment found. Please run a risk assessment first.');
          setLoading(false);
          return;
        }
      }

      if (!currentAssessmentId) {
        setError('Assessment ID not available');
        setLoading(false);
        return;
      }

      // Fetch the explanation
      const data = await explainRiskAssessment(currentAssessmentId);
      setExplanation(data);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load risk explanation';
      setError(errorMessage);
      toast.error('Failed to load explanation', {
        description: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  }, [merchantId, assessmentId]);

  useEffect(() => {
    fetchExplanation();
  }, [fetchExplanation]);

  // Prepare SHAP values chart data
  const shapChartData = useMemo(() => {
    if (!explanation?.shapValues) return [];
    
    return Object.entries(explanation.shapValues)
      .map(([name, value]) => ({
        name,
        value: Math.abs(value), // Use absolute value for visualization
        originalValue: value, // Keep original for display
      }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 10); // Top 10 features
  }, [explanation]);

  // Prepare feature importance chart data (from factors)
  const featureImportanceData = useMemo(() => {
    if (!explanation?.factors) return [];
    
    return explanation.factors
      .map((factor) => ({
        name: factor.name,
        value: factor.score * factor.weight, // Impact = score * weight
        score: factor.score,
        weight: factor.weight,
      }))
      .sort((a, b) => b.value - a.value);
  }, [explanation]);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Assessment Explainability</CardTitle>
          <CardDescription>SHAP values and feature importance analysis</CardDescription>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-64 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle>Risk Assessment Explainability</CardTitle>
          <CardDescription>SHAP values and feature importance analysis</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center p-6 space-y-4">
            <AlertCircle className="h-8 w-8 text-destructive" />
            <p className="text-sm text-destructive-foreground">{error}</p>
            <Button onClick={fetchExplanation} variant="outline" size="sm">
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!explanation) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Assessment Explainability</CardTitle>
          <CardDescription>SHAP values and feature importance analysis</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center p-6 space-y-4">
            <Info className="h-8 w-8 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              No explanation data available. Please run a risk assessment first.
            </p>
            <Button onClick={fetchExplanation} variant="outline" size="sm">
              <RefreshCw className="h-4 w-4 mr-2" />
              Reload
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Risk Assessment Explainability</CardTitle>
            <CardDescription>
              SHAP (SHapley Additive exPlanations) values and feature importance
            </CardDescription>
          </div>
          <Button onClick={fetchExplanation} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Summary Metrics */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="p-4 border rounded-lg">
            <p className="text-xs text-muted-foreground">Base Value</p>
            <p className="text-2xl font-bold">{explanation.baseValue.toFixed(3)}</p>
          </div>
          {explanation.prediction !== undefined && (
            <div className="p-4 border rounded-lg">
              <p className="text-xs text-muted-foreground">Prediction</p>
              <p className="text-2xl font-bold">{(explanation.prediction * 100).toFixed(1)}%</p>
            </div>
          )}
          <div className="p-4 border rounded-lg">
            <p className="text-xs text-muted-foreground">Features Analyzed</p>
            <p className="text-2xl font-bold">
              {Object.keys(explanation.shapValues || {}).length}
            </p>
          </div>
          <div className="p-4 border rounded-lg">
            <p className="text-xs text-muted-foreground">Risk Factors</p>
            <p className="text-2xl font-bold">{explanation.factors?.length || 0}</p>
          </div>
        </div>

        {/* SHAP Values Chart */}
        {shapChartData.length > 0 && (
          <ChartContainer
            title="SHAP Values (Top 10 Features)"
            description="Feature contributions to the risk score prediction"
            isLoading={false}
          >
            <BarChart
              data={shapChartData}
              dataKey="name"
              bars={[
                {
                  key: 'value',
                  name: 'SHAP Value',
                  color: '#8884d8',
                },
              ]}
              height={400}
              isLoading={false}
            />
          </ChartContainer>
        )}

        {/* Feature Importance Chart */}
        {featureImportanceData.length > 0 && (
          <ChartContainer
            title="Feature Importance"
            description="Impact of each risk factor (score × weight)"
            isLoading={false}
          >
            <BarChart
              data={featureImportanceData}
              dataKey="name"
              bars={[
                {
                  key: 'value',
                  name: 'Impact',
                  color: '#82ca9d',
                },
              ]}
              height={400}
              isLoading={false}
            />
          </ChartContainer>
        )}

        {/* SHAP Values Table */}
        {Object.keys(explanation.shapValues || {}).length > 0 && (
          <div className="space-y-2">
            <h3 className="text-lg font-semibold">SHAP Values</h3>
            <div className="border rounded-lg">
              <div className="max-h-64 overflow-y-auto">
                <div className="divide-y">
                  {Object.entries(explanation.shapValues)
                    .sort(([, a], [, b]) => Math.abs(b) - Math.abs(a))
                    .map(([feature, value]) => (
                      <div key={feature} className="p-3 flex items-center justify-between">
                        <span className="text-sm font-medium">{feature}</span>
                        <div className="flex items-center gap-2">
                          <div className="w-32 h-2 bg-muted rounded-full overflow-hidden">
                            <div
                              className={`h-full ${
                                value > 0 ? 'bg-red-500' : 'bg-green-500'
                              }`}
                              style={{ width: `${Math.min(100, Math.abs(value) * 100)}%` }}
                            />
                          </div>
                          <Badge variant={value > 0 ? 'destructive' : 'default'}>
                            {value > 0 ? '+' : ''}
                            {value.toFixed(3)}
                          </Badge>
                        </div>
                      </div>
                    ))}
                </div>
              </div>
            </div>
            <p className="text-xs text-muted-foreground">
              Positive values increase risk, negative values decrease risk
            </p>
          </div>
        )}

        {/* Feature Importance Details */}
        {explanation.factors && explanation.factors.length > 0 && (
          <div className="space-y-2">
            <h3 className="text-lg font-semibold">Risk Factors Impact</h3>
            <div className="space-y-2">
              {explanation.factors
                .sort((a, b) => b.score * b.weight - a.score * a.weight)
                .map((factor, index) => (
                  <div key={index} className="p-3 border rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium">{factor.name}</span>
                      <Badge variant="outline">
                        Impact: {(factor.score * factor.weight).toFixed(3)}
                      </Badge>
                    </div>
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <span className="text-muted-foreground">Score: </span>
                        <span className="font-medium">{factor.score.toFixed(3)}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Weight: </span>
                        <span className="font-medium">{(factor.weight * 100).toFixed(1)}%</span>
                      </div>
                    </div>
                  </div>
                ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

