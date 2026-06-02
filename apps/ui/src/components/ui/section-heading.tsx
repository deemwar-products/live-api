import type { ReactNode } from "react";
import { motion } from "framer-motion";
import { cn } from "@/lib/cn";

const ease = [0.22, 1, 0.36, 1] as const;

type Props = {
  eyebrow?: string;
  title: ReactNode;
  description?: ReactNode;
  align?: "left" | "center";
  className?: string;
  size?: "md" | "lg";
};

export function SectionHeading({ eyebrow, title, description, align = "center", className, size = "lg" }: Props) {
  const titleClass =
    size === "lg"
      ? "text-3xl font-semibold tracking-tight text-fg sm:text-4xl md:text-[48px] md:leading-[1.05]"
      : "text-2xl font-semibold tracking-tight text-fg sm:text-3xl";

  return (
    <div
      className={cn(
        "flex flex-col gap-5",
        align === "center" ? "items-center text-center" : "items-start text-left",
        className
      )}
    >
      {eyebrow && (
        <motion.span
          initial={{ opacity: 0, y: 8 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.6, ease }}
          className="inline-flex items-center gap-2 rounded-full border border-border bg-bg-elevated px-3 py-1 text-xs font-medium text-fg-muted"
        >
          <span className="size-1.5 rounded-full bg-accent" />
          {eyebrow}
        </motion.span>
      )}
      <motion.h2
        initial={{ opacity: 0, y: 16 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, margin: "-80px" }}
        transition={{ duration: 0.7, ease, delay: eyebrow ? 0.05 : 0 }}
        className={cn("text-balance", titleClass)}
      >
        {title}
      </motion.h2>
      {description && (
        <motion.p
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.7, ease, delay: eyebrow ? 0.1 : 0.05 }}
          className="max-w-2xl text-pretty text-base text-fg-muted sm:text-lg"
        >
          {description}
        </motion.p>
      )}
    </div>
  );
}
