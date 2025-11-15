import { MerchantDetailsLayout } from '@/components/merchant/MerchantDetailsLayout';

interface MerchantDetailsPageProps {
  params: Promise<{ id: string }>;
}

export default async function MerchantDetailsPage({ params }: MerchantDetailsPageProps) {
  const { id } = await params;
  return <MerchantDetailsLayout merchantId={id} />;
}

