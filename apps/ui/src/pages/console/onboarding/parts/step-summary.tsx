import { ArrowRight, CircleCheck, Sparkles } from "lucide-react";
import { motion } from "framer-motion";
import { Card, CardBody } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ONBOARDING_LABELS } from "@/labels";

const ease = [0.22, 1, 0.36, 1] as const;

const CHECKLIST = [
  { id: 1, label: ONBOARDING_LABELS.summary.checklist.knowledge },
  { id: 2, label: ONBOARDING_LABELS.summary.checklist.persona },
  { id: 3, label: ONBOARDING_LABELS.summary.checklist.tools },
  { id: 4, label: ONBOARDING_LABELS.summary.checklist.team },
];

type Props = {
  onGoToDashboard: () => void;
  onCustomize: () => void;
};

export function StepSummary({ onGoToDashboard, onCustomize }: Props) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, ease }}
    >
      <Card>
        <CardBody className="flex flex-col items-center gap-6 py-12 text-center">
          <span className="inline-flex size-14 items-center justify-center rounded-full bg-success-soft text-success">
            <Sparkles className="size-6" strokeWidth={1.5} />
          </span>
          <div>
            <h2 className="text-2xl font-semibold tracking-tight text-fg sm:text-3xl">
              {ONBOARDING_LABELS.summary.title}
            </h2>
            <p className="mt-2 max-w-md text-pretty text-sm text-fg-muted sm:text-base">
              {ONBOARDING_LABELS.summary.subtitle}
            </p>
          </div>

          <ul className="w-full max-w-sm space-y-2.5 text-left">
            {CHECKLIST.map((c) => (
              <li
                key={c.id}
                className="flex items-center gap-3 rounded-xl border border-border bg-bg-muted/40 px-4 py-2.5 text-sm"
              >
                <CircleCheck className="size-4 shrink-0 text-success" strokeWidth={2} />
                <span className="font-medium text-fg">{c.label}</span>
              </li>
            ))}
          </ul>

          <div className="flex flex-wrap items-center justify-center gap-3">
            <Button
              size="lg"
              onClick={onGoToDashboard}
              trailing={<ArrowRight className="size-4" />}
            >
              {ONBOARDING_LABELS.actions.goToDashboard}
            </Button>
            <Button size="lg" variant="secondary" onClick={onCustomize}>
              {ONBOARDING_LABELS.actions.customizeFurther}
            </Button>
          </div>
        </CardBody>
      </Card>
    </motion.div>
  );
}
