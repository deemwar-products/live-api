import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

type Props = InputHTMLAttributes<HTMLInputElement>;

/**
 * Plain text input. Rounded-full to match the rest of the surface
 * language; same border + focus-accent treatment as the topbar search.
 */
export const Input = forwardRef<HTMLInputElement, Props>(function Input(
  { className, ...rest },
  ref
) {
  return (
    <input
      ref={ref}
      className={cn(
        "h-10 w-full rounded-full border border-border bg-bg-elevated px-4 text-sm text-fg placeholder:text-fg-subtle",
        "transition-colors duration-300 ease-[var(--ease-out-soft)]",
        "hover:border-border-strong focus:border-accent focus:outline-none",
        className
      )}
      {...rest}
    />
  );
});
