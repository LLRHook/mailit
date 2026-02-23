"use client";

import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import { PlusIcon, WebhookIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { EmptyState } from "@/components/shared/empty-state";

interface Webhook {
  id: string;
  url: string;
  events: string[];
  active: boolean;
  created_at: string;
}

const columns: ColumnDef<Webhook>[] = [
  {
    accessorKey: "url",
    header: "URL",
    cell: ({ row }) => (
      <span className="font-mono text-sm max-w-[300px] truncate block">
        {row.getValue("url")}
      </span>
    ),
  },
  {
    accessorKey: "events",
    header: "Events",
    cell: ({ row }) => {
      const events = row.getValue("events") as string[];
      return (
        <div className="flex flex-wrap gap-1">
          {events.map((event) => (
            <Badge key={event} variant="secondary" className="text-xs">
              {event}
            </Badge>
          ))}
        </div>
      );
    },
  },
  {
    accessorKey: "active",
    header: "Active",
    cell: ({ row }) => (
      <span
        className={
          row.getValue("active") ? "text-emerald-400" : "text-muted-foreground"
        }
      >
        {row.getValue("active") ? "Active" : "Inactive"}
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

export default function WebhooksPage() {
  const router = useRouter();

  const { data, isLoading } = useQuery({
    queryKey: ["webhooks"],
    queryFn: () => api.get("/webhooks").then((res) => res.data),
  });

  const webhooks: Webhook[] = data?.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Webhooks"
        description="Receive real-time notifications for email events"
      >
        <Button onClick={() => router.push("/webhooks/new")}>
          <PlusIcon className="mr-2 size-4" />
          Add Webhook
        </Button>
      </PageHeader>

      {!isLoading && webhooks.length === 0 ? (
        <EmptyState
          icon={WebhookIcon}
          title="No webhooks yet"
          description="Add a webhook endpoint to receive event notifications."
          actionLabel="Add Webhook"
          onAction={() => router.push("/webhooks/new")}
        />
      ) : (
        <DataTable
          columns={columns}
          data={webhooks}
          isLoading={isLoading}
        />
      )}
    </div>
  );
}
