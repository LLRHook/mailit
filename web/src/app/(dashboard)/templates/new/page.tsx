"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { ArrowLeftIcon } from "lucide-react";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function NewTemplatePage() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [subject, setSubject] = useState("");
  const [html, setHtml] = useState("");
  const [text, setText] = useState("");

  const createMutation = useMutation({
    mutationFn: (payload: {
      name: string;
      description: string;
      subject: string;
      html: string;
      text: string;
    }) => api.post("/templates", payload).then((res) => res.data),
    onSuccess: (data) => {
      router.push(`/templates/${data.data.id}`);
    },
  });

  const isValid = name && subject && (html || text);

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => router.push("/templates")}
        >
          <ArrowLeftIcon className="mr-2 size-4" />
          Back
        </Button>
        <h1 className="text-xl font-semibold">New Template</h1>
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
                placeholder="Welcome email"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                placeholder="Sent to new users after signup"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="subject">Subject</Label>
            <Input
              id="subject"
              placeholder="Welcome to {{company_name}}"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="html">HTML Body</Label>
            <Textarea
              id="html"
              placeholder="<html><body>Your template content...</body></html>"
              className="min-h-[200px] font-mono text-sm"
              value={html}
              onChange={(e) => setHtml(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="text">Plain Text Body</Label>
            <Textarea
              id="text"
              placeholder="Your plain text template content..."
              className="min-h-[100px]"
              value={text}
              onChange={(e) => setText(e.target.value)}
            />
          </div>

          <div className="flex items-center gap-3 pt-4">
            <Button
              onClick={() =>
                createMutation.mutate({ name, description, subject, html, text })
              }
              disabled={!isValid || createMutation.isPending}
            >
              {createMutation.isPending ? "Creating..." : "Create Template"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
