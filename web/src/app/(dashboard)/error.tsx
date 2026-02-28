"use client";

import { AlertTriangleIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <Card className="bg-card border-border max-w-md w-full">
        <CardContent className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-destructive/10 p-4 mb-4">
            <AlertTriangleIcon className="h-8 w-8 text-destructive" />
          </div>
          <h3 className="text-lg font-medium">Something went wrong</h3>
          <p className="text-sm text-muted-foreground mt-1 max-w-sm">
            {error.message || "An unexpected error occurred while loading this page."}
          </p>
          {error.digest && (
            <p className="text-xs text-muted-foreground/60 mt-2 font-mono">
              Error ID: {error.digest}
            </p>
          )}
          <Button onClick={reset} size="sm" className="mt-4">
            Try Again
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
