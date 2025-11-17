'use client';

import { ExportButton } from '@/components/export/ExportButton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { getMerchantAnalytics, getWebsiteAnalysis } from '@/lib/api';
import { deferNonCriticalDataLoad } from '@/lib/lazy-loader';
import type { AnalyticsData, WebsiteAnalysisData } from '@/types/merchant';
import { useEffect, useState } from 'react';

interface BusinessAnalyticsTabProps {
  merchantId: string;
}

export function BusinessAnalyticsTab({ merchantId }: BusinessAnalyticsTabProps) {
  const [analytics, setAnalytics] = useState<AnalyticsData | null>(null);
  const [websiteAnalysis, setWebsiteAnalysis] = useState<WebsiteAnalysisData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadAnalytics() {
      try {
        setLoading(true);
        setError(null);
        // Load critical analytics data immediately
        const analyticsData = await getMerchantAnalytics(merchantId).catch(() => null);
        setAnalytics(analyticsData);
        setLoading(false);
        
        // Defer non-critical website analysis
        deferNonCriticalDataLoad(async () => {
          const websiteData = await getWebsiteAnalysis(merchantId).catch(() => null);
          setWebsiteAnalysis(websiteData);
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load analytics');
        setLoading(false);
      }
    }

    if (merchantId) {
      loadAnalytics();
    }
  }, [merchantId]);

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

  const getExportData = async () => {
    return {
      analytics,
      websiteAnalysis,
      merchantId,
      exportedAt: new Date().toISOString(),
    };
  };

  return (
    <div className="space-y-6">
      {analytics && (
        <>
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Classification</CardTitle>
                  <CardDescription>Industry classification data</CardDescription>
                </div>
                <ExportButton
                  data={getExportData}
                  exportType="analytics"
                  merchantId={merchantId}
                  formats={['csv', 'json', 'excel', 'pdf']}
                />
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Primary Industry</p>
                <p className="text-lg">{analytics.classification.primaryIndustry || 'N/A'}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Confidence Score</p>
                <p>{(analytics.classification.confidenceScore * 100).toFixed(1)}%</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Risk Level</p>
                <Badge variant="outline">{analytics.classification.riskLevel}</Badge>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Security</CardTitle>
              <CardDescription>Security metrics</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Trust Score</p>
                <p>{(analytics.security.trustScore * 100).toFixed(1)}%</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">SSL Valid</p>
                <Badge variant={analytics.security.sslValid ? 'default' : 'destructive'}>
                  {analytics.security.sslValid ? 'Valid' : 'Invalid'}
                </Badge>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Data Quality</CardTitle>
              <CardDescription>Data completeness metrics</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Completeness Score</p>
                <p>{(analytics.quality.completenessScore * 100).toFixed(1)}%</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Data Points</p>
                <p>{analytics.quality.dataPoints}</p>
              </div>
            </CardContent>
          </Card>
        </>
      )}

      {websiteAnalysis && (
        <Card>
          <CardHeader>
            <CardTitle>Website Analysis</CardTitle>
            <CardDescription>Website performance and security</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Website URL</p>
              <a
                href={websiteAnalysis.websiteUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline"
              >
                {websiteAnalysis.websiteUrl}
              </a>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Performance Score</p>
              <p>{websiteAnalysis.performance.score}/100</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Accessibility Score</p>
              <p>{(websiteAnalysis.accessibility.score * 100).toFixed(1)}%</p>
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
    </div>
  );
}

