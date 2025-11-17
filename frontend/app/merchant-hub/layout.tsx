import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Merchant Hub | KYB Platform',
  description: 'Merchant management hub',
};

export default function MerchantHubLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}

