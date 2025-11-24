'use client';

import { useState } from 'react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { Breadcrumbs, BreadcrumbItemType } from './Breadcrumbs';

interface AppLayoutProps {
  children: React.ReactNode;
  title?: string;
  description?: string;
  breadcrumbs?: BreadcrumbItemType[];
  headerActions?: React.ReactNode;
  className?: string;
}

export function AppLayout({
  children,
  title,
  description,
  breadcrumbs,
  headerActions,
  className,
}: AppLayoutProps) {
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false);

  return (
    <div className="min-h-screen bg-background" suppressHydrationWarning>
      {/* Skip link for keyboard navigation (WCAG 2.4.1) */}
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>
      
      <Sidebar
        mobileOpen={mobileSidebarOpen}
        onMobileClose={() => setMobileSidebarOpen(false)}
      />
      
      <div className="md:pl-64" suppressHydrationWarning>
        <Header
          title={title}
          description={description}
          actions={headerActions}
          onMenuClick={() => setMobileSidebarOpen(true)}
        />
        
        <main id="main-content" className={className} tabIndex={-1} suppressHydrationWarning>
          {breadcrumbs && breadcrumbs.length > 0 && (
            <div className="container px-4 py-4">
              <Breadcrumbs items={breadcrumbs} />
            </div>
          )}
          
          {children}
        </main>
      </div>
    </div>
  );
}

