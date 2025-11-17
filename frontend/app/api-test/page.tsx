import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { TestTube } from 'lucide-react';

export default function ApiTestPage() {
  return (
    <AppLayout
      title="API Testing"
      description="API endpoint testing and validation"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'API Testing' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <TestTube className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>API Testing</CardTitle>
                <CardDescription>API endpoint testing and validation</CardDescription>
              </div>
              <Badge variant="outline">Development Only</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              API testing interface will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

