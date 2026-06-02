import { motion, useScroll, useTransform } from "framer-motion";
import { useRef } from "react";
import { ArrowUpRight, Play } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Container } from "@/components/ui/container";
import { LANDING_LABELS } from "@/labels";
import { HeroVisual } from "@/pages/landing-page/parts/hero-visual";

const ease = [0.22, 1, 0.36, 1] as const;

export function Hero() {
  const ref = useRef<HTMLElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start start", "end start"],
  });
  // Subtle parallax: the visual drifts up as you scroll past
  const visualY = useTransform(scrollYProgress, [0, 1], [0, -60]);
  const visualOpacity = useTransform(scrollYProgress, [0, 0.8], [1, 0.7]);

  return (
    <section ref={ref} className="grain relative overflow-hidden">
      <div className="absolute inset-0 -z-10 bg-ambient" />
      <Container size="wide" className="pt-24 pb-28 sm:pt-32 sm:pb-36">
        <div className="grid items-center gap-16 lg:grid-cols-[1.05fr_1fr] lg:gap-24">
          <div className="flex flex-col items-start text-left">
            <motion.h1
              initial={{ opacity: 0, y: 24 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 1, ease, delay: 0.05 }}
              className="text-balance text-5xl font-semibold tracking-[-0.025em] text-fg sm:text-6xl md:text-[68px] md:leading-[1.02]"
            >
              {LANDING_LABELS.hero.title}
            </motion.h1>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 1, ease, delay: 0.18 }}
              className="mt-7 max-w-xl text-pretty text-lg leading-relaxed text-fg-muted sm:text-xl"
            >
              {LANDING_LABELS.hero.subtitle}
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 1, ease, delay: 0.3 }}
              className="mt-10 flex flex-wrap items-center gap-3"
            >
              <Button size="lg" trailing={<ArrowUpRight className="size-4 transition-transform duration-500 ease-[var(--ease-out-soft)] group-hover/btn:translate-x-0.5 group-hover/btn:-translate-y-0.5" />}>
                {LANDING_LABELS.hero.primaryCta}
              </Button>
              <Button size="lg" variant="secondary" leading={<Play className="size-4" />}>
                {LANDING_LABELS.hero.secondaryCta}
              </Button>
            </motion.div>

            <motion.p
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 1, ease, delay: 0.5 }}
              className="mt-5 text-sm text-fg-subtle"
            >
              {LANDING_LABELS.hero.trustNote}
            </motion.p>
          </div>

          <motion.div style={{ y: visualY, opacity: visualOpacity }}>
            <HeroVisual />
          </motion.div>
        </div>
      </Container>
    </section>
  );
}
