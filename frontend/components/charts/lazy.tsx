/**
 * Lazy-loaded chart components for code splitting
 */

import dynamic from 'next/dynamic';
import { Skeleton } from '@/components/ui/skeleton';

// Loading component for charts
const ChartSkeleton = ({ height = 300 }: { height?: number }) => (
  <Skeleton className="w-full" style={{ height: `${height}px` }} />
);

// Lazy load chart components
export const LineChart = dynamic(
  () => import('./LineChart').then((mod) => ({ default: mod.LineChart })),
  {
    loading: () => <ChartSkeleton />,
    ssr: false, // Disable SSR for charts (they're client-side only)
  }
);

export const BarChart = dynamic(
  () => import('./BarChart').then((mod) => ({ default: mod.BarChart })),
  {
    loading: () => <ChartSkeleton />,
    ssr: false,
  }
);

export const PieChart = dynamic(
  () => import('./PieChart').then((mod) => ({ default: mod.PieChart })),
  {
    loading: () => <ChartSkeleton />,
    ssr: false,
  }
);

export const AreaChart = dynamic(
  () => import('./AreaChart').then((mod) => ({ default: mod.AreaChart })),
  {
    loading: () => <ChartSkeleton />,
    ssr: false,
  }
);

export const RiskGauge = dynamic(
  () => import('./RiskGauge').then((mod) => ({ default: mod.RiskGauge })),
  {
    loading: () => <ChartSkeleton />,
    ssr: false,
  }
);

export const RiskTrendChart = dynamic(
  () => import('./RiskTrendChart').then((mod) => ({ default: mod.RiskTrendChart })),
  {
    loading: () => <ChartSkeleton />,
    ssr: false,
  }
);

export const RiskCategoryRadar = dynamic(
  () => import('./RiskCategoryRadar').then((mod) => ({ default: mod.RiskCategoryRadar })),
  {
    loading: () => <ChartSkeleton />,
    ssr: false,
  }
);

