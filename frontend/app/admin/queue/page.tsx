import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { DataTable, type Column } from '@/components/dashboards/DataTable';
import { ListTodo } from 'lucide-react';

interface QueueItem {
  id: string;
  task: string;
  status: string;
  priority: string;
  createdAt: string;
}

export default function AdminQueuePage() {
  const columns: Column<QueueItem>[] = [
    { key: 'task', header: 'Task', sortable: true },
    { key: 'status', header: 'Status', sortable: true },
    { key: 'priority', header: 'Priority', sortable: true },
    { key: 'createdAt', header: 'Created At', sortable: true },
  ];

  return (
    <AppLayout
      title="Queue Management"
      description="Task queue management and monitoring"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Admin', href: '/admin' },
        { label: 'Queue Management' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <ListTodo className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Queue Management</CardTitle>
                <CardDescription>Task queue management and monitoring</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <DataTable
              data={[]}
              columns={columns}
              searchable
              pagination={{ pageSize: 10 }}
              emptyMessage="No queue items available"
            />
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

