"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import {
  PlusIcon,
  UsersIcon,
  TagIcon,
  FilterIcon,
  BookmarkIcon,
} from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
import { EmptyState } from "@/components/shared/empty-state";

// --- Types ---

interface Audience {
  id: string;
  name: string;
}

interface Contact {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  subscribed: boolean;
  created_at: string;
}

interface Property {
  id: string;
  name: string;
  label: string;
  type: string;
  created_at: string;
}

interface Segment {
  id: string;
  name: string;
  conditions: unknown;
  created_at: string;
}

interface Topic {
  id: string;
  name: string;
  description: string;
  created_at: string;
}

// --- Contacts Tab ---

function ContactsTab() {
  const queryClient = useQueryClient();
  const [audienceId, setAudienceId] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [email, setEmail] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");

  const { data: audiencesData } = useQuery({
    queryKey: ["audiences"],
    queryFn: () => api.get("/audiences").then((res) => res.data),
  });

  const audiences: Audience[] = audiencesData?.data ?? [];

  const { data: contactsData, isLoading } = useQuery({
    queryKey: ["contacts", audienceId],
    queryFn: () =>
      api
        .get(`/audiences/${audienceId}/contacts`)
        .then((res) => res.data),
    enabled: !!audienceId,
  });

  const contacts: Contact[] = contactsData?.data ?? [];

  const createMutation = useMutation({
    mutationFn: (payload: {
      email: string;
      first_name: string;
      last_name: string;
    }) =>
      api
        .post(`/audiences/${audienceId}/contacts`, payload)
        .then((res) => res.data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["contacts", audienceId],
      });
      setDialogOpen(false);
      setEmail("");
      setFirstName("");
      setLastName("");
    },
  });

  const contactColumns: ColumnDef<Contact>[] = [
    {
      accessorKey: "email",
      header: "Email",
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue("email")}</span>
      ),
    },
    {
      accessorKey: "first_name",
      header: "First Name",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.getValue("first_name") || "-"}
        </span>
      ),
    },
    {
      accessorKey: "last_name",
      header: "Last Name",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.getValue("last_name") || "-"}
        </span>
      ),
    },
    {
      accessorKey: "subscribed",
      header: "Subscribed",
      cell: ({ row }) => (
        <span
          className={
            row.getValue("subscribed")
              ? "text-emerald-400"
              : "text-muted-foreground"
          }
        >
          {row.getValue("subscribed") ? "Yes" : "No"}
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

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Select value={audienceId} onValueChange={setAudienceId}>
          <SelectTrigger className="w-[240px]">
            <SelectValue placeholder="Select an audience" />
          </SelectTrigger>
          <SelectContent>
            {audiences.map((audience) => (
              <SelectItem key={audience.id} value={audience.id}>
                {audience.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button disabled={!audienceId}>
              <PlusIcon className="mr-2 size-4" />
              Add Contact
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Contact</DialogTitle>
              <DialogDescription>
                Add a new contact to this audience.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-2">
              <div className="space-y-2">
                <Label htmlFor="contact-email">Email</Label>
                <Input
                  id="contact-email"
                  type="email"
                  placeholder="john@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="contact-first-name">First Name</Label>
                  <Input
                    id="contact-first-name"
                    placeholder="John"
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="contact-last-name">Last Name</Label>
                  <Input
                    id="contact-last-name"
                    placeholder="Doe"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                  />
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button
                onClick={() =>
                  createMutation.mutate({
                    email,
                    first_name: firstName,
                    last_name: lastName,
                  })
                }
                disabled={!email || createMutation.isPending}
              >
                {createMutation.isPending ? "Adding..." : "Add Contact"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {!audienceId ? (
        <EmptyState
          icon={UsersIcon}
          title="Select an audience"
          description="Choose an audience from the dropdown to see its contacts."
        />
      ) : !isLoading && contacts.length === 0 ? (
        <EmptyState
          icon={UsersIcon}
          title="No contacts"
          description="This audience has no contacts yet."
          actionLabel="Add Contact"
          onAction={() => setDialogOpen(true)}
        />
      ) : (
        <DataTable
          columns={contactColumns}
          data={contacts}
          isLoading={isLoading}
        />
      )}
    </div>
  );
}

// --- Properties Tab ---

function PropertiesTab() {
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [name, setName] = useState("");
  const [label, setLabel] = useState("");
  const [type, setType] = useState("string");

  const { data, isLoading } = useQuery({
    queryKey: ["properties"],
    queryFn: () => api.get("/properties").then((res) => res.data),
  });

  const properties: Property[] = data?.data ?? [];

  const createMutation = useMutation({
    mutationFn: (payload: { name: string; label: string; type: string }) =>
      api.post("/properties", payload).then((res) => res.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["properties"] });
      setDialogOpen(false);
      setName("");
      setLabel("");
      setType("string");
    },
  });

  const propertyColumns: ColumnDef<Property>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <span className="font-medium font-mono text-sm">
          {row.getValue("name")}
        </span>
      ),
    },
    {
      accessorKey: "label",
      header: "Label",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.getValue("label")}
        </span>
      ),
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <span className="rounded bg-muted px-2 py-0.5 text-xs font-mono">
          {row.getValue("type")}
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

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <PlusIcon className="mr-2 size-4" />
              Add Property
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Property</DialogTitle>
              <DialogDescription>
                Create a new contact property.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-2">
              <div className="space-y-2">
                <Label htmlFor="prop-name">Name</Label>
                <Input
                  id="prop-name"
                  placeholder="company_name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="prop-label">Label</Label>
                <Input
                  id="prop-label"
                  placeholder="Company Name"
                  value={label}
                  onChange={(e) => setLabel(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Type</Label>
                <Select value={type} onValueChange={setType}>
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="string">String</SelectItem>
                    <SelectItem value="number">Number</SelectItem>
                    <SelectItem value="boolean">Boolean</SelectItem>
                    <SelectItem value="date">Date</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button
                onClick={() =>
                  createMutation.mutate({ name, label, type })
                }
                disabled={!name || !label || createMutation.isPending}
              >
                {createMutation.isPending ? "Adding..." : "Add Property"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {!isLoading && properties.length === 0 ? (
        <EmptyState
          icon={TagIcon}
          title="No properties"
          description="Properties let you store custom data on contacts."
          actionLabel="Add Property"
          onAction={() => setDialogOpen(true)}
        />
      ) : (
        <DataTable
          columns={propertyColumns}
          data={properties}
          isLoading={isLoading}
        />
      )}
    </div>
  );
}

// --- Segments Tab ---

function SegmentsTab() {
  const { data, isLoading } = useQuery({
    queryKey: ["segments"],
    queryFn: () => api.get("/segments").then((res) => res.data),
  });

  const segments: Segment[] = data?.data ?? [];

  const segmentColumns: ColumnDef<Segment>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue("name")}</span>
      ),
    },
    {
      accessorKey: "conditions",
      header: "Conditions",
      cell: ({ row }) => (
        <span className="text-muted-foreground font-mono text-xs max-w-[300px] truncate block">
          {JSON.stringify(row.getValue("conditions"))}
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

  return (
    <div className="space-y-4">
      {!isLoading && segments.length === 0 ? (
        <EmptyState
          icon={FilterIcon}
          title="No segments"
          description="Segments let you target specific groups of contacts."
        />
      ) : (
        <DataTable
          columns={segmentColumns}
          data={segments}
          isLoading={isLoading}
        />
      )}
    </div>
  );
}

// --- Topics Tab ---

function TopicsTab() {
  const { data, isLoading } = useQuery({
    queryKey: ["topics"],
    queryFn: () => api.get("/topics").then((res) => res.data),
  });

  const topics: Topic[] = data?.data ?? [];

  const topicColumns: ColumnDef<Topic>[] = [
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
          {row.getValue("description") || "-"}
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

  return (
    <div className="space-y-4">
      {!isLoading && topics.length === 0 ? (
        <EmptyState
          icon={BookmarkIcon}
          title="No topics"
          description="Topics let contacts choose what emails they want to receive."
        />
      ) : (
        <DataTable
          columns={topicColumns}
          data={topics}
          isLoading={isLoading}
        />
      )}
    </div>
  );
}

// --- Main Page ---

export default function AudiencePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Audience"
        description="Manage contacts, properties, segments, and topics"
      />

      <Tabs defaultValue="contacts">
        <TabsList>
          <TabsTrigger value="contacts">Contacts</TabsTrigger>
          <TabsTrigger value="properties">Properties</TabsTrigger>
          <TabsTrigger value="segments">Segments</TabsTrigger>
          <TabsTrigger value="topics">Topics</TabsTrigger>
        </TabsList>

        <TabsContent value="contacts" className="mt-4">
          <ContactsTab />
        </TabsContent>

        <TabsContent value="properties" className="mt-4">
          <PropertiesTab />
        </TabsContent>

        <TabsContent value="segments" className="mt-4">
          <SegmentsTab />
        </TabsContent>

        <TabsContent value="topics" className="mt-4">
          <TopicsTab />
        </TabsContent>
      </Tabs>
    </div>
  );
}
