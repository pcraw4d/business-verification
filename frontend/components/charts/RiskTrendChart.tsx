'use client';

import {
  LineChart as RechartsLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Area,
  AreaChart as RechartsAreaChart,
} from 'recharts';
import { Skeleton } from '@/components/ui/skeleton';

interface RiskTrendDataPoint {
  name: string;
  historical: number;
  prediction?: number;
  confidenceUpper?: number;
  confidenceLower?: number;
}

interface RiskTrendChartProps {
  data: RiskTrendDataPoint[];
  height?: number;
  isLoading?: boolean;
  showPrediction?: boolean;
  showConfidenceBands?: boolean;
}

export function RiskTrendChart({
  data,
  height = 300,
  isLoading = false,
  showPrediction = true,
  showConfidenceBands = true,
}: RiskTrendChartProps) {
  if (isLoading) {
    return <Skeleton className="w-full" style={{ height: `${height}px` }} />;
  }

  if (!data || data.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground" style={{ height: `${height}px` }}>
        No data available
      </div>
    );
  }

  if (showConfidenceBands && showPrediction && data.some((d) => d.confidenceUpper && d.confidenceLower)) {
    // Use AreaChart for confidence bands
    return (
      <ResponsiveContainer width="100%" height={height}>
        <RechartsAreaChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          <defs>
            <linearGradient id="confidenceGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
          <XAxis className="text-xs" dataKey="name" />
          <YAxis className="text-xs" domain={[0, 10]} />
          <Tooltip
            contentStyle={{
              backgroundColor: 'hsl(var(--popover))',
              border: '1px solid hsl(var(--border))',
              borderRadius: '6px',
            }}
          />
          <Legend />
          <Area
            type="monotone"
            dataKey="confidenceUpper"
            stroke="none"
            fill="url(#confidenceGradient)"
            fillOpacity={0.3}
          />
          <Area
            type="monotone"
            dataKey="confidenceLower"
            stroke="none"
            fill="url(#confidenceGradient)"
            fillOpacity={0.3}
          />
          <Line
            type="monotone"
            dataKey="historical"
            stroke="#3498db"
            strokeWidth={3}
            dot={{ r: 4 }}
            activeDot={{ r: 6 }}
            name="Historical Risk Score"
          />
          {showPrediction && (
            <Line
              type="monotone"
              dataKey="prediction"
              stroke="#e74c3c"
              strokeWidth={2}
              strokeDasharray="5 5"
              dot={{ r: 4 }}
              activeDot={{ r: 6 }}
              name="Prediction"
            />
          )}
        </RechartsAreaChart>
      </ResponsiveContainer>
    );
  }

  // Use LineChart for simple trend
  return (
    <ResponsiveContainer width="100%" height={height}>
      <RechartsLineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis className="text-xs" dataKey="name" />
        <YAxis className="text-xs" domain={[0, 10]} />
        <Tooltip
          contentStyle={{
            backgroundColor: 'hsl(var(--popover))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '6px',
          }}
        />
        <Legend />
        <Line
          type="monotone"
          dataKey="historical"
          stroke="#3498db"
          strokeWidth={3}
          dot={{ r: 4 }}
          activeDot={{ r: 6 }}
          name="Historical Risk Score"
        />
        {showPrediction && (
          <Line
            type="monotone"
            dataKey="prediction"
            stroke="#e74c3c"
            strokeWidth={2}
            strokeDasharray="5 5"
            dot={{ r: 4 }}
            activeDot={{ r: 6 }}
            name="Prediction"
          />
        )}
      </RechartsLineChart>
    </ResponsiveContainer>
  );
}

