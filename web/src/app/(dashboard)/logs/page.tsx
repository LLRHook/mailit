"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import { ScrollTextIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { EmptyState } from "@/components/shared/empty-state";

interface LogEntry {
  id: string;
  timestamp: string;
  level: string;
  method: string;
  path: string;
  status_code: number;
  duration_ms: number;
}

const levelColors: Record<string, string> = {
  debug: "bg-zinc-500/15 text-zinc-400 border-zinc-500/20",
  info: "bg-blue-500/15 text-blue-400 border-blue-500/20",
  warn: "bg-yellow-500/15 text-yellow-400 border-yellow-500/20",
  error: "bg-red-500/15 text-red-400 border-red-500/20",
};

const statusColor = (code: number) => {
  if (code >= 200 && code < 300) return "text-emerald-400";
  if (code >= 400 && code < 500) return "text-yellow-400";
  if (code >= 500) return "text-red-400";
  return "text-muted-foreground";
};

const columns: ColumnDef<LogEntry>[] = [
  {
    accessorKey: "timestamp",
    header: "Timestamp",
    cell: ({ row }) => (
      <span className="text-muted-foreground font-mono text-xs">
        {format(new Date(row.getValue("timestamp")), "MMM d HH:mm:ss.SSS")}
      </span>
    ),
  },
  {
    accessorKey: "level",
    header: "Level",
    cell: ({ row }) => {
      const level = row.getValue("level") as string;
      return (
        <Badge
          variant="outline"
          className={levelColors[level] ?? levelColors.info}
        >
          {level.toUpperCase()}
        </Badge>
      );
    },
  },
  {
    accessorKey: "method",
    header: "Method",
    cell: ({ row }) => (
      <span className="font-mono text-sm font-medium">
        {row.getValue("method")}
      </span>
    ),
  },
  {
    accessorKey: "path",
    header: "Path",
    cell: ({ row }) => (
      <span className="font-mono text-sm text-muted-foreground max-w-[300px] truncate block">
        {row.getValue("path")}
      </span>
    ),
  },
  {
    accessorKey: "status_code",
    header: "Status",
    cell: ({ row }) => {
      const code = row.getValue("status_code") as number;
      return (
        <span className={`font-mono text-sm font-medium ${statusColor(code)}`}>
          {code}
        </span>
      );
    },
  },
  {
    accessorKey: "duration_ms",
    header: "Duration",
    cell: ({ row }) => (
      <span className="text-muted-foreground font-mono text-sm tabular-nums">
        {row.getValue("duration_ms")}ms
      </span>
    ),
  },
];

export default function LogsPage() {
  const [level, setLevel] = useState("all");

  const { data, isLoading } = useQuery({
    queryKey: ["logs", level],
    queryFn: () =>
      api
        .get("/logs", { params: level !== "all" ? { level } : {} })
        .then((res) => res.data),
  });

  const logs: LogEntry[] = data?.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader title="API Logs" description="View API request logs">
        <Select value={level} onValueChange={setLevel}>
          <SelectTrigger className="w-[140px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Levels</SelectItem>
            <SelectItem value="debug">Debug</SelectItem>
            <SelectItem value="info">Info</SelectItem>
            <SelectItem value="warn">Warn</SelectItem>
            <SelectItem value="error">Error</SelectItem>
          </SelectContent>
        </Select>
      </PageHeader>

      {!isLoading && logs.length === 0 ? (
        <EmptyState
          icon={ScrollTextIcon}
          title="No logs yet"
          description="API request logs will appear here as they are generated."
        />
      ) : (
        <DataTable columns={columns} data={logs} isLoading={isLoading} />
      )}
    </div>
  );
}
