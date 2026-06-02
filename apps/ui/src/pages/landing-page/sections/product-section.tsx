import { BookOpen, MessageCircle, UserRound } from "lucide-react";
import { motion } from "framer-motion";
import { Container } from "@/components/ui/container";
import { RevealStagger, revealItem } from "@/components/ui/reveal";
import { SectionHeading } from "@/components/ui/section-heading";
import { LANDING_LABELS } from "@/labels";
import { cn } from "@/lib/cn";
import { ParallaxIllustration } from "@/pages/landing-page/parts/parallax-illustration";
import { KnowledgeGraph } from "@/pages/landing-page/parts/illustrations";

const STEPS = [
  { Icon: BookOpen, kicker: "Step 1" },
  { Icon: MessageCircle, kicker: "Step 2" },
  { Icon: UserRound, kicker: "Step 3" },
];

export function ProductSection() {
  return (
    <section id="how-it-works" className="grain relative overflow-hidden py-28 sm:py-36">
      <div className="absolute inset-0 -z-10 bg-ambient opacity-50" />
      <ParallaxIllustration strength={-30} className="-right-20 -top-20 h-72 w-[28rem] text-accent">
        <KnowledgeGraph />
      </ParallaxIllustration>
      <Container>
        <SectionHeading title={LANDING_LABELS.product.title} />
        <div className="relative mt-20">
          {/* connector line */}
          <div
            aria-hidden
            className="pointer-events-none absolute left-0 right-0 top-12 hidden h-px bg-gradient-to-r from-transparent via-border-strong to-transparent md:block"
          />
          <RevealStagger className="grid gap-10 md:grid-cols-3">
            {LANDING_LABELS.product.steps.map((step, i) => {
              const { Icon } = STEPS[i] ?? STEPS[0];
              return (
                <motion.div
                  key={step.title}
                  variants={revealItem}
                  transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1] }}
                  className="flex flex-col items-start"
                >
                  <div
                    className={cn(
                      "relative z-10 mb-7 grid size-24 place-items-center rounded-2xl",
                      "border border-border bg-bg-elevated shadow-md",
                      "transition-shadow duration-500 ease-[var(--ease-out-soft)] hover:shadow-xl"
                    )}
                  >
                    <Icon className="size-7 text-fg" strokeWidth={1.5} />
                    <span className="absolute -top-2 -right-2 rounded-full bg-fg px-2 py-0.5 text-[10px] font-semibold tracking-wide text-bg">
                      {String(i + 1).padStart(2, "0")}
                    </span>
                  </div>
                  <h3 className="text-xl font-semibold tracking-tight text-fg">{step.title}</h3>
                  <p className="mt-2.5 max-w-sm text-[15px] leading-relaxed text-fg-muted">{step.body}</p>
                </motion.div>
              );
            })}
          </RevealStagger>
        </div>
      </Container>
    </section>
  );
}
