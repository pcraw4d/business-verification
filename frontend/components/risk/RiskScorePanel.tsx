'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { ChevronDown, ChevronUp, Info } from 'lucide-react';
import { formatNumber, formatPercent } from '@/lib/number-format';
import type { RiskAssessment } from '@/types/merchant';

interface RiskScorePanelProps {
  assessment: RiskAssessment | null;
  collapsed?: boolean;
}

export function RiskScorePanel({ assessment, collapsed = true }: RiskScorePanelProps) {
  if (!assessment || !assessment.result) {
    return null;
  }

  const riskLevel = assessment.result.riskLevel.toLowerCase();
  const badgeVariant =
    riskLevel === 'low'
      ? 'default'
      : riskLevel === 'medium'
        ? 'secondary'
        : riskLevel === 'high'
          ? 'destructive'
          : 'destructive';

  return (
    <Card>
      <Collapsible defaultOpen={!collapsed}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Info className="h-5 w-5" />
              <CardTitle>Why This Score?</CardTitle>
            </div>
            <CollapsibleTrigger asChild>
              <button className="p-2 hover:bg-accent rounded-md" aria-label="Toggle risk score breakdown">
                <ChevronDown className="h-4 w-4" />
              </button>
            </CollapsibleTrigger>
          </div>
          <CardDescription>Risk score breakdown and factors</CardDescription>
        </CardHeader>
        <CollapsibleContent>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground mb-2">Overall Score</p>
              <div className="flex items-center gap-4">
                <p className="text-3xl font-bold">{formatNumber(assessment.result?.overallScore, 1)}</p>
                <Badge variant={badgeVariant}>{assessment.result.riskLevel}</Badge>
              </div>
            </div>

            {assessment.result.factors && assessment.result.factors.length > 0 && (
              <div>
                <p className="text-sm font-medium text-muted-foreground mb-3">Risk Factors</p>
                <div className="space-y-2">
                  {assessment.result.factors.map((factor, index) => (
                    <div key={index} className="flex justify-between items-center p-2 bg-muted rounded-md">
                      <span className="text-sm">{factor.name}</span>
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">{formatNumber(factor.score, 1)}</span>
                        <span className="text-xs text-muted-foreground">
                          (weight: {formatNumber(factor.weight, 2)})
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  );
}

