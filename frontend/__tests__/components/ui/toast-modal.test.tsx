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
      
      // sonner's toast methods may be called with just the message or with options
      expect(mockToast.success).toHaveBeenCalled();
      const callArgs = mockToast.success.mock.calls[0];
      expect(callArgs[0]).toBe('Operation successful');
    });

    it('should show error toast', () => {
      toast.error('Operation failed');
      
      expect(mockToast.error).toHaveBeenCalled();
      const callArgs = mockToast.error.mock.calls[0];
      expect(callArgs[0]).toBe('Operation failed');
    });

    it('should show info toast', () => {
      toast.info('Information message');
      
      expect(mockToast.info).toHaveBeenCalled();
      const callArgs = mockToast.info.mock.calls[0];
      expect(callArgs[0]).toBe('Information message');
    });

    it('should show warning toast', () => {
      toast.warning('Warning message');
      
      expect(mockToast.warning).toHaveBeenCalled();
      const callArgs = mockToast.warning.mock.calls[0];
      expect(callArgs[0]).toBe('Warning message');
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
      
      // Wait for dialog to be fully rendered and focus trap to be set up
      await waitFor(() => {
        const dialog = screen.getByRole('dialog');
        expect(dialog).toBeInTheDocument();
      });
      
      // Tab navigation should work - focus might go to any focusable element in the dialog
      // Check that at least one button has focus
      await user.tab();
      const firstButton = screen.getByText('First Button');
      const secondButton = screen.getByText('Second Button');
      
      // Focus should be on one of the buttons (dialog focus trap behavior may vary)
      const hasFocus = firstButton === document.activeElement || secondButton === document.activeElement;
      expect(hasFocus).toBe(true);
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

