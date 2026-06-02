import { Clock, Wrench, TrendingUp } from "lucide-react";
import { Container } from "@/components/ui/container";
import { Reveal, RevealStagger, revealItem } from "@/components/ui/reveal";
import { SectionHeading } from "@/components/ui/section-heading";
import { LANDING_LABELS } from "@/labels";
import { motion } from "framer-motion";

const ICONS = [Clock, Wrench, TrendingUp];

export function ProblemSection() {
  return (
    <section className="py-28 sm:py-36">
      <Container>
        <SectionHeading title={LANDING_LABELS.problem.title} />
        <RevealStagger className="mt-16 grid gap-5 md:grid-cols-3">
          {LANDING_LABELS.problem.items.map((item, i) => {
            const Icon = ICONS[i] ?? Clock;
            return (
              <motion.article
                key={item.title}
                variants={revealItem}
                transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1] }}
                className="group relative h-full overflow-hidden rounded-2xl border border-border bg-bg-elevated p-8 transition-all duration-500 ease-[var(--ease-out-soft)] hover:-translate-y-1 hover:border-border-strong hover:shadow-lg"
              >
                {/* Subtle hover glow */}
                <div
                  aria-hidden
                  className="pointer-events-none absolute -top-20 -right-20 size-40 rounded-full bg-accent-soft opacity-0 blur-3xl transition-opacity duration-700 ease-[var(--ease-out-soft)] group-hover:opacity-60"
                />
                <div className="relative">
                  <div className="mb-6 inline-flex size-11 items-center justify-center rounded-xl bg-bg-muted text-fg-muted transition-all duration-500 ease-[var(--ease-out-soft)] group-hover:bg-accent group-hover:text-neutral-0">
                    <Icon className="size-5" strokeWidth={1.75} />
                  </div>
                  <h3 className="text-lg font-semibold tracking-tight text-fg">{item.title}</h3>
                  <p className="mt-2.5 text-[15px] leading-relaxed text-fg-muted">{item.body}</p>
                </div>
              </motion.article>
            );
          })}
        </RevealStagger>
        <Reveal delay={0.2} className="mt-12 text-center">
          <p className="text-sm text-fg-subtle">
            Every minute in the queue is a minute a customer is shopping around.
          </p>
        </Reveal>
      </Container>
    </section>
  );
}
