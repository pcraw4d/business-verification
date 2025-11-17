import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { DataTable, type Column } from '@/components/dashboards/DataTable';
import { Shield } from 'lucide-react';

interface RiskAssessment {
  id: string;
  merchantName: string;
  riskScore: number;
  riskLevel: string;
  lastUpdated: string;
}

export default function RiskAssessmentPortfolioPage() {
  const columns: Column<RiskAssessment>[] = [
    { key: 'merchantName', header: 'Merchant', sortable: true },
    { key: 'riskScore', header: 'Risk Score', sortable: true },
    { key: 'riskLevel', header: 'Risk Level', sortable: true },
    { key: 'lastUpdated', header: 'Last Updated', sortable: true },
  ];

  return (
    <AppLayout
      title="Risk Assessment Portfolio"
      description="Portfolio-wide risk assessment overview"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Risk Dashboard', href: '/risk-dashboard' },
        { label: 'Portfolio' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Shield className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Risk Assessment Portfolio</CardTitle>
                <CardDescription>Portfolio-wide risk assessment overview</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <DataTable
              data={[]}
              columns={columns}
              searchable
              pagination={{ pageSize: 10 }}
              emptyMessage="No risk assessments available"
            />
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

