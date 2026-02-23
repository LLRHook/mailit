"use client";

import { useParams, useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeftIcon, ClockIcon } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { StatusBadge } from "@/components/shared/status-badge";
import { Skeleton } from "@/components/ui/skeleton";

interface EmailDetail {
  id: string;
  from: string;
  to: string;
  subject: string;
  status: string;
  html: string;
  text: string;
  created_at: string;
  sent_at: string | null;
  delivered_at: string | null;
  events: { type: string; created_at: string }[];
}

export default function EmailDetailPage() {
  const params = useParams<{ emailId: string }>();
  const router = useRouter();

  const { data, isLoading } = useQuery({
    queryKey: ["email", params.emailId],
    queryFn: () =>
      api.get(`/emails/${params.emailId}`).then((res) => res.data),
  });

  const email: EmailDetail | undefined = data?.data;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!email) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" onClick={() => router.push("/emails")}>
          <ArrowLeftIcon className="mr-2 size-4" />
          Back to Emails
        </Button>
        <p className="text-muted-foreground">Email not found.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" onClick={() => router.push("/emails")}>
          <ArrowLeftIcon className="mr-2 size-4" />
          Back
        </Button>
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold">{email.subject}</h1>
          <StatusBadge status={email.status} />
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Email Details</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                From
              </dt>
              <dd className="mt-1 text-sm">{email.from}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">To</dt>
              <dd className="mt-1 text-sm">{email.to}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                Subject
              </dt>
              <dd className="mt-1 text-sm">{email.subject}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                Status
              </dt>
              <dd className="mt-1">
                <StatusBadge status={email.status} />
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                Created
              </dt>
              <dd className="mt-1 text-sm">
                {format(new Date(email.created_at), "MMM d, yyyy HH:mm:ss")}
              </dd>
            </div>
            {email.sent_at && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Sent
                </dt>
                <dd className="mt-1 text-sm">
                  {format(new Date(email.sent_at), "MMM d, yyyy HH:mm:ss")}
                </dd>
              </div>
            )}
            {email.delivered_at && (
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  Delivered
                </dt>
                <dd className="mt-1 text-sm">
                  {format(
                    new Date(email.delivered_at),
                    "MMM d, yyyy HH:mm:ss"
                  )}
                </dd>
              </div>
            )}
          </dl>
        </CardContent>
      </Card>

      <Tabs defaultValue="preview">
        <TabsList>
          <TabsTrigger value="preview">Preview</TabsTrigger>
          <TabsTrigger value="source">Source</TabsTrigger>
          <TabsTrigger value="events">Events</TabsTrigger>
        </TabsList>

        <TabsContent value="preview" className="mt-4">
          <Card>
            <CardContent className="p-0">
              {email.html ? (
                <iframe
                  srcDoc={email.html}
                  className="w-full min-h-[400px] rounded-lg bg-white"
                  sandbox="allow-same-origin"
                  title="Email preview"
                />
              ) : (
                <div className="p-6 text-sm text-muted-foreground whitespace-pre-wrap">
                  {email.text || "No content available."}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="source" className="mt-4">
          <Card>
            <CardContent>
              <pre className="overflow-x-auto rounded-lg bg-muted/50 p-4 text-sm font-mono text-foreground">
                <code>{email.html || email.text || "No content available."}</code>
              </pre>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="events" className="mt-4">
          <Card>
            <CardContent>
              {email.events && email.events.length > 0 ? (
                <div className="space-y-3">
                  {email.events.map((event, i) => (
                    <div
                      key={i}
                      className="flex items-center gap-3 rounded-lg border border-border p-3"
                    >
                      <ClockIcon className="size-4 text-muted-foreground shrink-0" />
                      <div className="flex-1">
                        <StatusBadge status={event.type} />
                      </div>
                      <span className="text-sm text-muted-foreground">
                        {format(
                          new Date(event.created_at),
                          "MMM d, yyyy HH:mm:ss"
                        )}
                      </span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground py-4 text-center">
                  No events recorded.
                </p>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
