'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Download, FileText, FileSpreadsheet, FileJson, File } from 'lucide-react';
import { toast } from 'sonner';
import {
  exportToCSV,
  exportToJSON,
  exportToExcel,
  exportToPDF,
  exportViaAPI,
  generateFilename,
  type ExportFormat,
} from '@/lib/export';

interface ExportButtonProps {
  data: any | (() => Promise<any>);
  exportType?: 'merchant' | 'risk' | 'report' | 'analytics';
  formats?: ExportFormat[];
  filename?: string;
  merchantId?: string;
  onExportStart?: (format: ExportFormat) => void;
  onExportComplete?: (format: ExportFormat, result: any) => void;
  onExportError?: (format: ExportFormat, error: Error) => void;
  useAPI?: boolean;
  apiEndpoint?: string;
}

export function ExportButton({
  data,
  exportType = 'merchant',
  formats = ['csv', 'pdf', 'json', 'excel'],
  filename,
  merchantId,
  onExportStart,
  onExportComplete,
  onExportError,
  useAPI = false,
  apiEndpoint,
}: ExportButtonProps) {
  const [isExporting, setIsExporting] = useState(false);
  const [exportingFormat, setExportingFormat] = useState<ExportFormat | null>(null);

  const handleExport = async (format: ExportFormat) => {
    try {
      setIsExporting(true);
      setExportingFormat(format);

      // Call onExportStart callback
      if (onExportStart) {
        onExportStart(format);
      }

      // Get data
      const exportData = typeof data === 'function' ? await data() : data;
      
      if (!exportData) {
        throw new Error('No data available to export');
      }

      // Generate filename
      const exportFilename = filename || generateFilename(exportType, format, merchantId);

      let result;

      if (useAPI && apiEndpoint) {
        // Use API endpoint for export
        result = await exportViaAPI(exportData, format, apiEndpoint);
      } else {
        // Use client-side export
        switch (format) {
          case 'csv':
            result = exportToCSV(Array.isArray(exportData) ? exportData : [exportData], exportFilename);
            break;
          case 'json':
            result = exportToJSON(exportData, exportFilename);
            break;
          case 'excel':
          case 'xlsx':
            result = await exportToExcel(
              Array.isArray(exportData) ? exportData : [exportData],
              exportFilename
            );
            break;
          case 'pdf':
            result = await exportToPDF(exportData, exportFilename, {
              title: `${exportType.charAt(0).toUpperCase() + exportType.slice(1)} Export`,
            });
            break;
          default:
            throw new Error(`Unsupported export format: ${format}`);
        }
      }

      if (result.success) {
        toast.success('Export completed', {
          description: `File "${result.filename || exportFilename}" has been downloaded`,
        });

        // Call onExportComplete callback
        if (onExportComplete) {
          onExportComplete(format, result);
        }
      } else {
        throw new Error(result.error || 'Export failed');
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Export failed';
      toast.error('Export failed', {
        description: errorMessage,
      });

      // Call onExportError callback
      if (onExportError) {
        onExportError(format, error instanceof Error ? error : new Error(errorMessage));
      }
    } finally {
      setIsExporting(false);
      setExportingFormat(null);
    }
  };

  const getFormatIcon = (format: ExportFormat) => {
    switch (format) {
      case 'csv':
      case 'excel':
      case 'xlsx':
        return <FileSpreadsheet className="h-4 w-4" />;
      case 'pdf':
        return <FileText className="h-4 w-4" />;
      case 'json':
        return <FileJson className="h-4 w-4" />;
      default:
        return <File className="h-4 w-4" />;
    }
  };

  const getFormatLabel = (format: ExportFormat) => {
    switch (format) {
      case 'excel':
      case 'xlsx':
        return 'Excel';
      case 'csv':
        return 'CSV';
      case 'pdf':
        return 'PDF';
      case 'json':
        return 'JSON';
      default:
        return String(format).toUpperCase();
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          disabled={isExporting}
          className="gap-2"
        >
          <Download className={`h-4 w-4 ${isExporting ? 'animate-spin' : ''}`} />
          {isExporting && exportingFormat
            ? `Exporting ${getFormatLabel(exportingFormat)}...`
            : 'Export'}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {formats.map((format) => (
          <DropdownMenuItem
            key={format}
            onClick={() => handleExport(format)}
            disabled={isExporting}
            className="gap-2"
          >
            {getFormatIcon(format)}
            Export as {getFormatLabel(format)}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

