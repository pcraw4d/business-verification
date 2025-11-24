"use client";

import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import {
  AlertTriangle,
  BarChart3,
  ChartLine,
  CheckSquare,
  ClipboardCheck,
  Gauge,
  Home,
  LayoutDashboard,
  Menu,
  Network,
  PlusCircle,
  Search,
  Settings,
  Shield,
  Store,
  TrendingUp,
  Users,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

interface NavItem {
  href: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  badge?: string;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

const navigation: NavSection[] = [
  {
    title: "Platform",
    items: [
      { href: "/", label: "Home", icon: Home },
      { href: "/dashboard-hub", label: "Dashboard Hub", icon: LayoutDashboard },
    ],
  },
  {
    title: "Merchant Verification & Risk",
    items: [
      {
        href: "/add-merchant",
        label: "Add Merchant",
        icon: PlusCircle,
        badge: "NEW",
      },
      { href: "/dashboard", label: "Business Intelligence", icon: ChartLine },
      {
        href: "/risk-dashboard",
        label: "Risk Assessment",
        icon: AlertTriangle,
      },
      { href: "/risk-indicators", label: "Risk Indicators", icon: Gauge },
    ],
  },
  {
    title: "Compliance",
    items: [
      { href: "/compliance", label: "Compliance Status", icon: ClipboardCheck },
      {
        href: "/compliance/gap-analysis",
        label: "Gap Analysis",
        icon: Search,
        badge: "NEW",
      },
      {
        href: "/compliance/progress-tracking",
        label: "Progress Tracking",
        icon: CheckSquare,
      },
    ],
  },
  {
    title: "Merchant Management",
    items: [
      {
        href: "/merchant-hub",
        label: "Merchant Hub",
        icon: Network,
        badge: "NEW",
      },
      { href: "/merchant-portfolio", label: "Merchant Portfolio", icon: Store },
      {
        href: "/risk-assessment/portfolio",
        label: "Risk Assessment Portfolio",
        icon: Shield,
      },
    ],
  },
  {
    title: "Market Intelligence",
    items: [
      { href: "/market-analysis", label: "Market Analysis", icon: TrendingUp },
      {
        href: "/competitive-analysis",
        label: "Competitive Analysis",
        icon: BarChart3,
      },
    ],
  },
  {
    title: "Administration",
    items: [
      { href: "/admin", label: "Admin Dashboard", icon: Settings },
      { href: "/sessions", label: "Sessions", icon: Users },
    ],
  },
];

interface SidebarProps {
  className?: string;
  mobileOpen?: boolean;
  onMobileClose?: () => void;
}

// SidebarContent component - moved outside to prevent recreation during render
function SidebarContent({
  pathname,
  onMobileClose,
}: {
  pathname: string | null;
  onMobileClose?: () => void;
}) {
  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="flex items-center gap-2 p-4 border-b flex-shrink-0">
        <Shield className="h-6 w-6 text-primary" />
        <span className="font-bold text-lg">KYB Platform</span>
      </div>

      <ScrollArea className="flex-1 overflow-hidden">
        <nav className="p-4 space-y-6" aria-label="Main navigation">
          <h2 className="sr-only">Navigation Menu</h2>
          {navigation.map((section, sectionIndex) => (
            <div key={sectionIndex}>
              <h3
                className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 px-2"
                id={`nav-section-${sectionIndex}`}
              >
                {section.title}
              </h3>
              <ul className="space-y-1">
                {section.items.map((item) => {
                  const isActive =
                    pathname === item.href ||
                    pathname?.startsWith(item.href + "/");
                  const Icon = item.icon;

                  return (
                    <li key={item.href}>
                      <Link
                        href={item.href}
                        onClick={onMobileClose}
                        className={cn(
                          "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors",
                          "whitespace-nowrap",
                          isActive
                            ? "bg-primary text-primary-foreground"
                            : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        )}
                      >
                        <Icon className="h-4 w-4 flex-shrink-0" />
                        <span className="flex-1 min-w-0 truncate">
                          {item.label}
                        </span>
                        {item.badge && (
                          <span className="px-2 py-0.5 text-xs font-semibold bg-primary/20 text-primary rounded-full">
                            {item.badge}
                          </span>
                        )}
                      </Link>
                    </li>
                  );
                })}
              </ul>
              {sectionIndex < navigation.length - 1 && (
                <Separator className="mt-4" />
              )}
            </div>
          ))}
        </nav>
      </ScrollArea>
    </div>
  );
}

export function Sidebar({
  className,
  mobileOpen,
  onMobileClose,
}: SidebarProps) {
  const pathname = usePathname();

  return (
    <>
      {/* Desktop Sidebar - Always visible on desktop (md and above) */}
      <aside
        className={cn(
          "hidden md:flex md:w-64 md:flex-col md:fixed md:inset-y-0 md:left-0 md:z-50 md:border-r bg-background",
          className
        )}
        aria-label="Main navigation"
        style={{ display: 'none' }}
        suppressHydrationWarning
      >
        <SidebarContent pathname={pathname} onMobileClose={onMobileClose} />
      </aside>

      {/* Mobile Sidebar */}
      <Sheet
        open={mobileOpen}
        onOpenChange={(open) => !open && onMobileClose?.()}
      >
        <SheetContent side="left" className="w-64 p-0">
          <SidebarContent pathname={pathname} onMobileClose={onMobileClose} />
        </SheetContent>
      </Sheet>
    </>
  );
}

export function SidebarTrigger({ onOpen }: { onOpen: () => void }) {
  return (
    <Button
      variant="ghost"
      size="icon"
      className="md:hidden"
      onClick={onOpen}
      aria-label="Toggle sidebar"
    >
      <Menu className="h-5 w-5" />
      <span className="sr-only">Toggle sidebar</span>
    </Button>
  );
}
