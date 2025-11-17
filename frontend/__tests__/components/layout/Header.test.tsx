import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Header } from '@/components/layout/Header';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('next/navigation', () => ({
  usePathname: () => '/',
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
}));

describe('Header', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });
  it('should render header', () => {
    render(<Header />);
    
    const header = screen.getByRole('banner');
    expect(header).toBeInTheDocument();
  });

  it('should render title when provided', () => {
    render(<Header title="Test Title" />);
    
    expect(screen.getByText('Test Title')).toBeInTheDocument();
  });

  it('should render description when provided', () => {
    render(<Header title="Test Title" description="Test Description" />);
    
    expect(screen.getByText('Test Description')).toBeInTheDocument();
  });

  it('should render header actions when provided', () => {
    render(
      <Header
        title="Test Title"
        actions={<button>Action Button</button>}
      />
    );
    
    expect(screen.getByText('Action Button')).toBeInTheDocument();
  });

  it('should render sidebar trigger button', () => {
    render(<Header />);
    
    const menuButton = screen.getByLabelText(/toggle sidebar/i);
    expect(menuButton).toBeInTheDocument();
  });

  it('should call onMenuClick when menu button is clicked', async () => {
    const user = userEvent.setup();
    const onMenuClick = vi.fn();
    
    render(<Header onMenuClick={onMenuClick} />);
    
    const menuButton = screen.getByLabelText(/toggle sidebar/i);
    await user.click(menuButton);
    
    expect(onMenuClick).toHaveBeenCalled();
  });

  it('should not render title when not provided', () => {
    render(<Header />);
    
    const title = screen.queryByRole('heading', { level: 1 });
    expect(title).not.toBeInTheDocument();
  });

  it('should not render description when not provided', () => {
    render(<Header title="Test Title" />);
    
    const description = screen.queryByText(/test description/i);
    expect(description).not.toBeInTheDocument();
  });
});

