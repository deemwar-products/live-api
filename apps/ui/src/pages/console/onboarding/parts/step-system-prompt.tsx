import { Plus, Sparkles } from "lucide-react";
import { useState } from "react";
import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Select } from "@/components/ui/select";
import { Tag } from "@/components/ui/tag";
import { ONBOARDING_LABELS } from "@/labels";
import type { OnboardingState } from "@/mocks/business/onboarding";
import { cn } from "@/lib/cn";

const TONE_OPTIONS = [
  { value: "calm" as const, label: ONBOARDING_LABELS.persona.tone.options.calm },
  { value: "friendly" as const, label: ONBOARDING_LABELS.persona.tone.options.friendly },
  { value: "formal" as const, label: ONBOARDING_LABELS.persona.tone.options.formal },
  { value: "technical" as const, label: ONBOARDING_LABELS.persona.tone.options.technical },
];

const MODE_OPTIONS = [
  { value: "auto" as const, label: ONBOARDING_LABELS.persona.mode.auto },
  { value: "autoEdit" as const, label: ONBOARDING_LABELS.persona.mode.autoEdit },
  { value: "custom" as const, label: ONBOARDING_LABELS.persona.mode.custom },
];

type Props = {
  state: OnboardingState;
  onChange: (patch: Partial<OnboardingState>) => void;
};

export function StepSystemPrompt({ state, onChange }: Props) {
  const [langInput, setLangInput] = useState("");

  const addLanguage = () => {
    const v = langInput.trim().toLowerCase();
    if (!v || state.languages.includes(v)) {
      setLangInput("");
      return;
    }
    onChange({ languages: [...state.languages, v] });
    setLangInput("");
  };

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_360px]">
      <Card>
        <CardHeader
          title={ONBOARDING_LABELS.persona.title}
          caption={ONBOARDING_LABELS.persona.description}
        />
        <CardBody className="space-y-6">
          <SegmentedControl
            label={ONBOARDING_LABELS.persona.mode.label}
            value={state.mode}
            options={MODE_OPTIONS}
            onChange={(v) => onChange({ mode: v })}
            hint={ONBOARDING_LABELS.persona.mode.hint}
          />

          <Field label={ONBOARDING_LABELS.persona.tone.label}>
            <Select
              value={state.tone}
              onChange={(v) => onChange({ tone: v })}
              options={TONE_OPTIONS}
              ariaLabel={ONBOARDING_LABELS.persona.tone.label}
            />
          </Field>

          <Field label={ONBOARDING_LABELS.persona.languages.label} hint={ONBOARDING_LABELS.persona.languages.hint}>
            <div className="flex flex-wrap items-center gap-2">
              {state.languages.map((l) => (
                <Tag key={l} onRemove={() => onChange({ languages: state.languages.filter((x) => x !== l) })}>
                  {l}
                </Tag>
              ))}
              <div className="flex items-center gap-1.5">
                <Input
                  value={langInput}
                  onChange={(e) => setLangInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      addLanguage();
                    }
                  }}
                  placeholder={ONBOARDING_LABELS.persona.languages.addPlaceholder}
                  className="h-8 w-44 px-3 text-xs"
                />
                <button
                  type="button"
                  onClick={addLanguage}
                  aria-label="Add language"
                  className="inline-flex size-8 items-center justify-center rounded-full border border-border bg-bg-elevated text-fg-muted transition-colors duration-300 ease-[var(--ease-out-soft)] hover:border-accent hover:text-fg"
                >
                  <Plus className="size-3.5" />
                </button>
              </div>
            </div>
          </Field>

          <Field label={ONBOARDING_LABELS.persona.systemPrompt.label} hint={ONBOARDING_LABELS.persona.systemPrompt.hint}>
            <Textarea
              value={state.systemPrompt}
              onChange={(e) => onChange({ systemPrompt: e.target.value })}
              placeholder="You are a calm, accurate support specialist for Acme Corp. Always cite the source document, never guess pricing, and offer to escalate at the first sign of frustration."
              rows={6}
            />
          </Field>

          <Field
            label={
              <span className="inline-flex items-center gap-2">
                {ONBOARDING_LABELS.persona.greeting.label}
                <label className="ml-auto inline-flex cursor-pointer items-center gap-2 text-xs text-fg-muted">
                  <input
                    type="checkbox"
                    checked={state.greetingEnabled}
                    onChange={(e) => onChange({ greetingEnabled: e.target.checked })}
                    className="sr-only"
                  />
                  <span
                    className={cn(
                      "relative inline-flex h-5 w-9 items-center rounded-full transition-colors duration-300 ease-[var(--ease-out-soft)]",
                      state.greetingEnabled ? "bg-accent" : "bg-bg-muted"
                    )}
                  >
                    <span
                      className={cn(
                        "inline-block size-4 transform rounded-full bg-bg-elevated shadow-sm transition-transform duration-300 ease-[var(--ease-out-soft)]",
                        state.greetingEnabled ? "translate-x-4" : "translate-x-0.5"
                      )}
                    />
                  </span>
                  {ONBOARDING_LABELS.persona.greeting.enabled}
                </label>
              </span>
            }
            hint={ONBOARDING_LABELS.persona.greeting.hint}
          >
            <Textarea
              value={state.greeting}
              disabled={!state.greetingEnabled}
              onChange={(e) => onChange({ greeting: e.target.value })}
              rows={2}
            />
          </Field>
        </CardBody>
      </Card>

      <PreviewCard />
    </div>
  );
}

function Field({
  label,
  hint,
  children,
}: {
  label: React.ReactNode;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <div className="mb-1.5 flex flex-wrap items-center justify-between gap-2">
        <label className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">
          {label}
        </label>
      </div>
      {children}
      {hint && <p className="mt-1.5 text-xs text-fg-subtle">{hint}</p>}
    </div>
  );
}

function SegmentedControl<V extends string>({
  label,
  value,
  options,
  onChange,
  hint,
}: {
  label: string;
  value: V;
  options: { value: V; label: React.ReactNode }[];
  onChange: (v: V) => void;
  hint?: string;
}) {
  return (
    <div>
      <div className="mb-1.5 flex flex-wrap items-center justify-between gap-2">
        <label className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">
          {label}
        </label>
      </div>
      <div className="inline-flex rounded-full border border-border bg-bg-elevated p-1">
        {options.map((o) => {
          const selected = o.value === value;
          return (
            <button
              key={o.value}
              type="button"
              onClick={() => onChange(o.value)}
              className={cn(
                "rounded-full px-3.5 py-1.5 text-xs font-medium tracking-tight transition-all duration-300 ease-[var(--ease-out-soft)]",
                selected ? "bg-fg text-bg shadow-sm" : "text-fg-muted hover:text-fg"
              )}
            >
              {o.label}
            </button>
          );
        })}
      </div>
      {hint && <p className="mt-1.5 text-xs text-fg-subtle">{hint}</p>}
    </div>
  );
}

function PreviewCard() {
  return (
    <Card className="h-fit">
      <CardHeader
        title={
          <span className="inline-flex items-center gap-1.5">
            <Sparkles className="size-3.5 text-accent" />
            {ONBOARDING_LABELS.persona.preview.title}
          </span>
        }
      />
      <CardBody>
        <div className="space-y-3">
          <Bubble role="customer">{ONBOARDING_LABELS.persona.preview.customerSays}</Bubble>
          <Bubble role="ai">{ONBOARDING_LABELS.persona.preview.aiResponds}</Bubble>
        </div>
      </CardBody>
    </Card>
  );
}

function Bubble({ role, children }: { role: "customer" | "ai"; children: React.ReactNode }) {
  const isAi = role === "ai";
  return (
    <div className={cn("flex", isAi ? "justify-start" : "justify-end")}>
      <div
        className={cn(
          "max-w-[85%] rounded-2xl px-3.5 py-2.5 text-sm leading-relaxed shadow-sm",
          isAi ? "bg-bg-muted text-fg" : "bg-fg text-bg"
        )}
      >
        {children}
      </div>
    </div>
  );
}
