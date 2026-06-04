import { X } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "@/lib/cn";

type Props = {
  children: ReactNode;
  onRemove?: () => void;
  className?: string;
};

/**
 * Small pill used for languages, doc chips, and the like. Optional
 * remove button. The accent variant is reserved for "selected" states
 * the user can clear.
 */
export function Tag({ children, onRemove, className }: Props) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border border-border bg-bg-elevated px-2.5 py-1 text-xs font-medium text-fg",
        className
      )}
    >
      {children}
      {onRemove && (
        <button
          type="button"
          onClick={onRemove}
          aria-label="Remove"
          className="inline-flex size-4 items-center justify-center rounded-full text-fg-subtle transition-colors duration-200 hover:bg-bg-muted hover:text-fg"
        >
          <X className="size-2.5" />
        </button>
      )}
    </span>
  );
}
