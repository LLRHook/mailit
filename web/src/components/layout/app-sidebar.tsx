"use client";
import { usePathname, useRouter } from "next/navigation";
import Link from "next/link";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarFooter,
} from "@/components/ui/sidebar";
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
  Send,
  LogOut,
} from "lucide-react";
import { Button } from "@/components/ui/button";

const navItems = [
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

export function AppSidebar() {
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = () => {
    localStorage.removeItem("mailit_token");
    document.cookie = "mailit_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
    router.push("/login");
  };

  return (
    <Sidebar className="border-r border-border">
      <SidebarHeader className="p-4">
        <Link href="/emails" className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
            <Send className="h-4 w-4 text-primary-foreground" />
          </div>
          <span className="text-lg font-bold">MailIt</span>
        </Link>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.href}>
                  <SidebarMenuButton asChild isActive={pathname.startsWith(item.href)}>
                    <Link href={item.href}>
                      <item.icon className="h-4 w-4" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter className="p-4 space-y-2">
        <Button
          variant="ghost"
          size="sm"
          className="w-full justify-start text-muted-foreground hover:text-foreground"
          onClick={handleLogout}
        >
          <LogOut className="mr-2 h-4 w-4" />
          Log out
        </Button>
        <p className="text-xs text-muted-foreground">MailIt v0.1.0</p>
      </SidebarFooter>
    </Sidebar>
  );
}
