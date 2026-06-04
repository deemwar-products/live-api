import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import { Container } from "@/components/ui/container";
import { ConsoleShell } from "@/components/layout/console-shell";
import { KpiStrip } from "@/pages/console/parts/kpi-strip";
import { LiveConversations } from "@/pages/console/parts/live-conversations";
import { RecentActivity } from "@/pages/console/parts/recent-activity";
import { TopicBreakdown } from "@/pages/console/parts/topic-breakdown";
import { KnowledgeHealth } from "@/pages/console/parts/knowledge-health";
import { AgentQueue } from "@/pages/console/parts/agent-queue";
import { BUSINESS_LABELS } from "@/labels";
import { CONSOLE_PAGE_LABELS } from "@/pages/console/labels";
import { isOnboardingComplete } from "@/mocks/business/onboarding";
import { CONSOLE } from "@/lib/paths";

const ease = [0.22, 1, 0.36, 1] as const;

function greetingFor(hour: number) {
  if (hour < 12) return BUSINESS_LABELS.greeting.goodMorning;
  if (hour < 18) return BUSINESS_LABELS.greeting.goodAfternoon;
  return BUSINESS_LABELS.greeting.goodEvening;
}

export function ConsolePage() {
  const navigate = useNavigate();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    document.title = CONSOLE_PAGE_LABELS.meta.title;
  }, []);

  useEffect(() => {
    if (!isOnboardingComplete()) {
      navigate(CONSOLE.onboarding, { replace: true });
      return;
    }
    setReady(true);
  }, [navigate]);

  const greeting = useMemo(() => {
    const hour = new Date().getHours();
    return greetingFor(hour);
  }, []);

  if (!ready) {
    return null;
  }

  return (
    <ConsoleShell>
      <Container size="wide" className="py-8 sm:py-10">
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease }}
          className="mb-8 flex flex-wrap items-end justify-between gap-3"
        >
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-fg sm:text-3xl">
              {greeting}, {CONSOLE_PAGE_LABELS.greetingName}.
            </h1>
            <p className="mt-1 text-sm text-fg-muted">
              {CONSOLE_PAGE_LABELS.orgName} · {BUSINESS_LABELS.greeting.subtitle}
            </p>
          </div>
        </motion.div>

        <div className="space-y-6">
          <KpiStrip />

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            <div className="lg:col-span-2">
              <LiveConversations />
            </div>
            <div>
              <RecentActivity />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <TopicBreakdown />
            <KnowledgeHealth />
          </div>

          <AgentQueue />
        </div>
      </Container>
    </ConsoleShell>
  );
}
