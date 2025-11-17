import { server } from '@/__tests__/mocks/server';
import { ExportButton } from '@/components/common/ExportButton';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { toast } from 'sonner';

vi.mock('sonner');

const mockToast = vi.mocked(toast);

describe('ExportButton (common)', () => {
  const merchantId = 'merchant-123';
  const mockData = { id: merchantId, name: 'Test Business' };

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.success = vi.fn();
    mockToast.error = vi.fn();
    
    // Mock sessionStorage
    Object.defineProperty(window, 'sessionStorage', {
      value: {
        getItem: vi.fn(() => null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      writable: true,
    });

    // Mock URL.createObjectURL and URL.revokeObjectURL
    global.URL.createObjectURL = vi.fn(() => 'blob:mock-url');
    global.URL.revokeObjectURL = vi.fn();
  });

  describe('Component Rendering', () => {
    it('should render export button', () => {
      render(<ExportButton merchantId={merchantId} />);
      
      const button = screen.getByRole('button', { name: /export csv/i });
      expect(button).toBeInTheDocument();
    });

    it('should display correct format in button text', () => {
      render(<ExportButton merchantId={merchantId} format="pdf" />);
      
      const button = screen.getByRole('button', { name: /export pdf/i });
      expect(button).toBeInTheDocument();
    });

    it('should show exporting state when exporting', async () => {
      let resolveExport: (value: any) => void;
      const exportPromise = new Promise((resolve) => {
        resolveExport = resolve;
      });

      server.use(
        http.get('*/api/v1/merchants/:merchantId/export/:format', async () => {
          await exportPromise;
          return HttpResponse.json({});
        })
      );

      const user = userEvent.setup();
      render(<ExportButton merchantId={merchantId} format="csv" />);
      
      const button = screen.getByRole('button', { name: /export csv/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(button).toBeDisabled();
        expect(screen.getByText(/exporting/i)).toBeInTheDocument();
      });
      
      resolveExport!({});
    });
  });

  describe('JSON Export (Client-side)', () => {
    it('should export JSON data client-side', async () => {
      const user = userEvent.setup();
      render(<ExportButton merchantId={merchantId} format="json" data={mockData} />);
      
      const button = screen.getByRole('button', { name: /export json/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(mockToast.success).toHaveBeenCalledWith('Data exported successfully');
        expect(global.URL.createObjectURL).toHaveBeenCalled();
      });
    });

    it('should use custom filename for JSON export', async () => {
      const user = userEvent.setup();
      const filename = 'custom-export.json';
      render(<ExportButton merchantId={merchantId} format="json" data={mockData} filename={filename} />);
      
      const button = screen.getByRole('button', { name: /export json/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(mockToast.success).toHaveBeenCalled();
      });
    });
  });

  describe('Server-side Export', () => {
    it('should export CSV via API', async () => {
      const mockBlob = new Blob(['csv,data'], { type: 'text/csv' });
      
      server.use(
        http.get('*/api/v1/merchants/:merchantId/export/:format', () => {
          return HttpResponse.arrayBuffer(mockBlob.arrayBuffer());
        })
      );

      const user = userEvent.setup();
      render(<ExportButton merchantId={merchantId} format="csv" />);
      
      const button = screen.getByRole('button', { name: /export csv/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(mockToast.success).toHaveBeenCalledWith('Data exported successfully');
      });
    });

    it('should export PDF via API', async () => {
      const mockBlob = new Blob(['pdf content'], { type: 'application/pdf' });
      
      server.use(
        http.get('*/api/v1/merchants/:merchantId/export/:format', () => {
          return HttpResponse.arrayBuffer(mockBlob.arrayBuffer());
        })
      );

      const user = userEvent.setup();
      render(<ExportButton merchantId={merchantId} format="pdf" />);
      
      const button = screen.getByRole('button', { name: /export pdf/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(mockToast.success).toHaveBeenCalledWith('Data exported successfully');
      });
    });

    it('should include auth token in headers when available', async () => {
      const getItemSpy = vi.fn(() => 'test-token');
      Object.defineProperty(window, 'sessionStorage', {
        value: {
          getItem: getItemSpy,
          setItem: vi.fn(),
          removeItem: vi.fn(),
          clear: vi.fn(),
        },
        writable: true,
      });

      const mockBlob = new Blob(['data'], { type: 'text/csv' });
      
      server.use(
        http.get('*/api/v1/merchants/:merchantId/export/:format', ({ request }) => {
          const authHeader = request.headers.get('Authorization');
          expect(authHeader).toBe('Bearer test-token');
          return HttpResponse.arrayBuffer(mockBlob.arrayBuffer());
        })
      );

      const user = userEvent.setup();
      render(<ExportButton merchantId={merchantId} format="csv" />);
      
      const button = screen.getByRole('button', { name: /export csv/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(getItemSpy).toHaveBeenCalledWith('authToken');
      });
    });

    it('should handle export errors', async () => {
      server.use(
        http.get('*/api/v1/merchants/:merchantId/export/:format', () => {
          return HttpResponse.json({ error: 'Export failed' }, { status: 500 });
        })
      );

      const user = userEvent.setup();
      render(<ExportButton merchantId={merchantId} format="csv" />);
      
      const button = screen.getByRole('button', { name: /export csv/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to export data');
      });
    });

    it('should handle network errors', async () => {
      server.use(
        http.get('*/api/v1/merchants/:merchantId/export/:format', () => {
          return HttpResponse.error();
        })
      );

      const user = userEvent.setup();
      render(<ExportButton merchantId={merchantId} format="csv" />);
      
      const button = screen.getByRole('button', { name: /export csv/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith('Failed to export data');
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper aria-label', () => {
      render(<ExportButton merchantId={merchantId} format="csv" />);
      
      const button = screen.getByRole('button', { name: /export data as csv/i });
      expect(button).toBeInTheDocument();
    });

    it('should update aria-label when exporting', async () => {
      let resolveExport: (value: any) => void;
      const exportPromise = new Promise((resolve) => {
        resolveExport = resolve;
      });

      server.use(
        http.get('*/api/v1/merchants/:merchantId/export/:format', async () => {
          await exportPromise;
          return HttpResponse.json({});
        })
      );

      const user = userEvent.setup();
      render(<ExportButton merchantId={merchantId} format="csv" />);
      
      const button = screen.getByRole('button', { name: /export data as csv/i });
      await user.click(button);
      
      await waitFor(() => {
        expect(button).toHaveAttribute('aria-label', 'Exporting CSV');
      });
      
      resolveExport!({});
    });
  });
});

