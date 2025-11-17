'use client';

import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from './Sidebar';
import { Menu } from 'lucide-react';

interface HeaderProps {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
  onMenuClick?: () => void;
}

export function Header({ title, description, actions, onMenuClick }: HeaderProps) {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center gap-4 px-4">
        <SidebarTrigger onOpen={onMenuClick || (() => {})} />
        
        <Separator orientation="vertical" className="h-6" />
        
        <div className="flex flex-1 items-center justify-between">
          <div className="flex flex-col">
            {title && (
              <h1 className="text-lg font-semibold">{title}</h1>
            )}
            {description && (
              <p className="text-sm text-muted-foreground">{description}</p>
            )}
          </div>
          
          {actions && (
            <div className="flex items-center gap-2">
              {actions}
            </div>
          )}
        </div>
      </div>
    </header>
  );
}

