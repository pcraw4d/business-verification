import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { CheckSquare } from 'lucide-react';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Progress Tracking | Compliance',
  description: 'Track compliance progress and milestones',
};

export default function ComplianceProgressTrackingPage() {
  return (
    <AppLayout
      title="Progress Tracking"
      description="Track compliance progress and milestones"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Compliance', href: '/compliance' },
        { label: 'Progress Tracking' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <CheckSquare className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Compliance Progress Tracking</CardTitle>
                <CardDescription>Track compliance progress and milestones</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium">Overall Compliance</span>
                  <span className="text-sm text-muted-foreground">0%</span>
                </div>
                <Progress value={0} />
              </div>
              <div className="text-muted-foreground">
                Progress tracking details will be displayed here
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

