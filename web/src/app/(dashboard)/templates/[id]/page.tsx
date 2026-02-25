"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ArrowLeftIcon } from "lucide-react";
import { toast } from "sonner";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

interface TemplateDetail {
  id: string;
  name: string;
  description: string;
  subject: string;
  html: string;
  text: string;
  published: boolean;
  created_at: string;
  updated_at: string;
}

export default function EditTemplatePage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["template", params.id],
    queryFn: () =>
      api.get(`/templates/${params.id}`).then((res) => res.data),
  });

  const template: TemplateDetail | undefined = data?.data;

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [subject, setSubject] = useState("");
  const [html, setHtml] = useState("");
  const [text, setText] = useState("");

  // Sync form state when template data loads for the first time.
  const [initialized, setInitialized] = useState(false);
  if (template && !initialized) {
    setName(template.name);
    setDescription(template.description);
    setSubject(template.subject);
    setHtml(template.html);
    setText(template.text);
    setInitialized(true);
  }

  const updateMutation = useMutation({
    mutationFn: (payload: {
      name: string;
      description: string;
      subject: string;
      html: string;
      text: string;
    }) =>
      api.patch(`/templates/${params.id}`, payload).then((res) => res.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["template", params.id] });
      toast.success("Template saved");
    },
    onError: () => toast.error("Failed to save template"),
  });

  const publishMutation = useMutation({
    mutationFn: () =>
      api
        .post(`/templates/${params.id}/publish`)
        .then((res) => res.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["template", params.id] });
      toast.success("Template published");
    },
    onError: () => toast.error("Failed to publish template"),
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (!template) {
    return (
      <div className="space-y-6">
        <Button
          variant="ghost"
          onClick={() => router.push("/templates")}
        >
          <ArrowLeftIcon className="mr-2 size-4" />
          Back to Templates
        </Button>
        <p className="text-muted-foreground">Template not found.</p>
      </div>
    );
  }

  const isValid = name && subject && (html || text);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.push("/templates")}
          >
            <ArrowLeftIcon className="mr-2 size-4" />
            Back
          </Button>
          <h1 className="text-xl font-semibold">Edit Template</h1>
        </div>
        <Button
          variant="outline"
          onClick={() => publishMutation.mutate()}
          disabled={publishMutation.isPending}
        >
          {publishMutation.isPending ? "Publishing..." : "Publish"}
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Template Details</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="subject">Subject</Label>
            <Input
              id="subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="html">HTML Body</Label>
            <Textarea
              id="html"
              className="min-h-[200px] font-mono text-sm"
              value={html}
              onChange={(e) => setHtml(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="text">Plain Text Body</Label>
            <Textarea
              id="text"
              className="min-h-[100px]"
              value={text}
              onChange={(e) => setText(e.target.value)}
            />
          </div>

          <div className="flex items-center gap-3 pt-4">
            <Button
              onClick={() =>
                updateMutation.mutate({
                  name,
                  description,
                  subject,
                  html,
                  text,
                })
              }
              disabled={!isValid || updateMutation.isPending}
            >
              {updateMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
