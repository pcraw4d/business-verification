import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Plug } from 'lucide-react';

export default function MerchantHubIntegrationPage() {
  return (
    <AppLayout
      title="Merchant Hub Integration"
      description="Integration settings and configurations"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Merchant Hub', href: '/merchant-hub' },
        { label: 'Integration' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Plug className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Merchant Hub Integration</CardTitle>
                <CardDescription>Integration settings and configurations</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Integration settings will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

