import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Bell, AlertCircle } from 'lucide-react';

export default function ComplianceAlertsPage() {
  return (
    <AppLayout
      title="Compliance Alerts"
      description="Compliance alerts and notifications"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Compliance', href: '/compliance' },
        { label: 'Alerts' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Bell className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Compliance Alert System</CardTitle>
                <CardDescription>Compliance alerts and notifications</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Compliance alerts will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

