'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Users } from 'lucide-react';
import { ApiEndpoints } from '@/lib/api-config';
import { toast } from 'sonner';

interface Session {
  id: string;
  user_id?: string;
  userId?: string;
  ip_address?: string;
  ipAddress?: string;
  user_agent?: string;
  userAgent?: string;
  created_at?: string;
  createdAt?: string;
  startTime?: string;
  last_access_time?: string;
  lastAccessTime?: string;
  last_activity_time?: string;
  lastActivity?: string;
  lastActivityTime?: string;
  request_count?: number;
  requestCount?: number;
  is_active?: boolean;
  isActive?: boolean;
  status?: string;
  session_duration?: string;
  sessionDuration?: string;
  metadata?: Record<string, unknown>;
}

export default function SessionsPage() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [mounted, setMounted] = useState(false);

  // Client-side formatted dates to prevent hydration errors
  const [formattedDates, setFormattedDates] = useState<Record<string, { startTime: string; lastActivity: string }>>({});

  useEffect(() => {
    async function fetchSessions() {
      try {
        const token = typeof window !== 'undefined' ? sessionStorage.getItem('authToken') : null;
        
        const headers: HeadersInit = {
          'Content-Type': 'application/json',
        };
        
        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(ApiEndpoints.sessions.list(), {
          method: 'GET',
          headers,
        });

        if (response.ok) {
          const data = await response.json();
          // Handle different response formats
          let sessionsList: Session[] = [];
          
          if (Array.isArray(data)) {
            sessionsList = data;
          } else if (data.sessions && Array.isArray(data.sessions)) {
            sessionsList = data.sessions;
          } else if (data.success && Array.isArray(data.sessions)) {
            sessionsList = data.sessions;
          }
          
          // Normalize session data structure
          const normalizedSessions = sessionsList.map((session: Session) => ({
            id: session.id,
            userId: session.user_id || session.userId || '',
            startTime: session.created_at || session.createdAt || session.startTime || '',
            lastActivity: session.last_activity_time || session.lastActivityTime || 
                         session.last_access_time || session.lastAccessTime || 
                         session.lastActivity || '',
            status: session.is_active !== undefined ? (session.is_active ? 'active' : 'inactive') :
                   session.isActive !== undefined ? (session.isActive ? 'active' : 'inactive') :
                   session.status || 'unknown',
          }));
          
          setSessions(normalizedSessions);
        } else if (response.status === 404) {
          // Endpoint doesn't exist yet, use empty array
          setSessions([]);
        } else {
          throw new Error(`Failed to load sessions: ${response.statusText}`);
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to load sessions';
        toast.error('Failed to load sessions', {
          description: errorMessage,
        });
        setSessions([]);
      } finally {
        setLoading(false);
      }
    }

    fetchSessions();
    
    // Refresh sessions every 60 seconds
    const interval = setInterval(fetchSessions, 60000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Format dates on client side only to prevent hydration errors
  useEffect(() => {
    if (!mounted || sessions.length === 0) return;

    const formatted: Record<string, { startTime: string; lastActivity: string }> = {};
    sessions.forEach((session) => {
      const formattedSession: { startTime: string; lastActivity: string } = {
        startTime: 'N/A',
        lastActivity: 'N/A',
      };

      if (session.startTime) {
        try {
          formattedSession.startTime = new Date(session.startTime).toLocaleString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          });
        } catch {
          formattedSession.startTime = session.startTime;
        }
      }

      if (session.lastActivity) {
        try {
          formattedSession.lastActivity = new Date(session.lastActivity).toLocaleString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          });
        } catch {
          formattedSession.lastActivity = session.lastActivity;
        }
      }

      formatted[session.id] = formattedSession;
    });
    setFormattedDates(formatted);
  }, [mounted, sessions]);

  return (
    <AppLayout
      title="Session Management"
      description="Active session management and monitoring"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Admin', href: '/admin' },
        { label: 'Sessions' },
      ]}
    >
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Users className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>Session Management</CardTitle>
                <CardDescription>Active session management and monitoring</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="space-y-2">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-12 w-full" />
                ))}
              </div>
            ) : sessions.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <Users className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No active sessions</p>
              </div>
            ) : (
              <div className="border rounded-lg">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>User ID</TableHead>
                      <TableHead>Start Time</TableHead>
                      <TableHead>Last Activity</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {sessions.map((session) => (
                      <TableRow key={session.id}>
                        <TableCell className="font-medium">{session.userId}</TableCell>
                        <TableCell suppressHydrationWarning>
                          {mounted && formattedDates[session.id] ? formattedDates[session.id].startTime : session.startTime || 'N/A'}
                        </TableCell>
                        <TableCell suppressHydrationWarning>
                          {mounted && formattedDates[session.id] ? formattedDates[session.id].lastActivity : session.lastActivity || 'N/A'}
                        </TableCell>
                        <TableCell>
                          <Badge variant={session.status === 'active' ? 'default' : 'secondary'}>
                            {session.status}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

