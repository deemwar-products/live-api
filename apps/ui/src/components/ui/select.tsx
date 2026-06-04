import { ChevronDown } from "lucide-react";
import { useState, type ReactNode } from "react";
import { cn } from "@/lib/cn";

type Option<V extends string> = { value: V; label: ReactNode };

type Props<V extends string> = {
  value: V;
  onChange: (next: V) => void;
  options: Option<V>[];
  className?: string;
  ariaLabel?: string;
};

/**
 * Simple presentational select. Click toggles a panel; click outside
 * or pick an option closes it. No keyboard navigation beyond native
 * — kept minimal because this is a marketing/console surface, not
 * a data grid.
 */
export function Select<V extends string>({ value, onChange, options, className, ariaLabel }: Props<V>) {
  const [open, setOpen] = useState(false);
  const current = options.find((o) => o.value === value) ?? options[0];

  return (
    <div className={cn("relative inline-block w-full", className)}>
      <button
        type="button"
        aria-label={ariaLabel}
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => setOpen((o) => !o)}
        onBlur={() => setTimeout(() => setOpen(false), 120)}
        className={cn(
          "flex h-10 w-full items-center justify-between gap-2 rounded-full border border-border bg-bg-elevated px-4 text-sm text-fg",
          "transition-colors duration-300 ease-[var(--ease-out-soft)]",
          "hover:border-border-strong focus:border-accent focus:outline-none"
        )}
      >
        <span className="truncate">{current?.label}</span>
        <ChevronDown
          className={cn(
            "size-3.5 shrink-0 text-fg-subtle transition-transform duration-300 ease-[var(--ease-out-soft)]",
            open && "rotate-180"
          )}
        />
      </button>
      {open && (
        <ul
          role="listbox"
          className="absolute left-0 right-0 top-full z-30 mt-1.5 overflow-hidden rounded-xl border border-border bg-bg-elevated p-1 shadow-lg shadow-fg/[0.04]"
        >
          {options.map((o) => {
            const selected = o.value === value;
            return (
              <li key={o.value} role="option" aria-selected={selected}>
                <button
                  type="button"
                  onMouseDown={(e) => {
                    e.preventDefault();
                    onChange(o.value);
                    setOpen(false);
                  }}
                  className={cn(
                    "flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors duration-200 ease-[var(--ease-out-soft)]",
                    "hover:bg-bg-muted",
                    selected ? "text-fg" : "text-fg-muted"
                  )}
                >
                  <span className="truncate">{o.label}</span>
                  {selected && <span className="size-1.5 shrink-0 rounded-full bg-accent" />}
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
