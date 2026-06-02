import { LANDING_LABELS } from "@/labels";

/**
 * A slow, infinite marquee of product vocabulary. Replaces the trust strip.
 * No fake logos — just calm, factual phrases that describe what the product
 * does, in motion. Calmer than a banner, more useful than a logo wall.
 */
export function Ticker() {
  const items = LANDING_LABELS.ticker;

  return (
    <section
      aria-label="Live API capabilities"
      className="relative overflow-hidden border-y border-border bg-bg-elevated py-7"
    >
      <div
        className="pointer-events-none absolute inset-y-0 left-0 z-10 w-24 bg-gradient-to-r from-bg-elevated to-transparent"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute inset-y-0 right-0 z-10 w-24 bg-gradient-to-l from-bg-elevated to-transparent"
        aria-hidden
      />
      <div className="flex w-max animate-[ticker_40s_linear_infinite] gap-12 px-6 will-change-transform">
        {[...items, ...items].map((item, i) => (
          <div key={i} className="flex shrink-0 items-center gap-12">
            <span className="text-sm font-medium tracking-tight text-fg-muted">{item}</span>
            <span className="size-1.5 rounded-full bg-border-strong" aria-hidden />
          </div>
        ))}
      </div>
    </section>
  );
}
