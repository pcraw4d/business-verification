import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Target } from 'lucide-react';

export default function GapTrackingPage() {
  return (
    <AppLayout
      title="Gap Tracking"
      description="Track gap analysis progress and milestones"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Compliance', href: '/compliance' },
        { label: 'Gap Tracking' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Target className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Gap Tracking System</CardTitle>
                <CardDescription>Track gap analysis progress and milestones</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium">Gap Resolution Progress</span>
                  <span className="text-sm text-muted-foreground">0%</span>
                </div>
                <Progress value={0} />
              </div>
              <div className="text-muted-foreground">
                Gap tracking details will be displayed here
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

