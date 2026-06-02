import { motion } from "framer-motion";
import { Container } from "@/components/ui/container";
import { RevealStagger, revealItem } from "@/components/ui/reveal";
import { SectionHeading } from "@/components/ui/section-heading";
import { LANDING_LABELS } from "@/labels";
import { ParallaxIllustration } from "@/pages/landing-page/parts/parallax-illustration";
import { FlowDots } from "@/pages/landing-page/parts/illustrations";

export function OutcomesSection() {
  return (
    <section className="relative overflow-hidden py-28 sm:py-36">
      <ParallaxIllustration strength={-20} className="-right-10 top-1/2 h-32 w-[60rem] text-accent">
        <FlowDots />
      </ParallaxIllustration>
      <Container>
        <SectionHeading title={LANDING_LABELS.outcomes.title} />
        <RevealStagger className="mt-16 grid grid-cols-2 gap-px overflow-hidden rounded-3xl border border-border bg-border lg:grid-cols-4">
          {LANDING_LABELS.outcomes.items.map((item) => (
            <motion.div
              key={item.label}
              variants={revealItem}
              transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
              className="group relative flex h-full flex-col justify-between bg-bg-elevated p-8 transition-colors duration-500 ease-[var(--ease-out-soft)] hover:bg-bg"
            >
              <div className="text-5xl font-semibold tracking-tight text-fg sm:text-6xl">
                {item.metric}
              </div>
              <div className="mt-4 text-sm leading-relaxed text-fg-muted">{item.label}</div>
            </motion.div>
          ))}
        </RevealStagger>
      </Container>
    </section>
  );
}
