import { motion } from "framer-motion";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { MarketingHeader } from "@/components/layout/marketing-header";
import { LANDING_LABELS, COMMON_LABELS } from "@/labels";
import { PUBLIC_ROUTES } from "@/lib/paths";

const NAV = [
  { href: "#product", label: LANDING_LABELS.nav.product },
  { href: "#how-it-works", label: LANDING_LABELS.nav.howItWorks },
  { href: "#use-cases", label: LANDING_LABELS.nav.useCases },
  { href: "#pricing", label: LANDING_LABELS.nav.pricing },
];

const ease = [0.22, 1, 0.36, 1] as const;

export function Header() {
  return (
    <MarketingHeader
      rightSlot={
        <>
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
            <motion.div
              initial={{ opacity: 0, y: -6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, ease, delay: 0.1 }}
              className="hidden sm:block"
            >
              <Link
                to={PUBLIC_ROUTES.auth}
                className="rounded-full px-3.5 py-1.5 text-sm text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:text-fg"
              >
                {COMMON_LABELS.actions.signIn}
              </Link>
            </motion.div>
            <Button size="sm" className="hidden sm:inline-flex">
              {COMMON_LABELS.actions.startFree}
            </Button>
          </div>
        </>
      }
    />
  );
}
