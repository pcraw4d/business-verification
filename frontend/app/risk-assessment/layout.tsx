import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Risk Assessment | KYB Platform',
  description: 'Risk assessment and analysis',
};

export default function RiskAssessmentLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}

