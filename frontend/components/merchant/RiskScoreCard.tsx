'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { getMerchantRiskScore } from '@/lib/api';
import type { MerchantRiskScore } from '@/types/merchant';
import { useEffect, useState } from 'react';
import { AlertCircle, RefreshCw, Shield, TrendingUp, TrendingDown } from 'lucide-react';
import { toast } from 'sonner';
import { formatPercent } from '@/lib/number-format';
import { useRouter } from 'next/navigation';
import { ErrorCodes, formatErrorWithCode } from '@/lib/error-codes';

// Type guard for risk score validation
const isValidRiskScore = (score: unknown): score is number => {
  return typeof score === 'number' && !isNaN(score) && isFinite(score) && score >= 0 && score <= 1;
};

const hasValidMerchantRiskScore = (score: unknown): score is MerchantRiskScore => {
  if (!score || typeof score !== 'object') return false;
  const s = score as Record<string, unknown>;
  return (
    typeof s.risk_level === 'string' &&
    (s.risk_score === undefined || isValidRiskScore(s.risk_score)) &&
    (s.confidence_score === undefined || isValidRiskScore(s.confidence_score))
  );
};

interface RiskScoreCardProps {
  merchantId: string;
}

export function RiskScoreCard({ merchantId }: RiskScoreCardProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [riskScore, setRiskScore] = useState<MerchantRiskScore | null>(null);
  // Use client-side state for dates to avoid hydration issues
  const [formattedDate, setFormattedDate] = useState<string>('');

  const handleStartAssessment = () => {
    router.push(`/merchant-portfolio/${merchantId}?tab=risk`);
  };

  const fetchRiskScore = async () => {
    try {
      setLoading(true);
      setError(null);

      const data = await getMerchantRiskScore(merchantId);
      
      // Development logging
        if (process.env.NODE_ENV === 'development') {
          console.log('[RiskScoreCard] API Response:', {
            data,
            hasRiskScore: hasValidMerchantRiskScore(data),
            riskScoreType: typeof data.risk_score,
            riskLevel: data.risk_level,
          });
        }

      // Validate data structure
      if (hasValidMerchantRiskScore(data)) {
      setRiskScore(data);
        
        if (process.env.NODE_ENV === 'development') {
          console.log('[RiskScoreCard] Risk score loaded:', {
            risk_score: data.risk_score,
            risk_level: data.risk_level,
            confidence_score: data.confidence_score,
          });
        }
      } else {
        const errorMsg = 'Invalid risk score data structure received from API';
        if (process.env.NODE_ENV === 'development') {
          console.error('[RiskScoreCard] Invalid data structure:', data);
        }
        setError(errorMsg);
        toast.error('Invalid risk score data', {
          description: 'The risk score data format is invalid. Please try again.',
        });
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load risk score';
      setError(errorMessage);
      
      if (process.env.NODE_ENV === 'development') {
        console.error('[RiskScoreCard] Error fetching risk score:', err);
      }
      
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

  // Format date on client side only to avoid hydration issues
  useEffect(() => {
    if (riskScore?.assessment_date) {
      setFormattedDate(new Date(riskScore.assessment_date).toLocaleDateString());
    } else {
      setFormattedDate('');
    }
  }, [riskScore?.assessment_date]);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Score</CardTitle>
          <CardDescription>Loading risk assessment...</CardDescription>
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
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error Loading Risk Score</AlertTitle>
            <AlertDescription className="mt-2">
              {formatErrorWithCode(error, ErrorCodes.RISK_SCORE.FETCH_ERROR)}
              <Button 
                onClick={fetchRiskScore} 
                className="mt-3 w-full sm:w-auto"
                variant="outline"
              >
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  if (!riskScore) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Score</CardTitle>
          <CardDescription>Current merchant risk assessment</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>No Risk Assessment</AlertTitle>
            <AlertDescription className="mt-2">
              {formatErrorWithCode(
                'No risk assessment has been completed for this merchant. Start an assessment to view risk analysis.',
                ErrorCodes.RISK_SCORE.NOT_FOUND
              )}
              <Button 
                onClick={handleStartAssessment} 
                className="mt-3 w-full sm:w-auto"
                variant="default"
              >
                <Shield className="h-4 w-4 mr-2" />
                Start Risk Assessment
              </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
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
            {riskScore.risk_score !== undefined && isValidRiskScore(riskScore.risk_score) ? (
            <p className={`text-3xl font-bold ${getScoreColor()}`}>
              {formatPercent(riskScore.risk_score)}
            </p>
            ) : (
              <p className="text-3xl font-bold text-muted-foreground">N/A</p>
            )}
          </div>
          <div>{getRiskLevelBadge()}</div>
        </div>

        <div className="space-y-2">
          {riskScore.confidence_score !== undefined && isValidRiskScore(riskScore.confidence_score) && (
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">Confidence</span>
            <span className="text-sm font-medium">
              {formatPercent(riskScore.confidence_score)}
            </span>
          </div>
          )}
          {riskScore.assessment_date && formattedDate && (
            <div className="flex justify-between items-center">
              <span className="text-sm text-muted-foreground">Assessment Date</span>
              <span className="text-sm font-medium">
                {formattedDate}
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

