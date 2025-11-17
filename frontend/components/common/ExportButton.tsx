'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Download } from 'lucide-react';
import { toast } from 'sonner';

interface ExportButtonProps {
  merchantId: string;
  format?: 'csv' | 'pdf' | 'json' | 'excel';
  data?: unknown;
  filename?: string;
}

export function ExportButton({
  merchantId,
  format = 'csv',
  data,
  filename,
}: ExportButtonProps) {
  const [exporting, setExporting] = useState(false);

  async function handleExport() {
    try {
      setExporting(true);
      
      if (format === 'json' && data) {
        // Client-side JSON export
        const jsonStr = JSON.stringify(data, null, 2);
        const blob = new Blob([jsonStr], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename || `merchant-${merchantId}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        toast.success('Data exported successfully');
        return;
      }

      // Server-side export for other formats
      const token = typeof window !== 'undefined' ? sessionStorage.getItem('authToken') : null;
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const { ApiEndpoints } = await import('@/lib/api-config');
      const response = await fetch(ApiEndpoints.merchants.export(merchantId, format), {
        method: 'GET',
        headers,
      });

      if (response.ok) {
        const blob = await response.blob();
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename || `merchant-${merchantId}.${format}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        toast.success('Data exported successfully');
      } else {
        throw new Error('Export failed');
      }
    } catch (error) {
      toast.error('Failed to export data');
      console.error('Export error:', error);
    } finally {
      setExporting(false);
    }
  }

  return (
    <Button variant="outline" onClick={handleExport} disabled={exporting} aria-label={exporting ? `Exporting ${format.toUpperCase()}` : `Export data as ${format.toUpperCase()}`}>
      <Download className="h-4 w-4 mr-2" />
      {exporting ? 'Exporting...' : `Export ${format.toUpperCase()}`}
    </Button>
  );
}

