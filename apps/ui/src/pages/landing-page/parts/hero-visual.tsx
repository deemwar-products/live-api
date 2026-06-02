/**
 * Hero visual: a calm, animated voice waveform inside a "browser window"
 * card. The waveform is the product story — you're watching a conversation
 * happen. Subtle mouse-tracked tilt makes it feel alive without being busy.
 */
import { motion, useMotionValue, useSpring, useTransform } from "framer-motion";
import { useRef } from "react";
import { cn } from "@/lib/cn";

const BARS = 64;

export function HeroVisual({ className }: { className?: string }) {
  const cardRef = useRef<HTMLDivElement>(null);
  const mx = useMotionValue(0);
  const my = useMotionValue(0);
  const rotateX = useSpring(useTransform(my, [-0.5, 0.5], [4, -4]), { stiffness: 120, damping: 20 });
  const rotateY = useSpring(useTransform(mx, [-0.5, 0.5], [-4, 4]), { stiffness: 120, damping: 20 });

  // Deterministic pseudo-random heights so SSR/CSR match.
  const heights = Array.from({ length: BARS }, (_, i) => {
    const t = (i / BARS) * Math.PI * 4;
    return 0.35 + 0.5 * Math.abs(Math.sin(t * 0.9) * Math.cos(t * 0.4 + 0.7)) + 0.15 * Math.sin(t * 2.1);
  });

  const onMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const r = cardRef.current?.getBoundingClientRect();
    if (!r) return;
    mx.set((e.clientX - r.left) / r.width - 0.5);
    my.set((e.clientY - r.top) / r.height - 0.5);
  };
  const onMouseLeave = () => {
    mx.set(0);
    my.set(0);
  };

  return (
    <div className={cn("relative w-full [perspective:1200px]", className)}>
      <div className="absolute inset-0 -z-10 rounded-[2rem] bg-ambient opacity-90 blur-3xl" />
      <motion.div
        ref={cardRef}
        onMouseMove={onMouseMove}
        onMouseLeave={onMouseLeave}
        style={{ rotateX, rotateY, transformStyle: "preserve-3d" }}
        className={cn(
          "relative overflow-hidden rounded-3xl border border-border bg-bg-elevated p-6 shadow-xl",
          "transition-colors duration-500 ease-[var(--ease-out-soft)]"
        )}
      >
        {/* Window chrome */}
        <div className="mb-5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <span className="size-2.5 rounded-full bg-neutral-300 dark:bg-neutral-700" />
            <span className="size-2.5 rounded-full bg-neutral-300 dark:bg-neutral-700" />
            <span className="size-2.5 rounded-full bg-neutral-300 dark:bg-neutral-700" />
          </div>
          <div className="flex items-center gap-2 rounded-full bg-bg-muted px-3 py-1 text-[11px] font-medium tracking-wide text-fg-subtle">
            <span className="relative flex size-1.5">
              <span className="absolute inset-0 animate-ping rounded-full bg-accent opacity-75" />
              <span className="relative size-1.5 rounded-full bg-accent" />
            </span>
            live conversation · 01:14
          </div>
        </div>

        {/* Waveform */}
        <div className="flex h-44 items-center justify-between gap-[3px]">
          {heights.map((h, i) => (
            <motion.span
              key={i}
              className="block w-[3px] rounded-full bg-accent"
              initial={{ scaleY: 0.2, opacity: 0.4 }}
              animate={{ scaleY: [0.2, h, 0.2], opacity: [0.45, 1, 0.45] }}
              transition={{
                duration: 2.4,
                repeat: Infinity,
                ease: "easeInOut",
                delay: (i % 12) * 0.06,
              }}
              style={{ height: `${h * 100}%`, transformOrigin: "center" }}
            />
          ))}
        </div>

        {/* Captions */}
        <div className="mt-6 space-y-2.5">
          <Caption speaker="Customer" text="Hi — I was charged twice this month, can you take a look?" />
          <Caption
            speaker="Live API"
            text="I can see two charges. The second one looks like a duplicate from a failed retry — starting a refund now."
            accent
          />
        </div>
      </motion.div>
    </div>
  );
}

function Caption({ speaker, text, accent }: { speaker: string; text: string; accent?: boolean }) {
  return (
    <div className="flex items-start gap-3">
      <span
        className={cn(
          "mt-0.5 shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider",
          accent ? "bg-accent text-neutral-0" : "bg-bg-muted text-fg-muted"
        )}
      >
        {speaker}
      </span>
      <p className="text-[13px] leading-relaxed text-fg-muted">{text}</p>
    </div>
  );
}
