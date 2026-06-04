import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Container } from "@/components/ui/container";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { cn } from "@/lib/cn";

type Props = {
  /**
   * Right-side action area. Pass <Button>s, <Link>s, anything.
   * Falls back to the theme toggle if omitted.
   */
  rightSlot?: ReactNode;
  className?: string;
};

/**
 * Marketing header — the brand chrome at the top of any public page
 * (landing, auth, pricing, etc.). Logo on the left, whatever you want
 * on the right.
 *
 * Pages compose their own right-side actions (sign-in links, CTAs,
 * theme toggle, back-links) and pass them in via `rightSlot`.
 */
export function MarketingHeader({ rightSlot, className }: Props) {
  return (
    <header
      className={cn(
        "sticky top-0 z-50 w-full border-b border-border bg-bg/70 backdrop-blur-xl backdrop-saturate-150",
        className
      )}
    >
      <Container className="flex h-16 items-center justify-between gap-4">
        <Link
          to="/"
          className="group flex items-center gap-2 text-[15px] font-semibold tracking-tight text-fg"
        >
          <LogoMark />
          <span>Live API</span>
        </Link>

        <div className="flex items-center gap-2">
          {rightSlot ?? <ThemeToggle />}
        </div>
      </Container>
    </header>
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
