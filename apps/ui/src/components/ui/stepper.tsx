import { Check } from "lucide-react";
import { cn } from "@/lib/cn";

export type StepperItem = {
  id: number;
  label: string;
};

type Props = {
  steps: StepperItem[];
  current: number;
  completed: number[];
  onStepClick?: (id: number) => void;
  className?: string;
};

/**
 * Linear stepper. Filled accent + check = done. Filled accent ring =
 * current. Neutral outline = future. Future dots are non-clickable;
 * done dots jump back to that step via `onStepClick`.
 */
export function Stepper({ steps, current, completed, onStepClick, className }: Props) {
  return (
    <ol className={cn("flex w-full items-start", className)}>
      {steps.map((s, i) => {
        const isDone = completed.includes(s.id) && s.id !== current;
        const isCurrent = s.id === current;
        const isFuture = s.id > current && !completed.includes(s.id);
        const clickable = !isFuture && onStepClick;

        return (
          <li
            key={s.id}
            className={cn("flex flex-1 items-start", i === steps.length - 1 && "flex-none")}
          >
            <div className="flex flex-col items-center gap-2">
              <button
                type="button"
                disabled={!clickable}
                onClick={() => clickable && onStepClick(s.id)}
                aria-current={isCurrent ? "step" : undefined}
                className={cn(
                  "relative inline-flex size-8 items-center justify-center rounded-full text-xs font-semibold transition-all duration-300 ease-[var(--ease-out-soft)]",
                  isDone && "bg-accent text-bg shadow-sm",
                  isCurrent &&
                    "bg-bg-elevated text-fg ring-2 ring-accent ring-offset-2 ring-offset-bg",
                  isFuture && "border border-border bg-bg-elevated text-fg-subtle",
                  clickable && !isCurrent && "hover:scale-105",
                  !clickable && "cursor-not-allowed"
                )}
              >
                {isDone ? <Check className="size-3.5" strokeWidth={2.5} /> : s.id}
              </button>
              <span
                className={cn(
                  "text-[11px] font-medium tracking-tight",
                  isCurrent ? "text-fg" : "text-fg-subtle"
                )}
              >
                {s.label}
              </span>
            </div>
            {i < steps.length - 1 && (
              <div
                className={cn(
                  "mx-3 mt-4 h-px flex-1 transition-colors duration-500 ease-[var(--ease-out-soft)]",
                  completed.includes(s.id) ? "bg-accent" : "bg-border"
                )}
              />
            )}
          </li>
        );
      })}
    </ol>
  );
}
