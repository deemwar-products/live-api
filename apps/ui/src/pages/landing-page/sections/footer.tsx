import { ArrowUpRight } from "lucide-react";
import { Container } from "@/components/ui/container";
import { LANDING_LABELS, COMMON_LABELS } from "@/labels";

type FooterColumn = { title: string; links: readonly string[] };

export function Footer() {
  return (
    <footer className="border-t border-border bg-bg-elevated py-16">
      <Container>
        <div className="grid gap-10 md:grid-cols-[1.5fr_2fr]">
          <div>
            <div className="flex items-center gap-2 text-[15px] font-semibold tracking-tight text-fg">
              <span className="grid size-7 place-items-center rounded-lg bg-fg text-bg">
                <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden>
                  <path d="M2 7 L6 7 M8 4 L8 10 M10 5 L10 9 M12 3 L12 11" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
                </svg>
              </span>
              {COMMON_LABELS.app.name}
            </div>
            <p className="mt-4 max-w-sm text-sm text-fg-muted">{LANDING_LABELS.footer.tagline}</p>
          </div>
          <div className="grid grid-cols-2 gap-8 sm:grid-cols-4">
            {(Object.values(LANDING_LABELS.footer.columns) as FooterColumn[]).map((col) => (
              <div key={col.title}>
                <h4 className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">
                  {col.title}
                </h4>
                <ul className="mt-4 space-y-2.5">
                  {col.links.map((link) => (
                    <li key={link}>
                      <a
                        href="#"
                        className="group inline-flex items-center gap-1 text-sm text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:text-fg"
                      >
                        {link}
                        <ArrowUpRight className="size-3 opacity-0 transition-all duration-300 ease-[var(--ease-out-soft)] group-hover:opacity-60" />
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>
        <div className="mt-12 flex flex-col items-start justify-between gap-3 border-t border-border pt-6 text-xs text-fg-subtle sm:flex-row sm:items-center">
          <span>{LANDING_LABELS.footer.copyright}</span>
          <span>Made for support teams that measure outcomes.</span>
        </div>
      </Container>
    </footer>
  );
}
