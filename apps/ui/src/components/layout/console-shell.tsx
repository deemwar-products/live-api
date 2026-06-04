import { useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { ChevronDown, LogOut, Menu, Search, UserRound, X } from "lucide-react";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { Button } from "@/components/ui/button";
import { ConsoleNav } from "@/components/layout/console-nav";
import { BUSINESS_LABELS } from "@/labels";
import { cn } from "@/lib/cn";
import { LANDING } from "@/lib/paths";
import { getOrgSnapshot } from "@/mocks/business/console";

/**
 * App shell for the org console — sticky left sidebar (orgs, primary nav)
 * and a top bar (org switcher, search, theme, profile). Used by every
 * `/console/*` page.
 */
export function ConsoleShell({ children, rightSlot }: { children: ReactNode; rightSlot?: ReactNode }) {
  const [mobileOpen, setMobileOpen] = useState(false);
  const org = getOrgSnapshot();

  return (
    <div className="grain relative flex min-h-svh flex-col bg-bg lg:flex-row">
      <div className="absolute inset-0 -z-10 bg-ambient opacity-40" />

      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 flex w-64 shrink-0 flex-col border-r border-border bg-bg-elevated/80 backdrop-blur-xl transition-transform duration-500 ease-[var(--ease-out-soft)] lg:static lg:translate-x-0",
          mobileOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        <div className="flex h-16 items-center justify-between border-b border-border/60 px-5">
          <Link
            to={LANDING.home}
            className="group flex items-center gap-2 text-[15px] font-semibold tracking-tight text-fg"
          >
            <LogoMark />
            <span>{BUSINESS_LABELS.shell.brand}</span>
          </Link>
          <button
            type="button"
            onClick={() => setMobileOpen(false)}
            aria-label="Close navigation"
            className="inline-flex size-8 items-center justify-center rounded-full text-fg-muted hover:bg-bg-muted hover:text-fg lg:hidden"
          >
            <X className="size-4" />
          </button>
        </div>

        <div className="px-3 py-3">
          <button
            type="button"
            aria-label={BUSINESS_LABELS.shell.orgSwitcherLabel}
            className="group flex w-full items-center justify-between gap-2 rounded-lg border border-border bg-bg-elevated px-3 py-2 text-left text-sm transition-colors duration-300 ease-[var(--ease-out-soft)] hover:bg-bg-muted"
          >
            <div className="flex min-w-0 items-center gap-2.5">
              <span className="inline-flex size-7 shrink-0 items-center justify-center rounded-md bg-accent-soft text-[11px] font-semibold tracking-tight text-accent-strong">
                {org.name.slice(0, 2).toUpperCase()}
              </span>
              <div className="min-w-0">
                <div className="truncate font-medium text-fg">{org.name}</div>
                <div className="text-[11px] text-fg-subtle">
                  {org.plan} · {org.region}
                </div>
              </div>
            </div>
            <ChevronDown className="size-3.5 shrink-0 text-fg-subtle transition-transform duration-300 ease-[var(--ease-out-soft)] group-hover:translate-y-0.5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto">
          <ConsoleNav onNavigate={() => setMobileOpen(false)} />
        </div>

        <div className="border-t border-border/60 p-3">
          <Link
            to={LANDING.home}
            className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:bg-bg-muted/60 hover:text-fg"
          >
            <LogOut className="size-4" strokeWidth={1.75} />
            <span>{BUSINESS_LABELS.shell.signOut}</span>
          </Link>
        </div>
      </aside>

      {mobileOpen && (
        <button
          type="button"
          aria-label="Close navigation overlay"
          onClick={() => setMobileOpen(false)}
          className="fixed inset-0 z-30 bg-fg/30 backdrop-blur-sm lg:hidden"
        />
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex h-16 items-center gap-3 border-b border-border/60 bg-bg/70 px-4 backdrop-blur-xl sm:px-6">
          <button
            type="button"
            onClick={() => setMobileOpen(true)}
            aria-label="Open navigation"
            className="inline-flex size-9 items-center justify-center rounded-full border border-border bg-bg-elevated text-fg-muted hover:text-fg lg:hidden"
          >
            <Menu className="size-4" />
          </button>

          <div className="relative hidden flex-1 sm:block">
            <Search className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-fg-subtle" />
            <input
              type="search"
              placeholder={BUSINESS_LABELS.shell.searchPlaceholder}
              className="h-10 w-full max-w-md rounded-full border border-border bg-bg-elevated pl-10 pr-4 text-sm text-fg placeholder:text-fg-subtle transition-colors duration-300 ease-[var(--ease-out-soft)] hover:border-border-strong focus:border-accent focus:outline-none"
            />
          </div>

          <div className="ml-auto flex items-center gap-2">
            {rightSlot ?? (
              <>
                <ThemeToggle />
                <Button variant="secondary" size="sm" leading={<UserRound className="size-4" />}>
                  {BUSINESS_LABELS.shell.profile}
                </Button>
              </>
            )}
          </div>
        </header>

        <main className="min-w-0 flex-1">{children}</main>
      </div>
    </div>
  );
}

function LogoMark() {
  return (
    <span className="relative inline-flex size-7 items-center justify-center overflow-hidden rounded-lg bg-fg text-bg">
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden>
        <path
          d="M2 7 L6 7 M8 4 L8 10 M10 5 L10 9 M12 3 L12 11"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
        />
      </svg>
    </span>
  );
}
