import { AppLayout } from '@/components/layout/AppLayout';
import { MerchantForm } from '@/components/forms/MerchantForm';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { PlusCircle } from 'lucide-react';

export default function AddMerchantPage() {
  return (
    <AppLayout
      title="Add New Merchant"
      description="Enter merchant information to perform comprehensive business verification, risk assessment, and analytics analysis."
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Merchant Portfolio', href: '/merchant-portfolio' },
        { label: 'Add Merchant' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <PlusCircle className="h-6 w-6 text-primary" />
                <div>
                  <CardTitle>Add New Merchant</CardTitle>
                  <CardDescription className="mt-1">
                    Enter merchant information to perform comprehensive business verification
                  </CardDescription>
                </div>
              </div>
              <div className="flex gap-2">
                <Badge variant="default">Live</Badge>
                <Badge variant="secondary">Beta</Badge>
              </div>
            </div>
          </CardHeader>
        </Card>

        <MerchantForm />
      </div>
    </AppLayout>
  );
}

