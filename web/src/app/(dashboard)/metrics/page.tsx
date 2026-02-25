"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
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
import api from "@/lib/api";
import { PageHeader } from "@/components/shared/page-header";
import { StatCard } from "@/components/shared/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";

interface MetricsDataPoint {
  date: string;
  sent: number;
  delivered: number;
  bounced: number;
  failed: number;
  opened: number;
  clicked: number;
  complained: number;
}

interface MetricsTotals {
  sent: number;
  delivered: number;
  bounced: number;
  failed: number;
  opened: number;
  clicked: number;
  complained: number;
  delivery_rate: number;
  open_rate: number;
  bounce_rate: number;
}

interface MetricsResponse {
  period: string;
  from: string;
  to: string;
  totals: MetricsTotals;
  data: MetricsDataPoint[];
}

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
  const [period, setPeriod] = useState("7d");

  const { data: metricsData, isLoading } = useQuery<MetricsResponse>({
    queryKey: ["metrics", period],
    queryFn: () =>
      api.get(`/metrics?period=${period}`).then((res) => res.data),
  });

  const totals = metricsData?.totals;
  const chartData = metricsData?.data ?? [];
  const hasData = chartData.length > 0 && (totals?.sent ?? 0) > 0;

  const periodLabel =
    period === "24h"
      ? "Last 24 Hours"
      : period === "30d"
        ? "Last 30 Days"
        : "Last 7 Days";

  return (
    <div className="space-y-6">
      <PageHeader
        title="Metrics"
        description="Monitor your email performance"
      >
        <Select value={period} onValueChange={setPeriod}>
          <SelectTrigger className="w-[160px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="24h">Last 24 Hours</SelectItem>
            <SelectItem value="7d">Last 7 Days</SelectItem>
            <SelectItem value="30d">Last 30 Days</SelectItem>
          </SelectContent>
        </Select>
      </PageHeader>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Emails Sent"
          value={isLoading ? "..." : (totals?.sent ?? 0)}
          icon={MailIcon}
        />
        <StatCard
          title="Delivery Rate"
          value={
            isLoading ? "..." : `${(totals?.delivery_rate ?? 0).toFixed(1)}%`
          }
          icon={SendIcon}
        />
        <StatCard
          title="Open Rate"
          value={
            isLoading ? "..." : `${(totals?.open_rate ?? 0).toFixed(1)}%`
          }
          icon={EyeIcon}
        />
        <StatCard
          title="Bounce Rate"
          value={
            isLoading ? "..." : `${(totals?.bounce_rate ?? 0).toFixed(1)}%`
          }
          icon={AlertTriangleIcon}
        />
      </div>

      {hasData ? (
        <>
          <Card>
            <CardHeader>
              <CardTitle>Emails Sent ({periodLabel})</CardTitle>
            </CardHeader>
            <CardContent>
              <ChartContainer
                config={sentChartConfig}
                className="h-[300px] w-full"
              >
                <AreaChart data={chartData}>
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
              <CardTitle>Delivery Breakdown ({periodLabel})</CardTitle>
            </CardHeader>
            <CardContent>
              <ChartContainer
                config={deliveryChartConfig}
                className="h-[300px] w-full"
              >
                <BarChart data={chartData}>
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
