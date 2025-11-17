import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Search } from 'lucide-react';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Gap Analysis | Compliance',
  description: 'Identify compliance gaps and requirements',
};

export default function ComplianceGapAnalysisPage() {
  return (
    <AppLayout
      title="Gap Analysis"
      description="Identify compliance gaps and requirements"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Compliance', href: '/compliance' },
        { label: 'Gap Analysis' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Search className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Compliance Gap Analysis</CardTitle>
                <CardDescription>Identify compliance gaps and requirements</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Gap analysis content will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

