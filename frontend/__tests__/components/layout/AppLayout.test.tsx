import { AppLayout } from '@/components/layout/AppLayout';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

vi.mock('next/navigation', () => ({
  usePathname: () => '/',
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
}));

describe('AppLayout', () => {
  it('should render children', () => {
    render(
      <AppLayout>
        <div>Test Content</div>
      </AppLayout>
    );
    
    expect(screen.getByText('Test Content')).toBeInTheDocument();
  });

  it('should render skip link for accessibility', () => {
    render(<AppLayout>Content</AppLayout>);
    
    const skipLink = screen.getByText('Skip to main content');
    expect(skipLink).toBeInTheDocument();
    expect(skipLink).toHaveAttribute('href', '#main-content');
  });

  it('should render sidebar', () => {
    render(<AppLayout>Content</AppLayout>);
    
    // Sidebar should be rendered (check for KYB Platform text)
    expect(screen.getByText('KYB Platform')).toBeInTheDocument();
  });

  it('should render header with title', () => {
    render(<AppLayout title="Test Title">Content</AppLayout>);
    
    expect(screen.getByText('Test Title')).toBeInTheDocument();
  });

  it('should render header with description', () => {
    render(
      <AppLayout title="Test Title" description="Test Description">
        Content
      </AppLayout>
    );
    
    expect(screen.getByText('Test Description')).toBeInTheDocument();
  });

  it('should render header actions', () => {
    render(
      <AppLayout
        title="Test Title"
        headerActions={<button>Action Button</button>}
      >
        Content
      </AppLayout>
    );
    
    expect(screen.getByText('Action Button')).toBeInTheDocument();
  });

  it('should render breadcrumbs when provided', () => {
    const breadcrumbs = [
      { label: 'Home', href: '/' },
      { label: 'Dashboard', href: '/dashboard' },
      { label: 'Current Page' },
    ];
    
    render(
      <AppLayout breadcrumbs={breadcrumbs}>
        Content
      </AppLayout>
    );
    
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Current Page')).toBeInTheDocument();
  });

  it('should not render breadcrumbs when not provided', () => {
    render(<AppLayout>Content</AppLayout>);
    
    // Breadcrumbs container should not be visible
    const breadcrumbContainer = screen.queryByRole('navigation', { name: /breadcrumb/i });
    expect(breadcrumbContainer).not.toBeInTheDocument();
  });

  it('should apply custom className to main content', () => {
    const { container } = render(
      <AppLayout className="custom-class">
        Content
      </AppLayout>
    );
    
    const main = container.querySelector('main');
    expect(main).toHaveClass('custom-class');
  });

  it('should have main content with proper id for skip link', () => {
    render(<AppLayout>Content</AppLayout>);
    
    const main = screen.getByRole('main');
    expect(main).toHaveAttribute('id', 'main-content');
  });

  it('should toggle mobile sidebar', async () => {
    const user = userEvent.setup();
    render(<AppLayout>Content</AppLayout>);
    
    // Find menu button (mobile sidebar trigger)
    const menuButton = screen.getByLabelText(/toggle sidebar/i);
    await user.click(menuButton);
    
    // Mobile sidebar should open (check for navigation items)
    await waitFor(() => {
      expect(screen.getByText('Home')).toBeInTheDocument();
    });
  });
});

