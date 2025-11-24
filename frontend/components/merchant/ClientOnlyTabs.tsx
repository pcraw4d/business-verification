'use client';

import { useEffect, useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

interface ClientOnlyTabsProps {
  value: string;
  onValueChange: (value: string) => void;
  children: React.ReactNode;
  className?: string;
}

/**
 * ClientOnlyTabs - Wrapper to ensure Tabs component only renders on client
 * This prevents React hydration errors (#418) by ensuring server and client HTML match
 */
export function ClientOnlyTabs({ value, onValueChange, children, className }: ClientOnlyTabsProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return null;
  }

  return (
    <Tabs value={value} onValueChange={onValueChange} className={className}>
      {children}
    </Tabs>
  );
}

