import { render, screen } from '@testing-library/react';
import { Breadcrumbs } from '@/components/layout/Breadcrumbs';
import { describe, it, expect, vi } from 'vitest';

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
  }),
}));

describe('Breadcrumbs', () => {
  it('should render breadcrumb items', () => {
    const items = [
      { label: 'Home', href: '/' },
      { label: 'Dashboard', href: '/dashboard' },
      { label: 'Current Page' },
    ];
    
    render(<Breadcrumbs items={items} />);
    
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Current Page')).toBeInTheDocument();
  });

  it('should render links for items with href', () => {
    const items = [
      { label: 'Home', href: '/' },
      { label: 'Dashboard', href: '/dashboard' },
    ];
    
    render(<Breadcrumbs items={items} />);
    
    const homeLink = screen.getByRole('link', { name: 'Home' });
    const dashboardLink = screen.getByRole('link', { name: 'Dashboard' });
    
    expect(homeLink).toHaveAttribute('href', '/');
    expect(dashboardLink).toHaveAttribute('href', '/dashboard');
  });

  it('should render last item as page (not link)', () => {
    const items = [
      { label: 'Home', href: '/' },
      { label: 'Current Page' },
    ];
    
    render(<Breadcrumbs items={items} />);
    
    const homeLink = screen.getByRole('link', { name: 'Home' });
    const currentPage = screen.getByText('Current Page');
    
    expect(homeLink).toBeInTheDocument();
    expect(currentPage).toBeInTheDocument();
    expect(currentPage.closest('a')).not.toBeInTheDocument();
  });

  it('should render separators between items', () => {
    const items = [
      { label: 'Home', href: '/' },
      { label: 'Dashboard', href: '/dashboard' },
      { label: 'Current Page' },
    ];
    
    render(<Breadcrumbs items={items} />);
    
    // Separators should be present (ChevronRight icons)
    const separators = document.querySelectorAll('svg');
    expect(separators.length).toBeGreaterThan(0);
  });

  it('should handle single breadcrumb item', () => {
    const items = [{ label: 'Home' }];
    
    render(<Breadcrumbs items={items} />);
    
    expect(screen.getByText('Home')).toBeInTheDocument();
  });

  it('should apply custom className', () => {
    const items = [{ label: 'Home' }];
    const { container } = render(
      <Breadcrumbs items={items} className="custom-class" />
    );
    
    const breadcrumb = container.querySelector('nav');
    expect(breadcrumb).toHaveClass('custom-class');
  });

  it('should handle empty items array', () => {
    render(<Breadcrumbs items={[]} />);
    
    // Should render breadcrumb container but no items
    const breadcrumb = screen.queryByRole('navigation');
    expect(breadcrumb).toBeInTheDocument();
  });
});

