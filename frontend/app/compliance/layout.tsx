import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Compliance | KYB Platform',
  description: 'Compliance management and tracking',
};

export default function ComplianceLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}

