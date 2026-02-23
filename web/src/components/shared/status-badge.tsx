import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const statusVariants: Record<string, string> = {
  // Green
  delivered: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  verified: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  sent: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  active: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  // Red
  failed: "bg-red-500/10 text-red-400 border-red-500/20",
  bounced: "bg-red-500/10 text-red-400 border-red-500/20",
  complained: "bg-red-500/10 text-red-400 border-red-500/20",
  cancelled: "bg-red-500/10 text-red-400 border-red-500/20",
  // Yellow
  pending: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
  queued: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
  scheduled: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
  sending: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
  draft: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
  // Blue
  opened: "bg-blue-500/10 text-blue-400 border-blue-500/20",
  clicked: "bg-blue-500/10 text-blue-400 border-blue-500/20",
};

export function StatusBadge({ status }: { status: string }) {
  return (
    <Badge variant="outline" className={cn("capitalize", statusVariants[status] || "bg-zinc-500/10 text-zinc-400 border-zinc-500/20")}>
      {status}
    </Badge>
  );
}
