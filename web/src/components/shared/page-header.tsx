"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

interface PageHeaderProps {
  title: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
  children?: React.ReactNode;
}

export function PageHeader({ title, description, actionLabel, onAction, children }: PageHeaderProps) {
  useEffect(() => {
    document.title = `${title} — MailIt`;
  }, [title]);

  return (
    <div className="flex items-center justify-between pb-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {description && <p className="text-sm text-muted-foreground mt-1">{description}</p>}
      </div>
      <div className="flex items-center gap-2">
        {children}
        {actionLabel && onAction && (
          <Button onClick={onAction} size="sm">
            <Plus className="h-4 w-4 mr-1" />
            {actionLabel}
          </Button>
        )}
      </div>
    </div>
  );
}
