"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Mail,
  Radio,
  FileText,
  Users,
  BarChart3,
  Globe,
  ScrollText,
  Key,
  Webhook,
  Settings,
} from "lucide-react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";

const pages = [
  { title: "Emails", href: "/emails", icon: Mail },
  { title: "Broadcasts", href: "/broadcasts", icon: Radio },
  { title: "Templates", href: "/templates", icon: FileText },
  { title: "Audience", href: "/audience", icon: Users },
  { title: "Metrics", href: "/metrics", icon: BarChart3 },
  { title: "Domains", href: "/domains", icon: Globe },
  { title: "Logs", href: "/logs", icon: ScrollText },
  { title: "API Keys", href: "/api-keys", icon: Key },
  { title: "Webhooks", href: "/webhooks", icon: Webhook },
  { title: "Settings", href: "/settings", icon: Settings },
];

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const router = useRouter();

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  const navigate = (href: string) => {
    setOpen(false);
    router.push(href);
  };

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Search pages..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading="Pages">
          {pages.map((page) => (
            <CommandItem
              key={page.href}
              onSelect={() => navigate(page.href)}
            >
              <page.icon className="mr-2 h-4 w-4" />
              {page.title}
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  );
}
