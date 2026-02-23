"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import { PlusIcon, KeyIcon, TrashIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { CopyButton } from "@/components/shared/copy-button";
import { EmptyState } from "@/components/shared/empty-state";

interface ApiKey {
  id: string;
  name: string;
  key: string;
  permission: string;
  last_used_at: string | null;
  created_at: string;
}

function maskKey(key: string): string {
  if (!key || key.length < 8) return key;
  return `${key.slice(0, 8)}${"*".repeat(24)}`;
}

export default function ApiKeysPage() {
  const queryClient = useQueryClient();
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [keyDialogOpen, setKeyDialogOpen] = useState(false);
  const [newKeyValue, setNewKeyValue] = useState("");
  const [name, setName] = useState("");
  const [permission, setPermission] = useState("full_access");

  const { data, isLoading } = useQuery({
    queryKey: ["api-keys"],
    queryFn: () => api.get("/api-keys").then((res) => res.data),
  });

  const apiKeys: ApiKey[] = data?.data ?? [];

  const createMutation = useMutation({
    mutationFn: (payload: { name: string; permission: string }) =>
      api.post("/api-keys", payload).then((res) => res.data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      setCreateDialogOpen(false);
      setNewKeyValue(data.data.key);
      setKeyDialogOpen(true);
      setName("");
      setPermission("full_access");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      api.delete(`/api-keys/${id}`).then((res) => res.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
    },
  });

  const columns: ColumnDef<ApiKey>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue("name")}</span>
      ),
    },
    {
      accessorKey: "key",
      header: "Key",
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <span className="font-mono text-sm text-muted-foreground">
            {maskKey(row.getValue("key"))}
          </span>
          <CopyButton value={row.getValue("key")} />
        </div>
      ),
    },
    {
      accessorKey: "permission",
      header: "Permission",
      cell: ({ row }) => (
        <Badge variant="secondary" className="capitalize">
          {(row.getValue("permission") as string).replace("_", " ")}
        </Badge>
      ),
    },
    {
      accessorKey: "last_used_at",
      header: "Last Used",
      cell: ({ row }) => {
        const lastUsed = row.getValue("last_used_at") as string | null;
        return (
          <span className="text-muted-foreground">
            {lastUsed
              ? format(new Date(lastUsed), "MMM d, yyyy HH:mm")
              : "Never"}
          </span>
        );
      },
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
    {
      id: "actions",
      cell: ({ row }) => (
        <Button
          variant="ghost"
          size="icon-xs"
          className="text-muted-foreground hover:text-destructive"
          onClick={(e) => {
            e.stopPropagation();
            deleteMutation.mutate(row.original.id);
          }}
        >
          <TrashIcon className="size-3.5" />
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="API Keys"
        description="Manage your API keys for programmatic access"
      >
        <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <PlusIcon className="mr-2 size-4" />
              Create API Key
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create API Key</DialogTitle>
              <DialogDescription>
                Create a new API key to authenticate requests.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-2">
              <div className="space-y-2">
                <Label htmlFor="key-name">Name</Label>
                <Input
                  id="key-name"
                  placeholder="Production key"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Permission</Label>
                <Select value={permission} onValueChange={setPermission}>
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="full_access">Full Access</SelectItem>
                    <SelectItem value="sending_access">Sending Only</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button
                onClick={() => createMutation.mutate({ name, permission })}
                disabled={!name || createMutation.isPending}
              >
                {createMutation.isPending ? "Creating..." : "Create Key"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </PageHeader>

      <Dialog open={keyDialogOpen} onOpenChange={setKeyDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>API Key Created</DialogTitle>
            <DialogDescription>
              Copy your API key now. You will not be able to see it again.
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/50 p-3">
            <code className="flex-1 text-sm font-mono break-all">
              {newKeyValue}
            </code>
            <CopyButton value={newKeyValue} />
          </div>
          <DialogFooter>
            <Button
              onClick={() => {
                setKeyDialogOpen(false);
                setNewKeyValue("");
              }}
            >
              Done
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {!isLoading && apiKeys.length === 0 ? (
        <EmptyState
          icon={KeyIcon}
          title="No API keys"
          description="Create an API key to authenticate your requests."
          actionLabel="Create API Key"
          onAction={() => setCreateDialogOpen(true)}
        />
      ) : (
        <DataTable columns={columns} data={apiKeys} isLoading={isLoading} />
      )}
    </div>
  );
}
