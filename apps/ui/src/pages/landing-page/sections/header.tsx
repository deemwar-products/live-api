import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { LANDING_LABELS, COMMON_LABELS } from "@/labels";
import { cn } from "@/lib/cn";

const NAV = [
  { href: "#product", label: LANDING_LABELS.nav.product },
  { href: "#how-it-works", label: LANDING_LABELS.nav.howItWorks },
  { href: "#use-cases", label: LANDING_LABELS.nav.useCases },
  { href: "#pricing", label: LANDING_LABELS.nav.pricing },
];

const ease = [0.22, 1, 0.36, 1] as const;

export function Header() {
  return (
    <motion.header
      initial={{ y: -20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.8, ease }}
      className={cn(
        "sticky top-0 z-50 w-full",
        "border-b border-border",
        "bg-bg/70 backdrop-blur-xl backdrop-saturate-150"
      )}
    >
      <div className="mx-auto flex h-16 w-full max-w-7xl items-center justify-between px-6 sm:px-8">
        <a
          href="#"
          className="group flex items-center gap-2 text-[15px] font-semibold tracking-tight text-fg"
        >
          <LogoMark />
          <span>{COMMON_LABELS.app.name}</span>
        </a>

        <nav className="hidden items-center gap-1 md:flex">
          {NAV.map((item) => (
            <a
              key={item.href}
              href={item.href}
              className="rounded-full px-3.5 py-1.5 text-sm text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:text-fg"
            >
              {item.label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          <ThemeToggle />
          <a
            href="#signin"
            className="hidden rounded-full px-3.5 py-1.5 text-sm text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:text-fg sm:block"
          >
            {COMMON_LABELS.actions.signIn}
          </a>
          <Button size="sm" className="hidden sm:inline-flex">
            {COMMON_LABELS.actions.startFree}
          </Button>
        </div>
      </div>
    </motion.header>
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
