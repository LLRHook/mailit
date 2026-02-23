"use client";

import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import { MailIcon, SendIcon, AlertTriangleIcon, EyeIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { StatusBadge } from "@/components/shared/status-badge";
import { StatCard } from "@/components/shared/stat-card";
import { EmptyState } from "@/components/shared/empty-state";

interface Email {
  id: string;
  to: string;
  subject: string;
  status: string;
  created_at: string;
}

const columns: ColumnDef<Email>[] = [
  {
    accessorKey: "to",
    header: "To",
    cell: ({ row }) => (
      <span className="font-medium">{row.getValue("to")}</span>
    ),
  },
  {
    accessorKey: "subject",
    header: "Subject",
    cell: ({ row }) => (
      <span className="text-muted-foreground max-w-[300px] truncate block">
        {row.getValue("subject")}
      </span>
    ),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => <StatusBadge status={row.getValue("status")} />,
  },
  {
    accessorKey: "created_at",
    header: "Created At",
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {format(new Date(row.getValue("created_at")), "MMM d, yyyy HH:mm")}
      </span>
    ),
  },
];

export default function EmailsPage() {
  const router = useRouter();

  const { data, isLoading } = useQuery({
    queryKey: ["emails"],
    queryFn: () => api.get("/emails").then((res) => res.data),
  });

  const emails: Email[] = data?.data ?? [];

  const totalSent = emails.length;
  const delivered = emails.filter((e) => e.status === "delivered").length;
  const bounced = emails.filter((e) => e.status === "bounced").length;
  const opened = emails.filter((e) => e.status === "opened").length;
  const openRate = totalSent > 0 ? ((opened / totalSent) * 100).toFixed(1) : "0";

  return (
    <div className="space-y-6">
      <PageHeader title="Emails" description="View and manage sent emails" />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Sent"
          value={totalSent}
          icon={MailIcon}
        />
        <StatCard
          title="Delivered"
          value={delivered}
          icon={SendIcon}
        />
        <StatCard
          title="Bounced"
          value={bounced}
          icon={AlertTriangleIcon}
        />
        <StatCard
          title="Open Rate"
          value={`${openRate}%`}
          icon={EyeIcon}
        />
      </div>

      {!isLoading && emails.length === 0 ? (
        <EmptyState
          icon={MailIcon}
          title="No emails yet"
          description="Emails you send through the API will appear here."
        />
      ) : (
        <DataTable
          columns={columns}
          data={emails}
          isLoading={isLoading}
          onRowClick={(row) => router.push(`/emails/${row.id}`)}
        />
      )}
    </div>
  );
}
