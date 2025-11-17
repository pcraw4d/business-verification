'use client';

import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { Skeleton } from '@/components/ui/skeleton';

interface AreaChartData {
  name: string;
  value: number;
  [key: string]: string | number;
}

interface AreaChartProps {
  data: AreaChartData[];
  dataKey: string;
  areas?: Array<{
    key: string;
    name: string;
    color?: string;
    fillOpacity?: number;
  }>;
  xAxisLabel?: string;
  yAxisLabel?: string;
  height?: number;
  isLoading?: boolean;
}

export function AreaChart({
  data,
  dataKey,
  areas = [{ key: 'value', name: 'Value', color: '#8884d8', fillOpacity: 0.6 }],
  xAxisLabel,
  yAxisLabel,
  height = 300,
  isLoading = false,
}: AreaChartProps) {
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

  return (
    <ResponsiveContainer width="100%" height={height}>
      <RechartsAreaChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
        <defs>
          {areas.map((area, index) => (
            <linearGradient key={area.key} id={`color${index}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={area.color || '#8884d8'} stopOpacity={area.fillOpacity || 0.6} />
              <stop offset="95%" stopColor={area.color || '#8884d8'} stopOpacity={0} />
            </linearGradient>
          ))}
        </defs>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis
          dataKey={dataKey}
          label={xAxisLabel ? { value: xAxisLabel, position: 'insideBottom', offset: -5 } : undefined}
          className="text-xs"
        />
        <YAxis
          label={yAxisLabel ? { value: yAxisLabel, angle: -90, position: 'insideLeft' } : undefined}
          className="text-xs"
        />
        <Tooltip
          contentStyle={{
            backgroundColor: 'hsl(var(--popover))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '6px',
          }}
        />
        <Legend />
        {areas.map((area, index) => (
          <Area
            key={area.key}
            type="monotone"
            dataKey={area.key}
            name={area.name}
            stroke={area.color || '#8884d8'}
            fill={`url(#color${index})`}
            fillOpacity={area.fillOpacity || 0.6}
          />
        ))}
      </RechartsAreaChart>
    </ResponsiveContainer>
  );
}

