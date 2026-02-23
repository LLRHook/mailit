"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import { PlusIcon, GlobeIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { StatusBadge } from "@/components/shared/status-badge";
import { EmptyState } from "@/components/shared/empty-state";

interface Domain {
  id: string;
  name: string;
  status: string;
  dns_status: string;
  created_at: string;
}

const columns: ColumnDef<Domain>[] = [
  {
    accessorKey: "name",
    header: "Domain",
    cell: ({ row }) => (
      <span className="font-medium">{row.getValue("name")}</span>
    ),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => <StatusBadge status={row.getValue("status")} />,
  },
  {
    accessorKey: "dns_status",
    header: "DNS Records",
    cell: ({ row }) => <StatusBadge status={row.getValue("dns_status")} />,
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

export default function DomainsPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [domainName, setDomainName] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["domains"],
    queryFn: () => api.get("/domains").then((res) => res.data),
  });

  const domains: Domain[] = data?.data ?? [];

  const createMutation = useMutation({
    mutationFn: (payload: { name: string }) =>
      api.post("/domains", payload).then((res) => res.data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["domains"] });
      setDialogOpen(false);
      setDomainName("");
      router.push(`/domains/${data.data.id}`);
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Domains" description="Manage your sending domains">
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <PlusIcon className="mr-2 size-4" />
              Add Domain
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Domain</DialogTitle>
              <DialogDescription>
                Enter the domain name you want to send emails from.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-2">
              <div className="space-y-2">
                <Label htmlFor="domain-name">Domain Name</Label>
                <Input
                  id="domain-name"
                  placeholder="example.com"
                  value={domainName}
                  onChange={(e) => setDomainName(e.target.value)}
                />
              </div>
            </div>
            <DialogFooter>
              <Button
                onClick={() => createMutation.mutate({ name: domainName })}
                disabled={!domainName || createMutation.isPending}
              >
                {createMutation.isPending ? "Adding..." : "Add Domain"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </PageHeader>

      {!isLoading && domains.length === 0 ? (
        <EmptyState
          icon={GlobeIcon}
          title="No domains yet"
          description="Add a domain to start sending emails."
          actionLabel="Add Domain"
          onAction={() => setDialogOpen(true)}
        />
      ) : (
        <DataTable
          columns={columns}
          data={domains}
          isLoading={isLoading}
          onRowClick={(row) => router.push(`/domains/${row.id}`)}
        />
      )}
    </div>
  );
}
