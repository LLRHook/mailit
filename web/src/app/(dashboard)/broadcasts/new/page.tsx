"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useMutation } from "@tanstack/react-query";
import { ArrowLeftIcon } from "lucide-react";
import { toast } from "sonner";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Audience {
  id: string;
  name: string;
}

export default function NewBroadcastPage() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [from, setFrom] = useState("");
  const [subject, setSubject] = useState("");
  const [audienceId, setAudienceId] = useState("");
  const [html, setHtml] = useState("");
  const [text, setText] = useState("");

  const { data: audiencesData } = useQuery({
    queryKey: ["audiences"],
    queryFn: () => api.get("/audiences").then((res) => res.data),
  });

  const audiences: Audience[] = audiencesData?.data ?? [];

  const createMutation = useMutation({
    mutationFn: (payload: {
      name: string;
      from: string;
      subject: string;
      audience_id: string;
      html: string;
      text: string;
    }) => api.post("/broadcasts", payload).then((res) => res.data),
  });

  const sendMutation = useMutation({
    mutationFn: (id: string) =>
      api.post(`/broadcasts/${id}/send`).then((res) => res.data),
  });

  const handleSaveDraft = async () => {
    try {
      const result = await createMutation.mutateAsync({
        name,
        from,
        subject,
        audience_id: audienceId,
        html,
        text,
      });
      toast.success("Draft saved");
      router.push(`/broadcasts/${result.data.id}`);
    } catch {
      toast.error("Failed to save draft");
    }
  };

  const handleSendNow = async () => {
    try {
      const result = await createMutation.mutateAsync({
        name,
        from,
        subject,
        audience_id: audienceId,
        html,
        text,
      });
      await sendMutation.mutateAsync(result.data.id);
      toast.success("Broadcast sent");
      router.push(`/broadcasts/${result.data.id}`);
    } catch {
      toast.error("Failed to send broadcast");
    }
  };

  const isSubmitting = createMutation.isPending || sendMutation.isPending;
  const isValid = name && from && subject && audienceId && (html || text);

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
        <h1 className="text-xl font-semibold">New Broadcast</h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Broadcast Details</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="Weekly newsletter"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="from">From</Label>
              <Input
                id="from"
                placeholder="hello@example.com"
                value={from}
                onChange={(e) => setFrom(e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="subject">Subject</Label>
            <Input
              id="subject"
              placeholder="Your email subject line"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label>Audience</Label>
            <Select value={audienceId} onValueChange={setAudienceId}>
              <SelectTrigger className="w-full">
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
          </div>

          <div className="space-y-2">
            <Label htmlFor="html">HTML Body</Label>
            <Textarea
              id="html"
              placeholder="<html><body>Your email content...</body></html>"
              className="min-h-[200px] font-mono text-sm"
              value={html}
              onChange={(e) => setHtml(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="text">Plain Text Body</Label>
            <Textarea
              id="text"
              placeholder="Your plain text email content..."
              className="min-h-[100px]"
              value={text}
              onChange={(e) => setText(e.target.value)}
            />
          </div>

          <div className="flex items-center gap-3 pt-4">
            <Button
              variant="outline"
              onClick={handleSaveDraft}
              disabled={!isValid || isSubmitting}
            >
              Save Draft
            </Button>
            <Button
              onClick={handleSendNow}
              disabled={!isValid || isSubmitting}
            >
              Send Now
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
