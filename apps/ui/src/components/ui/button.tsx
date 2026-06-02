import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/cn";

type Variant = "primary" | "secondary" | "ghost";
type Size = "sm" | "md" | "lg";
/**
 * `tone` flips the button for use on inverted surfaces (e.g. a dark
 * card on a light page, or a light card on a dark page). Components
 * decide based on their own context — pass `tone="inverted"` when the
 * button sits on a surface that has inverted colors.
 */
type Tone = "default" | "inverted";

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant;
  size?: Size;
  tone?: Tone;
  leading?: ReactNode;
  trailing?: ReactNode;
};

const baseClass =
  "group/btn relative inline-flex cursor-pointer items-center justify-center gap-2 rounded-full font-medium tracking-tight " +
  "transition-[background-color,color,border-color,transform,box-shadow] duration-300 ease-[var(--ease-out-soft)] " +
  "active:scale-[0.97] disabled:opacity-50 disabled:pointer-events-none " +
  "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg";

const variantClass: Record<Tone, Record<Variant, string>> = {
  default: {
    primary:
      "bg-fg text-bg shadow-md shadow-fg/10 hover:shadow-lg hover:shadow-fg/15",
    secondary:
      "bg-bg-elevated text-fg border border-border shadow-sm hover:bg-bg-muted hover:border-border-strong",
    ghost: "bg-transparent text-fg hover:bg-bg-muted",
  },
  inverted: {
    // Use on a bg-fg (dark) card. Primary is light, secondary is dark-with-border, ghost is light.
    primary:
      "bg-bg-elevated text-fg shadow-md shadow-fg/20 hover:shadow-lg hover:shadow-fg/30",
    secondary:
      "bg-transparent text-fg-inverse border border-fg-inverse/20 hover:bg-fg-inverse/10 hover:border-fg-inverse/40",
    ghost: "bg-transparent text-fg-inverse hover:bg-fg-inverse/10",
  },
};

const sizeClass: Record<Size, string> = {
  sm: "h-9 px-3.5 text-sm",
  md: "h-11 px-5 text-sm",
  lg: "h-12 px-6 text-[15px]",
};

export function Button({
  variant = "primary",
  size = "md",
  tone = "default",
  leading,
  trailing,
  className,
  children,
  ...rest
}: Props) {
  return (
    <button
      type={rest.type ?? "button"}
      className={cn(baseClass, variantClass[tone][variant], sizeClass[size], className)}
      {...rest}
    >
      {leading}
      {children}
      {trailing}
    </button>
  );
}
