'use client';

import { useState, useEffect, createContext, useContext, ReactNode } from 'react';
import { getMerchant } from '@/lib/api';
import type { Merchant } from '@/types/merchant';

interface MerchantContextType {
  merchant: Merchant | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}

const MerchantContext = createContext<MerchantContextType | undefined>(undefined) as React.Context<MerchantContextType | undefined>;

export function MerchantProvider({
  children,
  merchantId,
}: {
  children: ReactNode;
  merchantId: string;
}) {
  const [merchant, setMerchant] = useState<Merchant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getMerchant(merchantId);
      setMerchant(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load merchant');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (merchantId) {
      refresh();
    }
  }, [merchantId]);

  return (
    <MerchantContext.Provider value={{ merchant, loading, error, refresh }}>
      {children}
    </MerchantContext.Provider>
  );
}

export function useMerchantContext() {
  const context = useContext(MerchantContext);
  if (context === undefined) {
    throw new Error('useMerchantContext must be used within a MerchantProvider');
  }
  return context;
}

