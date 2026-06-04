import { CircleCheck, FileText, Hand, Sparkles, Wrench } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { Reveal } from "@/components/ui/reveal";
import { BUSINESS_LABELS } from "@/labels";
import {
  getRecentActivity,
  type ActivityEvent,
} from "@/mocks/business/console";
import { cn } from "@/lib/cn";

const KIND_ICON: Record<ActivityEvent["kind"], LucideIcon> = {
  tool: Wrench,
  escalation: Hand,
  document: FileText,
  resolved: CircleCheck,
  gap: Sparkles,
};

const KIND_TONE: Record<ActivityEvent["kind"], string> = {
  tool: "text-fg-muted bg-bg-muted",
  escalation: "text-watch bg-watch-soft",
  document: "text-accent-strong bg-accent-soft",
  resolved: "text-success bg-success-soft",
  gap: "text-warning bg-warning-soft",
};

function formatAgo(min: number) {
  if (min < 1) return "just now";
  if (min < 60) return `${min} min ago`;
  const h = Math.floor(min / 60);
  return `${h} h ago`;
}

export function RecentActivity() {
  const events = getRecentActivity();

  return (
    <Reveal>
      <Card className="h-full">
        <CardHeader
          title={BUSINESS_LABELS.activity.title}
          caption={BUSINESS_LABELS.activity.caption}
        />
        <CardBody className="p-0">
          <ul className="divide-y divide-border/60">
            {events.map((e) => {
              const Icon = KIND_ICON[e.kind];
              return (
                <li key={e.id} className="flex items-start gap-3 px-5 py-3.5">
                  <span
                    className={cn(
                      "mt-0.5 inline-flex size-7 shrink-0 items-center justify-center rounded-md",
                      KIND_TONE[e.kind]
                    )}
                  >
                    <Icon className="size-3.5" strokeWidth={1.75} />
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm leading-snug text-fg">{e.message}</p>
                    <p className="mt-0.5 text-xs text-fg-subtle">{formatAgo(e.atMinAgo)}</p>
                  </div>
                </li>
              );
            })}
          </ul>
        </CardBody>
      </Card>
    </Reveal>
  );
}
