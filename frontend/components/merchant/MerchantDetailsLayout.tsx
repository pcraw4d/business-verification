'use client';

import { Alert, AlertDescription } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { getMerchant } from '@/lib/api';
import type { Merchant } from '@/types/merchant';
import dynamic from 'next/dynamic';
import { useEffect, useState } from 'react';

// Lazy load tabs - only load when needed
const MerchantOverviewTab = dynamic(
  () => import('./MerchantOverviewTab').then((mod) => ({ default: mod.MerchantOverviewTab })),
  { loading: () => <Skeleton className="h-64 w-full" />, ssr: false }
);

const BusinessAnalyticsTab = dynamic(
  () => import('./BusinessAnalyticsTab').then((mod) => ({ default: mod.BusinessAnalyticsTab })),
  { loading: () => <Skeleton className="h-64 w-full" />, ssr: false }
);

const RiskAssessmentTab = dynamic(
  () => import('./RiskAssessmentTab').then((mod) => ({ default: mod.RiskAssessmentTab })),
  { loading: () => <Skeleton className="h-64 w-full" />, ssr: false }
);

const RiskIndicatorsTab = dynamic(
  () => import('./RiskIndicatorsTab').then((mod) => ({ default: mod.RiskIndicatorsTab })),
  { loading: () => <Skeleton className="h-64 w-full" />, ssr: false }
);

interface MerchantDetailsLayoutProps {
  merchantId: string;
}

export function MerchantDetailsLayout({ merchantId }: MerchantDetailsLayoutProps) {
  const [merchant, setMerchant] = useState<Merchant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('overview');

  useEffect(() => {
    async function loadMerchant() {
      try {
        setLoading(true);
        setError(null);
        if (process.env.NODE_ENV === 'test') {
          console.log('[MerchantDetailsLayout] Starting to load merchant:', merchantId);
        }
        const data = await getMerchant(merchantId);
        if (process.env.NODE_ENV === 'test') {
          console.log('[MerchantDetailsLayout] Merchant data received:', data);
        }
        setMerchant(data);
        if (process.env.NODE_ENV === 'test') {
          console.log('[MerchantDetailsLayout] State updated with merchant data');
        }
      } catch (err) {
        if (process.env.NODE_ENV === 'test') {
          console.error('[MerchantDetailsLayout] Error loading merchant:', err);
        }
        setError(err instanceof Error ? err.message : 'Failed to load merchant data');
      } finally {
        setLoading(false);
        if (process.env.NODE_ENV === 'test') {
          console.log('[MerchantDetailsLayout] Loading complete');
        }
      }
    }

    if (merchantId) {
      loadMerchant();
    }
  }, [merchantId]);

  if (loading) {
    return (
      <div className="container mx-auto p-6 space-y-6">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    );
  }

  if (!merchant) {
    return (
      <div className="container mx-auto p-6">
        <Alert>
          <AlertDescription>Merchant not found</AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="border-b pb-4">
        <h1 className="text-3xl font-bold">{merchant.businessName}</h1>
        <p className="text-muted-foreground mt-2">
          {merchant.industry && `${merchant.industry} • `}
          Status: {merchant.status}
        </p>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-4 [@supports(display:grid)]:grid [@supports(display:-webkit-grid)]:grid">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="analytics">Business Analytics</TabsTrigger>
          <TabsTrigger value="risk">Risk Assessment</TabsTrigger>
          <TabsTrigger value="indicators">Risk Indicators</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-6">
          <MerchantOverviewTab merchant={merchant} />
        </TabsContent>

        <TabsContent value="analytics" className="mt-6">
          <BusinessAnalyticsTab merchantId={merchantId} />
        </TabsContent>

        <TabsContent value="risk" className="mt-6">
          <RiskAssessmentTab merchantId={merchantId} />
        </TabsContent>

        <TabsContent value="indicators" className="mt-6">
          <RiskIndicatorsTab merchantId={merchantId} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

