'use client';

import { useEffect, useCallback } from 'react';

interface KeyboardShortcut {
  key: string;
  ctrlKey?: boolean;
  metaKey?: boolean;
  shiftKey?: boolean;
  altKey?: boolean;
  handler: () => void;
  description: string;
}

/**
 * Hook for managing keyboard shortcuts
 * Supports R for refresh, E for enrichment, and other common shortcuts
 */
export function useKeyboardShortcuts(shortcuts: KeyboardShortcut[]) {
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Don't trigger shortcuts when typing in inputs, textareas, or contenteditable elements
      const target = event.target as HTMLElement;
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable ||
        (target.tagName === 'BUTTON' && target.getAttribute('aria-label')?.includes('Search'))
      ) {
        return;
      }

      for (const shortcut of shortcuts) {
        const keyMatches = event.key.toLowerCase() === shortcut.key.toLowerCase();
        const ctrlMatches = shortcut.ctrlKey === undefined ? true : event.ctrlKey === shortcut.ctrlKey;
        const metaMatches = shortcut.metaKey === undefined ? true : event.metaKey === shortcut.metaKey;
        const shiftMatches = shortcut.shiftKey === undefined ? true : event.shiftKey === shortcut.shiftKey;
        const altMatches = shortcut.altKey === undefined ? true : event.altKey === shortcut.altKey;

        if (keyMatches && ctrlMatches && metaMatches && shiftMatches && altMatches) {
          event.preventDefault();
          shortcut.handler();
          break;
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [shortcuts]);
}

/**
 * Common keyboard shortcuts for merchant details page
 */
export const MERCHANT_DETAILS_SHORTCUTS = {
  refresh: {
    key: 'r',
    description: 'Refresh data',
  },
  enrichment: {
    key: 'e',
    description: 'Open enrichment dialog',
  },
} as const;

