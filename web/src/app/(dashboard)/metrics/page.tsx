"use client";

import {
  MailIcon,
  SendIcon,
  EyeIcon,
  AlertTriangleIcon,
  BarChart3Icon,
} from "lucide-react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  XAxis,
  YAxis,
} from "recharts";
import { PageHeader } from "@/components/shared/page-header";
import { StatCard } from "@/components/shared/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";

// TODO: Replace with real API data from /metrics endpoint
const emailsSentData: { date: string; sent: number }[] = [];

const deliveryBreakdownData: {
  date: string;
  delivered: number;
  bounced: number;
  failed: number;
}[] = [];

const sentChartConfig: ChartConfig = {
  sent: {
    label: "Emails Sent",
    color: "var(--chart-1)",
  },
};

const deliveryChartConfig: ChartConfig = {
  delivered: {
    label: "Delivered",
    color: "var(--chart-1)",
  },
  bounced: {
    label: "Bounced",
    color: "var(--chart-4)",
  },
  failed: {
    label: "Failed",
    theme: { light: "#ef4444", dark: "#ef4444" },
  },
};

export default function MetricsPage() {
  const totalSent = emailsSentData.reduce((sum, d) => sum + d.sent, 0);
  const totalDelivered = deliveryBreakdownData.reduce(
    (sum, d) => sum + d.delivered,
    0
  );
  const totalBounced = deliveryBreakdownData.reduce(
    (sum, d) => sum + d.bounced,
    0
  );
  const deliveryRate =
    totalSent > 0 ? ((totalDelivered / totalSent) * 100).toFixed(1) : "0";
  const bounceRate =
    totalSent > 0 ? ((totalBounced / totalSent) * 100).toFixed(1) : "0";

  const hasData = emailsSentData.length > 0;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Metrics"
        description="Monitor your email performance"
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Emails Sent (24h)"
          value={totalSent}
          icon={MailIcon}
        />
        <StatCard
          title="Delivery Rate"
          value={`${deliveryRate}%`}
          icon={SendIcon}
        />
        <StatCard
          title="Open Rate"
          value="0%"
          icon={EyeIcon}
        />
        <StatCard
          title="Bounce Rate"
          value={`${bounceRate}%`}
          icon={AlertTriangleIcon}
        />
      </div>

      {hasData ? (
        <>
          <Card>
            <CardHeader>
              <CardTitle>Emails Sent (Last 7 Days)</CardTitle>
            </CardHeader>
            <CardContent>
              <ChartContainer config={sentChartConfig} className="h-[300px] w-full">
                <AreaChart data={emailsSentData}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} />
                  <XAxis dataKey="date" tickLine={false} axisLine={false} />
                  <YAxis tickLine={false} axisLine={false} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Area
                    type="monotone"
                    dataKey="sent"
                    stroke="var(--color-sent)"
                    fill="var(--color-sent)"
                    fillOpacity={0.15}
                    strokeWidth={2}
                  />
                </AreaChart>
              </ChartContainer>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Delivery Breakdown (Last 7 Days)</CardTitle>
            </CardHeader>
            <CardContent>
              <ChartContainer
                config={deliveryChartConfig}
                className="h-[300px] w-full"
              >
                <BarChart data={deliveryBreakdownData}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} />
                  <XAxis dataKey="date" tickLine={false} axisLine={false} />
                  <YAxis tickLine={false} axisLine={false} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Bar
                    dataKey="delivered"
                    fill="var(--color-delivered)"
                    radius={[4, 4, 0, 0]}
                  />
                  <Bar
                    dataKey="bounced"
                    fill="var(--color-bounced)"
                    radius={[4, 4, 0, 0]}
                  />
                  <Bar
                    dataKey="failed"
                    fill="var(--color-failed)"
                    radius={[4, 4, 0, 0]}
                  />
                </BarChart>
              </ChartContainer>
            </CardContent>
          </Card>
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <BarChart3Icon className="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 className="text-lg font-semibold">No metrics yet</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Metrics will appear here once you start sending emails.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
