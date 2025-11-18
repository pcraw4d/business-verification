'use client';

import { MerchantDetailsLayout } from '@/components/merchant/MerchantDetailsLayout';
import { use } from 'react';
import { useParams } from 'next/navigation';

// Configure route for client-side routing
// Use 'use client' to enable client-side rendering for static file serving
// This allows the page to work with client-side routing when served as static files
export const dynamicParams = true;

export default function MerchantDetailsPage() {
  // Use useParams for client-side routing instead of async params
  const params = useParams();
  const id = params?.id as string;
  
  if (!id) {
    return (
      <div className="container mx-auto p-6">
        <p>Merchant ID is required</p>
      </div>
    );
  }
  
  return <MerchantDetailsLayout merchantId={id} />;
}

