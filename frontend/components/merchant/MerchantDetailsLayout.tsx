'use client';

import { Alert, AlertDescription } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { getMerchant } from '@/lib/api';
import type { Merchant } from '@/types/merchant';
import dynamic from 'next/dynamic';
import { useEffect, useState, useCallback } from 'react';
import { RefreshCw } from 'lucide-react';
import { PortfolioContextBadge } from './PortfolioContextBadge';
import { EnrichmentButton } from './EnrichmentButton';
import { ErrorBoundary } from '@/components/ErrorBoundary';
import { TabErrorFallback } from './TabErrorFallback';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import { useRef } from 'react';

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

// Retry configuration
const MAX_RETRIES = 3;
const INITIAL_RETRY_DELAY = 1000; // 1 second

// Exponential backoff delay calculation
function getRetryDelay(attempt: number): number {
  return INITIAL_RETRY_DELAY * Math.pow(2, attempt);
}

export function MerchantDetailsLayout({ merchantId }: MerchantDetailsLayoutProps) {
  const [merchant, setMerchant] = useState<Merchant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [retryCount, setRetryCount] = useState(0);
  const [isRetrying, setIsRetrying] = useState(false);
  const [mounted, setMounted] = useState(false);
  const enrichmentButtonRef = useRef<HTMLButtonElement>(null);

  // Load merchant with retry logic
  const loadMerchant = useCallback(async (attempt: number = 0): Promise<void> => {
    try {
      if (attempt === 0) {
        setLoading(true);
      } else {
        setIsRetrying(true);
      }
      setError(null);
      
      if (process.env.NODE_ENV === 'test') {
        console.log(`[MerchantDetailsLayout] Loading merchant (attempt ${attempt + 1}):`, merchantId);
      }
      
      const data = await getMerchant(merchantId);
      
      if (process.env.NODE_ENV === 'test') {
        console.log('[MerchantDetailsLayout] Merchant data received:', data);
      }
      
      // Optimistic update: set merchant data immediately
      setMerchant(data);
      setRetryCount(0);
      setIsRetrying(false);
      
      if (process.env.NODE_ENV === 'test') {
        console.log('[MerchantDetailsLayout] State updated with merchant data');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load merchant data';
      
      if (process.env.NODE_ENV === 'test') {
        console.error(`[MerchantDetailsLayout] Error loading merchant (attempt ${attempt + 1}):`, err);
      }
      
      // Retry logic with exponential backoff
      if (attempt < MAX_RETRIES - 1) {
        const delay = getRetryDelay(attempt);
        setRetryCount(attempt + 1);
        
        if (process.env.NODE_ENV === 'test') {
          console.log(`[MerchantDetailsLayout] Retrying in ${delay}ms...`);
        }
        
        // Retry after delay
        setTimeout(() => {
          loadMerchant(attempt + 1);
        }, delay);
      } else {
        // Max retries reached
        setError(errorMessage);
        setRetryCount(0);
        setIsRetrying(false);
      }
    } finally {
      if (attempt === 0) {
        setLoading(false);
      }
      if (process.env.NODE_ENV === 'test') {
        console.log('[MerchantDetailsLayout] Loading complete');
      }
    }
  }, [merchantId]);

  useEffect(() => {
    if (merchantId) {
      loadMerchant();
    }
    // Cleanup: cancel any pending retries if component unmounts or merchantId changes
    return () => {
      setRetryCount(0);
      setIsRetrying(false);
    };
  }, [merchantId, loadMerchant]);

  // Set mounted state after component mounts to prevent hydration errors
  useEffect(() => {
    setMounted(true);
  }, []);

  // Manual retry function
  const handleRetry = useCallback(() => {
    setError(null);
    loadMerchant();
  }, [loadMerchant]);

  // Keyboard shortcut: E to open enrichment dialog
  const handleEnrichmentShortcut = useCallback(() => {
    if (enrichmentButtonRef.current) {
      enrichmentButtonRef.current.click();
    }
  }, []);

  useKeyboardShortcuts([
    {
      key: 'e',
      handler: handleEnrichmentShortcut,
      description: 'Open enrichment dialog',
    },
  ]);

  if (loading) {
    return (
      <div className="container mx-auto p-6 space-y-6">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-64 w-full" />
        {isRetrying && retryCount > 0 && (
          <div className="text-sm text-muted-foreground text-center">
            Retrying... (Attempt {retryCount + 1} of {MAX_RETRIES})
          </div>
        )}
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <Alert variant="destructive">
          <AlertDescription>
            <div className="flex items-center justify-between">
              <span>{error}</span>
              <Button
                onClick={handleRetry}
                variant="outline"
                size="sm"
                disabled={isRetrying}
                className="ml-4"
              >
                <RefreshCw className={`mr-2 h-4 w-4 ${isRetrying ? 'animate-spin' : ''}`} />
                {isRetrying ? 'Retrying...' : 'Try Again'}
              </Button>
            </div>
          </AlertDescription>
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
    <div className="container mx-auto px-4 py-6 space-y-6">
      <div className="border-b pb-4">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold" suppressHydrationWarning>{merchant.businessName || 'Unknown Merchant'}</h1>
              <PortfolioContextBadge merchantId={merchantId} variant="default" />
            </div>
            <p className="text-muted-foreground mt-2" suppressHydrationWarning>
              {merchant.industry && `${merchant.industry} • `}
              Status: {merchant.status || 'Unknown'}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <EnrichmentButton 
              merchantId={merchantId} 
              variant="outline" 
              size="sm"
              ref={enrichmentButtonRef}
              aria-label="Enrich merchant data (Press E)"
              title="Enrich data (E)"
            />
          </div>
        </div>
      </div>

      <section id="merchant-content" aria-label="Merchant details">

      {!mounted ? (
        <div className="space-y-4">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-64 w-full" />
        </div>
      ) : (
      <Tabs 
        value={activeTab} 
        onValueChange={setActiveTab} 
        className="w-full"
        suppressHydrationWarning
        key={`tabs-${mounted}`}
      >
        <TabsList className="grid w-full grid-cols-4" suppressHydrationWarning>
          <TabsTrigger value="overview" aria-label="Overview tab" suppressHydrationWarning>Overview</TabsTrigger>
          <TabsTrigger value="analytics" aria-label="Business Analytics tab" suppressHydrationWarning>Business Analytics</TabsTrigger>
          <TabsTrigger value="risk" aria-label="Risk Assessment tab" suppressHydrationWarning>Risk Assessment</TabsTrigger>
          <TabsTrigger value="indicators" aria-label="Risk Indicators tab" suppressHydrationWarning>Risk Indicators</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-6" suppressHydrationWarning>
          <ErrorBoundary
            fallback={
              <TabErrorFallback
                tabName="Overview"
                onRetry={() => {
                  // Force re-render by changing key
                  setActiveTab('overview');
                }}
              />
            }
            onError={(error, errorInfo) => {
              console.error('Overview tab error:', error, errorInfo);
            }}
          >
          <MerchantOverviewTab merchant={merchant} />
          </ErrorBoundary>
        </TabsContent>

        <TabsContent value="analytics" className="mt-6" suppressHydrationWarning>
          <ErrorBoundary
            fallback={
              <TabErrorFallback
                tabName="Business Analytics"
                onRetry={() => {
                  setActiveTab('analytics');
                }}
              />
            }
            onError={(error, errorInfo) => {
              console.error('Business Analytics tab error:', error, errorInfo);
            }}
          >
          <BusinessAnalyticsTab merchantId={merchantId} />
          </ErrorBoundary>
        </TabsContent>

        <TabsContent value="risk" className="mt-6" suppressHydrationWarning>
          <ErrorBoundary
            fallback={
              <TabErrorFallback
                tabName="Risk Assessment"
                onRetry={() => {
                  setActiveTab('risk');
                }}
              />
            }
            onError={(error, errorInfo) => {
              console.error('Risk Assessment tab error:', error, errorInfo);
            }}
          >
          <RiskAssessmentTab merchantId={merchantId} />
          </ErrorBoundary>
        </TabsContent>

        <TabsContent value="indicators" className="mt-6" suppressHydrationWarning>
          <ErrorBoundary
            fallback={
              <TabErrorFallback
                tabName="Risk Indicators"
                onRetry={() => {
                  setActiveTab('indicators');
                }}
              />
            }
            onError={(error, errorInfo) => {
              console.error('Risk Indicators tab error:', error, errorInfo);
            }}
          >
          <RiskIndicatorsTab merchantId={merchantId} />
          </ErrorBoundary>
        </TabsContent>
      </Tabs>
      )}
      </section>
    </div>
  );
}

