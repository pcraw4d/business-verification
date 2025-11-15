'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';

interface RiskIndicatorsTabProps {
  merchantId: string;
}

export function RiskIndicatorsTab({ merchantId }: RiskIndicatorsTabProps) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // TODO: Implement risk indicators API call
    setLoading(false);
  }, [merchantId]);

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Risk Indicators</CardTitle>
          <CardDescription>Active risk indicators for this merchant</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertDescription>Risk indicators feature coming soon</AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    </div>
  );
}

