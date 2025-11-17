'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { LucideIcon } from 'lucide-react';
import Link from 'next/link';

interface DashboardCardProps {
  title: string;
  description: string;
  href: string;
  icon: LucideIcon;
  badges?: string[];
  features?: string[];
  className?: string;
  variant?: 'default' | 'core' | 'compliance' | 'intelligence' | 'merchant';
}

const variantStyles = {
  default: 'border-l-4 border-l-primary',
  core: 'border-l-4 border-l-blue-500',
  compliance: 'border-l-4 border-l-red-500',
  intelligence: 'border-l-4 border-l-purple-500',
  merchant: 'border-l-4 border-l-indigo-500',
};

export function DashboardCard({
  title,
  description,
  href,
  icon: Icon,
  badges = [],
  features = [],
  className,
  variant = 'default',
}: DashboardCardProps) {
  return (
    <Card className={cn('hover:shadow-lg transition-shadow', variantStyles[variant], className)}>
      <CardHeader>
        <div className="flex items-start justify-between mb-2">
          <div className={cn(
            'p-2 rounded-lg',
            variant === 'core' && 'bg-blue-500/10',
            variant === 'compliance' && 'bg-red-500/10',
            variant === 'intelligence' && 'bg-purple-500/10',
            variant === 'merchant' && 'bg-indigo-500/10',
            variant === 'default' && 'bg-primary/10'
          )}>
            <Icon className={cn(
              'h-6 w-6',
              variant === 'core' && 'text-blue-500',
              variant === 'compliance' && 'text-red-500',
              variant === 'intelligence' && 'text-purple-500',
              variant === 'merchant' && 'text-indigo-500',
              variant === 'default' && 'text-primary'
            )} />
          </div>
          <div className="flex gap-1">
            {badges.map((badge, index) => (
              <Badge
                key={index}
                variant={badge === 'New' ? 'default' : badge === 'Enhanced' ? 'secondary' : 'outline'}
                className="text-xs"
              >
                {badge}
              </Badge>
            ))}
          </div>
        </div>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      {features.length > 0 && (
        <CardContent>
          <ul className="space-y-2 mb-4 text-sm text-muted-foreground">
            {features.map((feature, index) => (
              <li key={index} className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                {feature}
              </li>
            ))}
          </ul>
          <Button asChild className="w-full" aria-label={`Open ${title} dashboard`}>
            <Link href={href}>Open Dashboard</Link>
          </Button>
        </CardContent>
      )}
    </Card>
  );
}

