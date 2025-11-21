'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { getPortfolioStatistics, getMerchantRiskScore } from '@/lib/api';
import type { PortfolioStatistics, MerchantRiskScore, PortfolioComparison } from '@/types/merchant';
import { useEffect, useState, useCallback } from 'react';
import { TrendingUp, TrendingDown, Minus, RefreshCw, AlertCircle, Shield } from 'lucide-react';
import { toast } from 'sonner';
import { formatNumber, formatPercentile, formatPercentWithSign } from '@/lib/number-format';
import { useRouter } from 'next/navigation';
import { ErrorCodes, formatErrorWithCode } from '@/lib/error-codes';

// Simple relative time formatter
function formatRelativeTime(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) return 'just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
  return date.toLocaleDateString();
}

interface PortfolioComparisonCardProps {
  merchantId: string;
  merchantRiskLevel?: string;
}

// Type guards for validation
const isValidRiskScore = (score: unknown): score is number => {
  return typeof score === 'number' && !isNaN(score) && isFinite(score) && score >= 0 && score <= 1;
};

const hasValidPortfolioStats = (stats: unknown): stats is PortfolioStatistics => {
  if (!stats || typeof stats !== 'object') return false;
  const s = stats as Record<string, unknown>;
  return typeof s.averageRiskScore === 'number' && !isNaN(s.averageRiskScore) && isFinite(s.averageRiskScore);
};

const hasValidMerchantRiskScore = (score: unknown): score is MerchantRiskScore => {
  if (!score || typeof score !== 'object') return false;
  const s = score as Record<string, unknown>;
  return isValidRiskScore(s.risk_score);
};

export function PortfolioComparisonCard({ merchantId, merchantRiskLevel }: PortfolioComparisonCardProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [comparison, setComparison] = useState<PortfolioComparison | null>(null);
  const [portfolioStats, setPortfolioStats] = useState<PortfolioStatistics | null>(null);
  const [merchantRiskScore, setMerchantRiskScore] = useState<MerchantRiskScore | null>(null);
  const [hasRiskScore, setHasRiskScore] = useState(false);
  const [hasPortfolioStats, setHasPortfolioStats] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [formattedTotalMerchants, setFormattedTotalMerchants] = useState<string>('');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastRefreshTime, setLastRefreshTime] = useState<Date | null>(null);
  const [formattedLastRefresh, setFormattedLastRefresh] = useState<string>('');

  const handleStartAssessment = () => {
    router.push(`/merchant-portfolio/${merchantId}?tab=risk`);
  };

  const fetchComparisonData = useCallback(async (bypassCache = false) => {
    try {
      if (!bypassCache) {
        setLoading(true);
      } else {
        setIsRefreshing(true);
      }
      setError(null);

      // Fetch portfolio statistics and merchant risk score in parallel
      // If bypassing cache, clear cache first
      if (bypassCache) {
        // Clear cache by using a unique cache key or clearing the cache
        // The API cache will be bypassed by making fresh requests
      }
      
      const [statsResult, riskScoreResult] = await Promise.allSettled([
        getPortfolioStatistics(),
        getMerchantRiskScore(merchantId),
      ]);

      // Development logging
      if (process.env.NODE_ENV === 'development') {
        console.log('[PortfolioComparison] API Results:', {
          statsStatus: statsResult.status,
          riskScoreStatus: riskScoreResult.status,
          statsValue: statsResult.status === 'fulfilled' ? statsResult.value : null,
          riskScoreValue: riskScoreResult.status === 'fulfilled' ? riskScoreResult.value : null,
        });
      }

      // Process portfolio statistics
      let validStats: PortfolioStatistics | null = null;
      if (statsResult.status === 'fulfilled') {
        const stats = statsResult.value;
        if (hasValidPortfolioStats(stats)) {
          validStats = stats;
          setPortfolioStats(stats);
          setHasPortfolioStats(true);
          
          if (process.env.NODE_ENV === 'development') {
            console.log('[PortfolioComparison] Portfolio stats loaded:', {
              averageRiskScore: stats.averageRiskScore,
              totalMerchants: stats.totalMerchants,
              type: typeof stats.averageRiskScore,
            });
          }
        } else {
          if (process.env.NODE_ENV === 'development') {
            console.warn('[PortfolioComparison] Invalid portfolio stats structure:', stats);
      }
        }
      } else {
        if (process.env.NODE_ENV === 'development') {
          console.error('[PortfolioComparison] Failed to fetch portfolio stats:', statsResult.reason);
        }
      }

      // Process merchant risk score
      let validRiskScore: MerchantRiskScore | null = null;
      if (riskScoreResult.status === 'fulfilled') {
        const riskScore = riskScoreResult.value;
        if (hasValidMerchantRiskScore(riskScore)) {
          validRiskScore = riskScore;
          setMerchantRiskScore(riskScore);
          setHasRiskScore(true);
          
          if (process.env.NODE_ENV === 'development') {
            console.log('[PortfolioComparison] Risk score loaded:', {
              risk_score: riskScore.risk_score,
              risk_level: riskScore.risk_level,
              type: typeof riskScore.risk_score,
            });
          }
        } else {
          if (process.env.NODE_ENV === 'development') {
            console.warn('[PortfolioComparison] Invalid risk score structure:', riskScore);
          }
        }
      } else {
        if (process.env.NODE_ENV === 'development') {
          console.error('[PortfolioComparison] Failed to fetch risk score:', riskScoreResult.reason);
        }
      }

      // Log field availability
      if (process.env.NODE_ENV === 'development') {
        console.log('[PortfolioComparison] Fields available:', {
          hasRiskScore: !!validRiskScore,
          hasPortfolioAvg: !!validStats,
          merchantScoreType: validRiskScore ? typeof validRiskScore.risk_score : 'undefined',
        });
      }

      // Calculate comparison if we have both data
      if (validStats && validRiskScore) {
        const merchantScore = validRiskScore.risk_score;
        const portfolioAvg = validStats.averageRiskScore;
        
        if (!isValidRiskScore(merchantScore) || !isValidRiskScore(portfolioAvg)) {
          if (process.env.NODE_ENV === 'development') {
            console.warn('[PortfolioComparison] Invalid score values:', {
              merchantScore,
              portfolioAvg,
              merchantScoreValid: isValidRiskScore(merchantScore),
              portfolioAvgValid: isValidRiskScore(portfolioAvg),
            });
          }
          setLoading(false);
          return;
        }
        
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
      } else if (validStats && !validRiskScore) {
        // If we only have portfolio stats, show portfolio average with note
        // Don't create estimated comparison - show actionable message instead
        setPortfolioStats(validStats);
        setHasPortfolioStats(true);
      } else if (validRiskScore && !validStats) {
        // If we only have risk score, show merchant score only
        setMerchantRiskScore(validRiskScore);
        setHasRiskScore(true);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load portfolio comparison';
      setError(errorMessage);
      toast.error('Failed to load comparison data', {
        description: errorMessage,
      });
    } finally {
      setLoading(false);
      setIsRefreshing(false);
      setLastRefreshTime(new Date());
    }
  }, [merchantId, merchantRiskLevel]);
  
  const handleRefresh = useCallback(() => {
    // Optimistic update: show loading state immediately
    setIsRefreshing(true);
    // Fetch with cache bypass
    fetchComparisonData(true);
  }, [fetchComparisonData]);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    fetchComparisonData();
  }, [fetchComparisonData]);
  
  // Format last refresh time
  useEffect(() => {
    if (!mounted || !lastRefreshTime) {
      setFormattedLastRefresh('');
      return;
    }
    const updateTime = () => {
      setFormattedLastRefresh(formatRelativeTime(lastRefreshTime));
    };
    updateTime();
    const interval = setInterval(updateTime, 60000); // Update every minute
    return () => clearInterval(interval);
  }, [mounted, lastRefreshTime]);

  // Format numbers on client side only to prevent hydration errors
  useEffect(() => {
    if (!mounted || !portfolioStats) {
      setFormattedTotalMerchants('');
      return;
    }
    setFormattedTotalMerchants(portfolioStats.totalMerchants.toLocaleString());
  }, [mounted, portfolioStats]);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Comparison</CardTitle>
          <CardDescription>Loading portfolio comparison...</CardDescription>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-32 w-full" />
        </CardContent>
      </Card>
    );
  }

  // Handle missing risk score with actionable CTA
  if (!hasRiskScore && !hasPortfolioStats) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Comparison</CardTitle>
          <CardDescription>Comparing merchant performance to portfolio</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Risk Score Required</AlertTitle>
            <AlertDescription className="mt-2">
              {formatErrorWithCode(
                'A risk assessment must be completed before portfolio comparison can be displayed.',
                ErrorCodes.PORTFOLIO_COMPARISON.MISSING_BOTH
              )}
              <Button 
                onClick={handleStartAssessment} 
                className="mt-3 w-full sm:w-auto"
                variant="default"
              >
                <Shield className="h-4 w-4 mr-2" />
                Run Risk Assessment
              </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  // Handle missing risk score but portfolio stats available
  if (!hasRiskScore && hasPortfolioStats) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Comparison</CardTitle>
          <CardDescription>Comparing merchant performance to portfolio</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Risk Score Required</AlertTitle>
            <AlertDescription className="mt-2">
              {formatErrorWithCode(
                'Complete a risk assessment to see how this merchant compares to your portfolio.',
                ErrorCodes.PORTFOLIO_COMPARISON.MISSING_RISK_SCORE
              )}
              <div className="mt-3 space-y-2">
                <Button 
                  onClick={handleStartAssessment} 
                  className="w-full sm:w-auto"
                  variant="default"
                >
                  <Shield className="h-4 w-4 mr-2" />
                  Run Risk Assessment
                </Button>
                {portfolioStats && (
                  <div className="text-xs text-muted-foreground pt-2" suppressHydrationWarning>
                    Portfolio average: {formatNumber(portfolioStats.averageRiskScore, 2)} 
                    ({mounted ? formattedTotalMerchants || 'Loading...' : 'Loading...'} merchants)
                  </div>
                )}
              </div>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  // Handle missing portfolio stats but risk score available
  if (hasRiskScore && !hasPortfolioStats) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Portfolio Comparison</CardTitle>
          <CardDescription>Comparing merchant performance to portfolio</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Portfolio Statistics Unavailable</AlertTitle>
            <AlertDescription className="mt-2">
              {formatErrorWithCode(
                'Portfolio statistics are being calculated. Please try again in a few moments.',
                ErrorCodes.PORTFOLIO_COMPARISON.MISSING_PORTFOLIO_STATS
              )}
              <div className="mt-3 space-y-2">
                <Button 
                  onClick={fetchComparisonData} 
                  className="w-full sm:w-auto"
                  variant="outline"
                >
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Refresh Data
                </Button>
                <div className="text-xs text-muted-foreground pt-2">
                  Merchant risk score: {formatNumber(merchantRiskScore?.risk_score ?? 0, 2)}
                </div>
              </div>
            </AlertDescription>
          </Alert>
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
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error Loading Comparison</AlertTitle>
            <AlertDescription className="mt-2">
              {formatErrorWithCode(error, ErrorCodes.PORTFOLIO_COMPARISON.FETCH_ERROR)}
              <Button 
                onClick={fetchComparisonData} 
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
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle>Portfolio Comparison</CardTitle>
            <CardDescription>
              Comparing merchant performance to portfolio average
              {formattedLastRefresh && (
                <span className="ml-2 text-xs text-muted-foreground" suppressHydrationWarning>
                  • Updated {formattedLastRefresh}
                </span>
              )}
            </CardDescription>
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={handleRefresh}
            disabled={isRefreshing}
            aria-label="Refresh portfolio comparison data"
            className="ml-2"
          >
            <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Merchant Risk Score</p>
            <p className="text-2xl font-bold">{formatNumber(comparison.merchantScore, 2)}</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Portfolio Average</p>
            <p className="text-2xl font-bold">{formatNumber(comparison.portfolioAverage, 2)}</p>
          </div>
        </div>

        <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
          <div>
            <p className="text-sm font-medium">Position in Portfolio</p>
            <p className="text-lg font-semibold">{getPercentileLabel()}</p>
            <p className="text-xs text-muted-foreground">
              {formatPercentile(comparison.percentile, 1)} percentile
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
              {formatNumber(comparison.difference, 2)} ({formatPercentWithSign(comparison.differencePercentage)})
            </span>
          </div>
          {portfolioStats.totalMerchants > 0 && (
            <div className="flex justify-between items-center">
              <span className="text-sm text-muted-foreground">Portfolio Size</span>
              <span className="text-sm font-medium" suppressHydrationWarning>
                {mounted ? formattedTotalMerchants || 'Loading...' : 'Loading...'} merchants
              </span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

