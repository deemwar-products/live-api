import { ArrowUpRight } from "lucide-react";
import { Container } from "@/components/ui/container";
import { Reveal } from "@/components/ui/reveal";
import { Button } from "@/components/ui/button";
import { LANDING_LABELS } from "@/labels";

export function CtaSection() {
  return (
    <section className="py-24 sm:py-28">
      <Container size="default">
        <Reveal>
          <div className="grain relative overflow-hidden rounded-3xl border border-border bg-fg p-12 text-center text-bg sm:p-16">
            <div className="absolute inset-0 -z-0 bg-ambient opacity-30 mix-blend-overlay" />
            <div className="relative">
              <h2 className="text-balance text-3xl font-semibold tracking-tight sm:text-4xl md:text-[48px] md:leading-[1.05]">
                {LANDING_LABELS.cta.title}
              </h2>
              <p className="mx-auto mt-5 max-w-xl text-pretty text-base opacity-80 sm:text-lg">
                {LANDING_LABELS.cta.body}
              </p>
              <div className="mt-10 flex flex-wrap items-center justify-center gap-3">
                <Button
                  size="lg"
                  tone="inverted"
                  trailing={
                    <ArrowUpRight className="size-4 transition-transform duration-500 ease-[var(--ease-out-soft)] group-hover/btn:translate-x-0.5 group-hover/btn:-translate-y-0.5" />
                  }
                >
                  {LANDING_LABELS.cta.primary}
                </Button>
                <Button size="lg" tone="inverted" variant="ghost">
                  {LANDING_LABELS.cta.secondary}
                </Button>
              </div>
            </div>
          </div>
        </Reveal>
      </Container>
    </section>
  );
}
