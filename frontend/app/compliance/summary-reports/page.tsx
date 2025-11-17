import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { FileText, Download } from 'lucide-react';

export default function ComplianceSummaryReportsPage() {
  return (
    <AppLayout
      title="Summary Reports"
      description="Compliance summary reports and documentation"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Compliance', href: '/compliance' },
        { label: 'Summary Reports' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <FileText className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Compliance Summary Reports</CardTitle>
                <CardDescription>Compliance summary reports and documentation</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Summary reports will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

