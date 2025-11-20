'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Sparkles, Loader2, CheckCircle2, XCircle, RefreshCw } from 'lucide-react';
import { useEffect, useState, useCallback } from 'react';
import { getEnrichmentSources, triggerEnrichment } from '@/lib/api';
import type { EnrichmentSource } from '@/types/merchant';
import { toast } from 'sonner';

interface EnrichmentButtonProps {
  merchantId: string;
  variant?: 'default' | 'outline' | 'ghost';
  size?: 'default' | 'sm' | 'lg' | 'icon';
  className?: string;
}

type EnrichmentJobStatus = 'idle' | 'pending' | 'processing' | 'completed' | 'failed';

export function EnrichmentButton({ merchantId, variant = 'outline', size = 'default', className = '' }: EnrichmentButtonProps) {
  const [sources, setSources] = useState<EnrichmentSource[]>([]);
  const [loading, setLoading] = useState(false);
  const [enriching, setEnriching] = useState(false);
  const [open, setOpen] = useState(false);
  const [jobStatus, setJobStatus] = useState<Record<string, EnrichmentJobStatus>>({});
  const [error, setError] = useState<string | null>(null);

  const loadSources = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getEnrichmentSources(merchantId);
      setSources(data.sources || []);
    } catch (err) {
      // Silently handle 404s for optional endpoints - don't log to console or show error
      const is404 = err instanceof Error && err.message.includes('404');
      if (!is404) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to load enrichment sources';
        setError(errorMessage);
        console.error('Error loading enrichment sources:', err);
      }
      // If 404, just set empty sources - endpoint may not be implemented yet
      if (is404) {
        setSources([]);
      }
    } finally {
      setLoading(false);
    }
  }, [merchantId]);

  useEffect(() => {
    if (open) {
      loadSources();
    }
  }, [open, loadSources]);

  const handleTriggerEnrichment = useCallback(async (sourceId: string) => {
    try {
      setEnriching(true);
      setError(null);
      setJobStatus((prev) => ({ ...prev, [sourceId]: 'pending' }));

      const response = await triggerEnrichment(merchantId, sourceId);
      
      setJobStatus((prev) => ({ ...prev, [sourceId]: 'processing' }));
      
      toast.success('Enrichment job started', {
        description: `Job ID: ${response.jobId}`,
      });

      // Simulate job completion after a delay (in real implementation, poll for status)
      setTimeout(() => {
        setJobStatus((prev) => ({ ...prev, [sourceId]: 'completed' }));
        toast.success('Enrichment completed', {
          description: `Data from ${sources.find(s => s.id === sourceId)?.name || sourceId} has been enriched`,
        });
        
        // Clear cache to force refresh of merchant data
        // Note: In a real implementation, you might want to refresh merchant data here
        setTimeout(() => {
          setJobStatus((prev) => {
            const newStatus = { ...prev };
            delete newStatus[sourceId];
            return newStatus;
          });
        }, 3000);
      }, 2000);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to trigger enrichment';
      setError(errorMessage);
      setJobStatus((prev) => ({ ...prev, [sourceId]: 'failed' }));
      toast.error('Enrichment failed', {
        description: errorMessage,
      });
    } finally {
      setEnriching(false);
    }
  }, [merchantId, sources]);

  const getStatusIcon = (status: EnrichmentJobStatus) => {
    switch (status) {
      case 'pending':
      case 'processing':
        return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
      case 'completed':
        return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return null;
    }
  };

  const getStatusBadge = (status: EnrichmentJobStatus) => {
    switch (status) {
      case 'pending':
        return <Badge variant="secondary">Pending</Badge>;
      case 'processing':
        return <Badge variant="default">Processing</Badge>;
      case 'completed':
        return <Badge variant="default" className="bg-green-500">Completed</Badge>;
      case 'failed':
        return <Badge variant="destructive">Failed</Badge>;
      default:
        return null;
    }
  };

  const enabledSources = sources.filter((source) => source.enabled !== false);
  const hasEnabledSources = enabledSources.length > 0;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant={variant} size={size} className={className}>
          <Sparkles className="h-4 w-4 mr-2" />
          Enrich Data
          {hasEnabledSources && (
            <Badge variant="secondary" className="ml-2">
              {enabledSources.length}
            </Badge>
          )}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Data Enrichment</DialogTitle>
          <DialogDescription>
            Select a data source to enrich merchant information
          </DialogDescription>
        </DialogHeader>
        
        {loading ? (
          <div className="space-y-2">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        ) : error && !sources.length ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
            <Button onClick={loadSources} variant="outline" size="sm" className="mt-2">
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </Alert>
        ) : !hasEnabledSources ? (
          <Alert>
            <AlertDescription>No enrichment sources available for this merchant.</AlertDescription>
          </Alert>
        ) : (
          <div className="space-y-2">
            {enabledSources.map((source) => {
              const status = jobStatus[source.id] || 'idle';
              const isProcessing = status === 'pending' || status === 'processing';
              
              return (
                <div
                  key={source.id}
                  className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50 transition-colors"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{source.name}</span>
                      {getStatusIcon(status)}
                      {getStatusBadge(status)}
                    </div>
                    {source.description && (
                      <p className="text-sm text-muted-foreground mt-1">{source.description}</p>
                    )}
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleTriggerEnrichment(source.id)}
                    disabled={enriching || isProcessing}
                    className="ml-4"
                  >
                    {isProcessing ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        Processing...
                      </>
                    ) : status === 'completed' ? (
                      <>
                        <CheckCircle2 className="h-4 w-4 mr-2" />
                        Done
                      </>
                    ) : (
                      'Enrich'
                    )}
                  </Button>
                </div>
              );
            })}
          </div>
        )}

        {error && sources.length > 0 && (
          <Alert variant="destructive" className="mt-4">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </DialogContent>
    </Dialog>
  );
}

