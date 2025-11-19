'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { getPortfolioStatistics, getMerchantRiskScore } from '@/lib/api';
import type { PortfolioStatistics, MerchantRiskScore, PortfolioComparison } from '@/types/merchant';
import { useEffect, useState } from 'react';
import { TrendingUp, TrendingDown, Minus, RefreshCw, AlertCircle } from 'lucide-react';
import { toast } from 'sonner';

interface PortfolioComparisonCardProps {
  merchantId: string;
  merchantRiskLevel?: string;
}

export function PortfolioComparisonCard({ merchantId, merchantRiskLevel }: PortfolioComparisonCardProps) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [comparison, setComparison] = useState<PortfolioComparison | null>(null);
  const [portfolioStats, setPortfolioStats] = useState<PortfolioStatistics | null>(null);
  const [merchantRiskScore, setMerchantRiskScore] = useState<MerchantRiskScore | null>(null);

  const fetchComparisonData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch portfolio statistics and merchant risk score in parallel
      const [statsResult, riskScoreResult] = await Promise.allSettled([
        getPortfolioStatistics(),
        getMerchantRiskScore(merchantId),
      ]);

      if (statsResult.status === 'fulfilled') {
        setPortfolioStats(statsResult.value);
      }

      if (riskScoreResult.status === 'fulfilled') {
        setMerchantRiskScore(riskScoreResult.value);
      }

      // Calculate comparison if we have both data
      if (statsResult.status === 'fulfilled' && riskScoreResult.status === 'fulfilled') {
        const stats = statsResult.value;
        const riskScore = riskScoreResult.value;

        const merchantScore = riskScore.risk_score;
        const portfolioAvg = stats.averageRiskScore;
        
        // Calculate percentile (simplified - assumes normal distribution)
        // In a real implementation, you'd calculate this from the actual distribution
        const difference = merchantScore - portfolioAvg;
        const differencePercentage = portfolioAvg > 0 ? (difference / portfolioAvg) * 100 : 0;
        
        // Estimate percentile based on difference (simplified calculation)
        let percentile = 50; // Default to median
        if (difference > 0) {
          // Above average - estimate higher percentile
          percentile = Math.min(95, 50 + (difference / portfolioAvg) * 25);
        } else {
          // Below average - estimate lower percentile
          percentile = Math.max(5, 50 + (difference / portfolioAvg) * 25);
        }

        let position: 'above_average' | 'below_average' | 'average';
        if (differencePercentage > 5) {
          position = 'above_average';
        } else if (differencePercentage < -5) {
          position = 'below_average';
        } else {
          position = 'average';
        }

        setComparison({
          merchantScore,
          portfolioAverage: portfolioAvg,
          portfolioMedian: portfolioAvg, // Use average as median estimate
          percentile,
          position,
          difference,
          differencePercentage,
        });
      } else {
        // If we only have portfolio stats, create a basic comparison
        if (statsResult.status === 'fulfilled') {
          const stats = statsResult.value;
          // Use merchant risk level to estimate score if risk score not available
          let estimatedScore = stats.averageRiskScore;
          if (merchantRiskLevel === 'low') {
            estimatedScore = stats.averageRiskScore * 0.7;
          } else if (merchantRiskLevel === 'high') {
            estimatedScore = stats.averageRiskScore * 1.3;
          }

          setComparison({
            merchantScore: estimatedScore,
            portfolioAverage: stats.averageRiskScore,
            portfolioMedian: stats.averageRiskScore,
            percentile: 50,
            position: 'average',
            difference: estimatedScore - stats.averageRiskScore,
            differencePercentage: ((estimatedScore - stats.averageRiskScore) / stats.averageRiskScore) * 100,
          });
        }
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load portfolio comparison';
      setError(errorMessage);
      toast.error('Failed to load comparison data', {
        description: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchComparisonData();
  }, [merchantId, merchantRiskLevel]);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Comparison</CardTitle>
          <CardDescription>Comparing merchant performance to portfolio</CardDescription>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-32 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Comparison</CardTitle>
          <CardDescription>Comparing merchant performance to portfolio</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center p-6 space-y-4">
            <AlertCircle className="h-8 w-8 text-destructive" />
            <p className="text-sm text-muted-foreground">{error}</p>
            <Button variant="outline" size="sm" onClick={fetchComparisonData}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!comparison || !portfolioStats) {
    return null;
  }

  const getPositionBadge = () => {
    if (comparison.position === 'above_average') {
      return (
        <Badge variant="destructive" className="flex items-center gap-1">
          <TrendingUp className="h-3 w-3" />
          Above Average
        </Badge>
      );
    } else if (comparison.position === 'below_average') {
      return (
        <Badge variant="default" className="flex items-center gap-1">
          <TrendingDown className="h-3 w-3" />
          Below Average
        </Badge>
      );
    } else {
      return (
        <Badge variant="secondary" className="flex items-center gap-1">
          <Minus className="h-3 w-3" />
          Average
        </Badge>
      );
    }
  };

  const getPercentileLabel = () => {
    if (comparison.percentile >= 90) return 'Top 10%';
    if (comparison.percentile >= 75) return 'Top 25%';
    if (comparison.percentile >= 50) return 'Top 50%';
    if (comparison.percentile >= 25) return 'Bottom 50%';
    if (comparison.percentile >= 10) return 'Bottom 25%';
    return 'Bottom 10%';
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Portfolio Comparison</CardTitle>
        <CardDescription>Comparing merchant performance to portfolio average</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Merchant Risk Score</p>
            <p className="text-2xl font-bold">{comparison.merchantScore.toFixed(2)}</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Portfolio Average</p>
            <p className="text-2xl font-bold">{comparison.portfolioAverage.toFixed(2)}</p>
          </div>
        </div>

        <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
          <div>
            <p className="text-sm font-medium">Position in Portfolio</p>
            <p className="text-lg font-semibold">{getPercentileLabel()}</p>
            <p className="text-xs text-muted-foreground">
              {comparison.percentile.toFixed(1)}th percentile
            </p>
          </div>
          <div>{getPositionBadge()}</div>
        </div>

        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">Difference</span>
            <span
              className={`text-sm font-medium ${
                comparison.difference > 0 ? 'text-destructive' : 'text-green-600'
              }`}
            >
              {comparison.difference > 0 ? '+' : ''}
              {comparison.difference.toFixed(2)} (
              {comparison.differencePercentage > 0 ? '+' : ''}
              {comparison.differencePercentage.toFixed(1)}%)
            </span>
          </div>
          {portfolioStats.totalMerchants > 0 && (
            <div className="flex justify-between items-center">
              <span className="text-sm text-muted-foreground">Portfolio Size</span>
              <span className="text-sm font-medium">{portfolioStats.totalMerchants.toLocaleString()} merchants</span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

