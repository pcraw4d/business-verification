import { MerchantDetailsLayout } from '@/components/merchant/MerchantDetailsLayout';

interface MerchantDetailsPageProps {
  params: Promise<{ id: string }>;
}

// Configure route for static file serving
// Allow dynamic params so any merchant ID can be used
// This ensures Next.js generates the route template for client-side routing
export const dynamicParams = true;
// Don't force dynamic - allow static generation of the route template
// The template will be used for client-side routing by the Go service

export default async function MerchantDetailsPage({ params }: MerchantDetailsPageProps) {
  const { id } = await params;
  return <MerchantDetailsLayout merchantId={id} />;
}

