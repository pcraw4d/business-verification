'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { AlertCircle, AlertTriangle, Info, Bell, RefreshCw } from 'lucide-react';
import { useEffect, useState, useCallback, useMemo } from 'react';
import { getRiskAlerts } from '@/lib/api';
import type { RiskIndicatorsData, RiskIndicator } from '@/types/merchant';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { ChevronDown, ChevronUp } from 'lucide-react';

interface RiskAlertsSectionProps {
  merchantId: string;
}

type Severity = 'critical' | 'high' | 'medium' | 'low';

export function RiskAlertsSection({ merchantId }: RiskAlertsSectionProps) {
  const [alerts, setAlerts] = useState<RiskIndicatorsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedSeverity, setExpandedSeverity] = useState<Severity | null>(null);

  const fetchAlerts = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const data = await getRiskAlerts(merchantId);
      setAlerts(data);

      // Show toast notification if there are critical or high severity alerts
      if (data.indicators) {
        const criticalAlerts = data.indicators.filter(a => a.severity === 'critical');
        const highAlerts = data.indicators.filter(a => a.severity === 'high');
        
        if (criticalAlerts.length > 0) {
          toast.error(`${criticalAlerts.length} Critical Alert${criticalAlerts.length > 1 ? 's' : ''}`, {
            description: 'Immediate attention required',
            duration: 10000,
          });
        } else if (highAlerts.length > 0) {
          toast.warning(`${highAlerts.length} High Priority Alert${highAlerts.length > 1 ? 's' : ''}`, {
            description: 'Review recommended',
            duration: 8000,
          });
        }
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load risk alerts';
      setError(errorMessage);
      toast.error('Failed to load alerts', {
        description: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  }, [merchantId]);

  useEffect(() => {
    fetchAlerts();
    
    // Refresh alerts every 5 minutes
    const interval = setInterval(fetchAlerts, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, [fetchAlerts]);

  // Group alerts by severity
  const groupedAlerts = useMemo(() => {
    if (!alerts?.indicators) return {};
    
    const grouped: Record<Severity, RiskIndicator[]> = {
      critical: [],
      high: [],
      medium: [],
      low: [],
    };
    
    alerts.indicators.forEach((alert) => {
      const severity = (alert.severity || 'medium') as Severity;
      if (grouped[severity]) {
        grouped[severity].push(alert);
      }
    });
    
    return grouped;
  }, [alerts]);

  const totalAlerts = useMemo(() => {
    return Object.values(groupedAlerts).reduce((sum, alerts) => sum + alerts.length, 0);
  }, [groupedAlerts]);

  const getSeverityConfig = (severity: Severity) => {
    switch (severity) {
      case 'critical':
        return {
          label: 'Critical',
          icon: <AlertCircle className="h-4 w-4" />,
          variant: 'destructive' as const,
          color: 'text-red-600',
          bgColor: 'bg-red-50',
          borderColor: 'border-red-200',
        };
      case 'high':
        return {
          label: 'High',
          icon: <AlertTriangle className="h-4 w-4" />,
          variant: 'destructive' as const,
          color: 'text-orange-600',
          bgColor: 'bg-orange-50',
          borderColor: 'border-orange-200',
        };
      case 'medium':
        return {
          label: 'Medium',
          icon: <AlertTriangle className="h-4 w-4" />,
          variant: 'secondary' as const,
          color: 'text-yellow-600',
          bgColor: 'bg-yellow-50',
          borderColor: 'border-yellow-200',
        };
      case 'low':
        return {
          label: 'Low',
          icon: <Info className="h-4 w-4" />,
          variant: 'outline' as const,
          color: 'text-blue-600',
          bgColor: 'bg-blue-50',
          borderColor: 'border-blue-200',
        };
    }
  };

  const toggleSeverity = (severity: Severity) => {
    setExpandedSeverity(expandedSeverity === severity ? null : severity);
  };

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Alerts</CardTitle>
          <CardDescription>Active risk alerts for this merchant</CardDescription>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-32 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle>Risk Alerts</CardTitle>
          <CardDescription>Active risk alerts for this merchant</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-destructive-foreground">{error}</p>
          <Button onClick={fetchAlerts} className="mt-4" variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (totalAlerts === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5" /> Risk Alerts
          </CardTitle>
          <CardDescription>Active risk alerts for this merchant</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center p-8 text-center">
            <div className="rounded-full bg-green-100 p-3 mb-4">
              <Bell className="h-6 w-6 text-green-600" />
            </div>
            <p className="text-sm font-medium text-muted-foreground">No Active Alerts</p>
            <p className="text-xs text-muted-foreground mt-1">
              All risk indicators are within acceptable thresholds.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" /> Risk Alerts
            </CardTitle>
            <CardDescription>
              {totalAlerts} active alert{totalAlerts !== 1 ? 's' : ''} requiring attention
            </CardDescription>
          </div>
          <Button onClick={fetchAlerts} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {(['critical', 'high', 'medium', 'low'] as Severity[]).map((severity) => {
          const severityAlerts = groupedAlerts[severity];
          if (severityAlerts.length === 0) return null;

          const config = getSeverityConfig(severity);
          const isExpanded = expandedSeverity === severity;

          return (
            <Collapsible
              key={severity}
              open={isExpanded}
              onOpenChange={() => toggleSeverity(severity)}
            >
              <CollapsibleTrigger asChild>
                <Button
                  variant="ghost"
                  className={`w-full justify-between p-4 h-auto ${config.bgColor} ${config.borderColor} border rounded-lg hover:${config.bgColor}`}
                >
                  <div className="flex items-center gap-3">
                    <div className={config.color}>{config.icon}</div>
                    <div className="text-left">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold">{config.label}</span>
                        <Badge variant={config.variant}>{severityAlerts.length}</Badge>
                      </div>
                      <p className="text-xs text-muted-foreground mt-0.5">
                        {severityAlerts.length} alert{severityAlerts.length !== 1 ? 's' : ''}
                      </p>
                    </div>
                  </div>
                  {isExpanded ? (
                    <ChevronUp className="h-4 w-4" />
                  ) : (
                    <ChevronDown className="h-4 w-4" />
                  )}
                </Button>
              </CollapsibleTrigger>
              <CollapsibleContent className="mt-2">
                <div className="space-y-2 pl-7">
                  {severityAlerts.map((alert) => (
                    <div
                      key={alert.id}
                      className={`p-3 rounded-lg border ${config.borderColor} ${config.bgColor}`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex-1">
                          <h4 className="font-medium text-sm">{alert.title}</h4>
                          {alert.description && (
                            <p className="text-xs text-muted-foreground mt-1">
                              {alert.description}
                            </p>
                          )}
                          {alert.createdAt && (
                            <p className="text-xs text-muted-foreground mt-2">
                              Triggered: {new Date(alert.createdAt).toLocaleString()}
                            </p>
                          )}
                        </div>
                        <Badge variant={config.variant} className="shrink-0">
                          {config.label}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </CollapsibleContent>
            </Collapsible>
          );
        })}
      </CardContent>
    </Card>
  );
}

