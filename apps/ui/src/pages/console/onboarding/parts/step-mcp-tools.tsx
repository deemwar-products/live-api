import { Bell, CircleAlert, FileSearch, LifeBuoy, Wrench } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { ONBOARDING_LABELS } from "@/labels";
import type { OnboardingState, ToolKey } from "@/mocks/business/onboarding";

const TOOL_ICONS: Record<ToolKey, LucideIcon> = {
  rag: FileSearch,
  notify: Bell,
  ticket: LifeBuoy,
  gap: CircleAlert,
  credit: Wrench,
};

const TOOL_KEYS: ToolKey[] = ["rag", "notify", "ticket", "gap", "credit"];

type Props = {
  state: OnboardingState;
  onChange: (patch: Partial<OnboardingState>) => void;
};

export function StepMcpTools({ state, onChange }: Props) {
  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_320px]">
      <Card>
        <CardHeader
          title={ONBOARDING_LABELS.tools.title}
          caption={ONBOARDING_LABELS.tools.description}
        />
        <CardBody className="space-y-2.5">
          {TOOL_KEYS.map((key) => {
            const meta = ONBOARDING_LABELS.tools.items[key];
            const Icon = TOOL_ICONS[key];
            return (
              <Checkbox
                key={key}
                checked={state.tools[key]}
                onChange={(v) => onChange({ tools: { ...state.tools, [key]: v } })}
                label={
                  <span className="inline-flex items-center gap-2">
                    <Icon className="size-3.5 text-fg-muted" strokeWidth={1.75} />
                    {meta.name}
                  </span>
                }
                description={meta.description}
              />
            );
          })}
        </CardBody>
      </Card>

      <Card className="h-fit">
        <CardHeader title={ONBOARDING_LABELS.tools.whatAre.title} />
        <CardBody>
          <p className="text-sm leading-relaxed text-fg-muted">
            {ONBOARDING_LABELS.tools.whatAre.body}
          </p>
        </CardBody>
      </Card>
    </div>
  );
}
