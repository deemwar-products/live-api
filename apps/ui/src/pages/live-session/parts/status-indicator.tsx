/**
 * Live session status pill. Reuses the surface language of the existing
 * StatusPill component (rounded-full, leading dot, soft tinted bg) but
 * with a small pulsing dot for the "live" state so the running session
 * reads as alive at a glance.
 */

import { cn } from "@/lib/cn";
import type { StatusState } from "@/lib/live/protocol";

const toneClass: Record<StatusState, string> = {
 connecting: "bg-warning-soft text-warning",
 live: "bg-success-soft text-success",
 interrupted: "bg-watch-soft text-watch",
 ended: "bg-bg-muted text-fg-muted",
};

const dotClass: Record<StatusState, string> = {
 connecting: "bg-warning",
 live: "bg-success",
 interrupted: "bg-watch",
 ended: "bg-fg-subtle",
};

export function StatusIndicator({
 state,
 label,
 className,
}: {
 state: StatusState | "idle";
 label: string;
 className?: string;
}) {
 const tone = state === "idle" ? "bg-bg-muted text-fg-muted" : toneClass[state];
 const dot = state === "idle" ? "bg-fg-subtle" : dotClass[state];
 return (
 <span
 className={cn(
 "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium tracking-tight",
 tone,
 className,
 )}
 >
 <span
 className={cn(
 "size-1.5 rounded-full",
 dot,
 state === "live" && "animate-pulse",
 )}
 />
 {label}
 </span>
 );
}
