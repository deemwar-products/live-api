import { Mic, MessageSquare, Eye, Hand } from "lucide-react";
import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { StatusPill } from "@/components/ui/status-pill";
import { Reveal } from "@/components/ui/reveal";
import { BUSINESS_LABELS } from "@/labels";
import { getLiveConversations, type HealthBand } from "@/mocks/business/console";
import { cn } from "@/lib/cn";

const HEALTH_LABEL: Record<HealthBand, string> = {
  normal: BUSINESS_LABELS.live.health.normal,
  monitor: BUSINESS_LABELS.live.health.monitor,
  atRisk: BUSINESS_LABELS.live.health.atRisk,
  critical: BUSINESS_LABELS.live.health.critical,
};

function formatDuration(sec: number) {
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function LiveConversations() {
  const convos = getLiveConversations();

  return (
    <Reveal>
      <Card>
        <CardHeader
          title={BUSINESS_LABELS.live.title}
          caption={BUSINESS_LABELS.live.caption}
          action={
            <button
              type="button"
              className="rounded-full px-3 py-1 text-xs font-medium text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:bg-bg-muted hover:text-fg"
            >
              {BUSINESS_LABELS.live.viewAll} →
            </button>
          }
        />
        <CardBody className="p-0">
          <div className="grid grid-cols-[1fr_auto] gap-x-3 gap-y-1 px-5 py-3 text-[11px] font-semibold uppercase tracking-wider text-fg-subtle sm:grid-cols-[1fr_1.4fr_auto_auto_auto_auto]">
            <span>{BUSINESS_LABELS.live.columns.id}</span>
            <span className="hidden sm:block">{BUSINESS_LABELS.live.columns.topic}</span>
            <span className="hidden sm:block">{BUSINESS_LABELS.live.columns.channel}</span>
            <span className="hidden sm:block">{BUSINESS_LABELS.live.columns.duration}</span>
            <span className="hidden sm:block">{BUSINESS_LABELS.live.columns.health}</span>
            <span className="text-right">{BUSINESS_LABELS.live.columns.action}</span>
          </div>
          <ul className="divide-y divide-border/60 border-t border-border/60">
            {convos.map((c) => (
              <li
                key={c.id}
                className="grid grid-cols-[1fr_auto] items-center gap-x-3 gap-y-1 px-5 py-3.5 text-sm transition-colors duration-300 ease-[var(--ease-out-soft)] hover:bg-bg-muted/40 sm:grid-cols-[1fr_1.4fr_auto_auto_auto_auto]"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-fg">{c.id}</span>
                    <span className="truncate text-xs text-fg-subtle">· {c.customer}</span>
                  </div>
                  <p className="mt-0.5 truncate text-xs text-fg-muted sm:hidden">
                    {c.topic} · {formatDuration(c.durationSec)}
                  </p>
                </div>
                <span className="hidden truncate text-fg-muted sm:block">{c.topic}</span>
                <span className="hidden items-center gap-1.5 text-xs text-fg-muted sm:inline-flex">
                  {c.channel === "voice" ? (
                    <Mic className="size-3.5" strokeWidth={1.75} />
                  ) : (
                    <MessageSquare className="size-3.5" strokeWidth={1.75} />
                  )}
                  {c.channel === "voice" ? BUSINESS_LABELS.live.channels.voice : BUSINESS_LABELS.live.channels.chat}
                </span>
                <span className="hidden text-xs tabular-nums text-fg-muted sm:block">
                  {formatDuration(c.durationSec)}
                </span>
                <span className="hidden sm:inline-flex">
                  <StatusPill variant={c.healthBand} label={HEALTH_LABEL[c.healthBand]} />
                </span>
                <div className="flex items-center justify-end gap-1.5">
                  <ActionButton
                    icon={Eye}
                    label={BUSINESS_LABELS.live.actions.watch}
                    tone="ghost"
                    emphasis={c.healthBand === "critical" || c.healthBand === "atRisk"}
                  />
                  <ActionButton
                    icon={Hand}
                    label={BUSINESS_LABELS.live.actions.takeover}
                    tone="solid"
                    emphasis={c.healthBand === "critical" || c.healthBand === "atRisk"}
                  />
                </div>
              </li>
            ))}
          </ul>
        </CardBody>
      </Card>
    </Reveal>
  );
}

function ActionButton({
  icon: Icon,
  label,
  tone,
  emphasis,
}: {
  icon: typeof Eye;
  label: string;
  tone: "ghost" | "solid";
  emphasis: boolean;
}) {
  return (
    <button
      type="button"
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-medium tracking-tight transition-all duration-300 ease-[var(--ease-out-soft)] active:scale-[0.97]",
        tone === "ghost" && "text-fg-muted hover:bg-bg-muted hover:text-fg",
        tone === "solid" &&
          (emphasis
            ? "bg-fg text-bg shadow-sm hover:shadow-md"
            : "border border-border bg-bg-elevated text-fg hover:border-border-strong hover:bg-bg-muted")
      )}
    >
      <Icon className="size-3" strokeWidth={1.75} />
      <span className="hidden md:inline">{label}</span>
    </button>
  );
}
