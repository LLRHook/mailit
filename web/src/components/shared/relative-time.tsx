"use client";

import { formatDistanceToNow, format } from "date-fns";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface RelativeTimeProps {
  date: string;
  className?: string;
}

export function RelativeTime({ date, className }: RelativeTimeProps) {
  const parsed = new Date(date);
  const relative = formatDistanceToNow(parsed, { addSuffix: true });
  const absolute = format(parsed, "MMM d, yyyy HH:mm:ss");

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <span className={className}>{relative}</span>
        </TooltipTrigger>
        <TooltipContent>
          <p>{absolute}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
