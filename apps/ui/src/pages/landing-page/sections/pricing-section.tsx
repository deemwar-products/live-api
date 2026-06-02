import { ArrowUpRight } from "lucide-react";
import { Container } from "@/components/ui/container";
import { Reveal } from "@/components/ui/reveal";
import { SectionHeading } from "@/components/ui/section-heading";
import { Button } from "@/components/ui/button";
import { LANDING_LABELS, COMMON_LABELS } from "@/labels";

export function PricingSection() {
  return (
    <section id="pricing" className="py-28 sm:py-36">
      <Container size="narrow">
        <SectionHeading title={LANDING_LABELS.pricing.title} description={LANDING_LABELS.pricing.body} />
        <Reveal delay={0.1}>
          <div className="mt-14 flex flex-col items-center gap-6 rounded-3xl border border-border bg-bg-elevated p-10 text-center shadow-md sm:p-14">
            <p className="max-w-md text-sm text-fg-muted">{LANDING_LABELS.pricing.note}</p>
            <div className="flex flex-wrap items-center justify-center gap-3">
              <Button size="lg" trailing={<ArrowUpRight className="size-4 transition-transform duration-300 ease-[var(--ease-out-soft)] group-hover/btn:translate-x-0.5 group-hover/btn:-translate-y-0.5" />}>
                {LANDING_LABELS.pricing.cta}
              </Button>
              <Button size="lg" variant="secondary">
                {COMMON_LABELS.actions.contactSales}
              </Button>
            </div>
          </div>
        </Reveal>
      </Container>
    </section>
  );
}
