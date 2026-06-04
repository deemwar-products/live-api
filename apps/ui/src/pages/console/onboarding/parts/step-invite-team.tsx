import { useState } from "react";
import { Send, Users, X } from "lucide-react";
import { Card, CardHeader, CardBody } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Avatar } from "@/components/ui/avatar";
import { Tag } from "@/components/ui/tag";
import { ONBOARDING_LABELS } from "@/labels";
import type { Invite, OnboardingState, RoleKey } from "@/mocks/business/onboarding";

const ROLE_OPTIONS: { value: RoleKey; label: string }[] = [
  { value: "admin", label: ONBOARDING_LABELS.team.roles.admin.name },
  { value: "lead", label: ONBOARDING_LABELS.team.roles.lead.name },
  { value: "agent", label: ONBOARDING_LABELS.team.roles.agent.name },
  { value: "content", label: ONBOARDING_LABELS.team.roles.content.name },
  { value: "analyst", label: ONBOARDING_LABELS.team.roles.analyst.name },
];

const ROLE_DESCRIPTIONS: Record<RoleKey, string> = {
  admin: ONBOARDING_LABELS.team.roles.admin.perms,
  lead: ONBOARDING_LABELS.team.roles.lead.perms,
  agent: ONBOARDING_LABELS.team.roles.agent.perms,
  content: ONBOARDING_LABELS.team.roles.content.perms,
  analyst: ONBOARDING_LABELS.team.roles.analyst.perms,
};

type Props = {
  state: OnboardingState;
  onChange: (patch: Partial<OnboardingState>) => void;
};

export function StepInviteTeam({ state, onChange }: Props) {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<RoleKey>("agent");

  const sendInvite = () => {
    const v = email.trim();
    if (!v || !v.includes("@")) return;
    const invite: Invite = {
      id: `i${Date.now()}`,
      email: v,
      role,
      sentAt: "just now",
    };
    onChange({ invites: [...state.invites, invite] });
    setEmail("");
  };

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_360px]">
      <Card>
        <CardHeader
          title={ONBOARDING_LABELS.team.title}
          caption={ONBOARDING_LABELS.team.description}
        />
        <CardBody className="space-y-5">
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-[1fr_180px_auto]">
            <Input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  sendInvite();
                }
              }}
              placeholder={ONBOARDING_LABELS.team.inviteRow.emailPlaceholder}
              aria-label={ONBOARDING_LABELS.team.inviteRow.emailPlaceholder}
            />
            <Select
              value={role}
              onChange={(v) => setRole(v)}
              options={ROLE_OPTIONS}
              ariaLabel={ONBOARDING_LABELS.team.inviteRow.rolePlaceholder}
            />
            <Button onClick={sendInvite} leading={<Send className="size-3.5" />}>
              {ONBOARDING_LABELS.team.inviteRow.invite}
            </Button>
          </div>

          <div>
            <p className="mb-3 text-[11px] font-semibold uppercase tracking-wider text-fg-subtle">
              {ONBOARDING_LABELS.team.pending.title}
              {state.invites.length > 0 && (
                <span className="ml-2 text-fg-subtle">· {state.invites.length}</span>
              )}
            </p>
            {state.invites.length === 0 ? (
              <p className="rounded-xl border border-dashed border-border bg-bg-muted/30 px-4 py-6 text-center text-sm text-fg-muted">
                {ONBOARDING_LABELS.team.pending.empty}
              </p>
            ) : (
              <ul className="divide-y divide-border/60 rounded-xl border border-border bg-bg-elevated">
                {state.invites.map((inv) => (
                  <li key={inv.id} className="flex items-center gap-3 px-4 py-3 text-sm">
                    <Avatar name={inv.email} />
                    <div className="min-w-0 flex-1">
                      <div className="truncate font-medium text-fg">{inv.email}</div>
                      <div className="text-xs text-fg-subtle">Invited {inv.sentAt}</div>
                    </div>
                    <Tag>{ROLE_OPTIONS.find((r) => r.value === inv.role)?.label}</Tag>
                    <button
                      type="button"
                      onClick={() =>
                        onChange({ invites: state.invites.filter((i) => i.id !== inv.id) })
                      }
                      aria-label="Revoke invite"
                      className="inline-flex size-7 items-center justify-center rounded-full text-fg-subtle transition-colors duration-200 hover:bg-bg-muted hover:text-fg"
                    >
                      <X className="size-3.5" />
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </CardBody>
      </Card>

      <Card className="h-fit">
        <CardHeader
          title={
            <span className="inline-flex items-center gap-1.5">
              <Users className="size-3.5 text-accent" />
              {ONBOARDING_LABELS.team.roles.title}
            </span>
          }
        />
        <CardBody className="space-y-3">
          {ROLE_OPTIONS.map((r) => (
            <div key={r.value}>
              <div className="text-sm font-semibold text-fg">{r.label}</div>
              <p className="mt-0.5 text-xs leading-relaxed text-fg-muted">
                {ROLE_DESCRIPTIONS[r.value]}
              </p>
            </div>
          ))}
        </CardBody>
      </Card>
    </div>
  );
}
