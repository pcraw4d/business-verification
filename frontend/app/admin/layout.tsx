import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Admin | KYB Platform',
  description: 'System administration',
};

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}

