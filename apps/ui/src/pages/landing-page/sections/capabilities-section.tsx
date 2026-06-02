import { Mic, FileText, Workflow, Activity, Tag, ShieldCheck } from "lucide-react";
import { motion } from "framer-motion";
import { Container } from "@/components/ui/container";
import { RevealStagger, revealItem } from "@/components/ui/reveal";
import { SectionHeading } from "@/components/ui/section-heading";
import { LANDING_LABELS } from "@/labels";
import { ParallaxIllustration } from "@/pages/landing-page/parts/parallax-illustration";
import { ConcentricRings } from "@/pages/landing-page/parts/illustrations";
import type { LucideIcon } from "lucide-react";

const ICONS: LucideIcon[] = [Mic, FileText, Workflow, Activity, Tag, ShieldCheck];

export function CapabilitiesSection() {
  return (
    <section id="product" className="relative overflow-hidden py-28 sm:py-36">
      <ParallaxIllustration strength={40} className="-left-32 top-1/3 h-80 w-80 text-fg-subtle">
        <ConcentricRings />
      </ParallaxIllustration>
      <Container>
        <SectionHeading title={LANDING_LABELS.capabilities.title} />
        <RevealStagger className="mt-16 grid gap-px overflow-hidden rounded-3xl border border-border bg-border sm:grid-cols-2 lg:grid-cols-3">
          {LANDING_LABELS.capabilities.items.map((item, i) => {
            const Icon = ICONS[i] ?? Mic;
            return (
              <motion.div
                key={item.title}
                variants={revealItem}
                transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
                className="group relative h-full bg-bg-elevated p-8 transition-colors duration-500 ease-[var(--ease-out-soft)] hover:bg-bg"
              >
                <div
                  aria-hidden
                  className="pointer-events-none absolute -top-12 -right-12 size-32 rounded-full bg-accent-soft opacity-0 blur-2xl transition-opacity duration-700 ease-[var(--ease-out-soft)] group-hover:opacity-50"
                />
                <div className="relative">
                  <Icon className="mb-6 size-5 text-fg" strokeWidth={1.5} />
                  <h3 className="text-base font-semibold tracking-tight text-fg">{item.title}</h3>
                  <p className="mt-2.5 text-sm leading-relaxed text-fg-muted">{item.body}</p>
                </div>
              </motion.div>
            );
          })}
        </RevealStagger>
      </Container>
    </section>
  );
}
