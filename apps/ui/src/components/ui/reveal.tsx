import type { ReactNode } from "react";
import { motion, type Transition, type Variants } from "framer-motion";

type Props = {
  children: ReactNode;
  delay?: number;
  className?: string;
  y?: number;
  as?: "div" | "section" | "article" | "li";
};

const transition: Transition = {
  duration: 0.8,
  ease: [0.22, 1, 0.36, 1],
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  visible: { opacity: 1, y: 0 },
};

export function Reveal({ children, delay = 0, y = 16, className, as = "div" }: Props) {
  const MotionTag = motion[as] as typeof motion.div;
  return (
    <MotionTag
      initial={{ opacity: 0, y }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-60px" }}
      transition={{ ...transition, delay }}
      className={className}
    >
      {children}
    </MotionTag>
  );
}

/**
 * RevealChildren — wraps a list-like container and staggers its direct
 * motion children. Children opt in by setting `data-reveal` on themselves.
 * Cleaner than passing index into every Reveal.
 */
export function RevealStagger({
  children,
  className,
  step = 0.08,
  initialDelay = 0,
}: {
  children: ReactNode;
  className?: string;
  step?: number;
  initialDelay?: number;
}) {
  return (
    <motion.div
      className={className}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, margin: "-60px" }}
      variants={{
        hidden: {},
        visible: {
          transition: {
            staggerChildren: step,
            delayChildren: initialDelay,
          },
        },
      }}
    >
      {children}
    </motion.div>
  );
}

export const revealItem = itemVariants;
