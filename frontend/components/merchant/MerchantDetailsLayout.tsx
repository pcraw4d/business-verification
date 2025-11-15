'use client';

import { useState, useEffect } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { MerchantOverviewTab } from './MerchantOverviewTab';
import { BusinessAnalyticsTab } from './BusinessAnalyticsTab';
import { RiskAssessmentTab } from './RiskAssessmentTab';
import { RiskIndicatorsTab } from './RiskIndicatorsTab';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { getMerchant } from '@/lib/api';
import type { Merchant } from '@/types/merchant';

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
        const data = await getMerchant(merchantId);
        setMerchant(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load merchant data');
      } finally {
        setLoading(false);
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
        <TabsList className="grid w-full grid-cols-4">
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

