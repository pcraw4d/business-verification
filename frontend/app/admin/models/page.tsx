import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { DataTable, type Column } from '@/components/dashboards/DataTable';
import { Brain } from 'lucide-react';

interface MLModel {
  id: string;
  name: string;
  version: string;
  accuracy: number;
  status: string;
}

export default function AdminModelsPage() {
  const columns: Column<MLModel>[] = [
    { key: 'name', header: 'Model Name', sortable: true },
    { key: 'version', header: 'Version', sortable: true },
    { key: 'accuracy', header: 'Accuracy', sortable: true },
    { key: 'status', header: 'Status', sortable: true },
  ];

  return (
    <AppLayout
      title="ML Models"
      description="Machine learning model management"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Admin', href: '/admin' },
        { label: 'ML Models' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Brain className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>ML Model Management</CardTitle>
                <CardDescription>Machine learning model management and performance tracking</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <DataTable
              data={[]}
              columns={columns}
              searchable
              pagination={{ pageSize: 10 }}
              emptyMessage="No ML models available"
            />
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

