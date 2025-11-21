'use client';

import { createContext, useContext, useState, useCallback, ReactNode, useEffect } from 'react';

interface EnrichedField {
  fieldName: string;
  enrichedAt: Date;
  source: string;
  type: 'added' | 'updated';
}

interface EnrichmentContextType {
  enrichedFields: Map<string, EnrichedField[]>;
  addEnrichedFields: (merchantId: string, fields: { name: string; type: 'added' | 'updated'; source: string }[]) => void;
  clearEnrichedFields: (merchantId: string) => void;
  isFieldEnriched: (merchantId: string, fieldName: string) => boolean;
  getEnrichedFieldInfo: (merchantId: string, fieldName: string) => EnrichedField | null;
}

const EnrichmentContext = createContext<EnrichmentContextType | null>(null);

const STORAGE_KEY = 'enriched_fields';
const HIGHLIGHT_DURATION = 5 * 60 * 1000; // 5 minutes

export function EnrichmentProvider({ children }: { children: ReactNode }) {
  const [enrichedFields, setEnrichedFields] = useState<Map<string, EnrichedField[]>>(new Map());

  // Load from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        const map = new Map<string, EnrichedField[]>();
        
        Object.entries(parsed).forEach(([merchantId, fields]: [string, any]) => {
          map.set(
            merchantId,
            fields.map((f: any) => ({
              ...f,
              enrichedAt: new Date(f.enrichedAt),
            }))
          );
        });
        
        setEnrichedFields(map);
      }
    } catch (error) {
      console.error('Failed to load enriched fields from localStorage:', error);
    }
  }, []);

  // Save to localStorage whenever enrichedFields changes
  useEffect(() => {
    try {
      const serialized: Record<string, any> = {};
      enrichedFields.forEach((fields, merchantId) => {
        serialized[merchantId] = fields.map((f) => ({
          ...f,
          enrichedAt: f.enrichedAt.toISOString(),
        }));
      });
      localStorage.setItem(STORAGE_KEY, JSON.stringify(serialized));
    } catch (error) {
      console.error('Failed to save enriched fields to localStorage:', error);
    }
  }, [enrichedFields]);

  // Clean up expired highlights
  useEffect(() => {
    const interval = setInterval(() => {
      setEnrichedFields((prev) => {
        const now = new Date();
        const updated = new Map(prev);
        
        prev.forEach((fields, merchantId) => {
          const validFields = fields.filter((f) => {
            const age = now.getTime() - f.enrichedAt.getTime();
            return age < HIGHLIGHT_DURATION;
          });
          
          if (validFields.length === 0) {
            updated.delete(merchantId);
          } else {
            updated.set(merchantId, validFields);
          }
        });
        
        return updated;
      });
    }, 60000); // Check every minute

    return () => clearInterval(interval);
  }, []);

  const addEnrichedFields = useCallback(
    (merchantId: string, fields: { name: string; type: 'added' | 'updated'; source: string }[]) => {
      setEnrichedFields((prev) => {
        const updated = new Map(prev);
        const existing = updated.get(merchantId) || [];
        
        const newFields: EnrichedField[] = fields.map((f) => ({
          fieldName: f.name,
          enrichedAt: new Date(),
          source: f.source,
          type: f.type,
        }));
        
        // Merge with existing, avoiding duplicates
        const merged = [...existing];
        newFields.forEach((newField) => {
          const existingIndex = merged.findIndex(
            (f) => f.fieldName === newField.fieldName && f.type === newField.type
          );
          if (existingIndex >= 0) {
            merged[existingIndex] = newField; // Update timestamp
          } else {
            merged.push(newField);
          }
        });
        
        updated.set(merchantId, merged);
        return updated;
      });
    },
    []
  );

  const clearEnrichedFields = useCallback((merchantId: string) => {
    setEnrichedFields((prev) => {
      const updated = new Map(prev);
      updated.delete(merchantId);
      return updated;
    });
  }, []);

  const isFieldEnriched = useCallback(
    (merchantId: string, fieldName: string): boolean => {
      const fields = enrichedFields.get(merchantId) || [];
      const now = new Date();
      
      return fields.some((f) => {
        const age = now.getTime() - f.enrichedAt.getTime();
        return f.fieldName === fieldName && age < HIGHLIGHT_DURATION;
      });
    },
    [enrichedFields]
  );

  const getEnrichedFieldInfo = useCallback(
    (merchantId: string, fieldName: string): EnrichedField | null => {
      const fields = enrichedFields.get(merchantId) || [];
      const now = new Date();
      
      const field = fields.find((f) => {
        const age = now.getTime() - f.enrichedAt.getTime();
        return f.fieldName === fieldName && age < HIGHLIGHT_DURATION;
      });
      
      return field || null;
    },
    [enrichedFields]
  );

  return (
    <EnrichmentContext.Provider
      value={{
        enrichedFields,
        addEnrichedFields,
        clearEnrichedFields,
        isFieldEnriched,
        getEnrichedFieldInfo,
      }}
    >
      {children}
    </EnrichmentContext.Provider>
  );
}

export function useEnrichment() {
  const context = useContext(EnrichmentContext);
  if (!context) {
    // Return a no-op implementation if context is not available
    return {
      enrichedFields: new Map<string, EnrichedField[]>(),
      addEnrichedFields: () => {},
      clearEnrichedFields: () => {},
      isFieldEnriched: () => false,
      getEnrichedFieldInfo: () => null,
    };
  }
  return context;
}

