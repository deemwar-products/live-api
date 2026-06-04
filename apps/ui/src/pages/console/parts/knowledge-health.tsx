import { ArrowUpRight, Sparkles } from "lucide-react";
import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { Reveal } from "@/components/ui/reveal";
import { BUSINESS_LABELS } from "@/labels";
import { getKnowledgeGaps } from "@/mocks/business/console";

export function KnowledgeHealth() {
  const gaps = getKnowledgeGaps();
  const totalUnanswered = gaps.reduce((s, g) => s + g.askedTimes, 0);

  return (
    <Reveal>
      <Card className="h-full">
        <CardHeader
          title={BUSINESS_LABELS.knowledge.title}
          caption={BUSINESS_LABELS.knowledge.caption}
          action={
            <button
              type="button"
              className="inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-medium text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:bg-bg-muted hover:text-fg"
            >
              {BUSINESS_LABELS.knowledge.viewDocs}
              <ArrowUpRight className="size-3" />
            </button>
          }
        />
        <CardBody>
          <div className="mb-5 flex items-baseline gap-2">
            <span className="text-3xl font-semibold tracking-tight text-fg">
              {totalUnanswered}
            </span>
            <span className="text-sm text-fg-muted">
              {BUSINESS_LABELS.knowledge.unansweredToday}
            </span>
          </div>
          <p className="mb-3 text-[11px] font-semibold uppercase tracking-wider text-fg-subtle">
            {BUSINESS_LABELS.knowledge.samplesTitle}
          </p>
          <ul className="space-y-3">
            {gaps.map((g) => (
              <li key={g.id} className="flex items-start gap-3">
                <span className="mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-md bg-warning-soft text-warning">
                  <Sparkles className="size-3" strokeWidth={1.75} />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="line-clamp-2 text-sm leading-snug text-fg">
                    {g.question}
                  </p>
                  <p className="mt-0.5 text-xs text-fg-subtle">
                    Asked {g.askedTimes}× · {g.lastAskedAt}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        </CardBody>
      </Card>
    </Reveal>
  );
}
