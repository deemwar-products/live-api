import type { LucideIcon } from "lucide-react";
import {
  LayoutDashboard,
  MessageSquare,
  BookOpen,
  Wrench,
  Activity,
  Users,
  CreditCard,
  Settings,
} from "lucide-react";
import { NavLink } from "react-router-dom";
import { cn } from "@/lib/cn";
import { BUSINESS_LABELS } from "@/labels";
import { CONSOLE } from "@/lib/paths";

type NavItem = {
  to: string;
  label: string;
  icon: LucideIcon;
};

const NAV: NavItem[] = [
  { to: CONSOLE.home, label: BUSINESS_LABELS.nav.overview, icon: LayoutDashboard },
  { to: CONSOLE.conversations, label: BUSINESS_LABELS.nav.conversations, icon: MessageSquare },
  { to: CONSOLE.knowledge, label: BUSINESS_LABELS.nav.knowledge, icon: BookOpen },
  { to: CONSOLE.tools, label: BUSINESS_LABELS.nav.tools, icon: Wrench },
  { to: CONSOLE.monitoring, label: BUSINESS_LABELS.nav.monitoring, icon: Activity },
  { to: CONSOLE.team, label: BUSINESS_LABELS.nav.team, icon: Users },
  { to: CONSOLE.billing, label: BUSINESS_LABELS.nav.billing, icon: CreditCard },
  { to: CONSOLE.settings, label: BUSINESS_LABELS.nav.settings, icon: Settings },
];

export function ConsoleNav({ onNavigate }: { onNavigate?: () => void }) {
  return (
    <nav className="flex flex-col gap-0.5 p-3">
      {NAV.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={item.to === CONSOLE.home}
          onClick={onNavigate}
          className={({ isActive }) =>
            cn(
              "group flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium tracking-tight transition-all duration-300 ease-[var(--ease-out-soft)]",
              isActive
                ? "bg-bg-muted text-fg"
                : "text-fg-muted hover:bg-bg-muted/60 hover:text-fg"
            )
          }
        >
          <item.icon className="size-4 shrink-0" strokeWidth={1.75} />
          <span className="truncate">{item.label}</span>
        </NavLink>
      ))}
    </nav>
  );
}
