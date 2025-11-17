'use client';

import dynamic from 'next/dynamic';
import { AppLayout } from '@/components/layout/AppLayout';
import { Skeleton } from '@/components/ui/skeleton';

// Lazy load BulkOperationsManager (it's a heavy component)
const BulkOperationsManager = dynamic(
  () => import('@/components/bulk-operations/BulkOperationsManager').then((mod) => ({ default: mod.BulkOperationsManager })),
  {
    loading: () => <Skeleton className="h-96 w-full" />,
    ssr: false,
  }
);

export default function MerchantBulkOperationsPage() {
  return (
    <AppLayout
      title="Bulk Operations"
      description="Perform bulk operations on merchants"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Merchant Portfolio', href: '/merchant-portfolio' },
        { label: 'Bulk Operations' },
      ]}
    >
      <BulkOperationsManager />
    </AppLayout>
  );
}

