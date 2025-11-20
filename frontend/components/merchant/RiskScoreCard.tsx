'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { getMerchantRiskScore } from '@/lib/api';
import type { MerchantRiskScore } from '@/types/merchant';
import { useEffect, useState } from 'react';
import { AlertCircle, RefreshCw, Shield, TrendingUp, TrendingDown } from 'lucide-react';
import { toast } from 'sonner';
import { formatPercent } from '@/lib/number-format';

interface RiskScoreCardProps {
  merchantId: string;
}

export function RiskScoreCard({ merchantId }: RiskScoreCardProps) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [riskScore, setRiskScore] = useState<MerchantRiskScore | null>(null);

  const fetchRiskScore = async () => {
    try {
      setLoading(true);
      setError(null);

      const data = await getMerchantRiskScore(merchantId);
      setRiskScore(data);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load risk score';
      setError(errorMessage);
      toast.error('Failed to load risk score', {
        description: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRiskScore();
  }, [merchantId]);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Score</CardTitle>
          <CardDescription>Current merchant risk assessment</CardDescription>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-24 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Score</CardTitle>
          <CardDescription>Current merchant risk assessment</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center p-6 space-y-4">
            <AlertCircle className="h-8 w-8 text-destructive" />
            <p className="text-sm text-muted-foreground">{error}</p>
            <Button variant="outline" size="sm" onClick={fetchRiskScore}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!riskScore) {
    return null;
  }

  const getRiskLevelBadge = () => {
    const level = riskScore.risk_level;
    if (level === 'low') {
      return (
        <Badge variant="default" className="flex items-center gap-1 w-fit">
          <Shield className="h-3 w-3" />
          Low Risk
        </Badge>
      );
    } else if (level === 'medium') {
      return (
        <Badge variant="secondary" className="flex items-center gap-1 w-fit">
          <AlertCircle className="h-3 w-3" />
          Medium Risk
        </Badge>
      );
    } else {
      return (
        <Badge variant="destructive" className="flex items-center gap-1 w-fit">
          <AlertCircle className="h-3 w-3" />
          High Risk
        </Badge>
      );
    }
  };

  const getScoreColor = () => {
    const score = riskScore.risk_score;
    if (score == null) return 'text-gray-600';
    if (score < 0.3) return 'text-green-600';
    if (score < 0.7) return 'text-yellow-600';
    return 'text-red-600';
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Risk Score</CardTitle>
        <CardDescription>Current merchant risk assessment</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Risk Score</p>
            <p className={`text-3xl font-bold ${getScoreColor()}`}>
              {formatPercent(riskScore.risk_score)}
            </p>
          </div>
          <div>{getRiskLevelBadge()}</div>
        </div>

        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">Confidence</span>
            <span className="text-sm font-medium">
              {formatPercent(riskScore.confidence_score)}
            </span>
          </div>
          {riskScore.assessment_date && (
            <div className="flex justify-between items-center">
              <span className="text-sm text-muted-foreground">Assessment Date</span>
              <span className="text-sm font-medium">
                {new Date(riskScore.assessment_date).toLocaleDateString()}
              </span>
            </div>
          )}
        </div>

        {riskScore.factors && riskScore.factors.length > 0 && (
          <div className="pt-4 border-t">
            <p className="text-sm font-medium mb-2">Key Risk Factors</p>
            <div className="space-y-1">
              {riskScore.factors.slice(0, 3).map((factor, index) => (
                <div key={index} className="flex justify-between items-center text-sm">
                  <span className="text-muted-foreground">{factor.category}</span>
                  <span className="font-medium">{formatPercent(factor.score)}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

