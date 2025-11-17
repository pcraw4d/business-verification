/**
 * Export utilities for converting data to various formats
 */

export type ExportFormat = 'csv' | 'pdf' | 'json' | 'excel' | 'xlsx';

export interface ExportOptions {
  format: ExportFormat;
  filename?: string;
  includeCharts?: boolean;
  includeExplanations?: boolean;
  includeScenarios?: boolean;
}

export interface ExportResult {
  success: boolean;
  filename?: string;
  downloadUrl?: string;
  error?: string;
}

/**
 * Export data to CSV format
 */
export function exportToCSV(data: any[], filename: string = 'export.csv'): ExportResult {
  try {
    if (!Array.isArray(data) || data.length === 0) {
      throw new Error('Data must be a non-empty array');
    }

    // Get headers from first object
    const headers = Object.keys(data[0]);
    
    // Create CSV content
    const csvRows = [
      headers.join(','), // Header row
      ...data.map(row => 
        headers.map(header => {
          const value = row[header];
          // Escape commas and quotes
          if (value === null || value === undefined) return '';
          const stringValue = String(value);
          if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) {
            return `"${stringValue.replace(/"/g, '""')}"`;
          }
          return stringValue;
        }).join(',')
      )
    ];

    const csvContent = csvRows.join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    
    downloadFile(url, filename);
    URL.revokeObjectURL(url);

    return { success: true, filename, downloadUrl: url };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to export CSV',
    };
  }
}

/**
 * Export data to JSON format
 */
export function exportToJSON(data: any, filename: string = 'export.json'): ExportResult {
  try {
    const jsonContent = JSON.stringify(data, null, 2);
    const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    
    downloadFile(url, filename);
    URL.revokeObjectURL(url);

    return { success: true, filename, downloadUrl: url };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to export JSON',
    };
  }
}

/**
 * Export data to Excel format (XLSX)
 * Note: This requires a library like xlsx or exceljs
 */
export async function exportToExcel(
  data: any[],
  filename: string = 'export.xlsx',
  sheetName: string = 'Sheet1'
): Promise<ExportResult> {
  try {
    // Dynamic import of xlsx library
    const XLSX = await import('xlsx').catch(() => null);
    
    if (!XLSX) {
      // Fallback to CSV if xlsx library is not available
      console.warn('xlsx library not available, falling back to CSV');
      return exportToCSV(data, filename.replace(/\.xlsx?$/, '.csv'));
    }
    
    if (!Array.isArray(data) || data.length === 0) {
      throw new Error('Data must be a non-empty array');
    }

    // Create workbook and worksheet
    const worksheet = XLSX.utils.json_to_sheet(data);
    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, worksheet, sheetName);

    // Generate Excel file
    const excelBuffer = XLSX.write(workbook, { bookType: 'xlsx', type: 'array' });
    const blob = new Blob([excelBuffer], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    const url = URL.createObjectURL(blob);
    
    downloadFile(url, filename);
    URL.revokeObjectURL(url);

    return { success: true, filename, downloadUrl: url };
  } catch (error) {
    // Fallback to CSV on any error
    console.warn('Excel export failed, falling back to CSV:', error);
    return exportToCSV(data, filename.replace(/\.xlsx?$/, '.csv'));
  }
}

/**
 * Export data to PDF format
 * Note: This requires a library like jsPDF or pdfkit
 */
export async function exportToPDF(
  data: any,
  filename: string = 'export.pdf',
  options: { title?: string; includeCharts?: boolean } = {}
): Promise<ExportResult> {
  try {
    // Dynamic import of jsPDF
    const jsPDFModule = await import('jspdf').catch(() => null);
    
    if (!jsPDFModule) {
      throw new Error('jsPDF library not available');
    }
    
    // Handle both default and named exports
    const jsPDF = (jsPDFModule as any).default || jsPDFModule.jsPDF || jsPDFModule;
    const doc = new jsPDF();
    const pageWidth = doc.internal.pageSize.getWidth();
    const pageHeight = doc.internal.pageSize.getHeight();
    let yPosition = 20;
    const margin = 20;
    const lineHeight = 7;

    // Add title
    if (options.title) {
      doc.setFontSize(16);
      doc.text(options.title, margin, yPosition);
      yPosition += lineHeight * 2;
    }

    // Add content
    doc.setFontSize(10);
    
    if (Array.isArray(data)) {
      // Table format for arrays
      const headers = Object.keys(data[0] || {});
      const rows = data.map(row => headers.map(header => String(row[header] || '')));
      
      // Simple table rendering
      doc.text('Data Export', margin, yPosition);
      yPosition += lineHeight;
      
      // Add headers
      headers.forEach((header, index) => {
        const xPos = margin + (index * (pageWidth - 2 * margin) / headers.length);
        doc.text(header, xPos, yPosition);
      });
      yPosition += lineHeight;
      
      // Add rows (limited to fit on page)
      rows.slice(0, 20).forEach(row => {
        if (yPosition > pageHeight - 20) {
          doc.addPage();
          yPosition = 20;
        }
        row.forEach((cell, index) => {
          const xPos = margin + (index * (pageWidth - 2 * margin) / headers.length);
          doc.text(cell.substring(0, 20), xPos, yPosition);
        });
        yPosition += lineHeight;
      });
    } else {
      // Text format for objects
      const text = JSON.stringify(data, null, 2);
      const lines = doc.splitTextToSize(text, pageWidth - 2 * margin);
      
      lines.forEach((line: string) => {
        if (yPosition > pageHeight - 20) {
          doc.addPage();
          yPosition = 20;
        }
        doc.text(line, margin, yPosition);
        yPosition += lineHeight;
      });
    }

    // Save PDF
    doc.save(filename);

    return { success: true, filename };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to export PDF',
    };
  }
}

/**
 * Download file from URL
 */
function downloadFile(url: string, filename: string): void {
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

/**
 * Generate filename with timestamp
 */
export function generateFilename(
  prefix: string,
  format: ExportFormat,
  merchantId?: string
): string {
  const timestamp = new Date().toISOString().split('T')[0].replace(/-/g, '');
  const extension = format === 'excel' ? 'xlsx' : format;
  const merchantSuffix = merchantId ? `_${merchantId}` : '';
  return `${prefix}_${timestamp}${merchantSuffix}.${extension}`;
}

/**
 * Export data using API endpoint
 */
export async function exportViaAPI(
  data: any,
  format: ExportFormat,
  endpoint: string = '/api/v1/export'
): Promise<ExportResult> {
  try {
    const token = typeof window !== 'undefined' 
      ? sessionStorage.getItem('authToken') 
      : null;

    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
      },
      body: JSON.stringify({
        data,
        format,
      }),
    });

    if (!response.ok) {
      throw new Error(`Export failed: ${response.statusText}`);
    }

    const result = await response.json();
    
    // If API returns a download URL, open it
    if (result.download_url || result.file_url) {
      window.open(result.download_url || result.file_url, '_blank');
      return {
        success: true,
        filename: result.filename,
        downloadUrl: result.download_url || result.file_url,
      };
    }

    // If API returns blob data, download it
    if (response.headers.get('content-type')?.includes('application/')) {
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const filename = response.headers.get('content-disposition')?.split('filename=')[1] || 'export';
      downloadFile(url, filename);
      URL.revokeObjectURL(url);
      return { success: true, filename };
    }

    return { success: true, ...result };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to export via API',
    };
  }
}

