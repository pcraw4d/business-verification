'use client';

import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { TrendingUp, TrendingDown, Award, AlertCircle, BarChart3 } from 'lucide-react';
import { useEffect, useState, useCallback } from 'react';
import { getPortfolioStatistics, getMerchantRiskScore } from '@/lib/api';
import type { PortfolioStatistics, MerchantRiskScore } from '@/types/merchant';
import { useMemo } from 'react';

interface PortfolioContextBadgeProps {
  merchantId: string;
  variant?: 'default' | 'compact' | 'detailed';
  className?: string;
}

type Position = 'top_10' | 'top_25' | 'above_average' | 'average' | 'below_average' | 'bottom_25' | 'bottom_10';

export function PortfolioContextBadge({ merchantId, variant = 'default', className = '' }: PortfolioContextBadgeProps) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [percentile, setPercentile] = useState<number | null>(null);
  const [position, setPosition] = useState<Position | null>(null);
  const [merchantScore, setMerchantScore] = useState<number | null>(null);
  const [portfolioAverage, setPortfolioAverage] = useState<number | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const [portfolioStatsResult, merchantRiskScoreResult] = await Promise.allSettled([
        getPortfolioStatistics(),
        getMerchantRiskScore(merchantId),
      ]);

      let portfolioStats: PortfolioStatistics | null = null;
      if (portfolioStatsResult.status === 'fulfilled') {
        portfolioStats = portfolioStatsResult.value;
      } else {
        console.error('Failed to fetch portfolio statistics:', portfolioStatsResult.reason);
        return;
      }

      let merchantRiskScore: MerchantRiskScore | null = null;
      if (merchantRiskScoreResult.status === 'fulfilled') {
        merchantRiskScore = merchantRiskScoreResult.value;
      } else {
        console.error('Failed to fetch merchant risk score:', merchantRiskScoreResult.reason);
        return;
      }

      if (portfolioStats && merchantRiskScore) {
        const score = merchantRiskScore.risk_score * 100; // Convert to percentage
        const avg = portfolioStats.averageRiskScore * 100; // Convert to percentage

        setMerchantScore(score);
        setPortfolioAverage(avg);

        // Calculate percentile (simplified - can be improved with actual distribution)
        let calculatedPercentile = 50; // Default to median
        const difference = score - avg;
        
        if (difference > 0) {
          // Higher risk score (worse) - lower percentile
          calculatedPercentile = 50 - (difference / (100 - avg)) * 50;
        } else if (difference < 0) {
          // Lower risk score (better) - higher percentile
          calculatedPercentile = 50 + (Math.abs(difference) / avg) * 50;
        }
        
        calculatedPercentile = Math.min(100, Math.max(0, calculatedPercentile));
        setPercentile(calculatedPercentile);

        // Determine position
        let calculatedPosition: Position = 'average';
        if (calculatedPercentile >= 90) {
          calculatedPosition = 'top_10';
        } else if (calculatedPercentile >= 75) {
          calculatedPosition = 'top_25';
        } else if (calculatedPercentile >= 55) {
          calculatedPosition = 'above_average';
        } else if (calculatedPercentile >= 45) {
          calculatedPosition = 'average';
        } else if (calculatedPercentile >= 25) {
          calculatedPosition = 'below_average';
        } else if (calculatedPercentile >= 10) {
          calculatedPosition = 'bottom_25';
        } else {
          calculatedPosition = 'bottom_10';
        }
        setPosition(calculatedPosition);
      }
    } catch (err) {
      console.error('Error loading portfolio context:', err);
      setError('Failed to load context');
    } finally {
      setLoading(false);
    }
  }, [merchantId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const getPositionConfig = useMemo(() => {
    if (!position) return null;

    switch (position) {
      case 'top_10':
        return {
          label: 'Top 10%',
          description: 'Lowest Risk',
          variant: 'default' as const,
          icon: <Award className="h-3 w-3" />,
          color: 'text-green-600',
        };
      case 'top_25':
        return {
          label: 'Top 25%',
          description: 'Low Risk',
          variant: 'default' as const,
          icon: <TrendingDown className="h-3 w-3" />,
          color: 'text-green-600',
        };
      case 'above_average':
        return {
          label: 'Above Average',
          description: 'Better than average',
          variant: 'secondary' as const,
          icon: <TrendingDown className="h-3 w-3" />,
          color: 'text-green-500',
        };
      case 'average':
        return {
          label: 'Average',
          description: 'Portfolio average',
          variant: 'secondary' as const,
          icon: <BarChart3 className="h-3 w-3" />,
          color: 'text-gray-500',
        };
      case 'below_average':
        return {
          label: 'Below Average',
          description: 'Worse than average',
          variant: 'secondary' as const,
          icon: <TrendingUp className="h-3 w-3" />,
          color: 'text-yellow-500',
        };
      case 'bottom_25':
        return {
          label: 'Bottom 25%',
          description: 'High Risk',
          variant: 'destructive' as const,
          icon: <TrendingUp className="h-3 w-3" />,
          color: 'text-red-500',
        };
      case 'bottom_10':
        return {
          label: 'Bottom 10%',
          description: 'Highest Risk',
          variant: 'destructive' as const,
          icon: <AlertCircle className="h-3 w-3" />,
          color: 'text-red-600',
        };
      default:
        return null;
    }
  }, [position]);

  if (loading) {
    if (variant === 'compact') {
      return <Skeleton className="h-5 w-20" />;
    }
    return <Skeleton className="h-6 w-24" />;
  }

  if (error || !position || !getPositionConfig) {
    return null; // Fail silently
  }

  if (variant === 'compact') {
    return (
      <Badge variant={getPositionConfig.variant} className={`flex items-center gap-1 w-fit ${className}`}>
        {getPositionConfig.icon}
        <span>{getPositionConfig.label}</span>
      </Badge>
    );
  }

  if (variant === 'detailed') {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <Badge variant={getPositionConfig.variant} className="flex items-center gap-1">
          {getPositionConfig.icon}
          <span>{getPositionConfig.label}</span>
        </Badge>
        {percentile !== null && (
          <span className="text-sm text-muted-foreground">
            {percentile.toFixed(0)}th percentile
          </span>
        )}
        {merchantScore !== null && portfolioAverage !== null && (
          <span className="text-sm text-muted-foreground">
            ({merchantScore.toFixed(1)}% vs {portfolioAverage.toFixed(1)}% avg)
          </span>
        )}
      </div>
    );
  }

  // Default variant
  return (
    <Badge variant={getPositionConfig.variant} className={`flex items-center gap-1 w-fit ${className}`}>
      {getPositionConfig.icon}
      <span>{getPositionConfig.label}</span>
      {percentile !== null && (
        <span className="text-xs opacity-75">({percentile.toFixed(0)}th)</span>
      )}
    </Badge>
  );
}

