import { Sidebar } from '@/components/layout/Sidebar';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { usePathname } from 'next/navigation';
import { describe, expect, it, vi } from 'vitest';

vi.mock('next/navigation', () => ({
  usePathname: vi.fn(() => '/'),
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
}));

const mockUsePathname = vi.mocked(usePathname);

describe('Sidebar', () => {
  it('should render sidebar with navigation', () => {
    render(<Sidebar />);
    
    expect(screen.getByText('KYB Platform')).toBeInTheDocument();
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Dashboard Hub')).toBeInTheDocument();
  });

  it('should highlight active route', () => {
    mockUsePathname.mockReturnValue('/dashboard');
    
    render(<Sidebar />);
    
    const dashboardLink = screen.getByRole('link', { name: /business intelligence/i });
    expect(dashboardLink).toHaveClass('bg-primary');
  });

  it('should render navigation sections', () => {
    render(<Sidebar />);
    
    expect(screen.getByText('Platform')).toBeInTheDocument();
    expect(screen.getByText('Merchant Verification & Risk')).toBeInTheDocument();
    expect(screen.getByText('Compliance')).toBeInTheDocument();
    expect(screen.getByText('Merchant Management')).toBeInTheDocument();
  });

  it('should render badges for new items', () => {
    render(<Sidebar />);
    
    expect(screen.getByText('NEW')).toBeInTheDocument();
  });

  it('should close mobile sidebar when link is clicked', async () => {
    const user = userEvent.setup();
    const onMobileClose = vi.fn();
    
    render(<Sidebar mobileOpen={true} onMobileClose={onMobileClose} />);
    
    const homeLink = screen.getByRole('link', { name: /home/i });
    await user.click(homeLink);
    
    expect(onMobileClose).toHaveBeenCalled();
  });

  it('should not show mobile sidebar when mobileOpen is false', () => {
    render(<Sidebar mobileOpen={false} />);
    
    // Desktop sidebar should be visible
    const sidebar = document.querySelector('aside');
    expect(sidebar).toBeInTheDocument();
  });

  it('should show mobile sidebar when mobileOpen is true', () => {
    render(<Sidebar mobileOpen={true} />);
    
    // Mobile sidebar (Sheet) should be open
    const sheet = document.querySelector('[role="dialog"]');
    expect(sheet).toBeInTheDocument();
  });
});

