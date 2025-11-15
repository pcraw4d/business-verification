'use client';

import { AlertCircle, FileX, SearchX } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface EmptyStateProps {
  type?: 'noData' | 'error' | 'noResults';
  title?: string;
  message?: string;
  actionLabel?: string;
  onAction?: () => void;
  className?: string;
}

const icons = {
  noData: FileX,
  error: AlertCircle,
  noResults: SearchX,
};

const defaultTitles = {
  noData: 'No Data Available',
  error: 'Something Went Wrong',
  noResults: 'No Results Found',
};

const defaultMessages = {
  noData: 'There is no data to display at this time.',
  error: 'An error occurred while loading data. Please try again.',
  noResults: 'No results match your search criteria.',
};

export function EmptyState({
  type = 'noData',
  title,
  message,
  actionLabel,
  onAction,
  className,
}: EmptyStateProps) {
  const Icon = icons[type];
  const displayTitle = title || defaultTitles[type];
  const displayMessage = message || defaultMessages[type];

  return (
    <Card className={cn('border-dashed', className)}>
      <CardContent className="flex flex-col items-center justify-center py-12 text-center">
        <Icon className="h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-semibold mb-2">{displayTitle}</h3>
        <CardDescription className="mb-4">{displayMessage}</CardDescription>
        {onAction && actionLabel && (
          <Button variant="outline" onClick={onAction}>
            {actionLabel}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

