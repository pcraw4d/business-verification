import { renderHook } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';

describe('useKeyboardShortcuts', () => {
  let mockHandler: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockHandler = vi.fn();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Basic functionality', () => {
    it('should register keyboard shortcut handler', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      // Simulate keydown event
      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);
    });

    it('should call handler when matching key is pressed', () => {
      const shortcuts = [
        {
          key: 'e',
          handler: mockHandler,
          description: 'Enrichment',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 'e',
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);
    });

    it('should not call handler when non-matching key is pressed', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 'x',
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();
    });

    it('should handle case-insensitive key matching', () => {
      const shortcuts = [
        {
          key: 'R',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      // Press lowercase 'r'
      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);
    });
  });

  describe('Modifier keys', () => {
    it('should match when ctrlKey is required and pressed', () => {
      const shortcuts = [
        {
          key: 's',
          ctrlKey: true,
          handler: mockHandler,
          description: 'Save',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 's',
        ctrlKey: true,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);
    });

    it('should not match when ctrlKey is required but not pressed', () => {
      const shortcuts = [
        {
          key: 's',
          ctrlKey: true,
          handler: mockHandler,
          description: 'Save',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 's',
        ctrlKey: false,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();
    });

    it('should match when ctrlKey is undefined (any state allowed)', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      // With ctrlKey
      const event1 = new KeyboardEvent('keydown', {
        key: 'r',
        ctrlKey: true,
        bubbles: true,
      });
      window.dispatchEvent(event1);

      // Without ctrlKey
      const event2 = new KeyboardEvent('keydown', {
        key: 'r',
        ctrlKey: false,
        bubbles: true,
      });
      window.dispatchEvent(event2);

      expect(mockHandler).toHaveBeenCalledTimes(2);
    });

    it('should match when metaKey is required and pressed', () => {
      const shortcuts = [
        {
          key: 's',
          metaKey: true,
          handler: mockHandler,
          description: 'Save',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 's',
        metaKey: true,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);
    });

    it('should not match when metaKey is required but not pressed', () => {
      const shortcuts = [
        {
          key: 's',
          metaKey: true,
          handler: mockHandler,
          description: 'Save',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 's',
        metaKey: false,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();
    });

    it('should match when shiftKey is required and pressed', () => {
      const shortcuts = [
        {
          key: 'r',
          shiftKey: true,
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        shiftKey: true,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);
    });

    it('should not match when shiftKey is required but not pressed', () => {
      const shortcuts = [
        {
          key: 'r',
          shiftKey: true,
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        shiftKey: false,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();
    });

    it('should match when altKey is required and pressed', () => {
      const shortcuts = [
        {
          key: 'r',
          altKey: true,
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        altKey: true,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);
    });

    it('should not match when altKey is required but not pressed', () => {
      const shortcuts = [
        {
          key: 'r',
          altKey: true,
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        altKey: false,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();
    });

    it('should match when all modifier keys are required and pressed', () => {
      const shortcuts = [
        {
          key: 's',
          ctrlKey: true,
          shiftKey: true,
          altKey: true,
          handler: mockHandler,
          description: 'Save All',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 's',
        ctrlKey: true,
        shiftKey: true,
        altKey: true,
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);
    });
  });

  describe('Input filtering', () => {
    it('should not trigger shortcut when typing in INPUT element', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const input = document.createElement('input');
      document.body.appendChild(input);
      input.focus();

      // Create event and dispatch it on the input element so target is set correctly
      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
        cancelable: true,
      });
      Object.defineProperty(event, 'target', { value: input, writable: false });
      input.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();

      document.body.removeChild(input);
    });

    it('should not trigger shortcut when typing in TEXTAREA element', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const textarea = document.createElement('textarea');
      document.body.appendChild(textarea);
      textarea.focus();

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
        cancelable: true,
      });
      Object.defineProperty(event, 'target', { value: textarea, writable: false });
      textarea.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();

      document.body.removeChild(textarea);
    });

    it('should not trigger shortcut when typing in contenteditable element', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const div = document.createElement('div');
      div.contentEditable = 'true';
      document.body.appendChild(div);
      div.focus();

      // Dispatch event on window with target set to the contenteditable div
      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
        cancelable: true,
      });
      Object.defineProperty(event, 'target', { 
        value: div, 
        writable: false,
        configurable: true,
      });
      // Ensure isContentEditable is true
      Object.defineProperty(div, 'isContentEditable', {
        value: true,
        writable: false,
        configurable: true,
      });
      window.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();

      document.body.removeChild(div);
    });

    it('should not trigger shortcut when typing in search button', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const button = document.createElement('button');
      button.setAttribute('aria-label', 'Search merchants');
      document.body.appendChild(button);
      button.focus();

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
        cancelable: true,
      });
      Object.defineProperty(event, 'target', { value: button, writable: false });
      button.dispatchEvent(event);

      expect(mockHandler).not.toHaveBeenCalled();

      document.body.removeChild(button);
    });

    it('should trigger shortcut when typing in regular button (not search)', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const button = document.createElement('button');
      button.setAttribute('aria-label', 'Submit form');
      document.body.appendChild(button);
      button.focus();

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
        cancelable: true,
      });
      Object.defineProperty(event, 'target', { value: button, writable: false });
      button.dispatchEvent(event);

      expect(mockHandler).toHaveBeenCalledTimes(1);

      document.body.removeChild(button);
    });
  });

  describe('Multiple shortcuts', () => {
    it('should handle multiple shortcuts', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      const handler3 = vi.fn();

      const shortcuts = [
        {
          key: 'r',
          handler: handler1,
          description: 'Refresh',
        },
        {
          key: 'e',
          handler: handler2,
          description: 'Enrichment',
        },
        {
          key: 'x',
          handler: handler3,
          description: 'Export',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      // Press 'r'
      const event1 = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
      });
      window.dispatchEvent(event1);

      // Press 'e'
      const event2 = new KeyboardEvent('keydown', {
        key: 'e',
        bubbles: true,
      });
      window.dispatchEvent(event2);

      // Press 'x'
      const event3 = new KeyboardEvent('keydown', {
        key: 'x',
        bubbles: true,
      });
      window.dispatchEvent(event3);

      expect(handler1).toHaveBeenCalledTimes(1);
      expect(handler2).toHaveBeenCalledTimes(1);
      expect(handler3).toHaveBeenCalledTimes(1);
    });

    it('should only call first matching shortcut handler', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();

      const shortcuts = [
        {
          key: 'r',
          handler: handler1,
          description: 'Refresh 1',
        },
        {
          key: 'r',
          handler: handler2,
          description: 'Refresh 2',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
      });
      window.dispatchEvent(event);

      expect(handler1).toHaveBeenCalledTimes(1);
      expect(handler2).not.toHaveBeenCalled();
    });
  });

  describe('Event prevention', () => {
    it('should prevent default behavior when shortcut is triggered', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      renderHook(() => useKeyboardShortcuts(shortcuts));

      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
        cancelable: true,
      });
      const preventDefaultSpy = vi.spyOn(event, 'preventDefault');
      window.dispatchEvent(event);

      expect(preventDefaultSpy).toHaveBeenCalled();
      expect(mockHandler).toHaveBeenCalledTimes(1);
    });
  });

  describe('Cleanup', () => {
    it('should remove event listener on unmount', () => {
      const shortcuts = [
        {
          key: 'r',
          handler: mockHandler,
          description: 'Refresh',
        },
      ];

      const { unmount } = renderHook(() => useKeyboardShortcuts(shortcuts));

      // Unmount the hook
      unmount();

      // Try to trigger the shortcut after unmount
      const event = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
      });
      window.dispatchEvent(event);

      // Handler should not be called after unmount
      expect(mockHandler).not.toHaveBeenCalled();
    });

    it('should update shortcuts when dependencies change', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();

      const shortcuts1 = [
        {
          key: 'r',
          handler: handler1,
          description: 'Refresh',
        },
      ];

      const { rerender } = renderHook(
        ({ shortcuts }) => useKeyboardShortcuts(shortcuts),
        {
          initialProps: { shortcuts: shortcuts1 },
        }
      );

      // Trigger with first shortcut
      const event1 = new KeyboardEvent('keydown', {
        key: 'r',
        bubbles: true,
        cancelable: true,
      });
      window.dispatchEvent(event1);

      expect(handler1).toHaveBeenCalledTimes(1);
      handler1.mockClear();

      // Update shortcuts
      const shortcuts2 = [
        {
          key: 'e',
          handler: handler2,
          description: 'Enrichment',
        },
      ];

      rerender({ shortcuts: shortcuts2 });

      // Trigger with new shortcut
      const event2 = new KeyboardEvent('keydown', {
        key: 'e',
        bubbles: true,
        cancelable: true,
      });
      window.dispatchEvent(event2);

      expect(handler2).toHaveBeenCalledTimes(1);
      expect(handler1).not.toHaveBeenCalled();
    });
  });
});

