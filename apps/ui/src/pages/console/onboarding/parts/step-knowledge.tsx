import { useRef } from "react";
import {
  CheckCircle2,
  FileText,
  Globe,
  Link2,
  ListChecks,
  Sparkles,
  Upload,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { Tag } from "@/components/ui/tag";
import { ONBOARDING_LABELS } from "@/labels";
import type { DocumentRow, OnboardingState, Source } from "@/mocks/business/onboarding";
import { cn } from "@/lib/cn";

const SOURCES: { key: Source; label: string; icon: LucideIcon }[] = [
  { key: "notion", label: ONBOARDING_LABELS.knowledge.sources.notion, icon: FileText },
  { key: "confluence", label: ONBOARDING_LABELS.knowledge.sources.confluence, icon: FileText },
  { key: "zendesk", label: ONBOARDING_LABELS.knowledge.sources.zendesk, icon: ListChecks },
  { key: "url", label: ONBOARDING_LABELS.knowledge.sources.url, icon: Link2 },
];

type Props = {
  state: OnboardingState;
  onAddDocument: (doc: DocumentRow) => void;
};

export function StepKnowledge({ state, onAddDocument }: Props) {
  const fileInput = useRef<HTMLInputElement>(null);

  const handleAddMock = () => {
    const next: DocumentRow = {
      id: `d${Date.now()}`,
      name: `Knowledge-doc-${state.documents.length + 1}.pdf`,
      sizeKb: 220 + Math.floor(Math.random() * 800),
      status: "indexed",
      lang: "en",
      chunks: 18 + Math.floor(Math.random() * 60),
    };
    onAddDocument(next);
  };

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_320px]">
      <Card>
        <CardHeader
          title={ONBOARDING_LABELS.knowledge.title}
          caption={ONBOARDING_LABELS.knowledge.description}
        />
        <CardBody className="space-y-6">
          <button
            type="button"
            onClick={() => fileInput.current?.click()}
            onDragOver={(e) => e.preventDefault()}
            onDrop={(e) => {
              e.preventDefault();
              handleAddMock();
            }}
            className={cn(
              "group flex w-full flex-col items-center justify-center gap-3 rounded-2xl border-2 border-dashed border-border bg-bg-muted/30 px-6 py-12 text-center",
              "transition-colors duration-300 ease-[var(--ease-out-soft)] hover:border-accent/50 hover:bg-accent-soft/20"
            )}
          >
            <span className="inline-flex size-12 items-center justify-center rounded-full bg-accent-soft text-accent-strong transition-transform duration-300 ease-[var(--ease-out-soft)] group-hover:scale-105">
              <Upload className="size-5" strokeWidth={1.75} />
            </span>
            <div>
              <p className="text-sm font-semibold tracking-tight text-fg">
                {ONBOARDING_LABELS.knowledge.dropzone.title}
              </p>
              <p className="mt-1 text-xs text-fg-muted">
                {ONBOARDING_LABELS.knowledge.dropzone.subtitle}
              </p>
            </div>
            <input
              ref={fileInput}
              type="file"
              multiple
              className="hidden"
              onChange={handleAddMock}
            />
          </button>

          <div>
            <p className="mb-3 text-[11px] font-semibold uppercase tracking-wider text-fg-subtle">
              {ONBOARDING_LABELS.knowledge.sources.title}
            </p>
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
              {SOURCES.map((s) => (
                <button
                  key={s.key}
                  type="button"
                  onClick={handleAddMock}
                  className="group flex items-center gap-2.5 rounded-xl border border-border bg-bg-elevated px-3 py-2.5 text-left text-sm font-medium text-fg transition-all duration-300 ease-[var(--ease-out-soft)] hover:border-border-strong hover:bg-bg-muted"
                >
                  <s.icon className="size-3.5 text-fg-muted transition-colors group-hover:text-fg" />
                  <span className="truncate">{s.label}</span>
                </button>
              ))}
            </div>
          </div>

          <div>
            <p className="mb-3 text-[11px] font-semibold uppercase tracking-wider text-fg-subtle">
              {ONBOARDING_LABELS.knowledge.uploaded.title}
              {state.documents.length > 0 && (
                <span className="ml-2 text-fg-subtle">· {state.documents.length}</span>
              )}
            </p>
            {state.documents.length === 0 ? (
              <p className="rounded-xl border border-dashed border-border bg-bg-muted/30 px-4 py-6 text-center text-sm text-fg-muted">
                {ONBOARDING_LABELS.knowledge.uploaded.empty}
              </p>
            ) : (
              <ul className="divide-y divide-border/60 rounded-xl border border-border bg-bg-elevated">
                {state.documents.map((d) => (
                  <li key={d.id} className="flex items-center gap-3 px-4 py-3 text-sm">
                    <FileText className="size-4 shrink-0 text-fg-muted" />
                    <div className="min-w-0 flex-1">
                      <div className="truncate font-medium text-fg">{d.name}</div>
                      <div className="text-xs text-fg-subtle">
                        {(d.sizeKb / 1024).toFixed(2)} MB · {d.chunks} chunks
                      </div>
                    </div>
                    <Tag className="text-[10px]">
                      <Globe className="size-2.5" />
                      {d.lang}
                    </Tag>
                    <span className="inline-flex items-center gap-1 text-xs text-success">
                      <CheckCircle2 className="size-3" />
                      {d.status}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </CardBody>
      </Card>

      <TipsCard />
    </div>
  );
}

function TipsCard() {
  return (
    <Card className="h-fit">
      <CardHeader
        title={
          <span className="inline-flex items-center gap-1.5">
            <Sparkles className="size-3.5 text-accent" />
            {ONBOARDING_LABELS.knowledge.tips.title}
          </span>
        }
      />
      <CardBody>
        <ul className="space-y-3">
          {ONBOARDING_LABELS.knowledge.tips.items.map((tip) => (
            <li key={tip} className="flex items-start gap-2.5 text-sm text-fg-muted">
              <span className="mt-1.5 size-1.5 shrink-0 rounded-full bg-accent" />
              <span className="leading-relaxed">{tip}</span>
            </li>
          ))}
        </ul>
      </CardBody>
    </Card>
  );
}
