'use client';

import { useState, useEffect, useCallback } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { PlusCircle, Search, Store, Filter, AlertCircle, ChevronLeft, ChevronRight } from 'lucide-react';
import Link from 'next/link';
import { useRouter, usePathname } from 'next/navigation';
import { getMerchantsList, getMerchant } from '@/lib/api';
import dynamic from 'next/dynamic';
import type { MerchantListItem, MerchantListResponse } from '@/types/merchant';
import { MerchantDetailsLayout } from '@/components/merchant/MerchantDetailsLayout';

// Lazy load ExportButton (includes heavy libraries like xlsx, jspdf)
const ExportButton = dynamic(
  () => import('@/components/export/ExportButton').then((mod) => ({ default: mod.ExportButton })),
  {
    loading: () => <div className="h-9 w-24 animate-pulse bg-muted rounded" />,
    ssr: false,
  }
);
import { toast } from 'sonner';

export default function MerchantPortfolioPage() {
  const router = useRouter();
  const pathname = usePathname();
  const [merchants, setMerchants] = useState<MerchantListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [riskLevelFilter, setRiskLevelFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [hasPrevious, setHasPrevious] = useState(false);
  const [prefetchingMerchantId, setPrefetchingMerchantId] = useState<string | null>(null);

  // Client-side route detection: If this page is being served for a merchant-details route,
  // automatically route to the merchant-details page
  // This happens when the Go service serves merchant-portfolio page for merchant-details routes
  const [isRoutingToDetails, setIsRoutingToDetails] = useState(false);
  
  useEffect(() => {
    // Check if current pathname is a merchant-details route
    // Also check window.location as fallback (in case pathname isn't set yet)
    const currentPath = pathname || (typeof window !== 'undefined' ? window.location.pathname : '');
    
    if (currentPath && currentPath.startsWith('/merchant-details/')) {
      // Extract merchant ID from pathname
      const merchantId = currentPath.replace('/merchant-details/', '').split('/')[0].split('?')[0];
      if (merchantId && !isRoutingToDetails) {
        setIsRoutingToDetails(true);
        // Route to the merchant-details page - Next.js will handle client-side routing
        // Use replace to avoid adding to history and prevent back button issues
        // Use a small delay to ensure Next.js router is fully initialized
        const timer = setTimeout(() => {
          // Try router.replace first (cleaner, no history entry)
          // If that doesn't work, Next.js router should still handle it via the URL
          router.replace(`/merchant-details/${merchantId}`);
          
          // Fallback: If router.replace doesn't work, force navigation after a short delay
          // This ensures the merchant-details page loads even if client-side routing fails
          setTimeout(() => {
            const stillOnWrongPage = typeof window !== 'undefined' && 
              window.location.pathname.startsWith('/merchant-details/') &&
              !window.location.pathname.includes('merchant-portfolio');
            if (stillOnWrongPage && window.location.pathname !== `/merchant-details/${merchantId}`) {
              // Force navigation - this will trigger Next.js to load the correct page
              window.history.replaceState({}, '', `/merchant-details/${merchantId}`);
              // Trigger a route change event
              window.dispatchEvent(new PopStateEvent('popstate'));
            }
          }, 100);
        }, 10);
        return () => clearTimeout(timer);
      }
    }
  }, [pathname, router, isRoutingToDetails]);

  // If we're on a merchant-details route, render the merchant-details component directly
  // This happens when the Go service serves merchant-portfolio page for merchant-details routes
  const currentPath = pathname || (typeof window !== 'undefined' ? window.location.pathname : '');
  if (currentPath && currentPath.startsWith('/merchant-details/')) {
    // Extract merchant ID from pathname
    const merchantId = currentPath.replace('/merchant-details/', '').split('/')[0].split('?')[0];
    if (merchantId) {
      // Directly render the merchant-details component instead of trying to route
      // This ensures the merchant-details page is displayed even when served as merchant-portfolio HTML
      return <MerchantDetailsLayout merchantId={merchantId} />;
    }
  }

  // Stats
  const [stats, setStats] = useState({
    total: 0,
    verified: 0,
    pending: 0,
    highRisk: 0,
  });

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchTerm);
      setPage(1); // Reset to first page on search
    }, 500);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  // Fetch merchants
  const fetchMerchants = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const params = {
        page,
        pageSize,
        search: debouncedSearch || undefined,
        status: statusFilter !== 'all' ? statusFilter : undefined,
        riskLevel: riskLevelFilter !== 'all' ? riskLevelFilter : undefined,
        sortBy: 'created_at',
        sortOrder: 'desc' as const,
      };

      const response: MerchantListResponse = await getMerchantsList(params);
      
      setMerchants(response.merchants);
      setTotal(response.total);
      setTotalPages(response.total_pages);
      setHasNext(response.has_next);
      setHasPrevious(response.has_previous);

      // Calculate stats from current page data (or fetch separately if needed)
      const verified = response.merchants.filter(m => m.status === 'active' || m.compliance_status === 'compliant').length;
      const pending = response.merchants.filter(m => m.status === 'pending').length;
      const highRisk = response.merchants.filter(m => m.risk_level === 'high' || m.risk_level === 'critical').length;

      setStats({
        total: response.total,
        verified,
        pending,
        highRisk,
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load merchants';
      setError(errorMessage);
      toast.error('Failed to load merchants', {
        description: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, debouncedSearch, statusFilter, riskLevelFilter]);

  useEffect(() => {
    fetchMerchants();
  }, [fetchMerchants]);

  // Prefetch merchant data on hover
  const handleMerchantHover = useCallback(async (merchantId: string) => {
    // Only prefetch if not already prefetching this merchant
    if (prefetchingMerchantId === merchantId) {
      return;
    }

    setPrefetchingMerchantId(merchantId);
    
    try {
      // Prefetch merchant data - this will cache it for when user clicks
      await getMerchant(merchantId);
      
      // Prefetch the route - Next.js will prefetch the page
      router.prefetch(`/merchant-details/${merchantId}`);
    } catch (error) {
      // Silently fail - prefetching is optional
      if (process.env.NODE_ENV === 'development') {
        console.debug('Prefetch failed for merchant:', merchantId, error);
      }
    } finally {
      setPrefetchingMerchantId(null);
    }
  }, [router, prefetchingMerchantId]);

  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      });
    } catch {
      return dateString;
    }
  };

  const getRiskBadgeVariant = (riskLevel?: string) => {
    if (!riskLevel) return 'default';
    if (riskLevel === 'high' || riskLevel === 'critical') return 'destructive';
    if (riskLevel === 'medium') return 'secondary';
    return 'default';
  };

  const getStatusBadgeVariant = (status: string) => {
    if (status === 'active') return 'default';
    if (status === 'pending') return 'secondary';
    if (status === 'suspended' || status === 'inactive') return 'destructive';
    return 'outline';
  };

  return (
    <AppLayout
      title="Merchant Portfolio"
      description="Manage and view all merchants in your portfolio"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Merchant Portfolio' },
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
        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Total Merchants</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <Skeleton className="h-8 w-16" />
              ) : (
                <div className="text-2xl font-bold">{stats.total}</div>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Verified</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <Skeleton className="h-8 w-16" />
              ) : (
                <div className="text-2xl font-bold">{stats.verified}</div>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Pending</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <Skeleton className="h-8 w-16" />
              ) : (
                <div className="text-2xl font-bold">{stats.pending}</div>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>High Risk</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <Skeleton className="h-8 w-16" />
              ) : (
                <div className="text-2xl font-bold">{stats.highRisk}</div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Search and Filters */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Merchants</CardTitle>
                <CardDescription>Search and filter your merchant portfolio</CardDescription>
              </div>
              {!loading && merchants.length > 0 && (
                <ExportButton
                  data={async () => {
                    // Export all merchants (fetch all pages if needed)
                    const allMerchants: MerchantListItem[] = [];
                    let currentPage = 1;
                    let hasMore = true;
                    
                    while (hasMore && currentPage <= 10) { // Limit to 10 pages to avoid too many requests
                      const response = await getMerchantsList({
                        page: currentPage,
                        pageSize: 100,
                        search: debouncedSearch || undefined,
                        status: statusFilter !== 'all' ? statusFilter : undefined,
                        riskLevel: riskLevelFilter !== 'all' ? riskLevelFilter : undefined,
                      });
                      allMerchants.push(...response.merchants);
                      hasMore = response.has_next;
                      currentPage++;
                    }
                    
                    return {
                      merchants: allMerchants,
                      filters: {
                        search: debouncedSearch,
                        status: statusFilter,
                        riskLevel: riskLevelFilter,
                      },
                      exportedAt: new Date().toISOString(),
                      total: allMerchants.length,
                    };
                  }}
                  exportType="merchant"
                  formats={['csv', 'json', 'excel', 'pdf']}
                />
              )}
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col md:flex-row gap-4 mb-4">
              <div className="flex-1 relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" aria-hidden="true" />
                <Input
                  placeholder="Search merchants..."
                  className="pl-10"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  aria-label="Search merchants"
                />
              </div>
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                  <SelectItem value="suspended">Suspended</SelectItem>
                  <SelectItem value="inactive">Inactive</SelectItem>
                </SelectContent>
              </Select>
              <Select value={riskLevelFilter} onValueChange={setRiskLevelFilter}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Risk Level" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Risk Levels</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="critical">Critical</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {error && (
              <Alert variant="destructive" className="mb-4">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {/* Merchants Table */}
            <div className="border rounded-lg">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Business Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Risk Level</TableHead>
                    <TableHead>Last Updated</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {loading ? (
                    Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                      </TableRow>
                    ))
                  ) : merchants.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                        <Store className="h-12 w-12 mx-auto mb-4 opacity-50" />
                        <p>No merchants found</p>
                        {debouncedSearch || statusFilter !== 'all' || riskLevelFilter !== 'all' ? (
                          <Button
                            variant="outline"
                            className="mt-4"
                            onClick={() => {
                              setSearchTerm('');
                              setStatusFilter('all');
                              setRiskLevelFilter('all');
                            }}
                            aria-label="Clear all filters"
                          >
                            Clear Filters
                          </Button>
                        ) : (
                          <Button asChild className="mt-4" variant="outline" aria-label="Add your first merchant">
                            <Link href="/add-merchant">Add Your First Merchant</Link>
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ) : (
                    merchants.map((merchant) => (
                      <TableRow 
                        key={merchant.id}
                        onMouseEnter={() => handleMerchantHover(merchant.id)}
                        className="cursor-pointer hover:bg-muted/50 transition-colors"
                      >
                        <TableCell className="font-medium">{merchant.name}</TableCell>
                        <TableCell>
                          <Badge variant={getStatusBadgeVariant(merchant.status)}>
                            {merchant.status}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={getRiskBadgeVariant(merchant.risk_level)}>
                            {merchant.risk_level || 'N/A'}
                          </Badge>
                        </TableCell>
                        <TableCell>{formatDate(merchant.updated_at)}</TableCell>
                        <TableCell className="text-right">
                          <Button asChild variant="ghost" size="sm" aria-label={`View details for ${merchant.name}`}>
                            <Link href={`/merchant-details/${merchant.id}`}>View</Link>
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>

            {/* Pagination */}
            {!loading && merchants.length > 0 && (
              <div className="flex items-center justify-between mt-4">
                <div className="text-sm text-muted-foreground">
                  Showing {merchants.length} of {total} merchants
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={!hasPrevious}
                    aria-label="Go to previous page"
                  >
                    <ChevronLeft className="h-4 w-4" />
                    Previous
                  </Button>
                  <div className="flex items-center gap-2">
                    <span className="text-sm">
                      Page {page} of {totalPages}
                    </span>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    disabled={!hasNext}
                    aria-label="Go to next page"
                  >
                    Next
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}
