"use client";

import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import { PlusIcon, MegaphoneIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { StatusBadge } from "@/components/shared/status-badge";
import { EmptyState } from "@/components/shared/empty-state";

interface Broadcast {
  id: string;
  name: string;
  audience_id: string;
  audience_name: string;
  status: string;
  recipients: number;
  sent: number;
  created_at: string;
}

const columns: ColumnDef<Broadcast>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => (
      <span className="font-medium">{row.getValue("name")}</span>
    ),
  },
  {
    accessorKey: "audience_name",
    header: "Audience",
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {row.getValue("audience_name")}
      </span>
    ),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => <StatusBadge status={row.getValue("status")} />,
  },
  {
    accessorKey: "recipients",
    header: "Recipients",
    cell: ({ row }) => (
      <span className="text-muted-foreground tabular-nums">
        {(row.getValue("recipients") as number).toLocaleString()}
      </span>
    ),
  },
  {
    accessorKey: "sent",
    header: "Sent",
    cell: ({ row }) => (
      <span className="text-muted-foreground tabular-nums">
        {(row.getValue("sent") as number).toLocaleString()}
      </span>
    ),
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

export default function BroadcastsPage() {
  const router = useRouter();

  const { data, isLoading } = useQuery({
    queryKey: ["broadcasts"],
    queryFn: () => api.get("/broadcasts").then((res) => res.data),
  });

  const broadcasts: Broadcast[] = data?.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader title="Broadcasts" description="Send emails to your audiences">
        <Button onClick={() => router.push("/broadcasts/new")}>
          <PlusIcon className="mr-2 size-4" />
          New Broadcast
        </Button>
      </PageHeader>

      {!isLoading && broadcasts.length === 0 ? (
        <EmptyState
          icon={MegaphoneIcon}
          title="No broadcasts yet"
          description="Create your first broadcast to send emails to an audience."
          actionLabel="New Broadcast"
          onAction={() => router.push("/broadcasts/new")}
        />
      ) : (
        <DataTable
          columns={columns}
          data={broadcasts}
          isLoading={isLoading}
          onRowClick={(row) => router.push(`/broadcasts/${row.id}`)}
        />
      )}
    </div>
  );
}
