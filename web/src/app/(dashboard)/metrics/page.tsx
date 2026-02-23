"use client";

import {
  MailIcon,
  SendIcon,
  EyeIcon,
  AlertTriangleIcon,
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

const emailsSentData = [
  { date: "Mon", sent: 120 },
  { date: "Tue", sent: 180 },
  { date: "Wed", sent: 150 },
  { date: "Thu", sent: 210 },
  { date: "Fri", sent: 190 },
  { date: "Sat", sent: 80 },
  { date: "Sun", sent: 60 },
];

const deliveryBreakdownData = [
  { date: "Mon", delivered: 115, bounced: 3, failed: 2 },
  { date: "Tue", delivered: 170, bounced: 5, failed: 5 },
  { date: "Wed", delivered: 142, bounced: 4, failed: 4 },
  { date: "Thu", delivered: 200, bounced: 6, failed: 4 },
  { date: "Fri", delivered: 182, bounced: 4, failed: 4 },
  { date: "Sat", delivered: 76, bounced: 2, failed: 2 },
  { date: "Sun", delivered: 56, bounced: 2, failed: 2 },
];

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

  return (
    <div className="space-y-6">
      <PageHeader
        title="Metrics"
        description="Monitor your email performance"
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Emails Sent (24h)"
          value={210}
          icon={MailIcon}
          change="+12%"
          trend="up"
        />
        <StatCard
          title="Delivery Rate"
          value={`${deliveryRate}%`}
          icon={SendIcon}
          change="+0.5%"
          trend="up"
        />
        <StatCard
          title="Open Rate"
          value="34.2%"
          icon={EyeIcon}
          change="+2.1%"
          trend="up"
        />
        <StatCard
          title="Bounce Rate"
          value={`${bounceRate}%`}
          icon={AlertTriangleIcon}
          change="-0.3%"
          trend="down"
        />
      </div>

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
    </div>
  );
}
