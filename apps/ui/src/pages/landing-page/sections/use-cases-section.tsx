import { motion } from "framer-motion";
import { Container } from "@/components/ui/container";
import { RevealStagger, revealItem } from "@/components/ui/reveal";
import { SectionHeading } from "@/components/ui/section-heading";
import { LANDING_LABELS } from "@/labels";

export function UseCasesSection() {
  return (
    <section id="use-cases" className="grain relative overflow-hidden py-28 sm:py-36">
      <div className="absolute inset-0 -z-10 bg-ambient opacity-50" />
      <Container>
        <SectionHeading title={LANDING_LABELS.useCases.title} />
        <RevealStagger className="mt-16 grid gap-5 sm:grid-cols-2">
          {LANDING_LABELS.useCases.items.map((item, i) => (
            <motion.article
              key={item.title}
              variants={revealItem}
              transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
              className="group flex h-full flex-col justify-between rounded-2xl border border-border bg-bg-elevated p-8 transition-all duration-500 ease-[var(--ease-out-soft)] hover:-translate-y-1 hover:border-border-strong hover:shadow-lg"
            >
              <div>
                <div className="mb-5 inline-flex items-center gap-2 rounded-full bg-bg-muted px-2.5 py-1 text-[11px] font-medium uppercase tracking-wider text-fg-subtle">
                  0{i + 1}
                </div>
                <h3 className="text-xl font-semibold tracking-tight text-fg">{item.title}</h3>
                <p className="mt-2.5 text-[15px] leading-relaxed text-fg-muted">{item.body}</p>
              </div>
              <div className="mt-6 h-px w-full bg-gradient-to-r from-border via-border to-transparent" />
            </motion.article>
          ))}
        </RevealStagger>
      </Container>
    </section>
  );
}
