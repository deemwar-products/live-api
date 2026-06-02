import { motion, useScroll, useTransform, type MotionValue } from "framer-motion";
import { useRef } from "react";
import { cn } from "@/lib/cn";

type Props = {
  className?: string;
  /** Multiplier on the parallax strength. Negative moves opposite to scroll. */
  strength?: number;
  /** Rendered inside the parallax layer. */
  children: React.ReactNode;
};

/**
 * Wraps any visual content in a scroll-linked parallax transform.
 * Use sparingly — one per section, max. Children can be any node, but
 * illustrations with intrinsic aspect ratios work best.
 */
export function ParallaxIllustration({ className, strength = -40, children }: Props) {
  const ref = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start end", "end start"],
  });
  const y = useTransform(scrollYProgress, [0, 1], [strength, -strength]);

  return (
    <div ref={ref} className={cn("pointer-events-none absolute inset-0 overflow-hidden", className)}>
      <ParallaxLayer y={y}>{children}</ParallaxLayer>
    </div>
  );
}

function ParallaxLayer({ y, children }: { y: MotionValue<number>; children: React.ReactNode }) {
  return (
    <motion.div style={{ y }} className="h-full w-full">
      {children}
    </motion.div>
  );
}
