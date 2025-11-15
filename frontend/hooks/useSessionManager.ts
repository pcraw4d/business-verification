'use client';

import { useState, useEffect, useCallback } from 'react';

interface SessionData {
  merchantId: string;
  merchantName?: string;
  timestamp: number;
}

const SESSION_STORAGE_KEY = 'merchant_session';
const SESSION_TIMEOUT = 30 * 60 * 1000; // 30 minutes

export function useSessionManager() {
  const [currentSession, setCurrentSession] = useState<SessionData | null>(null);
  const [sessionHistory, setSessionHistory] = useState<SessionData[]>([]);

  // Load session from storage
  useEffect(() => {
    if (typeof window === 'undefined') return;

    try {
      const stored = sessionStorage.getItem(SESSION_STORAGE_KEY);
      if (stored) {
        const session: SessionData = JSON.parse(stored);
        // Check if session is still valid (not expired)
        if (Date.now() - session.timestamp < SESSION_TIMEOUT) {
          setCurrentSession(session);
        } else {
          sessionStorage.removeItem(SESSION_STORAGE_KEY);
        }
      }

      // Load history
      const historyStored = sessionStorage.getItem('merchant_session_history');
      if (historyStored) {
        setSessionHistory(JSON.parse(historyStored));
      }
    } catch (error) {
      console.error('Error loading session:', error);
    }
  }, []);

  // Start a new session
  const startSession = useCallback((merchantId: string, merchantName?: string) => {
    const session: SessionData = {
      merchantId,
      merchantName,
      timestamp: Date.now(),
    };

    setCurrentSession(session);
    sessionStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));

    // Add to history
    setSessionHistory((prev) => {
      const updated = [session, ...prev.filter((s) => s.merchantId !== merchantId)].slice(0, 10);
      sessionStorage.setItem('merchant_session_history', JSON.stringify(updated));
      return updated;
    });
  }, []);

  // End current session
  const endSession = useCallback(() => {
    setCurrentSession(null);
    sessionStorage.removeItem(SESSION_STORAGE_KEY);
  }, []);

  // Switch to a different merchant
  const switchSession = useCallback(
    (merchantId: string, merchantName?: string) => {
      endSession();
      startSession(merchantId, merchantName);
    },
    [endSession, startSession]
  );

  return {
    currentSession,
    sessionHistory,
    startSession,
    endSession,
    switchSession,
    isSessionActive: currentSession !== null,
  };
}

