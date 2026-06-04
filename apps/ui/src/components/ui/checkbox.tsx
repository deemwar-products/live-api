import { Check } from "lucide-react";
import { useState, type ReactNode } from "react";
import { cn } from "@/lib/cn";

type Props = {
  checked: boolean;
  onChange: (next: boolean) => void;
  label: ReactNode;
  description?: ReactNode;
  disabled?: boolean;
  className?: string;
};

/**
 * Single checkbox row. Square indicator on the left, label + optional
 * description on the right. Click anywhere on the row to toggle.
 */
export function Checkbox({ checked, onChange, label, description, disabled, className }: Props) {
  const [focused, setFocused] = useState(false);
  return (
    <label
      className={cn(
        "group flex cursor-pointer items-start gap-3 rounded-xl border border-border bg-bg-elevated p-3.5 transition-all duration-300 ease-[var(--ease-out-soft)]",
        "hover:border-border-strong hover:bg-bg-muted/50",
        checked && "border-accent/40 bg-accent-soft/30",
        disabled && "cursor-not-allowed opacity-60",
        className
      )}
    >
      <span
        className={cn(
          "relative mt-0.5 inline-flex size-4 shrink-0 items-center justify-center rounded-md border transition-all duration-300 ease-[var(--ease-out-soft)]",
          checked
            ? "border-accent bg-accent text-bg"
            : "border-border-strong bg-bg-elevated text-transparent",
          focused && "ring-2 ring-accent ring-offset-2 ring-offset-bg"
        )}
      >
        <Check className="size-3" strokeWidth={3} />
      </span>
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        className="sr-only"
      />
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium text-fg">{label}</div>
        {description && <div className="mt-0.5 text-xs text-fg-muted">{description}</div>}
      </div>
    </label>
  );
}
