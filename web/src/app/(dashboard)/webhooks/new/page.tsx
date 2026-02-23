"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { ArrowLeftIcon } from "lucide-react";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const availableEvents = [
  { id: "email.sent", label: "Email Sent" },
  { id: "email.delivered", label: "Email Delivered" },
  { id: "email.bounced", label: "Email Bounced" },
  { id: "email.opened", label: "Email Opened" },
  { id: "email.clicked", label: "Email Clicked" },
  { id: "email.complained", label: "Email Complained" },
];

export default function NewWebhookPage() {
  const router = useRouter();
  const [url, setUrl] = useState("");
  const [selectedEvents, setSelectedEvents] = useState<string[]>([]);

  const createMutation = useMutation({
    mutationFn: (payload: { url: string; events: string[] }) =>
      api.post("/webhooks", payload).then((res) => res.data),
    onSuccess: () => {
      router.push("/webhooks");
    },
  });

  const toggleEvent = (eventId: string) => {
    setSelectedEvents((prev) =>
      prev.includes(eventId)
        ? prev.filter((e) => e !== eventId)
        : [...prev, eventId]
    );
  };

  const isValid = url && selectedEvents.length > 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => router.push("/webhooks")}
        >
          <ArrowLeftIcon className="mr-2 size-4" />
          Back
        </Button>
        <h1 className="text-xl font-semibold">New Webhook</h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Webhook Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="webhook-url">Endpoint URL</Label>
            <Input
              id="webhook-url"
              type="url"
              placeholder="https://example.com/webhooks/mailit"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
            />
          </div>

          <div className="space-y-3">
            <Label>Events</Label>
            <p className="text-sm text-muted-foreground">
              Select which events should trigger this webhook.
            </p>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {availableEvents.map((event) => (
                <label
                  key={event.id}
                  className="flex items-center gap-3 rounded-lg border border-border p-3 cursor-pointer hover:bg-muted/50 transition-colors"
                >
                  <Checkbox
                    checked={selectedEvents.includes(event.id)}
                    onCheckedChange={() => toggleEvent(event.id)}
                  />
                  <div>
                    <span className="text-sm font-medium">{event.label}</span>
                    <p className="text-xs text-muted-foreground font-mono">
                      {event.id}
                    </p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          <div className="flex items-center gap-3 pt-4">
            <Button
              onClick={() =>
                createMutation.mutate({ url, events: selectedEvents })
              }
              disabled={!isValid || createMutation.isPending}
            >
              {createMutation.isPending
                ? "Creating..."
                : "Create Webhook"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
