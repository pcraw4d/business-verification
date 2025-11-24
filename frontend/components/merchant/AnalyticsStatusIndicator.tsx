'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { getMerchantAnalyticsStatus, type MerchantAnalyticsStatus } from '@/lib/api';
import { Loader2, CheckCircle2, XCircle, Clock } from 'lucide-react';

interface AnalyticsStatusIndicatorProps {
  merchantId: string;
  type: 'classification' | 'websiteAnalysis';
}

export function AnalyticsStatusIndicator({ merchantId, type }: AnalyticsStatusIndicatorProps) {
  const [status, setStatus] = useState<'pending' | 'processing' | 'completed' | 'failed' | 'skipped'>('pending');
  const [isPolling, setIsPolling] = useState(false);

  useEffect(() => {
    let pollInterval: NodeJS.Timeout | null = null;

    const fetchStatus = async () => {
      try {
        const statusData: MerchantAnalyticsStatus = await getMerchantAnalyticsStatus(merchantId);
        const currentStatus = statusData.status[type === 'classification' ? 'classification' : 'websiteAnalysis'];
        setStatus(currentStatus);

        // Poll if status is processing or pending
        if (currentStatus === 'processing' || currentStatus === 'pending') {
          setIsPolling(true);
          if (!pollInterval) {
            pollInterval = setInterval(fetchStatus, 3000); // Poll every 3 seconds
          }
        } else {
          setIsPolling(false);
          if (pollInterval) {
            clearInterval(pollInterval);
            pollInterval = null;
          }
        }
      } catch (error) {
        console.error('Failed to fetch analytics status:', error);
        setIsPolling(false);
        if (pollInterval) {
          clearInterval(pollInterval);
          pollInterval = null;
        }
      }
    };

    // Initial fetch
    fetchStatus();

    // Cleanup on unmount
    return () => {
      if (pollInterval) {
        clearInterval(pollInterval);
      }
    };
  }, [merchantId, type]);

  const getStatusBadge = () => {
    switch (status) {
      case 'processing':
        return (
          <Badge variant="outline" className="flex items-center gap-1">
            <Loader2 className="h-3 w-3 animate-spin" />
            Processing...
          </Badge>
        );
      case 'completed':
        return (
          <Badge variant="outline" className="flex items-center gap-1 text-green-600 border-green-600">
            <CheckCircle2 className="h-3 w-3" />
            Completed
          </Badge>
        );
      case 'failed':
        return (
          <Badge variant="outline" className="flex items-center gap-1 text-red-600 border-red-600">
            <XCircle className="h-3 w-3" />
            Failed
          </Badge>
        );
      case 'skipped':
        return (
          <Badge variant="outline" className="flex items-center gap-1 text-gray-500">
            <Clock className="h-3 w-3" />
            Skipped
          </Badge>
        );
      case 'pending':
      default:
        return (
          <Badge variant="outline" className="flex items-center gap-1 text-gray-500">
            <Clock className="h-3 w-3" />
            Pending
          </Badge>
        );
    }
  };

  return <div className="inline-flex">{getStatusBadge()}</div>;
}

