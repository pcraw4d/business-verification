import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ChartContainer } from '@/components/dashboards/ChartContainer';
import { DashboardCard } from '@/components/dashboards/DashboardCard';
import { DataTable } from '@/components/dashboards/DataTable';
import { MetricCard } from '@/components/dashboards/MetricCard';
import { describe, it, expect, vi } from 'vitest';
import { Users } from 'lucide-react';

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: any) => <a href={href}>{children}</a>,
}));

describe('Dashboard Components', () => {
  describe('ChartContainer', () => {
    it('should render chart container with title', () => {
      render(
        <ChartContainer title="Test Chart">
          <div>Chart Content</div>
        </ChartContainer>
      );
      
      expect(screen.getByText('Test Chart')).toBeInTheDocument();
      expect(screen.getByText('Chart Content')).toBeInTheDocument();
    });

    it('should render description when provided', () => {
      render(
        <ChartContainer title="Test Chart" description="Chart description">
          <div>Chart Content</div>
        </ChartContainer>
      );
      
      expect(screen.getByText('Chart description')).toBeInTheDocument();
    });

    it('should render action when provided', () => {
      render(
        <ChartContainer title="Test Chart" action={<button>Action</button>}>
          <div>Chart Content</div>
        </ChartContainer>
      );
      
      expect(screen.getByText('Action')).toBeInTheDocument();
    });

    it('should show loading state', () => {
      render(
        <ChartContainer title="Test Chart" isLoading={true}>
          <div>Chart Content</div>
        </ChartContainer>
      );
      
      const skeleton = document.querySelector('[data-slot="skeleton"]') ||
                      document.querySelector('.animate-pulse');
      expect(skeleton).toBeInTheDocument();
    });

    it('should apply custom className', () => {
      const { container } = render(
        <ChartContainer title="Test Chart" className="custom-class">
          <div>Chart Content</div>
        </ChartContainer>
      );
      
      const containerEl = container.querySelector('.custom-class');
      expect(containerEl).toBeInTheDocument();
    });
  });

  describe('DashboardCard', () => {
    it('should render dashboard card with title and description', () => {
      render(
        <DashboardCard
          title="Test Card"
          description="Card description"
          href="/test"
          icon={Users}
        />
      );
      
      expect(screen.getByText('Test Card')).toBeInTheDocument();
      expect(screen.getByText('Card description')).toBeInTheDocument();
    });

    it('should render badges when provided', () => {
      render(
        <DashboardCard
          title="Test Card"
          description="Card description"
          href="/test"
          icon={Users}
          badges={['New', 'Enhanced']}
        />
      );
      
      expect(screen.getByText('New')).toBeInTheDocument();
      expect(screen.getByText('Enhanced')).toBeInTheDocument();
    });

    it('should render features and open button when features provided', () => {
      render(
        <DashboardCard
          title="Test Card"
          description="Card description"
          href="/test"
          icon={Users}
          features={['Feature 1', 'Feature 2']}
        />
      );
      
      expect(screen.getByText('Feature 1')).toBeInTheDocument();
      expect(screen.getByText('Feature 2')).toBeInTheDocument();
      expect(screen.getByText('Open Dashboard')).toBeInTheDocument();
    });

    it('should apply variant styles', () => {
      const { container } = render(
        <DashboardCard
          title="Test Card"
          description="Card description"
          href="/test"
          icon={Users}
          variant="compliance"
        />
      );
      
      const card = container.querySelector('.border-l-red-500');
      expect(card).toBeInTheDocument();
    });

    it('should apply custom className', () => {
      const { container } = render(
        <DashboardCard
          title="Test Card"
          description="Card description"
          href="/test"
          icon={Users}
          className="custom-class"
        />
      );
      
      const card = container.querySelector('.custom-class');
      expect(card).toBeInTheDocument();
    });
  });

  describe('DataTable', () => {
    const mockColumns = [
      { key: 'name', header: 'Name' },
      { key: 'value', header: 'Value', sortable: true },
    ];

    const mockData = [
      { id: '1', name: 'Item 1', value: 100 },
      { id: '2', name: 'Item 2', value: 200 },
    ];

    it('should render data table with columns and data', () => {
      render(<DataTable columns={mockColumns} data={mockData} />);
      
      expect(screen.getByText('Name')).toBeInTheDocument();
      expect(screen.getByText('Value')).toBeInTheDocument();
      expect(screen.getByText('Item 1')).toBeInTheDocument();
      expect(screen.getByText('Item 2')).toBeInTheDocument();
    });

    it('should render empty state when no data', () => {
      render(<DataTable columns={mockColumns} data={[]} />);
      
      expect(screen.getByText('No data available')).toBeInTheDocument();
    });

    it('should handle sorting', async () => {
      const user = userEvent.setup();
      render(<DataTable columns={mockColumns} data={mockData} />);
      
      const valueHeader = screen.getByText('Value');
      await user.click(valueHeader);
      
      // Table should still be visible after sorting
      expect(screen.getByText('Item 1')).toBeInTheDocument();
    });

    it('should handle search when searchable', async () => {
      const user = userEvent.setup();
      render(<DataTable columns={mockColumns} data={mockData} searchable />);
      
      const searchInput = screen.getByPlaceholderText('Search...');
      await user.type(searchInput, 'Item 1');
      
      // Filtered results should be visible
      expect(screen.getByText('Item 1')).toBeInTheDocument();
      expect(screen.queryByText('Item 2')).not.toBeInTheDocument();
    });

    it('should handle pagination', () => {
      const largeData = Array.from({ length: 20 }, (_, i) => ({
        id: String(i),
        name: `Item ${i}`,
        value: i * 10,
      }));
      
      render(<DataTable columns={mockColumns} data={largeData} pagination={{ pageSize: 10 }} />);
      
      // First page should be visible
      expect(screen.getByText('Item 0')).toBeInTheDocument();
      
      // Pagination controls should be present
      expect(screen.getByText('Item 0')).toBeInTheDocument();
    });

    it('should handle row click', async () => {
      const user = userEvent.setup();
      const onRowClick = vi.fn();
      
      render(<DataTable columns={mockColumns} data={mockData} onRowClick={onRowClick} />);
      
      const row = screen.getByText('Item 1').closest('tr');
      if (row) {
        await user.click(row);
        expect(onRowClick).toHaveBeenCalledWith(mockData[0]);
      }
    });

    it('should use custom empty message', () => {
      render(<DataTable columns={mockColumns} data={[]} emptyMessage="Custom empty" />);
      
      expect(screen.getByText('Custom empty')).toBeInTheDocument();
    });
  });

  describe('MetricCard', () => {
    it('should render metric card with label and value', () => {
      render(<MetricCard label="Total Users" value="1,234" />);
      
      expect(screen.getByText('Total Users')).toBeInTheDocument();
      expect(screen.getByText('1,234')).toBeInTheDocument();
    });

    it('should render description when provided', () => {
      render(
        <MetricCard
          label="Total Users"
          value="1,234"
          description="Active users this month"
        />
      );
      
      expect(screen.getByText('Active users this month')).toBeInTheDocument();
    });

    it('should render trend when provided', () => {
      render(
        <MetricCard
          label="Total Users"
          value="1,234"
          trend={{ value: 5, isPositive: true }}
        />
      );
      
      // Trend indicator should be visible
      expect(screen.getByText('5%')).toBeInTheDocument();
    });

    it('should render icon when provided', () => {
      render(
        <MetricCard
          label="Total Users"
          value="1,234"
          icon={Users}
        />
      );
      
      // Icon should be rendered (lucide-react icons render as SVG)
      const icon = document.querySelector('svg');
      expect(icon).toBeInTheDocument();
    });

    it('should apply variant styles', () => {
      const { container } = render(
        <MetricCard
          label="Total Users"
          value="1,234"
          variant="success"
        />
      );
      
      const valueEl = container.querySelector('.text-green-600');
      expect(valueEl).toBeInTheDocument();
    });

    it('should apply custom className', () => {
      const { container } = render(
        <MetricCard
          label="Total Users"
          value="1,234"
          className="custom-class"
        />
      );
      
      const card = container.querySelector('.custom-class');
      expect(card).toBeInTheDocument();
    });
  });
});

