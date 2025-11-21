'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { AlertCircle, BarChart3, Info, RefreshCw, HelpCircle, Download } from 'lucide-react';
import { useEffect, useState, useCallback, useMemo } from 'react';
import { explainRiskAssessment, getRiskAssessment, startRiskAssessment } from '@/lib/api';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { BarChart } from '@/components/charts/lazy';
import { formatNumber, formatPercent } from '@/lib/number-format';
import { ErrorCodes, formatErrorWithCode } from '@/lib/error-codes';
import { ExportButton } from '@/components/export/ExportButton';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

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
          setError(formatErrorWithCode(
            'No risk assessment found. Please run a risk assessment first.',
            ErrorCodes.RISK_ASSESSMENT.NOT_FOUND
          ));
          setLoading(false);
          return;
        }
      }

      if (!currentAssessmentId) {
        setError(formatErrorWithCode(
          'Assessment ID not available',
          ErrorCodes.RISK_ASSESSMENT.NOT_FOUND
        ));
        setLoading(false);
        return;
      }

      // Fetch the explanation
      const data = await explainRiskAssessment(currentAssessmentId);
      setExplanation(data);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load risk explanation';
      const formattedError = formatErrorWithCode(
        errorMessage,
        ErrorCodes.RISK_ASSESSMENT.FETCH_ERROR
      );
      setError(formattedError);
      toast.error('Failed to load explanation', {
        description: formattedError,
      });
    } finally {
      setLoading(false);
    }
  }, [merchantId, assessmentId]);

  useEffect(() => {
    fetchExplanation();
  }, [fetchExplanation]);

  // Export data function
  const getExportData = useCallback(() => {
    if (!explanation) return null;
    
    return {
      assessmentId: explanation.assessmentId,
      baseValue: explanation.baseValue,
      prediction: explanation.prediction,
      factors: explanation.factors,
      shapValues: explanation.shapValues,
      exportedAt: new Date().toISOString(),
    };
  }, [explanation]);

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

  const handleStartAssessment = useCallback(async () => {
    try {
      toast.info('Starting risk assessment...');
      const response = await startRiskAssessment({
        merchantId,
        options: {
          includeHistory: true,
          includePredictions: true,
        },
      });
      toast.success('Risk assessment started successfully');
      
      // Poll for completion
      const pollInterval = setInterval(async () => {
        try {
          const assessment = await getRiskAssessment(merchantId);
          if (assessment && assessment.status === 'completed') {
            clearInterval(pollInterval);
            await fetchExplanation();
            toast.success('Risk assessment completed');
          } else if (assessment && assessment.status === 'failed') {
            clearInterval(pollInterval);
            toast.error('Risk assessment failed');
          }
        } catch (err) {
          clearInterval(pollInterval);
          console.error('Error polling assessment status:', err);
        }
      }, 2000);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to start assessment';
      toast.error('Failed to start assessment', {
        description: errorMessage,
      });
    }
  }, [merchantId, fetchExplanation]);

  if (error) {
    const isNoAssessmentError = error.includes('No risk assessment found') || 
                                error.includes('Assessment ID not available');
    
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle>Risk Assessment Explainability</CardTitle>
          <CardDescription>SHAP values and feature importance analysis</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center p-6 space-y-4">
            <AlertCircle className="h-8 w-8 text-destructive" />
            <div className="text-center space-y-2">
              <p className="text-sm font-medium text-destructive-foreground">{error}</p>
              {isNoAssessmentError && (
                <p className="text-xs text-muted-foreground">
                  A risk assessment must be completed before explainability data can be displayed.
                </p>
              )}
            </div>
            <div className="flex gap-2">
              {isNoAssessmentError ? (
                <Button onClick={handleStartAssessment} variant="default" size="sm">
                  Run Risk Assessment
                </Button>
              ) : (
                <Button onClick={fetchExplanation} variant="outline" size="sm">
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Retry
                </Button>
              )}
            </div>
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
    <TooltipProvider>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2">
                <CardTitle>Risk Assessment Explainability</CardTitle>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <HelpCircle className="h-4 w-4 text-muted-foreground cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent className="max-w-xs">
                    <p className="font-medium mb-1">SHAP Values Explained</p>
                    <p className="text-xs">
                      SHAP (SHapley Additive exPlanations) values show how each feature contributes to the risk score prediction. 
                      Positive values increase risk, negative values decrease risk. The magnitude indicates the feature's importance.
                    </p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <CardDescription>
                SHAP (SHapley Additive exPlanations) values and feature importance
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <ExportButton
                data={getExportData}
                exportType="risk"
                merchantId={merchantId}
                formats={['csv', 'json', 'excel', 'pdf']}
              />
              <Button onClick={fetchExplanation} variant="outline" size="sm">
                <RefreshCw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
            </div>
          </div>
        </CardHeader>
      <CardContent className="space-y-6">
        {/* Summary Metrics */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="p-4 border rounded-lg">
            <p className="text-xs text-muted-foreground">Base Value</p>
            <p className="text-2xl font-bold">{formatNumber(explanation.baseValue, 3)}</p>
          </div>
          {explanation.prediction !== undefined && (
            <div className="p-4 border rounded-lg">
              <p className="text-xs text-muted-foreground">Prediction</p>
              <p className="text-2xl font-bold">{formatPercent(explanation.prediction)}</p>
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
            title={
              <div className="flex items-center gap-2">
                <span>SHAP Values (Top 10 Features)</span>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <HelpCircle className="h-3 w-3 text-muted-foreground cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent className="max-w-xs">
                    <p className="font-medium mb-1">Understanding SHAP Values</p>
                    <p className="text-xs">
                      This chart shows the top 10 features that most influence the risk score. 
                      Features are sorted by absolute SHAP value (importance). 
                      Positive values (red) increase risk, negative values (green) decrease risk.
                    </p>
                  </TooltipContent>
                </Tooltip>
              </div>
            }
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
            title={
              <div className="flex items-center gap-2">
                <span>Feature Importance</span>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <HelpCircle className="h-3 w-3 text-muted-foreground cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent className="max-w-xs">
                    <p className="font-medium mb-1">Feature Importance</p>
                    <p className="text-xs">
                      This chart shows the impact of each risk factor, calculated as score × weight. 
                      Higher values indicate factors that have a greater influence on the overall risk assessment.
                    </p>
                  </TooltipContent>
                </Tooltip>
              </div>
            }
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
                            {formatNumber(value, 3)}
                          </Badge>
                        </div>
                      </div>
                    ))}
                </div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <p className="text-xs text-muted-foreground">
                Positive values increase risk, negative values decrease risk
              </p>
              <Tooltip>
                <TooltipTrigger asChild>
                  <HelpCircle className="h-3 w-3 text-muted-foreground cursor-help" />
                </TooltipTrigger>
                <TooltipContent className="max-w-xs">
                  <p className="font-medium mb-1">SHAP Value Interpretation</p>
                  <p className="text-xs">
                    Each SHAP value represents how much a feature pushes the prediction away from the base value. 
                    Positive values (red) push toward higher risk, negative values (green) push toward lower risk.
                  </p>
                </TooltipContent>
              </Tooltip>
            </div>
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
                        Impact: {formatNumber(factor.score * factor.weight, 3)}
                      </Badge>
                    </div>
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <span className="text-muted-foreground">Score: </span>
                        <span className="font-medium">{formatNumber(factor.score, 3)}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Weight: </span>
                        <span className="font-medium">{formatPercent(factor.weight)}</span>
                      </div>
                    </div>
                  </div>
                ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
    </TooltipProvider>
  );
}

