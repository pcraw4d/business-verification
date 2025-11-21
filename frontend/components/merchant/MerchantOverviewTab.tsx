'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Button } from '@/components/ui/button';
import { ChevronDown } from 'lucide-react';
import type { Merchant } from '@/types/merchant';
import { useEffect, useState, type ReactNode } from 'react';
import { PortfolioComparisonCard } from './PortfolioComparisonCard';
import { RiskScoreCard } from './RiskScoreCard';
import { PortfolioContextBadge } from './PortfolioContextBadge';
import { EnrichmentButton } from './EnrichmentButton';
import { useEnrichment } from '@/contexts/EnrichmentContext';
import { Sparkles } from 'lucide-react';

interface MerchantOverviewTabProps {
  merchant: Merchant;
}

// Helper function to calculate data completeness
function calculateCompleteness(merchant: Merchant): number {
  const fields = [
    merchant.foundedDate,
    merchant.employeeCount,
    merchant.annualRevenue,
    merchant.address?.street1 || merchant.address?.street,
    merchant.address?.city,
    merchant.address?.country,
    merchant.email,
    merchant.phone,
    merchant.website,
    merchant.legalName,
    merchant.registrationNumber,
    merchant.taxId,
    merchant.industry,
    merchant.industryCode,
  ];
  const filled = fields.filter(Boolean).length;
  return Math.round((filled / fields.length) * 100);
}

// Helper function to format relative time
function formatRelativeTime(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) return 'just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
  return date.toLocaleDateString();
}

export function MerchantOverviewTab({ merchant }: MerchantOverviewTabProps) {
  const { isFieldEnriched, getEnrichedFieldInfo } = useEnrichment();
  // Use client-side state for dates to avoid hydration issues
  const [createdDate, setCreatedDate] = useState<string>('');
  const [updatedDate, setUpdatedDate] = useState<string>('');
  const [foundedDate, setFoundedDate] = useState<string>('');
  const [formattedRevenue, setFormattedRevenue] = useState<string>('');
  const [mounted, setMounted] = useState(false);
  const [formattedEmployeeCount, setFormattedEmployeeCount] = useState<string>('');
  const [lastUpdatedRelative, setLastUpdatedRelative] = useState<string>('');
  const [lastUpdatedFull, setLastUpdatedFull] = useState<string>('');

  // Field name mapping from enrichment results to merchant fields
  const fieldNameMap: Record<string, string> = {
    'Founded Date': 'foundedDate',
    'Employee Count': 'employeeCount',
    'Annual Revenue': 'annualRevenue',
    'Business Name': 'name',
    'Address': 'address',
  };

  const getFieldKey = (displayName: string): string => {
    return fieldNameMap[displayName] || displayName.toLowerCase().replace(/\s+/g, '');
  };

  const FieldHighlight = ({ fieldName, children }: { fieldName: string; children: ReactNode }) => {
    const fieldKey = getFieldKey(fieldName);
    const isEnriched = isFieldEnriched(merchant.id, fieldKey);
    const fieldInfo = getEnrichedFieldInfo(merchant.id, fieldKey);

    if (!isEnriched) {
      return <>{children}</>;
    }

    return (
      <div className="relative">
        <div className="flex items-center gap-2">
          {children}
          <Badge
            variant={fieldInfo?.type === 'added' ? 'default' : 'secondary'}
            className="text-xs animate-pulse"
            title={`${fieldInfo?.type === 'added' ? 'Added' : 'Updated'} from ${fieldInfo?.source} at ${fieldInfo?.enrichedAt.toLocaleTimeString()}`}
          >
            <Sparkles className="h-3 w-3 mr-1" />
            {fieldInfo?.type === 'added' ? 'New' : 'Updated'}
          </Badge>
        </div>
      </div>
    );
  };

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    // Format dates on client side only to prevent hydration errors
    // eslint-disable-next-line react-hooks/exhaustive-deps
    if (merchant.createdAt) {
      setCreatedDate(new Date(merchant.createdAt).toLocaleDateString());
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    if (merchant.updatedAt) {
      const updatedDateObj = new Date(merchant.updatedAt);
      setUpdatedDate(updatedDateObj.toLocaleDateString());
      setLastUpdatedFull(updatedDateObj.toLocaleString());
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    if (merchant.foundedDate) {
      setFoundedDate(new Date(merchant.foundedDate).toLocaleDateString());
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    if (merchant.annualRevenue) {
      setFormattedRevenue(
        new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency: 'USD',
          minimumFractionDigits: 0,
          maximumFractionDigits: 0,
        }).format(merchant.annualRevenue)
      );
    }
    // Format employee count on client side only
    if (merchant.employeeCount !== undefined && merchant.employeeCount !== null) {
      setFormattedEmployeeCount(merchant.employeeCount.toLocaleString());
    } else {
      setFormattedEmployeeCount('');
    }
  }, [merchant.createdAt, merchant.updatedAt, merchant.foundedDate, merchant.annualRevenue, merchant.employeeCount]);

  // Format relative time when merchant.updatedAt changes
  useEffect(() => {
    if (!mounted || !merchant.updatedAt) {
      setLastUpdatedRelative('');
      return;
    }
    setLastUpdatedRelative(formatRelativeTime(new Date(merchant.updatedAt)));
  }, [mounted, merchant.updatedAt]);

  const displayName = merchant.businessName || merchant.name || 'Unknown Merchant';
  const hasAdditionalInfo = merchant.legalName || merchant.registrationNumber || merchant.taxId || merchant.businessType;
  const completeness = calculateCompleteness(merchant);
  const completenessColor = completeness >= 80 ? 'default' : completeness >= 50 ? 'secondary' : 'destructive';

  return (
    <div className="space-y-6">
      {/* Portfolio Context Badge and Enrichment Button */}
      <div className="flex justify-between items-center">
        <EnrichmentButton merchantId={merchant.id} variant="outline" size="sm" />
        <div className="flex items-center gap-2">
          <Badge variant={completenessColor}>
            Data Completeness: {completeness}%
          </Badge>
        <PortfolioContextBadge merchantId={merchant.id} variant="detailed" />
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <div className="flex items-start justify-between">
              <div>
            <CardTitle>Business Information</CardTitle>
            <CardDescription>Basic merchant details</CardDescription>
              </div>
              {merchant.updatedAt && mounted && (
                <p 
                  className="text-xs text-muted-foreground mt-1" 
                  title={lastUpdatedFull}
                  suppressHydrationWarning
                >
                  Updated {lastUpdatedRelative || 'Loading...'}
                </p>
              )}
            </div>
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
            <div className="flex items-start justify-between">
              <div>
            <CardTitle>Contact Information</CardTitle>
            <CardDescription>Contact details</CardDescription>
              </div>
              {merchant.updatedAt && mounted && (
                <p 
                  className="text-xs text-muted-foreground mt-1" 
                  title={lastUpdatedFull}
                  suppressHydrationWarning
                >
                  Updated {lastUpdatedRelative || 'Loading...'}
                </p>
              )}
            </div>
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

        {/* Financial Information Card */}
        {(merchant.foundedDate || merchant.employeeCount || merchant.annualRevenue) && (
          <Card>
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
              <CardTitle>Financial Information</CardTitle>
              <CardDescription>Business financial details</CardDescription>
                </div>
                {merchant.updatedAt && mounted && (
                  <p 
                    className="text-xs text-muted-foreground mt-1" 
                    title={lastUpdatedFull}
                    suppressHydrationWarning
                  >
                    Updated {lastUpdatedRelative || 'Loading...'}
                  </p>
                )}
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {merchant.foundedDate && (
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Founded Date</p>
                  <FieldHighlight fieldName="Founded Date">
                    <p suppressHydrationWarning>{foundedDate || 'N/A'}</p>
                  </FieldHighlight>
                </div>
              )}
              {merchant.employeeCount !== undefined && (
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Employee Count</p>
                  <FieldHighlight fieldName="Employee Count">
                    <p suppressHydrationWarning>{mounted ? formattedEmployeeCount || 'Loading...' : 'Loading...'}</p>
                  </FieldHighlight>
                </div>
              )}
              {merchant.annualRevenue !== undefined && (
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Annual Revenue</p>
                  <FieldHighlight fieldName="Annual Revenue">
                    <p suppressHydrationWarning>{formattedRevenue || 'N/A'}</p>
                  </FieldHighlight>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>

        {/* Registration and Compliance Information */}
        {hasAdditionalInfo && (
          <Card>
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
              <CardTitle>Registration & Compliance</CardTitle>
              <CardDescription>Business registration and compliance details</CardDescription>
                </div>
                {merchant.updatedAt && mounted && (
                  <p 
                    className="text-xs text-muted-foreground mt-1" 
                    title={lastUpdatedFull}
                    suppressHydrationWarning
                  >
                    Updated {lastUpdatedRelative || 'Loading...'}
                  </p>
                )}
              </div>
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
            <div className="flex items-start justify-between">
              <div>
            <CardTitle>Address</CardTitle>
            <CardDescription>Business location</CardDescription>
              </div>
              {merchant.updatedAt && mounted && (
                <p 
                  className="text-xs text-muted-foreground mt-1" 
                  title={lastUpdatedFull}
                  suppressHydrationWarning
                >
                  Updated {lastUpdatedRelative || 'Loading...'}
                </p>
              )}
            </div>
          </CardHeader>
          <CardContent className="space-y-2">
            {(merchant.address.street1 || merchant.address.street) && (
              <FieldHighlight fieldName="Address">
                <p>{merchant.address.street1 || merchant.address.street}</p>
              </FieldHighlight>
            )}
            {merchant.address.street2 && (
              <FieldHighlight fieldName="Address">
                <p>{merchant.address.street2}</p>
              </FieldHighlight>
            )}
            <FieldHighlight fieldName="Address">
              <p>
                {[
                  merchant.address.city,
                  merchant.address.state,
                  merchant.address.postalCode,
                ]
                  .filter(Boolean)
                  .join(', ')}
              </p>
            </FieldHighlight>
            {merchant.address.country && (
              <FieldHighlight fieldName="Address">
                <p>
                  {merchant.address.country}
                  {merchant.address.countryCode && ` (${merchant.address.countryCode})`}
                </p>
              </FieldHighlight>
            )}
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
          <div className="flex items-start justify-between">
            <div>
          <CardTitle>Metadata</CardTitle>
          <CardDescription>System information</CardDescription>
            </div>
            {merchant.updatedAt && mounted && (
              <p 
                className="text-xs text-muted-foreground mt-1" 
                title={lastUpdatedFull}
                suppressHydrationWarning
              >
                Updated {lastUpdatedRelative || 'Loading...'}
              </p>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell className="font-medium text-muted-foreground w-1/3">Merchant ID</TableCell>
                <TableCell className="font-mono text-sm">{merchant.id}</TableCell>
              </TableRow>
              {merchant.createdBy && (
                <TableRow>
                  <TableCell className="font-medium text-muted-foreground w-1/3">Created By</TableCell>
                  <TableCell>{merchant.createdBy}</TableCell>
                </TableRow>
              )}
              <TableRow>
                <TableCell className="font-medium text-muted-foreground w-1/3">Created</TableCell>
                <TableCell suppressHydrationWarning>{createdDate || 'N/A'}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell className="font-medium text-muted-foreground w-1/3">Last Updated</TableCell>
                <TableCell suppressHydrationWarning>{updatedDate || 'N/A'}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
          
          {/* Metadata JSON Viewer */}
          {merchant.metadata && Object.keys(merchant.metadata).length > 0 && (
            <div className="mt-4">
              <Collapsible>
                <CollapsibleTrigger asChild>
                  <Button variant="outline" className="w-full justify-between">
                    <span>View Metadata JSON</span>
                    <ChevronDown className="h-4 w-4" />
                  </Button>
                </CollapsibleTrigger>
                <CollapsibleContent className="mt-2">
                  <div className="rounded-md bg-muted p-4">
                    <pre className="text-xs overflow-auto max-h-96">
                      {JSON.stringify(merchant.metadata, null, 2)}
                    </pre>
                  </div>
                </CollapsibleContent>
              </Collapsible>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

