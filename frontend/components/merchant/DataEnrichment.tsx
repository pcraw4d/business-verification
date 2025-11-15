'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { toast } from 'sonner';

interface DataEnrichmentProps {
  merchantId: string;
}

interface EnrichmentSource {
  id: string;
  name: string;
  description?: string;
}

export function DataEnrichment({ merchantId }: DataEnrichmentProps) {
  const [sources, setSources] = useState<EnrichmentSource[]>([]);
  const [loading, setLoading] = useState(false);
  const [enriching, setEnriching] = useState(false);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    loadSources();
  }, [merchantId]);

  async function loadSources() {
    try {
      setLoading(true);
      const token = typeof window !== 'undefined' ? sessionStorage.getItem('authToken') : null;
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch(`/api/v1/merchants/${merchantId}/enrichment/sources`, {
        headers,
      });

      if (response.ok) {
        const data = await response.json();
        setSources(data.sources || []);
      }
    } catch (error) {
      console.error('Error loading sources:', error);
    } finally {
      setLoading(false);
    }
  }

  async function triggerEnrichment(source: string) {
    try {
      setEnriching(true);
      const token = typeof window !== 'undefined' ? sessionStorage.getItem('authToken') : null;
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch(`/api/v1/merchants/${merchantId}/enrichment/trigger`, {
        method: 'POST',
        headers,
        body: JSON.stringify({ source }),
      });

      if (response.ok) {
        const job = await response.json();
        toast.success('Enrichment job started successfully');
        setOpen(false);
      } else {
        throw new Error('Failed to trigger enrichment');
      }
    } catch (error) {
      toast.error('Failed to trigger enrichment');
      console.error('Enrichment error:', error);
    } finally {
      setEnriching(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">Enrich Data</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Data Enrichment</DialogTitle>
          <DialogDescription>Select a data source to enrich merchant information</DialogDescription>
        </DialogHeader>
        {loading ? (
          <div className="space-y-2">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        ) : sources.length === 0 ? (
          <Alert>
            <AlertDescription>No enrichment sources available</AlertDescription>
          </Alert>
        ) : (
          <div className="space-y-2">
            {sources.map((source) => (
              <Button
                key={source.id}
                variant="outline"
                className="w-full justify-start"
                onClick={() => triggerEnrichment(source.id)}
                disabled={enriching}
              >
                {source.name}
              </Button>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

