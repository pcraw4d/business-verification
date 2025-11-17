'use client';

import { useState, useEffect, useCallback } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Progress } from '@/components/ui/progress';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Play,
  Pause,
  Square,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Info,
  Upload,
  Download,
  Users,
  Filter,
  CheckSquare,
  Square as SquareIcon,
} from 'lucide-react';
import { toast } from 'sonner';
import { getMerchantsList } from '@/lib/api';
import { ExportButton } from '@/components/export/ExportButton';
import type { MerchantListItem } from '@/types/merchant';

type OperationType =
  | 'update-portfolio'
  | 'update-risk'
  | 'export-data'
  | 'send-notifications'
  | 'schedule-review'
  | 'bulk-deactivate';

interface OperationState {
  status: 'ready' | 'running' | 'paused' | 'completed' | 'failed';
  progress: number;
  completed: number;
  failed: number;
  total: number;
  currentIndex: number;
  operationId: string | null;
}

interface OperationLog {
  timestamp: string;
  level: 'info' | 'success' | 'error' | 'warning';
  message: string;
}

export function BulkOperationsManager() {
  const [merchants, setMerchants] = useState<MerchantListItem[]>([]);
  const [selectedMerchants, setSelectedMerchants] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [currentOperation, setCurrentOperation] = useState<OperationType | null>(null);
  const [operationState, setOperationState] = useState<OperationState>({
    status: 'ready',
    progress: 0,
    completed: 0,
    failed: 0,
    total: 0,
    currentIndex: 0,
    operationId: null,
  });
  const [operationLogs, setOperationLogs] = useState<OperationLog[]>([]);
  const [filters, setFilters] = useState({
    search: '',
    status: 'all',
    riskLevel: 'all',
  });

  // Operation configuration
  const [operationConfig, setOperationConfig] = useState<Record<string, any>>({});

  // Load merchants
  useEffect(() => {
    loadMerchants();
  }, [filters]);

  const loadMerchants = async () => {
    try {
      setLoading(true);
      const response = await getMerchantsList({
        page: 1,
        pageSize: 100,
        search: filters.search || undefined,
        status: filters.status !== 'all' ? filters.status : undefined,
        riskLevel: filters.riskLevel !== 'all' ? filters.riskLevel : undefined,
      });
      setMerchants(response.merchants);
      addLog('info', `Loaded ${response.merchants.length} merchants`);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load merchants';
      toast.error('Failed to load merchants', { description: errorMessage });
      addLog('error', errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const addLog = (level: OperationLog['level'], message: string) => {
    setOperationLogs((prev) => [
      ...prev,
      {
        timestamp: new Date().toISOString(),
        level,
        message,
      },
    ]);
  };

  const toggleMerchantSelection = (merchantId: string) => {
    setSelectedMerchants((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(merchantId)) {
        newSet.delete(merchantId);
      } else {
        newSet.add(merchantId);
      }
      return newSet;
    });
  };

  const selectAll = () => {
    setSelectedMerchants(new Set(merchants.map((m) => m.id)));
    addLog('info', `Selected all ${merchants.length} merchants`);
  };

  const deselectAll = () => {
    setSelectedMerchants(new Set());
    addLog('info', 'Deselected all merchants');
  };

  const selectByFilter = () => {
    const filtered = merchants.filter(
      (m) => m.status === 'pending' || m.risk_level === 'high' || m.risk_level === 'critical'
    );
    setSelectedMerchants(new Set(filtered.map((m) => m.id)));
    addLog('info', `Selected ${filtered.length} merchants by filter (pending or high risk)`);
  };

  const handleOperationSelect = (operation: OperationType) => {
    setCurrentOperation(operation);
    addLog('info', `Selected operation: ${operation}`);
  };

  const handleStartOperation = async () => {
    if (!currentOperation) {
      toast.error('Please select an operation type');
      return;
    }

    if (selectedMerchants.size === 0) {
      toast.error('Please select at least one merchant');
      return;
    }

    const merchantIds = Array.from(selectedMerchants);
    setOperationState({
      status: 'running',
      progress: 0,
      completed: 0,
      failed: 0,
      total: merchantIds.length,
      currentIndex: 0,
      operationId: `op_${Date.now()}`,
    });

    addLog('info', `Starting ${currentOperation} operation on ${merchantIds.length} merchants`);

    try {
      switch (currentOperation) {
        case 'update-portfolio':
          await performBulkPortfolioUpdate(merchantIds);
          break;
        case 'update-risk':
          await performBulkRiskUpdate(merchantIds);
          break;
        case 'export-data':
          await performBulkExport(merchantIds);
          break;
        default:
          toast.error('Operation not yet implemented');
          setOperationState((prev) => ({ ...prev, status: 'failed' }));
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Operation failed';
      addLog('error', errorMessage);
      setOperationState((prev) => ({ ...prev, status: 'failed' }));
    }
  };

  const performBulkPortfolioUpdate = async (merchantIds: string[]) => {
    try {
      const token = typeof window !== 'undefined' ? sessionStorage.getItem('authToken') : null;
      const { ApiEndpoints } = await import('@/lib/api-config');
      const response = await fetch(ApiEndpoints.merchants.bulkUpdate(), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { Authorization: `Bearer ${token}` }),
        },
        body: JSON.stringify({
          merchant_ids: merchantIds,
          operation: 'update_portfolio_type',
          portfolio_type: operationConfig.portfolioType || 'onboarded',
        }),
      });

      if (!response.ok) {
        throw new Error(`Bulk update failed: ${response.statusText}`);
      }

      const result = await response.json();
      addLog('success', `Bulk portfolio update completed: ${result.processed || merchantIds.length} merchants updated`);
      setOperationState((prev) => ({
        ...prev,
        status: 'completed',
        progress: 100,
        completed: result.processed || merchantIds.length,
      }));
      toast.success('Bulk operation completed');
      
      // Reload merchants to reflect changes
      await loadMerchants();
    } catch (error) {
      throw error;
    }
  };

  const performBulkRiskUpdate = async (merchantIds: string[]) => {
    try {
      const token = typeof window !== 'undefined' ? sessionStorage.getItem('authToken') : null;
      const { ApiEndpoints } = await import('@/lib/api-config');
      const response = await fetch(ApiEndpoints.merchants.bulkUpdate(), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { Authorization: `Bearer ${token}` }),
        },
        body: JSON.stringify({
          merchant_ids: merchantIds,
          operation: 'update_risk_level',
          risk_level: operationConfig.riskLevel || 'medium',
        }),
      });

      if (!response.ok) {
        throw new Error(`Bulk risk update failed: ${response.statusText}`);
      }

      const result = await response.json();
      addLog('success', `Bulk risk update completed: ${result.processed || merchantIds.length} merchants updated`);
      setOperationState((prev) => ({
        ...prev,
        status: 'completed',
        progress: 100,
        completed: result.processed || merchantIds.length,
      }));
      toast.success('Bulk operation completed');
      
      // Reload merchants to reflect changes
      await loadMerchants();
    } catch (error) {
      throw error;
    }
  };

  const performBulkExport = async (merchantIds: string[]) => {
    // Export is handled by ExportButton component
    addLog('info', 'Export operation - use export button');
    setOperationState((prev) => ({
      ...prev,
      status: 'completed',
      progress: 100,
    }));
  };

  const handlePause = () => {
    setOperationState((prev) => ({ ...prev, status: 'paused' }));
    addLog('info', 'Operation paused');
  };

  const handleResume = () => {
    setOperationState((prev) => ({ ...prev, status: 'running' }));
    addLog('info', 'Operation resumed');
  };

  const handleCancel = () => {
    setOperationState({
      status: 'ready',
      progress: 0,
      completed: 0,
      failed: 0,
      total: 0,
      currentIndex: 0,
      operationId: null,
    });
    addLog('info', 'Operation cancelled');
  };

  const getLogIcon = (level: OperationLog['level']) => {
    switch (level) {
      case 'success':
        return <CheckCircle2 className="h-4 w-4 text-green-600" />;
      case 'error':
        return <XCircle className="h-4 w-4 text-red-600" />;
      case 'warning':
        return <AlertCircle className="h-4 w-4 text-yellow-600" />;
      default:
        return <Info className="h-4 w-4 text-blue-600" />;
    }
  };

  return (
    <div className="space-y-6">
      {/* Selection Stats */}
      <Card>
        <CardHeader>
          <CardTitle>Merchant Selection</CardTitle>
          <CardDescription>
            Select merchants to perform bulk operations on
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="text-sm">
                <span className="font-medium">{merchants.length}</span> total merchants
              </div>
              <div className="text-sm">
                <span className="font-medium">{selectedMerchants.size}</span> selected
              </div>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={selectAll} aria-label="Select all merchants">
                <CheckSquare className="h-4 w-4 mr-2" />
                Select All
              </Button>
              <Button variant="outline" size="sm" onClick={deselectAll} aria-label="Deselect all merchants">
                <SquareIcon className="h-4 w-4 mr-2" />
                Deselect All
              </Button>
              <Button variant="outline" size="sm" onClick={selectByFilter} aria-label="Select merchants by current filter">
                <Filter className="h-4 w-4 mr-2" />
                Select by Filter
              </Button>
            </div>
          </div>

          {/* Filters */}
          <div className="flex gap-4">
            <Input
              placeholder="Search merchants..."
              value={filters.search}
              onChange={(e) => setFilters((prev) => ({ ...prev, search: e.target.value }))}
              className="flex-1"
              aria-label="Search merchants"
            />
            <Select
              value={filters.status}
              onValueChange={(value) => setFilters((prev) => ({ ...prev, status: value }))}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="suspended">Suspended</SelectItem>
              </SelectContent>
            </Select>
            <Select
              value={filters.riskLevel}
              onValueChange={(value) => setFilters((prev) => ({ ...prev, riskLevel: value }))}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Risk Level" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Risk Levels</SelectItem>
                <SelectItem value="low">Low</SelectItem>
                <SelectItem value="medium">Medium</SelectItem>
                <SelectItem value="high">High</SelectItem>
                <SelectItem value="critical">Critical</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Merchant List */}
          <ScrollArea className="h-[400px] border rounded-lg p-4">
            {loading ? (
              <div className="text-center text-muted-foreground py-8">Loading merchants...</div>
            ) : merchants.length === 0 ? (
              <div className="text-center text-muted-foreground py-8">No merchants found</div>
            ) : (
              <div className="space-y-2">
                {merchants.map((merchant) => (
                  <div
                    key={merchant.id}
                    className="flex items-center gap-3 p-2 hover:bg-muted rounded-lg"
                  >
                    <Checkbox
                      checked={selectedMerchants.has(merchant.id)}
                      onCheckedChange={() => toggleMerchantSelection(merchant.id)}
                    />
                    <div className="flex-1">
                      <div className="font-medium">{merchant.name}</div>
                      <div className="text-sm text-muted-foreground">
                        {merchant.industry || 'N/A'} • {merchant.status}
                      </div>
                    </div>
                    <Badge variant={merchant.risk_level === 'high' ? 'destructive' : 'outline'}>
                      {merchant.risk_level || 'N/A'}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </ScrollArea>
        </CardContent>
      </Card>

      {/* Operation Selection */}
      <Card>
        <CardHeader>
          <CardTitle>Operation Type</CardTitle>
          <CardDescription>Select the type of bulk operation to perform</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            {[
              { type: 'update-portfolio', label: 'Update Portfolio Type', icon: Users },
              { type: 'update-risk', label: 'Update Risk Level', icon: AlertCircle },
              { type: 'export-data', label: 'Export Data', icon: Download },
              { type: 'send-notifications', label: 'Send Notifications', icon: Upload },
              { type: 'schedule-review', label: 'Schedule Review', icon: CheckCircle2 },
              { type: 'bulk-deactivate', label: 'Bulk Deactivate', icon: XCircle },
            ].map(({ type, label, icon: Icon }) => (
              <Button
                key={type}
                variant={currentOperation === type ? 'default' : 'outline'}
                className="h-auto flex-col gap-2 p-4"
                onClick={() => handleOperationSelect(type as OperationType)}
                aria-label={`Select ${label} operation`}
                aria-pressed={currentOperation === type}
              >
                <Icon className="h-6 w-6" />
                <span className="text-sm">{label}</span>
              </Button>
            ))}
          </div>

          {/* Operation Configuration */}
          {currentOperation === 'update-portfolio' && (
            <div className="mt-6 space-y-4 p-4 border rounded-lg">
              <div>
                <Label htmlFor="portfolioType">New Portfolio Type</Label>
                <Select
                  value={operationConfig.portfolioType || 'onboarded'}
                  onValueChange={(value) =>
                    setOperationConfig((prev) => ({ ...prev, portfolioType: value }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="onboarded">Onboarded</SelectItem>
                    <SelectItem value="pending">Pending</SelectItem>
                    <SelectItem value="deactivated">Deactivated</SelectItem>
                    <SelectItem value="prospective">Prospective</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label htmlFor="portfolioReason">Reason</Label>
                <Textarea
                  id="portfolioReason"
                  placeholder="Enter reason for portfolio type change..."
                  value={operationConfig.reason || ''}
                  onChange={(e) =>
                    setOperationConfig((prev) => ({ ...prev, reason: e.target.value }))
                  }
                />
              </div>
            </div>
          )}

          {currentOperation === 'update-risk' && (
            <div className="mt-6 space-y-4 p-4 border rounded-lg">
              <div>
                <Label htmlFor="riskLevel">New Risk Level</Label>
                <Select
                  value={operationConfig.riskLevel || 'medium'}
                  onValueChange={(value) =>
                    setOperationConfig((prev) => ({ ...prev, riskLevel: value }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="low">Low Risk</SelectItem>
                    <SelectItem value="medium">Medium Risk</SelectItem>
                    <SelectItem value="high">High Risk</SelectItem>
                    <SelectItem value="critical">Critical Risk</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label htmlFor="riskNotes">Risk Assessment Notes</Label>
                <Textarea
                  id="riskNotes"
                  placeholder="Enter risk assessment details..."
                  value={operationConfig.notes || ''}
                  onChange={(e) =>
                    setOperationConfig((prev) => ({ ...prev, notes: e.target.value }))
                  }
                />
              </div>
            </div>
          )}

          {currentOperation === 'export-data' && selectedMerchants.size > 0 && (
            <div className="mt-6 p-4 border rounded-lg">
              <ExportButton
                data={async () => {
                  const selected = merchants.filter((m) => selectedMerchants.has(m.id));
                  return {
                    merchants: selected,
                    exportedAt: new Date().toISOString(),
                    total: selected.length,
                  };
                }}
                exportType="merchant"
                formats={['csv', 'json', 'excel', 'pdf']}
              />
            </div>
          )}

          {/* Operation Controls */}
          {currentOperation && selectedMerchants.size > 0 && currentOperation !== 'export-data' && (
            <div className="mt-6 flex gap-2">
              {operationState.status === 'ready' && (
                <Button onClick={handleStartOperation} aria-label="Start bulk operation">
                  <Play className="h-4 w-4 mr-2" />
                  Start Operation
                </Button>
              )}
              {operationState.status === 'running' && (
                <>
                  <Button variant="outline" onClick={handlePause} aria-label="Pause bulk operation">
                    <Pause className="h-4 w-4 mr-2" />
                    Pause
                  </Button>
                  <Button variant="destructive" onClick={handleCancel} aria-label="Cancel bulk operation">
                    <Square className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                </>
              )}
              {operationState.status === 'paused' && (
                <>
                  <Button onClick={handleResume} aria-label="Resume bulk operation">
                    <Play className="h-4 w-4 mr-2" />
                    Resume
                  </Button>
                  <Button variant="destructive" onClick={handleCancel} aria-label="Cancel bulk operation">
                    <Square className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                </>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Progress Tracking */}
      {operationState.status !== 'ready' && (
        <Card>
          <CardHeader>
            <CardTitle>Operation Progress</CardTitle>
            <CardDescription>
              Status: <Badge>{operationState.status}</Badge>
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span>Progress</span>
                <span>
                  {operationState.completed} / {operationState.total} completed
                </span>
              </div>
              <Progress value={operationState.progress} />
            </div>
            {operationState.failed > 0 && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  {operationState.failed} operations failed
                </AlertDescription>
              </Alert>
            )}
          </CardContent>
        </Card>
      )}

      {/* Operation Logs */}
      <Card>
        <CardHeader>
          <CardTitle>Operation Logs</CardTitle>
          <CardDescription>Real-time operation status and messages</CardDescription>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-[300px]">
            <div className="space-y-2">
              {operationLogs.length === 0 ? (
                <div className="text-center text-muted-foreground py-8">No logs yet</div>
              ) : (
                operationLogs.map((log, index) => (
                  <div key={index} className="flex items-start gap-2 text-sm">
                    {getLogIcon(log.level)}
                    <div className="flex-1">
                      <div className="text-muted-foreground">
                        {new Date(log.timestamp).toLocaleTimeString()}
                      </div>
                      <div>{log.message}</div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  );
}

