'use client';

import { ExportButton } from '@/components/export/ExportButton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { BarChart, PieChart } from '@/components/charts/lazy';
import { getMerchant, getMerchantAnalytics, getWebsiteAnalysis } from '@/lib/api';
import { deferNonCriticalDataLoad } from '@/lib/lazy-loader';
import { formatPercent } from '@/lib/number-format';
import type { AnalyticsData, WebsiteAnalysisData, IndustryCode, Merchant } from '@/types/merchant';
import { useEffect, useState, useMemo } from 'react';
import { RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import { AnalyticsComparison } from './AnalyticsComparison';

interface BusinessAnalyticsTabProps {
  merchantId: string;
}

export function BusinessAnalyticsTab({ merchantId }: BusinessAnalyticsTabProps) {
  const [analytics, setAnalytics] = useState<AnalyticsData | null>(null);
  const [websiteAnalysis, setWebsiteAnalysis] = useState<WebsiteAnalysisData | null>(null);
  const [merchant, setMerchant] = useState<Merchant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastRefreshTime, setLastRefreshTime] = useState<Date | null>(null);
  const [formattedEmployeeCount, setFormattedEmployeeCount] = useState<string>('');
  const [formattedMerchantEmployeeCount, setFormattedMerchantEmployeeCount] = useState<string>('');
  const [formattedAnnualRevenue, setFormattedAnnualRevenue] = useState<string>('');
  const [formattedMerchantRevenue, setFormattedMerchantRevenue] = useState<string>('');

  useEffect(() => {
    setMounted(true);
  }, []);

  const loadAnalytics = async (bypassCache = false) => {
      try {
      if (!bypassCache) {
        setLoading(true);
      } else {
        setIsRefreshing(true);
      }
        setError(null);
        // Load critical analytics data immediately
        const analyticsData = await getMerchantAnalytics(merchantId).catch(() => null);
        setAnalytics(analyticsData);
        
        // Load merchant data for comparison with intelligence data
        const merchantData = await getMerchant(merchantId).catch(() => null);
        setMerchant(merchantData);
        
        setLoading(false);
      setIsRefreshing(false);
      setLastRefreshTime(new Date());
        
        // Defer non-critical website analysis
        deferNonCriticalDataLoad(async () => {
          const websiteData = await getWebsiteAnalysis(merchantId).catch(() => null);
          setWebsiteAnalysis(websiteData);
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load analytics');
        setLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    if (merchantId) {
      loadAnalytics();
    }
  }, [merchantId]);

  const handleRefresh = () => {
    loadAnalytics(true);
  };

  // Format relative time for last refresh
  const formatRelativeTime = (date: Date): string => {
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
  };

  // Keyboard shortcut: R to refresh
  useKeyboardShortcuts([
    {
      key: 'r',
      handler: handleRefresh,
      description: 'Refresh business analytics data',
    },
  ]);

  // Format numbers on client side only to prevent hydration errors
  useEffect(() => {
    if (!mounted) return;

    if (analytics?.intelligence?.employeeCount !== undefined && analytics.intelligence.employeeCount !== null) {
      setFormattedEmployeeCount(analytics.intelligence.employeeCount.toLocaleString());
    } else {
      setFormattedEmployeeCount('');
    }

    if (merchant?.employeeCount !== undefined && merchant.employeeCount !== null) {
      setFormattedMerchantEmployeeCount(merchant.employeeCount.toLocaleString());
    } else {
      setFormattedMerchantEmployeeCount('');
    }

    if (analytics?.intelligence?.annualRevenue !== undefined && analytics.intelligence.annualRevenue !== null) {
      setFormattedAnnualRevenue(
        new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency: 'USD',
          minimumFractionDigits: 0,
          maximumFractionDigits: 0,
        }).format(analytics.intelligence.annualRevenue)
      );
    } else {
      setFormattedAnnualRevenue('');
    }

    if (merchant?.annualRevenue !== undefined && merchant.annualRevenue !== null) {
      setFormattedMerchantRevenue(
        new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency: 'USD',
          minimumFractionDigits: 0,
          maximumFractionDigits: 0,
        }).format(merchant.annualRevenue)
      );
    } else {
      setFormattedMerchantRevenue('');
    }
  }, [mounted, analytics?.intelligence?.employeeCount, analytics?.intelligence?.annualRevenue, merchant?.employeeCount, merchant?.annualRevenue]);

  // Get top 3 industry codes per type
  const getTopCodes = (codes: IndustryCode[] | undefined, limit = 3): IndustryCode[] => {
    if (!codes || codes.length === 0) return [];
    return [...codes]
      .sort((a, b) => b.confidence - a.confidence)
      .slice(0, limit);
  };

  // ALL HOOKS MUST BE CALLED BEFORE ANY EARLY RETURNS
  // This fixes React Error #310: "Rendered more hooks than during the previous render"
  const topMccCodes = useMemo(() => getTopCodes(analytics?.classification?.mccCodes), [analytics]);
  const topSicCodes = useMemo(() => getTopCodes(analytics?.classification?.sicCodes), [analytics]);
  const topNaicsCodes = useMemo(() => getTopCodes(analytics?.classification?.naicsCodes), [analytics]);

  // Prepare chart data
  const classificationChartData = useMemo(() => {
    if (!analytics?.classification) return [];
    return [
      { name: 'Confidence', value: (analytics.classification.confidenceScore ?? 0) * 100 },
      { name: 'Remaining', value: 100 - ((analytics.classification.confidenceScore ?? 0) * 100) },
    ];
  }, [analytics]);

  const dataQualityChartData = useMemo(() => {
    if (!analytics?.quality) return [];
    return [
      { name: 'Complete', value: (analytics.quality.completenessScore ?? 0) * 100 },
      { name: 'Missing', value: 100 - ((analytics.quality.completenessScore ?? 0) * 100) },
    ];
  }, [analytics]);

  const securityChartData = useMemo(() => {
    if (!analytics?.security) return [];
    return [
      { name: 'Trust Score', value: (analytics.security.trustScore ?? 0) * 100 },
      { name: 'Remaining', value: 100 - ((analytics.security.trustScore ?? 0) * 100) },
    ];
  }, [analytics]);

  const industryCodeDistributionData = useMemo(() => {
    const data = [];
    if (topMccCodes.length > 0) {
      data.push({ name: 'MCC', value: topMccCodes.length });
    }
    if (topSicCodes.length > 0) {
      data.push({ name: 'SIC', value: topSicCodes.length });
    }
    if (topNaicsCodes.length > 0) {
      data.push({ name: 'NAICS', value: topNaicsCodes.length });
    }
    return data;
  }, [topMccCodes, topSicCodes, topNaicsCodes]);

  const getExportData = async () => {
    return {
      analytics,
      websiteAnalysis,
      merchantId,
      exportedAt: new Date().toISOString(),
    };
  };

  // Early returns AFTER all hooks
  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <section className="space-y-6" aria-labelledby="analytics-heading">
      {/* Header with refresh button */}
      <header className="flex items-center justify-between">
        <div>
          <h2 id="analytics-heading" className="text-2xl font-bold">Business Analytics</h2>
          {lastRefreshTime && (
            <p className="text-sm text-muted-foreground mt-1" aria-live="polite">
              Updated {formatRelativeTime(lastRefreshTime)}
            </p>
          )}
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleRefresh}
          disabled={isRefreshing || loading}
          aria-label="Refresh business analytics data"
          title="Refresh data (R)"
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${isRefreshing ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </header>

      {analytics && (
        <>
          {/* Portfolio Analytics Comparison */}
          <AnalyticsComparison merchantId={merchantId} merchantAnalytics={analytics} />

          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Classification</CardTitle>
                  <CardDescription>Industry classification data</CardDescription>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant="outline">From Analytics API</Badge>
                <ExportButton
                  data={getExportData}
                  exportType="analytics"
                  merchantId={merchantId}
                  formats={['csv', 'json', 'excel', 'pdf']}
                />
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Primary Industry</p>
                <p className="text-lg">{analytics.classification?.primaryIndustry || 'N/A'}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Confidence Score</p>
                <p>{formatPercent(analytics.classification?.confidenceScore)}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Risk Level</p>
                <Badge variant="outline">{analytics.classification?.riskLevel || 'N/A'}</Badge>
              </div>
            </CardContent>
          </Card>

          {/* Industry Codes Tables */}
          {(topMccCodes.length > 0 || topSicCodes.length > 0 || topNaicsCodes.length > 0) && (
            <div className="grid gap-6 md:grid-cols-3">
              {topMccCodes.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle>MCC Codes</CardTitle>
                    <CardDescription>Top {topMccCodes.length} Merchant Category Codes</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Code</TableHead>
                          <TableHead>Description</TableHead>
                          <TableHead className="text-right">Confidence</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {topMccCodes.map((code, index) => (
                          <TableRow key={index}>
                            <TableCell className="font-mono text-sm">{code.code}</TableCell>
                            <TableCell className="max-w-[200px] truncate" title={code.description}>
                              {code.description}
                            </TableCell>
                            <TableCell className="text-right">
                              {formatPercent(code.confidence)}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </CardContent>
                </Card>
              )}

              {topSicCodes.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle>SIC Codes</CardTitle>
                    <CardDescription>Top {topSicCodes.length} Standard Industrial Classification Codes</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Code</TableHead>
                          <TableHead>Description</TableHead>
                          <TableHead className="text-right">Confidence</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {topSicCodes.map((code, index) => (
                          <TableRow key={index}>
                            <TableCell className="font-mono text-sm">{code.code}</TableCell>
                            <TableCell className="max-w-[200px] truncate" title={code.description}>
                              {code.description}
                            </TableCell>
                            <TableCell className="text-right">
                              {formatPercent(code.confidence)}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </CardContent>
                </Card>
              )}

              {topNaicsCodes.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle>NAICS Codes</CardTitle>
                    <CardDescription>Top {topNaicsCodes.length} North American Industry Classification Codes</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Code</TableHead>
                          <TableHead>Description</TableHead>
                          <TableHead className="text-right">Confidence</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {topNaicsCodes.map((code, index) => (
                          <TableRow key={index}>
                            <TableCell className="font-mono text-sm">{code.code}</TableCell>
                            <TableCell className="max-w-[200px] truncate" title={code.description}>
                              {code.description}
                            </TableCell>
                            <TableCell className="text-right">
                              {formatPercent(code.confidence)}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </CardContent>
                </Card>
              )}
            </div>
          )}

          {/* Charts Section */}
          <div className="grid gap-6 md:grid-cols-2">
            {classificationChartData.length > 0 && (
              <ChartContainer
                title="Classification Confidence"
                description="Confidence score distribution"
                isLoading={false}
              >
                <PieChart
                  data={classificationChartData}
                  height={250}
                  isLoading={false}
                />
              </ChartContainer>
            )}

            {industryCodeDistributionData.length > 0 && (
              <ChartContainer
                title="Industry Code Distribution"
                description="Number of codes per classification type"
                isLoading={false}
              >
                <BarChart
                  data={industryCodeDistributionData}
                  dataKey="value"
                  bars={[{ key: 'value', name: 'Codes', color: '#8884d8' }]}
                  height={250}
                  isLoading={false}
                />
              </ChartContainer>
            )}
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                <CardTitle>Security</CardTitle>
                <CardDescription>Security metrics</CardDescription>
                  </div>
                  <Badge variant="outline">From Analytics API</Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Trust Score</p>
                  <p>{formatPercent(analytics.security?.trustScore)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">SSL Valid</p>
                  <Badge variant={analytics.security?.sslValid ? 'default' : 'destructive'}>
                    {analytics.security?.sslValid ? 'Valid' : 'Invalid'}
                  </Badge>
                </div>
                {securityChartData.length > 0 && (
                  <div className="mt-4">
                    <ChartContainer
                      title="Security Trust Score"
                      description="Trust score visualization"
                      isLoading={false}
                    >
                      <PieChart
                        data={securityChartData}
                        height={200}
                        isLoading={false}
                      />
                    </ChartContainer>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                <CardTitle>Data Quality</CardTitle>
                <CardDescription>Data completeness metrics</CardDescription>
                  </div>
                  <Badge variant="outline">From Analytics API</Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Completeness Score</p>
                  <p>{formatPercent(analytics.quality?.completenessScore)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Data Points</p>
                  <p>{analytics.quality?.dataPoints ?? 'N/A'}</p>
                </div>
                {dataQualityChartData.length > 0 && (
                  <div className="mt-4">
                    <ChartContainer
                      title="Data Quality"
                      description="Completeness score visualization"
                      isLoading={false}
                    >
                      <PieChart
                        data={dataQualityChartData}
                        height={200}
                        isLoading={false}
                      />
                    </ChartContainer>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Business Intelligence Card */}
          {analytics.intelligence && (
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Business Intelligence</CardTitle>
                    <CardDescription>Intelligence data from analytics API</CardDescription>
                  </div>
                  <Badge variant="outline">From Analytics API</Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {analytics.intelligence.businessAge !== undefined && analytics.intelligence.businessAge !== null && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Business Age</p>
                    <p>
                      {analytics.intelligence.businessAge >= 12
                        ? `${Math.floor(analytics.intelligence.businessAge / 12)} years`
                        : `${analytics.intelligence.businessAge} months`}
                    </p>
                  </div>
                )}
                
                {analytics.intelligence.employeeCount !== undefined && analytics.intelligence.employeeCount !== null && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Employee Count</p>
                    <div className="space-y-1">
                      <p suppressHydrationWarning>{mounted ? formattedEmployeeCount || 'Loading...' : 'Loading...'}</p>
                      {merchant?.employeeCount !== undefined && merchant.employeeCount !== null && 
                       merchant.employeeCount !== analytics.intelligence.employeeCount && (
                        <p className="text-xs text-muted-foreground" suppressHydrationWarning>
                          Note: Merchant data shows {mounted ? formattedMerchantEmployeeCount || 'Loading...' : 'Loading...'} employees
                        </p>
                      )}
                    </div>
                  </div>
                )}
                
                {analytics.intelligence.annualRevenue !== undefined && analytics.intelligence.annualRevenue !== null && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Annual Revenue</p>
                    <div className="space-y-1">
                      <p suppressHydrationWarning>{mounted ? formattedAnnualRevenue || 'Loading...' : 'Loading...'}</p>
                      {merchant?.annualRevenue !== undefined && merchant.annualRevenue !== null && 
                       Math.abs(merchant.annualRevenue - analytics.intelligence.annualRevenue) > 1000 && (
                        <p className="text-xs text-muted-foreground" suppressHydrationWarning>
                          Note: Merchant data shows {mounted ? formattedMerchantRevenue || 'Loading...' : 'Loading...'}
                        </p>
                      )}
                    </div>
                  </div>
                )}
                
                {(!analytics.intelligence.businessAge && 
                  !analytics.intelligence.employeeCount && 
                  !analytics.intelligence.annualRevenue) && (
                  <p className="text-sm text-muted-foreground">No intelligence data available</p>
                )}
              </CardContent>
            </Card>
          )}
        </>
      )}

      {websiteAnalysis && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
            <CardTitle>Website Analysis</CardTitle>
            <CardDescription>Website performance and security</CardDescription>
              </div>
              <Badge variant="outline">From Website Analysis API</Badge>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Website URL</p>
              {websiteAnalysis.websiteUrl ? (
                <a
                  href={websiteAnalysis.websiteUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {websiteAnalysis.websiteUrl}
                </a>
              ) : (
                <p>N/A</p>
              )}
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Performance Score</p>
              <p>{websiteAnalysis.performance?.score != null 
                ? `${websiteAnalysis.performance.score}/100`
                : 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Accessibility Score</p>
              <p>{formatPercent(websiteAnalysis.accessibility?.score)}</p>
            </div>
          </CardContent>
        </Card>
      )}

      {!analytics && !websiteAnalysis && !loading && (
        <EmptyState
          type="noData"
          title="No Analytics Data"
          message="Analytics data is not available for this merchant at this time."
        />
      )}
    </section>
  );
}

