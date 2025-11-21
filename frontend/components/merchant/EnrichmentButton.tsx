'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Sparkles, Loader2, CheckCircle2, XCircle, RefreshCw, History, Info, Clock } from 'lucide-react';
import { useEffect, useState, useCallback, forwardRef } from 'react';
import { getEnrichmentSources, triggerEnrichment } from '@/lib/api';
import type { EnrichmentSource } from '@/types/merchant';
import { toast } from 'sonner';
import { useEnrichment } from '@/contexts/EnrichmentContext';

interface EnrichmentButtonProps {
  merchantId: string;
  variant?: 'default' | 'outline' | 'ghost';
  size?: 'default' | 'sm' | 'lg' | 'icon';
  className?: string;
}

type EnrichmentJobStatus = 'idle' | 'pending' | 'processing' | 'completed' | 'failed';

interface EnrichmentJob {
  id: string;
  sourceId: string;
  sourceName: string;
  status: EnrichmentJobStatus;
  progress?: number;
  startedAt: Date;
  completedAt?: Date;
  results?: {
    added: string[];
    updated: string[];
    unchanged: string[];
  };
  error?: string;
}

export const EnrichmentButton = forwardRef<HTMLButtonElement, EnrichmentButtonProps>(
  ({ merchantId, variant = 'outline', size = 'default', className = '' }, ref) => {
  const { addEnrichedFields } = useEnrichment();
  const [sources, setSources] = useState<EnrichmentSource[]>([]);
  const [loading, setLoading] = useState(false);
  const [enriching, setEnriching] = useState(false);
  const [open, setOpen] = useState(false);
  const [jobStatus, setJobStatus] = useState<Record<string, EnrichmentJobStatus>>({});
  const [enrichmentHistory, setEnrichmentHistory] = useState<EnrichmentJob[]>([]);
  const [selectedSources, setSelectedSources] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);

  // Client-side formatted timestamps to prevent hydration errors
  const [formattedJobTimestamps, setFormattedJobTimestamps] = useState<Record<string, string>>({});

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
    setMounted(true);
  }, []);

  // Format job timestamps on client side only to prevent hydration errors
  useEffect(() => {
    if (!mounted || enrichmentHistory.length === 0) return;

    const formatted: Record<string, string> = {};
    enrichmentHistory.forEach((job) => {
      if (job.startedAt) {
        try {
          formatted[job.id] = job.startedAt.toLocaleTimeString();
        } catch {
          formatted[job.id] = job.startedAt.toString();
        }
      }
    });
    setFormattedJobTimestamps(formatted);
  }, [mounted, enrichmentHistory]);

  useEffect(() => {
    if (open) {
      loadSources();
    }
  }, [open, loadSources]);

  const handleTriggerEnrichment = useCallback(async (sourceId?: string) => {
    const sourcesToEnrich = sourceId 
      ? [sourceId]
      : Array.from(selectedSources);
    
    if (sourcesToEnrich.length === 0) {
      toast.error('No sources selected', {
        description: 'Please select at least one enrichment source.',
      });
      return;
    }

    try {
      setEnriching(true);
      setError(null);

      // Create jobs for each selected source
      const jobs: EnrichmentJob[] = sourcesToEnrich.map((sid) => {
        const source = sources.find((s) => s.id === sid);
        return {
          id: `job-${Date.now()}-${sid}`,
          sourceId: sid,
          sourceName: source?.name || sid,
          status: 'pending',
          progress: 0,
          startedAt: new Date(),
        };
      });

      // Add jobs to history
      setEnrichmentHistory((prev) => [...jobs, ...prev]);

      // Start enrichment for each source
      for (const job of jobs) {
        setJobStatus((prev) => ({ ...prev, [job.sourceId]: 'pending' }));
        
        try {
          const response = await triggerEnrichment(merchantId, job.sourceId);
          
          setJobStatus((prev) => ({ ...prev, [job.sourceId]: 'processing' }));
          setEnrichmentHistory((prev) =>
            prev.map((j) =>
              j.id === job.id ? { ...j, status: 'processing', progress: 25 } : j
            )
          );

          toast.success('Enrichment job started', {
            description: `Job ID: ${response.jobId || job.id}`,
          });

          // Simulate job progress (in real implementation, poll for status)
          const progressSteps = [25, 50, 75, 100];
          for (const progress of progressSteps) {
            await new Promise((resolve) => setTimeout(resolve, 500));
            setEnrichmentHistory((prev) =>
              prev.map((j) =>
                j.id === job.id ? { ...j, progress } : j
              )
            );
          }

          // Simulate completion with results
          const mockResults = {
            added: ['Founded Date', 'Employee Count'],
            updated: ['Annual Revenue'],
            unchanged: ['Business Name', 'Address'],
          };

          setJobStatus((prev) => ({ ...prev, [job.sourceId]: 'completed' }));
          setEnrichmentHistory((prev) =>
            prev.map((j) =>
              j.id === job.id
                ? {
                    ...j,
                    status: 'completed',
                    progress: 100,
                    completedAt: new Date(),
                    results: mockResults,
                  }
                : j
            )
          );

          // Track enriched fields for highlighting
          const enrichedFieldsData = [
            ...mockResults.added.map((name) => ({ name, type: 'added' as const, source: job.sourceName })),
            ...mockResults.updated.map((name) => ({ name, type: 'updated' as const, source: job.sourceName })),
          ];
          addEnrichedFields(merchantId, enrichedFieldsData);

          toast.success('Enrichment completed', {
            description: `Data from ${job.sourceName} has been enriched`,
          });
        } catch (err) {
          const errorMessage = err instanceof Error ? err.message : 'Failed to trigger enrichment';
          setJobStatus((prev) => ({ ...prev, [job.sourceId]: 'failed' }));
          setEnrichmentHistory((prev) =>
            prev.map((j) =>
              j.id === job.id
                ? {
                    ...j,
                    status: 'failed',
                    error: errorMessage,
                    completedAt: new Date(),
                  }
                : j
            )
          );
          toast.error('Enrichment failed', {
            description: `${job.sourceName}: ${errorMessage}`,
          });
        }
      }

      // Clear selection after completion
      setSelectedSources(new Set());
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to trigger enrichment';
      setError(errorMessage);
      toast.error('Enrichment failed', {
        description: errorMessage,
      });
    } finally {
      setEnriching(false);
    }
  }, [merchantId, sources, selectedSources]);

  const toggleSourceSelection = useCallback((sourceId: string) => {
    setSelectedSources((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(sourceId)) {
        newSet.delete(sourceId);
      } else {
        newSet.add(sourceId);
      }
      return newSet;
    });
  }, []);

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
        <Button 
          ref={ref}
          variant={variant} 
          size={size} 
          className={className}
          aria-label="Enrich merchant data from third-party vendors (Press E)"
          title="Enrich data (E)"
        >
          <Sparkles className="h-4 w-4 mr-2" aria-hidden="true" />
          Enrich Data
          {hasEnabledSources && (
            <Badge variant="secondary" className="ml-2" aria-label={`${enabledSources.length} sources available`}>
              {enabledSources.length}
            </Badge>
          )}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[700px] max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Data Enrichment</DialogTitle>
          <DialogDescription>
            Select data sources to enrich merchant information from third-party vendors
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="sources" className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="sources">Enrichment Sources</TabsTrigger>
            <TabsTrigger value="history">
              History
              {enrichmentHistory.length > 0 && (
                <Badge variant="secondary" className="ml-2">
                  {enrichmentHistory.length}
                </Badge>
              )}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="sources" className="space-y-4">
        
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
          <div className="space-y-4">
            <div className="space-y-2">
              {enabledSources.map((source) => {
                const status = jobStatus[source.id] || 'idle';
                const isProcessing = status === 'pending' || status === 'processing';
                const isSelected = selectedSources.has(source.id);
                
                return (
                  <Card
                    key={source.id}
                    className={`cursor-pointer transition-all ${
                      isSelected ? 'border-primary bg-primary/5' : ''
                    } ${isProcessing ? 'opacity-75' : ''}`}
                    onClick={() => !isProcessing && toggleSourceSelection(source.id)}
                  >
                    <CardContent className="p-4">
                      <div className="flex items-start justify-between">
                        <div className="flex-1 space-y-2">
                          <div className="flex items-center gap-2">
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={() => toggleSourceSelection(source.id)}
                              disabled={isProcessing}
                              className="h-4 w-4"
                            />
                            <span className="font-medium">{source.name}</span>
                            {getStatusIcon(status)}
                            {getStatusBadge(status)}
                          </div>
                          {source.description && (
                            <p className="text-sm text-muted-foreground">{source.description}</p>
                          )}
                          {source.dataProvided && (
                            <div className="flex flex-wrap gap-1 mt-2">
                              {source.dataProvided.map((data, idx) => (
                                <Badge key={idx} variant="outline" className="text-xs">
                                  {data}
                                </Badge>
                              ))}
                            </div>
                          )}
                        </div>
                        {isProcessing && (
                          <div className="ml-4 w-32">
                            <Progress value={enrichmentHistory.find(j => j.sourceId === source.id)?.progress || 0} />
                          </div>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>

            <div className="flex items-center justify-between pt-4 border-t">
              <div className="text-sm text-muted-foreground">
                {selectedSources.size > 0
                  ? `${selectedSources.size} source${selectedSources.size !== 1 ? 's' : ''} selected`
                  : 'Select one or more sources to enrich'}
              </div>
              <Button
                onClick={() => handleTriggerEnrichment()}
                disabled={enriching || selectedSources.size === 0}
                className="ml-auto"
              >
                {enriching ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Processing...
                  </>
                ) : (
                  <>
                    <Sparkles className="h-4 w-4 mr-2" />
                    Enrich Selected ({selectedSources.size})
                  </>
                )}
              </Button>
            </div>
          </div>
        )}
          </TabsContent>

          <TabsContent value="history" className="space-y-4">
            {enrichmentHistory.length === 0 ? (
              <div className="text-center py-8 text-sm text-muted-foreground">
                <History className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>No enrichment history yet.</p>
                <p className="text-xs mt-1">Enrichment jobs will appear here after they are started.</p>
              </div>
            ) : (
              <div className="space-y-3">
                {enrichmentHistory.map((job) => (
                  <Card key={job.id}>
                    <CardHeader className="pb-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <CardTitle className="text-sm">{job.sourceName}</CardTitle>
                          {getStatusBadge(job.status)}
                        </div>
                        <div className="text-xs text-muted-foreground flex items-center gap-1" suppressHydrationWarning>
                          <Clock className="h-3 w-3" />
                          {mounted && formattedJobTimestamps[job.id] ? formattedJobTimestamps[job.id] : job.startedAt.toString()}
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="space-y-3">
                      {job.status === 'processing' && job.progress !== undefined && (
                        <div className="space-y-1">
                          <div className="flex items-center justify-between text-xs">
                            <span>Progress</span>
                            <span>{job.progress}%</span>
                          </div>
                          <Progress value={job.progress} />
                        </div>
                      )}
                      
                      {job.results && (
                        <div className="space-y-2 text-sm">
                          {job.results.added.length > 0 && (
                            <div>
                              <span className="font-medium text-green-600">Added: </span>
                              <span className="text-muted-foreground">
                                {job.results.added.join(', ')}
                              </span>
                            </div>
                          )}
                          {job.results.updated.length > 0 && (
                            <div>
                              <span className="font-medium text-blue-600">Updated: </span>
                              <span className="text-muted-foreground">
                                {job.results.updated.join(', ')}
                              </span>
                            </div>
                          )}
                          {job.results.unchanged.length > 0 && (
                            <div>
                              <span className="font-medium text-muted-foreground">Unchanged: </span>
                              <span className="text-muted-foreground">
                                {job.results.unchanged.join(', ')}
                              </span>
                            </div>
                          )}
                        </div>
                      )}

                      {job.error && (
                        <Alert variant="destructive">
                          <AlertDescription className="text-xs">{job.error}</AlertDescription>
                        </Alert>
                      )}

                      {job.status === 'failed' && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleTriggerEnrichment(job.sourceId)}
                          disabled={enriching}
                        >
                          <RefreshCw className="h-3 w-3 mr-2" />
                          Retry
                        </Button>
                      )}
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </TabsContent>
        </Tabs>

        {error && sources.length > 0 && (
          <Alert variant="destructive" className="mt-4">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </DialogContent>
    </Dialog>
  );
  }
);

EnrichmentButton.displayName = 'EnrichmentButton';
