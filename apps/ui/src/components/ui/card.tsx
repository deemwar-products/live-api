import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/cn";

/**
 * Surface card — the workhorse container for dashboard panels. Sits on
 * `--color-bg` (the page background) and uses the same surface language
 * as the landing-page outcomes section: elevated, lightly translucent,
 * backdrop-blurred, soft border.
 */
export function Card({
  children,
  className,
  ...rest
}: HTMLAttributes<HTMLDivElement> & { children: ReactNode }) {
  return (
    <div
      className={cn(
        "rounded-2xl border border-border bg-bg-elevated/60 shadow-sm shadow-fg/[0.02] backdrop-blur-sm",
        className
      )}
      {...rest}
    >
      {children}
    </div>
  );
}

export function CardHeader({
  title,
  caption,
  action,
  className,
}: {
  title: ReactNode;
  caption?: ReactNode;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex items-start justify-between gap-4 border-b border-border/60 px-5 py-4",
        className
      )}
    >
      <div className="min-w-0">
        <h3 className="truncate text-sm font-semibold tracking-tight text-fg">{title}</h3>
        {caption && <p className="mt-0.5 text-xs text-fg-subtle">{caption}</p>}
      </div>
      {action && <div className="shrink-0">{action}</div>}
    </div>
  );
}

export function CardBody({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return <div className={cn("p-5", className)}>{children}</div>;
}
