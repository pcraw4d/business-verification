'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import type { Merchant } from '@/types/merchant';
import { useEffect, useState } from 'react';
import { PortfolioComparisonCard } from './PortfolioComparisonCard';
import { RiskScoreCard } from './RiskScoreCard';
import { PortfolioContextBadge } from './PortfolioContextBadge';
import { EnrichmentButton } from './EnrichmentButton';

interface MerchantOverviewTabProps {
  merchant: Merchant;
}

export function MerchantOverviewTab({ merchant }: MerchantOverviewTabProps) {
  // Use client-side state for dates to avoid hydration issues
  const [createdDate, setCreatedDate] = useState<string>('');
  const [updatedDate, setUpdatedDate] = useState<string>('');

  useEffect(() => {
    // Format dates on client side only
    if (merchant.createdAt) {
      setCreatedDate(new Date(merchant.createdAt).toLocaleDateString());
    }
    if (merchant.updatedAt) {
      setUpdatedDate(new Date(merchant.updatedAt).toLocaleDateString());
    }
  }, [merchant.createdAt, merchant.updatedAt]);

  const displayName = merchant.businessName || merchant.name || 'Unknown Merchant';
  const hasAdditionalInfo = merchant.legalName || merchant.registrationNumber || merchant.taxId || merchant.businessType;

  return (
    <div className="space-y-6">
      {/* Portfolio Context Badge and Enrichment Button */}
      <div className="flex justify-between items-center">
        <EnrichmentButton merchantId={merchant.id} variant="outline" size="sm" />
        <PortfolioContextBadge merchantId={merchant.id} variant="detailed" />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Business Information</CardTitle>
            <CardDescription>Basic merchant details</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Business Name</p>
              <p className="text-lg">{displayName}</p>
            </div>
            {merchant.legalName && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Legal Name</p>
                <p>{merchant.legalName}</p>
              </div>
            )}
            {merchant.industry && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Industry</p>
                <p>{merchant.industry}</p>
              </div>
            )}
            {merchant.industryCode && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Industry Code</p>
                <p className="font-mono text-sm">{merchant.industryCode}</p>
              </div>
            )}
            {merchant.businessType && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Business Type</p>
                <p>{merchant.businessType}</p>
              </div>
            )}
            {merchant.description && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Description</p>
                <p className="text-sm">{merchant.description}</p>
              </div>
            )}
            <div>
              <p className="text-sm font-medium text-muted-foreground">Status</p>
              <Badge variant={merchant.status === 'active' ? 'default' : 'secondary'}>
                {merchant.status}
              </Badge>
            </div>
            {merchant.portfolioType && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Portfolio Type</p>
                <Badge variant="outline">{merchant.portfolioType}</Badge>
              </div>
            )}
            {merchant.riskLevel && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Risk Level</p>
                <Badge
                  variant={
                    merchant.riskLevel === 'low'
                      ? 'default'
                      : merchant.riskLevel === 'medium'
                      ? 'secondary'
                      : 'destructive'
                  }
                >
                  {merchant.riskLevel}
                </Badge>
              </div>
            )}
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

        {/* Registration and Compliance Information */}
        {hasAdditionalInfo && (
          <Card>
            <CardHeader>
              <CardTitle>Registration & Compliance</CardTitle>
              <CardDescription>Business registration and compliance details</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableBody>
                  {merchant.registrationNumber && (
                    <TableRow>
                      <TableCell className="font-medium text-muted-foreground w-1/3">
                        Registration Number
                      </TableCell>
                      <TableCell className="font-mono text-sm">{merchant.registrationNumber}</TableCell>
                    </TableRow>
                  )}
                  {merchant.taxId && (
                    <TableRow>
                      <TableCell className="font-medium text-muted-foreground w-1/3">Tax ID</TableCell>
                      <TableCell className="font-mono text-sm">{merchant.taxId}</TableCell>
                    </TableRow>
                  )}
                  {merchant.complianceStatus && (
                    <TableRow>
                      <TableCell className="font-medium text-muted-foreground w-1/3">
                        Compliance Status
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{merchant.complianceStatus}</Badge>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        )}

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

      {/* Risk Score and Portfolio Comparison */}
      <div className="grid gap-6 md:grid-cols-2">
        <RiskScoreCard merchantId={merchant.id} />
        <PortfolioComparisonCard merchantId={merchant.id} merchantRiskLevel={merchant.riskLevel} />
      </div>

      {/* Metadata Card */}
      <Card>
        <CardHeader>
          <CardTitle>Metadata</CardTitle>
          <CardDescription>System information</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell className="font-medium text-muted-foreground w-1/3">Merchant ID</TableCell>
                <TableCell className="font-mono text-sm">{merchant.id}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell className="font-medium text-muted-foreground w-1/3">Created</TableCell>
                <TableCell>{createdDate || 'N/A'}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell className="font-medium text-muted-foreground w-1/3">Last Updated</TableCell>
                <TableCell>{updatedDate || 'N/A'}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}

