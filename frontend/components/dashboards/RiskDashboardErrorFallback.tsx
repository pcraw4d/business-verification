'use client';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { AlertCircle, RefreshCw } from 'lucide-react';

interface RiskDashboardErrorFallbackProps {
  error: Error | null;
  resetError: () => void;
}

/**
 * Error fallback component for Risk Dashboard page
 * Provides user-friendly error message and retry functionality
 */
export function RiskDashboardErrorFallback({ error, resetError }: RiskDashboardErrorFallbackProps) {
  return (
    <div className="container mx-auto p-6">
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Failed to Load Risk Dashboard</AlertTitle>
        <AlertDescription className="mt-2">
          <p className="mb-2">
            We couldn't load the risk dashboard data. This might be due to a network issue or a temporary server problem.
          </p>
          {process.env.NODE_ENV === 'development' && error && (
            <details className="mt-2 text-sm">
              <summary className="cursor-pointer font-semibold">Error details (development only)</summary>
              <pre className="mt-2 overflow-auto rounded bg-muted p-2 text-xs">
                {error.toString()}
              </pre>
            </details>
          )}
        </AlertDescription>
      </Alert>
      <div className="mt-4">
        <Button onClick={resetError} variant="outline">
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry
        </Button>
      </div>
    </div>
  );
}

