import { useId, useState } from "react";
import { motion } from "framer-motion";
import { Link } from "react-router-dom";
import { LANDING } from "@/lib/paths";
import { AUTH_LABELS } from "@/pages/auth-page/labels";
import { MarketingHeader } from "@/components/layout/marketing-header";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { cn } from "@/lib/cn";

const ease = [0.22, 1, 0.36, 1] as const;

export function AuthPage() {
  return (
    <main className="grain relative flex h-svh flex-col overflow-hidden">
      <div className="absolute inset-0 -z-10 bg-ambient opacity-70" />

      <MarketingHeader
        rightSlot={
          <>
            <ThemeToggle />
            <Link
              to={LANDING.home}
              className="rounded-full px-3.5 py-1.5 text-sm text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:text-fg"
            >
              {AUTH_LABELS.footer.backToHome}
            </Link>
          </>
        }
      />

      <div className="flex flex-1 items-center justify-center overflow-hidden px-6 py-6">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, ease }}
          className="w-full max-w-md"
        >
          <div className="rounded-3xl border border-border bg-bg-elevated/80 p-7 shadow-xl backdrop-blur-sm sm:p-8">
            <h1 className="text-balance text-2xl font-semibold tracking-tight text-fg">
              {AUTH_LABELS.meta.title}
            </h1>
            <p className="mt-1.5 text-sm leading-relaxed text-fg-muted">
              {AUTH_LABELS.meta.description}
            </p>

            <div className="mt-6 space-y-3">
              <ProviderGroup label={AUTH_LABELS.providers.personal}>
                <GoogleButton />
                <MicrosoftButton />
              </ProviderGroup>

              <ProviderGroup label={AUTH_LABELS.providers.team}>
                <SsoButton />
              </ProviderGroup>
            </div>

            <p className="mt-6 text-xs leading-relaxed text-fg-subtle">
              {AUTH_LABELS.legal.termsPrefix}{" "}
              <a href="#" className="underline underline-offset-2 hover:text-fg-muted">
                {AUTH_LABELS.legal.terms}
              </a>{" "}
              {AUTH_LABELS.legal.and}{" "}
              <a href="#" className="underline underline-offset-2 hover:text-fg-muted">
                {AUTH_LABELS.legal.privacy}
              </a>
              .
            </p>
          </div>

          <p className="mt-4 text-center text-xs text-fg-subtle">
            {AUTH_LABELS.brand.tagline}
          </p>
        </motion.div>
      </div>
    </main>
  );
}

function ProviderGroup({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-2.5">
      <div className="flex items-center gap-3">
        <span className="text-[10px] font-semibold uppercase tracking-wider text-fg-subtle">
          {label}
        </span>
        <span className="h-px flex-1 bg-border" />
      </div>
      <div className="space-y-2.5">{children}</div>
    </div>
  );
}

function GoogleButton() {
  return (
    <button
      type="button"
      className="group flex w-full items-center justify-center gap-3 rounded-full border border-border bg-bg-elevated px-5 py-3 text-sm font-medium text-fg shadow-sm transition-all duration-300 ease-[var(--ease-out-soft)] hover:border-border-strong hover:bg-bg-muted hover:shadow-md active:scale-[0.99]"
    >
      <GoogleMark />
      <span>{AUTH_LABELS.providers.google.label}</span>
    </button>
  );
}

function MicrosoftButton() {
  return (
    <button
      type="button"
      className="group flex w-full items-center justify-center gap-3 rounded-full border border-border bg-bg-elevated px-5 py-3 text-sm font-medium text-fg shadow-sm transition-all duration-300 ease-[var(--ease-out-soft)] hover:border-border-strong hover:bg-bg-muted hover:shadow-md active:scale-[0.99]"
    >
      <MicrosoftMark />
      <span>{AUTH_LABELS.providers.microsoft.label}</span>
    </button>
  );
}

function SsoButton() {
  const [hovered, setHovered] = useState(false);
  const [focused, setFocused] = useState(false);
  const tooltipId = useId();
  const showTooltip = hovered || focused;

  return (
    <div className="relative">
      <button
        type="button"
        disabled
        aria-describedby={showTooltip ? tooltipId : undefined}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        className="group flex w-full items-center justify-center gap-3 rounded-full border border-border bg-bg-muted/40 px-5 py-3 text-sm font-medium text-fg-subtle opacity-60 shadow-sm transition-all duration-300 ease-[var(--ease-out-soft)] hover:opacity-80"
      >
        <KeyMark />
        <span>{AUTH_LABELS.providers.sso.label}</span>
        <span className="ml-2 rounded-full border border-border bg-bg-elevated px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-fg-subtle">
          Soon
        </span>
      </button>

      <div
        id={tooltipId}
        role="tooltip"
        className={cn(
          "pointer-events-none absolute left-1/2 top-full z-20 mt-2 -translate-x-1/2 whitespace-nowrap rounded-full border border-border bg-bg-elevated px-3 py-1.5 text-xs font-medium text-fg shadow-md transition-all duration-200 ease-[var(--ease-out-soft)]",
          showTooltip ? "translate-y-0 opacity-100" : "-translate-y-1 opacity-0"
        )}
      >
        {AUTH_LABELS.providers.sso.comingSoonNote}
        <span
          aria-hidden
          className="absolute -top-1 left-1/2 size-2 -translate-x-1/2 rotate-45 border-l border-t border-border bg-bg-elevated"
        />
      </div>
    </div>
  );
}

function GoogleMark() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden>
      <path
        fill="#4285F4"
        d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.79 2.72v2.26h2.9c1.7-1.56 2.69-3.86 2.69-6.62z"
      />
      <path
        fill="#34A853"
        d="M9 18c2.43 0 4.47-.81 5.96-2.18l-2.9-2.26c-.81.54-1.83.86-3.06.86-2.35 0-4.34-1.59-5.05-3.72H.96v2.34A9 9 0 0 0 9 18z"
      />
      <path
        fill="#FBBC05"
        d="M3.95 10.7A5.4 5.4 0 0 1 3.66 9c0-.59.1-1.16.29-1.7V4.96H.96A9 9 0 0 0 0 9c0 1.45.35 2.83.96 4.04l2.99-2.34z"
      />
      <path
        fill="#EA4335"
        d="M9 3.58c1.32 0 2.5.45 3.44 1.35l2.58-2.58A9 9 0 0 0 .96 4.96L3.95 7.3C4.66 5.16 6.65 3.58 9 3.58z"
      />
    </svg>
  );
}

function MicrosoftMark() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden>
      <rect width="8" height="8" fill="#F25022" />
      <rect x="10" width="8" height="8" fill="#7FBA00" />
      <rect y="10" width="8" height="8" fill="#00A4EF" />
      <rect x="10" y="10" width="8" height="8" fill="#FFB900" />
    </svg>
  );
}

function KeyMark() {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 18 18"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <circle cx="7" cy="11" r="3" />
      <path d="M9.2 8.8 16 2" />
      <path d="m13 5 2 2" />
      <path d="m15 3 2 2" />
    </svg>
  );
}
