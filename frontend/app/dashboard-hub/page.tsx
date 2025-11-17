import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Shield,
  ChartLine,
  AlertTriangle,
  Gauge,
  ClipboardCheck,
  Store,
  TrendingUp,
  Settings,
} from 'lucide-react';
import Link from 'next/link';

const dashboards = [
  {
    category: 'Merchant Verification & Risk',
    items: [
      {
        title: 'Business Intelligence',
        description: 'Comprehensive business analytics and insights',
        href: '/dashboard',
        icon: ChartLine,
        badges: ['Enhanced'],
        features: ['Real-time analytics', 'Data visualization', 'Trend analysis'],
      },
      {
        title: 'Risk Assessment',
        description: 'Advanced risk scoring and assessment tools',
        href: '/risk-dashboard',
        icon: AlertTriangle,
        badges: ['Live'],
        features: ['Risk scoring', 'Scenario analysis', 'Risk history'],
      },
      {
        title: 'Risk Indicators',
        description: 'Real-time risk monitoring and alerts',
        href: '/risk-indicators',
        icon: Gauge,
        badges: ['New'],
        features: ['Live monitoring', 'Alert system', 'Risk trends'],
      },
    ],
  },
  {
    category: 'Compliance',
    items: [
      {
        title: 'Compliance Status',
        description: 'Track compliance across all frameworks',
        href: '/compliance',
        icon: ClipboardCheck,
        badges: ['Live'],
        features: ['FATF compliance', 'Regulatory tracking', 'Status reports'],
      },
    ],
  },
  {
    category: 'Merchant Management',
    items: [
      {
        title: 'Merchant Portfolio',
        description: 'Manage and view all merchants',
        href: '/merchant-portfolio',
        icon: Store,
        badges: ['Live'],
        features: ['Portfolio overview', 'Merchant search', 'Bulk operations'],
      },
    ],
  },
  {
    category: 'Administration',
    items: [
      {
        title: 'Admin Dashboard',
        description: 'System administration and monitoring',
        href: '/admin',
        icon: Settings,
        badges: ['Beta'],
        features: ['System metrics', 'User management', 'Configuration'],
      },
    ],
  },
];

export default function DashboardHubPage() {
  return (
    <AppLayout
      title="Dashboard Hub"
      description="Comprehensive business intelligence, risk assessment, and compliance management tools"
      breadcrumbs={[
        { label: 'Home', href: '/' },
        { label: 'Dashboard Hub' },
      ]}
    >
      <div className="space-y-8">
        {/* Hero Section */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3 mb-2">
              <Shield className="h-8 w-8 text-primary" />
              <div>
                <h1 className="text-3xl font-semibold">KYB Platform Dashboard Hub</h1>
                <CardDescription className="mt-2">
                  Comprehensive business intelligence, risk assessment, and compliance management tools for modern enterprises.
                </CardDescription>
              </div>
            </div>
            <div className="flex gap-2 mt-4">
              <Badge>Live</Badge>
              <Badge variant="secondary">Enhanced</Badge>
              <Badge variant="outline">Beta Testing</Badge>
            </div>
          </CardHeader>
        </Card>

        {/* Statistics */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Total Merchants</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">0</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Risk Assessments</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-orange-600">0</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Compliance Status</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-red-600">0</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Growth Rate</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-green-600">0%</div>
            </CardContent>
          </Card>
        </div>

        {/* Dashboard Grid */}
        {dashboards.map((section, sectionIndex) => (
          <div key={sectionIndex} className="space-y-4">
            <h2 className="text-2xl font-semibold" id={`section-${sectionIndex}`}>{section.category}</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {section.items.map((dashboard, itemIndex) => {
                const Icon = dashboard.icon;
                return (
                  <Card key={itemIndex} className="hover:shadow-lg transition-shadow">
                    <CardHeader>
                      <div className="flex items-start justify-between mb-2">
                        <div className="p-2 rounded-lg bg-primary/10">
                          <Icon className="h-6 w-6 text-primary" />
                        </div>
                        <div className="flex gap-1">
                          {dashboard.badges.map((badge, badgeIndex) => (
                            <Badge key={badgeIndex} variant={badge === 'New' ? 'default' : 'secondary'} className="text-xs">
                              {badge}
                            </Badge>
                          ))}
                        </div>
                      </div>
                      <CardTitle>{dashboard.title}</CardTitle>
                      <CardDescription>{dashboard.description}</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <ul className="space-y-2 mb-4 text-sm text-muted-foreground">
                        {dashboard.features.map((feature, featureIndex) => (
                          <li key={featureIndex} className="flex items-center gap-2">
                            <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                            {feature}
                          </li>
                        ))}
                      </ul>
                      <Button asChild className="w-full" aria-label={`Open ${dashboard.title} dashboard`}>
                        <Link href={dashboard.href}>Open Dashboard</Link>
                      </Button>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </AppLayout>
  );
}

