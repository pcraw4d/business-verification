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
 * 
 * According to React docs (https://react.dev/errors/418), hydration errors can occur when:
 * - Server/client branch differences
 * - Variable input like Date.now() or Math.random()
 * - Date formatting in user's locale
 * - External changing data without snapshot
 * - Invalid HTML tag nesting
 */
export function ClientOnlyTabs({ value, onValueChange, children, className }: ClientOnlyTabsProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // Use requestAnimationFrame to ensure this runs after initial render
    requestAnimationFrame(() => {
      setMounted(true);
    });
  }, []);

  // Return a placeholder div during SSR to maintain layout
  if (!mounted) {
    return (
      <div className={className} suppressHydrationWarning>
        <div className="h-10 w-full" /> {/* Placeholder for TabsList */}
        <div className="mt-6" /> {/* Placeholder for TabsContent */}
      </div>
    );
  }

  return (
    <Tabs value={value} onValueChange={onValueChange} className={className} suppressHydrationWarning>
      {children}
    </Tabs>
  );
}

