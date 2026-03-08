"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import {
  MailIcon,
  SendIcon,
  EyeIcon,
  AlertTriangleIcon,
  UsersIcon,
  GlobeIcon,
  KeyIcon,
  WebhookIcon,
  ArrowRightIcon,
} from "lucide-react";
import api from "@/lib/api";
import { PageHeader } from "@/components/shared/page-header";
import { StatCard } from "@/components/shared/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface MetricsTotals {
  sent: number;
  delivered: number;
  bounced: number;
  delivery_rate: number;
  open_rate: number;
  bounce_rate: number;
}

interface MetricsResponse {
  totals: MetricsTotals;
}

interface UsageResponse {
  emails_sent_today: number;
  emails_sent_month: number;
  domains: number;
  api_keys: number;
  webhooks: number;
  contacts: number;
}

const quickLinks = [
  { title: "Domains", href: "/domains", icon: GlobeIcon, description: "Manage verified domains" },
  { title: "API Keys", href: "/api-keys", icon: KeyIcon, description: "Manage API access" },
  { title: "Audience", href: "/audience", icon: UsersIcon, description: "Contacts & segments" },
  { title: "Webhooks", href: "/webhooks", icon: WebhookIcon, description: "Event notifications" },
];

export default function DashboardHome() {
  const { data: metricsData, isLoading: metricsLoading } = useQuery<MetricsResponse>({
    queryKey: ["metrics", "7d"],
    queryFn: () => api.get("/metrics?period=7d").then((res) => res.data),
  });

  const { data: usageData, isLoading: usageLoading } = useQuery<UsageResponse>({
    queryKey: ["settings-usage"],
    queryFn: () => api.get("/settings/usage").then((res) => res.data),
  });

  const totals = metricsData?.totals;
  const loading = metricsLoading || usageLoading;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description="Overview of your email sending activity"
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Emails Sent (7d)"
          value={loading ? "..." : (totals?.sent ?? 0)}
          icon={MailIcon}
        />
        <StatCard
          title="Delivery Rate"
          value={loading ? "..." : `${(totals?.delivery_rate ?? 0).toFixed(1)}%`}
          icon={SendIcon}
        />
        <StatCard
          title="Open Rate"
          value={loading ? "..." : `${(totals?.open_rate ?? 0).toFixed(1)}%`}
          icon={EyeIcon}
        />
        <StatCard
          title="Bounce Rate"
          value={loading ? "..." : `${(totals?.bounce_rate ?? 0).toFixed(1)}%`}
          icon={AlertTriangleIcon}
        />
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Today</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold tabular-nums">
              {usageLoading ? "..." : (usageData?.emails_sent_today ?? 0).toLocaleString()}
            </p>
            <p className="text-sm text-muted-foreground mt-1">emails sent today</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>This Month</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold tabular-nums">
              {usageLoading ? "..." : (usageData?.emails_sent_month ?? 0).toLocaleString()}
            </p>
            <p className="text-sm text-muted-foreground mt-1">emails sent this month</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Contacts</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold tabular-nums">
              {usageLoading ? "..." : (usageData?.contacts ?? 0).toLocaleString()}
            </p>
            <p className="text-sm text-muted-foreground mt-1">total contacts</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Quick Links</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {quickLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="flex items-center gap-3 rounded-lg border border-border p-3 hover:bg-muted/50 transition-colors"
              >
                <div className="rounded-full bg-primary/10 p-2">
                  <link.icon className="h-4 w-4 text-primary" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-medium">{link.title}</p>
                  <p className="text-xs text-muted-foreground">{link.description}</p>
                </div>
                <ArrowRightIcon className="h-4 w-4 text-muted-foreground" />
              </Link>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
