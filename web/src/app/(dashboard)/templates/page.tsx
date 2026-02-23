"use client";

import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import { PlusIcon, FileTextIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { EmptyState } from "@/components/shared/empty-state";

interface Template {
  id: string;
  name: string;
  description: string;
  versions: number;
  created_at: string;
}

const columns: ColumnDef<Template>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => (
      <span className="font-medium">{row.getValue("name")}</span>
    ),
  },
  {
    accessorKey: "description",
    header: "Description",
    cell: ({ row }) => (
      <span className="text-muted-foreground max-w-[300px] truncate block">
        {row.getValue("description") || "No description"}
      </span>
    ),
  },
  {
    accessorKey: "versions",
    header: "Versions",
    cell: ({ row }) => (
      <span className="text-muted-foreground tabular-nums">
        {row.getValue("versions")}
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

export default function TemplatesPage() {
  const router = useRouter();

  const { data, isLoading } = useQuery({
    queryKey: ["templates"],
    queryFn: () => api.get("/templates").then((res) => res.data),
  });

  const templates: Template[] = data?.data ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Templates"
        description="Manage reusable email templates"
      >
        <Button onClick={() => router.push("/templates/new")}>
          <PlusIcon className="mr-2 size-4" />
          New Template
        </Button>
      </PageHeader>

      {!isLoading && templates.length === 0 ? (
        <EmptyState
          icon={FileTextIcon}
          title="No templates yet"
          description="Create reusable email templates to speed up your workflow."
          actionLabel="New Template"
          onAction={() => router.push("/templates/new")}
        />
      ) : (
        <DataTable
          columns={columns}
          data={templates}
          isLoading={isLoading}
          onRowClick={(row) => router.push(`/templates/${row.id}`)}
        />
      )}
    </div>
  );
}
