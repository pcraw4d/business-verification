'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { AlertCircle, CheckCircle2, Info, RefreshCw, ArrowRight } from 'lucide-react';
import { useEffect, useState, useCallback, useMemo } from 'react';
import { getRiskRecommendations } from '@/lib/api';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { ChevronDown, ChevronUp } from 'lucide-react';

interface RiskRecommendationsSectionProps {
  merchantId: string;
}

type Recommendation = {
  id: string;
  type: string;
  priority: 'high' | 'medium' | 'low';
  title: string;
  description: string;
  actionItems: string[];
};

type RiskRecommendationsResponse = {
  merchantId: string;
  recommendations: Recommendation[];
  timestamp: string;
};

export function RiskRecommendationsSection({ merchantId }: RiskRecommendationsSectionProps) {
  const [recommendations, setRecommendations] = useState<RiskRecommendationsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedPriority, setExpandedPriority] = useState<'high' | 'medium' | 'low' | null>(null);
  const [mounted, setMounted] = useState(false);

  const fetchRecommendations = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const data = await getRiskRecommendations(merchantId);
      // Map API response to component type, ensuring priority is correctly typed
      const mappedData: RiskRecommendationsResponse = {
        ...data,
        recommendations: data.recommendations.map((rec) => ({
          ...rec,
          priority: (rec.priority === 'high' || rec.priority === 'medium' || rec.priority === 'low' 
            ? rec.priority 
            : 'medium') as 'high' | 'medium' | 'low',
        })),
      };
      setRecommendations(mappedData);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load risk recommendations';
      setError(errorMessage);
      toast.error('Failed to load recommendations', {
        description: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  }, [merchantId]);

  useEffect(() => {
    fetchRecommendations();
  }, [fetchRecommendations]);

  // Group recommendations by priority
  const groupedRecommendations = useMemo((): Record<'high' | 'medium' | 'low', Recommendation[]> => {
    const grouped: Record<'high' | 'medium' | 'low', Recommendation[]> = {
      high: [],
      medium: [],
      low: [],
    };
    
    if (!recommendations?.recommendations) return grouped;
    
    recommendations.recommendations.forEach((rec) => {
      const priority = rec.priority || 'medium';
      if (grouped[priority]) {
        grouped[priority].push(rec);
      }
    });
    
    return grouped;
  }, [recommendations]);

  const totalRecommendations = useMemo(() => {
    return (Object.values(groupedRecommendations) as Recommendation[][]).reduce((sum: number, recs: Recommendation[]) => sum + recs.length, 0);
  }, [groupedRecommendations]);

  const getPriorityConfig = (priority: 'high' | 'medium' | 'low') => {
    switch (priority) {
      case 'high':
        return {
          label: 'High Priority',
          icon: <AlertCircle className="h-4 w-4" />,
          variant: 'destructive' as const,
          color: 'text-red-600',
          bgColor: 'bg-red-50',
          borderColor: 'border-red-200',
        };
      case 'medium':
        return {
          label: 'Medium Priority',
          icon: <Info className="h-4 w-4" />,
          variant: 'secondary' as const,
          color: 'text-yellow-600',
          bgColor: 'bg-yellow-50',
          borderColor: 'border-yellow-200',
        };
      case 'low':
        return {
          label: 'Low Priority',
          icon: <CheckCircle2 className="h-4 w-4" />,
          variant: 'outline' as const,
          color: 'text-blue-600',
          bgColor: 'bg-blue-50',
          borderColor: 'border-blue-200',
        };
    }
  };

  const togglePriority = (priority: 'high' | 'medium' | 'low') => {
    setExpandedPriority(expandedPriority === priority ? null : priority);
  };

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Recommendations</CardTitle>
          <CardDescription>Actionable recommendations to improve risk profile</CardDescription>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-64 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle>Risk Recommendations</CardTitle>
          <CardDescription>Actionable recommendations to improve risk profile</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center p-6 space-y-4">
            <AlertCircle className="h-8 w-8 text-destructive" />
            <p className="text-sm text-destructive-foreground">{error}</p>
            <Button onClick={fetchRecommendations} variant="outline" size="sm">
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (totalRecommendations === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Risk Recommendations</CardTitle>
          <CardDescription>Actionable recommendations to improve risk profile</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center p-8 text-center">
            <div className="rounded-full bg-green-100 p-3 mb-4">
              <CheckCircle2 className="h-6 w-6 text-green-600" />
            </div>
            <p className="text-sm font-medium text-muted-foreground">No Recommendations</p>
            <p className="text-xs text-muted-foreground mt-1">
              All risk factors are within acceptable thresholds. No immediate actions required.
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
            <CardTitle>Risk Recommendations</CardTitle>
            <CardDescription>
              {totalRecommendations} actionable recommendation{totalRecommendations !== 1 ? 's' : ''} to improve risk profile
            </CardDescription>
          </div>
          <Button onClick={fetchRecommendations} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {(['high', 'medium', 'low'] as const).map((priority) => {
          const priorityRecs = groupedRecommendations[priority];
          if (priorityRecs.length === 0) return null;

          const config = getPriorityConfig(priority);
          const isExpanded = expandedPriority === priority;

          return (
            <Collapsible
              key={priority}
              open={isExpanded}
              onOpenChange={() => togglePriority(priority)}
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
                        <Badge variant={config.variant}>{priorityRecs.length}</Badge>
                      </div>
                      <p className="text-xs text-muted-foreground mt-0.5">
                        {priorityRecs.length} recommendation{priorityRecs.length !== 1 ? 's' : ''}
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
                <div className="space-y-3 pl-7">
                  {priorityRecs.map((rec) => (
                    <div
                      key={rec.id}
                      className={`p-4 rounded-lg border ${config.borderColor} ${config.bgColor}`}
                    >
                      <div className="space-y-3">
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-1">
                              <h4 className="font-semibold text-sm">{rec.title}</h4>
                              {rec.type && (
                                <Badge variant="outline" className="text-xs">
                                  {rec.type}
                                </Badge>
                              )}
                            </div>
                            {rec.description && (
                              <p className="text-sm text-muted-foreground">{rec.description}</p>
                            )}
                          </div>
                          <Badge variant={config.variant} className="shrink-0">
                            {config.label}
                          </Badge>
                        </div>
                        
                        {rec.actionItems && rec.actionItems.length > 0 && (
                          <div className="mt-3 pt-3 border-t border-current/20">
                            <p className="text-xs font-medium text-muted-foreground mb-2">
                              Action Items:
                            </p>
                            <ul className="space-y-1.5">
                              {rec.actionItems.map((action, index) => (
                                <li key={index} className="flex items-start gap-2 text-sm">
                                  <ArrowRight className="h-3 w-3 mt-1.5 shrink-0 text-muted-foreground" />
                                  <span className="text-muted-foreground">{action}</span>
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </CollapsibleContent>
            </Collapsible>
          );
        })}

        {recommendations?.timestamp && (
          <div className="pt-4 border-t">
            <p className="text-xs text-muted-foreground text-center">
              Last updated: {mounted ? new Date(recommendations.timestamp).toLocaleString() : 'Loading...'}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

