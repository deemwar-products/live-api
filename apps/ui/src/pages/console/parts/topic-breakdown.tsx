import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { Reveal } from "@/components/ui/reveal";
import { BUSINESS_LABELS } from "@/labels";
import { getTopicBreakdown } from "@/mocks/business/console";

export function TopicBreakdown() {
  const topics = getTopicBreakdown();
  const total = topics.reduce((s, t) => s + t.count, 0);

  return (
    <Reveal>
      <Card className="h-full">
        <CardHeader
          title={BUSINESS_LABELS.topics.title}
          caption={BUSINESS_LABELS.topics.caption}
        />
        <CardBody>
          <ul className="space-y-4">
            {topics.map((t) => {
              const pct = Math.round(t.share * 100);
              return (
                <li key={t.category}>
                  <div className="mb-1.5 flex items-baseline justify-between gap-3 text-sm">
                    <span className="font-medium text-fg">{t.category}</span>
                    <span className="tabular-nums text-xs text-fg-muted">
                      {t.count.toLocaleString()} · {pct}%
                    </span>
                  </div>
                  <div className="h-2 w-full overflow-hidden rounded-full bg-bg-muted">
                    <div
                      className="h-full rounded-full bg-accent"
                      style={{ width: `${pct}%`, opacity: 0.45 + t.share * 1.1 }}
                    />
                  </div>
                </li>
              );
            })}
          </ul>
          <p className="mt-5 border-t border-border/60 pt-4 text-xs text-fg-subtle">
            <span className="font-semibold text-fg">{total.toLocaleString()}</span>{" "}
            {BUSINESS_LABELS.topics.totalLabel}
          </p>
        </CardBody>
      </Card>
    </Reveal>
  );
}
