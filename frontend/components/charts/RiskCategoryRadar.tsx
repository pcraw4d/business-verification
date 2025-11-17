'use client';

import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import { Skeleton } from '@/components/ui/skeleton';

interface RiskCategoryData {
  category: string;
  score: number; // 0-10
}

interface RiskCategoryRadarProps {
  data: RiskCategoryData[];
  height?: number;
  width?: number;
  isLoading?: boolean;
  maxScore?: number;
}

export function RiskCategoryRadar({
  data,
  height = 400,
  width = 400,
  isLoading = false,
  maxScore = 10,
}: RiskCategoryRadarProps) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (isLoading || !svgRef.current || !data || data.length === 0) return;

    // Clear previous content
    d3.select(svgRef.current).selectAll('*').remove();

    const svg = d3.select(svgRef.current);
    const margin = { top: 40, right: 40, bottom: 40, left: 40 };
    const chartWidth = width - margin.left - margin.right;
    const chartHeight = height - margin.top - margin.bottom;
    const radius = Math.min(chartWidth, chartHeight) / 2;

    const g = svg
      .append('g')
      .attr('transform', `translate(${width / 2}, ${height / 2})`);

    // Create scales
    const angleScale = d3.scalePoint<number>().domain(data.map((_, i) => i)).range([0, 2 * Math.PI]);
    const radiusScale = d3.scaleLinear().domain([0, maxScore]).range([0, radius]);

    // Draw grid circles
    const gridLevels = [2, 4, 6, 8, 10];
    gridLevels.forEach((level) => {
      g.append('circle')
        .attr('r', radiusScale(level))
        .attr('fill', 'none')
        .attr('stroke', '#e5e7eb')
        .attr('stroke-width', 1)
        .attr('opacity', 0.5);
    });

    // Draw grid lines
    data.forEach((_, i) => {
      const angle = (i / data.length) * 2 * Math.PI - Math.PI / 2;
      g.append('line')
        .attr('x1', 0)
        .attr('y1', 0)
        .attr('x2', Math.cos(angle) * radius)
        .attr('y2', Math.sin(angle) * radius)
        .attr('stroke', '#e5e7eb')
        .attr('stroke-width', 1)
        .attr('opacity', 0.5);
    });

    // Draw data polygon
    const line = d3
      .lineRadial<RiskCategoryData>()
      .angle((_, i) => (i / data.length) * 2 * Math.PI - Math.PI / 2)
      .radius((d) => radiusScale(d.score))
      .curve(d3.curveLinearClosed);

    g.append('path')
      .datum(data)
      .attr('d', line)
      .attr('fill', '#3498db')
      .attr('fill-opacity', 0.3)
      .attr('stroke', '#3498db')
      .attr('stroke-width', 2);

    // Draw data points
    data.forEach((d, i) => {
      const angle = (i / data.length) * 2 * Math.PI - Math.PI / 2;
      const r = radiusScale(d.score);

      g.append('circle')
        .attr('cx', Math.cos(angle) * r)
        .attr('cy', Math.sin(angle) * r)
        .attr('r', 4)
        .attr('fill', '#3498db');

      // Add labels
      const labelAngle = angle;
      const labelRadius = radius + 20;
      g.append('text')
        .attr('x', Math.cos(labelAngle) * labelRadius)
        .attr('y', Math.sin(labelAngle) * labelRadius)
        .attr('text-anchor', 'middle')
        .attr('dominant-baseline', 'middle')
        .attr('font-size', '12px')
        .attr('fill', 'hsl(var(--foreground))')
        .text(d.category);
    });
  }, [data, height, width, isLoading, maxScore]);

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
    <div className="flex items-center justify-center">
      <svg ref={svgRef} width={width} height={height} />
    </div>
  );
}

