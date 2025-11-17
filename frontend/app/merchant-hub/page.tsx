import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Network, PlusCircle } from 'lucide-react';
import Link from 'next/link';

export default function MerchantHubPage() {
  return (
    <AppLayout
      title="Merchant Hub"
      description="Central hub for merchant management and operations"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub', href: '/dashboard-hub' },
        { label: 'Merchant Hub' },
      ]}
      headerActions={
        <Button asChild aria-label="Add new merchant">
          <Link href="/add-merchant">
            <PlusCircle className="h-4 w-4 mr-2" />
            Add Merchant
          </Link>
        </Button>
      }
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Network className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Merchant Hub</CardTitle>
                <CardDescription>Central hub for merchant management and operations</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground">
              Merchant hub content will be displayed here
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

