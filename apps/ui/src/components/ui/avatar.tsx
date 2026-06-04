import { cn } from "@/lib/cn";

type Props = {
  name: string;
  className?: string;
};

/**
 * Initials avatar — two letters from the start of a name. Used for
 * invite rows and any other "small portrait" spot. Color stays neutral;
 * differentiation comes from text, not background hue.
 */
export function Avatar({ name, className }: Props) {
  const initials = name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? "")
    .join("");

  return (
    <span
      className={cn(
        "inline-flex size-8 shrink-0 items-center justify-center rounded-full bg-bg-muted text-[11px] font-semibold tracking-tight text-fg-muted",
        className
      )}
      aria-hidden
    >
      {initials || "·"}
    </span>
  );
}
