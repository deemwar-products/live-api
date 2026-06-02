import { useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { Play, Pause, Sparkles } from "lucide-react";
import { Container } from "@/components/ui/container";
import { Reveal } from "@/components/ui/reveal";
import { SectionHeading } from "@/components/ui/section-heading";
import { Button } from "@/components/ui/button";
import { LANDING_LABELS } from "@/labels";
import { DEMO_TRANSCRIPT } from "@/pages/landing-page/mocks";
import { cn } from "@/lib/cn";

const ease = [0.22, 1, 0.36, 1] as const;

export function LiveDemoSection() {
  const [playing, setPlaying] = useState(false);
  const [step, setStep] = useState(0);

  const toggle = () => {
    if (playing) {
      setPlaying(false);
      return;
    }
    setPlaying(true);
    setStep(0);
    const id = setInterval(() => {
      setStep((s) => {
        const next = s + 1;
        if (next >= DEMO_TRANSCRIPT.length) {
          setPlaying(false);
          clearInterval(id);
          return 0;
        }
        return next;
      });
    }, 1600);
  };

  return (
    <section className="py-28 sm:py-36">
      <Container size="wide">
        <SectionHeading title={LANDING_LABELS.liveDemo.title} description={LANDING_LABELS.liveDemo.body} />

        <Reveal delay={0.1}>
          <div className="mx-auto mt-14 max-w-4xl overflow-hidden rounded-3xl border border-border bg-bg-elevated shadow-xl">
            <div className="flex items-center justify-between border-b border-border px-6 py-4">
              <div className="flex items-center gap-3">
                <span
                  className={cn(
                    "size-2.5 rounded-full transition-colors duration-500",
                    playing ? "bg-accent" : "bg-neutral-300 dark:bg-neutral-700"
                  )}
                />
                <span className="text-sm font-medium tracking-tight text-fg">
                  {LANDING_LABELS.liveDemo.cta}
                </span>
                <span className="text-xs text-fg-subtle">· {LANDING_LABELS.liveDemo.duration}</span>
              </div>
              <Button
                size="sm"
                variant={playing ? "secondary" : "primary"}
                onClick={toggle}
                leading={playing ? <Pause className="size-3.5" /> : <Play className="size-3.5" />}
              >
                {playing ? "Pause" : "Play sample"}
              </Button>
            </div>

            <div className="space-y-3 p-6 sm:p-8">
              {DEMO_TRANSCRIPT.map((turn, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: i <= step ? 1 : 0.25, y: 0 }}
                  transition={{ duration: 0.5, ease }}
                  className={cn(
                    "flex items-start gap-3 rounded-2xl border border-border p-4",
                    turn.speaker === "Live API" ? "bg-bg-muted/50" : "bg-bg-elevated"
                  )}
                >
                  <span
                    className={cn(
                      "mt-0.5 shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider",
                      turn.speaker === "Live API"
                        ? "bg-accent text-neutral-0"
                        : "bg-bg-muted text-fg-muted"
                    )}
                  >
                    {turn.speaker}
                  </span>
                  <div className="flex-1">
                    <p className="text-[14.5px] leading-relaxed text-fg">{turn.text}</p>
                    <span className="mt-1.5 inline-block text-[11px] text-fg-subtle">{turn.at}</span>
                  </div>
                </motion.div>
              ))}
            </div>

            <AnimatePresence>
              {!playing && step === 0 && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="flex items-center gap-2 border-t border-border bg-bg-muted/40 px-6 py-3 text-xs text-fg-muted"
                >
                  <Sparkles className="size-3.5 text-accent" />
                  The AI knew the customer's billing history before they asked. No screen, no script.
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </Reveal>
      </Container>
    </section>
  );
}
