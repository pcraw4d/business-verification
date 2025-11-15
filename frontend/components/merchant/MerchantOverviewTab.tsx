'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { Merchant } from '@/types/merchant';

interface MerchantOverviewTabProps {
  merchant: Merchant;
}

export function MerchantOverviewTab({ merchant }: MerchantOverviewTabProps) {
  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Business Information</CardTitle>
            <CardDescription>Basic merchant details</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Business Name</p>
              <p className="text-lg">{merchant.businessName}</p>
            </div>
            {merchant.industry && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Industry</p>
                <p>{merchant.industry}</p>
              </div>
            )}
            <div>
              <p className="text-sm font-medium text-muted-foreground">Status</p>
              <Badge variant={merchant.status === 'active' ? 'default' : 'secondary'}>
                {merchant.status}
              </Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Contact Information</CardTitle>
            <CardDescription>Contact details</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {merchant.email && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Email</p>
                <p>{merchant.email}</p>
              </div>
            )}
            {merchant.phone && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Phone</p>
                <p>{merchant.phone}</p>
              </div>
            )}
            {merchant.website && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Website</p>
                <a
                  href={merchant.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {merchant.website}
                </a>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {merchant.address && (
        <Card>
          <CardHeader>
            <CardTitle>Address</CardTitle>
            <CardDescription>Business location</CardDescription>
          </CardHeader>
          <CardContent>
            <p>
              {[
                merchant.address.street,
                merchant.address.city,
                merchant.address.state,
                merchant.address.postalCode,
                merchant.address.country,
              ]
                .filter(Boolean)
                .join(', ')}
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

