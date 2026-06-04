import { Link } from "react-router-dom";
import type { ReactNode } from "react";
import { Stepper, type StepperItem } from "@/components/ui/stepper";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { Button } from "@/components/ui/button";
import { ONBOARDING_LABELS } from "@/labels";
import { LANDING } from "@/lib/paths";

type Props = {
  children: ReactNode;
  steps: StepperItem[];
  current: number;
  completed: number[];
  onStepClick?: (id: number) => void;
  onSaveAndExit?: () => void;
};

/**
 * Lighter shell for the onboarding flow. No sidebar — single column,
 * stepper across the top, save-and-exit in the corner. The same brand
 * mark as the console shell so it reads as part of the same product.
 */
export function OnboardingShell({
  children,
  steps,
  current,
  completed,
  onStepClick,
  onSaveAndExit,
}: Props) {
  return (
    <div className="grain relative flex min-h-svh flex-col bg-bg">
      <div className="absolute inset-0 -z-10 bg-ambient opacity-40" />

      <header className="sticky top-0 z-20 flex h-16 items-center gap-3 border-b border-border/60 bg-bg/70 px-4 backdrop-blur-xl sm:px-6">
        <Link
          to={LANDING.home}
          className="flex items-center gap-2 text-[15px] font-semibold tracking-tight text-fg"
        >
          <LogoMark />
          <span>{ONBOARDING_LABELS.shell.brand}</span>
        </Link>

        <p className="ml-3 hidden text-sm font-medium tracking-tight text-fg-muted md:block">
          {ONBOARDING_LABELS.shell.title}
        </p>

        <div className="ml-auto flex items-center gap-2">
          {onSaveAndExit && (
            <Button variant="ghost" size="sm" onClick={onSaveAndExit}>
              {ONBOARDING_LABELS.shell.saveAndExit}
            </Button>
          )}
          <ThemeToggle />
        </div>
      </header>

      <div className="border-b border-border/60 bg-bg-elevated/40 px-4 py-5 backdrop-blur-sm sm:px-6">
        <div className="mx-auto w-full max-w-3xl">
          <Stepper
            steps={steps}
            current={current}
            completed={completed}
            onStepClick={onStepClick}
          />
        </div>
      </div>

      <main className="min-w-0 flex-1">
        <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-10">{children}</div>
      </main>
    </div>
  );
}

function LogoMark() {
  return (
    <span className="relative inline-flex size-7 items-center justify-center overflow-hidden rounded-lg bg-fg text-bg">
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden>
        <path
          d="M2 7 L6 7 M8 4 L8 10 M10 5 L10 9 M12 3 L12 11"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
        />
      </svg>
    </span>
  );
}
