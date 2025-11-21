'use client';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { AlertCircle, RefreshCw } from 'lucide-react';

interface TabErrorFallbackProps {
  tabName: string;
  onRetry?: () => void;
  error?: Error;
}

/**
 * Error fallback component for individual tabs
 * Provides user-friendly error messages and retry functionality
 */
export function TabErrorFallback({
  tabName,
  onRetry,
  error,
}: TabErrorFallbackProps) {
  const handleRetry = () => {
    if (onRetry) {
      onRetry();
    } else {
      // Default: reload the page
      window.location.reload();
    }
  };

  return (
    <div className="container mx-auto p-6">
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Error loading {tabName}</AlertTitle>
        <AlertDescription className="mt-2">
          <p className="mb-2">
            An error occurred while loading the {tabName} tab. This tab may be temporarily unavailable.
          </p>
          {process.env.NODE_ENV === 'development' && error && (
            <details className="mt-2 text-sm">
              <summary className="cursor-pointer font-semibold">Error details (development only)</summary>
              <pre className="mt-2 overflow-auto rounded bg-muted p-2 text-xs">
                {error.toString()}
                {error.stack && `\n${error.stack}`}
              </pre>
            </details>
          )}
        </AlertDescription>
      </Alert>
      <div className="mt-4">
        <Button onClick={handleRetry} variant="outline" aria-label={`Retry loading ${tabName} tab`}>
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry
        </Button>
      </div>
    </div>
  );
}

