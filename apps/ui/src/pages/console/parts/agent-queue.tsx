import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { StatusPill } from "@/components/ui/status-pill";
import { Reveal } from "@/components/ui/reveal";
import { BUSINESS_LABELS } from "@/labels";
import { getAgentQueue, type Escalation } from "@/mocks/business/console";

const PRIORITY_VARIANT: Record<Escalation["priority"], "critical" | "monitor" | "neutral"> = {
  high: "critical",
  medium: "monitor",
  low: "neutral",
};

const PRIORITY_LABEL: Record<Escalation["priority"], string> = {
  high: BUSINESS_LABELS.agents.priorities.high,
  medium: BUSINESS_LABELS.agents.priorities.medium,
  low: BUSINESS_LABELS.agents.priorities.low,
};

function formatWaiting(sec: number) {
  if (sec < 60) return `${sec}s`;
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function AgentQueue() {
  const queue = getAgentQueue();

  return (
    <Reveal>
      <Card>
        <CardHeader
          title={BUSINESS_LABELS.agents.title}
          caption={BUSINESS_LABELS.agents.caption}
        />
        <CardBody className="p-0">
          <div className="grid grid-cols-[1fr_auto] gap-x-3 gap-y-1 px-5 py-3 text-[11px] font-semibold uppercase tracking-wider text-fg-subtle sm:grid-cols-[1fr_1.4fr_1.4fr_auto_auto_auto]">
            <span>{BUSINESS_LABELS.agents.columns.id}</span>
            <span className="hidden sm:block">{BUSINESS_LABELS.agents.columns.customer}</span>
            <span className="hidden sm:block">{BUSINESS_LABELS.agents.columns.topic}</span>
            <span className="hidden sm:block">{BUSINESS_LABELS.agents.columns.waiting}</span>
            <span className="hidden sm:block">{BUSINESS_LABELS.agents.columns.priority}</span>
            <span />
          </div>
          <ul className="divide-y divide-border/60 border-t border-border/60">
            {queue.map((e) => (
              <li
                key={e.id}
                className="grid grid-cols-[1fr_auto] items-center gap-x-3 gap-y-1 px-5 py-3.5 text-sm transition-colors duration-300 ease-[var(--ease-out-soft)] hover:bg-bg-muted/40 sm:grid-cols-[1fr_1.4fr_1.4fr_auto_auto_auto]"
              >
                <div className="min-w-0">
                  <span className="font-medium text-fg">{e.id}</span>
                  <p className="mt-0.5 line-clamp-1 text-xs text-fg-muted sm:hidden">
                    {e.topic} · {formatWaiting(e.waitingSec)} · {e.reason}
                  </p>
                </div>
                <span className="hidden truncate text-fg-muted sm:block">{e.customer}</span>
                <span className="hidden truncate text-fg-muted sm:block">{e.topic}</span>
                <span className="hidden text-xs tabular-nums text-fg-muted sm:block">
                  {formatWaiting(e.waitingSec)}
                </span>
                <span className="hidden sm:inline-flex">
                  <StatusPill variant={PRIORITY_VARIANT[e.priority]} label={PRIORITY_LABEL[e.priority]} />
                </span>
                <button
                  type="button"
                  className="inline-flex items-center justify-end gap-1.5 rounded-full bg-fg px-3 py-1 text-[11px] font-medium tracking-tight text-bg shadow-sm transition-all duration-300 ease-[var(--ease-out-soft)] hover:shadow-md active:scale-[0.97]"
                >
                  {BUSINESS_LABELS.agents.claim}
                </button>
              </li>
            ))}
          </ul>
        </CardBody>
      </Card>
    </Reveal>
  );
}
