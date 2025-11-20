'use client';

import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import { Skeleton } from '@/components/ui/skeleton';

interface RiskGaugeProps {
  value: number; // 0-10 (risk score scale)
  min?: number;
  max?: number;
  height?: number;
  width?: number;
  isLoading?: boolean;
  label?: string;
  showNeedle?: boolean;
  colorScheme?: {
    low: string;
    medium: string;
    high: string;
    critical: string;
  };
}

export function RiskGauge({
  value,
  min = 0,
  max = 10,
  height = 300,
  width = 300,
  isLoading = false,
  label = 'Risk Score',
  showNeedle = true,
  colorScheme = {
    low: '#22c55e',
    medium: '#eab308',
    high: '#ef4444',
    critical: '#991b1b',
  },
}: RiskGaugeProps) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (isLoading || !svgRef.current) return;

    // Clear previous content
    d3.select(svgRef.current).selectAll('*').remove();

    const svg = d3.select(svgRef.current);
    const radius = Math.min(width, height) / 2 - 20;

    const g = svg
      .append('g')
      .attr('transform', `translate(${width / 2}, ${height / 2})`);

    // Create arc generators
    const backgroundArc = d3
      .arc<{ endAngle: number }>()
      .innerRadius(radius * 0.6)
      .outerRadius(radius)
      .startAngle(0)
      .endAngle((d) => d.endAngle);

    const valueArc = d3
      .arc<{ endAngle: number }>()
      .innerRadius(radius * 0.6)
      .outerRadius(radius)
      .startAngle(-Math.PI / 2)
      .endAngle((d) => d.endAngle);

    // Add background arc
    g.append('path')
      .datum({ endAngle: 2 * Math.PI })
      .attr('d', backgroundArc)
      .attr('fill', '#f0f0f0')
      .attr('stroke', '#ddd')
      .attr('stroke-width', 2);

    // Create risk level arcs
    const riskLevels = [
      { level: 'low', start: -Math.PI / 2, end: -Math.PI / 2 + Math.PI * 0.5, color: colorScheme.low },
      { level: 'medium', start: -Math.PI / 2 + Math.PI * 0.5, end: -Math.PI / 2 + Math.PI * 0.75, color: colorScheme.medium },
      { level: 'high', start: -Math.PI / 2 + Math.PI * 0.75, end: -Math.PI / 2 + Math.PI * 0.9, color: colorScheme.high },
      { level: 'critical', start: -Math.PI / 2 + Math.PI * 0.9, end: Math.PI / 2, color: colorScheme.critical },
    ];

    riskLevels.forEach((level) => {
      const arc = d3
        .arc<{ startAngle: number; endAngle: number }>()
        .innerRadius(radius * 0.6)
        .outerRadius(radius)
        .startAngle((d) => d.startAngle)
        .endAngle((d) => d.endAngle);

      g.append('path')
        .datum({ startAngle: level.start, endAngle: level.end })
        .attr('d', arc)
        .attr('fill', level.color)
        .attr('opacity', 0.3);
    });

    // Calculate angle for value (0-10 scale mapped to 0-180 degrees)
    const normalizedValue = Math.max(min, Math.min(max, value));
    const angle = -Math.PI / 2 + (normalizedValue / max) * Math.PI;

    // Create value arc
    const valuePath = g
      .append('path')
      .datum({ endAngle: -Math.PI / 2 })
      .attr('fill', 'none')
      .attr('stroke', () => {
        if (value <= 2.5) return colorScheme.low;
        if (value <= 5) return colorScheme.medium;
        if (value <= 7.5) return colorScheme.high;
        return colorScheme.critical;
      })
      .attr('stroke-width', 8)
      .attr('stroke-linecap', 'round')
      .attr('opacity', 0);

    // Animate value arc
    valuePath
      .transition()
      .duration(1000)
      .attrTween('d', function (d) {
        const interpolate = d3.interpolate(d.endAngle, angle);
        return function (t) {
          d.endAngle = interpolate(t);
          return valueArc(d) || '';
        };
      })
      .attr('opacity', 1);

    // Create needle if enabled
    if (showNeedle) {
      const needle = g.append('g').attr('class', 'needle');

      needle
        .append('line')
        .attr('x1', 0)
        .attr('y1', 0)
        .attr('x2', 0)
        .attr('y2', -radius * 0.8)
        .attr('stroke', '#2c3e50')
        .attr('stroke-width', 3)
        .attr('stroke-linecap', 'round')
        .attr('transform', `rotate(${-90})`)
        .transition()
        .duration(1000)
        .attr('transform', `rotate(${(angle * 180) / Math.PI})`);

      needle
        .append('circle')
        .attr('cx', 0)
        .attr('cy', 0)
        .attr('r', 8)
        .attr('fill', '#2c3e50');
    }

    // Add center text
    const centerText = g.append('g').attr('class', 'center-text');

    centerText
      .append('text')
      .attr('text-anchor', 'middle')
      .attr('dy', '-0.5em')
      .attr('font-size', '2em')
      .attr('font-weight', 'bold')
      .attr('fill', 'hsl(var(--foreground))')
      .text('0.0')
      .transition()
      .duration(1000)
      .tween('text', function () {
        const current = parseFloat(this.textContent || '0') || 0;
        const targetValue = value != null && !isNaN(value) ? value : 0;
        const interpolate = d3.interpolate(current, targetValue);
        return function (t) {
          const interpolatedValue = interpolate(t);
          this.textContent = (interpolatedValue != null && !isNaN(interpolatedValue)) 
            ? interpolatedValue.toFixed(1) 
            : '0.0';
        };
      });

    centerText
      .append('text')
      .attr('text-anchor', 'middle')
      .attr('dy', '1.5em')
      .attr('font-size', '0.8em')
      .attr('fill', 'hsl(var(--muted-foreground))')
      .text(label);
  }, [value, min, max, height, width, isLoading, label, showNeedle, colorScheme]);

  if (isLoading) {
    return <Skeleton className="w-full" style={{ height: `${height}px` }} />;
  }

  return (
    <div className="flex items-center justify-center">
      <svg ref={svgRef} width={width} height={height} />
    </div>
  );
}

