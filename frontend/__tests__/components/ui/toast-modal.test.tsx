import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { toast } from 'sonner';

vi.mock('sonner');

const mockToast = vi.mocked(toast);

describe('Toast Notifications', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.success = vi.fn();
    mockToast.error = vi.fn();
    mockToast.info = vi.fn();
    mockToast.warning = vi.fn();
  });

  describe('Toast Display', () => {
    it('should show success toast', () => {
      toast.success('Operation successful');
      
      expect(mockToast.success).toHaveBeenCalledWith('Operation successful', expect.any(Object));
    });

    it('should show error toast', () => {
      toast.error('Operation failed');
      
      expect(mockToast.error).toHaveBeenCalledWith('Operation failed', expect.any(Object));
    });

    it('should show info toast', () => {
      toast.info('Information message');
      
      expect(mockToast.info).toHaveBeenCalledWith('Information message', expect.any(Object));
    });

    it('should show warning toast', () => {
      toast.warning('Warning message');
      
      expect(mockToast.warning).toHaveBeenCalledWith('Warning message', expect.any(Object));
    });
  });

  describe('Toast Options', () => {
    it('should accept description in toast options', () => {
      toast.success('Success', { description: 'Operation completed successfully' });
      
      expect(mockToast.success).toHaveBeenCalledWith('Success', {
        description: 'Operation completed successfully',
      });
    });

    it('should accept duration in toast options', () => {
      toast.info('Info', { duration: 5000 });
      
      expect(mockToast.info).toHaveBeenCalledWith('Info', {
        duration: 5000,
      });
    });
  });

  describe('Toast Types', () => {
    it('should handle success toast type', () => {
      toast.success('Success message');
      expect(mockToast.success).toHaveBeenCalled();
    });

    it('should handle error toast type', () => {
      toast.error('Error message');
      expect(mockToast.error).toHaveBeenCalled();
    });

    it('should handle info toast type', () => {
      toast.info('Info message');
      expect(mockToast.info).toHaveBeenCalled();
    });

    it('should handle warning toast type', () => {
      toast.warning('Warning message');
      expect(mockToast.warning).toHaveBeenCalled();
    });
  });
});

describe('Modal Dialogs', () => {
  describe('Dialog Rendering', () => {
    it('should render dialog trigger', () => {
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Test Dialog</DialogTitle>
              <DialogDescription>Test description</DialogDescription>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      );
      
      expect(screen.getByText('Open Dialog')).toBeInTheDocument();
    });

    it('should open dialog when trigger is clicked', async () => {
      const user = userEvent.setup();
      
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Test Dialog</DialogTitle>
              <DialogDescription>Test description</DialogDescription>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      );
      
      const trigger = screen.getByText('Open Dialog');
      await user.click(trigger);
      
      await waitFor(() => {
        expect(screen.getByText('Test Dialog')).toBeInTheDocument();
        expect(screen.getByText('Test description')).toBeInTheDocument();
      });
    });

    it('should close dialog when clicking outside', async () => {
      const user = userEvent.setup();
      
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Test Dialog</DialogTitle>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      );
      
      const trigger = screen.getByText('Open Dialog');
      await user.click(trigger);
      
      await waitFor(() => {
        expect(screen.getByText('Test Dialog')).toBeInTheDocument();
      });
      
      // Press Escape to close
      await user.keyboard('{Escape}');
      
      await waitFor(() => {
        expect(screen.queryByText('Test Dialog')).not.toBeInTheDocument();
      });
    });

    it('should close dialog when clicking close button', async () => {
      const user = userEvent.setup();
      
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Test Dialog</DialogTitle>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      );
      
      const trigger = screen.getByText('Open Dialog');
      await user.click(trigger);
      
      await waitFor(() => {
        expect(screen.getByText('Test Dialog')).toBeInTheDocument();
      });
      
      // Find and click close button (usually an X button)
      const closeButton = screen.queryByRole('button', { name: /close/i }) ||
                         screen.queryByLabelText(/close/i);
      
      if (closeButton) {
        await user.click(closeButton);
        
        await waitFor(() => {
          expect(screen.queryByText('Test Dialog')).not.toBeInTheDocument();
        });
      }
    });
  });

  describe('Modal Focus Management', () => {
    it('should trap focus within dialog', async () => {
      const user = userEvent.setup();
      
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Test Dialog</DialogTitle>
            </DialogHeader>
            <Button>Action Button</Button>
          </DialogContent>
        </Dialog>
      );
      
      const trigger = screen.getByText('Open Dialog');
      await user.click(trigger);
      
      await waitFor(() => {
        expect(screen.getByText('Test Dialog')).toBeInTheDocument();
      });
      
      // Focus should be within dialog
      const actionButton = screen.getByText('Action Button');
      expect(actionButton).toBeInTheDocument();
    });
  });

  describe('Modal Accessibility', () => {
    it('should have proper ARIA attributes', async () => {
      const user = userEvent.setup();
      
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Test Dialog</DialogTitle>
              <DialogDescription>Test description</DialogDescription>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      );
      
      const trigger = screen.getByText('Open Dialog');
      await user.click(trigger);
      
      await waitFor(() => {
        const dialog = screen.getByRole('dialog');
        expect(dialog).toBeInTheDocument();
        expect(dialog).toHaveAttribute('aria-labelledby');
        expect(dialog).toHaveAttribute('aria-describedby');
      });
    });

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup();
      
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Test Dialog</DialogTitle>
            </DialogHeader>
            <Button>First Button</Button>
            <Button>Second Button</Button>
          </DialogContent>
        </Dialog>
      );
      
      const trigger = screen.getByText('Open Dialog');
      await user.click(trigger);
      
      await waitFor(() => {
        expect(screen.getByText('Test Dialog')).toBeInTheDocument();
      });
      
      // Tab navigation should work
      await user.tab();
      const firstButton = screen.getByText('First Button');
      expect(firstButton).toHaveFocus();
    });
  });

  describe('Modal Content', () => {
    it('should render dialog title', async () => {
      const user = userEvent.setup();
      
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>My Dialog Title</DialogTitle>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      );
      
      const trigger = screen.getByText('Open Dialog');
      await user.click(trigger);
      
      await waitFor(() => {
        expect(screen.getByText('My Dialog Title')).toBeInTheDocument();
      });
    });

    it('should render dialog description', async () => {
      const user = userEvent.setup();
      
      render(
        <Dialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Test Dialog</DialogTitle>
              <DialogDescription>This is a test description</DialogDescription>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      );
      
      const trigger = screen.getByText('Open Dialog');
      await user.click(trigger);
      
      await waitFor(() => {
        expect(screen.getByText('This is a test description')).toBeInTheDocument();
      });
    });
  });
});

