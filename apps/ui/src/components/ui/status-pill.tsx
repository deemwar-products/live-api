import type { HealthBand } from "@/mocks/business/console";
import { cn } from "@/lib/cn";

type Variant = HealthBand | "neutral";

const variantClass: Record<Variant, string> = {
  normal: "bg-success-soft text-success",
  monitor: "bg-warning-soft text-warning",
  atRisk: "bg-watch-soft text-watch",
  critical: "bg-danger-soft text-danger",
  neutral: "bg-bg-muted text-fg-muted",
};

const dotClass: Record<Variant, string> = {
  normal: "bg-success",
  monitor: "bg-warning",
  atRisk: "bg-watch",
  critical: "bg-danger",
  neutral: "bg-fg-subtle",
};

/**
 * Small status pill with a leading dot. Used for live-conversation health
 * bands (PRD §4.8), but the `neutral` variant covers non-health uses
 * like priority chips.
 */
export function StatusPill({
  variant,
  label,
  className,
}: {
  variant: Variant;
  label: string;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium tracking-tight",
        variantClass[variant],
        className
      )}
    >
      <span className={cn("size-1.5 rounded-full", dotClass[variant])} />
      {label}
    </span>
  );
}
