"use client";

import { useQuery } from "@tanstack/react-query";
import {
  BarChart3Icon,
  UsersIcon,
  ServerIcon,
  PuzzleIcon,
  PlusIcon,
} from "lucide-react";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageHeader } from "@/components/shared/page-header";
import { CopyButton } from "@/components/shared/copy-button";
import { EmptyState } from "@/components/shared/empty-state";

// --- Usage Tab ---

function UsageTab() {
  const { data } = useQuery({
    queryKey: ["settings-usage"],
    queryFn: () => api.get("/settings/usage").then((res) => res.data),
  });

  const usage = data?.data ?? {
    emails_sent_today: 0,
    emails_sent_month: 0,
    domains: 0,
    api_keys: 0,
    webhooks: 0,
    contacts: 0,
  };

  const stats = [
    { label: "Emails Sent Today", value: usage.emails_sent_today },
    { label: "Emails Sent This Month", value: usage.emails_sent_month },
    { label: "Domains", value: usage.domains },
    { label: "API Keys", value: usage.api_keys },
    { label: "Webhooks", value: usage.webhooks },
    { label: "Contacts", value: usage.contacts },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Current Usage</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {stats.map((stat) => (
            <div
              key={stat.label}
              className="rounded-lg border border-border p-4"
            >
              <p className="text-sm text-muted-foreground">{stat.label}</p>
              <p className="text-2xl font-bold mt-1 tabular-nums">
                {stat.value.toLocaleString()}
              </p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

// --- Team Tab ---

interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: string;
}

function TeamTab() {
  const { data } = useQuery({
    queryKey: ["settings-team"],
    queryFn: () => api.get("/settings/team").then((res) => res.data),
  });

  const members: TeamMember[] = data?.data ?? [];

  const roleColor: Record<string, string> = {
    owner: "bg-primary/15 text-primary border-primary/20",
    admin: "bg-blue-500/15 text-blue-400 border-blue-500/20",
    member: "bg-zinc-500/15 text-zinc-400 border-zinc-500/20",
  };

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between">
        <CardTitle>Team Members</CardTitle>
        <Button size="sm">
          <PlusIcon className="mr-2 size-4" />
          Invite Member
        </Button>
      </CardHeader>
      <CardContent>
        {members.length === 0 ? (
          <EmptyState
            icon={UsersIcon}
            title="No team members"
            description="Invite team members to collaborate on your project."
          />
        ) : (
          <div className="space-y-3">
            {members.map((member) => (
              <div
                key={member.id}
                className="flex items-center justify-between rounded-lg border border-border p-4"
              >
                <div className="flex items-center gap-3">
                  <div className="flex size-9 items-center justify-center rounded-full bg-muted text-sm font-medium">
                    {member.name
                      .split(" ")
                      .map((n) => n[0])
                      .join("")
                      .toUpperCase()}
                  </div>
                  <div>
                    <p className="text-sm font-medium">{member.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {member.email}
                    </p>
                  </div>
                </div>
                <Badge
                  variant="outline"
                  className={`capitalize ${roleColor[member.role] ?? ""}`}
                >
                  {member.role}
                </Badge>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// --- SMTP Tab ---

function SmtpTab() {
  const smtpConfig = {
    host: "smtp.mailit.local",
    port: "587",
    username: "mailit",
    password: "your-api-key",
    encryption: "STARTTLS",
  };

  const fields = [
    { label: "Host", value: smtpConfig.host },
    { label: "Port", value: smtpConfig.port },
    { label: "Username", value: smtpConfig.username },
    { label: "Password", value: smtpConfig.password },
    { label: "Encryption", value: smtpConfig.encryption },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>SMTP Credentials</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground mb-4">
          Use these credentials to send emails via SMTP.
        </p>
        <div className="space-y-3">
          {fields.map((field) => (
            <div
              key={field.label}
              className="flex items-center justify-between rounded-lg border border-border p-3"
            >
              <div>
                <p className="text-xs text-muted-foreground">{field.label}</p>
                <p className="text-sm font-mono mt-0.5">{field.value}</p>
              </div>
              <CopyButton value={field.value} />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

// --- Integrations Tab ---

function IntegrationsTab() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Integrations</CardTitle>
      </CardHeader>
      <CardContent>
        <EmptyState
          icon={PuzzleIcon}
          title="Coming soon"
          description="Third-party integrations will be available in a future release."
        />
      </CardContent>
    </Card>
  );
}

// --- Main Page ---

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Settings"
        description="Manage your account and project settings"
      />

      <Tabs defaultValue="usage">
        <TabsList>
          <TabsTrigger value="usage">
            <BarChart3Icon className="mr-1.5 size-4" />
            Usage
          </TabsTrigger>
          <TabsTrigger value="team">
            <UsersIcon className="mr-1.5 size-4" />
            Team
          </TabsTrigger>
          <TabsTrigger value="smtp">
            <ServerIcon className="mr-1.5 size-4" />
            SMTP
          </TabsTrigger>
          <TabsTrigger value="integrations">
            <PuzzleIcon className="mr-1.5 size-4" />
            Integrations
          </TabsTrigger>
        </TabsList>

        <TabsContent value="usage" className="mt-4">
          <UsageTab />
        </TabsContent>

        <TabsContent value="team" className="mt-4">
          <TeamTab />
        </TabsContent>

        <TabsContent value="smtp" className="mt-4">
          <SmtpTab />
        </TabsContent>

        <TabsContent value="integrations" className="mt-4">
          <IntegrationsTab />
        </TabsContent>
      </Tabs>
    </div>
  );
}
