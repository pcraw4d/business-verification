import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Merchant | KYB Platform',
  description: 'Merchant management',
};

export default function MerchantLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}

