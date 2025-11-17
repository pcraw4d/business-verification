import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { DataTable, type Column } from '@/components/dashboards/DataTable';
import { Scale } from 'lucide-react';

interface Merchant {
  id: string;
  name: string;
  riskScore: number;
  status: string;
}

export default function MerchantComparisonPage() {
  const columns: Column<Merchant>[] = [
    { key: 'name', header: 'Merchant Name', sortable: true },
    { key: 'riskScore', header: 'Risk Score', sortable: true },
    { key: 'status', header: 'Status', sortable: true },
  ];

  return (
    <AppLayout
      title="Merchant Comparison"
      description="Compare merchants side by side"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Merchant Portfolio', href: '/merchant-portfolio' },
        { label: 'Comparison' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Scale className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Merchant Comparison</CardTitle>
                <CardDescription>Compare merchants side by side</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <DataTable
              data={[]}
              columns={columns}
              searchable
              emptyMessage="No merchants selected for comparison"
            />
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

