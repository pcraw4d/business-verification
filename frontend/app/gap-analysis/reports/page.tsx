import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { FileText, Download } from 'lucide-react';

export default function GapAnalysisReportsPage() {
  return (
    <AppLayout
      title="Gap Analysis Reports"
      description="Gap analysis reports and documentation"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Compliance', href: '/compliance' },
        { label: 'Gap Analysis Reports' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <FileText className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Gap Analysis Reports</CardTitle>
                <CardDescription>Gap analysis reports and documentation</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Gap analysis reports will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

