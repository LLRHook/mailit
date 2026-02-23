"use client";

import { useParams, useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeftIcon, UsersIcon, SendIcon, MailIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBadge } from "@/components/shared/status-badge";
import { StatCard } from "@/components/shared/stat-card";
import { Skeleton } from "@/components/ui/skeleton";

interface BroadcastDetail {
  id: string;
  name: string;
  from: string;
  subject: string;
  audience_id: string;
  audience_name: string;
  status: string;
  recipients: number;
  sent: number;
  delivered: number;
  opened: number;
  html: string;
  text: string;
  created_at: string;
  sent_at: string | null;
  completed_at: string | null;
}

export default function BroadcastDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();

  const { data, isLoading } = useQuery({
    queryKey: ["broadcast", params.id],
    queryFn: () =>
      api.get(`/broadcasts/${params.id}`).then((res) => res.data),
  });

  const broadcast: BroadcastDetail | undefined = data?.data;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
        </div>
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!broadcast) {
    return (
      <div className="space-y-6">
        <Button
          variant="ghost"
          onClick={() => router.push("/broadcasts")}
        >
          <ArrowLeftIcon className="mr-2 size-4" />
          Back to Broadcasts
        </Button>
        <p className="text-muted-foreground">Broadcast not found.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => router.push("/broadcasts")}
        >
          <ArrowLeftIcon className="mr-2 size-4" />
          Back
        </Button>
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold">{broadcast.name}</h1>
          <StatusBadge status={broadcast.status} />
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard
          title="Recipients"
          value={broadcast.recipients.toLocaleString()}
          icon={UsersIcon}
        />
        <StatCard
          title="Sent"
          value={broadcast.sent.toLocaleString()}
          icon={SendIcon}
        />
        <StatCard
          title="Delivered"
          value={broadcast.delivered.toLocaleString()}
          icon={MailIcon}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Broadcast Details</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                From
              </dt>
              <dd className="mt-1 text-sm">{broadcast.from}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                Subject
              </dt>
              <dd className="mt-1 text-sm">{broadcast.subject}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                Audience
              </dt>
              <dd className="mt-1 text-sm">{broadcast.audience_name}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                Status
              </dt>
              <dd className="mt-1">
                <StatusBadge status={broadcast.status} />
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                Created
              </dt>
              <dd className="mt-1 text-sm">
                {format(
                  new Date(broadcast.created_at),
                  "MMM d, yyyy HH:mm:ss"
                )}
              </dd>
            </div>
            {broadcast.sent_at && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Sent At
                </dt>
                <dd className="mt-1 text-sm">
                  {format(
                    new Date(broadcast.sent_at),
                    "MMM d, yyyy HH:mm:ss"
                  )}
                </dd>
              </div>
            )}
            {broadcast.completed_at && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Completed At
                </dt>
                <dd className="mt-1 text-sm">
                  {format(
                    new Date(broadcast.completed_at),
                    "MMM d, yyyy HH:mm:ss"
                  )}
                </dd>
              </div>
            )}
          </dl>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Preview</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {broadcast.html ? (
            <iframe
              srcDoc={broadcast.html}
              className="w-full min-h-[400px] rounded-b-lg bg-white"
              sandbox="allow-same-origin"
              title="Broadcast preview"
            />
          ) : (
            <div className="p-6 text-sm text-muted-foreground whitespace-pre-wrap">
              {broadcast.text || "No content available."}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
