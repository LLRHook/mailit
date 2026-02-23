"use client";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Code } from "lucide-react";
import { CodeBlock } from "./code-block";

interface ApiExample {
  title: string;
  method: string;
  endpoint: string;
  body?: string;
}

interface ApiDrawerProps {
  examples: ApiExample[];
}

function buildCurl(example: ApiExample): string {
  let cmd = `curl -X ${example.method} \\
  https://api.yourdomain.com${example.endpoint} \\
  -H "Authorization: Bearer re_xxxxx" \\
  -H "Content-Type: application/json"`;

  if (example.body) {
    cmd += ` \\
  -d '${example.body}'`;
  }

  return cmd;
}

export function ApiDrawer({ examples }: ApiDrawerProps) {
  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline" size="icon" className="h-8 w-8">
          <Code className="h-4 w-4" />
        </Button>
      </SheetTrigger>
      <SheetContent className="w-[480px] sm:max-w-[480px] overflow-y-auto">
        <SheetHeader>
          <SheetTitle>API Reference</SheetTitle>
        </SheetHeader>
        <div className="mt-6 space-y-6">
          {examples.map((example, i) => (
            <div key={i}>
              <div className="flex items-center gap-2 mb-2">
                <span className="text-xs font-mono font-semibold px-1.5 py-0.5 rounded bg-teal-500/10 text-teal-400">
                  {example.method}
                </span>
                <span className="text-sm font-mono text-muted-foreground">
                  {example.endpoint}
                </span>
              </div>
              <p className="text-sm text-muted-foreground mb-2">{example.title}</p>
              <CodeBlock code={buildCurl(example)} />
            </div>
          ))}
        </div>
      </SheetContent>
    </Sheet>
  );
}
