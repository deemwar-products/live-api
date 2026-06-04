import { forwardRef, type TextareaHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

type Props = TextareaHTMLAttributes<HTMLTextAreaElement>;

/**
 * Multi-line text input. Same surface treatment as `Input` but with
 * a larger radius (xl) and a min-height default for prompt bodies.
 */
export const Textarea = forwardRef<HTMLTextAreaElement, Props>(function Textarea(
  { className, rows = 4, ...rest },
  ref
) {
  return (
    <textarea
      ref={ref}
      rows={rows}
      className={cn(
        "w-full rounded-xl border border-border bg-bg-elevated px-4 py-3 text-sm text-fg placeholder:text-fg-subtle",
        "transition-colors duration-300 ease-[var(--ease-out-soft)]",
        "hover:border-border-strong focus:border-accent focus:outline-none",
        "resize-y leading-relaxed",
        className
      )}
      {...rest}
    />
  );
});
