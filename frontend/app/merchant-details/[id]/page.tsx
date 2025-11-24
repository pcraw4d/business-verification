'use client';

import { Suspense } from 'react';
import { MerchantDetailsLayout } from '@/components/merchant/MerchantDetailsLayout';
import { useParams } from 'next/navigation';
import { ErrorBoundary } from '@/components/ErrorBoundary';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { EnrichmentProvider } from '@/contexts/EnrichmentContext';
import { AppLayout } from '@/components/layout/AppLayout';

// Configure route for client-side routing
// Use 'use client' to enable client-side rendering for static file serving
// This allows the page to work with client-side routing when served as static files
export const dynamicParams = true;

// Loading fallback component for Suspense
function MerchantDetailsLoading() {
  return (
    <div className="container mx-auto p-6 space-y-6">
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-64 w-full" />
      <Skeleton className="h-64 w-full" />
    </div>
  );
}

// Error fallback component for ErrorBoundary
function MerchantDetailsError() {
  return (
    <div className="container mx-auto p-6">
      <Alert variant="destructive">
        <AlertDescription>
          Failed to load merchant details page. Please try refreshing the page.
        </AlertDescription>
      </Alert>
    </div>
  );
}

// Inner component that uses useParams (must be in client component)
function MerchantDetailsContent() {
  const params = useParams();
  const id = params?.id as string;
  
  if (!id) {
    return (
      <div className="container mx-auto p-6">
        <Alert variant="destructive">
          <AlertDescription>Merchant ID is required</AlertDescription>
        </Alert>
      </div>
    );
  }
  
  return <MerchantDetailsLayout merchantId={id} />;
}

export default function MerchantDetailsPage() {
  return (
    <EnrichmentProvider>
      <AppLayout
        title="Merchant Details"
        breadcrumbs={[
          { label: 'Home', href: '/' },
          { label: 'Merchant Portfolio', href: '/merchant-portfolio' },
          { label: 'Merchant Details' },
        ]}
      >
        <ErrorBoundary
          fallback={<MerchantDetailsError />}
          onError={(error, errorInfo) => {
            // Log error for monitoring/debugging
            console.error('MerchantDetailsPage error:', error, errorInfo);
          }}
        >
          <Suspense fallback={<MerchantDetailsLoading />}>
            <MerchantDetailsContent />
          </Suspense>
        </ErrorBoundary>
      </AppLayout>
    </EnrichmentProvider>
  );
}

