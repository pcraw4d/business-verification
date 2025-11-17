'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Shield, Search, TrendingUp, Zap } from 'lucide-react';
import Link from 'next/link';

export default function HomePage() {
  const router = useRouter();

  useEffect(() => {
    // Auto-redirect to merchant-portfolio after 3 seconds
    const timer = setTimeout(() => {
      router.push('/merchant-portfolio');
    }, 3000);

    return () => clearTimeout(timer);
  }, [router]);

  return (
    <AppLayout
      title="KYB Platform"
      description="Business Verification & Risk Management"
    >
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-8">
        <div className="text-center space-y-4">
          <div className="flex items-center justify-center gap-3 mb-4">
            <Shield className="h-12 w-12 text-primary" />
            <h1 className="text-5xl font-bold">KYB Platform</h1>
          </div>
          <p className="text-xl text-muted-foreground max-w-2xl">
            Comprehensive Know Your Business (KYB) platform for merchant verification,
            risk assessment, and compliance management. Streamline your onboarding process
            with advanced AI-powered business intelligence.
          </p>
          
          <div className="flex items-center justify-center gap-4 flex-wrap mt-6">
            <Badge variant="default" className="px-4 py-2">
              <span className="w-2 h-2 bg-green-500 rounded-full mr-2 inline-block animate-pulse" />
              Live
            </Badge>
            <Badge variant="secondary" className="px-4 py-2">
              Beta
            </Badge>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 w-full max-w-4xl mt-12">
          <Card>
            <CardHeader>
              <Search className="h-8 w-8 text-primary mb-2" />
              <CardTitle>Business Verification</CardTitle>
            </CardHeader>
            <CardContent>
              <CardDescription>
                Advanced merchant verification and due diligence
              </CardDescription>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <TrendingUp className="h-8 w-8 text-primary mb-2" />
              <CardTitle>Risk Assessment</CardTitle>
            </CardHeader>
            <CardContent>
              <CardDescription>
                Comprehensive risk analysis and scoring
              </CardDescription>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <Zap className="h-8 w-8 text-primary mb-2" />
              <CardTitle>Real-time Processing</CardTitle>
            </CardHeader>
            <CardContent>
              <CardDescription>
                Instant verification and decision making
              </CardDescription>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <Shield className="h-8 w-8 text-primary mb-2" />
              <CardTitle>Compliance</CardTitle>
            </CardHeader>
            <CardContent>
              <CardDescription>
                FATF and regulatory compliance tracking
              </CardDescription>
            </CardContent>
          </Card>
        </div>

        <div className="mt-8">
          <Button asChild size="lg">
            <Link href="/merchant-portfolio">
              Enter Merchant Portfolio
            </Link>
          </Button>
          <p className="text-sm text-muted-foreground mt-4 text-center">
            Redirecting automatically in 3 seconds...
          </p>
        </div>
      </div>
    </AppLayout>
  );
}
